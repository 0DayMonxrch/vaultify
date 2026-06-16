package tokens

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Run("Format and Prefix", func(t *testing.T) {
		rawToken, prefix, err := Generate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(rawToken, TokenPrefix) {
			t.Errorf("expected token to start with %s, got %s", TokenPrefix, rawToken)
		}

		if len(prefix) != 8 {
			t.Errorf("expected prefix to be 8 characters, got %d", len(prefix))
		}

		// TokenPrefix (3) + 44 base64 characters = 47 characters total
		if len(rawToken) != 47 {
			t.Errorf("expected raw token length to be 47, got %d", len(rawToken))
		}

		if !strings.HasPrefix(rawToken, TokenPrefix+prefix) {
			t.Errorf("expected raw token to contain prefix correctly")
		}
	})

	t.Run("Entropy Uniqueness", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			rawToken, _, err := Generate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if seen[rawToken] {
				t.Fatalf("duplicate token generated: %s", rawToken)
			}
			seen[rawToken] = true
		}
	})
}

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		rawToken string
	}{
		{
			name:     "Basic test token",
			rawToken: "vt_hello_world",
		},
		{
			name:     "Long token matching spec format",
			rawToken: "vt_8a7f9b2cA1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8S9t0U1v",
		},
		{
			name:     "Empty string",
			rawToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Hash(tt.rawToken)

			// Compute expected manually to verify exact match
			h := sha256.Sum256([]byte(tt.rawToken))
			expected := hex.EncodeToString(h[:])

			if got != expected {
				t.Errorf("Hash(%q) = %v; want %v", tt.rawToken, got, expected)
			}

			// Determinism check: hashing same token should yield same result
			got2 := Hash(tt.rawToken)
			if got != got2 {
				t.Errorf("Hash is not deterministic for %q", tt.rawToken)
			}
		})
	}
}
