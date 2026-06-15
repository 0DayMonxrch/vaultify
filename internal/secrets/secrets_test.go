package secrets

import (
	"context"
	"encoding/base64"
	"reflect"
	"strings"
	"testing"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockAudit struct {
	logs []struct {
		ProjectID uuid.UUID
		Action    string
		KeyName   string
	}
}

func (m *mockAudit) WriteLog(ctx context.Context, projectID uuid.UUID, action string, keyName *string) error {
	k := ""
	if keyName != nil {
		k = *keyName
	}
	m.logs = append(m.logs, struct {
		ProjectID uuid.UUID
		Action    string
		KeyName   string
	}{ProjectID: projectID, Action: action, KeyName: k})
	return nil
}

type mockDBTX struct {
	t          *testing.T
	queryRowFn func(sql string, args ...interface{}) pgx.Row
}

func (m *mockDBTX) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockDBTX) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockDBTX) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.queryRowFn != nil {
		return m.queryRowFn(sql, args...)
	}
	return &mockRow{}
}

type mockRow struct {
	scanFn func(dest ...interface{}) error
}

func (m *mockRow) Scan(dest ...interface{}) error {
	if m.scanFn != nil {
		return m.scanFn(dest...)
	}
	return nil
}

func TestSecretService_RoundTripFlow(t *testing.T) {
	userID := uuid.New()
	projectID := uuid.New()
	secretID := uuid.New()
	masterKey := []byte("01234567890123456789012345678912") // 32 bytes
	kekSalt := []byte("salt1234")
	keyName := "DB_PASSWORD"
	plaintext := []byte("my-super-secret-password")

	ctx := context.WithValue(context.Background(), ctxkey.UserID, userID.String())

	var storedCiphertext, storedNonce string

	mDB := &mockDBTX{
		t: t,
		queryRowFn: func(sql string, args ...interface{}) pgx.Row {
			if strings.Contains(sql, "FROM projects") {
				return &mockRow{
					scanFn: func(dest ...interface{}) error {
						// Set project.KekSalt
						// returning: id, name, slug, kek_salt, created_by, created_at
						reflect.ValueOf(dest[3]).Elem().Set(reflect.ValueOf(kekSalt))
						return nil
					},
				}
			}
			if strings.Contains(sql, "INSERT INTO secrets") {
				// intercept the args to check ciphertext is passed, not plaintext
				storedCiphertext = args[3].(string) // EncryptedValue
				storedNonce = args[4].(string)      // Nonce
				
				// assert it's not plaintext
				rawCipher, _ := base64.StdEncoding.DecodeString(storedCiphertext)
				if string(rawCipher) == string(plaintext) {
					t.Errorf("plaintext was passed directly to the DB without encryption")
				}

				return &mockRow{
					scanFn: func(dest ...interface{}) error {
						// return a secret struct
						reflect.ValueOf(dest[0]).Elem().Set(reflect.ValueOf(pgtype.UUID{Bytes: secretID, Valid: true}))
						reflect.ValueOf(dest[1]).Elem().Set(reflect.ValueOf(pgtype.UUID{Bytes: projectID, Valid: true}))
						reflect.ValueOf(dest[2]).Elem().SetString(keyName)
						reflect.ValueOf(dest[4]).Elem().SetString(storedCiphertext)
						reflect.ValueOf(dest[5]).Elem().SetString(storedNonce)
						return nil
					},
				}
			}
			if strings.Contains(sql, "SELECT * FROM secrets") || strings.Contains(sql, "SELECT id, project_id, key_name") {
				return &mockRow{
					scanFn: func(dest ...interface{}) error {
						// return the secret struct exactly as it would be mapped by sqlc
						// fields: id, project_id, key_name, environment, encrypted_value, nonce, created_by, updated_at, created_at
						reflect.ValueOf(dest[0]).Elem().Set(reflect.ValueOf(pgtype.UUID{Bytes: secretID, Valid: true}))
						reflect.ValueOf(dest[1]).Elem().Set(reflect.ValueOf(pgtype.UUID{Bytes: projectID, Valid: true}))
						reflect.ValueOf(dest[2]).Elem().SetString(keyName)
						reflect.ValueOf(dest[3]).Elem().SetString("production")
						reflect.ValueOf(dest[4]).Elem().SetString(storedCiphertext)
						reflect.ValueOf(dest[5]).Elem().SetString(storedNonce)
						return nil
					},
				}
			}
			return &mockRow{scanFn: func(dest ...interface{}) error { return nil }}
		},
	}

	auditMock := &mockAudit{}
	queries := db.New(mDB)
	service := NewSecretService(queries, auditMock, masterKey)

	// Test Create
	plaintextCopy := make([]byte, len(plaintext))
	copy(plaintextCopy, plaintext)
	_, err := service.CreateSecret(ctx, projectID, keyName, "production", plaintextCopy)
	if err != nil {
		t.Fatalf("failed to create secret: %v", err)
	}

	if auditMock.logs[0].Action != "SECRET_WRITE" {
		t.Errorf("expected SECRET_WRITE audit log, got %s", auditMock.logs[0].Action)
	}

	// Plaintext slice should be cleared
	for _, b := range plaintextCopy {
		if b != 0 {
			t.Errorf("plaintext was not cleared")
			break
		}
	}

	// Test Get
	retrieved, err := service.GetSecret(ctx, projectID, secretID)
	if err != nil {
		t.Fatalf("failed to get secret: %v", err)
	}

	if string(retrieved) != string(plaintext) {
		t.Errorf("retrieved secret %q does not match original %q", retrieved, plaintext)
	}

	if auditMock.logs[1].Action != "SECRET_READ" {
		t.Errorf("expected SECRET_READ audit log, got %s", auditMock.logs[1].Action)
	}
}
