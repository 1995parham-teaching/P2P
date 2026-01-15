package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"

	"github.com/1995parham-teaching/P2P/internal/cluster"
	"github.com/1995parham-teaching/P2P/internal/config"
	"github.com/1995parham-teaching/P2P/internal/tcp/client"
	tcp "github.com/1995parham-teaching/P2P/internal/tcp/server"
	udp "github.com/1995parham-teaching/P2P/internal/udp/server"
)

const (
	menuList     = "List cluster members"
	menuGet      = "Download a file"
	menuQuit     = "Quit"
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
			pterm.Error.Printf("TCP server error: %v\n", err)
		}
	}()

	// Start TCP client
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.TCPClient.Connect(n.ctx, n.Addr, n.fileName); err != nil {
			pterm.Error.Printf("TCP client error: %v\n", err)
		}
	}()

	// Start UDP server
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		if err := n.UDPServer.Up(n.ctx, n.TCPPort, n.Addr, n.fileName); err != nil {
			pterm.Error.Printf("UDP server error: %v\n", err)
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
	options := []string{menuList, menuGet, menuQuit}

	for {
		select {
		case <-n.ctx.Done():
			return nil
		default:
		}

		pterm.Println()
		selectedOption, err := pterm.DefaultInteractiveSelect.
			WithOptions(options).
			WithDefaultText("What would you like to do?").
			Show()

		if err != nil {
			// Handle Ctrl+C or other interrupts
			if err.Error() == "^C" {
				n.Shutdown()
				return nil
			}
			pterm.Error.Printf("Error reading input: %v\n", err)
			continue
		}

		switch selectedOption {
		case menuList:
			n.showClusterMembers()

		case menuGet:
			n.downloadFile()

		case menuQuit:
			n.Shutdown()
			return nil
		}
	}
}

func (n *Node) showClusterMembers() {
	list := n.UDPServer.Cluster.List()

	pterm.Println()
	if len(list) == 0 {
		pterm.Warning.Println("No cluster members found")
		return
	}

	// Create table data
	tableData := pterm.TableData{
		{"#", "Address"},
	}

	for i, addr := range list {
		tableData = append(tableData, []string{
			fmt.Sprintf("%d", i+1),
			addr,
		})
	}

	pterm.DefaultTable.
		WithHasHeader().
		WithBoxed().
		WithData(tableData).
		Render()

	pterm.Info.Printf("Total: %d member(s)\n", len(list))
}

func (n *Node) downloadFile() {
	fileName, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText("").
		Show("Enter filename to download")

	if err != nil {
		pterm.Error.Printf("Error: %v\n", err)
		return
	}

	if fileName == "" {
		pterm.Warning.Println("No filename provided")
		return
	}

	pterm.Info.Printf("Requesting file: %s\n", fileName)

	spinner, _ := pterm.DefaultSpinner.
		WithRemoveWhenDone(true).
		Start("Searching for file in cluster...")

	n.UDPServer.Req = fileName
	n.UDPServer.File(n.ctx)

	spinner.Stop()
}

// Shutdown gracefully stops the node
func (n *Node) Shutdown() {
	pterm.Println()
	spinner, _ := pterm.DefaultSpinner.Start("Shutting down...")

	n.cancel()

	// Close servers
	n.TCPServer.Close()
	n.UDPServer.Close()

	// Wait for goroutines to finish
	n.wg.Wait()

	spinner.Success("Shutdown complete")
}
