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
	ip string
	Cluster []string
	mutex *sync.Mutex
	server chan string
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
	ticker := time.NewTicker(3 * time.Second)
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-ticker.C:
			n.discover()
		default:
			print("Enter a file you want to download")
			text, err := reader.ReadString('\n')

			if err != nil {
				return
			}

			text = strings.TrimSuffix(text, "\n")
		}
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
	p := make([]byte, 2040)
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

		_,err = conn.Write([]byte("get"))

		if err != nil {
			fmt.Println(err)
		}
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
		n.merge(protocol[1:len(protocol)])
	}
}
//func (n *Node) request(file string) {
//	ready := make([]stri[]ng, 0)
//
//	for _, node := range n.Cluster {
//		//if node.get(file){
//		//	ready = append(ready, node)
//		//}
//	}
//}
//
//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}

