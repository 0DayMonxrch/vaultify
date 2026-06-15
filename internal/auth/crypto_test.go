package auth

import (
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "supersecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error during hashing, got: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash string")
	}

	// Verify correct password
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("expected no error during verification, got: %v", err)
	}
	if !match {
		t.Fatal("expected password to match")
	}

	// Verify incorrect password
	match, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("expected no error during verification of incorrect password, got: %v", err)
	}
	if match {
		t.Fatal("expected wrong password to not match")
	}
}

func TestVerifyPassword_MalformedHash(t *testing.T) {
	malformedHashes := []string{
		"not-a-hash",
		"$argon2id$v=19$m=65536,t=3,p=4$salt",
		"$argon2id$v=18$m=65536,t=3,p=4$salt$hash", // wrong version
		"$argon2i$v=19$m=65536,t=3,p=4$salt$hash",  // wrong algorithm
		"$argon2id$v=19$m=abc,t=3,p=4$salt$hash",   // malformed parameters
	}

	for _, h := range malformedHashes {
		match, err := VerifyPassword("password", h)
		if err == nil {
			t.Errorf("expected error for malformed hash %q, but got nil (match=%v)", h, match)
		}
	}
}
