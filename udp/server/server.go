package server

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"net"
	"strings"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/message"
)

type Server struct {
	IP              string
	Cluster         *cluster.Cluster
	DiscoveryTicker *time.Ticker
	waiting         bool
	waitingTicker   *time.Ticker
	Req             string
	folder          string
	conn            *net.UDPConn
}

func New(cluster *cluster.Cluster, ticker *time.Ticker, folder string) Server {
	return Server{
		IP:              "127.0.0.1",
		Cluster:         cluster,
		DiscoveryTicker: ticker,
		folder:          folder,
	}
}

func (s *Server) Up(tcpPort chan int, request chan string) {
	tPort := <-tcpPort
	addr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: 1378,
	}

	add, err := net.ResolveUDPAddr("udp", addr.String())
	print(add)

	ser, err := net.ListenUDP("udp", &addr)

	s.conn = ser

	if err != nil {
		fmt.Println(err)
		return
	}

	m := make([]byte, 2048)

	for {
		_, remoteAddr, err := ser.ReadFromUDP(m)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := strings.Split(string(m), "\n")[0]

		r = strings.TrimSuffix(r, "\n")

		fmt.Println(r)

		res := message.Unmarshal(r)

		s.protocol(res, remoteAddr, tPort, request)
	}
}

func (s *Server) protocol(res message.Message, remoteAddr *net.UDPAddr, tcpPort int, request chan string) {
	fmt.Println("protocol")
	switch t := res.(type) {
	case *message.Discover:
		s.Cluster.Merge(t.List)
	case *message.Get:
		if s.Search(t.Name) {
			go s.transfer(remoteAddr, (&message.File{TcpPort: tcpPort}).Marshal())
		}
	case *message.File:
		ip := remoteAddr.IP.String()
		s.waiting = false
		request <- fmt.Sprintf("%s:%d", ip, t.TcpPort)
	}

	//case request.File:
	//	f := req.(request.File)
	//	if n.Search(f.Name) {
	//		go transfer(ser, remoteAddr, response.File{Answer: true, TcpPort: n.TcpPort}.Marshal())
	//	}
	// if t == "list" {
	//	n.merge(protocol[1:])
	//}
	//} else if t == "ans" {
	//	if n.waiting {
	//		n.check(protocol[1:])
	//	}
	//}
}

func (s *Server) transfer(addr *net.UDPAddr, message string) {
	_, err := s.conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Server) Discover() {
	for {
		<-s.DiscoveryTicker.C
		s.Cluster.Broadcast(s.conn, (&message.Discover{List: s.Cluster.List}).Marshal())
	}
}

func (s *Server) File() {
	s.waiting = true
	s.waitingTicker = time.NewTicker(5 * time.Second)
	s.Cluster.Broadcast(s.conn, (&message.Get{Name: s.Req}).Marshal())
	<-s.waitingTicker.C
	s.waiting = false
	s.waitingTicker.Stop()
}

func (s *Server) Search(file string) bool {
	found := false

	err := filepath.Walk(s.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if file == info.Name() {
			found = true
			return nil
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return found
}

func Connect(port int) {

}
