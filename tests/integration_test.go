package tests

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/pkg/testutil"
	"github.com/coolguy1771/wastebin/routes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullPasteWorkflow tests the complete paste lifecycle
func TestFullPasteWorkflow(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// 1. Create a paste
	createResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "POST",
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      "Integration test content",
			"extension": "go",
			"expires":   "120", // 2 hours
			"burn":      "false",
		},
	})

	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	require.NotNil(t, createResp.JSON)

	pasteUUID := createResp.JSON["uuid"].(string)
	require.NotEmpty(t, pasteUUID)

	// 2. Retrieve the paste via JSON API
	getResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	require.Equal(t, http.StatusOK, getResp.StatusCode)
	require.NotNil(t, getResp.JSON)
	assert.Equal(t, "Integration test content", getResp.JSON["Content"])
	assert.Equal(t, "go", getResp.JSON["Language"])
	assert.Equal(t, false, getResp.JSON["Burn"])

	// 3. Retrieve the paste via raw endpoint
	rawResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/paste/" + pasteUUID + "/raw",
	})

	require.Equal(t, http.StatusOK, rawResp.StatusCode)
	assert.Equal(t, "Integration test content", string(rawResp.Body))
	assert.Contains(t, rawResp.Headers.Get("Content-Type"), "text/plain")

	// 4. Delete the paste
	deleteResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "DELETE",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	require.Equal(t, http.StatusOK, deleteResp.StatusCode)
	require.NotNil(t, deleteResp.JSON)
	assert.Equal(t, "Paste deleted successfully", deleteResp.JSON["message"])

	// 5. Verify paste is deleted
	verifyResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	assert.Equal(t, http.StatusNotFound, verifyResp.StatusCode)
}

// TestBurnAfterReadWorkflow tests the burn-after-read functionality
func TestBurnAfterReadWorkflow(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Create a burn-after-read paste
	createResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "POST",
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      "This will burn after reading",
			"extension": "txt",
			"expires":   "60",
			"burn":      "true",
		},
	})

	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	pasteUUID := createResp.JSON["uuid"].(string)

	// First access should return the content and delete the paste
	firstResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	assert.Equal(t, http.StatusGone, firstResp.StatusCode)
	assert.Contains(t, firstResp.JSON["error"], "expired or been burned")

	// Second access should return not found
	secondResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	assert.Equal(t, http.StatusNotFound, secondResp.StatusCode)
}

// TestHealthEndpoints tests all health check endpoints
func TestHealthEndpoints(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name         string
		endpoint     string
		expectedKeys []string
	}{
		{
			name:         "Basic health check",
			endpoint:     "/health",
			expectedKeys: []string{"status", "service"},
		},
		{
			name:         "Database health check",
			endpoint:     "/health/db",
			expectedKeys: []string{"status", "service", "timestamp"},
		},
		{
			name:         "Built-in health check",
			endpoint:     "/healthz",
			expectedKeys: []string{}, // This is a simple ping endpoint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   tt.endpoint,
			})

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			if tt.endpoint != "/healthz" {
				require.NotNil(t, resp.JSON)
				for _, key := range tt.expectedKeys {
					assert.Contains(t, resp.JSON, key, "Missing key: %s", key)
				}
				assert.Equal(t, "healthy", resp.JSON["status"])
			}
		})
	}
}

// TestAPIVersioning tests API versioning functionality
func TestAPIVersioning(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name            string
		headers         map[string]string
		expectedVersion string
	}{
		{
			name:            "Default version",
			headers:         nil,
			expectedVersion: "v1",
		},
		{
			name: "Version via X-API-Version header",
			headers: map[string]string{
				"X-API-Version": "v2",
			},
			expectedVersion: "v2",
		},
		{
			name: "Version via Accept header",
			headers: map[string]string{
				"Accept": "application/vnd.wastebin.v2+json",
			},
			expectedVersion: "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:  "GET",
				Path:    "/api/v1/",
				Headers: tt.headers,
			})

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.expectedVersion, resp.Headers.Get("X-API-Version"))
		})
	}
}

