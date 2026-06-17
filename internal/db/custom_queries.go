package db

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

type ListProjectMembersRow struct {
	UserID pgtype.UUID
	Email  string
	Role   string
}

func (q *Queries) ListProjectMembers(ctx context.Context, projectID pgtype.UUID) ([]ListProjectMembersRow, error) {
	query := `
		SELECT pm.user_id, u.email, pm.role 
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1
	`
	rows, err := q.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListProjectMembersRow
	for rows.Next() {
		var i ListProjectMembersRow
		if err := rows.Scan(&i.UserID, &i.Email, &i.Role); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	return items, nil
}

type AuditLogWithEmailRow struct {
	ID        int64
	ProjectID pgtype.UUID
	UserEmail string
	Action    string
	KeyName   pgtype.Text
	IpAddress *netip.Addr
	CreatedAt pgtype.Timestamptz
}

func (q *Queries) ListAuditLogsWithEmail(ctx context.Context, projectID pgtype.UUID, limit, offset int32) ([]AuditLogWithEmailRow, error) {
	query := `
		SELECT a.id, a.project_id, u.email, a.action, a.key_name, a.ip_address, a.created_at
		FROM audit_log a
		JOIN users u ON u.id = a.user_id
		WHERE a.project_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := q.db.Query(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AuditLogWithEmailRow
	for rows.Next() {
		var i AuditLogWithEmailRow
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.UserEmail, &i.Action, &i.KeyName, &i.IpAddress, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (q *Queries) CountAuditLogs(ctx context.Context, projectID pgtype.UUID) (int, error) {
	var total int
	err := q.db.QueryRow(ctx, "SELECT COUNT(*) FROM audit_log WHERE project_id = $1", projectID).Scan(&total)
	return total, err
}
