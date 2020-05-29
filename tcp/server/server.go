package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/elahe-dstn/p2p/message"
)

const BUFFERSIZE = 1024

type Server struct {
	TcpPort	int
	folder  string
}

func New(folder string) Server {
	return Server{folder:folder}
}

func (s *Server) Up(TcpPort chan int)  {
	addr := net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: 0,
	}

	l, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	s.TcpPort = l.Addr().(*net.TCPAddr).Port

	TcpPort<- s.TcpPort

	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	m := make([]byte, 2048)

	_, err = c.Read(m)

	fmt.Println(string(m))

	if err != nil {
		fmt.Println(err)
		return
	}

	res := message.Unmarshal(string(m))
	g := res.(*message.Get)

	go s.send(c, g.Name)

}

func (s *Server) send(conn net.Conn, name string)  {
	fmt.Println("A client has connected!")
	defer conn.Close()
	file, err := os.Open(s.folder + "/" + name)
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Println("Sending filename and filesize!")
	conn.Write([]byte(fileSize))
	conn.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		conn.Write(sendBuffer)
	}
	fmt.Println("File has been sent, closing connection!")
	return
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}