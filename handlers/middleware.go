package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

		// Skip CSRF protection for API endpoints only if using API key authentication
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// Check if request uses API key authentication
			if r.Header.Get("X-API-Key") != "" || r.Header.Get("Authorization") != "" {
				next.ServeHTTP(w, r)
				return
			}
			// For cookie-based API auth, continue with CSRF validation
		}

		// Generate and validate CSRF token for web forms using double-submit cookie pattern
		if config.Conf.CSRFKey != "" {
			// Get token from header or form
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			// Get session ID from cookie or generate one
			sessionID, err := getOrCreateSessionID(w, r)
			if err != nil {
				log.Error("Failed to get or create session ID for CSRF validation", zap.Error(err))
				respondWithError(w, http.StatusInternalServerError, "Session management failure")
				return
			}
			
			if !validateCSRFToken(token, sessionID, config.Conf.CSRFKey) {
				log.Warn("CSRF validation failed",
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("user_agent", r.UserAgent()),
					zap.String("path", r.URL.Path),
					zap.String("session_id", sessionID))
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

const (
	csrfTokenTTL = 24 * time.Hour // Token expires after 24 hours
	sessionCookieName = "wastebin_session"
	csrfCookieName = "wastebin_csrf"
)

// getOrCreateSessionID gets the session ID from cookie or creates a new one
func getOrCreateSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	// Try to get existing session ID from cookie
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if cookie.Value != "" {
			return cookie.Value, nil
		}
	}

	// Generate new session ID
	sessionID, err := generateSecureRandomString(32)
	if err != nil {
		log.Error("Failed to generate secure session ID", zap.Error(err))
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	
	// Set session cookie (HttpOnly, Secure in production, SameSite)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   !config.Conf.Dev, // Only require HTTPS in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(csrfTokenTTL.Seconds()),
	})

	return sessionID, nil
}

// generateSecureRandomString generates a cryptographically secure random string
func generateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fail securely - do not use predictable fallbacks
		return "", fmt.Errorf("failed to generate secure random string: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// generateCSRFToken creates a secure CSRF token using HMAC with timestamp
func generateCSRFToken(sessionID, secretKey string) string {
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("%s:%d", sessionID, timestamp)
	
	// Create HMAC with SHA256
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))
	
	// Format: sessionID:timestamp:signature
	token := fmt.Sprintf("%s:%d:%s", sessionID, timestamp, signature)
	return base64.StdEncoding.EncodeToString([]byte(token))
}

// validateCSRFToken validates a CSRF token using HMAC and checks expiration
func validateCSRFToken(tokenStr, sessionID, secretKey string) bool {
	if tokenStr == "" || sessionID == "" || secretKey == "" {
		return false
	}

	// Decode base64 token
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		log.Warn("Invalid base64 CSRF token", zap.Error(err))
		return false
	}

	token := string(tokenBytes)
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		log.Warn("Invalid CSRF token format")
		return false
	}

	tokenSessionID := parts[0]
	timestampStr := parts[1]
	providedSignature := parts[2]

	// Verify session ID matches
	if subtle.ConstantTimeCompare([]byte(tokenSessionID), []byte(sessionID)) != 1 {
		log.Warn("CSRF token session ID mismatch")
		return false
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		log.Warn("Invalid CSRF token timestamp", zap.Error(err))
		return false
	}

	// Check if token has expired
	tokenTime := time.Unix(timestamp, 0)
	if time.Since(tokenTime) > csrfTokenTTL {
		log.Warn("CSRF token expired", zap.Time("token_time", tokenTime))
		return false
	}

	// Recreate expected signature
	message := fmt.Sprintf("%s:%d", sessionID, timestamp)
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare signatures using constant time comparison
	return subtle.ConstantTimeCompare([]byte(providedSignature), []byte(expectedSignature)) == 1
}

// GetCSRFToken generates a new CSRF token for the current session (for use in templates/frontend)
func GetCSRFToken(w http.ResponseWriter, r *http.Request) string {
	if config.Conf.CSRFKey == "" {
		return ""
	}

	sessionID, err := getOrCreateSessionID(w, r)
	if err != nil {
		log.Error("Failed to get or create session ID for CSRF token generation", zap.Error(err))
		return "" // Fail securely by returning empty token
	}
	
	token := generateCSRFToken(sessionID, config.Conf.CSRFKey)

	// Also set as cookie for double-submit pattern
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Needs to be accessible by JavaScript
		Secure:   !config.Conf.Dev, // Only require HTTPS in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(csrfTokenTTL.Seconds()),
	})

	return token
}