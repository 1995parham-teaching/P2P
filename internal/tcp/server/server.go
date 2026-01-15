package server

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"

	"github.com/1995parham-teaching/P2P/internal/config"
	"github.com/1995parham-teaching/P2P/internal/message"
)

type Server struct {
	TCPPort  int
	folder   string
	listener *net.TCPListener
	host     string
}

func New(folder string, host string) *Server {
	return &Server{
		folder: folder,
		host:   host,
	}
}

// Up starts the TCP server and listens for incoming connections
func (s *Server) Up(ctx context.Context, tcpPort chan<- int) error {
	addr := net.TCPAddr{
		IP:   net.ParseIP(s.host),
		Port: 0, // Let OS assign a port
	}

	listener, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		return err
	}
	s.listener = listener

	s.TCPPort = listener.Addr().(*net.TCPAddr).Port
	pterm.Success.Printf("TCP server listening on port %d\n", s.TCPPort)
	tcpPort <- s.TCPPort

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // Graceful shutdown
			default:
				pterm.Error.Printf("Failed to accept connection: %v\n", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	remoteAddr := conn.RemoteAddr().String()
	pterm.Info.Printf("TCP connection from %s\n", remoteAddr)

	buffer := make([]byte, config.UDPBufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		pterm.Error.Printf("Failed to read from %s: %v\n", remoteAddr, err)
		return
	}

	msg, err := message.Unmarshal(string(buffer[:n]))
	if err != nil {
		pterm.Error.Printf("Failed to unmarshal message from %s: %v\n", remoteAddr, err)
		return
	}

	getMsg, ok := msg.(*message.Get)
	if !ok {
		pterm.Warning.Printf("Expected Get message from %s, got something else\n", remoteAddr)
		return
	}

	pterm.Info.Printf("Peer %s requesting file '%s'\n", remoteAddr, getMsg.Name)

	if err := s.send(conn, getMsg.Name); err != nil {
		pterm.Error.Printf("Failed to send file to %s: %v\n", remoteAddr, err)
	}
}

func (s *Server) send(conn io.Writer, name string) error {
	// Use safe path to prevent directory traversal attacks
	filePath := safePath(s.folder, name)
	pterm.Debug.Printf("Resolved file path: %s\n", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), config.FileSizeLength, ':')
	fileName := fillString(fileInfo.Name(), config.FileNameLength, ':')

	pterm.Info.Printf("Sending file: %s (%d bytes)\n", fileInfo.Name(), fileInfo.Size())

	if _, err := conn.Write([]byte(fileSize)); err != nil {
		return err
	}

	if _, err := conn.Write([]byte(fileName)); err != nil {
		return err
	}

	// Create progress bar for upload
	progressBar, _ := pterm.DefaultProgressbar.
		WithTotal(int(fileInfo.Size())).
		WithTitle("Uploading " + fileInfo.Name()).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()

	sendBuffer := make([]byte, config.BufferSize)

	for {
		n, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			_, _ = progressBar.Stop()
			return err
		}

		if _, err := conn.Write(sendBuffer[:n]); err != nil {
			_, _ = progressBar.Stop()
			return err
		}

		progressBar.Add(n)
	}

	_, _ = progressBar.Stop()
	pterm.Success.Println("File sent successfully!")
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// safePath ensures the file path is within the allowed folder
func safePath(folder, filename string) string {
	safeName := filepath.Base(filename)
	return filepath.Join(folder, safeName)
}

// fillString pads a string to the specified length
func fillString(s string, length int, padding byte) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(string(padding), length-len(s))
}
