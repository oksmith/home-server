package chain

import (
	"os"
	"testing"

	"github.com/oksmith/home-server/blockchain/pkg/block"
)

func TestNew(t *testing.T) {
	c := New(2)

	if c.Difficulty != 2 {
		t.Errorf("expected difficulty 2, got %d", c.Difficulty)
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
	c := New(2)

	err := c.AddBlock("first block")
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
	if newBlock.Data != "first block" {
		t.Errorf("expected data 'first block', got %s", newBlock.Data)
	}
}

func TestAddMultipleBlocks(t *testing.T) {
	c := New(2)

	blocks := []string{"block 1", "block 2", "block 3"}
	for _, data := range blocks {
		if err := c.AddBlock(data); err != nil {
			t.Fatalf("failed to add block: %v", err)
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
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

	if !c.IsValid() {
		t.Errorf("valid chain should return true")
	}
}

func TestIsValidDetectsTampering(t *testing.T) {
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

	// Tamper with data in middle block
	c.Blocks[1].Data = "tampered data"

	if c.IsValid() {
		t.Errorf("chain should be invalid after tampering with data")
	}
}

func TestIsValidDetectsHashTampering(t *testing.T) {
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

	// Tamper with hash
	c.Blocks[1].Hash = "fake_hash"

	if c.IsValid() {
		t.Errorf("chain should be invalid after tampering with hash")
	}
}

func TestIsValidDetectsBrokenLinks(t *testing.T) {
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

	// Break the chain link
	c.Blocks[2].PreviousHash = "wrong_hash"

	if c.IsValid() {
		t.Errorf("chain should be invalid with broken links")
	}
}

func TestValidateNewBlock(t *testing.T) {
	c := New(2)

	tests := []struct {
		name    string
		setup   func() (*block.Block, *block.Block)
		wantErr bool
	}{
		{
			name: "valid block",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				new := block.New(1, "test", prev.Hash)
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: false,
		},
		{
			name: "wrong index",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				new := block.New(5, "test", prev.Hash) // Should be 1, not 5
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: true,
		},
		{
			name: "wrong previous hash",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				new := block.New(1, "test", "wrong_hash")
				new.Mine(c.Difficulty)
				return new, prev
			},
			wantErr: true,
		},
		{
			name: "insufficient proof of work",
			setup: func() (*block.Block, *block.Block) {
				prev := c.Blocks[0]
				new := block.New(1, "test", prev.Hash)
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
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

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

	// Verify loaded chain is valid
	if !loaded.IsValid() {
		t.Errorf("loaded chain should be valid")
	}

	// Verify block data matches
	for i := range c.Blocks {
		if loaded.Blocks[i].Hash != c.Blocks[i].Hash {
			t.Errorf("block %d hash mismatch", i)
		}
		if loaded.Blocks[i].Data != c.Blocks[i].Data {
			t.Errorf("block %d data mismatch", i)
		}
	}
}

func TestGetLatestBlock(t *testing.T) {
	c := New(2)
	c.AddBlock("block 1")
	c.AddBlock("block 2")

	latest := c.GetLatestBlock()
	if latest.Index != 2 {
		t.Errorf("expected latest block index 2, got %d", latest.Index)
	}
	if latest.Data != "block 2" {
		t.Errorf("expected latest block data 'block 2', got %s", latest.Data)
	}
}

func TestLength(t *testing.T) {
	c := New(2)
	if c.Length() != 1 {
		t.Errorf("expected length 1, got %d", c.Length())
	}

	c.AddBlock("block 1")
	if c.Length() != 2 {
		t.Errorf("expected length 2, got %d", c.Length())
	}
}

func TestChainIntegrity(t *testing.T) {
	// This test verifies that you can't easily tamper with the chain
	c := New(3) // Higher difficulty for this test
	c.AddBlock("transaction 1")
	c.AddBlock("transaction 2")
	c.AddBlock("transaction 3")

	// Attempt to tamper with middle block and recalculate its hash
	c.Blocks[1].Data = "fraudulent transaction"
	c.Blocks[1].Hash = c.Blocks[1].CalculateHash()

	// Chain should still be invalid because the next block's
	// PreviousHash won't match the new hash
	// This is important! It's not easy to tamper with history.
	if c.IsValid() {
		t.Errorf("chain should detect tampering even with recalculated hash")
	}
}
