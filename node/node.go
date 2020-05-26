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

	"github.com/elahe-dstn/p2p/tcp"
	"github.com/elahe-dstn/p2p/udp/client"
	"github.com/elahe-dstn/p2p/udp/server"
)

type Node struct {
	IP            string
	folder        string
	Cluster       []string
	TcpPort       int
	Mutex         *sync.Mutex
	answeringNode string
	Req           string
	UdpClient     client.Client
}

func New(ip string, folder string, cluster []string) Node {
	return Node{
		IP:        ip,
		folder:    folder,
		Cluster:   cluster,
		Mutex:     &sync.Mutex{},
		UdpClient: client.New(time.NewTicker(3 * time.Second)),
	}
}

func (n *Node) Run() {
	go server.Server(n)
	time.Sleep(time.Second)

	go tcp.Server(n)
	time.Sleep(time.Second)

	reader := bufio.NewReader(os.Stdin)

	go n.UdpClient.Discover(n)
	time.Sleep(time.Second)

	for {
		print("Enter a file you want to download")
		text, err := reader.ReadString('\n')

		if err != nil {
			return
		}

		text = strings.TrimSuffix(text, "\n")
		n.Req = text
		n.UdpClient.File(n)
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
			n.Mutex.Lock()
			n.Cluster = append(n.Cluster, ip)
			n.Mutex.Unlock()
		}
	}
}


func (n *Node) Search(file string) bool {
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

	_, err = c.Write([]byte(n.Req))
	if err != nil {
		fmt.Println(err)
		return
	}


}
//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}
