package health

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeropass/zeropass/core/vault/types"
)

// knownBreachedPassword is "password" — a famously breached password.
// SHA-1: 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
const knownBreachedPassword = "password"

// sha1Hex returns the uppercase SHA-1 hex digest.
func sha1Hex(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%X", h.Sum(nil))
}

// newMockHIBPServer creates an httptest server that simulates the HIBP range API.
// It records which prefix was requested and returns the given suffixes.
func newMockHIBPServer(t *testing.T, suffixResponses map[string]string) (*httptest.Server, *atomic.Value) {
	t.Helper()

	var lastPrefix atomic.Value

	mux := http.NewServeMux()
	mux.HandleFunc("/range/", func(w http.ResponseWriter, r *http.Request) {
		prefix := strings.TrimPrefix(r.URL.Path, "/range/")
		prefix = strings.ToUpper(prefix)
		lastPrefix.Store(prefix)

		body, ok := suffixResponses[prefix]
		if !ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "0000000000000000000000000000000000000:0")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv, &lastPrefix
}

func TestCheckPasswordBreached(t *testing.T) {
	// "password" → SHA1 = 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
	hash := sha1Hex(knownBreachedPassword)
	prefix := hash[:5] // 5BAA6
	suffix := hash[5:] // 1E4C9B93F3F0682250B6CF8331B7EE68FD8

	responseBody := fmt.Sprintf(
		"0000000000000000000000000000000000000:3\r\n"+
			"%s:3861493\r\n"+
			"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:1\r\n",
		suffix,
	)

	srv, lastPrefix := newMockHIBPServer(t, map[string]string{
		prefix: responseBody,
	})

	client := NewHIBPClient(WithBaseURL(srv.URL))

	breached, count, err := client.CheckPassword(knownBreachedPassword)
	require.NoError(t, err)
	assert.True(t, breached, "expected 'password' to be breached")
	assert.Equal(t, 3861493, count)

	// Verify k-anonymity: only prefix was sent.
	sentPrefix, ok := lastPrefix.Load().(string)
	require.True(t, ok)
	assert.Equal(t, prefix, sentPrefix, "should only send first 5 chars of hash")
	assert.Len(t, sentPrefix, hashPrefixLength, "prefix must be exactly 5 chars")
}

func TestCheckPasswordNotBreached(t *testing.T) {
	// Use a password whose suffix won't appear in the mock response.
	safePassword := "ThisIsAVeryUniquePassword!@#$%^2024XYZ"
	hash := sha1Hex(safePassword)
	prefix := hash[:5]

	// Response with unrelated suffixes.
	responseBody := "0000000000000000000000000000000000000:5\r\nFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF:2\r\n"

	srv, _ := newMockHIBPServer(t, map[string]string{
		prefix: responseBody,
	})

	client := NewHIBPClient(WithBaseURL(srv.URL))

	breached, count, err := client.CheckPassword(safePassword)
	require.NoError(t, err)
	assert.False(t, breached, "expected unique password to not be breached")
	assert.Equal(t, 0, count)
}

func TestKAnonymityPrefixOnly(t *testing.T) {
	// Verify that the full hash is NEVER sent to the server — only the 5-char prefix.
	testPassword := "MySecretTestPassword123!"
	hash := sha1Hex(testPassword)
	prefix := hash[:5]
	suffix := hash[5:]

	var receivedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s:1\r\n", suffix)
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))
	_, _, err := client.CheckPassword(testPassword)
	require.NoError(t, err)

	// The path must be /range/<prefix> — only 5 chars.
	expectedPath := "/range/" + prefix
	assert.Equal(t, expectedPath, receivedPath)

	// The suffix must NOT appear in the path.
	assert.NotContains(t, receivedPath, suffix,
		"full hash suffix must never be sent to the API")
}

func TestCheckPasswordNetworkError(t *testing.T) {
	// Point client at a closed server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // close immediately

	client := NewHIBPClient(WithBaseURL(srv.URL))

	_, _, err := client.CheckPassword("anything")
	assert.Error(t, err, "expected error for unreachable server")
	assert.Contains(t, err.Error(), "hibp request failed")
}

func TestCheckPasswordServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))

	_, _, err := client.CheckPassword("anything")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 429")
}

