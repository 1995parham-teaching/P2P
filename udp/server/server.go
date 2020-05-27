package server

import (
	"fmt"
	"time"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/request"
	"github.com/elahe-dstn/p2p/response"

	"net"
	"strings"
)

type Server struct {
	IP              string
	Cluster         *cluster.Cluster
	DiscoveryTicker *time.Ticker
}

func New(cluster *cluster.Cluster, ticker *time.Ticker) Server {
	return Server{
		IP:              "127.0.0.1",
		Cluster:         cluster,
		DiscoveryTicker: ticker,
	}
}

func (s *Server) Up() {
	addr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: 1378,
	}

	add, err := net.ResolveUDPAddr("udp", addr.String())
	print(add)

	ser, err := net.ListenUDP("udp", &addr)

	if err != nil {
		fmt.Println(err)
		return
	}

	message := make([]byte, 2048)

	for {
		_, remoteAddr, err := ser.ReadFromUDP(message)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := strings.Split(string(message), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		req := request.Unmarshal(r)

		s.protocol(req, ser, remoteAddr)
	}
}

func (s *Server) protocol(req request.Request, ser *net.UDPConn, remoteAddr *net.UDPAddr) {
	switch req.(type) {
	case request.Discover:
		go transfer(ser, remoteAddr, (&response.Discover{List: s.Cluster.List}).Marshal())
		//case request.File:
		//	f := req.(request.File)
		//	if n.Search(f.Name) {
		//		go transfer(ser, remoteAddr, response.File{Answer: true, TcpPort: n.TcpPort}.Marshal())
		//	}
	}
	// if t == "list" {
	//	n.merge(protocol[1:])
	//}
	//} else if t == "ans" {
	//	if n.waiting {
	//		n.check(protocol[1:])
	//	}
	//}
}

func transfer(conn *net.UDPConn, addr *net.UDPAddr, message string) {
	_, err := conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Server) Discover() {
	for {
		<-s.DiscoveryTicker.C
		s.broadcast()
	}
}
