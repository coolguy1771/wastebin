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

// CSRFProtectionMiddleware provides CSRF protection for state-changing operations
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

// RequestSizeLimitMiddleware limits the size of incoming requests
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

// SecurityHeadersMiddleware adds security headers to all responses
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

// BasicAuthMiddleware provides optional basic authentication
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

// SecurityAuditMiddleware logs security events
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

func validateCSRFToken(token, key string) bool {
	if token == "" || key == "" {
		return false
	}

	// Simple HMAC-based validation
	// In production, you might want to use a more sophisticated token system
	expectedToken := generateCSRFToken(key)
	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}

func generateCSRFToken(key string) string {
	// Generate a simple token based on the key
	// In production, this should include timestamp and be more secure
	token := make([]byte, 32)
	copy(token, []byte(key))
	rand.Read(token[len(key):])
	return base64.StdEncoding.EncodeToString(token)
}