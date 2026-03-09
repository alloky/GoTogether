package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/getsentry/sentry-go"
	"github.com/gotogether/backend/internal/config"
	"github.com/gotogether/backend/internal/handler"
	"github.com/gotogether/backend/internal/repository/postgres"
	"github.com/gotogether/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	// Initialize OpenTelemetry tracing
	shutdownTracer := handler.InitTracer(ctx, cfg.OTELEndpoint)
	defer shutdownTracer(ctx)

	// Initialize Sentry (no-op if DSN is empty)
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			TracesSampleRate: 1.0,
			Environment:      "production",
		}); err != nil {
			log.Printf("WARN: Sentry init failed: %v", err)
		} else {
			log.Println("Sentry initialized")
			defer sentry.Flush(5 * time.Second)
		}
	}

	// Connect to database with OTel tracing
	pgxCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}
	pgxCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to database")

	// Run migrations
	if err := postgres.RunMigrations(cfg.DatabaseURL, cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed")

	// Initialize repositories
	userRepo := postgres.NewUserRepo(pool)
	meetingRepo := postgres.NewMeetingRepo(pool)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	meetingService := service.NewMeetingService(meetingRepo, userRepo)

	// Initialize router
	router := handler.NewRouter(authService, meetingService, cfg.CORSOrigin, pool)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
