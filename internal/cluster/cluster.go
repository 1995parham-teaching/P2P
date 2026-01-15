package cluster

import (
	"fmt"
	"net"
	"sync"

	"github.com/pterm/pterm"
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
	for _, address := range list {
		// Use ResolveUDPAddr to handle both IP addresses and hostnames
		addr, err := net.ResolveUDPAddr("udp", address)
		if err != nil {
			lastErr = fmt.Errorf("failed to resolve address %s: %w", address, err)
			pterm.Error.Printf("Failed to resolve %s: %v\n", address, err)
			continue
		}

		_, err = conn.WriteToUDP([]byte(message), addr)
		if err != nil {
			lastErr = fmt.Errorf("failed to send to %s: %w", address, err)
			pterm.Error.Printf("Failed to send to %s: %v\n", address, err)
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

		if !contains(c.list, ip) {
			c.list = append(c.list, ip)
			pterm.Success.Printf("Discovered new peer: %s\n", ip)
		}
	}
}

// Add adds a single address to the cluster if not already present
func (c *Cluster) Add(addr string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if addr != "" && !contains(c.list, addr) {
		c.list = append(c.list, addr)
		pterm.Success.Printf("Discovered new peer: %s\n", addr)
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

// contains checks if a string slice contains a specific item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
