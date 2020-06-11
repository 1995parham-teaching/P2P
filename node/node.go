package node

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	reliableUDPClient "github.com/elahe-dastan/reliable_UDP/udp/client"
	reliableUDPServer "github.com/elahe-dastan/reliable_UDP/udp/server"
	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/config"
	"github.com/elahe-dstn/p2p/tcp/client"
	tcp "github.com/elahe-dstn/p2p/tcp/server"
	udp "github.com/elahe-dstn/p2p/udp/server"
)

type Node struct {
	UDPServer udp.Server
	TCPServer tcp.Server
	TCPClient client.Client
	TCPPort   chan int
	Addr      chan string
	fName     chan string
	approach  int
	reliableUDPServer reliableUDPServer.Server
	reliableUDPClient reliableUDPClient.Client
}

func New(folder string, c []string, approach int) Node {
	clu := cluster.New(c)

	cfg := config.Read()

	ip := cfg.Host
	port := cfg.Port
	d := cfg.DiscoveryPeriod
	waitingDuration := cfg.WaitingTime

	udpServer := udp.New(ip, port, &clu, time.NewTicker(time.Duration(d)*time.Second), waitingDuration, folder, approach)

	if approach == 1 {
		return Node {
			UDPServer: udpServer,
			TCPServer: tcp.New(folder),
			TCPClient: client.New(folder),
			TCPPort:   make(chan int),
			Addr:      make(chan string, 1),
			fName:     make(chan string),
			approach:approach,
		}
	}else {
		return Node {
			UDPServer: udp.Server{},
			TCPServer: tcp.Server{},
			TCPClient: client.Client{},
			TCPPort:   nil,
			Addr:      nil,
			fName:     nil,
			approach:approach,
			reliableUDPServer:reliableUDPServer.New("127.0.0.1", 1995, folder),
			reliableUDPClient:reliableUDPClient.New(folder),
		}
	}
}

func (n *Node) Run() {
	reader := bufio.NewReader(os.Stdin)

	//if n.approach == 1 {
		go n.TCPServer.Up(n.TCPPort)

		go n.TCPClient.Connect(n.Addr, n.fName)

		go n.reliableUDPServer.Up()

		go n.reliableUDPClient.Connect()
	//}

	go n.UDPServer.Up(n.TCPPort, n.Addr, n.fName)

	go n.UDPServer.Discover()

	for {
		fmt.Println("Enter a file you want to download or list to see the cluster")

		text, err := reader.ReadString('\n')


		if err != nil {
			fmt.Println(err)
			return
		}

		text = strings.TrimSuffix(text, "\n")

		fmt.Println(text)

		req := strings.Split(text, " ")

		if req[0] == "list" {
			fmt.Println(n.UDPServer.Cluster.List)
		}else if req[0] == "get" {
			n.UDPServer.Req = req[1]
			n.UDPServer.File()
			//n.fName <- req[1]
		}
	}
}
