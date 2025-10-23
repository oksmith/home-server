package chain

import (
	"os"
	"testing"
	"time"

	"github.com/oksmith/home-server/blockchain/pkg/block"
	"github.com/oksmith/home-server/blockchain/pkg/transaction"
)

// createTestTransaction creates a simple test transaction
func createTestTransaction(from, to string, amount float64) *transaction.Transaction {
	tx := transaction.New(from, to, amount)
	// Set a fixed timestamp for deterministic testing
	tx.Timestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	// Generate ID manually for testing
	tx.ID = tx.Hash()
	return tx
}

func TestNew(t *testing.T) {
	c := New(2, 10.0)

	if c.Difficulty != 2 {
		t.Errorf("expected difficulty 2, got %d", c.Difficulty)
	}

	if c.MiningReward != 10.0 {
		t.Errorf("expected mining reward 10.0, got %f", c.MiningReward)
	}

	// Should have genesis block
	if len(c.Blocks) != 1 {
		t.Errorf("expected 1 block (genesis), got %d", len(c.Blocks))
	}

	genesis := c.Blocks[0]
	if genesis.Index != 0 {
		t.Errorf("genesis block should have index 0, got %d", genesis.Index)
	}
	if genesis.PreviousHash != "0" {
		t.Errorf("genesis block should have previous hash '0', got %s", genesis.PreviousHash)
	}
}

func TestAddBlock(t *testing.T) {
	c := New(2, 10.0)

	// Create a test transaction
	tx := createTestTransaction("alice", "bob", 5.0)
	transactions := []*transaction.Transaction{tx}

	err := c.AddBlock(transactions, "miner")
	if err != nil {
		t.Fatalf("failed to add block: %v", err)
	}

	if len(c.Blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(c.Blocks))
	}

	// Check block was properly linked
	newBlock := c.Blocks[1]
	prevBlock := c.Blocks[0]

	if newBlock.Index != 1 {
		t.Errorf("expected index 1, got %d", newBlock.Index)
	}
	if newBlock.PreviousHash != prevBlock.Hash {
		t.Errorf("new block's previous hash doesn't match previous block's hash")
	}

	// Should have coinbase transaction + our transaction
	if len(newBlock.Transactions) != 2 {
		t.Errorf("expected 2 transactions (coinbase + user), got %d", len(newBlock.Transactions))
	}
}

func TestAddMultipleBlocks(t *testing.T) {
	c := New(2, 10.0)

	// Create multiple transactions for different blocks
	transactions1 := []*transaction.Transaction{createTestTransaction("alice", "bob", 5.0)}
	transactions2 := []*transaction.Transaction{createTestTransaction("bob", "charlie", 3.0)}
	transactions3 := []*transaction.Transaction{createTestTransaction("charlie", "alice", 2.0)}

	allTransactions := [][]*transaction.Transaction{transactions1, transactions2, transactions3}
	for i, txs := range allTransactions {
		if err := c.AddBlock(txs, "miner"); err != nil {
			t.Fatalf("failed to add block %d: %v", i+1, err)
		}
	}

	if len(c.Blocks) != 4 { // genesis + 3 blocks
		t.Errorf("expected 4 blocks, got %d", len(c.Blocks))
	}

	// Verify chain integrity
	if !c.IsValid() {
		t.Errorf("chain should be valid after adding blocks")
	}
}

func TestIsValid(t *testing.T) {
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	if !c.IsValid() {
		t.Errorf("valid chain should return true")
	}
}

func TestIsValidDetectsTampering(t *testing.T) {
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	// Tamper with transaction in middle block
	c.Blocks[1].Transactions[1].Amount = 999.0 // Tamper with the user transaction (index 1, not coinbase)

	if c.IsValid() {
		t.Errorf("chain should be invalid after tampering with transaction")
	}
}

func TestIsValidDetectsHashTampering(t *testing.T) {
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	// Tamper with hash
	c.Blocks[1].Hash = "fake_hash"

	if c.IsValid() {
		t.Errorf("chain should be invalid after tampering with hash")
	}
}

func TestIsValidDetectsBrokenLinks(t *testing.T) {
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	// Break the chain link
	c.Blocks[2].PreviousHash = "wrong_hash"

	if c.IsValid() {
		t.Errorf("chain should be invalid with broken links")
	}
}

