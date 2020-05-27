package node

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/udp/client"
	"github.com/elahe-dstn/p2p/udp/server"
)

type Node struct {
	folder        string
	TcpPort       int
	answeringNode string
	UdpClient     client.Client
	UdpServer     server.Server
}

func New(folder string, c []string) Node {
	clu := cluster.New(c)
	return Node{
		folder:    folder,
		UdpClient: client.New(&clu),
		UdpServer: server.New(&clu, time.NewTicker(3*time.Second)),
	}
}

func (n *Node) Run() {
	go n.UdpServer.Up()
	time.Sleep(time.Second)

	//go tcp.Server(n)
	//time.Sleep(time.Second)

	reader := bufio.NewReader(os.Stdin)

	go n.UdpClient.Discover()
	time.Sleep(time.Second)

	for {
		print("Enter a file you want to download")
		text, err := reader.ReadString('\n')

		fmt.Println(text)

		if err != nil {
			return
		}

		text = strings.TrimSuffix(text, "\n")

		fmt.Println(text)
		n.UdpClient.Req = text
		//n.UdpClient.File(n)
	}

}

func (n *Node) Search(file string) bool {
	found := false

	err := filepath.Walk(n.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file == info.Name() {
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

//func (n *Node) connect() {
//	c, err := net.Dial("tcp", n.answeringNode)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	_, err = c.Write([]byte(n.Req))
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//}

//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}
