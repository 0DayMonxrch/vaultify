package tokens

import (
	"encoding/json"
	"net/http"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handlers struct {
	queries *db.Queries
}

func NewHandlers(queries *db.Queries) *Handlers {
	return &Handlers{queries: queries}
}

func (h *Handlers) RegisterRoutes(r chi.Router, authenticator func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(authenticator)
		r.Post("/tokens", h.CreateToken)
		r.Get("/tokens", h.ListTokens)
		r.Delete("/tokens/{id}", h.RevokeToken)
	})
}

// requireJWT ensures that API tokens cannot be used to call these endpoints.
func requireJWT(r *http.Request) bool {
	return r.Header.Get("X-Vaultify-Token") == ""
}

type CreateTokenRequest struct {
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Role      string `json:"role"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

func (h *Handlers) CreateToken(w http.ResponseWriter, r *http.Request) {
	if !requireJWT(r) {
		http.Error(w, "API tokens cannot create other tokens", http.StatusForbidden)
		return
	}

	userIDStr, _ := r.Context().Value(ctxkey.UserID).(string)
	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "invalid user context", http.StatusInternalServerError)
		return
	}

	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	projUUID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		http.Error(w, "invalid project_id format", http.StatusBadRequest)
		return
	}

	// Verify the user is actually an owner of the project to create tokens?
	// The prompt doesn't strictly say owner only, but normally yes. Let's just create it.
	// Actually, verify membership first.
	member, err := h.queries.GetProjectMember(r.Context(), db.GetProjectMemberParams{
		ProjectID: pgtype.UUID{Bytes: projUUID, Valid: true},
		UserID:    pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err != nil || member.Role != "owner" {
		http.Error(w, "forbidden: must be project owner to create tokens", http.StatusForbidden)
		return
	}

	rawToken, prefix, err := Generate()
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	hash := Hash(rawToken)

	_, err = h.queries.CreateToken(r.Context(), db.CreateTokenParams{
		UserID:      pgtype.UUID{Bytes: userUUID, Valid: true},
		ProjectID:   pgtype.UUID{Bytes: projUUID, Valid: true},
		Name:        req.Name,
		TokenHash:   hash,
		TokenPrefix: prefix,
		Role:        req.Role,
		ExpiresAt:   pgtype.Timestamptz{Valid: false}, // No expiry by default for now
	})
	if err != nil {
		http.Error(w, "failed to store token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateTokenResponse{Token: rawToken})
}

func (h *Handlers) ListTokens(w http.ResponseWriter, r *http.Request) {
	if !requireJWT(r) {
		http.Error(w, "API tokens cannot list tokens", http.StatusForbidden)
		return
	}

	userIDStr, _ := r.Context().Value(ctxkey.UserID).(string)
	userUUID, _ := uuid.Parse(userIDStr)

	tokens, err := h.queries.ListUserTokens(r.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		http.Error(w, "failed to fetch tokens", http.StatusInternalServerError)
		return
	}

	// We don't return the hash. We can map to a safe response.
	type SafeToken struct {
		ID          string  `json:"id"`
		ProjectID   string  `json:"project_id"`
		Name        string  `json:"name"`
		TokenPrefix string  `json:"token_prefix"`
		Role        string  `json:"role"`
		Revoked     bool    `json:"revoked"`
		CreatedAt   string  `json:"created_at"`
		LastUsedAt  *string `json:"last_used_at"`
	}

	var resp []SafeToken
	for _, t := range tokens {
		var lastUsed *string
		if t.LastUsedAt.Valid {
			str := t.LastUsedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			lastUsed = &str
		}

		var createdAt string
		if t.CreatedAt.Valid {
			createdAt = t.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		resp = append(resp, SafeToken{
			ID:          uuid.UUID(t.ID.Bytes).String(),
			ProjectID:   uuid.UUID(t.ProjectID.Bytes).String(),
			Name:        t.Name,
			TokenPrefix: t.TokenPrefix,
			Role:        t.Role,
			Revoked:     t.Revoked,
			CreatedAt:   createdAt,
			LastUsedAt:  lastUsed,
		})
	}

	if resp == nil {
		resp = []SafeToken{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if !requireJWT(r) {
		http.Error(w, "API tokens cannot revoke tokens", http.StatusForbidden)
		return
	}

	userIDStr, _ := r.Context().Value(ctxkey.UserID).(string)
	userUUID, _ := uuid.Parse(userIDStr)

	tokenIDStr := chi.URLParam(r, "id")
	tokenUUID, err := uuid.Parse(tokenIDStr)
	if err != nil {
		http.Error(w, "invalid token id", http.StatusBadRequest)
		return
	}

	err = h.queries.SoftRevokeToken(r.Context(), db.SoftRevokeTokenParams{
		ID:     pgtype.UUID{Bytes: tokenUUID, Valid: true},
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
