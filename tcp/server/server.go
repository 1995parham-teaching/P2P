package server

import (
	"fmt"
	"net"
)

type Server struct {
	TcpPort	int
}

func New() Server {
	return Server{}
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

	message := make([]byte, 2048)

	_, err = c.Read(message)

	fmt.Println(message)

	if err != nil {
		fmt.Println(err)
		return
	}

	//go send(c)

}
