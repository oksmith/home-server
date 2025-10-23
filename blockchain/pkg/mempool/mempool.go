package mempool

import (
	"fmt"
	"sync"

	"github.com/oksmith/home-server/blockchain/pkg/transaction"
)

// Mempool holds pending transactions waiting to be mined
type Mempool struct {
	transactions map[string]*transaction.Transaction
	mu           sync.RWMutex // a lock that prevents data races when multiple goroutines access the same data
}

// New creates a new mempool
func New() *Mempool {
	return &Mempool{
		transactions: make(map[string]*transaction.Transaction),
	}
}

// Add adds a transaction to the mempool
func (m *Mempool) Add(tx *transaction.Transaction) error {
	if err := tx.IsValid(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if transaction already exists
	if _, exists := m.transactions[tx.ID]; exists {
		return fmt.Errorf("transaction %s already in mempool", tx.ID)
	}

	m.transactions[tx.ID] = tx
	return nil
}

// Remove removes a transaction from the mempool
func (m *Mempool) Remove(txID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.transactions, txID)
}

// Get retrieves a transaction by ID
func (m *Mempool) Get(txID string) (*transaction.Transaction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	tx, exists := m.transactions[txID]
	return tx, exists
}

// GetAll returns all pending transactions
func (m *Mempool) GetAll() []*transaction.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make([]*transaction.Transaction, 0, len(m.transactions))
	for _, tx := range m.transactions {
		txs = append(txs, tx)
	}
	return txs
}

// GetN returns up to n transactions for mining
func (m *Mempool) GetN(n int) []*transaction.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make([]*transaction.Transaction, 0, n)
	count := 0
	for _, tx := range m.transactions {
		if count >= n {
			break
		}
		txs = append(txs, tx)
		count++
	}
	return txs
}

// Size returns the number of transactions in the mempool
func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.transactions)
}

// Clear removes all transactions from the mempool
func (m *Mempool) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.transactions = make(map[string]*transaction.Transaction)
}

// RemoveTransactions removes multiple transactions (used after mining a block)
func (m *Mempool) RemoveTransactions(txs []*transaction.Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tx := range txs {
		delete(m.transactions, tx.ID)
	}
}
