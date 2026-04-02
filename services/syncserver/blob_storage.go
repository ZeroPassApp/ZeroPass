package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeropass/zeropass/core/sync/protocol"
)

const blobRefPrefix = "s3://"

// BlobStorageConfig configures S3-compatible blob storage.
type BlobStorageConfig struct {
	Endpoint  string // e.g., "localhost:9000" or "s3.amazonaws.com"
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string // default "us-east-1"
}

// BlobStorage wraps a base storage and offloads payloads to S3/MinIO.
type BlobStorage struct {
	base   StorageWithTables
	config BlobStorageConfig
	client *http.Client
}

// StorageWithTables extends the sync server Storage interface with lifecycle methods.
type StorageWithTables interface {
	SaveItem(item protocol.SyncItem) error
	GetItemsSince(since int64) ([]protocol.SyncItem, error)
	GetItem(itemID string) (*protocol.SyncItem, error)
	RegisterDevice(device protocol.DeviceInfo) error
	UpdateDeviceSync(deviceID string, timestamp int64) error
	CreateTables() error
	Close() error
}

// NewBlobStorage creates a BlobStorage that wraps base and offloads payloads to S3.
func NewBlobStorage(base StorageWithTables, config BlobStorageConfig) *BlobStorage {
	if config.Region == "" {
		config.Region = "us-east-1"
	}
	return &BlobStorage{
		base:   base,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SaveItem uploads the payload to S3 (if non-empty) and stores an s3:// reference
// in the database instead of the full payload.
func (b *BlobStorage) SaveItem(item protocol.SyncItem) error {
	if item.Payload != "" {
		key := fmt.Sprintf("items/%s/%d", item.ItemID, item.Version)
		if err := b.putObject(key, []byte(item.Payload)); err != nil {
			return fmt.Errorf("upload blob for %s: %w", item.ItemID, err)
		}
		item.Payload = fmt.Sprintf("%s%s/%s", blobRefPrefix, b.config.Bucket, key)
	}
	return b.base.SaveItem(item)
}

// GetItem retrieves an item and resolves any s3:// payload reference.
func (b *BlobStorage) GetItem(itemID string) (*protocol.SyncItem, error) {
	item, err := b.base.GetItem(itemID)
	if err != nil || item == nil {
		return item, err
	}
	if err := b.resolvePayload(item); err != nil {
		return nil, err
	}
	return item, nil
}

// GetItemsSince retrieves items and resolves any s3:// payload references.
func (b *BlobStorage) GetItemsSince(since int64) ([]protocol.SyncItem, error) {
	items, err := b.base.GetItemsSince(since)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if err := b.resolvePayload(&items[i]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// RegisterDevice delegates to the base storage.
func (b *BlobStorage) RegisterDevice(device protocol.DeviceInfo) error {
	return b.base.RegisterDevice(device)
}

// UpdateDeviceSync delegates to the base storage.
func (b *BlobStorage) UpdateDeviceSync(deviceID string, timestamp int64) error {
	return b.base.UpdateDeviceSync(deviceID, timestamp)
}

// CreateTables delegates to the base storage.
func (b *BlobStorage) CreateTables() error {
	return b.base.CreateTables()
}

// Close delegates to the base storage.
func (b *BlobStorage) Close() error {
	return b.base.Close()
}

// resolvePayload downloads the actual payload from S3 when the stored value
// is an s3:// reference.
func (b *BlobStorage) resolvePayload(item *protocol.SyncItem) error {
	if !strings.HasPrefix(item.Payload, blobRefPrefix) {
		return nil
	}
	ref := strings.TrimPrefix(item.Payload, blobRefPrefix)
	// ref format: "bucket/path/to/key"
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid blob reference: %s", item.Payload)
	}
	key := parts[1]

	data, err := b.getObject(key)
	if err != nil {
		return fmt.Errorf("download blob for %s: %w", item.ItemID, err)
	}
	item.Payload = string(data)
	return nil
}

// buildURL constructs the full HTTP URL for an S3 object.
func (b *BlobStorage) buildURL(key string) string {
	endpoint := b.config.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		scheme := "http"
		if b.config.UseSSL {
			scheme = "https"
		}
		endpoint = scheme + "://" + endpoint
	}
	return endpoint + "/" + b.config.Bucket + "/" + key
}

// putObject uploads data to the S3-compatible endpoint using HTTP PUT with basic auth.
func (b *BlobStorage) putObject(key string, data []byte) error {
	u := b.buildURL(key)
	req, err := http.NewRequest(http.MethodPut, u, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create put request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if b.config.AccessKey != "" {
		req.SetBasicAuth(b.config.AccessKey, b.config.SecretKey)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("put object status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// getObject downloads data from the S3-compatible endpoint using HTTP GET with basic auth.
func (b *BlobStorage) getObject(key string) ([]byte, error) {
	u := b.buildURL(key)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create get request: %w", err)
	}
	if b.config.AccessKey != "" {
		req.SetBasicAuth(b.config.AccessKey, b.config.SecretKey)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get object status %d: %s", resp.StatusCode, string(body))
	}
	return io.ReadAll(resp.Body)
}
