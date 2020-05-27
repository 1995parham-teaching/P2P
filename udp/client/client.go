package client

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/elahe-dstn/p2p/request"
)

type Client struct {
	DiscoveryTicker *time.Ticker
	waiting         bool
	waitingTicker   *time.Ticker
	Cluster         []string
	Mutex           *sync.Mutex
	Req             string
}

func New(ticker *time.Ticker, cluster []string) Client {
	return Client{
		DiscoveryTicker: ticker,
		Cluster:         cluster,
		Mutex:           &sync.Mutex{},
	}
}

func (c *Client) Discover() {
	for {
		<-c.DiscoveryTicker.C
		c.broadcast(request.Discover{}.Marshal())
	}
}

func (c *Client) broadcast(t string) {
	for i, ip := range c.Cluster {
		conn, err := net.Dial("udp", ip)
		if err != nil {
			c.Mutex.Lock()

			c.Cluster[i] = c.Cluster[len(c.Cluster)-1] // Copy last element to index i.
			c.Cluster[len(c.Cluster)-1] = ""           // Erase last element (write zero value).
			c.Cluster = c.Cluster[:len(c.Cluster)-1]   // Truncate slice.

			c.Mutex.Unlock()
			return
		}

		_, err = conn.Write([]byte(t))

		if err != nil {
			fmt.Println(err)
		}

		go read(conn)
	}
}

func read(conn net.Conn) {
	m := make([]byte, 1024)

	for {
		conn.Read(m)

		print(string(m))
	}
}

func (c *Client) File() {
	c.waiting = true
	c.waitingTicker = time.NewTicker(5 * time.Second)
	c.broadcast(request.File{Name: c.Req}.Marshal())
	<-c.waitingTicker.C
	c.waiting = false
	c.waitingTicker.Stop()
}

//func (n *Node) check(ans []string) {
//	if ans[0] == "y" {
//		n.waiting = false
//		n.waitingTicker.Stop()
//		n.answeringNode = ans[1]
//	}
//
//	connect()
//}
