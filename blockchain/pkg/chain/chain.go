package chain

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/oksmith/home-server/blockchain/pkg/block"
	"github.com/oksmith/home-server/blockchain/pkg/transaction"
)

// Chain represents the blockchain with account state
type Chain struct {
	Blocks       []*block.Block     `json:"blocks"`
	Difficulty   int                `json:"difficulty"`
	MiningReward float64            `json:"mining_reward"`
	balances     map[string]float64 // Address -> Balance
	publicKeys   map[string]*ecdsa.PublicKey
}

// New creates a new blockchain with a genesis block
func New(difficulty int, miningReward float64) *Chain {
	c := &Chain{
		Blocks:       make([]*block.Block, 0),
		Difficulty:   difficulty,
		MiningReward: miningReward,
		balances:     make(map[string]float64),
		publicKeys:   make(map[string]*ecdsa.PublicKey),
	}
	c.createGenesisBlock()
	return c
}

// createGenesisBlock creates the first block in the chain
func (c *Chain) createGenesisBlock() {
	genesis := block.New(0, []*transaction.Transaction{}, "0")
	genesis.Mine(c.Difficulty)
	c.Blocks = append(c.Blocks, genesis)
}

// RegisterPublicKey associates a public key with an address
// This is needed for signature verification
func (c *Chain) RegisterPublicKey(address string, publicKey *ecdsa.PublicKey) {
	c.publicKeys[address] = publicKey
}

// GetBalance returns the balance for an address
func (c *Chain) GetBalance(address string) float64 {
	return c.balances[address]
}

// AddBlock mines a new block with the given transactions
func (c *Chain) AddBlock(transactions []*transaction.Transaction, minerAddress string) error {
	// Validate all transactions
	if err := c.validateTransactions(transactions); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Add coinbase transaction (mining reward)
	coinbase := transaction.New("COINBASE", minerAddress, c.MiningReward)
	coinbase.ID = coinbase.Hash()
	allTransactions := append([]*transaction.Transaction{coinbase}, transactions...)

	prevBlock := c.Blocks[len(c.Blocks)-1]
	newBlock := block.New(
		prevBlock.Index+1,
		allTransactions,
		prevBlock.Hash,
	)
	newBlock.Mine(c.Difficulty)

	if err := c.validateNewBlock(newBlock, prevBlock); err != nil {
		return fmt.Errorf("block validation failed: %w", err)
	}

	c.Blocks = append(c.Blocks, newBlock)

	// Apply transactions to update balances
	c.applyTransactions(allTransactions)

	return nil
}

// validateTransactions checks if all transactions are valid
func (c *Chain) validateTransactions(transactions []*transaction.Transaction) error {
	// Create a copy of current balances to simulate transaction application
	tempBalances := make(map[string]float64)
	for addr, balance := range c.balances {
		tempBalances[addr] = balance
	}

	for _, tx := range transactions {
		// Basic validation
		if err := tx.IsValid(); err != nil {
			return err
		}

		// Skip signature check for coinbase
		if tx.IsCoinbase() {
			continue
		}

		// Verify signature
		pubKey, exists := c.publicKeys[tx.From]
		if !exists {
			return fmt.Errorf("public key not registered for address %s", tx.From)
		}

		if !tx.Verify(pubKey) {
			return fmt.Errorf("invalid signature for transaction %s", tx.ID)
		}

		// Check balance against simulated state (prevents double-spending in same block)
		if tempBalances[tx.From] < tx.Amount {
			return fmt.Errorf("insufficient balance: address %s has %.2f but tried to send %.2f",
				tx.From, tempBalances[tx.From], tx.Amount)
		}

		// Update simulated balances
		tempBalances[tx.From] -= tx.Amount
		tempBalances[tx.To] += tx.Amount
	}
	return nil
}

// applyTransactions updates account balances
func (c *Chain) applyTransactions(transactions []*transaction.Transaction) {
	for _, tx := range transactions {
		if !tx.IsCoinbase() {
			c.balances[tx.From] -= tx.Amount
		}
		c.balances[tx.To] += tx.Amount
	}
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
	// Rebuild state from scratch
	tempBalances := make(map[string]float64)

	for i := 1; i < len(c.Blocks); i++ {
		currentBlock := c.Blocks[i]
		prevBlock := c.Blocks[i-1]

		// Validate block structure
		if err := c.validateNewBlock(currentBlock, prevBlock); err != nil {
			fmt.Printf("Chain validation failed at block %d: %v\n", i, err)
			return false
		}

		// Validate and apply transactions
		for _, tx := range currentBlock.Transactions {
			if !tx.IsCoinbase() {
				if tempBalances[tx.From] < tx.Amount {
					fmt.Printf("Invalid transaction in block %d: insufficient balance\n", i)
					return false
				}
				tempBalances[tx.From] -= tx.Amount
			}
			tempBalances[tx.To] += tx.Amount
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

	// Rebuild balances from blocks
	c.balances = make(map[string]float64)
	c.publicKeys = make(map[string]*ecdsa.PublicKey)

	for _, block := range c.Blocks {
		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				c.balances[tx.From] -= tx.Amount
			}
			c.balances[tx.To] += tx.Amount
		}
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
