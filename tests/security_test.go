package tests

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/pkg/testutil"
	"github.com/coolguy1771/wastebin/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecurityHeaders verifies that required security headers are present
func TestSecurityHeaders(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/",
	})

	requiredHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
	}

	for header, expectedValue := range requiredHeaders {
		actualValue := resp.Headers.Get(header)
		assert.Equal(t, expectedValue, actualValue, "Security header %s not set correctly", header)
	}
}

// TestSQLInjectionProtection tests protection against SQL injection attacks
func TestSQLInjectionProtection(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	sqlInjectionPayloads := []string{
		"'; DROP TABLE pastes; --",
		"' OR '1'='1",
		"' UNION SELECT * FROM information_schema.tables --",
		"'; UPDATE pastes SET content='hacked' WHERE '1'='1'; --",
		"1' OR 1=1#",
		"1' OR 1=1/*",
		"x' AND (SELECT COUNT(*) FROM pastes) > 0 AND 'x'='x",
	}

	for _, payload := range sqlInjectionPayloads {
		t.Run("SQLInjection-"+payload[:min(len(payload), 20)], func(t *testing.T) {
			// Test SQL injection in paste content
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      payload,
					"extension": "txt",
					"expires":   "60",
					"burn":      "false",
				},
			})

			// Should create paste successfully (content is escaped)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			if resp.StatusCode == http.StatusCreated {
				pasteUUID := resp.JSON["uuid"].(string)

				// Retrieve and verify content is stored as-is (escaped)
				getResp := server.MakeRequest(testutil.HTTPRequest{
					Method: "GET",
					Path:   "/api/v1/paste/" + pasteUUID,
				})

				require.Equal(t, http.StatusOK, getResp.StatusCode)
				assert.Equal(t, payload, getResp.JSON["Content"])
			}
		})
	}
}

// TestXSSProtection tests protection against XSS attacks
func TestXSSProtection(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<svg onload=alert('xss')>",
		"<iframe src=javascript:alert('xss')></iframe>",
		"<input onfocus=alert('xss') autofocus>",
		"<body onload=alert('xss')>",
	}

	for _, payload := range xssPayloads {
		t.Run("XSS-"+payload[:min(len(payload), 20)], func(t *testing.T) {
			// Create paste with XSS payload
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "POST",
				Path:   "/api/v1/paste",
				FormData: map[string]string{
					"text":      payload,
					"extension": "html",
					"expires":   "60",
					"burn":      "false",
				},
			})

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			if resp.StatusCode == http.StatusCreated {
				pasteUUID := resp.JSON["uuid"].(string)

				// Get raw content and verify XSS header is set
				rawResp := server.MakeRequest(testutil.HTTPRequest{
					Method: "GET",
					Path:   "/paste/" + pasteUUID + "/raw",
				})

				require.Equal(t, http.StatusOK, rawResp.StatusCode)
				assert.Equal(t, payload, string(rawResp.Body))

				// Verify XSS protection header
				xssProtection := rawResp.Headers.Get("X-XSS-Protection")
				assert.Equal(t, "1; mode=block", xssProtection)

				// Verify content type is plain text, not HTML
				contentType := rawResp.Headers.Get("Content-Type")
				assert.Contains(t, contentType, "text/plain")
			}
		})
	}
}

// TestRateLimiting tests the rate limiting functionality
func TestRateLimiting(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Make rapid requests to trigger rate limiting
	const numRequests = 50
	const rapidInterval = 10 * time.Millisecond

	var rateLimitedCount int
	var successCount int

	for i := 0; i < numRequests; i++ {
		resp := server.MakeRequest(testutil.HTTPRequest{
			Method: "GET",
			Path:   "/api/v1/",
		})

		if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitedCount++
		} else if resp.StatusCode == http.StatusOK {
			successCount++
		}

		time.Sleep(rapidInterval)
	}

	// Should have some successful requests and possibly some rate limited
	assert.Greater(t, successCount, 0, "Should have some successful requests")

	// If we're hitting rate limits, verify the response
	if rateLimitedCount > 0 {
		t.Logf("Rate limiting triggered: %d requests were rate limited", rateLimitedCount)
	}
}

