package main

import (
	"context"
	"encoding/hex"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/0DayMonxrch/vaultify/frontend"
	"github.com/0DayMonxrch/vaultify/internal/audit"
	"github.com/0DayMonxrch/vaultify/internal/auth"
	"github.com/0DayMonxrch/vaultify/internal/ctxkey"
	"github.com/0DayMonxrch/vaultify/internal/db"
	"github.com/0DayMonxrch/vaultify/internal/middleware"
	"github.com/0DayMonxrch/vaultify/internal/projects"
	"github.com/0DayMonxrch/vaultify/internal/secrets"
	"github.com/0DayMonxrch/vaultify/internal/tokens"
)

func main() {
	// Setup zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Info().Msg("No .env file found, relying on environment variables")
	}

	ctx := context.Background()

	// Parse configuration
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://vaultify:vaultify_password@localhost:5432/vaultify?sslmode=disable" // #nosec G101
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	// Connect to Postgres
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create connection pool to database")
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("Unable to ping database")
	}
	log.Info().Msg("Connected to PostgreSQL")

	// Apply migrations on startup
	migrationPaths := []string{
		"db/migrations/000001_init_schema.up.sql",
		"db/migrations/004_secrets.up.sql",
		"db/migrations/005_audit_log.up.sql",
	}
	for _, migrationPath := range migrationPaths {
		// #nosec G304
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			log.Warn().Err(err).Msgf("Unable to read migration file %s", migrationPath)
			continue
		}
		_, err = dbPool.Exec(ctx, string(sqlBytes))
		if err != nil {
			log.Warn().Err(err).Msgf("Unable to execute database migration %s", migrationPath)
		}
	}
	log.Info().Msg("Database migrations applied successfully")

	// Connect to Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse Redis URL")
	}
	rdb := redis.NewClient(opt)
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Unable to ping Redis")
	}
	log.Info().Msg("Connected to Redis")

	// Initialize DB queries and session manager
	queries := db.New(dbPool)
	sessionMgr := auth.NewSessionManager(rdb)

	env := os.Getenv("ENV")
	isProd := env == "production" || env == "prod"

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if isProd {
			log.Fatal().Msg("JWT_SECRET environment variable is required in production environment")
		}
		jwtSecret = "development_only_vaultify_super_secret_key_change_me_in_production"
		log.Warn().Msg("JWT_SECRET not set, using default development secret key")
	} else if isProd && jwtSecret == "development_only_vaultify_super_secret_key_change_me_in_production" {
		log.Fatal().Msg("JWT_SECRET cannot be the default development key in production environment")
	}

	// Initialize Domain Handlers
	authHandlers := auth.NewHandlers(queries, sessionMgr, []byte(jwtSecret))
	projectHandlers := projects.NewHandlers(queries)
	authMiddleware := middleware.NewAuthMiddleware(queries, []byte(jwtSecret))

	auditSvc := audit.NewAuditService(queries)

	masterKeyStr := os.Getenv("MASTER_KEY")
	if masterKeyStr == "" {
		if isProd {
			log.Fatal().Msg("MASTER_KEY environment variable is required in production environment")
		}
		log.Warn().Msg("MASTER_KEY not set, using default development master key")
		masterKeyStr = "development_only_master_key_1234"
	} else if isProd && masterKeyStr == "development_only_master_key_1234" {
		log.Fatal().Msg("MASTER_KEY cannot be the default development key in production environment")
	}

	var masterKey []byte
	// If the MASTER_KEY is a 64-character hex string, decode it to its 32-byte representation.
	if len(masterKeyStr) == 64 {
		decoded, err := hex.DecodeString(masterKeyStr)
		if err == nil {
			masterKey = decoded
		}
	}

	// If not parsed as hex, use the raw string bytes.
	if len(masterKey) == 0 {
		masterKey = []byte(masterKeyStr)
	}

	// Enforce 32-byte key length constraint
	if len(masterKey) != 32 {
		if isProd {
			log.Fatal().Msgf("MASTER_KEY must be exactly 32 bytes (or 64 hex characters) in production, got %d bytes", len(masterKey))
		}
		// In development/testing, pad or truncate to 32 bytes for backwards compatibility
		if len(masterKey) > 32 {
			masterKey = masterKey[:32]
		} else {
			padded := make([]byte, 32)
			copy(padded, masterKey)
			masterKey = padded
		}
	}
	secretsSvc := secrets.NewSecretService(queries, auditSvc, masterKey)

	auditHandlers := audit.NewHandlers(auditSvc)
	secretsHandlers := secrets.NewHandlers(secretsSvc)
	tokenHandlers := tokens.NewHandlers(queries)

	// Setup Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip := req.Header.Get("Fly-Client-IP")
			if ip == "" {
				ip = req.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = req.RemoteAddr
			}
			ctx := context.WithValue(req.Context(), ctxkey.IPAddress, ip)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Register API Routes under /api/v1
	r.Route("/api/v1", func(r chi.Router) {
		authHandlers.RegisterRoutes(r, authMiddleware.Authenticator)
		projectHandlers.RegisterRoutes(r, authMiddleware)
		auditHandlers.RegisterRoutes(r, authMiddleware)
		secretsHandlers.RegisterRoutes(r, authMiddleware)
		tokenHandlers.RegisterRoutes(r, authMiddleware.Authenticator)
	})

	// Setup static file serving with React Router fallback
	subFS, err := fs.Sub(frontend.DistFS, "dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get sub fs for frontend")
	}
	fileServer := http.FileServer(http.FS(subFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Check if the requested file exists in the embedded FS
		f, err := subFS.Open(path)
		if err != nil {
			// React Router fallback: serve index.html directly
			indexHtml, err := fs.ReadFile(subFS, "index.html")
			if err != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(indexHtml)
			return
		}
		_ = f.Close()

		// Let the standard file server handle existing files (and / directory root)
		fileServer.ServeHTTP(w, r)
	})

	// Setup HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown setup
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Info().Msg("Received interrupt signal, shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("HTTP server Shutdown Error")
		}
		close(idleConnsClosed)
	}()

	log.Info().Msg("Starting HTTP server on port 8080")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server ListenAndServe Error")
	}

	<-idleConnsClosed
	log.Info().Msg("Server stopped")
}
