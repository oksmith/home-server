package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/oksmith/home-server/blockchain/pkg/chain"
	"github.com/oksmith/home-server/blockchain/pkg/mempool"
	"github.com/oksmith/home-server/blockchain/pkg/transaction"
	"github.com/oksmith/home-server/blockchain/pkg/wallet"
)

// Node represents a blockchain node with networking capabilities
type Node struct {
	Chain       *chain.Chain
	Mempool     *mempool.Mempool
	Wallet      *wallet.Wallet
	Address     string   // This node's address (e.g., "localhost:8080")
	Peers       []string // List of peer addresses
	peersMutex  sync.RWMutex
	isMining    bool
	miningMutex sync.Mutex
}

// New creates a new blockchain node
func New(address string, difficulty int, miningReward float64) (*Node, error) {
	w, err := wallet.New()
	if err != nil {
		return nil, err
	}

	c := chain.New(difficulty, miningReward)
	c.RegisterPublicKey(w.Address(), w.PublicKey)

	return &Node{
		Chain:   c,
		Mempool: mempool.New(),
		Wallet:  w,
		Address: address,
		Peers:   make([]string, 0),
	}, nil
}

// AddPeer adds a peer to the node's peer list
func (n *Node) AddPeer(peerAddress string) {
	n.peersMutex.Lock()
	defer n.peersMutex.Unlock()

	// Don't add self or duplicates
	if peerAddress == n.Address {
		return
	}
	for _, peer := range n.Peers {
		if peer == peerAddress {
			return
		}
	}

	n.Peers = append(n.Peers, peerAddress)
	fmt.Printf("[%s] Added peer: %s\n", n.Address, peerAddress)
}

// GetPeers returns a copy of the peer list
func (n *Node) GetPeers() []string {
	n.peersMutex.RLock()
	defer n.peersMutex.RUnlock()

	peers := make([]string, len(n.Peers))
	copy(peers, n.Peers)
	return peers
}

// BroadcastTransaction sends a transaction to all peers
func (n *Node) BroadcastTransaction(tx *transaction.Transaction) {
	peers := n.GetPeers()
	for _, peer := range peers {
		go func(peerAddr string) {
			url := fmt.Sprintf("http://%s/transaction", peerAddr)
			data, _ := json.Marshal(tx)

			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Node-Address", n.Address)

			client := &http.Client{Timeout: 5 * time.Second}
			client.Do(req)
		}(peer)
	}
}

// BroadcastBlock sends a block to all peers
func (n *Node) BroadcastBlock() {
	latestBlock := n.Chain.GetLatestBlock()
	peers := n.GetPeers()

	for _, peer := range peers {
		go func(peerAddr string) {
			url := fmt.Sprintf("http://%s/block", peerAddr)
			data, _ := json.Marshal(latestBlock)

			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Node-Address", n.Address)

			client := &http.Client{Timeout: 5 * time.Second}
			client.Do(req)
		}(peer)
	}
}

// SyncWithPeers synchronizes the chain with peers
func (n *Node) SyncWithPeers() error {
	peers := n.GetPeers()
	if len(peers) == 0 {
		return nil
	}

	// Announce ourselves to peers (helps establish bidirectional connections)
	for _, peer := range peers {
		go func(peerAddr string) {
			url := fmt.Sprintf("http://%s/peers", peerAddr)
			data := map[string]string{"peer": n.Address}
			jsonData, _ := json.Marshal(data)
			http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		}(peer)
	}

	var longestChain *chain.Chain
	maxLength := n.Chain.Length()

	for _, peer := range peers {
		url := fmt.Sprintf("http://%s/chain", peer)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}

		var peerChain chain.Chain
		if err := json.NewDecoder(resp.Body).Decode(&peerChain); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// Rebuild the chain state (balances and public keys from blocks)
		if err := peerChain.RebuildState(); err != nil {
			continue
		}

		// Check if peer's chain is longer and valid
		if peerChain.Length() > maxLength && peerChain.IsValid() {
			maxLength = peerChain.Length()
			longestChain = &peerChain
		}
	}

	// Replace chain if a longer valid chain was found
	if longestChain != nil {
		fmt.Printf("[%s] Replacing chain with longer chain (length: %d)\n", n.Address, maxLength)
		// Re-register our own public key with the new chain
		longestChain.RegisterPublicKey(n.Wallet.Address(), n.Wallet.PublicKey)
		n.Chain = longestChain
		return nil
	}

	return nil
}

// Mine attempts to mine a block with pending transactions
func (n *Node) Mine() error {
	n.miningMutex.Lock()
	if n.isMining {
		n.miningMutex.Unlock()
		return fmt.Errorf("already mining")
	}
	n.isMining = true
	n.miningMutex.Unlock()

	defer func() {
		n.miningMutex.Lock()
		n.isMining = false
		n.miningMutex.Unlock()
	}()

	// Get transactions from mempool
	transactions := n.Mempool.GetAll()

	fmt.Printf("[%s] Mining block with %d transactions...\n", n.Address, len(transactions))

	// Add block to chain
	if err := n.Chain.AddBlock(transactions, n.Wallet.Address()); err != nil {
		return err
	}

	// Remove mined transactions from mempool
	n.Mempool.RemoveTransactions(transactions)

	// Broadcast the new block
	n.BroadcastBlock()

	fmt.Printf("[%s] Mined block %d!\n", n.Address, n.Chain.GetLatestBlock().Index)

	return nil
}

// StartMining continuously mines blocks
func (n *Node) StartMining(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if n.Mempool.Size() > 0 {
				n.Mine()
			}
		}
	}()
}

// ReceiveTransaction handles incoming transactions from peers
func (n *Node) ReceiveTransaction(tx *transaction.Transaction) error {
	// Add to mempool
	if err := n.Mempool.Add(tx); err != nil {
		return err
	}

	fmt.Printf("[%s] Received transaction: %s -> %s (%.2f coins)\n",
		n.Address, tx.From[:8], tx.To[:8], tx.Amount)

	// Relay to other peers
	n.BroadcastTransaction(tx)

	return nil
}

// ReceiveBlock handles incoming blocks from peers
func (n *Node) ReceiveBlock(newBlock []byte) error {
	// Sync with peers to get the full chain
	return n.SyncWithPeers()
}
