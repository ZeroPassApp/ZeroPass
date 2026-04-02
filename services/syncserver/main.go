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

// serverConfig holds all configuration for the sync server.
type serverConfig struct {
	Port          int
	DBPath        string
	APIKey        string
	StorageType   string // "sqlite" (default) or "postgres"
	PostgresURL   string
	BlobEnabled   bool
	BlobEndpoint  string
	BlobBucket    string
	BlobAccessKey string
	BlobSecretKey string
}

func main() {
	var cfg serverConfig
	flag.IntVar(&cfg.Port, "port", 8443, "HTTP listen port")
	flag.StringVar(&cfg.DBPath, "db-path", "zeropass-sync.db", "SQLite database file path")
	flag.StringVar(&cfg.APIKey, "api-key", "", "API key for Bearer auth (empty = no auth)")
	flag.StringVar(&cfg.StorageType, "storage", "sqlite", "Storage backend: sqlite or postgres")
	flag.StringVar(&cfg.PostgresURL, "postgres-url", "", "PostgreSQL connection string")
	flag.BoolVar(&cfg.BlobEnabled, "blob-enabled", false, "Enable S3/MinIO blob storage for payloads")
	flag.StringVar(&cfg.BlobEndpoint, "blob-endpoint", "", "S3/MinIO endpoint (e.g. localhost:9000)")
	flag.StringVar(&cfg.BlobBucket, "blob-bucket", "zeropass", "S3/MinIO bucket name")
	flag.StringVar(&cfg.BlobAccessKey, "blob-access-key", "", "S3/MinIO access key")
	flag.StringVar(&cfg.BlobSecretKey, "blob-secret-key", "", "S3/MinIO secret key")
	flag.Parse()

	if err := run(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// run initializes storage and starts the server with graceful shutdown.
func run(cfg serverConfig) error {
	if cfg.StorageType == "" {
		cfg.StorageType = "sqlite"
	}

	var store StorageWithTables
	var err error

	switch cfg.StorageType {
	case "sqlite":
		store, err = NewSQLiteStorage(cfg.DBPath)
	case "postgres":
		if cfg.PostgresURL == "" {
			return fmt.Errorf("--postgres-url required when --storage=postgres")
		}
		store, err = NewPostgresStorage(cfg.PostgresURL)
	default:
		return fmt.Errorf("unsupported storage type: %s", cfg.StorageType)
	}
	if err != nil {
		return fmt.Errorf("initialize storage: %w", err)
	}
	defer store.Close()

	if err := store.CreateTables(); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	if cfg.BlobEnabled {
		store = NewBlobStorage(store, BlobStorageConfig{
			Endpoint:  cfg.BlobEndpoint,
			Bucket:    cfg.BlobBucket,
			AccessKey: cfg.BlobAccessKey,
			SecretKey: cfg.BlobSecretKey,
		})
	}

	srv := NewServer(cfg.Port, store, cfg.APIKey)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("ZeroPass sync server listening on :%d", cfg.Port)
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

// NewServer creates a configured *http.Server with the given storage backend.
func NewServer(port int, store syncserver.Storage, apiKey string) *http.Server {
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
