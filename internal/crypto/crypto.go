// Package crypto provides the foundational cryptographic operations for Vaultify.
//
// Memory Safety Requirement:
// Plaintext data (such as decrypted secrets or keys) must be zeroed out of memory
// immediately after it is no longer needed to prevent exposure in core dumps or memory swapping.
// Consumers of this package MUST use the Go 1.21+ built-in clear() function on the plaintext
// byte slices returned by Decrypt.
// Example:
//
//	plaintext, err := crypto.Decrypt(cipher, nonce, key)
//	// ... use plaintext ...
//	clear(plaintext)
//
// DO NOT use os.Getenv inside this package. All keys must be explicitly passed.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters
	time    = 1
	memory  = 64 * 1024
	threads = 4
	keyLen  = 32

	// nonceSize is the standard size for AES-GCM nonces
	nonceSize = 12
)

var (
	ErrInvalidCiphertext = errors.New("crypto: invalid ciphertext or nonce")
)

// DeriveKey derives a 32-byte key suitable for AES-256 from a master key and a salt
// using Argon2id.
func DeriveKey(masterKey []byte, salt []byte) ([]byte, error) {
	if len(masterKey) == 0 {
		return nil, errors.New("crypto: master key cannot be empty")
	}
	if len(salt) == 0 {
		return nil, errors.New("crypto: salt cannot be empty")
	}

	// argon2.IDKey(password, salt, time, memory, threads, keyLen)
	key := argon2.IDKey(masterKey, salt, time, memory, threads, keyLen)
	return key, nil
}

// Encrypt encrypts the plaintext using AES-256-GCM.
// It returns the ciphertext and the securely generated 12-byte nonce.
func Encrypt(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts the ciphertext using AES-256-GCM.
// Consumers MUST use clear() on the returned plaintext after use.
func Decrypt(ciphertext []byte, nonce []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != nonceSize {
		return nil, errors.New("crypto: invalid nonce size")
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}
