package client

import (
	"bufio"
	"os"
	"strings"

	"github.com/elahe-dstn/p2p/node"
	"github.com/spf13/cobra"
)

func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "node",
			Short: "",
			Long:  "",
			Run: func(cmd *cobra.Command, args []string) {
				// Ask user to give its cluster members
				print("Enter your cluster members list(enter q for quit)")
				reader := bufio.NewReader(os.Stdin)

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

				n := node.New("0", cluster)
				n.Run()
			},
		},
	)
}
