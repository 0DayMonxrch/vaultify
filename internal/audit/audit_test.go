package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockDB struct {
	lastArgs []interface{}
}

func (m *mockDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	m.lastArgs = args
	return pgconn.CommandTag{}, nil
}

func (m *mockDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return nil
}

func TestWriteLog_ContextExtraction(t *testing.T) {
	mDB := &mockDB{}
	queries := db.New(mDB)
	service := NewAuditService(queries)

	projectID := uuid.New()
	userID := uuid.New()
	ipStr := "192.168.1.100"

	ctx := context.WithValue(context.Background(), ctxkey.UserID, userID.String())
	ctx = context.WithValue(ctx, ctxkey.IPAddress, ipStr)

	keyName := "my-secret-key"
	err := service.WriteLog(ctx, projectID, "create", &keyName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mDB.lastArgs) != 5 {
		t.Fatalf("expected 5 args for InsertAuditLog, got %d", len(mDB.lastArgs))
	}
}

func TestWriteLog_ContextFailure(t *testing.T) {
	queries := db.New(&mockDB{})
	service := NewAuditService(queries)

	projectID := uuid.New()

	// Empty context
	err := service.WriteLog(context.Background(), projectID, "create", nil)
	if !errors.Is(err, ErrMissingUserContext) {
		t.Fatalf("expected ErrMissingUserContext, got %v", err)
	}

	// Invalid UUID in context
	ctx := context.WithValue(context.Background(), ctxkey.UserID, "not-a-uuid")
	err = service.WriteLog(ctx, projectID, "create", nil)
	if !errors.Is(err, ErrMissingUserContext) {
		t.Fatalf("expected ErrMissingUserContext, got %v", err)
	}
}
