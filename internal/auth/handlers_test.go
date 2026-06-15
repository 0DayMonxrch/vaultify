package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
)

func TestHandlers_RegisterAndLoginLifecycle(t *testing.T) {
	ctx := context.Background()

	// Connect to local PostgreSQL
	dbURL := "postgres://vaultify:vaultify_password@localhost:5432/vaultify?sslmode=disable"
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skip("Postgres is not available locally, skipping handlers tests")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		t.Skipf("Postgres is not available locally, skipping handlers tests: %v", err)
	}

	// Apply migrations automatically
	migrationPath := "../../db/migrations/000001_init_schema.up.sql"
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	_, err = dbPool.Exec(ctx, string(sqlBytes))
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Connect to local Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not available locally, skipping handlers tests")
	}

	// Clean up database users to avoid collisions
	_, err = dbPool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate users table: %v", err)
	}

	queries := db.New(dbPool)
	sessionMgr := NewSessionManager(rdb)
	jwtSecret := []byte("test-jwt-secret-key-12345")

	handlers := NewHandlers(queries, sessionMgr, jwtSecret)

	// Setup Router
	r := chi.NewRouter()

	authenticator := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing auth header", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}
			claims, err := ParseAccessToken(parts[1], jwtSecret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxkey.UserID, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	handlers.RegisterRoutes(r, authenticator)

	// 1. Test POST /auth/register
	regReq := RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	reqBody, _ := json.Marshal(regReq)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got: %d (body: %s)", rec.Code, rec.Body.String())
	}

	var regResp RegisterResponse
	if err := json.NewDecoder(rec.Body).Decode(&regResp); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}

	if regResp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got: %s", regResp.Email)
	}

	if regResp.ID == "" {
		t.Error("expected non-empty user ID")
	}

	// Try registering again with the same email (should fail with 409 Conflict)
	req = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409 Conflict for duplicate email, got: %d", rec.Code)
	}

	// 2. Test POST /auth/login
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	reqBody, _ = json.Marshal(loginReq)
	req = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got: %d (body: %s)", rec.Code, rec.Body.String())
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("failed to decode login token response: %v", err)
	}

	if tokenResp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}

	// Test GET /auth/me
	meReq := httptest.NewRequest("GET", "/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	meRec := httptest.NewRecorder()
	r.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for GET /auth/me, got: %d (body: %s)", meRec.Code, meRec.Body.String())
	}

	var meResp RegisterResponse
	if err := json.NewDecoder(meRec.Body).Decode(&meResp); err != nil {
		t.Fatalf("failed to decode me response: %v", err)
	}

	if meResp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got: %s", meResp.Email)
	}

	if meResp.ID != regResp.ID {
		t.Errorf("expected ID %s, got: %s", regResp.ID, meResp.ID)
	}

	// Extract refresh token from cookie
	cookies := rec.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}

	if refreshCookie == nil {
		t.Fatal("expected refresh_token cookie in response, but got nil")
	}

	if !strings.HasPrefix(refreshCookie.Value, regResp.ID+":") {
		t.Errorf("expected cookie value to start with user ID %s, got: %s", regResp.ID, refreshCookie.Value)
	}

	// 3. Test POST /auth/refresh
	req = httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(refreshCookie)
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for refresh, got: %d (body: %s)", rec.Code, rec.Body.String())
	}

	var refreshResp TokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&refreshResp); err != nil {
		t.Fatalf("failed to decode refresh token response: %v", err)
	}

	if refreshResp.AccessToken == "" {
		t.Error("expected non-empty access token after refresh")
	}

	newCookies := rec.Result().Cookies()
	var newRefreshCookie *http.Cookie
	for _, c := range newCookies {
		if c.Name == "refresh_token" {
			newRefreshCookie = c
			break
		}
	}

	if newRefreshCookie == nil {
		t.Fatal("expected new refresh_token cookie after rotation, but got nil")
	}

	if newRefreshCookie.Value == refreshCookie.Value {
		t.Fatal("expected refresh token to be rotated, but it has the same value")
	}

	// 4. Test Replay Attack
	req = httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(refreshCookie) // old token
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for replay attack, got: %d", rec.Code)
	}

	// 5. Test DELETE /auth/logout
	req = httptest.NewRequest("DELETE", "/auth/logout", nil)
	req.AddCookie(newRefreshCookie)
	rec = httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content for logout, got: %d", rec.Code)
	}

	// Verify cookie is cleared
	logoutCookies := rec.Result().Cookies()
	var clearedCookie *http.Cookie
	for _, c := range logoutCookies {
		if c.Name == "refresh_token" {
			clearedCookie = c
			break
		}
	}

	if clearedCookie == nil {
		t.Fatal("expected cleared refresh_token cookie, got nil")
	}

	if clearedCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got: %d", clearedCookie.MaxAge)
	}
}