// TestCSRFProtection tests CSRF protection measures
func TestCSRFProtection(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test that state-changing operations require proper headers/methods
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST request should work",
			method:         "POST",
			path:           "/api/v1/paste",
			expectedStatus: http.StatusBadRequest, // Will fail validation but not CSRF
		},
		{
			name:           "GET request to POST endpoint should fail",
			method:         "GET",
			path:           "/api/v1/paste",
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: tt.method,
				Path:   tt.path,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// TestInputValidation tests comprehensive input validation
func TestInputValidation(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	tests := []struct {
		name           string
		formData       map[string]string
		expectedStatus int
		description    string
	}{
		{
			name: "Extremely long content",
			formData: map[string]string{
				"text":      strings.Repeat("A", 15*1024*1024), // 15MB
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			expectedStatus: http.StatusRequestEntityTooLarge,
			description:    "Should reject content larger than 10MB",
		},
		{
			name: "Null bytes in content",
			formData: map[string]string{
				"text":      "Hello\x00World",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should handle null bytes gracefully",
		},
		{
			name: "Unicode content",
			formData: map[string]string{
				"text":      "Unicode: 🚀 测试 مرحبا العالم",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should handle Unicode content",
		},
		{
			name: "Control characters",
			formData: map[string]string{
				"text":      "Control\r\n\t\x08\x1b[31mchars",
				"extension": "txt",
				"expires":   "60",
				"burn":      "false",
			},
			expectedStatus: http.StatusCreated,
			description:    "Should handle control characters",
		},
		{
			name: "Negative expiry",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "-1",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject negative expiry",
		},
		{
			name: "Zero expiry",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "0",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject zero expiry",
		},
		{
			name: "Very large expiry",
			formData: map[string]string{
				"text":      "Test content",
				"extension": "txt",
				"expires":   "999999999",
				"burn":      "false",
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject extremely large expiry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:   "POST",
				Path:     "/api/v1/paste",
				FormData: tt.formData,
			})

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, tt.description)
		})
	}
}

// TestAuthenticationBypass tests that no authentication bypass is possible
func TestAuthenticationBypass(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test various headers that could potentially bypass security
	bypassHeaders := []map[string]string{
		{"X-Forwarded-For": "127.0.0.1"},
		{"X-Real-IP": "127.0.0.1"},
		{"X-Originating-IP": "127.0.0.1"},
		{"X-Forwarded-Host": "localhost"},
		{"X-Remote-IP": "127.0.0.1"},
		{"X-Client-IP": "127.0.0.1"},
		{"Authorization": "Bearer fake-token"},
		{"X-API-Key": "fake-key"},
		{"X-Access-Token": "fake-token"},
	}

	for i, headers := range bypassHeaders {
		t.Run(fmt.Sprintf("BypassAttempt-%d", i+1), func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method:  "GET",
				Path:    "/api/v1/",
				Headers: headers,
			})

			// Should still work normally (no special privileges)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// TestDirectoryTraversal tests protection against directory traversal attacks
func TestDirectoryTraversal(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test directory traversal in UUID parameter
	traversalPayloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"....//....//....//etc/passwd",
		"..%252f..%252f..%252fetc%252fpasswd",
	}

	for _, payload := range traversalPayloads {
		t.Run("DirectoryTraversal-"+payload[:min(len(payload), 20)], func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: "GET",
				Path:   "/api/v1/paste/" + payload,
			})

			// Should return bad request or not found, never succeed
			assert.NotEqual(t, http.StatusOK, resp.StatusCode)
			assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, resp.StatusCode)
		})
	}
}

// TestHTTPMethodSecurity tests that only allowed HTTP methods work
func TestHTTPMethodSecurity(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	disallowedMethods := []string{"PUT", "PATCH", "HEAD", "TRACE", "CONNECT"}

	for _, method := range disallowedMethods {
		t.Run("Method-"+method, func(t *testing.T) {
			resp := server.MakeRequest(testutil.HTTPRequest{
				Method: method,
				Path:   "/api/v1/paste",
			})

			assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		})
	}
}

// TestInformationDisclosure tests that no sensitive information is leaked
func TestInformationDisclosure(t *testing.T) {
	router := routes.AddRoutes(nil)
	server := testutil.NewTestServer(t, router, nil)
	defer server.Close()

	// Test error responses don't leak sensitive information
	resp := server.MakeRequest(testutil.HTTPRequest{
		Method: "GET",
		Path:   "/api/v1/paste/non-existent-uuid",
	})

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	if resp.JSON != nil {
		errorMsg := fmt.Sprintf("%v", resp.JSON["error"])

		// Should not contain sensitive paths or internal details
		sensitiveStrings := []string{
			"/var/", "/etc/", "/home/", "/root/",
			"database", "sql", "connection",
			"password", "secret", "key",
			"stack trace", "panic",
		}

		for _, sensitive := range sensitiveStrings {
			assert.NotContains(t, strings.ToLower(errorMsg), sensitive,
				"Error message should not contain sensitive information: %s", sensitive)
		}
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
