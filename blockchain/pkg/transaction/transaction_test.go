package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"
)

func createTestWallet() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func TestNew(t *testing.T) {
	tx := New("alice", "bob", 10.0)

	if tx.From != "alice" {
		t.Errorf("expected from 'alice', got %s", tx.From)
	}
	if tx.To != "bob" {
		t.Errorf("expected to 'bob', got %s", tx.To)
	}
	if tx.Amount != 10.0 {
		t.Errorf("expected amount 10.0, got %f", tx.Amount)
	}
	if tx.Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}
	if tx.ID != "" {
		t.Error("new transaction should not have ID until signed")
	}
	if len(tx.Signature) != 0 {
		t.Error("new transaction should not be signed")
	}
}

func TestHash(t *testing.T) {
	tx := New("alice", "bob", 10.0)
	tx.Timestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	hash1 := tx.Hash()
	hash2 := tx.Hash()

	// Same transaction should produce same hash
	if hash1 != hash2 {
		t.Error("hash should be deterministic")
	}

	// Hash should be 64 characters (SHA-256 in hex)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}

	// Changing amount should change hash
	tx.Amount = 20.0
	hash3 := tx.Hash()
	if hash1 == hash3 {
		t.Error("changing amount should change hash")
	}
}

func TestSign(t *testing.T) {
	privateKey, err := createTestWallet()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	tx := New("alice", "bob", 10.0)
	err = tx.Sign(privateKey)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	// Signature should be set
	if len(tx.Signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(tx.Signature))
	}

	// ID should be set after signing
	if tx.ID == "" {
		t.Error("transaction ID should be set after signing")
	}

	// ID should match hash
	if tx.ID != tx.Hash() {
		t.Error("transaction ID should match hash")
	}
}

func TestVerify(t *testing.T) {
	privateKey, err := createTestWallet()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	tx := New("alice", "bob", 10.0)
	tx.Sign(privateKey)

	// Valid signature should verify
	if !tx.Verify(&privateKey.PublicKey) {
		t.Error("valid signature should verify")
	}

	// Wrong public key should not verify
	wrongKey, _ := createTestWallet()
	if tx.Verify(&wrongKey.PublicKey) {
		t.Error("signature should not verify with wrong public key")
	}

	// Tampering with transaction should invalidate signature
	tx.Amount = 999.0
	if tx.Verify(&privateKey.PublicKey) {
		t.Error("tampered transaction should not verify")
	}
}

func TestIsValid(t *testing.T) {
	privateKey, err := createTestWallet()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	tests := []struct {
		name    string
		setup   func() *Transaction
		wantErr bool
	}{
		{
			name: "valid transaction",
			setup: func() *Transaction {
				tx := New("alice", "bob", 10.0)
				tx.Sign(privateKey)
				return tx
			},
			wantErr: false,
		},
		{
			name: "missing from address",
			setup: func() *Transaction {
				tx := New("", "bob", 10.0)
				tx.Sign(privateKey)
				return tx
			},
			wantErr: true,
		},
		{
			name: "missing to address",
			setup: func() *Transaction {
				tx := New("alice", "", 10.0)
				tx.Sign(privateKey)
				return tx
			},
			wantErr: true,
		},
		{
			name: "zero amount",
			setup: func() *Transaction {
				tx := New("alice", "bob", 0)
				tx.Sign(privateKey)
				return tx
			},
			wantErr: true,
		},
		{
			name: "negative amount",
			setup: func() *Transaction {
				tx := New("alice", "bob", -5.0)
				tx.Sign(privateKey)
				return tx
			},
			wantErr: true,
		},
		{
			name: "missing signature",
			setup: func() *Transaction {
				return New("alice", "bob", 10.0)
			},
			wantErr: true,
		},
		{
			name: "missing ID",
			setup: func() *Transaction {
				tx := New("alice", "bob", 10.0)
				tx.Signature = []byte("fake signature fake signature fake signature fake signature fake sig")
				return tx
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := tt.setup()
			err := tx.IsValid()

			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsCoinbase(t *testing.T) {
	coinbase := New("COINBASE", "miner", 50.0)
	if !coinbase.IsCoinbase() {
		t.Error("transaction from COINBASE should be identified as coinbase")
	}

	regular := New("alice", "bob", 10.0)
	if regular.IsCoinbase() {
		t.Error("regular transaction should not be identified as coinbase")
	}
}

func TestDataToSign(t *testing.T) {
	tx := New("alice", "bob", 10.0)
	tx.Timestamp = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	data1 := tx.DataToSign()
	data2 := tx.DataToSign()

	// Should be deterministic
	if string(data1) != string(data2) {
		t.Error("DataToSign should be deterministic")
	}

	// Changing transaction should change data
	tx.Amount = 20.0
	data3 := tx.DataToSign()
	if string(data1) == string(data3) {
		t.Error("changing transaction should change DataToSign")
	}
}
