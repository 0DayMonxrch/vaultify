package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrReplayDetected  = errors.New("session replay detected")
)

type SessionMetadata struct {
	UserID            string    `json:"user_id"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	CreatedAt         time.Time `json:"created_at"`
}

type SessionManager struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewSessionManager(rdb *redis.Client) *SessionManager {
	return &SessionManager{
		rdb: rdb,
		ttl: 7 * 24 * time.Hour, // 7 days
	}
}

// CreateSession creates a new session in Redis, adds it to the user's active sessions set, and returns the composite refresh token (user_id:token_uuid).
func (sm *SessionManager) CreateSession(ctx context.Context, userID, deviceFingerprint string) (string, error) {
	tokenUUID := uuid.NewString()

	metadata := SessionMetadata{
		UserID:            userID,
		DeviceFingerprint: deviceFingerprint,
		CreatedAt:         time.Now(),
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	pipe := sm.rdb.TxPipeline()

	// 1. SET session:{user_id}:{token_uuid} -> metadata JSON
	sessionKey := fmt.Sprintf("session:%s:%s", userID, tokenUUID)
	pipe.Set(ctx, sessionKey, metadataBytes, sm.ttl)

	// 2. SADD user:sessions:{user_id} -> token_uuid
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	pipe.SAdd(ctx, userSessionsKey, tokenUUID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", userID, tokenUUID), nil
}

// ValidateAndRotateSession validates the old refresh token, handles replay detection, rotates the token, and returns the new composite refresh token.
func (sm *SessionManager) ValidateAndRotateSession(ctx context.Context, compositeToken, deviceFingerprint string) (string, error) {
	parts := strings.SplitN(compositeToken, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid refresh token format")
	}
	userID := parts[0]
	tokenUUID := parts[1]

	sessionKey := fmt.Sprintf("session:%s:%s", userID, tokenUUID)
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)
	rotatedKey := fmt.Sprintf("rotated_session:%s:%s", userID, tokenUUID)

	// 1. Get session metadata
	val, err := sm.rdb.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		// Session not active. Check if it was recently rotated to detect replay.
		rotatedVal, err := sm.rdb.Get(ctx, rotatedKey).Result()
		if err == nil && rotatedVal == "1" {
			// Replay detected! Revoke all sessions for this user.
			_ = sm.RevokeAllSessions(ctx, userID)
			return "", ErrReplayDetected
		}
		return "", ErrSessionNotFound
	} else if err != nil {
		return "", err
	}

	// 2. Parse metadata
	var metadata SessionMetadata
	if err := json.Unmarshal([]byte(val), &metadata); err != nil {
		return "", err
	}

	// 3. Perform rotation (pipelined)
	newUUID := uuid.NewString()
	newMetadata := SessionMetadata{
		UserID:            userID,
		DeviceFingerprint: deviceFingerprint,
		CreatedAt:         time.Now(),
	}
	newMetadataBytes, err := json.Marshal(newMetadata)
	if err != nil {
		return "", err
	}

	newSessionKey := fmt.Sprintf("session:%s:%s", userID, newUUID)

	pipe := sm.rdb.TxPipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, tokenUUID)
	pipe.Set(ctx, rotatedKey, "1", 5*time.Minute)

	pipe.Set(ctx, newSessionKey, newMetadataBytes, sm.ttl)
	pipe.SAdd(ctx, userSessionsKey, newUUID)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", userID, newUUID), nil
}

// RevokeSession revokes a single active session.
func (sm *SessionManager) RevokeSession(ctx context.Context, compositeToken string) error {
	parts := strings.SplitN(compositeToken, ":", 2)
	if len(parts) != 2 {
		return errors.New("invalid refresh token format")
	}
	userID := parts[0]
	tokenUUID := parts[1]

	sessionKey := fmt.Sprintf("session:%s:%s", userID, tokenUUID)
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)

	pipe := sm.rdb.TxPipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, tokenUUID)

	_, err := pipe.Exec(ctx)
	return err
}

// RevokeAllSessions revokes all sessions associated with a specific user.
func (sm *SessionManager) RevokeAllSessions(ctx context.Context, userID string) error {
	userSessionsKey := fmt.Sprintf("user:sessions:%s", userID)

	uuids, err := sm.rdb.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return err
	}

	if len(uuids) == 0 {
		return nil
	}

	pipe := sm.rdb.TxPipeline()
	for _, tokenUUID := range uuids {
		sessionKey := fmt.Sprintf("session:%s:%s", userID, tokenUUID)
		pipe.Del(ctx, sessionKey)
	}
	pipe.Del(ctx, userSessionsKey)

	_, err = pipe.Exec(ctx)
	return err
}
