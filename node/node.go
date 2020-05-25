package node

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Node struct {
	ip              string
	folder          string
	Cluster         []string
	tcpPort			int
	mutex           *sync.Mutex
	server          chan string
	discoveryTicker *time.Ticker
	waiting         bool
	waitingTicker   *time.Ticker
	answeringNode   string
	req				string
}

func New(ip string, folder string, cluster []string) Node {
	return Node{
		ip:      ip,
		folder:  folder,
		Cluster: cluster,
		mutex:   &sync.Mutex{},
		server:  make(chan string),
	}
}

func (n *Node) Run() {
	go n.udpServer()
	time.Sleep(time.Second)

	go n.tcpServer()
	time.Sleep(time.Second)

	reader := bufio.NewReader(os.Stdin)

	n.discoveryTicker = time.NewTicker(3 * time.Second)

	go n.discover()
	time.Sleep(time.Second)

	for {
		print("Enter a file you want to download")
		text, err := reader.ReadString('\n')

		if err != nil {
			return
		}

		text = strings.TrimSuffix(text, "\n")
		n.req = text
		n.request(n.req)
	}

}

func (n *Node) udpServer() {
	addr := net.UDPAddr{
		IP:   net.ParseIP(n.ip),
		Port: 1378,
	}

	add, err := net.ResolveUDPAddr("udp", addr.String())
	print(add)

	ser, err := net.ListenUDP("udp", &addr)

	if err != nil {
		fmt.Println(err)
		return
	}

	message := make([]byte, 2048)

	_, remoteAddr, err := ser.ReadFromUDP(message)
	if err != nil {
		fmt.Println(err)
		return
	}

	n.protocol(message, ser, remoteAddr)
}

func transfer(conn *net.UDPConn, addr *net.UDPAddr, message string) {
	_, err := conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (n *Node) tcpServer()  {
	addr := net.TCPAddr{
		IP: net.ParseIP(n.ip),
		Port: 0,
	}

	l, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	n.tcpPort = l.Addr().(*net.TCPAddr).Port
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	message := make([]byte, 2048)

	_, err = c.Read(message)
	if err != nil {
		fmt.Println(err)
		return
	}

	go send(c)

}

func (n *Node) discover() {

	for {
		<-n.discoveryTicker.C
		n.UdpBroadcast("get")
	}
}

func (n *Node) merge(list []string) {
	for _, ip := range list {
		exists := false
		for _, c := range n.Cluster {
			if ip == c {
				exists = true
			}
		}

		if !exists {
			n.mutex.Lock()
			n.Cluster = append(n.Cluster, ip)
			n.mutex.Unlock()
		}
	}
}

func (n *Node) protocol(message []byte, ser *net.UDPConn, remoteAddr *net.UDPAddr) {
	protocol := strings.Split(string(message), ",")
	t := protocol[0]
	if t == "get" {
		go transfer(ser, remoteAddr, "this should be the list")
	} else if t == "list" {
		n.merge(protocol[1:])
	} else if t == "file" {
		if n.search(protocol[1]) {
			go transfer(ser, remoteAddr, fmt.Sprintf("ans,y,%d", n.tcpPort))
		}
	} else if t == "ans" {
		if n.waiting {
			n.check(protocol[1:])
		}
	}
}
func (n *Node) request(file string) {
	n.waiting = true
	n.waitingTicker = time.NewTicker(5 * time.Second)
	n.UdpBroadcast("file", file)
	<-n.waitingTicker.C
	n.waiting = false
	n.waitingTicker.Stop()
}

func (n *Node) UdpBroadcast(t string, options ...string) {
	for i, ip := range n.Cluster {
		conn, err := net.Dial("udp", ip+":1373")
		if err != nil {
			n.mutex.Lock()

			n.Cluster[i] = n.Cluster[len(n.Cluster)-1] // Copy last element to index i.
			n.Cluster[len(n.Cluster)-1] = ""           // Erase last element (write zero value).
			n.Cluster = n.Cluster[:len(n.Cluster)-1]   // Truncate slice.

			n.mutex.Unlock()
			return
		}

		o := ""
		for _, option := range options {
			o += option
			o += ","
		}
		o = strings.TrimSuffix(o, ",")

		r := fmt.Sprintf("%s,%s", t, o)

		_, err = conn.Write([]byte(r))

		if err != nil {
			fmt.Println(err)
		}
	}
}

func (n *Node) check(ans []string) {
	if ans[0] == "y" {
		n.waiting = false
		n.waitingTicker.Stop()
		n.answeringNode = ans[1]
	}

	connect()
}

func (n *Node) search(file string) bool {
	found := false

	err := filepath.Walk(n.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file == info.Name(){
			found = true
			return nil
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return found
}

func (n *Node) connect() {
	c, err := net.Dial("tcp", n.answeringNode)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = c.Write([]byte(n.req))
	if err != nil {
		fmt.Println(err)
		return
	}


}
//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}
