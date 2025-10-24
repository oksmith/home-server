package node

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/oksmith/home-server/blockchain/pkg/block"
	"github.com/oksmith/home-server/blockchain/pkg/transaction"
)

// StartServer starts the HTTP server for the node
func (n *Node) StartServer() error {
	http.HandleFunc("/chain", n.handleGetChain)
	http.HandleFunc("/transaction", n.handleTransaction)
	http.HandleFunc("/block", n.handleBlock)
	http.HandleFunc("/peers", n.handlePeers)
	http.HandleFunc("/balance", n.handleBalance)
	http.HandleFunc("/mine", n.handleMine)

	fmt.Printf("[%s] Starting server...\n", n.Address)
	return http.ListenAndServe(n.Address, nil)
}

// handleGetChain returns the full blockchain
func (n *Node) handleGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n.Chain)
}

// handleTransaction handles incoming transactions
func (n *Node) handleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add sender as peer (peer discovery)
	senderAddr := r.Header.Get("X-Node-Address")
	if senderAddr != "" {
		n.AddPeer(senderAddr)
	}

	var tx transaction.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := n.ReceiveTransaction(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Transaction received")
}

// handleBlock handles incoming blocks
func (n *Node) handleBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Add sender as peer (peer discovery)
	senderAddr := r.Header.Get("X-Node-Address")
	if senderAddr != "" {
		n.AddPeer(senderAddr)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var newBlock block.Block
	if err := json.Unmarshal(body, &newBlock); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := n.ReceiveBlock(body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Block received, syncing chain")
}

// handlePeers handles peer management
func (n *Node) handlePeers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(n.GetPeers())

	case http.MethodPost:
		var req struct {
			Peer string `json:"peer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		n.AddPeer(req.Peer)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Peer added")

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleBalance returns balance for an address
func (n *Node) handleBalance(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "address parameter required", http.StatusBadRequest)
		return
	}

	balance := n.Chain.GetBalance(address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}

// handleMine triggers mining of a new block
func (n *Node) handleMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := n.Mine(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Block mined successfully")
}
