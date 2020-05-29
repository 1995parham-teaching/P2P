package node

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/tcp/client"
	tcp "github.com/elahe-dstn/p2p/tcp/server"
	udp "github.com/elahe-dstn/p2p/udp/server"
)

type Node struct {
	UdpServer udp.Server
	TcpServer tcp.Server
	TcpClient client.Client
	TcpPort   chan int
	Addr      chan string
	fName     chan string
}

func New(folder string, c []string) Node {
	clu := cluster.New(c)
	return Node{
		UdpServer: udp.New(&clu, time.NewTicker(20*time.Second), folder),
		TcpServer: tcp.New(folder),
		TcpClient: client.New(folder),
		TcpPort:   make(chan int, 0),
		Addr:      make(chan string, 0),
		fName:     make(chan string, 0),
	}
}

func (n *Node) Run() {
	reader := bufio.NewReader(os.Stdin)

	go n.TcpServer.Up(n.TcpPort)

	go n.UdpServer.Up(n.TcpPort, n.Addr)

	go n.UdpServer.Discover()

	go n.TcpClient.Connect(n.Addr, n.fName)

	for {
		print("Enter a file you want to download")
		text, err := reader.ReadString('\n')

		fmt.Println(text)

		if err != nil {
			return
		}

		text = strings.TrimSuffix(text, "\n")

		fmt.Println(text)
		n.UdpServer.Req = text
		n.UdpServer.File()
		n.fName <- text
	}

}

//// returns true if has the file
//func (n *Node) get(file string) bool {
//
//}