func TestRateLimiting(t *testing.T) {
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "0000000000000000000000000000000000000:0")
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))

	// Make 3 requests and verify they respect rate limiting.
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, _, err := client.CheckPassword(fmt.Sprintf("pw%d", i))
		require.NoError(t, err)
	}
	elapsed := time.Since(start)

	// 3 requests with 1.5s interval between them → minimum ~3s elapsed
	// (first fires immediately, then 1.5s wait, then 1.5s wait).
	assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(2900),
		"rate limiter should enforce ~1.5s intervals")
	assert.Equal(t, int32(3), atomic.LoadInt32(&requestCount))
}

func TestCheckBatch(t *testing.T) {
	hash1 := sha1Hex("password1")
	prefix1 := hash1[:5]
	suffix1 := hash1[5:]

	hash2 := sha1Hex("secureP@ss99")
	prefix2 := hash2[:5]

	responses := map[string]string{
		prefix1: fmt.Sprintf("%s:2420242\r\n0000000000000000000000000000000000000:1\r\n", suffix1),
		prefix2: "0000000000000000000000000000000000000:1\r\n",
	}

	srv, _ := newMockHIBPServer(t, responses)
	client := NewHIBPClient(WithBaseURL(srv.URL))

	items := []*batchItem{
		{ItemID: "id-1", ItemName: "Email", Password: "password1"},
		{ItemID: "id-2", ItemName: "Bank", Password: "secureP@ss99"},
	}

	results, err := client.CheckBatch(items)
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First password is breached.
	assert.Equal(t, "id-1", results[0].ItemID)
	assert.Equal(t, "Email", results[0].ItemName)
	assert.True(t, results[0].Breached)
	assert.Equal(t, 2420242, results[0].OccurrenceCount)

	// Second password is not breached.
	assert.Equal(t, "id-2", results[1].ItemID)
	assert.Equal(t, "Bank", results[1].ItemName)
	assert.False(t, results[1].Breached)
	assert.Equal(t, 0, results[1].OccurrenceCount)
}

func TestCheckBatchNetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))

	items := []*batchItem{
		{ItemID: "id-1", ItemName: "Test", Password: "pw"},
	}

	_, err := client.CheckBatch(items)
	assert.Error(t, err)
}

func TestBuildBatchItems(t *testing.T) {
	passwords := map[string]string{
		"id-1": "pass1",
		"id-2": "pass2",
	}
	names := map[string]string{
		"id-1": "Site A",
		"id-2": "Site B",
	}

	items := BuildBatchItems(passwords, names)
	assert.Len(t, items, 2)

	// Verify all items are present (order is non-deterministic with maps).
	found := make(map[string]bool)
	for _, item := range items {
		found[item.ItemID] = true
		assert.Equal(t, names[item.ItemID], item.ItemName)
		assert.Equal(t, passwords[item.ItemID], item.Password)
	}
	assert.True(t, found["id-1"])
	assert.True(t, found["id-2"])
}

