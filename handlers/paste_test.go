package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/pkg/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// createTestRouter creates a minimal router for testing handlers.
func createTestRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/paste", CreatePaste)
		r.Get("/paste/{uuid}", GetPaste)
		r.Delete("/paste/{uuid}", DeletePaste)
	})
	r.Route("/paste/{uuid}", func(r chi.Router) {
		r.Get("/raw", GetRawPaste)
	})

	return r
}

func TestCreatePaste(t *testing.T) {
	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

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
				"text":      testutil.TestData.ValidPasteContent,
				"extension": testutil.TestData.ValidLanguage,
				"expires":   strconv.Itoa(testutil.TestData.ValidExpiryMinutes),
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
			// Clean up before each test
			server.CleanupPastes()

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:   "POST",
				Path:     "/api/v1/paste",
				FormData: tt.formData,
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

				// Verify UUID is valid
				parsedUUID, err := uuid.Parse(uuidStr.(string))
				assert.NoError(t, err, "UUID should be valid")

				// Verify paste was created in database
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
	defer server.Close()

	tests := []struct {
		name           string
		setupPaste     func() *models.Paste
		pasteUUID      string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name: "Get existing paste",
			setupPaste: func() *models.Paste {
				return server.CreateTestPaste("Test content", "txt", 60, false)
			},
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name: "Get non-existent paste",
			setupPaste: func() *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			shouldExist:    false,
		},
		{
			name: "Get paste with invalid UUID",
			setupPaste: func() *models.Paste {
				return nil
			},
			pasteUUID:      "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			shouldExist:    false,
		},
		{
			name: "Get expired paste",
			setupPaste: func() *models.Paste {
				return server.CreateExpiredPaste()
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
				paste = tt.setupPaste()
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + pasteUUID,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.shouldExist && paste != nil {
				assert.NotNil(t, resp.JSON)
				assert.Equal(t, paste.Content, resp.JSON["Content"])
				assert.Equal(t, paste.Language, resp.JSON["Language"])
				assert.Equal(t, paste.Burn, resp.JSON["Burn"])
			}
		})
	}
}

func TestGetRawPaste(t *testing.T) {
	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name            string
		setupPaste      func() *models.Paste
		pasteUUID       string
		expectedStatus  int
		expectedContent string
	}{
		{
			name: "Get raw content of existing paste",
			setupPaste: func() *models.Paste {
				return server.CreateTestPaste("Raw content test", "txt", 60, false)
			},
			expectedStatus:  http.StatusOK,
			expectedContent: "Raw content test",
		},
		{
			name: "Get raw content of non-existent paste",
			setupPaste: func() *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Get raw content with invalid UUID",
			setupPaste: func() *models.Paste {
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
				paste = tt.setupPaste()
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/paste/" + pasteUUID + "/raw",
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
	defer server.Close()

	tests := []struct {
		name           string
		setupPaste     func() *models.Paste
		pasteUUID      string
		expectedStatus int
	}{
		{
			name: "Delete existing paste",
			setupPaste: func() *models.Paste {
				return server.CreateTestPaste("Delete me", "txt", 60, false)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Delete non-existent paste",
			setupPaste: func() *models.Paste {
				return nil
			},
			pasteUUID:      uuid.New().String(),
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Delete with invalid UUID",
			setupPaste: func() *models.Paste {
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
				paste = tt.setupPaste()
			}

			pasteUUID := tt.pasteUUID
			if paste != nil {
				pasteUUID = paste.UUID.String()
			}

			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "DELETE",
				Path:   "/api/v1/paste/" + pasteUUID,
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
	defer server.Close()

	// Create a burn-after-read paste
	paste := server.CreateTestPaste("Burn after reading", "txt", 60, true)

	// First read should succeed
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + paste.UUID.String(),
	})

	assert.Equal(t, http.StatusGone, resp.StatusCode)
	assert.NotNil(t, resp.JSON)
	assert.Contains(t, resp.JSON["error"], "expired or been burned")

	// Verify paste was deleted from database
	deletedPaste := server.GetPasteFromDB(paste.UUID)
	assert.Nil(t, deletedPaste, "Burn-after-read paste should be deleted")

	// Second read should fail
	resp2 := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + paste.UUID.String(),
	})

	assert.Equal(t, http.StatusNotFound, resp2.StatusCode)
}

func TestPasteExpiry(t *testing.T) {
	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Create an expired paste
	expiredPaste := server.CreateExpiredPaste()

	// Try to access expired paste
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/" + expiredPaste.UUID.String(),
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
	defer server.Close()

	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/health/db",
	})

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NotNil(t, resp.JSON)
	assert.Equal(t, "healthy", resp.JSON["status"])
	assert.Equal(t, "database", resp.JSON["service"])
	assert.NotEmpty(t, resp.JSON["timestamp"])
}

