package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const TokenPrefix = "vt_"

// Generate creates a new API token with 32 bytes of entropy.
// It returns the full raw token string and the 8-character display prefix.
func Generate() (rawToken string, displayPrefix string, err error) {
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		return "", "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(entropy)

	// The specification dictates the display prefix is the first 8 chars of the encoded payload.
	displayPrefix = encoded[:8]
	rawToken = TokenPrefix + encoded

	return rawToken, displayPrefix, nil
}

// Hash computes the SHA-256 hash of the raw token and returns it as a hex string.
func Hash(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}
