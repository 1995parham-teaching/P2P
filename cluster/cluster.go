package cluster

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/1995parham-teaching/P2P/internal/utils"
)

type Cluster struct {
	list  []string
	mutex sync.RWMutex
}

func New(list []string) *Cluster {
	// Make a copy of the input list to avoid external mutations
	listCopy := make([]string, len(list))
	copy(listCopy, list)

	return &Cluster{
		list: listCopy,
	}
}

// List returns a copy of the cluster list (thread-safe)
func (c *Cluster) List() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	listCopy := make([]string, len(c.list))
	copy(listCopy, c.list)
	return listCopy
}

// Broadcast sends a message to all nodes in the cluster
func (c *Cluster) Broadcast(conn *net.UDPConn, message string) error {
	list := c.List() // Get thread-safe copy

	var lastErr error
	for _, ip := range list {
		arr := strings.Split(ip, ":")
		if len(arr) != 2 {
			lastErr = fmt.Errorf("invalid address format: %s", ip)
			fmt.Println(lastErr)
			continue
		}

		port, err := strconv.Atoi(arr[1])
		if err != nil {
			lastErr = fmt.Errorf("invalid port in address %s: %w", ip, err)
			fmt.Println(lastErr)
			continue
		}

		addr := net.UDPAddr{
			IP:   net.ParseIP(arr[0]),
			Port: port,
		}

		_, err = conn.WriteToUDP([]byte(message), &addr)
		if err != nil {
			lastErr = fmt.Errorf("failed to send to %s: %w", ip, err)
			fmt.Println(lastErr)
		}
	}

	return lastErr
}

// Merge adds new addresses to the cluster list, excluding duplicates and the host itself
func (c *Cluster) Merge(host string, newList []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, ip := range newList {
		if ip == "" || ip == host {
			continue
		}

		if !utils.Contains(c.list, ip) {
			c.list = append(c.list, ip)
		}
	}
}

// Add adds a single address to the cluster if not already present
func (c *Cluster) Add(addr string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if addr != "" && !utils.Contains(c.list, addr) {
		c.list = append(c.list, addr)
	}
}

// Remove removes an address from the cluster
func (c *Cluster) Remove(addr string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, ip := range c.list {
		if ip == addr {
			c.list = append(c.list[:i], c.list[i+1:]...)
			return
		}
	}
}

// Size returns the number of nodes in the cluster
func (c *Cluster) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.list)
}