func TestParseRangeResponseMalformed(t *testing.T) {
	// Lines without colons should be skipped.
	body := strings.NewReader("MALFORMED_LINE\r\n0000000000000000000000000000000000000:5\r\n")
	breached, count, err := parseRangeResponse(body, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	require.NoError(t, err)
	assert.False(t, breached)
	assert.Equal(t, 0, count)
}

func TestParseRangeResponseEmptyBody(t *testing.T) {
	body := strings.NewReader("")
	breached, count, err := parseRangeResponse(body, "ANYTHING")
	require.NoError(t, err)
	assert.False(t, breached)
	assert.Equal(t, 0, count)
}

func TestSHA1Hash(t *testing.T) {
	// "password" → 5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8
	got := sha1Hash("password")
	assert.Equal(t, "5BAA61E4C9B93F3F0682250B6CF8331B7EE68FD8", got)
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	client := NewHIBPClient(WithHTTPClient(customClient))
	assert.Equal(t, customClient, client.httpClient)
}

func TestCheckBatchEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "0000000000000000000000000000000000000:0")
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))
	results, err := client.CheckBatch(nil)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAnalyzeWithBreachFindsBreached(t *testing.T) {
	// Setup a mock server that reports "password" as breached.
	hash := sha1Hex("password")
	prefix := hash[:5]
	suffix := hash[5:]

	responses := map[string]string{
		prefix: fmt.Sprintf("%s:9999\r\n0000000000000000000000000000000000000:1\r\n", suffix),
	}
	srv, _ := newMockHIBPServer(t, responses)
	client := NewHIBPClient(WithBaseURL(srv.URL))

	analyzer := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	items := []*types.Item{
		loginItem("BreachedSite", "password", now),
		loginItem("SafeSite", "Xk9!mNpQ2#vL8@wR", now),
	}

	report := analyzer.AnalyzeWithBreach(items, client)

	assert.Equal(t, 1, report.BreachedCount)
	require.Len(t, report.BreachedPasswords, 1)
	assert.Equal(t, "id-BreachedSite", report.BreachedPasswords[0].ItemID)
	assert.Equal(t, 9999, report.BreachedPasswords[0].OccurrenceCount)
	assert.Contains(t, report.BreachedPasswords[0].Recommendation, "Change it immediately")

	// Verify breach finding in Findings list.
	var breachFindings []Finding
	for _, f := range report.Findings {
		if f.Category == "breached" {
			breachFindings = append(breachFindings, f)
		}
	}
	require.Len(t, breachFindings, 1)
	assert.Equal(t, SeverityCritical, breachFindings[0].Severity)
	assert.Contains(t, breachFindings[0].Message, "9999")

	// Score should be lower due to breached password.
	assert.Less(t, report.OverallScore, 100)
}

func TestAnalyzeWithBreachNilClient(t *testing.T) {
	analyzer := NewAnalyzer(DefaultAnalyzerConfig())
	items := []*types.Item{
		loginItem("Site", "Xk9!mNpQ2#vL8@wR", time.Now()),
	}
	report := analyzer.AnalyzeWithBreach(items, nil)

	// Should behave like normal Analyze.
	assert.Equal(t, 0, report.BreachedCount)
	assert.Nil(t, report.BreachedPasswords)
}

func TestAnalyzeWithBreachSkipsEmptyPasswords(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "0000000000000000000000000000000000000:0")
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))
	analyzer := NewAnalyzer(DefaultAnalyzerConfig())
	items := []*types.Item{
		loginItem("NoPass", "", time.Now()),
	}

	report := analyzer.AnalyzeWithBreach(items, client)
	assert.Equal(t, 0, report.BreachedCount)
}

func TestAnalyzeWithBreachSkipsNonLoginItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make HIBP request for non-login items")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewHIBPClient(WithBaseURL(srv.URL))
	analyzer := NewAnalyzer(DefaultAnalyzerConfig())
	items := []*types.Item{
		{
			ID:   "note-1",
			Type: types.ItemTypeSecureNote,
			Name: "My Note",
			Fields: map[string]string{
				"content": "secret",
			},
		},
	}

	report := analyzer.AnalyzeWithBreach(items, client)
	assert.Equal(t, 0, report.BreachedCount)
}

func TestAnalyzeWithBreachScoreReducedByBreach(t *testing.T) {
	// All strong passwords, but one is breached — score should drop.
	hash := sha1Hex("Xk9!mNpQ2#vL8@wR")
	prefix := hash[:5]
	suffix := hash[5:]

	responses := map[string]string{
		prefix: fmt.Sprintf("%s:100\r\n", suffix),
	}
	srv, _ := newMockHIBPServer(t, responses)
	client := NewHIBPClient(WithBaseURL(srv.URL))

	analyzer := NewAnalyzer(DefaultAnalyzerConfig())
	now := time.Now()
	items := []*types.Item{
		loginItem("Breached", "Xk9!mNpQ2#vL8@wR", now),
		loginItem("Safe1", "Yz7$bTfH4&jS6*cE", now),
		loginItem("Safe2", "Qw3#rTyU8!pAs0dF", now),
	}

	report := analyzer.AnalyzeWithBreach(items, client)

	// Without breach: score = 100 (all strong, not reused, not old).
	// With 1 breached (counts as 2 issues) out of 3 items → penalty = 2/3 * 100 ≈ 66.
	assert.Equal(t, 1, report.BreachedCount)
	assert.Less(t, report.OverallScore, 100)
	assert.Greater(t, report.OverallScore, 0)
}
