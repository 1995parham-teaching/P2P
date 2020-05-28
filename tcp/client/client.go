package client

import (
	"fmt"
	"net"

	"github.com/elahe-dstn/p2p/message"
)

type Client struct {

}

func (c *Client) Connect(addr chan string, fName chan string) {
		conn, err := net.Dial("tcp", <-addr)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = conn.Write([]byte((&message.Get{Name:<-fName}).Marshal()))
		if err != nil {
			fmt.Println(err)
			return
		}
}
