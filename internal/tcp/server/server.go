package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	s.listener = listener

	s.TCPPort = listener.Addr().(*net.TCPAddr).Port
	tcpPort <- s.TCPPort

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // Graceful shutdown
			default:
				fmt.Printf("Failed to accept connection: %v\n", err)
				continue
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, config.UDPBufferSize)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read from connection: %v\n", err)
		return
	}

	msg, err := message.Unmarshal(string(buffer[:n]))
	if err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	getMsg, ok := msg.(*message.Get)
	if !ok {
		fmt.Println("Expected Get message, got something else")
		return
	}

	if err := s.send(conn, getMsg.Name); err != nil {
		fmt.Printf("Failed to send file: %v\n", err)
	}
}

func (s *Server) send(conn io.Writer, name string) error {
	fmt.Println("A client has connected!")

	// Use safe path to prevent directory traversal attacks
	filePath := safePath(s.folder, name)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), config.FileSizeLength, ':')
	fileName := fillString(fileInfo.Name(), config.FileNameLength, ':')

	fmt.Println("Sending filename and filesize!")

	if _, err := conn.Write([]byte(fileSize)); err != nil {
		return fmt.Errorf("failed to write file size: %w", err)
	}

	if _, err := conn.Write([]byte(fileName)); err != nil {
		return fmt.Errorf("failed to write file name: %w", err)
	}

	sendBuffer := make([]byte, config.BufferSize)

	fmt.Println("Start sending file")

	for {
		n, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		if _, err := conn.Write(sendBuffer[:n]); err != nil {
			return fmt.Errorf("failed to write file data: %w", err)
		}
	}

	fmt.Println("File has been sent, closing connection!")
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
