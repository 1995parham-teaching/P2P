package node

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elahe-dastan/reliable_UDP/request"
	"github.com/elahe-dastan/reliable_UDP/response"
	reliableUDPClient "github.com/elahe-dastan/reliable_UDP/udp/client"
	reliableUDPServer "github.com/elahe-dastan/reliable_UDP/udp/server"
	"github.com/1995parham-teaching/P2P/cluster"
	"github.com/1995parham-teaching/P2P/config"
	"github.com/1995parham-teaching/P2P/internal/utils"
	"github.com/1995parham-teaching/P2P/message"
	"github.com/1995parham-teaching/P2P/tcp/client"
	tcp "github.com/1995parham-teaching/P2P/tcp/server"
	udp "github.com/1995parham-teaching/P2P/udp/server"
)

type Node struct {
	UDPServer         *udp.Server
	TCPServer         *tcp.Server
	TCPClient         *client.Client
	TCPPort           chan int
	Addr              chan string
	fileName          chan string
	reliableUDPServer reliableUDPServer.Server
	reliableUDPClient reliableUDPClient.Client
	UDPAddr           chan string
	UDPFileName       chan string
	folder            string

	// File reception state
	fileSize    int64
	currentFile *os.File
	received    int
	finished    bool
	fileMutex   sync.Mutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(folder string, clusterList []string) (*Node, error) {
	clu := cluster.New(clusterList)
	cfg := config.Read()

	udpPort, err := strconv.Atoi(strings.Split(cfg.ReliableUDPServer, ":")[1])
	if err != nil {
		return nil, fmt.Errorf("invalid reliable UDP server address: %w", err)
	}

	udpServer := udp.New(
		cfg.Host,
		cfg.Port,
		clu,
		time.NewTicker(time.Duration(cfg.DiscoveryPeriod)*time.Second),
		cfg.WaitingTime,
		folder,
		cfg.Type,
		udpPort,
	)

	ctx, cancel := context.WithCancel(context.Background())

	return &Node{
		UDPServer:         udpServer,
		TCPServer:         tcp.New(folder, cfg.Host),
		TCPClient:         client.New(folder),
		TCPPort:           make(chan int, 1),
		Addr:              make(chan string, 1),
		fileName:          make(chan string, 1),
		reliableUDPServer: reliableUDPServer.New(cfg.ReliableUDPServer, folder),
		reliableUDPClient: reliableUDPClient.New(folder),
		UDPAddr:           make(chan string, 1),
		UDPFileName:       make(chan string, 1),
		folder:            folder,
		ctx:               ctx,
		cancel:            cancel,
	}, nil
}

// Run starts the node and all its services
func (n *Node) Run() error {
	// Start TCP server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.TCPServer.Up(n.ctx, n.TCPPort); err != nil {
			fmt.Printf("TCP server error: %v\n", err)
		}
	}()

	// Start TCP client
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.TCPClient.Connect(n.ctx, n.Addr, n.fileName); err != nil {
			fmt.Printf("TCP client error: %v\n", err)
		}
	}()

	// Start reliable UDP server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.reliableUDPServer.Up()
	}()

	// Start reliable UDP sender handler
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.handleReliableUDPSend()
	}()

	// Start reliable UDP client
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.reliableUDPClient.Connect(n.UDPAddr)
	}()

	// Start file request sender for reliable UDP
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		select {
		case <-n.ctx.Done():
			return
		case fName := <-n.UDPFileName:
			n.reliableUDPClient.Send([]byte((&request.Get{Name: fName}).Marshal()))
		}
	}()

	// Start reliable UDP receiver handler
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.handleReliableUDPReceive()
	}()

	// Start UDP server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.UDPServer.Up(n.ctx, n.TCPPort, n.Addr, n.fileName, n.UDPAddr, n.UDPFileName); err != nil {
			fmt.Printf("UDP server error: %v\n", err)
		}
	}()

	// Start discovery broadcasts
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.UDPServer.Discover(n.ctx)
	}()

	// Handle user input
	return n.handleUserInput()
}

func (n *Node) handleUserInput() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-n.ctx.Done():
			return nil
		default:
		}

		fmt.Println("Enter a file you want to download or 'list' to see the cluster ('quit' to exit)")

		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		parts := strings.SplitN(text, " ", 2)
		command := parts[0]

		switch command {
		case "list":
			list := n.UDPServer.Cluster.List()
			fmt.Println("Cluster members:")
			for i, addr := range list {
				fmt.Printf("  %d. %s\n", i+1, addr)
			}
			if len(list) == 0 {
				fmt.Println("  (no members)")
			}

		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <filename>")
				continue
			}
			fileName := parts[1]
			n.UDPServer.Req = fileName
			n.UDPServer.File(n.ctx)

		case "quit", "exit":
			n.Shutdown()
			return nil

		default:
			fmt.Println("Unknown command. Use 'list', 'get <filename>', or 'quit'")
		}
	}
}

