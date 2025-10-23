package block

import (
	"strings"
	"testing"
	"time"

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
	tx := createTestTransaction("alice", "bob", 10.0)
	transactions := []*transaction.Transaction{tx}
	b := New(1, transactions, "prev_hash")

	if b.Index != 1 {
		t.Errorf("expected index 1, got %d", b.Index)
	}
	if len(b.Transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(b.Transactions))
	}
	if b.Transactions[0].From != "alice" {
		t.Errorf("expected transaction from 'alice', got %s", b.Transactions[0].From)
	}
	if b.PreviousHash != "prev_hash" {
		t.Errorf("expected previous hash 'prev_hash', got %s", b.PreviousHash)
	}
	if b.Nonce != 0 {
		t.Errorf("expected nonce 0, got %d", b.Nonce)
	}
}

func TestCalculateHash(t *testing.T) {
	tx := createTestTransaction("genesis", "alice", 100.0)
	transactions := []*transaction.Transaction{tx}
	b := New(0, transactions, "0")
	b.Timestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	hash1 := b.CalculateHash()
	hash2 := b.CalculateHash()

	// Same block should produce same hash
	if hash1 != hash2 {
		t.Errorf("hash should be deterministic")
	}

	// Hash should be 64 characters (SHA-256 in hex)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}

	// Changing transaction should change hash
	b.Transactions[0].Amount = 200.0
	hash3 := b.CalculateHash()
	if hash1 == hash3 {
		t.Errorf("changing transaction should change hash")
	}
}

func TestMine(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
	}{
		{"difficulty 1", 1},
		{"difficulty 2", 2},
		{"difficulty 3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := createTestTransaction("alice", "bob", 10.0)
			transactions := []*transaction.Transaction{tx}
			b := New(0, transactions, "0")
			b.Mine(tt.difficulty)

			// Check that hash has required leading zeros
			expectedPrefix := strings.Repeat("0", tt.difficulty)
			if !strings.HasPrefix(b.Hash, expectedPrefix) {
				t.Errorf("expected hash to start with %d zeros, got hash: %s",
					tt.difficulty, b.Hash)
			}

			// Verify the hash is actually correct
			if !b.IsValid() {
				t.Errorf("mined block should have valid hash")
			}

			// Nonce should have been incremented
			if b.Nonce == 0 {
				t.Errorf("expected nonce to be incremented during mining")
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tx := createTestTransaction("alice", "bob", 10.0)
	transactions := []*transaction.Transaction{tx}
	b := New(0, transactions, "0")
	b.Mine(2)

	// Should be valid after mining
	if !b.IsValid() {
		t.Errorf("freshly mined block should be valid")
	}

	// Tampering with transaction should invalidate
	b.Transactions[0].Amount = 999.0
	if b.IsValid() {
		t.Errorf("block with tampered transaction should be invalid")
	}

	// Recalculate hash - should be valid again
	b.Hash = b.CalculateHash()
	if !b.IsValid() {
		t.Errorf("block with recalculated hash should be valid")
	}
}

func TestHashDeterminism(t *testing.T) {
	// Two blocks with identical properties should have identical hashes
	timestamp := time.Now()
	tx1 := createTestTransaction("alice", "bob", 10.0)
	tx2 := createTestTransaction("alice", "bob", 10.0)

	// Set same timestamp for both transactions
	tx1.Timestamp = timestamp
	tx2.Timestamp = timestamp
	tx1.ID = tx1.Hash()
	tx2.ID = tx2.Hash()

	b1 := &Block{
		Index:        1,
		Timestamp:    timestamp,
		Transactions: []*transaction.Transaction{tx1},
		PreviousHash: "prev",
		Nonce:        42,
	}

	b2 := &Block{
		Index:        1,
		Timestamp:    timestamp,
		Transactions: []*transaction.Transaction{tx2},
		PreviousHash: "prev",
		Nonce:        42,
	}

	hash1 := b1.CalculateHash()
	hash2 := b2.CalculateHash()

	if hash1 != hash2 {
		t.Errorf("identical blocks should produce identical hashes")
	}
}

func TestNonceImpactsHash(t *testing.T) {
	tx := createTestTransaction("alice", "bob", 10.0)
	transactions := []*transaction.Transaction{tx}
	b := New(0, transactions, "0")

	hash1 := b.CalculateHash()
	b.Nonce = 1
	hash2 := b.CalculateHash()

	if hash1 == hash2 {
		t.Errorf("changing nonce should change hash")
	}
}

func BenchmarkMine(b *testing.B) {
	// Benchmark mining at different difficulties
	difficulties := []int{1, 2, 3, 4}

	for _, diff := range difficulties {
		b.Run(string(rune('0'+diff)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tx := createTestTransaction("alice", "bob", 10.0)
				transactions := []*transaction.Transaction{tx}
				block := New(0, transactions, "0")
				block.Mine(diff)
			}
		})
	}
}
