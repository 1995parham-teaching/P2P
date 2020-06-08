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
	seq             int
	fileName        string
	fileSize        int64
	newFile		   *os.File
}

func New(ip string, port int, cluster *cluster.Cluster,
	ticker *time.Ticker, waitingDuration int, folder string) Server {
	return Server{
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
		port := strconv.Itoa(s.Port)
		s.Cluster.Merge(s.IP+":"+port, t.List)
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

		s.seq = 0
		s.AskFile()

	case *message.AskFile:
		go s.StopWait(t.Name)
	case *message.Size:
		if t.Seq == s.seq {
			s.seq += 1
			s.seq %= 2
		}

		s.fileSize = t.Size

	case *message.FileName:
		if t.Seq == s.seq {
			s.seq += 1
			s.seq %= 2
		}

		s.fileName = t.Name

		newFile, err := os.Create(filepath.Join(s.folder, filepath.Base(s.fileName+"getting")))
		if err != nil {
			fmt.Println(err)
		}

		s.newFile = newFile

	case *message.Segment:
		if t.Seq == s.seq {
			s.seq += 1
			s.seq %= 2
		}

		segment := t.Part

		//var receivedBytes int64

		for {
			//if (s.fileSize - receivedBytes) < BUFFERSIZE {
			//	s.newFile.Write(), connection, fileSize-receivedBytes)
			//	if err != nil {
			//		fmt.Println(err)
			//	}
			//
			//	_, err = connection.Read(make([]byte, (receivedBytes+BUFFER)-fileSize))
			//	if err != nil {
			//		fmt.Println(err)
			//	}
			//
			//	break
			//}

			_, err := s.newFile.Write(segment)
			if err != nil {
				fmt.Println(err)
			}

			//receivedBytes += BUFFER
		}
	case *message.Acknowledgment:
		s.SWAck <- t.Seq
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
	_, err := s.conn.WriteToUDP([]byte((&message.AskFile{Name: s.Req}).Marshal()), s.SWAddr)
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

	fmt.Println("Sending filename and filesize!")

	seq := 0

	fileSize := (&message.Size{Size: fileInfo.Size(), Seq: seq}).Marshal()

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

	fileName := (&message.FileName{Name: fileInfo.Name(), Seq: seq}).Marshal()

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

	sendBuffer := make([]byte, BUFFERSIZE-9)

	fmt.Println("Start sending file")

	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}

		buffer := (&message.Segment{
			Part: string(sendBuffer),
			Seq:  seq,
		}).Marshal()

		send := []byte(buffer)

		_, err = s.conn.WriteToUDP(send, s.SWAddr)
		if err != nil {
			fmt.Println(err)
		}

		for {
			ticker := time.NewTicker(6 * time.Second)
			b := false

			select {
			case <-ticker.C:
				_, err = s.conn.WriteToUDP(send, s.SWAddr)
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
