package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/1995parham-teaching/P2P/internal/node"
)

func main() {
	// Check for environment variables (for Docker/automated deployment)
	folder := os.Getenv("P2P_FOLDER")
	clusterEnv := os.Getenv("P2P_CLUSTER")

	var clusterList []string
	var err error

	if folder != "" && clusterEnv != "" {
		// Use environment variables
		fmt.Printf("Using environment configuration:\n")
		fmt.Printf("  Folder: %s\n", folder)
		fmt.Printf("  Cluster: %s\n", clusterEnv)

		// Parse cluster list from comma-separated string
		if clusterEnv != "" {
			for _, addr := range strings.Split(clusterEnv, ",") {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					clusterList = append(clusterList, addr)
				}
			}
		}
	} else {
		// Interactive mode
		reader := bufio.NewReader(os.Stdin)

		folder, err = getFolder(reader)
		if err != nil {
			fmt.Printf("Error getting folder: %v\n", err)
			os.Exit(1)
		}

		clusterList, err = getClusterMembers(reader)
		if err != nil {
			fmt.Printf("Error getting cluster members: %v\n", err)
			os.Exit(1)
		}
	}

	n, err := node.New(folder, clusterList)
	if err != nil {
		fmt.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}

	if err := n.Run(); err != nil {
		fmt.Printf("Node error: %v\n", err)
		os.Exit(1)
	}
}

func getFolder(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println("Enter the folder you want to share:")

		folder, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		folder = strings.TrimSpace(folder)

		info, err := os.Stat(folder)
		if err != nil {
			fmt.Println("Couldn't find the folder, please try again")
			continue
		}

		if !info.IsDir() {
			fmt.Println("Path is not a directory, please try again")
			continue
		}

		return folder, nil
	}
}

func getClusterMembers(reader *bufio.Reader) ([]string, error) {
	fmt.Println("Enter your cluster members list (one per line, enter 'q' to finish):")

	var cluster []string

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		text = strings.TrimSpace(text)

		if text == "q" || text == "Q" {
			break
		}

		if text == "" {
			continue
		}

		if !strings.Contains(text, ":") {
			fmt.Println("Invalid address format. Use IP:Port (e.g., 127.0.0.1:1378)")
			continue
		}

		cluster = append(cluster, text)
		fmt.Printf("Added: %s\n", text)
	}

	return cluster, nil
}
