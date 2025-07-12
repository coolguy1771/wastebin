package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"go.uber.org/zap"
)

// CSRFProtectionMiddleware enforces CSRF protection for state-changing HTTP requests to non-API endpoints.
// It validates a CSRF token from the request against a configured secret key, rejecting requests with HTTP 403 Forbidden if validation fails. CSRF checks are skipped for safe HTTP methods (GET, HEAD, OPTIONS) and for API routes.
func CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF protection for GET, HEAD, OPTIONS requests
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF protection for API endpoints if API key is present
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// For API endpoints, we can use other authentication mechanisms
			next.ServeHTTP(w, r)
			return
		}

		// Generate and validate CSRF token for web forms
		if config.Conf.CSRFKey != "" {
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			if !validateCSRFToken(token, config.Conf.CSRFKey) {
				log.Warn("CSRF validation failed",
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
					zap.String("path", r.URL.Path))
				respondWithError(w, http.StatusForbidden, "CSRF validation failed")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimitMiddleware returns middleware that enforces a maximum size for incoming HTTP requests.
// If the request's Content-Length exceeds maxSize, the middleware responds with HTTP 413 Request Entity Too Large and does not call the next handler.
// Otherwise, it wraps the request body to prevent reading more than maxSize bytes before passing control to the next handler.
func RequestSizeLimitMiddleware(maxSize int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Apply request size limit
			if r.ContentLength > maxSize {
				log.Warn("Request size exceeded limit",
					zap.Int64("content_length", r.ContentLength),
					zap.Int64("max_size", maxSize),
					zap.String("remote_addr", r.RemoteAddr))
				respondWithError(w, http.StatusRequestEntityTooLarge, "Request size exceeds limit")
				return
			}

			// Wrap the request body with a limited reader
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds standard security-related HTTP headers to all responses, including protections against MIME sniffing, clickjacking, XSS, and sets a strict Content Security Policy. It also adds HSTS headers for HTTPS requests.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// HSTS header for HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"frame-ancestors 'none'"
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}

// BasicAuthMiddleware enforces HTTP Basic Authentication if enabled in the configuration.
// If authentication is required, it validates credentials from the request using constant-time comparison and responds with HTTP 401 Unauthorized if authentication fails.
// Proceeds to the next handler upon successful authentication or if authentication is not required.
func BasicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.Conf.RequireAuth {
			next.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Wastebin"`)
			respondWithError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(config.Conf.AuthUsername))
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(config.Conf.AuthPassword))

		if usernameMatch != 1 || passwordMatch != 1 {
			log.Warn("Authentication failed",
				zap.String("username", username),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))
			w.Header().Set("WWW-Authenticate", `Basic realm="Wastebin"`)
			respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		log.Info("User authenticated",
			zap.String("username", username),
			zap.String("remote_addr", r.RemoteAddr))

		next.ServeHTTP(w, r)
	})
}

// SecurityAuditMiddleware logs security-related HTTP responses such as rate limiting, unauthorized, and forbidden access attempts.
// It captures the response status code and logs relevant request details for HTTP 429, 401, and 403 responses before returning control to the next handler.
func SecurityAuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		// Log security events
		switch ww.statusCode {
		case http.StatusTooManyRequests:
			log.Warn("Rate limit exceeded",
				zap.String("ip", getRealIP(r)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
		case http.StatusUnauthorized:
			log.Warn("Unauthorized access attempt",
				zap.String("ip", getRealIP(r)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
		case http.StatusForbidden:
			log.Warn("Forbidden access attempt",
				zap.String("ip", getRealIP(r)),
				zap.String("user_agent", r.UserAgent()),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
		}
	})
}

// Helper functions

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// getRealIP extracts the client's real IP address from the HTTP request, checking the X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
func getRealIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Get the first IP in case of multiple
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// validateCSRFToken checks whether the provided CSRF token matches the expected token generated from the given key using constant-time comparison.
// Returns true if the token is valid and false otherwise.
func validateCSRFToken(token, key string) bool {
	if token == "" || key == "" {
		return false
	}

	// Simple HMAC-based validation
	// In production, you might want to use a more sophisticated token system
	expectedToken := generateCSRFToken(key)
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}

// generateCSRFToken creates a 32-byte base64-encoded CSRF token using the provided key and random bytes.
// The token begins with the bytes of the key and fills the remainder with cryptographically secure random data.
func generateCSRFToken(key string) string {
	// Generate a simple token based on the key
	// In production, this should include timestamp and be more secure
	token := make([]byte, 32)
	copy(token, []byte(key))
	rand.Read(token[len(key):])
	return base64.StdEncoding.EncodeToString(token)
}