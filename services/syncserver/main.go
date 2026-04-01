// Package main provides the ZeroPass sync server entry point.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	syncserver "github.com/zeropass/zeropass/core/sync/server"
)

func main() {
	port := flag.Int("port", 8443, "HTTP listen port")
	dbPath := flag.String("db-path", "zeropass-sync.db", "SQLite database file path")
	apiKey := flag.String("api-key", "", "API key for Bearer auth (empty = no auth)")
	flag.Parse()

	if err := run(*port, *dbPath, *apiKey); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// run initializes storage and starts the server with graceful shutdown.
func run(port int, dbPath, apiKey string) error {
	store, err := NewSQLiteStorage(dbPath)
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	defer store.Close()

	if err := store.CreateTables(); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	srv := NewServer(port, store, apiKey)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("ZeroPass sync server listening on :%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down...")
	case err := <-errCh:
		return err
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.Println("Server stopped")
	return nil
}

// NewServer creates a configured *http.Server.
func NewServer(port int, store *SQLiteStorage, apiKey string) *http.Server {
	mux := http.NewServeMux()
	handler := syncserver.NewHandler(store, apiKey)
	handler.RegisterRoutes(mux)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
