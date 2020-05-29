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

const BUFFERSIZE = 1024

func New(folder string) Client {
	return Client{folder: folder}
}

func (c *Client) Connect(addr chan string, fName chan string) {
	for {
		conn, err := net.Dial("tcp", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = conn.Write([]byte((&message.Get{Name: <-fName}).Marshal()))
		if err != nil {
			fmt.Println(err)
			return
		}

		bufferFileName := make([]byte, 64)
		bufferFileSize := make([]byte, 10)

		_, err = conn.Read(bufferFileSize)
		if err != nil {
			fmt.Println(err)
		}

		fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

		_, err = conn.Read(bufferFileName)
		if err != nil {
			fmt.Println(err)
		}

		fileName := strings.Trim(string(bufferFileName), ":")

		newFile, err := os.Create(filepath.Join(c.folder, filepath.Base(fileName+"getting")))

		if err != nil {
			panic(err)
		}

		var receivedBytes int64

		for {
			if (fileSize - receivedBytes) < BUFFERSIZE {
				_, err = io.CopyN(newFile, conn, fileSize-receivedBytes)
				if err != nil {
					fmt.Println(err)
				}

				_, err = conn.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
				if err != nil {
					fmt.Println(err)
				}

				break
			}
			
			_, err = io.CopyN(newFile, conn, BUFFERSIZE)
			if err != nil {
				fmt.Println(err)
			}
			receivedBytes += BUFFERSIZE
		}
		fmt.Println("Received file completely!")
	}
}
