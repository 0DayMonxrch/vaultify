package audit

import (
	"context"
	"errors"
	"net/netip"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrMissingUserContext = errors.New("audit: missing user context")
)

type AuditService struct {
	queries *db.Queries
}

func NewAuditService(queries *db.Queries) *AuditService {
	return &AuditService{queries: queries}
}

// WriteLog appends a new audit log. UserID and IPAddress are extracted from the context.
func (s *AuditService) WriteLog(ctx context.Context, projectID uuid.UUID, action string, keyName *string) error {
	userIDStr, ok := ctx.Value(ctxkey.UserID).(string)
	if !ok || userIDStr == "" {
		return ErrMissingUserContext
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrMissingUserContext
	}

	var ipAddr *netip.Addr
	ipStr, ok := ctx.Value(ctxkey.IPAddress).(string)
	if ok && ipStr != "" {
		if addr, err := netip.ParseAddr(ipStr); err == nil {
			ipAddr = &addr
		}
	}

	var keyNamePg pgtype.Text
	if keyName != nil {
		keyNamePg = pgtype.Text{String: *keyName, Valid: true}
	}

	return s.queries.InsertAuditLog(ctx, db.InsertAuditLogParams{
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		Action:    action,
		KeyName:   keyNamePg,
		IpAddress: ipAddr,
	})
}

// ListProjectLogs retrieves logs for a specific project with pagination.
func (s *AuditService) ListProjectLogs(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]db.AuditLog, error) {
	return s.queries.ListAuditLogsByProject(ctx, db.ListAuditLogsByProjectParams{
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		Limit:     limit,
		Offset:    offset,
	})
}

// ListProjectLogsWithEmail retrieves logs joined with user emails
func (s *AuditService) ListProjectLogsWithEmail(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]db.AuditLogWithEmailRow, error) {
	return s.queries.ListAuditLogsWithEmail(ctx, pgtype.UUID{Bytes: projectID, Valid: true}, limit, offset)
}

func (s *AuditService) CountProjectLogs(ctx context.Context, projectID uuid.UUID) (int, error) {
	return s.queries.CountAuditLogs(ctx, pgtype.UUID{Bytes: projectID, Valid: true})
}
