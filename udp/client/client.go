package client

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/elahe-dstn/p2p/cluster"
	"github.com/elahe-dstn/p2p/request"
	"github.com/elahe-dstn/p2p/response"
)

type Client struct {
	Cluster       *cluster.Cluster
	Req           string
	waiting       bool
	waitingTicker *time.Ticker
}

func New(cluster *cluster.Cluster) Client {
	return Client{
		Cluster: cluster,
	}
}

func (c *Client) read(conn net.Conn) {
	m := make([]byte, 1024)

	for {
		_, err := conn.Read(m)
		if err != nil {
			fmt.Println(err)
		}

		res := strings.TrimSuffix(string(m), "\n")

		r := response.Unmarshal(res)

		switch r.(type) {
		case *response.Discover:
			clu := r.(*response.Discover)
			c.Cluster.Merge(clu.List)
		}

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
