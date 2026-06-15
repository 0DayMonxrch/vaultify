package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/0DayMonxrch/vaultify/internal/auth"
	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)



type AuthMiddleware struct {
	jwtSecret []byte
	queries   *db.Queries
}

func NewAuthMiddleware(queries *db.Queries, jwtSecret []byte) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
		queries:   queries,
	}
}

func (m *AuthMiddleware) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims, err := auth.ParseAccessToken(tokenStr, m.jwtSecret)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		userID := claims.Subject
		if userID == "" {
			http.Error(w, "invalid token subject", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxkey.UserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) ContextEnricher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr, ok := r.Context().Value(ctxkey.UserID).(string)
		if !ok || userIDStr == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		projectIDStr := chi.URLParam(r, "id")
		if projectIDStr == "" {
			http.Error(w, "missing project id", http.StatusBadRequest)
			return
		}

		var projectUUID pgtype.UUID
		if err := projectUUID.Scan(projectIDStr); err != nil {
			http.Error(w, "invalid project id format", http.StatusBadRequest)
			return
		}

		var userUUID pgtype.UUID
		if err := userUUID.Scan(userIDStr); err != nil {
			http.Error(w, "invalid user id format", http.StatusUnauthorized)
			return
		}

		member, err := m.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
			ProjectID: projectUUID,
			UserID:    userUUID,
		})
		if err != nil {
			http.Error(w, "forbidden: not a project member", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), ctxkey.ProjectID, projectIDStr)
		ctx = context.WithValue(ctx, ctxkey.Role, member.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ctxkey.Role).(string)
		if !ok || role != "owner" {
			http.Error(w, "forbidden: requires owner role", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) RequireMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ctxkey.Role).(string)
		if !ok || role == "" {
			http.Error(w, "forbidden: requires project membership", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
