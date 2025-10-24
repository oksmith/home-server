package mempool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"sync"
	"testing"

	"github.com/oksmith/home-server/blockchain/pkg/transaction"
)

func createSignedTransaction(from, to string, amount float64) *transaction.Transaction {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tx := transaction.New(from, to, amount)
	tx.Sign(privateKey)
	return tx
}

func TestNew(t *testing.T) {
	m := New()

	if m == nil {
		t.Fatal("mempool should not be nil")
	}

	if m.Size() != 0 {
		t.Errorf("new mempool should be empty, got size %d", m.Size())
	}
}

func TestAdd(t *testing.T) {
	m := New()
	tx := createSignedTransaction("alice", "bob", 10.0)

	err := m.Add(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	if m.Size() != 1 {
		t.Errorf("expected size 1, got %d", m.Size())
	}

	// Verify transaction can be retrieved
	retrieved, exists := m.Get(tx.ID)
	if !exists {
		t.Error("transaction should exist in mempool")
	}
	if retrieved.ID != tx.ID {
		t.Error("retrieved transaction should match added transaction")
	}
}

func TestAddDuplicate(t *testing.T) {
	m := New()
	tx := createSignedTransaction("alice", "bob", 10.0)

	err := m.Add(tx)
	if err != nil {
		t.Fatalf("failed to add transaction: %v", err)
	}

	// Try to add same transaction again
	err = m.Add(tx)
	if err == nil {
		t.Error("adding duplicate transaction should return error")
	}

	if m.Size() != 1 {
		t.Errorf("duplicate transaction should not increase size, got %d", m.Size())
	}
}

func TestAddInvalidTransaction(t *testing.T) {
	m := New()

	// Create invalid transaction (not signed)
	invalidTx := transaction.New("alice", "bob", 10.0)

	err := m.Add(invalidTx)
	if err == nil {
		t.Error("adding invalid transaction should return error")
	}

	if m.Size() != 0 {
		t.Errorf("invalid transaction should not be added, got size %d", m.Size())
	}
}

func TestRemove(t *testing.T) {
	m := New()
	tx := createSignedTransaction("alice", "bob", 10.0)

	m.Add(tx)
	if m.Size() != 1 {
		t.Fatal("transaction should be added")
	}

	m.Remove(tx.ID)
	if m.Size() != 0 {
		t.Errorf("expected size 0 after removal, got %d", m.Size())
	}

	_, exists := m.Get(tx.ID)
	if exists {
		t.Error("removed transaction should not exist")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	m := New()

	// Removing non-existent transaction should not panic
	m.Remove("nonexistent")

	if m.Size() != 0 {
		t.Errorf("expected size 0, got %d", m.Size())
	}
}

func TestGet(t *testing.T) {
	m := New()
	tx := createSignedTransaction("alice", "bob", 10.0)

	m.Add(tx)

	retrieved, exists := m.Get(tx.ID)
	if !exists {
		t.Fatal("transaction should exist")
	}

	if retrieved.ID != tx.ID {
		t.Error("retrieved transaction ID should match")
	}
	if retrieved.Amount != tx.Amount {
		t.Error("retrieved transaction amount should match")
	}
}

func TestGetNonExistent(t *testing.T) {
	m := New()

	_, exists := m.Get("nonexistent")
	if exists {
		t.Error("non-existent transaction should not exist")
	}
}

func TestGetAll(t *testing.T) {
	m := New()

	tx1 := createSignedTransaction("alice", "bob", 10.0)
	tx2 := createSignedTransaction("bob", "charlie", 5.0)
	tx3 := createSignedTransaction("charlie", "alice", 3.0)

	m.Add(tx1)
	m.Add(tx2)
	m.Add(tx3)

	all := m.GetAll()
	if len(all) != 3 {
		t.Errorf("expected 3 transactions, got %d", len(all))
	}

	// Verify all transactions are present
	ids := make(map[string]bool)
	for _, tx := range all {
		ids[tx.ID] = true
	}

	if !ids[tx1.ID] || !ids[tx2.ID] || !ids[tx3.ID] {
		t.Error("all added transactions should be present")
	}
}

func TestGetAllEmpty(t *testing.T) {
	m := New()

	all := m.GetAll()
	if len(all) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(all))
	}
}

func TestGetN(t *testing.T) {
	m := New()

	tx1 := createSignedTransaction("alice", "bob", 10.0)
	tx2 := createSignedTransaction("bob", "charlie", 5.0)
	tx3 := createSignedTransaction("charlie", "alice", 3.0)

	m.Add(tx1)
	m.Add(tx2)
	m.Add(tx3)

	// Get 2 transactions
	txs := m.GetN(2)
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}

	// Get more than available
	txs = m.GetN(10)
	if len(txs) != 3 {
		t.Errorf("expected 3 transactions (max available), got %d", len(txs))
	}

	// Get 0 transactions
	txs = m.GetN(0)
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txs))
	}
}

func TestClear(t *testing.T) {
	m := New()

	tx1 := createSignedTransaction("alice", "bob", 10.0)
	tx2 := createSignedTransaction("bob", "charlie", 5.0)

	m.Add(tx1)
	m.Add(tx2)

	if m.Size() != 2 {
		t.Fatal("transactions should be added")
	}

	m.Clear()

	if m.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", m.Size())
	}

	_, exists := m.Get(tx1.ID)
	if exists {
		t.Error("transactions should be removed after clear")
	}
}

func TestRemoveTransactions(t *testing.T) {
	m := New()

	tx1 := createSignedTransaction("alice", "bob", 10.0)
	tx2 := createSignedTransaction("bob", "charlie", 5.0)
	tx3 := createSignedTransaction("charlie", "alice", 3.0)

	m.Add(tx1)
	m.Add(tx2)
	m.Add(tx3)

	// Remove tx1 and tx2
	m.RemoveTransactions([]*transaction.Transaction{tx1, tx2})

	if m.Size() != 1 {
		t.Errorf("expected size 1, got %d", m.Size())
	}

	_, exists := m.Get(tx3.ID)
	if !exists {
		t.Error("tx3 should still exist")
	}

	_, exists = m.Get(tx1.ID)
	if exists {
		t.Error("tx1 should be removed")
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := New()
	var wg sync.WaitGroup

	// Pre-create transactions to avoid timing issues with ID generation
	transactions := make([]*transaction.Transaction, 10)
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		from := fmt.Sprintf("alice%d", i)
		to := fmt.Sprintf("bob%d", i)
		transactions[i] = createSignedTransaction(from, to, float64(i+1))
		if ids[transactions[i].ID] {
			t.Fatalf("duplicate transaction ID at index %d: %s", i, transactions[i].ID)
		}
		ids[transactions[i].ID] = true
	}

	// Concurrent adds
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Add(transactions[n])
		}(i)
	}

	wg.Wait()

	if m.Size() != 10 {
		t.Errorf("expected 10 transactions after concurrent adds, got %d", m.Size())
	}

	// Concurrent reads and removes
	wg = sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(2)

		// Reader
		go func() {
			defer wg.Done()
			m.GetAll()
		}()

		// Remover
		go func() {
			defer wg.Done()
			txs := m.GetN(1)
			if len(txs) > 0 {
				m.Remove(txs[0].ID)
			}
		}()
	}

	wg.Wait()

	// Should not panic and final size should be consistent
	finalSize := m.Size()
	if finalSize > 10 {
		t.Errorf("size should not exceed initial count, got %d", finalSize)
	}
}
