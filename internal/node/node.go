package node

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/1995parham-teaching/P2P/internal/cluster"
	"github.com/1995parham-teaching/P2P/internal/config"
	"github.com/1995parham-teaching/P2P/internal/tcp/client"
	tcp "github.com/1995parham-teaching/P2P/internal/tcp/server"
	udp "github.com/1995parham-teaching/P2P/internal/udp/server"
)

type Node struct {
	UDPServer *udp.Server
	TCPServer *tcp.Server
	TCPClient *client.Client
	TCPPort   chan int
	Addr      chan string
	fileName  chan string
	folder    string

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(folder string, clusterList []string) (*Node, error) {
	clu := cluster.New(clusterList)
	cfg := config.Read()

	udpServer := udp.New(
		cfg.Host,
		cfg.Port,
		clu,
		time.NewTicker(time.Duration(cfg.DiscoveryPeriod)*time.Second),
		cfg.WaitingTime,
		folder,
	)

	ctx, cancel := context.WithCancel(context.Background())

	return &Node{
		UDPServer: udpServer,
		TCPServer: tcp.New(folder, cfg.Host),
		TCPClient: client.New(folder),
		TCPPort:   make(chan int, 1),
		Addr:      make(chan string, 1),
		fileName:  make(chan string, 1),
		folder:    folder,
		ctx:       ctx,
		cancel:    cancel,
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

	// Start UDP server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.UDPServer.Up(n.ctx, n.TCPPort, n.Addr, n.fileName); err != nil {
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

// Shutdown gracefully stops the node
func (n *Node) Shutdown() {
	fmt.Println("Shutting down...")
	n.cancel()

	// Close servers
	n.TCPServer.Close()
	n.UDPServer.Close()

	// Wait for goroutines to finish
	n.wg.Wait()
	fmt.Println("Shutdown complete")
}
