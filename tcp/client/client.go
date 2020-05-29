package client

import (
	"fmt"
	"net"

	"github.com/elahe-dstn/p2p/message"
)

type Client struct {

}

func (c *Client) Connect(test chan string, addr chan string, fName chan string) {
	fmt.Println(<-test)
	conn, err := net.Dial("tcp", <-addr)
	fmt.Println("rad")
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
