package tokens

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlers_TokenLifecycle(t *testing.T) {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vaultify:vaultify_password@localhost:5432/vaultify?sslmode=disable"
	}
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skip("Postgres is not available locally, skipping handlers tests")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		t.Skipf("Postgres is not available locally, skipping handlers tests: %v", err)
	}

	queries := db.New(dbPool)

	// Setup mock user and project
	userID := uuid.New()
	user, err := queries.CreateUser(ctx, db.CreateUserParams{
		Email:        userID.String() + "@example.com",
		PasswordHash: "testhash",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	userID = user.ID.Bytes

	projID := uuid.New()
	project, err := queries.CreateProject(ctx, db.CreateProjectParams{
		Name:      "Token Test Project",
		Slug:      projID.String()[:8],
		KekSalt:   []byte("salt"),
		CreatedBy: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}
	projID = project.ID.Bytes

	// Add user as owner to project
	err = queries.AddProjectMember(ctx, db.AddProjectMemberParams{
		ProjectID: pgtype.UUID{Bytes: projID, Valid: true},
		UserID:    pgtype.UUID{Bytes: userID, Valid: true},
		Role:      "owner",
	})
	if err != nil {
		t.Fatalf("Failed to add project member: %v", err)
	}

	// Setup Router and Handlers
	handlers := NewHandlers(queries)

	r := chi.NewRouter()

	// Mock JWT middleware behavior for tests that bypass actual JWT generation
	mockAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vaultifyToken := r.Header.Get("X-Vaultify-Token")
			if vaultifyToken != "" {
				// Mock what the real middleware does
				hash := Hash(vaultifyToken)
				tkn, err := queries.GetTokenByHash(r.Context(), hash)
				if err != nil {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), ctxkey.UserID, uuid.UUID(tkn.UserID.Bytes).String())
				ctx = context.WithValue(ctx, ctxkey.ProjectID, uuid.UUID(tkn.ProjectID.Bytes).String())
				ctx = context.WithValue(ctx, ctxkey.Role, tkn.Role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Mock JWT by checking a simple custom header for tests
			if r.Header.Get("X-Test-Auth") == "valid" {
				ctx := context.WithValue(r.Context(), ctxkey.UserID, userID.String())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}

	r.Group(func(r chi.Router) {
		r.Use(mockAuth)
		r.Post("/tokens", handlers.CreateToken)
		r.Get("/tokens", handlers.ListTokens)
		r.Delete("/tokens/{id}", handlers.RevokeToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(mockAuth)
		// We add a dummy endpoint to test X-Vaultify-Token injection
		r.Get("/test-secret", func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(ctxkey.Role).(string)
			_, _ = w.Write([]byte(role))
		})
	})

	// 1. Create Token
	reqBody, _ := json.Marshal(CreateTokenRequest{
		ProjectID: projID.String(),
		Name:      "CLI Token",
		Role:      "read",
	})
	req := httptest.NewRequest("POST", "/tokens", bytes.NewBuffer(reqBody))
	req.Header.Set("X-Test-Auth", "valid") // Simulate valid JWT
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for token creation, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var createResp CreateTokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("Failed to decode token response: %v", err)
	}
	if !strings.HasPrefix(createResp.Token, "vt_") {
		t.Errorf("Expected token prefix vt_, got %s", createResp.Token)
	}

	rawToken := createResp.Token

	// 2. Try creating token WITH an API Token (should fail)
	req = httptest.NewRequest("POST", "/tokens", bytes.NewBuffer(reqBody))
	req.Header.Set("X-Vaultify-Token", rawToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when creating token via API token, got %d", rec.Code)
	}

	// 3. Test token auth context injection
	req = httptest.NewRequest("GET", "/test-secret", nil)
	req.Header.Set("X-Vaultify-Token", rawToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for token auth, got %d", rec.Code)
	}
	if rec.Body.String() != "read" {
		t.Errorf("Expected injected role to be 'read', got '%s'", rec.Body.String())
	}

	// 4. List Tokens
	req = httptest.NewRequest("GET", "/tokens", nil)
	req.Header.Set("X-Test-Auth", "valid")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list tokens, got %d", rec.Code)
	}

	var listResp []map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&listResp); err != nil {
		t.Fatalf("Failed to decode list response: %v", err)
	}
	if len(listResp) != 1 {
		t.Fatalf("Expected 1 token in list, got %d", len(listResp))
	}

	tokenID := listResp[0]["id"].(string)

	// 5. Revoke Token
	req = httptest.NewRequest("DELETE", "/tokens/"+tokenID, nil)
	req.Header.Set("X-Test-Auth", "valid")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected 204 No Content for revoke, got %d", rec.Code)
	}

	// 6. Test revoked token (should fail auth)
	req = httptest.NewRequest("GET", "/test-secret", nil)
	req.Header.Set("X-Vaultify-Token", rawToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized for revoked token, got %d", rec.Code)
	}
}
