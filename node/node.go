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
	reliableUDPServer reliableUDPServer.Server
	reliableUDPClient reliableUDPClient.Client
	UAdder    chan string
	UFName    chan string
}

func New(folder string, c []string) Node {
	clu := cluster.New(c)

	cfg := config.Read()

	ip := cfg.Host
	port := cfg.Port
	d := cfg.DiscoveryPeriod
	waitingDuration := cfg.WaitingTime
	method := cfg.Type

	udpServer := udp.New(ip, port, &clu, time.NewTicker(time.Duration(d)*time.Second), waitingDuration, folder, method)

	return Node {
		UDPServer: udpServer,
		TCPServer: tcp.New(folder),
		TCPClient: client.New(folder),
		TCPPort:   make(chan int),
		Addr:      make(chan string, 1),
		fName:     make(chan string),
		UAdder:    make(chan string),
		UFName:	   make(chan string),
	}
}

func (n *Node) Run() {
	reader := bufio.NewReader(os.Stdin)

	go n.TCPServer.Up(n.TCPPort)

	go n.TCPClient.Connect(n.Addr, n.fName)

	go n.reliableUDPServer.Up()

	go n.reliableUDPClient.Connect(<-n.UAdder,<- n.UFName)
	fmt.Println("pass")

	go n.UDPServer.Up(n.TCPPort, n.Addr, n.fName, n.UAdder, n.UFName)

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
