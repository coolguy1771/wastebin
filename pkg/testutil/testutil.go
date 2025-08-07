package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
)

// testDBMutex protects concurrent access to the global storage.DBConn during tests.
//
//nolint:gochecknoglobals // Required for test synchronization
var testDBMutex sync.Mutex

const (
	// HTTP client timeout for tests.
	httpClientTimeout = 10 * time.Second

	// HTTP client connection pool settings.
	maxIdleConns        = 100
	maxIdleConnsPerHost = 100
	idleConnTimeout     = 90 * time.Second

	// Wait condition check interval.
	waitCheckInterval = 10 * time.Millisecond

	// Test data constants.
	validExpiryMinutes = 60

	OneMiB           = 1 << 20
	largeContentSize = 11 * OneMiB // 11MB (exceeds 10MB limit)
	DefaultMaxSize   = 10 * OneMiB // 10MB default for tests

	expiredTimeOffset = -1 * time.Hour

	// Default database configuration constants.
	defaultDBMaxIdleConns = 5
	defaultDBMaxOpenConns = 10
	defaultDBPort         = 5432
)

// Shared HTTP client for tests with connection pooling
//
//nolint:gochecknoglobals // Shared client for performance
var testHTTPClient = &http.Client{
	Timeout: httpClientTimeout,
	Transport: &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
	},
}

// TestServer represents a test server instance.
type TestServer struct {
	Server *httptest.Server
	Router http.Handler
	DB     *gorm.DB
	t      *testing.T
}

// TestConfig holds test configuration.
type TestConfig struct {
	UseInMemoryDB bool
	EnableLogging bool
}

// NewTestServer creates a new test server with provided router.
func NewTestServer(t *testing.T, router http.Handler, cfg *TestConfig) *TestServer {
	t.Helper()

	if cfg == nil {
		cfg = &TestConfig{
			UseInMemoryDB: true,
			EnableLogging: false,
		}
	}

	// Setup test database
	db := setupTestDB(t, cfg.UseInMemoryDB)

	// Setup test configuration
	setupTestConfig(cfg.EnableLogging)

	// Create test server
	server := httptest.NewServer(router)

	return &TestServer{
		Server: server,
		Router: router,
		DB:     db,
		t:      t,
	}
}

// Close closes the test server and cleans up resources.
func (ts *TestServer) Close() {
	// Close the test server first
	if ts.Server != nil {
		ts.Server.Close()
	}

	// Then close the database
	if ts.DB != nil {
		sqlDB, _ := ts.DB.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}
}

