package db_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestTokenQueries(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vaultify:vaultify_password@localhost:5432/vaultify?sslmode=disable"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Apply migration for testing
	upSQL, _ := os.ReadFile(filepath.Join("..", "..", "db", "migrations", "006_api_tokens.up.sql"))
	if len(upSQL) > 0 {
		conn.Exec(ctx, string(upSQL))
	}

	queries := db.New(conn)

	// Create user
	email := uuid.New().String() + "@example.com"
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: "testhash",
	})
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create project
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		Name:      "Test Project",
		Slug:      uuid.New().String()[:8],
		KekSalt:   []byte("salt"),
		CreatedBy: pgtype.UUID{Bytes: user.ID.Bytes, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create token
	tokenHash := "testhash12345"
	tokenPrefix := "testpref"
	role := "read"
	tokenName := "Test Token"

	token, err := queries.CreateToken(ctx, db.CreateTokenParams{
		UserID:      pgtype.UUID{Bytes: user.ID.Bytes, Valid: true},
		ProjectID:   pgtype.UUID{Bytes: project.ID.Bytes, Valid: true},
		Name:        tokenName,
		TokenHash:   tokenHash,
		TokenPrefix: tokenPrefix,
		Role:        role,
		ExpiresAt:   pgtype.Timestamptz{Valid: false}, // No expiry
	})
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}
	if token.TokenHash != tokenHash || token.TokenPrefix != tokenPrefix {
		t.Fatalf("Token mismatch")
	}

	// Get Token by hash
	retrieved, err := queries.GetTokenByHash(ctx, tokenHash)
	if err != nil {
		t.Fatalf("Failed to get token by hash: %v", err)
	}
	if retrieved.ID.Bytes != token.ID.Bytes {
		t.Fatalf("Retrieved token ID mismatch")
	}

	// List User Tokens
	tokens, err := queries.ListUserTokens(ctx, pgtype.UUID{Bytes: user.ID.Bytes, Valid: true})
	if err != nil {
		t.Fatalf("Failed to list tokens: %v", err)
	}
	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	// Update last used
	err = queries.UpdateTokenLastUsed(ctx, token.ID)
	if err != nil {
		t.Fatalf("Failed to update last used: %v", err)
	}

	updated, err := queries.GetTokenByHash(ctx, tokenHash)
	if err != nil || !updated.LastUsedAt.Valid {
		t.Fatalf("LastUsedAt not updated")
	}

	// Soft Revoke
	err = queries.SoftRevokeToken(ctx, db.SoftRevokeTokenParams{
		ID:     token.ID,
		UserID: pgtype.UUID{Bytes: user.ID.Bytes, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to soft revoke token: %v", err)
	}

	// Get Token by hash again should fail
	_, err = queries.GetTokenByHash(ctx, tokenHash)
	if err != pgx.ErrNoRows {
		t.Fatalf("Expected ErrNoRows, got: %v", err)
	}
}
