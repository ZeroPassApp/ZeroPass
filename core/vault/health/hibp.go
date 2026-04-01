// Package health provides password health analysis for vault items.
package health

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// defaultHIBPBaseURL is the default Have I Been Pwned API base URL.
	defaultHIBPBaseURL = "https://api.pwnedpasswords.com"

	// hibpRateInterval is the minimum interval between HIBP API requests
	// to respect the free-tier rate limit.
	hibpRateInterval = 1500 * time.Millisecond

	// hashPrefixLength is the number of hex characters sent to the HIBP API
	// for k-anonymity. Only the first 5 characters of the SHA-1 hash are sent.
	hashPrefixLength = 5
)

// BreachResult holds the HIBP breach check result for a single item.
type BreachResult struct {
	ItemID          string `json:"item_id"`
	ItemName        string `json:"item_name"`
	Breached        bool   `json:"breached"`
	OccurrenceCount int    `json:"occurrence_count"`
}

// HIBPClient checks passwords against the Have I Been Pwned API
// using the k-anonymity model. Only the first 5 characters of the
// SHA-1 hash are ever sent over the network.
type HIBPClient struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.Mutex
	lastReq    time.Time
}

// HIBPOption configures an HIBPClient.
type HIBPOption func(*HIBPClient)

// WithBaseURL overrides the HIBP API base URL (useful for testing).
func WithBaseURL(url string) HIBPOption {
	return func(c *HIBPClient) {
		c.baseURL = url
	}
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(hc *http.Client) HIBPOption {
	return func(c *HIBPClient) {
		c.httpClient = hc
	}
}

// NewHIBPClient creates a new HIBP client with the given options.
func NewHIBPClient(opts ...HIBPOption) *HIBPClient {
	c := &HIBPClient{
		baseURL: defaultHIBPBaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CheckPassword checks a single password against the HIBP database.
// It returns whether the password has been breached and the occurrence count.
// Only the first 5 characters of the SHA-1 hash are sent (k-anonymity).
func (c *HIBPClient) CheckPassword(password string) (bool, int, error) {
	hash := sha1Hash(password)
	prefix := hash[:hashPrefixLength]
	suffix := hash[hashPrefixLength:]

	c.rateLimit()

	url := fmt.Sprintf("%s/range/%s", c.baseURL, prefix)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return false, 0, fmt.Errorf("hibp request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, 0, fmt.Errorf("hibp returned status %d", resp.StatusCode)
	}

	return parseRangeResponse(resp.Body, suffix)
}

// CheckBatch checks multiple passwords against the HIBP database.
// The passwords map is keyed by item ID with values of "itemName\x00password".
// Use BuildBatchInput to construct the map correctly.
func (c *HIBPClient) CheckBatch(items []*batchItem) ([]BreachResult, error) {
	results := make([]BreachResult, 0, len(items))

	for _, item := range items {
		breached, count, err := c.CheckPassword(item.Password)
		if err != nil {
			return nil, fmt.Errorf("checking item %s: %w", item.ItemID, err)
		}
		results = append(results, BreachResult{
			ItemID:          item.ItemID,
			ItemName:        item.ItemName,
			Breached:        breached,
			OccurrenceCount: count,
		})
	}

	return results, nil
}

// batchItem holds the data needed for a batch HIBP check.
type batchItem struct {
	ItemID   string
	ItemName string
	Password string
}

// BuildBatchItems constructs the input for CheckBatch from a password map.
// The map key is item ID, value is the password.
func BuildBatchItems(passwords map[string]string, names map[string]string) []*batchItem {
	items := make([]*batchItem, 0, len(passwords))
	for id, pw := range passwords {
		items = append(items, &batchItem{
			ItemID:   id,
			ItemName: names[id],
			Password: pw,
		})
	}
	return items
}

// rateLimit enforces a minimum interval between API requests.
func (c *HIBPClient) rateLimit() {
	c.mu.Lock()
	defer c.mu.Unlock()

	since := time.Since(c.lastReq)
	if since < hibpRateInterval {
		time.Sleep(hibpRateInterval - since)
	}
	c.lastReq = time.Now()
}

// sha1Hash returns the uppercase hex-encoded SHA-1 hash of the input string.
func sha1Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%X", h.Sum(nil))
}

// parseRangeResponse parses the HIBP range API response body and checks
// whether the given hash suffix appears. The response format is:
//
//	SUFFIX:COUNT\r\n
//
// where SUFFIX is the remaining characters of the SHA-1 hash (after the
// 5-char prefix) and COUNT is the number of times that password appeared
// in breaches.
func parseRangeResponse(body io.Reader, suffix string) (bool, int, error) {
	suffix = strings.ToUpper(suffix)
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		if strings.ToUpper(parts[0]) == suffix {
			count, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return false, 0, fmt.Errorf("invalid count in HIBP response: %w", err)
			}
			return true, count, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, 0, fmt.Errorf("reading HIBP response: %w", err)
	}

	return false, 0, nil
}