func TestValidateNewBlock(t *testing.T) {
	c := New(2, 10.0)

	tests := []struct {
		name    string
		setup   func() (*block.Block, *block.Block)
		wantErr bool
	}{
		{
			name: "valid block",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				tx := createTestTransaction("alice", "bob", 5.0)
				new := block.New(1, []*transaction.Transaction{tx}, prev.Hash)
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: false,
		},
		{
			name: "wrong index",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				tx := createTestTransaction("alice", "bob", 5.0)
				new := block.New(5, []*transaction.Transaction{tx}, prev.Hash) // Should be 1, not 5
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: true,
		},
		{
			name: "wrong previous hash",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				tx := createTestTransaction("alice", "bob", 5.0)
				new := block.New(1, []*transaction.Transaction{tx}, "wrong_hash")
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: true,
		},
		{
			name: "insufficient proof of work",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				tx := createTestTransaction("alice", "bob", 5.0)
				new := block.New(1, []*transaction.Transaction{tx}, prev.Hash)
				new.Mine(1) // Mine with lower difficulty than required
				return new, prev
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newBlock, prevBlock := tt.setup()
			err := c.validateNewBlock(newBlock, prevBlock)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateNewBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoadFromFile(t *testing.T) {
	// Create a test chain
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	filename := "test_blockchain.json"
	defer os.Remove(filename) // Clean up after test

	// Save to file
	if err := c.SaveToFile(filename); err != nil {
		t.Fatalf("failed to save chain: %v", err)
	}

	// Load from file
	loaded, err := LoadFromFile(filename)
	if err != nil {
		t.Fatalf("failed to load chain: %v", err)
	}

	// Verify loaded chain matches original
	if len(loaded.Blocks) != len(c.Blocks) {
		t.Errorf("expected %d blocks, got %d", len(c.Blocks), len(loaded.Blocks))
	}

	if loaded.Difficulty != c.Difficulty {
		t.Errorf("expected difficulty %d, got %d", c.Difficulty, loaded.Difficulty)
	}

	if loaded.MiningReward != c.MiningReward {
		t.Errorf("expected mining reward %f, got %f", c.MiningReward, loaded.MiningReward)
	}

	// Verify loaded chain is valid
	if !loaded.IsValid() {
		t.Errorf("loaded chain should be valid")
	}

	// Verify block data matches
	for i := range c.Blocks {
		if loaded.Blocks[i].Hash != c.Blocks[i].Hash {
			t.Errorf("block %d hash mismatch", i)
		}
		if len(loaded.Blocks[i].Transactions) != len(c.Blocks[i].Transactions) {
			t.Errorf("block %d transaction count mismatch", i)
		}
	}
}

func TestGetLatestBlock(t *testing.T) {
	c := New(2, 10.0)

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")

	latest := c.GetLatestBlock()
	if latest.Index != 2 {
		t.Errorf("expected latest block index 2, got %d", latest.Index)
	}
	if len(latest.Transactions) != 2 {
		t.Errorf("expected 2 transactions in latest block, got %d", len(latest.Transactions))
	}
}

func TestLength(t *testing.T) {
	c := New(2, 10.0)
	if c.Length() != 1 {
		t.Errorf("expected length 1, got %d", c.Length())
	}

	tx := createTestTransaction("alice", "bob", 5.0)
	c.AddBlock([]*transaction.Transaction{tx}, "miner")
	if c.Length() != 2 {
		t.Errorf("expected length 2, got %d", c.Length())
	}
}

func TestChainIntegrity(t *testing.T) {
	// This test verifies that you can't easily tamper with the chain
	c := New(3, 10.0) // Higher difficulty for this test

	tx1 := createTestTransaction("alice", "bob", 5.0)
	tx2 := createTestTransaction("bob", "charlie", 3.0)
	tx3 := createTestTransaction("charlie", "alice", 2.0)

	c.AddBlock([]*transaction.Transaction{tx1}, "miner")
	c.AddBlock([]*transaction.Transaction{tx2}, "miner")
	c.AddBlock([]*transaction.Transaction{tx3}, "miner")

	// Attempt to tamper with middle block and recalculate its hash
	c.Blocks[1].Transactions[1].Amount = 999.0 // Tamper with user transaction
	c.Blocks[1].Hash = c.Blocks[1].CalculateHash()

	// Chain should still be invalid because the next block's
	// PreviousHash won't match the new hash
	// This is important! It's not easy to tamper with history.
	if c.IsValid() {
		t.Errorf("chain should detect tampering even with recalculated hash")
	}
}
