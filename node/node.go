package node

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Node struct {
	ip string
	Cluster []string
	mutex *sync.Mutex
}

func New(ip string, cluster []string) Node {
	return Node {
		ip:      ip,
		Cluster: cluster,
		mutex:	&sync.Mutex{},
	}
}

func (n *Node) Run() {
	go udpServer()
	go n.discover()
	reader := bufio.NewReader(os.Stdin)

	for {
		print("Enter a file you want to download")
		text, err := reader.ReadString('\n')

		if err != nil {
			return
		}

		text = strings.TrimSuffix(text, "\n")

	}

}

func udpServer() {
	addr := net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 1373,
	}
	
	ser, err := net.ListenUDP("udp", &addr)
	
	if err != nil {
		fmt.Println(err)
		return
	}

	_, remoteAddr, err := ser.ReadFromUDP(make([]byte, 2048))
	if err != nil {
		fmt.Println(err)
		return
	}
	
	go sendList(ser, remoteAddr)
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
	for {
		for i ,ip := range n.Cluster {
			conn, err := net.Dial("udp", ip + ":1373")
			if err != nil {
				n.mutex.Lock()

				n.Cluster[i] = n.Cluster[len(n.Cluster)-1] // Copy last element to index i.
				n.Cluster[len(n.Cluster)-1] = ""   // Erase last element (write zero value).
				n.Cluster = n.Cluster[:len(n.Cluster)-1]   // Truncate slice.

				n.mutex.Unlock()
			}

			_, err = bufio.NewReader(conn).Read(p)

			if err != nil {
				fmt.Println(err)
			}

			merge(fmt.Sprintf("%s",p))
		}
	}

}

func merge(list string)  {
	// this needs a protocol
}
//func (n *Node) request(file string) {
//	ready := make([]string, 0)
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

