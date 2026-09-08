package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"github.com/futomaru/todo-app-go/internal/clock"
	"github.com/futomaru/todo-app-go/internal/handler"
	"github.com/futomaru/todo-app-go/internal/middleware"
	"github.com/futomaru/todo-app-go/internal/repository"
	"github.com/futomaru/todo-app-go/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	addr := envOr("PORT", ":8080")
	dbPath := envOr("DB_PATH", "./data/todo.db")

	db, err := openDB(dbPath)
	if err != nil {
		logger.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.Migrate(context.Background(), db); err != nil {
		logger.Error("failed to migrate database", "err", err)
		os.Exit(1)
	}

	root := newRouter(db, logger)

	srv := &http.Server{
		Addr:              addr,
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
	}
	runWithGracefulShutdown(srv, logger)
}

// newRouter wires all layers bottom-up (Composition Root) and returns the
// fully chained handler. Split out from main so integration tests (STEP 11)
// can build the same stack against a test database.
func newRouter(db *sql.DB, logger *slog.Logger) http.Handler {
	repo := repository.New(db)
	svc := service.New(repo, clock.System())
	h := handler.New(svc)

	return middleware.Chain(h.Routes(),
		middleware.Recoverer(logger),
		middleware.RequestLogger(logger),
	)
}

// openDB creates the DB directory if needed and opens a SQLite connection
// tuned for a single-writer, file-backed workload: WAL improves read/write
// concurrency, busy_timeout absorbs brief lock contention instead of
// failing immediately, and capping the pool at 1 connection serializes
// writes so concurrent requests don't trip "database is locked".
func openDB(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	dsn := dbPath + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// runWithGracefulShutdown starts srv and blocks until SIGINT/SIGTERM,
// then gives in-flight requests up to 10s to finish before returning.
func runWithGracefulShutdown(srv *http.Server, logger *slog.Logger) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
