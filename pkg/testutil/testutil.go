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
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestServer represents a test server instance
type TestServer struct {
	Server *httptest.Server
	Router http.Handler
	DB     *gorm.DB
	t      *testing.T
}

// TestConfig holds test configuration
type TestConfig struct {
	UseInMemoryDB bool
	EnableLogging bool
}

// NewTestServer creates a new test server with provided router
func NewTestServer(t *testing.T, router http.Handler, cfg *TestConfig) *TestServer {
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

// Close closes the test server and cleans up resources
func (ts *TestServer) Close() {
	ts.Server.Close()
	if ts.DB != nil {
		sqlDB, _ := ts.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// URL returns the base URL of the test server
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// setupTestDB sets up a test database
func setupTestDB(t *testing.T, useInMemory bool) *gorm.DB {
	var db *gorm.DB
	var err error

	if useInMemory {
		// Use in-memory SQLite for tests
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err, "Failed to connect to in-memory database")
	} else {
		// Use temporary file database
		tempDB := fmt.Sprintf("test_%s.db", uuid.New().String())
		db, err = gorm.Open(sqlite.Open(tempDB), &gorm.Config{})
		require.NoError(t, err, "Failed to connect to temp database")

		// Clean up temp file after test
		t.Cleanup(func() {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
			os.Remove(tempDB)
		})
	}

	// Set the global DB connection for handlers to use
	storage.DBConn = db

	// Run migrations
	err = db.AutoMigrate(&models.Paste{})
	require.NoError(t, err, "Failed to run migrations")

	return db
}

// setupTestConfig sets up test configuration
func setupTestConfig(enableLogging bool) {
	config.Conf = config.Config{
		WebappPort:     "3000",
		DBMaxIdleConns: 5,
		DBMaxOpenConns: 10,
		DBPort:         5432,
		DBHost:         "localhost",
		DBUser:         "test",
		DBName:         "test",
		LogLevel:       "INFO",
		LocalDB:        true,
		Dev:            true,
	}

	if !enableLogging {
		config.Conf.LogLevel = "ERROR" // Reduce log noise in tests
	}
}

// CreateTestPaste creates a test paste in the database
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

// HTTPRequest represents an HTTP request for testing
type HTTPRequest struct {
	Method      string
	Path        string
	Headers     map[string]string
	Body        interface{}
	QueryParams map[string]string
	FormData    map[string]string
}

// HTTPResponse represents an HTTP response for testing
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	JSON       map[string]interface{}
}

// MakeRequest makes an HTTP request to the test server
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
	httpReq, err := http.NewRequest(req.Method, ts.URL()+req.Path, bodyReader)
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

	// Make request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(ts.t, err, "Failed to make HTTP request")
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(ts.t, err, "Failed to read response body")

	response := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	// Try to parse JSON
	if len(body) > 0 && strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		var jsonBody map[string]interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			response.JSON = jsonBody
		}
	}

	return response
}

// AssertJSONResponse asserts that the response contains expected JSON
func (ts *TestServer) AssertJSONResponse(resp *HTTPResponse, expectedStatus int, expectedFields map[string]interface{}) {
	require.Equal(ts.t, expectedStatus, resp.StatusCode, "Unexpected status code")
	require.NotNil(ts.t, resp.JSON, "Response is not JSON")

	for key, expectedValue := range expectedFields {
		actualValue, exists := resp.JSON[key]
		require.True(ts.t, exists, "Missing field: %s", key)
		require.Equal(ts.t, expectedValue, actualValue, "Unexpected value for field: %s", key)
	}
}

// AssertError asserts that the response is an error with expected message
func (ts *TestServer) AssertError(resp *HTTPResponse, expectedStatus int, expectedError string) {
	require.Equal(ts.t, expectedStatus, resp.StatusCode, "Unexpected status code")
	require.NotNil(ts.t, resp.JSON, "Response is not JSON")

	errorMsg, exists := resp.JSON["error"]
	require.True(ts.t, exists, "Missing error field")
	require.Equal(ts.t, expectedError, errorMsg, "Unexpected error message")
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, fmt.Sprintf("Condition not met within timeout: %s", message))
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// GetPasteFromDB retrieves a paste from the database by UUID
func (ts *TestServer) GetPasteFromDB(pasteUUID uuid.UUID) *models.Paste {
	var paste models.Paste
	err := ts.DB.First(&paste, "uuid = ?", pasteUUID).Error
	if err != nil {
		return nil
	}
	return &paste
}

// CountPastesInDB returns the number of pastes in the database
func (ts *TestServer) CountPastesInDB() int64 {
	var count int64
	ts.DB.Model(&models.Paste{}).Count(&count)
	return count
}

// CleanupPastes removes all pastes from the database
func (ts *TestServer) CleanupPastes() {
	ts.DB.Exec("DELETE FROM pastes")
}

// CreateExpiredPaste creates an expired paste for testing
func (ts *TestServer) CreateExpiredPaste() *models.Paste {
	paste := &models.Paste{
		UUID:            uuid.New(),
		Content:         "This paste is expired",
		Language:        "txt",
		Burn:            false,
		ExpiryTimestamp: time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	err := ts.DB.Create(paste).Error
	require.NoError(ts.t, err, "Failed to create expired paste")

	return paste
}

// MockTimeNow can be used to mock time.Now() in tests
type MockTime struct {
	current time.Time
}

func NewMockTime(t time.Time) *MockTime {
	return &MockTime{current: t}
}

func (m *MockTime) Now() time.Time {
	return m.current
}

func (m *MockTime) Add(d time.Duration) {
	m.current = m.current.Add(d)
}

// TestData contains commonly used test data
var TestData = struct {
	ValidPasteContent   string
	ValidLanguage       string
	ValidExpiryMinutes  int
	LargePasteContent   string
	InvalidExpiryString string
}{
	ValidPasteContent:   "Hello, World! This is a test paste.",
	ValidLanguage:       "txt",
	ValidExpiryMinutes:  60,
	LargePasteContent:   strings.Repeat("A", 11*1024*1024), // 11MB (exceeds 10MB limit)
	InvalidExpiryString: "invalid-expiry",
}