func TestAPIVersioning(t *testing.T) {
	router := createTestRouter()

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
			name: "Explicit v1 via header",
			headers: map[string]string{
				"X-API-Version": "v1",
			},
			expectedVersion: "v1",
		},
		{
			name: "v2 via header",
			headers: map[string]string{
				"X-API-Version": "v2",
			},
			expectedVersion: "v2",
		},
		{
			name: "v1 via content negotiation",
			headers: map[string]string{
				"Accept": "application/vnd.wastebin.v1+json",
			},
			expectedVersion: "v1",
		},
		{
			name: "v2 via content negotiation",
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

func TestContentSizeLimit(t *testing.T) {
	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test with content that exceeds the 10MB limit
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "POST",
		Path:   "/api/v1/paste",
		FormData: map[string]string{
			"text":      testutil.TestData.LargePasteContent,
			"extension": "txt",
			"expires":   "60",
			"burn":      "false",
		},
	})

	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestRateLimiting(t *testing.T) {
	t.Skip("Rate limiting test requires specific timing - implement in integration tests")
}

func TestConcurrentPasteCreation(t *testing.T) {
	router := createTestRouter()

	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	const numGoroutines = 10

	results := make(chan *testutil.HTTPResponse, numGoroutines)

	// Create multiple pastes concurrently
	for i := range numGoroutines {
		go func(index int) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      "Concurrent test " + strconv.Itoa(index),
					"extension": "txt",
					"expires":   "60",
					"burn":      "false",
				},
			})
			results <- resp
		}(i)
	}

	// Collect results
	var successful int

	for range numGoroutines {
		resp := <-results
		if resp.StatusCode == http.StatusCreated {
			successful++
		}
	}

	// All should succeed
	assert.Equal(t, numGoroutines, successful)

	// Verify count in database
	count := server.CountPastesInDB()
	assert.Equal(t, int64(numGoroutines), count)
}

// Helper function tests.
func TestValidateCreatePasteRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     models.CreatePasteRequest
		expectError bool
		errorType   error
	}{
		{
			name: "Valid request",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			expectError: false,
		},
		{
			name: "Empty content",
			request: models.CreatePasteRequest{
				Content:    "",
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			expectError: true,
			errorType:   ErrEmptyContent,
		},
		{
			name: "Content too large",
			request: models.CreatePasteRequest{
				Content:    strings.Repeat("A", MaxPasteSize+1),
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			},
			expectError: true,
			errorType:   ErrContentTooLarge,
		},
		{
			name: "Empty expiry time",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: "",
			},
			expectError: true,
			errorType:   ErrInvalidExpiry,
		},
		{
			name: "Past expiry time",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			expectError: true,
			errorType:   ErrExpiryInPast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreatePasteRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err)

				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseExpiryTime(t *testing.T) {
	// Test valid RFC3339 timestamp
	validTime := time.Now().Add(1 * time.Hour)
	validTimeStr := validTime.Format(time.RFC3339)

	parsed := parseExpiryTime(validTimeStr)
	assert.True(t, parsed.Equal(validTime.Truncate(time.Second)))

	// Test invalid timestamp - should return zero time
	parsed = parseExpiryTime("invalid-time")
	assert.True(t, parsed.IsZero())
}

// Benchmark tests.
func BenchmarkCreatePaste(b *testing.B) {
	router := createTestRouter()

	server := testutil.NewTestServer(&testing.T{}, router, nil)
	defer server.Close()

	b.ResetTimer()

	for range b.N {
		server.MakeRequest(testutil.HTTPRequest{
			Method: "POST",
			Path:   "/api/v1/paste",
			FormData: map[string]string{
				"text":      "Benchmark test content",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
		})
	}
}