// URL returns the base URL of the test server.
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// setupTestDB sets up a test database.
func setupTestDB(t *testing.T, useInMemory bool) *gorm.DB {
	t.Helper()

	// Lock to prevent concurrent modification of global storage.DBConn
	testDBMutex.Lock()
	defer testDBMutex.Unlock()

	var (
		db  *gorm.DB
		err error
	)

	if useInMemory {
		// Use a unique file-based SQLite database for each test to avoid conflicts
		// In-memory databases don't work well with global storage.DBConn in parallel tests
		tempDB := fmt.Sprintf("test_mem_%s.db", uuid.New().String())

		// Use Zap logger adapter for GORM in tests
		gormLogger := storage.NewGormZapLogger(log.Default())

		db, err = gorm.Open(sqlite.Open(tempDB), &gorm.Config{
			Logger:                                   gormLogger,
			SkipDefaultTransaction:                   false,
			DisableAutomaticPing:                     false,
			DisableForeignKeyConstraintWhenMigrating: false,
		})
		require.NoError(t, err, "Failed to connect to test database")

		// Clean up temp file after test
		t.Cleanup(func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
			_ = os.Remove(tempDB)
		})
	} else {
		// Use temporary file database
		tempDB := fmt.Sprintf("test_%s.db", uuid.New().String())

		// Use Zap logger adapter for GORM in tests
		gormLogger := storage.NewGormZapLogger(log.Default())

		db, err = gorm.Open(sqlite.Open(tempDB), &gorm.Config{
			Logger:                                   gormLogger,
			SkipDefaultTransaction:                   false,
			DisableAutomaticPing:                     false,
			DisableForeignKeyConstraintWhenMigrating: false,
		})
		require.NoError(t, err, "Failed to connect to temp database")

		// Clean up temp file after test
		t.Cleanup(func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}

			_ = os.Remove(tempDB)
		})
	}

	// Set the global DB connection for handlers to use
	//nolint:reassign // DBConn is reassigned for test setup.
	storage.DBConn = db

	// Run migrations
	err = db.AutoMigrate(&models.Paste{})
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// setupTestConfig sets up test configuration.
func setupTestConfig(enableLogging bool) {
	//nolint:reassign // Conf is reassigned for test setup.
	config.Conf = config.Config{
		WebappPort:          "3000",
		DBMaxIdleConns:      defaultDBMaxIdleConns,
		DBMaxOpenConns:      defaultDBMaxOpenConns,
		DBPort:              defaultDBPort,
		DBHost:              "localhost",
		DBUser:              "test",
		DBPassword:          "",
		DBName:              "test",
		Dev:                 true,
		TLSEnabled:          false,
		TLSCertFile:         "",
		TLSKeyFile:          "",
		AllowedOrigins:      "",
		RequireAuth:         false,
		AuthUsername:        "",
		AuthPassword:        "",
		CSRFKey:             "",
		MaxRequestSize:      DefaultMaxSize, // 10MB default for tests
		LogLevel:            "INFO",
		LocalDB:             true,
		TracingEnabled:      false,
		MetricsEnabled:      false,
		ServiceName:         "",
		ServiceVersion:      "",
		Environment:         "",
		OTLPTraceEndpoint:   "",
		OTLPMetricsEndpoint: "",
		MetricsInterval:     0,
	}

	if !enableLogging {
		config.Conf.LogLevel = "ERROR" // Reduce log noise in tests
	}
}

// CreateTestPaste creates a test paste in the database.
func (ts *TestServer) CreateTestPaste(content, language string, expiryMinutes int, burn bool) *models.Paste {
	paste := &models.Paste{
		UUID:            uuid.New(),
		Content:         content,
		Language:        language,
		Burn:            burn,
		ExpiryTimestamp: time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
	}

	err := ts.DB.Create(paste).Error
	require.NoError(ts.t, err, "Failed to create test paste")

	return paste
}

// HTTPRequest represents an HTTP request for testing.
type HTTPRequest struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        interface{}
	QueryParams map[string]string
	FormData    map[string]string
}

// HTTPResponse represents an HTTP response for testing.
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	JSON       map[string]interface{}
}

// MakeRequest makes an HTTP request to the test server.
func (ts *TestServer) MakeRequest(req HTTPRequest) *HTTPResponse {
	var bodyReader io.Reader

	// Handle different body types
	if req.Body != nil {
		switch body := req.Body.(type) {
		case string:
			bodyReader = strings.NewReader(body)
		case []byte:
			bodyReader = bytes.NewReader(body)
		case map[string]interface{}:
			jsonBody, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonBody)
		default:
			jsonBody, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(jsonBody)
		}
	}

	// Handle form data
	if req.FormData != nil {
		formValues := url.Values{}
		for key, value := range req.FormData {
			formValues.Set(key, value)
		}

		bodyReader = strings.NewReader(formValues.Encode())

		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}

		req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(context.Background(), req.Method, ts.URL()+req.Path, bodyReader)
	require.NoError(ts.t, err, "Failed to create HTTP request")

	// Add headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Add query parameters
	if req.QueryParams != nil {
		q := httpReq.URL.Query()
		for key, value := range req.QueryParams {
			q.Add(key, value)
		}

		httpReq.URL.RawQuery = q.Encode()
	}

	// Make request using shared client
	resp, err := testHTTPClient.Do(httpReq)
	require.NoError(ts.t, err, "Failed to make HTTP request")

	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(ts.t, err, "Failed to read response body")

	response := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		JSON:       nil,
	}

	// Try to parse JSON
	if len(body) > 0 && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonBody map[string]interface{}

		jsonErr := json.Unmarshal(body, &jsonBody)
		if jsonErr == nil {
			response.JSON = jsonBody
		}
	}

	return response
}

