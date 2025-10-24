package wallet

import (
	"testing"
)

func TestNew(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	if w.PrivateKey == nil {
		t.Error("wallet should have a private key")
	}

	if w.PublicKey == nil {
		t.Error("wallet should have a public key")
	}

	// Public key should match the one derived from private key
	if w.PublicKey != &w.PrivateKey.PublicKey {
		t.Error("public key should be derived from private key")
	}
}

func TestAddress(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	address := w.Address()

	// Address should be 64 characters (SHA-256 hex)
	if len(address) != 64 {
		t.Errorf("expected address length 64, got %d", len(address))
	}

	// Same wallet should produce same address
	address2 := w.Address()
	if address != address2 {
		t.Error("address should be deterministic")
	}
}

func TestDifferentWalletsHaveDifferentAddresses(t *testing.T) {
	w1, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet 1: %v", err)
	}

	w2, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet 2: %v", err)
	}

	if w1.Address() == w2.Address() {
		t.Error("different wallets should have different addresses")
	}
}

func TestSign(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	data := []byte("test message")
	signature, err := w.Sign(data)
	if err != nil {
		t.Fatalf("failed to sign data: %v", err)
	}

	// Signature should be 64 bytes (r || s)
	if len(signature) != 64 {
		t.Errorf("expected signature length 64, got %d", len(signature))
	}

	// Same data should produce different signatures (due to random k)
	signature2, err := w.Sign(data)
	if err != nil {
		t.Fatalf("failed to sign data again: %v", err)
	}

	// While signatures may differ, both should be valid
	if !VerifySignature(w.PublicKey, data, signature) {
		t.Error("signature should verify against public key")
	}

	if !VerifySignature(w.PublicKey, data, signature2) {
		t.Error("second signature should verify against public key")
	}
}

func TestVerifySignature(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	data := []byte("test message")
	signature, err := w.Sign(data)
	if err != nil {
		t.Fatalf("failed to sign data: %v", err)
	}

	// Valid signature should verify
	if !VerifySignature(w.PublicKey, data, signature) {
		t.Error("valid signature should verify")
	}

	// Wrong data should not verify
	wrongData := []byte("wrong message")
	if VerifySignature(w.PublicKey, wrongData, signature) {
		t.Error("signature should not verify with wrong data")
	}

	// Wrong public key should not verify
	w2, _ := New()
	if VerifySignature(w2.PublicKey, data, signature) {
		t.Error("signature should not verify with wrong public key")
	}

	// Invalid signature length should not verify
	invalidSig := []byte("invalid")
	if VerifySignature(w.PublicKey, data, invalidSig) {
		t.Error("invalid signature should not verify")
	}
}

func TestPublicKeyToAddress(t *testing.T) {
	w, err := New()
	if err != nil {
		t.Fatalf("failed to create wallet: %v", err)
	}

	// PublicKeyToAddress should produce same result as Address method
	address1 := w.Address()
	address2 := PublicKeyToAddress(w.PublicKey)

	if address1 != address2 {
		t.Error("PublicKeyToAddress should match Address method")
	}
}
