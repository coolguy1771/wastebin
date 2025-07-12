package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Common errors
var (
	ErrEmptyContent       = errors.New("content cannot be empty")
	ErrInvalidExpiry      = errors.New("invalid expiry time")
	ErrExpiryInPast       = errors.New("expiry time cannot be in the past")
	ErrInvalidUUID        = errors.New("invalid UUID format")
	ErrPasteNotFound      = errors.New("paste not found")
	ErrContentTooLarge    = errors.New("content exceeds maximum size")
	ErrInvalidContentType = errors.New("invalid content type")
)

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// respondWithDetailedError sends a JSON error response with additional details
func respondWithDetailedError(w http.ResponseWriter, code int, message, errorCode, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Code:    errorCode,
		Details: details,
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
