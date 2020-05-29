package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/elahe-dstn/p2p/message"
)

type Client struct {
	folder string
}

const BUFFER = 1024

func New(folder string) Client {
	return Client{folder: folder}
}

func (c *Client) Connect(addr chan string, fName chan string) {
	for {
		connection, err := net.Dial("tcp", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		request(connection, fName)

		bufferFileName := make([]byte, 64)
		bufferFileSize := make([]byte, 10)

		_, err = connection.Read(bufferFileSize)
		if err != nil {
			fmt.Println(err)
		}

		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		_, err = connection.Read(bufferFileName)
		if err != nil {
			fmt.Println(err)
		}

		fileName := strings.Trim(string(bufferFileName), ":")

		newFile, err := os.Create(filepath.Join(c.folder, filepath.Base(fileName+"getting")))

		read(connection, fileSize, newFile)

		if err != nil {
			panic(err)
		}

		fmt.Println("Received file completely!")
	}
}

func request(conn io.Writer, fName chan string) {
	_, err := conn.Write([]byte((&message.Get{Name: <-fName}).Marshal()))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func read(connection io.Reader, fileSize int64, newFile io.Writer) {
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFER {
			_, err := io.CopyN(newFile, connection, fileSize-receivedBytes)
			if err != nil {
				fmt.Println(err)
			}

			_, err = connection.Read(make([]byte, (receivedBytes+BUFFER)-fileSize))
			if err != nil {
				fmt.Println(err)
			}

			break
		}

		_, err := io.CopyN(newFile, connection, BUFFER)
		if err != nil {
			fmt.Println(err)
		}

		receivedBytes += BUFFER
	}
}