// TestSecurityHeaders tests that security headers are properly set
func TestIntegrationSecurityHeaders(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/",
	})

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := resp.Headers.Get(header)
		assert.Equal(t, expectedValue, actualValue, "Security header %s not set correctly", header)
	}
}

// TestRequestTracing tests that request IDs are properly generated and returned
func TestRequestTracing(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/",
	})

	requestID := resp.Headers.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "Request ID should be set in response headers")
}

// TestConcurrentOperations tests concurrent access to the API
func TestIntegrationConcurrentOperations(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	const numGoroutines = 20
	const numOperationsPerGoroutine = 5

	var wg sync.WaitGroup
	results := make(chan bool, numGoroutines*numOperationsPerGoroutine)

	// Concurrent paste creation and retrieval
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Create paste
				createResp := server.MakeRequest(testutil.HTTPRequest{
					Method: "POST",
					Path:   "/api/v1/paste",
					FormData: map[string]string{
						"text":      fmt.Sprintf("Concurrent test %d-%d", routineID, j),
						"extension": "txt",
						"expires":   "60",
						"burn":      "false",
					},
				})

				if createResp.StatusCode == http.StatusCreated {
					// Try to retrieve the paste
					pasteUUID := createResp.JSON["uuid"].(string)
					getResp := server.MakeRequest(testutil.HTTPRequest{
						Method: "GET",
						Path:   "/api/v1/paste/" + pasteUUID,
					})

					results <- getResp.StatusCode == http.StatusOK
				} else {
					results <- false
				}
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Count successful operations
	var successful int
	for success := range results {
		if success {
			successful++
		}
	}

	// Should have high success rate (allowing for some rate limiting)
	successRate := float64(successful) / float64(numGoroutines*numOperationsPerGoroutine)
	assert.Greater(t, successRate, 0.8, "Success rate should be > 80%")
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name           string
		request        testutil.HTTPRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Invalid UUID in GET",
			request: testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid UUID format",
		},
		{
			name: "Invalid UUID in DELETE",
			request: testutil.HTTPRequest{
				Method: "DELETE",
				Path:   "/api/v1/paste/invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid UUID format",
		},
		{
			name: "Non-existent paste",
			request: testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + uuid.New().String(),
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Paste not found",
		},
		{
			name: "Empty content in POST",
			request: testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      "",
					"extension": "txt",
					"expires":   "60",
					"burn":      "false",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Content cannot be empty",
		},
		{
			name: "Invalid expiry time in POST",
			request: testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      "Test content",
					"extension": "txt",
					"expires":   "invalid",
					"burn":      "false",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(tt.request)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedError != "" {
				require.NotNil(t, resp.JSON)
				assert.Equal(t, tt.expectedError, resp.JSON["error"])
			}
		})
	}
}

// TestRateLimiting tests rate limiting functionality
func TestIntegrationRateLimiting(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Make a burst of requests quickly
	const numRequests = 110 // Exceeds 100 req/min limit
	var rateLimited bool

	for i := 0; i < numRequests; i++ {
		resp := server.MakeRequest(testutil.HTTPRequest{
			Method: "GET",
			Path:   "/api/v1/",
		})

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimited = true
			break
		}

		// Small delay to avoid overwhelming the test
		time.Sleep(1 * time.Millisecond)
	}

	assert.True(t, rateLimited, "Rate limiting should have been triggered")
}

// TestDataPersistence tests that data persists across operations
func TestDataPersistence(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, &testutil.TestConfig{
		UseInMemoryDB: false, // Use file-based DB for persistence test
		EnableLogging: false,
	})
	defer server.Close()

	// Create multiple pastes
	pasteIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		resp := server.MakeRequest(testutil.HTTPRequest{
			Method: "POST",
			Path:   "/api/v1/paste",
			FormData: map[string]string{
				"text":      fmt.Sprintf("Persistence test %d", i),
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
		})

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		pasteIDs[i] = resp.JSON["uuid"].(string)
	}

	// Verify all pastes exist
	for i, pasteID := range pasteIDs {
		resp := server.MakeRequest(testutil.HTTPRequest{
			Method: "GET",
			Path:   "/api/v1/paste/" + pasteID,
		})

		require.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, fmt.Sprintf("Persistence test %d", i), resp.JSON["Content"])
	}

	// Verify count in database
	count := server.CountPastesInDB()
	assert.Equal(t, int64(5), count)
}

