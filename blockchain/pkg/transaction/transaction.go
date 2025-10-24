package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// Transaction represents a transfer of value between addresses
type Transaction struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
	Signature []byte    `json:"signature"`
}

// New creates a new unsigned transaction
func New(from, to string, amount float64) *Transaction {
	tx := &Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now(),
	}
	return tx
}

// Hash generates a unique identifier for the transaction
func (tx *Transaction) Hash() string {
	data := fmt.Sprintf("%s%s%f%s",
		tx.From,
		tx.To,
		tx.Amount,
		tx.Timestamp.Format(time.RFC3339Nano),
	)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// DataToSign returns the transaction data that should be signed
func (tx *Transaction) DataToSign() []byte {
	data := fmt.Sprintf("%s%s%f%s",
		tx.From,
		tx.To,
		tx.Amount,
		tx.Timestamp.Format(time.RFC3339Nano),
	)
	return []byte(data)
}

// Sign signs the transaction with the given private key
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey) error {
	dataToSign := tx.DataToSign()
	hash := sha256.Sum256(dataToSign)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Encode signature as r || s, padded to 32 bytes each
	signature := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)
	tx.Signature = signature
	tx.ID = tx.Hash()
	return nil
}

// Verify checks if the transaction signature is valid
func (tx *Transaction) Verify(publicKey *ecdsa.PublicKey) bool {
	if len(tx.Signature) != 64 {
		return false
	}

	dataToSign := tx.DataToSign()
	hash := sha256.Sum256(dataToSign)

	// Split signature into r and s
	r := new(big.Int).SetBytes(tx.Signature[:32])
	s := new(big.Int).SetBytes(tx.Signature[32:])

	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// IsValid performs basic validation checks
func (tx *Transaction) IsValid() error {
	if tx.From == "" {
		return fmt.Errorf("from address is required")
	}
	if tx.To == "" {
		return fmt.Errorf("to address is required")
	}
	if tx.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if len(tx.Signature) == 0 {
		return fmt.Errorf("transaction must be signed")
	}
	if tx.ID == "" {
		return fmt.Errorf("transaction must have an ID")
	}
	return nil
}

// IsCoinbase checks if this is a coinbase transaction (mining reward)
func (tx *Transaction) IsCoinbase() bool {
	return tx.From == "COINBASE"
}

// MarshalJSON implements custom JSON marshaling
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
		*Alias
	}{
		Timestamp: tx.Timestamp.Format(time.RFC3339Nano),
		Signature: hex.EncodeToString(tx.Signature),
		Alias:     (*Alias)(tx),
	})
}
