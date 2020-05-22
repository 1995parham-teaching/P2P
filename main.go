package main

import (
	"bufio"
	"os"
	"strings"
)

func main() {
	// Ask user to give its cluster members
	print("Enter your cluster members list(enter q for quit)")
	reader := bufio.NewReader(os.Stdin)

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
	}
}
