package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"

	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
)

type Handlers struct {
	queries    *db.Queries
	sessionMgr *SessionManager
	jwtSecret  []byte
}

func NewHandlers(queries *db.Queries, sessionMgr *SessionManager, jwtSecret []byte) *Handlers {
	return &Handlers{
		queries:    queries,
		sessionMgr: sessionMgr,
		jwtSecret:  jwtSecret,
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to hash password")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		// Detect unique violation (could use pgconn.PgError check if needed, but error string check is fine for simplified logic)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "user with this email already exists", http.StatusConflict)
			return
		}
		log.Error().Err(err).Msg("failed to create user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userIDStr := uuidToString(user.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(RegisterResponse{
		ID:    userIDStr,
		Email: user.Email,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Log failed auth but return generic 401
		log.Warn().Str("email", req.Email).Msg("login failed: user not found")
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	match, err := VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		log.Warn().Str("email", req.Email).Msg("login failed: wrong password")
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	userIDStr := uuidToString(user.ID)
	userAgent := r.Header.Get("User-Agent")

	// Create session in Redis
	compositeToken, err := h.sessionMgr.CreateSession(r.Context(), userIDStr, userAgent)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session in redis")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Set HttpOnly cookie
	isProd := os.Getenv("ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    compositeToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600, // 7 days
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	})

	// Generate JWT
	accessToken, err := GenerateAccessToken(userIDStr, h.jwtSecret)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   600, // 10 minutes
	})
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "missing refresh token", http.StatusUnauthorized)
		return
	}

	compositeToken := cookie.Value
	userAgent := r.Header.Get("User-Agent")

	newCompositeToken, err := h.sessionMgr.ValidateAndRotateSession(r.Context(), compositeToken, userAgent)
	if err != nil {
		if errors.Is(err, ErrReplayDetected) {
			log.Warn().Str("token", compositeToken).Msg("REPLAY ATTACK DETECTED! Revoking all sessions.")
			// Clear cookie
			h.clearCookie(w)
			http.Error(w, "security violation: replay detected", http.StatusForbidden)
			return
		}
		if errors.Is(err, ErrSessionNotFound) {
			h.clearCookie(w)
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}
		log.Error().Err(err).Msg("failed to validate/rotate session")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Parse User ID from new composite token
	parts := strings.SplitN(newCompositeToken, ":", 2)
	if len(parts) != 2 {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr := parts[0]

	// Set new HttpOnly cookie
	isProd := os.Getenv("ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newCompositeToken,
		Path:     "/",
		MaxAge:   7 * 24 * 3600, // 7 days
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	})

	// Generate new JWT
	accessToken, err := GenerateAccessToken(userIDStr, h.jwtSecret)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate access token")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   600, // 10 minutes
	})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		// Already logged out or no token, just clear cookie and return 204
		h.clearCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	compositeToken := cookie.Value
	globalLogout := r.URL.Query().Get("global") == "true"

	if globalLogout {
		parts := strings.SplitN(compositeToken, ":", 2)
		if len(parts) == 2 {
			userIDStr := parts[0]
			if err := h.sessionMgr.RevokeAllSessions(r.Context(), userIDStr); err != nil {
				log.Error().Err(err).Msg("failed to revoke all sessions during global logout")
			}
		}
	} else {
		if err := h.sessionMgr.RevokeSession(r.Context(), compositeToken); err != nil {
			log.Error().Err(err).Msg("failed to revoke session during logout")
		}
	}

	h.clearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) clearCookie(w http.ResponseWriter) {
	isProd := os.Getenv("ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProd,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handlers) GetMe(w http.ResponseWriter, r *http.Request) {
	// The user ID is put in the context by the Authenticator middleware
	userIDStr, ok := r.Context().Value(ctxkey.UserID).(string)
	if !ok || userIDStr == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var userUUID pgtype.UUID
	if err := userUUID.Scan(userIDStr); err != nil {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserById(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(RegisterResponse{
		ID:    uuidToString(user.ID),
		Email: user.Email,
	})
}

// RegisterRoutes registers auth endpoints into Chi router
func (h *Handlers) RegisterRoutes(r chi.Router, authenticator func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Delete("/logout", h.Logout) // Not strictly behind JWT in router, handled by Cookie

		r.With(authenticator).Get("/me", h.GetMe)
	})
}

// Helper to convert pgtype.UUID to string
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// UUID byte array layout is 16 bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}
