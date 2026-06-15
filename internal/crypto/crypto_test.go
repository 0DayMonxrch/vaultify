package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	masterKey := []byte("super-secret-master-key")
	salt := []byte("random-salt-12345")

	key1, err := DeriveKey(masterKey, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Expected 32-byte key, got %d bytes", len(key1))
	}

	// Ensure determinism
	key2, err := DeriveKey(masterKey, salt)
	if err != nil {
		t.Fatalf("DeriveKey failed: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey should be deterministic for the same inputs")
	}

	// Different salt -> different key
	key3, _ := DeriveKey(masterKey, []byte("different-salt"))
	if bytes.Equal(key1, key3) {
		t.Error("DeriveKey with different salt should produce a different key")
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("this is a very secret message")

	ciphertext, nonce, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if len(nonce) != 12 {
		t.Errorf("Expected 12-byte nonce, got %d bytes", len(nonce))
	}

	decrypted, err := Decrypt(ciphertext, nonce, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	defer clear(decrypted)

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Round-trip failed. Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptDecrypt_AuthFailures(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("this is a very secret message")
	ciphertext, nonce, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	t.Run("Modified Ciphertext", func(t *testing.T) {
		modCiphertext := make([]byte, len(ciphertext))
		copy(modCiphertext, ciphertext)
		modCiphertext[0] ^= 0xFF // flip bits in the first byte

		_, err := Decrypt(modCiphertext, nonce, key)
		if err == nil {
			t.Error("Expected error when decrypting modified ciphertext, got nil")
		}
	})

	t.Run("Modified Nonce", func(t *testing.T) {
		modNonce := make([]byte, len(nonce))
		copy(modNonce, nonce)
		modNonce[0] ^= 0xFF

		_, err := Decrypt(ciphertext, modNonce, key)
		if err == nil {
			t.Error("Expected error when decrypting with modified nonce, got nil")
		}
	})

	t.Run("Incorrect Key", func(t *testing.T) {
		wrongKey := make([]byte, 32)
		copy(wrongKey, key)
		wrongKey[0] ^= 0xFF

		_, err := Decrypt(ciphertext, nonce, wrongKey)
		if err == nil {
			t.Error("Expected error when decrypting with incorrect key, got nil")
		}
	})
}

func TestEncrypt_NonceUniqueness(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("this is a very secret message")

	ciphertext1, nonce1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt 1 failed: %v", err)
	}

	ciphertext2, nonce2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt 2 failed: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Error("Nonces should be unique across encryptions")
	}

	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Ciphertexts should be different for identical plaintexts due to different nonces")
	}
}
