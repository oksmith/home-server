package chain

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/oksmith/home-server/blockchain/pkg/block"
)

// Chain represents the blockchain
type Chain struct {
	Blocks     []*block.Block `json:"blocks"`
	Difficulty int            `json:"difficulty"`
}

// New creates a new blockchain with a genesis block
func New(difficulty int) *Chain {
	c := &Chain{
		Blocks:     make([]*block.Block, 0),
		Difficulty: difficulty,
	}
	c.createGenesisBlock()
	return c
}

// createGenesisBlock creates the first block in the chain
func (c *Chain) createGenesisBlock() {
	genesis := block.New(0, "Genesis Block", "0")
	genesis.Mine(c.Difficulty)
	c.Blocks = append(c.Blocks, genesis)
}

// AddBlock adds a new block to the chain
func (c *Chain) AddBlock(data string) error {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	newBlock := block.New(
		prevBlock.Index+1,
		data,
		prevBlock.Hash,
	)
	newBlock.Mine(c.Difficulty)

	if err := c.validateNewBlock(newBlock, prevBlock); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	c.Blocks = append(c.Blocks, newBlock)
	return nil
}

// validateNewBlock checks if a new block is valid
func (c *Chain) validateNewBlock(newBlock, prevBlock *block.Block) error {
	if newBlock.Index != prevBlock.Index+1 {
		return fmt.Errorf("invalid index: expected %d, got %d", prevBlock.Index+1, newBlock.Index)
	}

	if newBlock.PreviousHash != prevBlock.Hash {
		return fmt.Errorf("invalid previous hash")
	}

	if !newBlock.IsValid() {
		return fmt.Errorf("invalid hash")
	}

	// Verify proof-of-work
	target := ""
	for i := 0; i < c.Difficulty; i++ {
		target += "0"
	}
	if newBlock.Hash[:c.Difficulty] != target {
		return fmt.Errorf("insufficient proof-of-work")
	}

	return nil
}

// IsValid validates the entire blockchain
func (c *Chain) IsValid() bool {
	for i := 1; i < len(c.Blocks); i++ {
		currentBlock := c.Blocks[i]
		prevBlock := c.Blocks[i-1]

		if err := c.validateNewBlock(currentBlock, prevBlock); err != nil {
			fmt.Printf("Chain validation failed at block %d: %v\n", i, err)
			return false
		}
	}
	return true
}

// SaveToFile persists the blockchain to a JSON file
func (c *Chain) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// LoadFromFile loads the blockchain from a JSON file
func LoadFromFile(filename string) (*Chain, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var c Chain
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// GetLatestBlock returns the most recent block
func (c *Chain) GetLatestBlock() *block.Block {
	return c.Blocks[len(c.Blocks)-1]
}

// Length returns the number of blocks in the chain
func (c *Chain) Length() int {
	return len(c.Blocks)
}