func (n *Node) handleReliableUDPSend() {
	reader := bufio.NewReader(&n.reliableUDPClient)

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}

		m, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Reliable UDP read error: %v\n", err)
			}
			continue
		}

		msg, err := message.ReliableUDPUnmarshal(string(m))
		if err != nil {
			fmt.Printf("Failed to unmarshal reliable UDP message: %v\n", err)
			continue
		}

		switch t := msg.(type) {
		case *message.Get:
			if err := n.sendFileViaReliableUDP(t.Name); err != nil {
				fmt.Printf("Failed to send file: %v\n", err)
			}
		}
	}
}

func (n *Node) sendFileViaReliableUDP(fileName string) error {
	// Use safe path to prevent directory traversal
	filePath := utils.SafePath(n.folder, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fmt.Println("Sending filename and filesize!")

	// Send file size
	sizeMsg := (&message.Size{Size: fileInfo.Size()}).Marshal()
	n.reliableUDPClient.Send([]byte(sizeMsg))

	// Send file name
	nameMsg := (&response.FileName{Name: fileInfo.Name()}).Marshal()
	n.reliableUDPClient.Send([]byte(nameMsg))

	// Send file content in segments
	sendBuffer := make([]byte, config.BufferSize-9) // Leave room for protocol overhead

	fmt.Println("Start sending file")

	for {
		bytesRead, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		segment := (&response.Segment{
			Part: sendBuffer[:bytesRead],
		}).Marshal()

		n.reliableUDPClient.Send([]byte(segment))
	}

	fmt.Println("File has been sent!")
	return nil
}

func (n *Node) handleReliableUDPReceive() {
	reader := bufio.NewReader(&n.reliableUDPClient)

	n.fileMutex.Lock()
	n.finished = false
	n.received = 0
	n.fileMutex.Unlock()

	for {
		select {
		case <-n.ctx.Done():
			return
		default:
		}

		n.fileMutex.Lock()
		isFinished := n.finished
		n.fileMutex.Unlock()

		if isFinished {
			// Reset for next file
			n.fileMutex.Lock()
			n.finished = false
			n.received = 0
			n.fileMutex.Unlock()
		}

		m, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Reliable UDP receive error: %v\n", err)
			}
			continue
		}

		msg, err := message.ReliableUDPUnmarshal(string(m))
		if err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			continue
		}

		if err := n.processReceivedMessage(msg); err != nil {
			fmt.Printf("Error processing message: %v\n", err)
		}
	}
}

func (n *Node) processReceivedMessage(msg message.Message) error {
	switch t := msg.(type) {
	case *response.Size:
		fmt.Printf("Received file size: %d bytes\n", t.Size)
		n.fileMutex.Lock()
		n.fileSize = t.Size
		n.fileMutex.Unlock()

	case *response.FileName:
		fmt.Printf("Received file name: %s\n", t.Name)

		// Create output file with safe path
		safeName := filepath.Base(t.Name)
		outputPath := filepath.Join(n.folder, "receiving_"+safeName)

		newFile, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		n.fileMutex.Lock()
		n.currentFile = newFile
		n.fileMutex.Unlock()

	case *response.Segment:
		fmt.Println("Received segment")

		n.fileMutex.Lock()
		defer n.fileMutex.Unlock()

		if n.currentFile == nil {
			return fmt.Errorf("received segment but no file is open")
		}

		bytesWritten, err := n.currentFile.Write(t.Part)
		if err != nil {
			return fmt.Errorf("failed to write segment: %w", err)
		}

		n.received += bytesWritten

		if int64(n.received) >= n.fileSize {
			n.finished = true
			if err := n.currentFile.Close(); err != nil {
				return fmt.Errorf("failed to close file: %w", err)
			}
			n.currentFile = nil
			fmt.Println("File received completely!")
		}
	}

	return nil
}

// Shutdown gracefully stops the node
func (n *Node) Shutdown() {
	fmt.Println("Shutting down...")
	n.cancel()

	// Close any open file
	n.fileMutex.Lock()
	if n.currentFile != nil {
		n.currentFile.Close()
		n.currentFile = nil
	}
	n.fileMutex.Unlock()

	// Close servers
	n.TCPServer.Close()
	n.UDPServer.Close()

	// Wait for goroutines to finish
	n.wg.Wait()
	fmt.Println("Shutdown complete")
}
