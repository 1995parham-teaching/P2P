package node

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
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
	reliableUDPServeAddr := cfg.ReliableUDPServer

	UPort, _ := strconv.Atoi(strings.Split(reliableUDPServeAddr, ":")[1])

	udpServer := udp.New(ip, port, &clu, time.NewTicker(time.Duration(d)*time.Second), waitingDuration, folder, method, UPort)

	return Node {
		UDPServer: udpServer,
		TCPServer: tcp.New(folder),
		TCPClient: client.New(folder),
		TCPPort:   make(chan int),
		Addr:      make(chan string, 1),
		fName:     make(chan string),
		reliableUDPServer:reliableUDPServer.New(reliableUDPServeAddr, folder),
		reliableUDPClient:reliableUDPClient.New(folder),
		UAdder:    make(chan string),
		UFName:	   make(chan string),
	}
}

func (n *Node) Run() {
	reader := bufio.NewReader(os.Stdin)

	go n.TCPServer.Up(n.TCPPort)

	go n.TCPClient.Connect(n.Addr, n.fName)

	go n.reliableUDPServer.Up()

	go n.reliableUDPClient.Connect(n.UAdder)

	go n.reliableUDPClient.Send([]byte((&request.Get{Name: <-n.UFName}).Marshal()))

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
