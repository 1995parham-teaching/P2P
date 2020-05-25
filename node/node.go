package node

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Node struct {
	ip              string
	Cluster         []string
	mutex           *sync.Mutex
	server          chan string
	discoveryTicker *time.Ticker
	waiting         bool
	waitingTicker *time.Ticker
	answeringNode 	string
}

func New(ip string, cluster []string) Node {
	return Node {
		ip:      ip,
		Cluster: cluster,
		mutex:	&sync.Mutex{},
		server:	make(chan string),
	}
}

func (n *Node) Run() {
	go n.udpServer()
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
		n.request(text)
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

func sendList(conn *net.UDPConn, addr *net.UDPAddr) {
	_,err := conn.WriteToUDP([]byte("This should be the list"), addr)
	if err != nil {
		fmt.Println(err)
		return
	}
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
		go sendList(ser, remoteAddr)
	}else if t == "list" {
		n.merge(protocol[1:])
	}else if t == "file" {
		//send yes or no
	}else if t == "ans" {
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

func (n *Node) UdpBroadcast(t string, options ...string)  {
	for i ,ip := range n.Cluster {
		conn, err := net.Dial("udp", ip + ":1373")
		if err != nil {
			n.mutex.Lock()

			n.Cluster[i] = n.Cluster[len(n.Cluster)-1] // Copy last element to index i.
			n.Cluster[len(n.Cluster)-1] = ""   // Erase last element (write zero value).
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

		_,err = conn.Write([]byte(r))

		if err != nil {
			fmt.Println(err)
		}
	}
}

func (n *Node) check(ans []string)  {
	if ans[0] == "y" {
		n.waiting = false
		n.waitingTicker.Stop()
		n.answeringNode = ans[1]
	}

	// ask for file
}
//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}

