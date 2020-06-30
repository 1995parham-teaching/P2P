package node

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
	reliableUDPClient "github.com/elahe-dastan/reliable_UDP/udp/client"
	reliableUDPServer "github.com/elahe-dastan/reliable_UDP/udp/server"
	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/config"
	"github.com/elahe-dstn/p2p/message"
	"github.com/elahe-dstn/p2p/tcp/client"
	tcp "github.com/elahe-dstn/p2p/tcp/server"
	udp "github.com/elahe-dstn/p2p/udp/server"
)

const BUFFERSIZE = 1015

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
	fileSize  int64
	fileName  string
	folder    string
	newFile   *os.File
	Fin       bool
	received  int
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

	go func() {
		reader := bufio.NewReader(&n.reliableUDPClient)

		for {
			m, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
			}

			mes := message.ReliableUDPUnmarshal(string(m))

			switch t := mes.(type) {
			case *message.Get:
				file, err := os.Open(n.folder + "/" + t.Name)
				if err != nil {
					fmt.Println(err)
					return
				}

				fileInfo, err := file.Stat()
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Println("Sending filename and filesize!")

				fileSize := (&message.Size{Size: fileInfo.Size()}).Marshal()

				n.reliableUDPClient.Send([]byte(fileSize))

				fileName := (&response.FileName{Name: fileInfo.Name()}).Marshal()

				n.reliableUDPClient.Send([]byte(fileName))

				sendBuffer := make([]byte, BUFFERSIZE)

				fmt.Println("Start sending file")

				for {
					read, err := file.Read(sendBuffer)
					if err == io.EOF {
						break
					}

					sendBuff := sendBuffer[0:read]
					buffer := (&response.Segment{
						Part: sendBuff,
					}).Marshal()

					n.reliableUDPClient.Send([]byte(buffer))
				}

				fmt.Println("File has been sent, closing connection!")
			}
		}
	}()

	go n.reliableUDPClient.Connect(n.UAdder)

	go n.reliableUDPClient.Send([]byte((&request.Get{Name: <-n.UFName}).Marshal()))

	go func() {
		reader := bufio.NewReader(&n.reliableUDPClient)

		n.Fin = false
		n.received = 0

		for {
			if n.Fin {
				break
			}
			m, err := reader.ReadBytes('\n')
			if err != nil {
				fmt.Println(err)
			}

			mes := message.ReliableUDPUnmarshal(string(m))

			switch t := mes.(type) {
			case *response.Size:
				fmt.Println("received size the seq is")
				fmt.Println(t.Seq)

				n.fileSize = t.Size


			case *response.FileName:
				fmt.Println("received file name the seq is")
				fmt.Println(t.Seq)
				n.fileName = t.Name

				newFile, err := os.Create(filepath.Join(n.folder, filepath.Base("yep"+n.fileName)))
				if err != nil {
					fmt.Println(err)
				}

				n.newFile = newFile

			case *response.Segment:
				fmt.Println("received segment the seq is")
				fmt.Println(t.Seq)

				segment := t.Part

				received, err := n.newFile.Write(segment)
				if err != nil {
					fmt.Println(err)
				}

				n.received += received
				if int64(n.received) == n.fileSize {
					n.Fin = true
				}
			}
		}
	}()

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
