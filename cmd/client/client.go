package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/1995parham-teaching/P2P/node"
	"github.com/spf13/cobra"
)

func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "node",
			Short: "Start a P2P node for file sharing",
			Run: func(cmd *cobra.Command, args []string) {
				reader := bufio.NewReader(os.Stdin)

				folder, err := getFolder(reader)
				if err != nil {
					fmt.Printf("Error getting folder: %v\n", err)
					return
				}

				clusterList, err := getClusterMembers(reader)
				if err != nil {
					fmt.Printf("Error getting cluster members: %v\n", err)
					return
				}

				n, err := node.New(folder, clusterList)
				if err != nil {
					fmt.Printf("Failed to create node: %v\n", err)
					return
				}

				if err := n.Run(); err != nil {
					fmt.Printf("Node error: %v\n", err)
				}
			},
		},
	)
}

func getFolder(reader *bufio.Reader) (string, error) {
	for {
		fmt.Println("Enter the folder you want to share:")

		folder, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		folder = strings.TrimSpace(folder)

		// Check if the folder exists
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

		// Basic validation of address format
		if !strings.Contains(text, ":") {
			fmt.Println("Invalid address format. Use IP:Port (e.g., 127.0.0.1:1378)")
			continue
		}

		cluster = append(cluster, text)
		fmt.Printf("Added: %s\n", text)
	}

	return cluster, nil
}
