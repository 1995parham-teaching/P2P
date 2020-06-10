package client

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/elahe-dstn/p2p/node"
	"github.com/spf13/cobra"
)

func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "node",
			Short: "",
			Run: func(cmd *cobra.Command, args []string) {
				reader := bufio.NewReader(os.Stdin)

				folder := ""

				for {
					fmt.Println("Enter the folder you want to share")

					folder, err := reader.ReadString('\n')

					if err != nil {
						print(err)
						return
					}


					folder = strings.TrimSuffix(folder, "\n")

					// Check if the folder exists
					_, err = os.Open(folder)
					if err == nil {
						break
					}

					fmt.Println("Couldn't find the folder")
				}

				// Ask user to give its cluster members
				fmt.Println("Enter your cluster members list(enter q for quit)")

				cluster := make([]string, 0)

				for {
					text, err := reader.ReadString('\n')

					if err != nil {
						print(err)
						return
					}

					text = strings.TrimSuffix(text, "\n")

					if text == "q" {
						break
					}

					cluster = append(cluster, text)
				}

				fmt.Println("Choose the approach for transferring file\n" +
					"1-TCP\n" +
					"2-Stop and wait")

				approach, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
				}

				way, err := strconv.Atoi(approach)
				if err != nil {
					fmt.Println(err)
				}

				n := node.New(folder, cluster, way)
				n.Run()
			},
		},
	)
}
