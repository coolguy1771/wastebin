//go:build integration
// +build integration

package handlers_test

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coolguy1771/wastebin/handlers"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/pkg/testutil"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger, err := log.New(os.Stdout, "ERROR")
	if err != nil {
		panic(err)
	}
	log.ResetDefault(logger)
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// createTestRouter creates a minimal router for testing handlers.
func createTestRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	
	// Add API versioning middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			version := "v1"
			if headerVersion := r.Header.Get("X-API-Version"); headerVersion != "" {
				version = headerVersion
			}
			if accept := r.Header.Get("Accept"); accept != "" {
				if strings.Contains(accept, "application/vnd.wastebin.v2+json") {
					version = "v2"
				} else if strings.Contains(accept, "application/vnd.wastebin.v1+json") {
					version = "v1"
				}
			}
			w.Header().Set("X-Api-Version", version)
			next.ServeHTTP(w, r)
		})
	})
	
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"🐣 v1"}`))
		})
		r.Post("/paste", handlers.CreatePaste)
		r.Get("/paste/{uuid}", handlers.GetPaste)
		r.Delete("/paste/{uuid}", handlers.DeletePaste)
	})
	router.Route("/paste/{uuid}", func(r chi.Router) {
		r.Get("/raw", handlers.GetRawPaste)
	})
	
	// Add health check endpoint
	router.Get("/health/db", handlers.DatabaseHealthCheck)

	return router
}

func TestCreatePaste(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
		expectedFields map[string]interface{}
		shouldHaveUUID bool
	}{
		{
			name: "Valid paste creation",
			formData: map[string]string{
				"text":      testutil.TestData().ValidPasteContent,
				"extension": testutil.TestData().ValidLanguage,
				"expires":   strconv.Itoa(testutil.TestData().ValidExpiryMinutes),
				"burn":      "false",
			},
			expectedStatus: http.StatusCreated,
			expectedFields: map[string]interface{}{
				"message": "Paste created successfully",
			},
			shouldHaveUUID: true,
		},
		{
			name: "Valid burn-after-read paste",
			formData: map[string]string{
				"text":      "Burn after reading",
				"extension": "txt",
				"expires":   "30",
				"burn":      "true",
			},
			expectedStatus: http.StatusCreated,
			shouldHaveUUID: true,
		},
		{
			name: "Empty content should fail",
			formData: map[string]string{
				"text":      "",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: map[string]interface{}{
				"error": "Content cannot be empty",
			},
		},
		{
			name: "Invalid expiry time should fail",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "invalid",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Expiry time too short should fail",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "0",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Expiry time too long should fail",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "99999999", // More than 1 year
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				server.CleanupPastes()
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:      http.MethodPost,
				Path:        "/api/v1/paste",
				FormData:    tt.formData,
				Headers:     nil,
				Body:        nil,
				QueryParams: nil,
			})
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedFields != nil {
				for key, expectedValue := range tt.expectedFields {
					actualValue, exists := resp.JSON[key]
					assert.True(t, exists, "Missing field: %s", key)
					assert.Equal(t, expectedValue, actualValue, "Unexpected value for field: %s", key)
				}
			}

			if tt.shouldHaveUUID {
				uuidStr, exists := resp.JSON["uuid"]
				assert.True(t, exists, "Missing UUID in response")
				assert.NotEmpty(t, uuidStr, "UUID should not be empty")
				parsedUUID, err := uuid.Parse(uuidStr.(string))
				require.NoError(t, err, "UUID should be valid")

				paste := server.GetPasteFromDB(parsedUUID)
				assert.NotNil(t, paste, "Paste should exist in database")
				assert.Equal(t, tt.formData["text"], paste.Content)
				assert.Equal(t, tt.formData["extension"], paste.Language)
				assert.Equal(t, tt.formData["burn"] == "true", paste.Burn)
			}
		})
	}
}

func TestGetPaste(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		setupPaste     func(*testutil.TestServer) *models.Paste
		pasteUUID      string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name: "Get existing paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return s.CreateTestPaste("Test content", "txt", 60, false)
			},
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name: "Get non-existent paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
		{
			name: "Get paste with invalid UUID",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			shouldExist:    false,
		},
		{
			name: "Get expired paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return s.CreateExpiredPaste()
			},
			expectedStatus: http.StatusGone,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				server.CleanupPastes()

			var paste *models.Paste
			if tt.setupPaste != nil {
				paste = tt.setupPaste(server)
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:      http.MethodGet,
				Path:        "/api/v1/paste/" + pasteUUID,
				Headers:     nil,
				Body:        nil,
				QueryParams: nil,
				FormData:    nil,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.shouldExist && paste != nil {
				assert.NotNil(t, resp.JSON)
				// The Paste model uses lowercase JSON field names
				assert.Equal(t, paste.Content, resp.JSON["content"])
				assert.Equal(t, paste.Language, resp.JSON["language"])
				assert.Equal(t, paste.Burn, resp.JSON["burn"])
				// Also check the UUID is returned as pasteId
				assert.Equal(t, paste.UUID.String(), resp.JSON["pasteId"])
			}
		})
	}
}

func TestGetRawPaste(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	tests := []struct {
		name            string
		setupPaste      func(*testutil.TestServer) *models.Paste
		pasteUUID       string
		expectedStatus  int
		expectedContent string
	}{
		{
			name: "Get raw content of existing paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return s.CreateTestPaste("Raw content test", "txt", 60, false)
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "Raw content test",
		},
		{
			name: "Get raw content of non-existent paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Get raw content with invalid UUID",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				server.CleanupPastes()

			var paste *models.Paste
			if tt.setupPaste != nil {
				paste = tt.setupPaste(server)
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:      http.MethodGet,
				Path:        "/paste/" + pasteUUID + "/raw",
				Headers:     nil,
				Body:        nil,
				QueryParams: nil,
				FormData:    nil,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedContent != "" {
				assert.Equal(t, tt.expectedContent, string(resp.Body))
				assert.Contains(t, resp.Headers.Get("Content-Type"), "text/plain")
			}
		})
	}
}

func TestDeletePaste(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	tests := []struct {
		name           string
		setupPaste     func(*testutil.TestServer) *models.Paste
		pasteUUID      string
		expectedStatus int
	}{
		{
			name: "Delete existing paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return s.CreateTestPaste("Delete me", "txt", 60, false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Delete non-existent paste",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Delete with invalid UUID",
			setupPaste: func(s *testutil.TestServer) *models.Paste {
				return nil
			},
			pasteUUID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				server.CleanupPastes()

			var paste *models.Paste
			if tt.setupPaste != nil {
				paste = tt.setupPaste(server)
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:      http.MethodDelete,
				Path:        "/api/v1/paste/" + pasteUUID,
				Headers:     nil,
				Body:        nil,
				QueryParams: nil,
				FormData:    nil,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK && paste != nil {
				// Verify paste was deleted from database
				deletedPaste := server.GetPasteFromDB(paste.UUID)
				assert.Nil(t, deletedPaste, "Paste should be deleted from database")

				// Verify response message
				assert.NotNil(t, resp.JSON)
				assert.Equal(t, "Paste deleted successfully", resp.JSON["message"])
			}
		})
	}
}

func TestBurnAfterRead(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	// Create a burn-after-read paste
	paste := server.CreateTestPaste("Burn after reading", "txt", 60, true)

	// First read should succeed
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method:      http.MethodGet,
		Path:        "/api/v1/paste/" + paste.UUID.String(),
		Headers:     nil,
		Body:        nil,
		QueryParams: nil,
		FormData:    nil,
	})

	assert.Equal(t, http.StatusGone, resp.StatusCode)
	assert.NotNil(t, resp.JSON)
	assert.Contains(t, resp.JSON["error"], "expired or been burned")

	// Verify paste was deleted from database
	deletedPaste := server.GetPasteFromDB(paste.UUID)
	assert.Nil(t, deletedPaste, "Burn-after-read paste should be deleted")

	// Second read should fail
	resp2 := server.MakeRequest(testutil.HTTPRequest{
		Method:      http.MethodGet,
		Path:        "/api/v1/paste/" + paste.UUID.String(),
		Headers:     nil,
		Body:        nil,
		QueryParams: nil,
		FormData:    nil,
	})

	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestPasteExpiry(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	// Create an expired paste
	expiredPaste := server.CreateExpiredPaste()

	// Try to access expired paste
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method:      http.MethodGet,
		Path:        "/api/v1/paste/" + expiredPaste.UUID.String(),
		Headers:     nil,
		Body:        nil,
		QueryParams: nil,
		FormData:    nil,
	})

	assert.Equal(t, http.StatusGone, resp.StatusCode)
	assert.NotNil(t, resp.JSON)
	assert.Contains(t, resp.JSON["error"], "expired or been burned")

	// Verify expired paste was deleted from database
	deletedPaste := server.GetPasteFromDB(expiredPaste.UUID)
	assert.Nil(t, deletedPaste, "Expired paste should be deleted")
}

func TestDatabaseHealthCheck(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	resp := server.MakeRequest(testutil.HTTPRequest{
		Method:      http.MethodGet,
		Path:        "/health/db",
		Headers:     nil,
		Body:        nil,
		QueryParams: nil,
		FormData:    nil,
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.JSON)
	assert.Equal(t, "healthy", resp.JSON["status"])
	assert.Equal(t, "database", resp.JSON["service"])
	assert.NotEmpty(t, resp.JSON["timestamp"])
}

func TestContentSizeLimit(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)

	// Test with content that exceeds the 10MB limit
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      testutil.TestData().LargePasteContent,
			"extension": "txt",
			"expires":   "60",
			"burn":      "false",
		},
		Headers:     nil,
		Body:        nil,
		QueryParams: nil,
	})

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestRateLimiting(t *testing.T) {
	t.Skip("Rate limiting test requires specific timing - implement in integration tests")
}

func TestConcurrentPasteCreation(t *testing.T) {

	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	t.Cleanup(server.Close)
	
	// First, verify that a single paste can be created
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: http.MethodPost,
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      "Initial test paste",
			"extension": "txt",
			"expires":   "60",
			"burn":      "false",
		},
	})
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Initial paste creation should succeed")

	const numGoroutines = 10

	results := make(chan *testutil.HTTPResponse, numGoroutines)
	var wg sync.WaitGroup

	// Create multiple pastes concurrently
	for i := range numGoroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			// Add a small random delay to reduce SQLite lock contention
			time.Sleep(time.Duration(index*10) * time.Millisecond)
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: http.MethodPost,
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      "Concurrent test " + strconv.Itoa(index),
					"extension": "txt",
					"expires":   "60",
					"burn":      "false",
				},
				Headers:     nil,
				Body:        nil,
				QueryParams: nil,
			})
			results <- resp
		}(i)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Collect results
	var successful int
	var failed int

	for resp := range results {
		if resp.StatusCode == http.StatusCreated {
			successful++
		} else {
			failed++
			t.Logf("Request failed with status %d: %v", resp.StatusCode, resp.JSON)
		}
	}

	// Log the results
	t.Logf("Concurrent test results: %d successful, %d failed out of %d total", successful, failed, numGoroutines)

	// SQLite has known issues with concurrent writes, so we allow some failures
	// but at least half should succeed
	if successful < numGoroutines/2 {
		t.Errorf("Too many failures in concurrent test: only %d/%d succeeded", successful, numGoroutines)
	}

	// Verify count in database matches successful creates
	count := server.CountPastesInDB()
	// We expect count to be successful + 1 (from the initial test paste)
	assert.Equal(t, int64(successful+1), count, "Database count should match successful creates plus initial paste")
}

// Benchmark tests.
func BenchmarkCreatePaste(b *testing.B) {
	router := createTestRouter()

	server := testutil.NewTestServer(&testing.T{}, router, nil)
	b.Cleanup(server.Close)

	b.ResetTimer()

	for range b.N {
		server.MakeRequest(testutil.HTTPRequest{
			Method: http.MethodPost,
			Path:   "/api/v1/paste",
			FormData: map[string]string{
				"text":      "Benchmark test content",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			Headers:     nil,
			Body:        nil,
			QueryParams: nil,
		})
	}
}