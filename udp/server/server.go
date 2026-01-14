package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/1995parham-teaching/P2P/cluster"
	"github.com/1995parham-teaching/P2P/config"
	"github.com/1995parham-teaching/P2P/internal/utils"
	"github.com/1995parham-teaching/P2P/message"
)

type Server struct {
	IP              string
	Port            int
	Cluster         *cluster.Cluster
	DiscoveryTicker *time.Ticker
	waitingDuration time.Duration
	Req             string
	folder          string
	conn            *net.UDPConn
	method          int
	UDPPort         int

	// File request state
	waiting       bool
	waitingMutex  sync.Mutex
	waitingCancel context.CancelFunc

	// Priority responders tracking
	prior      []string
	priorMutex sync.RWMutex

	// File index cache
	fileIndex      map[string]string // filename -> full path
	fileIndexMutex sync.RWMutex
}

func New(ip string, port int, cluster *cluster.Cluster,
	ticker *time.Ticker, waitingDuration int, folder string, method int, udpPort int) *Server {
	s := &Server{
		IP:              ip,
		Port:            port,
		Cluster:         cluster,
		DiscoveryTicker: ticker,
		waitingDuration: time.Duration(waitingDuration) * time.Second,
		folder:          folder,
		prior:           make([]string, 0),
		method:          method,
		UDPPort:         udpPort,
		fileIndex:       make(map[string]string),
	}

	// Build initial file index
	s.rebuildFileIndex()

	return s
}

// Up starts the UDP server and listens for incoming messages
func (s *Server) Up(ctx context.Context, tcpPort <-chan int, request chan<- string, fName chan<- string, uRequest chan<- string, uFName chan<- string) error {
	tPort := <-tcpPort

	addr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: s.Port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %w", err)
	}
	s.conn = conn

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buffer := make([]byte, config.UDPBufferSize)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Printf("UDP read error: %v\n", err)
				continue
			}
		}

		// Parse the message
		msgStr := strings.TrimSpace(string(buffer[:n]))
		fmt.Println("Received:", msgStr)

		msg, err := message.Unmarshal(msgStr)
		if err != nil {
			fmt.Printf("Failed to unmarshal message: %v\n", err)
			continue
		}

		s.handleMessage(ctx, msg, remoteAddr, tPort, request, fName, uRequest, uFName)
	}
}

func (s *Server) handleMessage(ctx context.Context, msg message.Message, remoteAddr *net.UDPAddr, tcpPort int, request chan<- string, fName chan<- string, uRequest chan<- string, uFName chan<- string) {
	fmt.Println("Processing message")

	switch t := msg.(type) {
	case *message.Discover:
		host := fmt.Sprintf("%s:%d", s.IP, s.Port)
		s.Cluster.Merge(host, t.List)

	case *message.Get:
		if s.Search(t.Name) {
			go s.transfer(remoteAddr, (&message.File{
				Method:  s.method,
				TCPPort: tcpPort,
				UDPPort: s.UDPPort,
			}).Marshal())
		}

	case *message.File:
		s.waitingMutex.Lock()
		isWaiting := s.waiting
		s.waitingMutex.Unlock()

		if isWaiting {
			// Add to prior list
			s.addToPrior(remoteAddr.String())

			// Cancel waiting
			s.waitingMutex.Lock()
			s.waiting = false
			if s.waitingCancel != nil {
				s.waitingCancel()
			}
			s.waitingMutex.Unlock()

			ip := remoteAddr.IP.String()

			if t.Method == config.TransferMethodTCP {
				request <- fmt.Sprintf("%s:%d", ip, t.TCPPort)
				fName <- s.Req
			} else {
				uRequest <- fmt.Sprintf("%s:%d", ip, t.UDPPort)
				uFName <- s.Req
			}
		}
	}
}

func (s *Server) transfer(addr *net.UDPAddr, msg string) {
	// Check if this is a priority responder
	if !s.isPrior(addr.String()) {
		time.Sleep(config.NonPriorResponseDelay)
	}

	if _, err := s.conn.WriteToUDP([]byte(msg), addr); err != nil {
		fmt.Printf("Failed to send transfer message: %v\n", err)
	}
}

// Discover periodically broadcasts cluster information
func (s *Server) Discover(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.DiscoveryTicker.C:
			list := s.Cluster.List()
			msg := (&message.Discover{List: list}).Marshal()
			if err := s.Cluster.Broadcast(s.conn, msg); err != nil {
				fmt.Printf("Discovery broadcast error: %v\n", err)
			}
		}
	}
}

// File broadcasts a file request to the cluster
func (s *Server) File(ctx context.Context) {
	s.waitingMutex.Lock()
	s.waiting = true
	waitCtx, cancel := context.WithTimeout(ctx, s.waitingDuration)
	s.waitingCancel = cancel
	s.waitingMutex.Unlock()

	msg := (&message.Get{Name: s.Req}).Marshal()
	if err := s.Cluster.Broadcast(s.conn, msg); err != nil {
		fmt.Printf("File request broadcast error: %v\n", err)
	}

	// Wait for timeout or response
	<-waitCtx.Done()

	s.waitingMutex.Lock()
	s.waiting = false
	s.waitingCancel = nil
	s.waitingMutex.Unlock()
}

// Search checks if a file exists in the shared folder
func (s *Server) Search(filename string) bool {
	// Sanitize filename to prevent path traversal
	safeFilename := filepath.Base(filename)

	// Check file index cache first
	s.fileIndexMutex.RLock()
	_, found := s.fileIndex[safeFilename]
	s.fileIndexMutex.RUnlock()

	if found {
		return true
	}

	// Rebuild index and check again (file might have been added)
	s.rebuildFileIndex()

	s.fileIndexMutex.RLock()
	_, found = s.fileIndex[safeFilename]
	s.fileIndexMutex.RUnlock()

	return found
}

// rebuildFileIndex scans the folder and rebuilds the file index
func (s *Server) rebuildFileIndex() {
	s.fileIndexMutex.Lock()
	defer s.fileIndexMutex.Unlock()

	// Clear existing index
	s.fileIndex = make(map[string]string)

	err := filepath.Walk(s.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if !info.IsDir() {
			s.fileIndex[info.Name()] = path
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error rebuilding file index: %v\n", err)
	}
}

// GetFilePath returns the full path for a filename if it exists
func (s *Server) GetFilePath(filename string) (string, bool) {
	safeFilename := filepath.Base(filename)

	s.fileIndexMutex.RLock()
	path, found := s.fileIndex[safeFilename]
	s.fileIndexMutex.RUnlock()

	return path, found
}

func (s *Server) addToPrior(addr string) {
	s.priorMutex.Lock()
	defer s.priorMutex.Unlock()

	if !utils.Contains(s.prior, addr) {
		s.prior = append(s.prior, addr)
	}
}

func (s *Server) isPrior(addr string) bool {
	s.priorMutex.RLock()
	defer s.priorMutex.RUnlock()

	return utils.Contains(s.prior, addr)
}

// Close shuts down the UDP server
func (s *Server) Close() error {
	s.DiscoveryTicker.Stop()
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
