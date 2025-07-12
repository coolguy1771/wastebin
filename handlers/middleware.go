package handlers

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/coolguy1771/wastebin/log"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// WrapResponseWriter wraps the standard http.ResponseWriter to capture status and size.
type WrapResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

// NewWrapResponseWriter creates a new instance of WrapResponseWriter.
func NewWrapResponseWriter(w http.ResponseWriter, protoMajor int) *WrapResponseWriter {
	return &WrapResponseWriter{ResponseWriter: w, status: 200}
}

// Status returns the response status code.
func (ww *WrapResponseWriter) Status() int {
	return ww.status
}

// BytesWritten returns the number of bytes written in the response.
func (ww *WrapResponseWriter) BytesWritten() int {
	return ww.bytes
}

// WriteHeader overrides the default WriteHeader to capture the status code.
func (ww *WrapResponseWriter) WriteHeader(code int) {
	if ww.wroteHeader {
		return
	}
	ww.status = code
	ww.ResponseWriter.WriteHeader(code)
	ww.wroteHeader = true
}

// Write overrides the default Write to capture the size of the response.
func (ww *WrapResponseWriter) Write(b []byte) (int, error) {
	if !ww.wroteHeader {
		ww.WriteHeader(http.StatusOK)
	}
	size, err := ww.ResponseWriter.Write(b)
	ww.bytes += size
	return size, err
}

// Hijack allows the middleware to support hijacking.
func (ww *WrapResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := ww.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hj.Hijack()
}

// Flush allows the middleware to support flushing.
func (ww *WrapResponseWriter) Flush() {
	fl, ok := ww.ResponseWriter.(http.Flusher)
	if ok {
		fl.Flush()
	}
}

// ZapLogger is a middleware that logs HTTP requests using zap.Logger in JSON format.
func ZapLogger(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from middleware (chi.RequestID middleware adds this)
			requestID := GetRequestID(r)

			// Create a response writer to capture response status and size
			ww := NewWrapResponseWriter(w, r.ProtoMajor)

			// Add request ID to response headers for traceability
			if requestID != "" {
				ww.Header().Set("X-Request-ID", requestID)
			}

			// Call the next handler
			next.ServeHTTP(ww, r)

			// Log the request details with structured logging
			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
				zap.String("protocol", r.Proto),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Int("status_code", ww.Status()),
				zap.Int("response_size", ww.BytesWritten()),
				zap.Duration("duration", time.Since(start)),
				zap.String("timestamp", start.UTC().Format(time.RFC3339)),
			}

			// Add referer if present
			if referer := r.Referer(); referer != "" {
				fields = append(fields, zap.String("referer", referer))
			}

			// Log based on status code
			if ww.Status() >= 500 {
				logger.Error("Request handled with server error", fields...)
			} else if ww.Status() >= 400 {
				logger.Warn("Request handled with client error", fields...)
			} else {
				logger.Info("Request handled successfully", fields...)
			}
		})
	}
}

// GetRequestID extracts the request ID from the context using chi's middleware
func GetRequestID(r *http.Request) string {
	// Use chi's built-in request ID getter
	return middleware.GetReqID(r.Context())
}