// AssertJSONResponse asserts that the response contains expected JSON.
func (ts *TestServer) AssertJSONResponse(
	resp *HTTPResponse,
	expectedStatus int,
	expectedFields map[string]interface{},
) {
	require.Equal(ts.t, expectedStatus, resp.StatusCode, "Unexpected status code")
	require.NotNil(ts.t, resp.JSON, "Response is not JSON")

	for key, expectedValue := range expectedFields {
		actualValue, exists := resp.JSON[key]
		require.True(ts.t, exists, "Missing field: %s", key)
		require.Equal(ts.t, expectedValue, actualValue, "Unexpected value for field: %s", key)
	}
}

// AssertError asserts that the response is an error with expected message.
func (ts *TestServer) AssertError(resp *HTTPResponse, expectedStatus int, expectedError string) {
	require.Equal(ts.t, expectedStatus, resp.StatusCode, "Unexpected status code")
	require.NotNil(ts.t, resp.JSON, "Response is not JSON")

	errorMsg, exists := resp.JSON["error"]
	require.True(ts.t, exists, "Missing error field")
	require.Equal(ts.t, expectedError, errorMsg, "Unexpected error message")
}

// WaitForCondition waits for a condition to be true with timeout.
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(waitCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "Condition not met within timeout: "+message)
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// GetPasteFromDB retrieves a paste from the database by UUID.
func (ts *TestServer) GetPasteFromDB(pasteUUID uuid.UUID) *models.Paste {
	var paste models.Paste

	err := ts.DB.First(&paste, "uuid = ?", pasteUUID).Error
	if err != nil {
		return nil
	}

	return &paste
}

// CountPastesInDB returns the number of pastes in the database.
func (ts *TestServer) CountPastesInDB() int64 {
	var count int64

	ts.DB.Model(&models.Paste{}).Count(&count)

	return count
}

// CleanupPastes removes all pastes from the database.
func (ts *TestServer) CleanupPastes() {
	ts.DB.Exec("DELETE FROM pastes")
}

// CreateExpiredPaste creates an expired paste for testing.
func (ts *TestServer) CreateExpiredPaste() *models.Paste {
	paste := &models.Paste{
		UUID:            uuid.New(),
		Content:         "This paste is expired",
		Burn:            false,
		Language:        "txt",
		ExpiryTimestamp: time.Now().Add(expiredTimeOffset), // 1 hour ago
	}

	err := ts.DB.Create(paste).Error
	require.NoError(ts.t, err, "Failed to create expired paste")

	return paste
}

// MockTime can be used to mock time.Now() in tests.
type MockTime struct {
	current time.Time
}

// NewMockTime creates a new mock time instance.
func NewMockTime(t time.Time) *MockTime {
	return &MockTime{current: t}
}

// Now returns the current mock time.
func (m *MockTime) Now() time.Time {
	return m.current
}

// Add advances the mock time by the given duration.
func (m *MockTime) Add(d time.Duration) {
	m.current = m.current.Add(d)
}

// TestData contains commonly used test data.
func TestData() struct {
	ValidPasteContent   string
	ValidLanguage       string
	ValidExpiryMinutes  int
	LargePasteContent   string
	InvalidExpiryString string
} {
	return struct {
		ValidPasteContent   string
		ValidLanguage       string
		ValidExpiryMinutes  int
		LargePasteContent   string
		InvalidExpiryString string
	}{
		ValidPasteContent:   "Hello, World! This is a test paste.",
		ValidLanguage:       "txt",
		ValidExpiryMinutes:  validExpiryMinutes,
		LargePasteContent:   strings.Repeat("A", largeContentSize), // 11MB (exceeds 10MB limit)
		InvalidExpiryString: "invalid-expiry",
	}
}
