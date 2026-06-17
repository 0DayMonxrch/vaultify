package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestSessionManager(t *testing.T) {
	ctx := context.Background()

	// Try to connect to local Redis running via compose.yaml
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not available locally, skipping SessionManager tests")
	}

	sm := NewSessionManager(rdb)
	userID := "user-123-uuid"
	deviceFingerprint := "mozilla-firefox-windows"

	// Clean up user sessions first to ensure clean state
	_ = sm.RevokeAllSessions(ctx, userID)

	// 1. Create session
	compositeToken, err := sm.CreateSession(ctx, userID, deviceFingerprint)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// 2. Validate and Rotate (normal flow)
	newCompositeToken, err := sm.ValidateAndRotateSession(ctx, compositeToken, deviceFingerprint)
	if err != nil {
		t.Fatalf("failed to rotate session: %v", err)
	}

	if newCompositeToken == compositeToken {
		t.Fatal("expected new composite token to be different after rotation")
	}

	// 3. Replay attack: try using the old token again
	_, err = sm.ValidateAndRotateSession(ctx, compositeToken, deviceFingerprint)
	if !errors.Is(err, ErrReplayDetected) {
		t.Fatalf("expected ErrReplayDetected, got: %v", err)
	}

	// 4. Verify all sessions were revoked for that user because of the replay attack
	_, err = sm.ValidateAndRotateSession(ctx, newCompositeToken, deviceFingerprint)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound for rotated token, got: %v", err)
	}

	// Clean up
	_ = sm.RevokeAllSessions(ctx, userID)
}
