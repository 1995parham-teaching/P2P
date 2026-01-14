package client

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

type Client struct {
	folder string
}

func New(folder string) *Client {
	return &Client{folder: folder}
}

// Connect listens for file download requests and processes them
func (c *Client) Connect(ctx context.Context, addr <-chan string, fileName <-chan string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case serverAddr := <-addr:
			fName := <-fileName
			if err := c.downloadFile(serverAddr, fName); err != nil {
				fmt.Printf("Failed to download file: %v\n", err)
			}
		}
	}
}

func (c *Client) downloadFile(serverAddr, fileName string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", serverAddr, err)
	}
	defer conn.Close()

	// Send the file request
	if err := c.sendRequest(conn, fileName); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read file size
	bufferFileSize := make([]byte, config.FileSizeLength)
	if _, err := io.ReadFull(conn, bufferFileSize); err != nil {
		return fmt.Errorf("failed to read file size: %w", err)
	}

	fileSize, err := strconv.ParseInt(strings.TrimRight(string(bufferFileSize), ":"), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid file size: %w", err)
	}

	// Read file name
	bufferFileName := make([]byte, config.FileNameLength)
	if _, err := io.ReadFull(conn, bufferFileName); err != nil {
		return fmt.Errorf("failed to read file name: %w", err)
	}

	receivedFileName := strings.TrimRight(string(bufferFileName), ":")

	// Create output file with "downloading_" prefix to indicate in-progress download
	outputPath := filepath.Join(c.folder, "downloading_"+filepath.Base(receivedFileName))
	finalPath := filepath.Join(c.folder, filepath.Base(receivedFileName))

	newFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}

	// Read file content
	if err := c.readFileContent(conn, fileSize, newFile); err != nil {
		newFile.Close()
		os.Remove(outputPath) // Clean up partial file
		return fmt.Errorf("failed to read file content: %w", err)
	}

	if err := newFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Rename to final path after successful download
	if err := os.Rename(outputPath, finalPath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	fmt.Println("Received file completely!")
	return nil
}

func (c *Client) sendRequest(conn io.Writer, fileName string) error {
	msg := (&message.Get{Name: fileName}).Marshal()
	_, err := conn.Write([]byte(msg))
	return err
}

func (c *Client) readFileContent(conn io.Reader, fileSize int64, dest io.Writer) error {
	written, err := io.CopyN(dest, conn, fileSize)
	if err != nil {
		return fmt.Errorf("copy failed after %d bytes: %w", written, err)
	}

	if written != fileSize {
		return fmt.Errorf("expected %d bytes, got %d", fileSize, written)
	}

	return nil
}
