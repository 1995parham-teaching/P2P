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

	"github.com/pterm/pterm"

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
				pterm.Error.Printf("Failed to download file: %v\n", err)
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
		return err
	}

	// Read file size
	bufferFileSize := make([]byte, config.FileSizeLength)
	if _, err := io.ReadFull(conn, bufferFileSize); err != nil {
		return err
	}

	fileSize, err := strconv.ParseInt(strings.TrimRight(string(bufferFileSize), ":"), 10, 64)
	if err != nil {
		return err
	}

	// Read file name
	bufferFileName := make([]byte, config.FileNameLength)
	if _, err := io.ReadFull(conn, bufferFileName); err != nil {
		return err
	}

	receivedFileName := strings.TrimRight(string(bufferFileName), ":")

	// Create output file with "downloading_" prefix to indicate in-progress download
	outputPath := filepath.Join(c.folder, "downloading_"+filepath.Base(receivedFileName))
	finalPath := filepath.Join(c.folder, filepath.Base(receivedFileName))

	newFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	// Create progress bar
	progressBar, _ := pterm.DefaultProgressbar.
		WithTotal(int(fileSize)).
		WithTitle("Downloading " + receivedFileName).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()

	// Read file content with progress
	if err := c.readFileContentWithProgress(conn, fileSize, newFile, progressBar); err != nil {
		newFile.Close()
		os.Remove(outputPath) // Clean up partial file
		return err
	}

	progressBar.Stop()

	if err := newFile.Close(); err != nil {
		return err
	}

	// Rename to final path after successful download
	if err := os.Rename(outputPath, finalPath); err != nil {
		return err
	}

	pterm.Success.Printf("File saved: %s\n", finalPath)
	return nil
}

func (c *Client) sendRequest(conn io.Writer, fileName string) error {
	msg := (&message.Get{Name: fileName}).Marshal()
	_, err := conn.Write([]byte(msg))
	return err
}

func (c *Client) readFileContentWithProgress(conn io.Reader, fileSize int64, dest io.Writer, progressBar *pterm.ProgressbarPrinter) error {
	buffer := make([]byte, config.BufferSize)
	var totalWritten int64

	for totalWritten < fileSize {
		n, err := conn.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n > 0 {
			written, err := dest.Write(buffer[:n])
			if err != nil {
				return err
			}
			totalWritten += int64(written)
			progressBar.Add(written)
		}

		if err == io.EOF {
			break
		}
	}

	if totalWritten != fileSize {
		pterm.Warning.Printf("Expected %d bytes, got %d\n", fileSize, totalWritten)
	}

	return nil
}
