package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/0DayMonxrch/vaultify/internal/audit"
	"github.com/0DayMonxrch/vaultify/internal/auth"
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
		dbURL = "postgres://vaultify:vaultify_password@localhost:5432/vaultify?sslmode=disable"
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
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Unable to ping Redis")
	}
	log.Info().Msg("Connected to Redis")

	// Initialize DB queries and session manager
	queries := db.New(dbPool)
	sessionMgr := auth.NewSessionManager(rdb)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "development_only_vaultify_super_secret_key_change_me_in_production"
		log.Warn().Msg("JWT_SECRET not set, using default development secret key")
	}

	// Initialize Domain Handlers
	authHandlers := auth.NewHandlers(queries, sessionMgr, []byte(jwtSecret))
	projectHandlers := projects.NewHandlers(queries)
	authMiddleware := middleware.NewAuthMiddleware(queries, []byte(jwtSecret))

	auditSvc := audit.NewAuditService(queries)
	masterKeyStr := os.Getenv("MASTER_KEY")
	if masterKeyStr == "" {
		masterKeyStr = "development_only_master_key_1234"
	}
	masterKey := []byte(masterKeyStr)
	if len(masterKey) > 32 {
		masterKey = masterKey[:32]
	} else if len(masterKey) < 32 {
		padded := make([]byte, 32)
		copy(padded, masterKey)
		masterKey = padded
	}
	secretsSvc := secrets.NewSecretService(queries, auditSvc, masterKey)

	auditHandlers := audit.NewHandlers(auditSvc)
	secretsHandlers := secrets.NewHandlers(secretsSvc)
	tokenHandlers := tokens.NewHandlers(queries)

	// Setup Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register Auth Routes
	authHandlers.RegisterRoutes(r, authMiddleware.Authenticator)

	// Register Project Routes
	projectHandlers.RegisterRoutes(r, authMiddleware)
	auditHandlers.RegisterRoutes(r, authMiddleware)
	secretsHandlers.RegisterRoutes(r, authMiddleware)
	tokenHandlers.RegisterRoutes(r, authMiddleware.Authenticator)

	// Setup HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
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
