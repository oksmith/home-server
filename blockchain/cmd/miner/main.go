package main

import (
	"fmt"
	"log"

	"github.com/oksmith/home-server/blockchain/pkg/chain"
)

func main() {
	// Create a new blockchain with difficulty 4
	// (requires 4 leading zeros in block hashes)
	fmt.Println("Creating blockchain...")
	bc := chain.New(4) // note: difficulty of 6 takes a good 30 seconds per block

	// Add some blocks
	blocks := []string{
		"First transaction: Alice sends 10 coins to Bob",
		"Second transaction: Bob sends 5 coins to Charlie",
		"Third transaction: Charlie sends 3 coins to Alice",
	}

	for i, data := range blocks {
		fmt.Printf("\nMining block %d...\n", i+1)
		if err := bc.AddBlock(data); err != nil {
			log.Fatalf("Failed to add block: %v", err)
		}
	}

	// Display the blockchain
	fmt.Println("\n=== BLOCKCHAIN ===")
	for _, block := range bc.Blocks {
		fmt.Printf("\nBlock #%d\n", block.Index)
		fmt.Printf("Timestamp: %s\n", block.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
	}

	// Validate the chain
	fmt.Printf("\nBlockchain valid? %v\n", bc.IsValid())

	// Save to file
	filename := "blockchain.json"
	if err := bc.SaveToFile(filename); err != nil {
		log.Fatalf("Failed to save blockchain: %v", err)
	}
	fmt.Printf("\nBlockchain saved to %s\n", filename)

	// Try tampering with the chain
	fmt.Println("\n=== TAMPERING TEST ===")
	// bc.Blocks[1].Data = "TAMPERED DATA"
	// bc.Blocks[1].Data = "First transaction: Alice sends 10 coins to Bob"
	bc.Blocks[2].Data = "TAMPERED DATA"
	fmt.Printf("After tampering, blockchain valid? %v\n", bc.IsValid())
}
