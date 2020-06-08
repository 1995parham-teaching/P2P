package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"net"
	"strings"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/message"
)

const BUFFERSIZE = 1024

type Server struct {
	IP              string
	Port            int
	Cluster         *cluster.Cluster
	DiscoveryTicker *time.Ticker
	waiting         bool
	WaitingDuration int
	waitingTicker   *time.Ticker
	Req             string
	folder          string
	conn            *net.UDPConn
	prior           []string
	SWAddr          *net.UDPAddr
	SWAck           chan int
}

func New(ip string, port int, cluster *cluster.Cluster,
	ticker *time.Ticker, waitingDuration int, folder string) Server {
	return Server {
		IP:              ip,
		Port:            port,
		Cluster:         cluster,
		DiscoveryTicker: ticker,
		WaitingDuration: waitingDuration,
		folder:          folder,
		prior:           make([]string, 0),
		SWAck:           make(chan int),
	}
}

func (s *Server) Up(tcpPort chan int, request chan string, fName chan string) {
	tPort := <-tcpPort
	addr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: s.Port,
	}

	_, err := net.ResolveUDPAddr("udp", addr.String())
	if err != nil {
		fmt.Println(err)
	}

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

		s.protocol(res, remoteAddr, tPort, request, fName)
	}
}

func (s *Server) protocol(res message.Message, remoteAddr *net.UDPAddr, tcpPort int, request chan string, fName chan string) {
	fmt.Println("protocol")

	switch t := res.(type) {
	case *message.Discover:
		port:= strconv.Itoa(s.Port)
		s.Cluster.Merge(s.IP + ":" + port, t.List)
	case *message.Get:
		if s.Search(t.Name) {
			//go s.transfer(remoteAddr, (&message.File{TCPPort: tcpPort}).Marshal())
			go s.transfer(remoteAddr, (&message.StopWait{}).Marshal())
		}
	case *message.File:
		if s.waiting {
			// Add to prior list
			exists := false

			for _, ip := range s.prior {
				if ip == remoteAddr.String() {
					exists = true
					break
				}
			}

			if !exists {
				s.prior = append(s.prior, remoteAddr.String())
			}

			ip := remoteAddr.IP.String()
			s.waiting = false
			request <- fmt.Sprintf("%s:%d", ip, t.TCPPort)
			fName <- s.Req
		}
	case *message.StopWait:
		if s.waiting {
			// Add to prior list
			exists := false

			for _, ip := range s.prior {
				if ip == remoteAddr.String() {
					exists = true
					break
				}
			}

			if !exists {
				s.prior = append(s.prior, remoteAddr.String())
			}

			s.SWAddr = remoteAddr
			s.waiting = false
		}

		s.AskFile()

	case *message.AskFile:
		go s.StopWait(t.Name)
	}
}

func (s *Server) transfer(addr *net.UDPAddr, message string) {
	exists := false

	for _, ip := range s.prior {
		if ip == addr.String() {
			exists = true
			break
		}
	}

	if !exists {
		time.Sleep(10 * time.Second)
	}

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
	s.waitingTicker = time.NewTicker(time.Duration(s.WaitingDuration) * time.Second)
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

func (s *Server) AskFile() {
	_, err := s.conn.WriteToUDP([]byte((&message.AskFile{Name:s.Req}).Marshal()), s.SWAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Server) StopWait(name string) {
	fmt.Println("A stop and wait client has connected!")

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

	seq := 0

	_, err = s.conn.WriteToUDP([]byte(fileSize), s.SWAddr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		ticker := time.NewTicker(6 * time.Second)
		b := false

		select {
		case <-ticker.C:
			_, err = s.conn.WriteToUDP([]byte(fileSize), s.SWAddr)
			if err != nil {
				fmt.Println(err)
			}
		case ack := <-s.SWAck:
			if ack == seq {
				seq += 1
				seq %= 2
				b = true
				break
			}
		}

		if b {
			break
		}
	}

	_, err = s.conn.WriteToUDP([]byte(fileName), s.SWAddr)
	if err != nil {
		fmt.Println(err)
	}

	for {
		ticker := time.NewTicker(6 * time.Second)
		b := false

		select {
		case <-ticker.C:
			_, err = s.conn.WriteToUDP([]byte(fileName), s.SWAddr)
			if err != nil {
				fmt.Println(err)
			}
		case ack := <-s.SWAck:
			if ack == seq {
				seq += 1
				seq %= 2
				b = true
				break
			}
		}

		if b {
			break
		}
	}

	sendBuffer := make([]byte, BUFFERSIZE)

	fmt.Println("Start sending file")

	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}

		_, err = s.conn.WriteToUDP(sendBuffer, s.SWAddr)
		if err != nil {
			fmt.Println(err)
		}

		for {
			ticker := time.NewTicker(6 * time.Second)
			b := false

			select {
			case <-ticker.C:
				_, err = s.conn.WriteToUDP(sendBuffer, s.SWAddr)
				if err != nil {
					fmt.Println(err)
				}
			case ack := <-s.SWAck:
				if ack == seq {
					seq += 1
					seq %= 2
					b = true
					break
				}
			}

			if b {
				break
			}
		}
	}

	fmt.Println("File has been sent, closing connection!")
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString += ":"
			continue
		}

		break
	}

	return retunString
}