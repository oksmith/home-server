package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Wallet represents a blockchain wallet with public/private key pair
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// New generates a new wallet with a random key pair
func New() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// Address returns the wallet's public address (derived from public key)
func (w *Wallet) Address() string {
	// In production, this would use more sophisticated address derivation
	// For learning, we'll use a simple hash of the public key
	pubKeyBytes := append(w.PublicKey.X.Bytes(), w.PublicKey.Y.Bytes()...)
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[:])
}

// Sign creates a signature for the given data using the wallet's private key
// this is general purpose to sign any data, currently not used anywhere in the
// blockchain. transaction.Sign will do the actual signing.
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature as r || s
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature verifies a signature against data and a public key
func VerifySignature(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)

	// Split signature into r and s
	if len(signature) != 64 {
		return false
	}

	r := new(ecdsa.PublicKey).X
	s := new(ecdsa.PublicKey).Y
	r.SetBytes(signature[:32])
	s.SetBytes(signature[32:])

	return ecdsa.Verify(publicKey, hash[:], r, s)
}

// PublicKeyFromAddress is a simplified lookup function
// In production, you'd maintain a mapping of addresses to public keys
// For now, we'll store this mapping in the chain state
func PublicKeyToAddress(pubKey *ecdsa.PublicKey) string {
	pubKeyBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	hash := sha256.Sum256(pubKeyBytes)
	return hex.EncodeToString(hash[:])
}
