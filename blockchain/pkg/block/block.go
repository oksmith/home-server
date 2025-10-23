package block

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block represents a single block in the blockchain
type Block struct {
	Index        int64     `json:"index"`
	Timestamp    time.Time `json:"timestamp"`
	Data         string    `json:"data"`
	PreviousHash string    `json:"previous_hash"`
	Hash         string    `json:"hash"`
	Nonce        int64     `json:"nonce"`
}

// New creates a new block with the given data
func New(index int64, data string, previousHash string) *Block {
	b := &Block{
		Index:        index,
		Timestamp:    time.Now(),
		Data:         data,
		PreviousHash: previousHash,
		Nonce:        0,
	}
	return b
}

// CalculateHash computes the SHA-256 hash of the block's contents
func (b *Block) CalculateHash() string {
	record := fmt.Sprintf("%d%s%s%s%d",
		b.Index,
		b.Timestamp.Format(time.RFC3339Nano),
		b.Data,
		b.PreviousHash,
		b.Nonce,
	)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Mine performs proof-of-work to find a valid hash with the specified difficulty
// difficulty is the number of leading zeros required in the hash
func (b *Block) Mine(difficulty int) {
	target := make([]byte, difficulty)
	for i := range target {
		target[i] = '0'
	}
	targetStr := string(target)

	for {
		b.Hash = b.CalculateHash()
		if b.Hash[:difficulty] == targetStr {
			fmt.Printf("Mined block %d with hash: %s (nonce: %d)\n", b.Index, b.Hash, b.Nonce)
			return
		}
		b.Nonce++
	}
}

// IsValid checks if the block's hash is correct
func (b *Block) IsValid() bool {
	return b.Hash == b.CalculateHash()
}

// MarshalJSON implements custom JSON marshaling
func (b *Block) MarshalJSON() ([]byte, error) {
	type Alias Block
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Timestamp: b.Timestamp.Format(time.RFC3339Nano),
		Alias:     (*Alias)(b),
	})
}
