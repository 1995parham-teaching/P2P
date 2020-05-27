package cluster

import (
	"fmt"
	"net"
	"sync"
)

type Cluster struct {
	List  []string
	Mutex *sync.Mutex
}

func New(list []string) Cluster {
	return Cluster{
		List:  list,
		Mutex: &sync.Mutex{},
	}
}

func (c *Cluster) broadcast(t string) {
	for i, ip := range c.List {
		conn, err := net.Dial("udp", ip)
		if err != nil {
			c.Mutex.Lock()

			c.List[i] = c.List[len(c.List)-1] // Copy last element to index i.
			c.List[len(c.List)-1] = ""        // Erase last element (write zero value).
			c.List = c.List[:len(c.List)-1]   // Truncate slice.

			c.Mutex.Unlock()
			return
		}

		_, err = conn.Write([]byte(t))

		if err != nil {
			fmt.Println(err)
		}
	}
}

func (c *Cluster) Merge(list []string) {
	for _, ip := range list {
		exists := false
		for _, c := range c.List {
			if ip == c {
				exists = true
			}
		}

		if !exists {
			c.Mutex.Lock()
			c.List = append(c.List, ip)
			c.Mutex.Unlock()
		}
	}
}
