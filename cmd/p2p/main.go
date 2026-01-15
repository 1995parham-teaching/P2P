package main

import (
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"

	"github.com/1995parham-teaching/P2P/internal/node"
)

func main() {
	// Print header
	_ = pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("P2P", pterm.NewStyle(pterm.FgCyan)),
	).Render()

	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).
		WithTextStyle(pterm.NewStyle(pterm.FgLightCyan)).
		Println("Peer-to-Peer File Sharing")

	pterm.Println()

	// Check for environment variables (for Docker/automated deployment)
	folder := os.Getenv("P2P_FOLDER")
	clusterEnv := os.Getenv("P2P_CLUSTER")

	var clusterList []string

	if folder != "" && clusterEnv != "" {
		// Use environment variables
		pterm.Info.Println("Using environment configuration")

		_ = pterm.DefaultBulletList.WithItems([]pterm.BulletListItem{
			{Level: 0, Text: "Folder: " + pterm.LightCyan(folder)},
			{Level: 0, Text: "Cluster: " + pterm.LightCyan(clusterEnv)},
		}).Render()

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
		var err error

		folder, err = getFolder()
		if err != nil {
			pterm.Error.Printf("Error getting folder: %v\n", err)
			os.Exit(1)
		}

		clusterList, err = getClusterMembers()
		if err != nil {
			pterm.Error.Printf("Error getting cluster members: %v\n", err)
			os.Exit(1)
		}
	}

	pterm.Println()
	pterm.Success.Println("Configuration complete!")
	pterm.Println()

	n, err := node.New(folder, clusterList)
	if err != nil {
		pterm.Error.Printf("Failed to create node: %v\n", err)
		os.Exit(1)
	}

	if err := n.Run(); err != nil {
		pterm.Error.Printf("Node error: %v\n", err)
		os.Exit(1)
	}
}

func getFolder() (string, error) {
	for {
		folder, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("./shared").
			Show("Enter the folder you want to share")
		if err != nil {
			return "", err
		}

		folder = strings.TrimSpace(folder)
		if folder == "" {
			folder = "./shared"
		}

		info, err := os.Stat(folder)
		if err != nil {
			pterm.Warning.Println("Couldn't find the folder, please try again")
			continue
		}

		if !info.IsDir() {
			pterm.Warning.Println("Path is not a directory, please try again")
			continue
		}

		pterm.Success.Printf("Using folder: %s\n", folder)
		return folder, nil
	}
}

func getClusterMembers() ([]string, error) {
	var cluster []string

	pterm.Info.Println("Add cluster members (IP:Port format)")
	pterm.DefaultBasicText.WithStyle(pterm.NewStyle(pterm.FgGray)).
		Println("Press Enter with empty input when done")
	pterm.Println()

	for {
		text, err := pterm.DefaultInteractiveTextInput.
			WithDefaultText("").
			Show("Add peer address (or press Enter to finish)")
		if err != nil {
			return nil, err
		}

		text = strings.TrimSpace(text)

		if text == "" {
			break
		}

		if !strings.Contains(text, ":") {
			pterm.Warning.Println("Invalid format. Use IP:Port (e.g., 127.0.0.1:1378)")
			continue
		}

		cluster = append(cluster, text)
		pterm.Success.Printf("Added: %s\n", text)
	}

	if len(cluster) > 0 {
		pterm.Println()
		pterm.Info.Printf("Added %d cluster member(s)\n", len(cluster))
	} else {
		pterm.Info.Println("No cluster members added (standalone mode)")
	}

	return cluster, nil
}
