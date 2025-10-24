package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/oksmith/home-server/blockchain/pkg/node"
)

func main() {
	// Command line flags
	port := flag.Int("port", 8080, "Port to run the node on")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses (e.g., localhost:8081,localhost:8082)")
	difficulty := flag.Int("difficulty", 3, "Mining difficulty")
	reward := flag.Float64("reward", 50.0, "Mining reward")
	flag.Parse()

	address := fmt.Sprintf("localhost:%d", *port)

	// Create node
	n, err := node.New(address, *difficulty, *reward)
	if err != nil {
		log.Fatal(err)
	}

	// Add peers
	if *peers != "" {
		peerList := strings.Split(*peers, ",")
		for _, peer := range peerList {
			peer = strings.TrimSpace(peer)
			if peer != "" {
				n.AddPeer(peer)
			}
		}
	}

	// Sync with peers on startup
	if len(n.GetPeers()) > 0 {
		fmt.Printf("[%s] Syncing with peers...\n", address)
		if err := n.SyncWithPeers(); err != nil {
			fmt.Printf("[%s] Sync warning: %v\n", address, err)
		}
	}

	fmt.Printf("\n=== NODE INFO ===\n")
	fmt.Printf("Address: %s\n", address)
	fmt.Printf("Wallet Address: %s\n", n.Wallet.Address())
	fmt.Printf("Chain Length: %d blocks\n", n.Chain.Length())
	fmt.Printf("Balance: %.2f coins\n", n.Chain.GetBalance(n.Wallet.Address()))
	fmt.Printf("Peers: %v\n\n", n.GetPeers())

	// Start server
	log.Fatal(n.StartServer())
}