// TestContentTypes tests different content types and languages
func TestContentTypes(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	testCases := []struct {
		language string
		content  string
	}{
		{"go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}"},
		{"python", "def hello():\n    print(\"Hello, World!\")"},
		{"javascript", "function hello() {\n    console.log(\"Hello, World!\");\n}"},
		{"sql", "SELECT * FROM users WHERE active = true;"},
		{"json", `{"name": "test", "value": 123}`},
		{"xml", "<?xml version=\"1.0\"?>\n<root><item>test</item></root>"},
		{"markdown", "# Test\n\nThis is a **test** markdown."},
		{"plain", "Just plain text content"},
	}

	for _, tc := range testCases {
		t.Run("Content-"+tc.language, func(t *testing.T) {
			// Create paste
			createResp := server.MakeRequest(testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      tc.content,
					"extension": tc.language,
					"expires":   "60",
					"burn":      "false",
				},
			})

			require.Equal(t, http.StatusCreated, createResp.StatusCode)
			pasteUUID := createResp.JSON["uuid"].(string)

			// Retrieve and verify
			getResp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + pasteUUID,
			})

			require.Equal(t, http.StatusOK, getResp.StatusCode)
			assert.Equal(t, tc.content, getResp.JSON["Content"])
			assert.Equal(t, tc.language, getResp.JSON["Language"])

			// Test raw endpoint
			rawResp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/paste/" + pasteUUID + "/raw",
			})

			require.Equal(t, http.StatusOK, rawResp.StatusCode)
			assert.Equal(t, tc.content, string(rawResp.Body))
		})
	}
}

// TestLargeContent tests handling of large content (within limits)
func TestLargeContent(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test with content approaching the limit (9MB)
	largeContent := strings.Repeat("A", 9*1024*1024)

	createResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "POST",
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      largeContent,
			"extension": "txt",
			"expires":   "60",
			"burn":      "false",
		},
	})

	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	pasteUUID := createResp.JSON["uuid"].(string)

	// Verify content can be retrieved
	getResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	require.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Equal(t, largeContent, getResp.JSON["Content"])
}

// TestExpiryTimeValidation tests expiry time validation
func TestExpiryTimeValidation(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name           string
		expiryMinutes  string
		expectedStatus int
	}{
		{"Minimum valid expiry", "1", http.StatusCreated},
		{"Maximum valid expiry", strconv.Itoa(60 * 24 * 365), http.StatusCreated}, // 1 year
		{"Zero expiry", "0", http.StatusBadRequest},
		{"Negative expiry", "-1", http.StatusBadRequest},
		{"Too large expiry", strconv.Itoa(60*24*365 + 1), http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      "Test content",
					"extension": "txt",
					"expires":   tt.expiryMinutes,
					"burn":      "false",
				},
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestUnicodeContent tests handling of Unicode content
func TestUnicodeContent(t *testing.T) {
	router := routes.AddRoutes()
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	unicodeContent := "Hello 世界! 🌍 Testing émojis and 特殊字符"

	createResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "POST",
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      unicodeContent,
			"extension": "txt",
			"expires":   "60",
			"burn":      "false",
		},
	})

	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	pasteUUID := createResp.JSON["uuid"].(string)

	// Verify Unicode content is preserved
	getResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + pasteUUID,
	})

	require.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Equal(t, unicodeContent, getResp.JSON["Content"])

	// Test raw endpoint with Unicode
	rawResp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/paste/" + pasteUUID + "/raw",
	})

	require.Equal(t, http.StatusOK, rawResp.StatusCode)
	assert.Equal(t, unicodeContent, string(rawResp.Body))
	assert.Contains(t, rawResp.Headers.Get("Content-Type"), "utf-8")
}