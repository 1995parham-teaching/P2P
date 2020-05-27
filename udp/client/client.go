package client

import (
	"fmt"
	"net"
	"time"

	"github.com/elahe-dstn/p2p/node"
	"github.com/elahe-dstn/p2p/request"
)

type Client struct {
	DiscoveryTicker *time.Ticker
	waiting       bool
	waitingTicker *time.Ticker
}

func New(ticker *time.Ticker) Client {
	return Client{
		DiscoveryTicker: ticker,
	}
}

func (c *Client) Discover(n *node.Node)  {
	for {
		<-c.DiscoveryTicker.C
		broadcast(n, request.Discover{}.Marshal())
	}
}

func broadcast(n *node.Node, t string) {
	for i, ip := range n.Cluster {
		conn, err := net.Dial("udp", ip)
		if err != nil {
			n.Mutex.Lock()

			n.Cluster[i] = n.Cluster[len(n.Cluster)-1] // Copy last element to index i.
			n.Cluster[len(n.Cluster)-1] = ""           // Erase last element (write zero value).
			n.Cluster = n.Cluster[:len(n.Cluster)-1]   // Truncate slice.

			n.Mutex.Unlock()
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

func (c *Client) File(n *node.Node) {
	c.waiting = true
	c.waitingTicker = time.NewTicker(5 * time.Second)
	broadcast(n, request.File{Name: n.Req}.Marshal())
	<-c.waitingTicker.C
	c.waiting = false
	c.waitingTicker.Stop()
}

func (n *Node) check(ans []string) {
	if ans[0] == "y" {
		n.waiting = false
		n.waitingTicker.Stop()
		n.answeringNode = ans[1]
	}

	connect()
}