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
	return Client{folder:folder}
}

func (c *Client) Connect(addr chan string, fName chan string) {
	conn, err := net.Dial("tcp", <-addr)
	fmt.Println("rad")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte((&message.Get{Name:<-fName}).Marshal()))
	if err != nil {
		fmt.Println(err)
		return
	}


	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	conn.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	conn.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	newFile, err := os.Create(filepath.Join(c.folder, filepath.Base(fileName + "getting")))

	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < BUFFERSIZE {
			io.CopyN(newFile, conn, fileSize - receivedBytes)
			conn.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(newFile, conn, BUFFERSIZE)
		receivedBytes += BUFFERSIZE
	}
	fmt.Println("Received file completely!")
}
