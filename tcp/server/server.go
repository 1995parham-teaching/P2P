package server

import (
	"fmt"
	"net"

	"github.com/elahe-dstn/p2p/node"
)

func Server(n *node.Node)  {
	addr := net.TCPAddr{
		IP: net.ParseIP(n.IP),
		Port: 0,
	}

	l, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	n.TcpPort = l.Addr().(*net.TCPAddr).Port
	defer l.Close()

	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	message := make([]byte, 2048)

	_, err = c.Read(message)
	if err != nil {
		fmt.Println(err)
		return
	}

	go send(c)

}
