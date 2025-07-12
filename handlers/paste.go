package handlers

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetRawPaste retrieves a paste's raw content by its UUID.
func GetRawPaste(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pasteUUID := chi.URLParam(r, "uuid")

	paste, err := getPasteByUUID(ctx, pasteUUID)
	if err != nil {
		switch err {
		case ErrPasteNotFound:
			respondWithError(w, http.StatusNotFound, "Paste not found")
		case ErrInvalidUUID:
			respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		default:
			log.Error("Error retrieving paste", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve paste")
		}

		return
	}

	// Check if the paste has expired or is a burn-after-read paste.
	if shouldDelete, err := handlePasteExpiryAndBurn(ctx, paste); err != nil {
		log.Error("Error handling paste expiry/burn", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to process paste")

		return
	} else if shouldDelete {
		respondWithError(w, http.StatusGone, "Paste has expired or been burned")

		return
	}

	// Set content type and return raw content.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(paste.Content))
}

// GetPaste retrieves a paste by its UUID and returns it as JSON.
func GetPaste(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pasteUUID := chi.URLParam(r, "uuid")

	paste, err := getPasteByUUID(ctx, pasteUUID)
	if err != nil {
		switch err {
		case ErrPasteNotFound:
			respondWithError(w, http.StatusNotFound, "Paste not found")
		case ErrInvalidUUID:
			respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		default:
			log.Error("Error retrieving paste", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve paste")
		}

		return
	}

	// Check if the paste has expired or is a burn-after-read paste.
	if shouldDelete, err := handlePasteExpiryAndBurn(ctx, paste); err != nil {
		log.Error("Error handling paste expiry/burn", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to process paste")

		return
	} else if shouldDelete {
		respondWithError(w, http.StatusGone, "Paste has expired or been burned")

		return
	}

	respondWithJSON(w, http.StatusOK, paste)
}

// Constants for paste constraints.
const (
	MaxPasteSize     = 10 * 1024 * 1024 // 10MB
	MinExpiryMinutes = 1
	MaxExpiryMinutes = 60 * 24 * 365 // 1 year
)

// Security patterns to detect potentially dangerous content.
var (
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)data:text/html`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // onclick, onload, etc.
	}

	// Allowed languages for syntax highlighting.
	allowedLanguages = map[string]bool{
		"":           true, // plain text
		"txt":        true,
		"javascript": true,
		"python":     true,
		"go":         true,
		"java":       true,
		"c":          true,
		"cpp":        true,
		"html":       true,
		"css":        true,
		"json":       true,
		"xml":        true,
		"yaml":       true,
		"markdown":   true,
		"sql":        true,
		"bash":       true,
		"shell":      true,
		"php":        true,
		"ruby":       true,
		"rust":       true,
	}
)

// CreatePaste processes an HTTP request to create a new paste, including content sanitization, language and expiry validation, and database storage.
// On success, responds with HTTP 201 and the paste UUID in JSON; on failure, returns an appropriate error response.
func CreatePaste(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.Info("CreatePaste called")

	// Parse form with size limit
	r.Body = http.MaxBytesReader(w, r.Body, MaxPasteSize)
	if err := r.ParseForm(); err != nil {
		log.Error("Error parsing form data", zap.Error(err))

		if err.Error() == "http: request body too large" {
			respondWithError(w, http.StatusRequestEntityTooLarge, "Request body too large")
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid form data")
		}

		return
	}

	// Parse and validate expiry time
	expireTime, err := strconv.ParseInt(r.FormValue("expires"), 10, 64)
	if err != nil {
		log.Error("Error parsing expiry time", zap.Error(err))
		respondWithDetailedError(w, http.StatusBadRequest, "Invalid expiry time", "INVALID_EXPIRY", "Expiry time must be a number")

		return
	}

	// Validate expiry time range
	if expireTime < MinExpiryMinutes || expireTime > MaxExpiryMinutes {
		respondWithDetailedError(w, http.StatusBadRequest, "Invalid expiry time", "EXPIRY_OUT_OF_RANGE",
			"Expiry time must be between 1 minute and 1 year")

		return
	}

	req := models.CreatePasteRequest{
		Content:    r.FormValue("text"),
		Burn:       r.FormValue("burn") == "true",
		Language:   r.FormValue("extension"),
		ExpiryTime: time.Now().Add(time.Duration(expireTime) * time.Minute).Format(time.RFC3339),
	}

	// Sanitize content before validation
	sanitizedContent, err := sanitizeContent(req.Content)
	if err != nil {
		log.Error("Error sanitizing content", zap.Error(err))

		if errors.Is(err, ErrInvalidUTF8) {
			respondWithError(w, http.StatusBadRequest, "Content contains invalid UTF-8 encoding")
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid content")
		}

		return
	}

	req.Content = sanitizedContent

	if err := validateCreatePasteRequest(req); err != nil {
		log.Error("Error validating create paste request", zap.Error(err))

		if err == ErrEmptyContent {
			respondWithError(w, http.StatusBadRequest, "Content cannot be empty")
		} else if err == ErrContentTooLarge {
			respondWithError(w, http.StatusRequestEntityTooLarge, "Content exceeds maximum size")
		} else if errors.Is(err, ErrInvalidLanguage) {
			respondWithError(w, http.StatusBadRequest, "Invalid or unsupported language")
		} else {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}

		return
	}

	pasteUUID, err := uuid.NewRandom()
	if err != nil {
		log.Error("Error generating UUID", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to generate paste ID")

		return
	}

	paste := models.Paste{
		UUID:            pasteUUID,
		Content:         req.Content,
		Burn:            req.Burn,
		Language:        req.Language,
		ExpiryTimestamp: parseExpiryTime(req.ExpiryTime),
	}

	if err := storage.DBConn.WithContext(ctx).Create(&paste).Error; err != nil {
		log.Error("Error saving paste to database", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to save paste")

		return
	}

	log.Info("Paste saved to database", zap.String("uuid", pasteUUID.String()))
	respondWithJSON(w, http.StatusCreated, map[string]string{
		"message": "Paste created successfully",
		"uuid":    pasteUUID.String(),
	})
}

// DeletePaste deletes a paste by its UUID.
func DeletePaste(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pasteUUIDStr := chi.URLParam(r, "uuid")

	if pasteUUIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "UUID parameter is required")

		return
	}

	pasteUUID, err := uuid.Parse(pasteUUIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid UUID format")

		return
	}

	if err := deletePasteByUUID(ctx, pasteUUID); err != nil {
		if err == ErrPasteNotFound {
			respondWithError(w, http.StatusNotFound, "Paste not found")
		} else {
			log.Error("Error deleting paste", zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to delete paste")
		}

		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Paste deleted successfully",
	})
}

// Helper function to retrieve a paste by its UUID.
func getPasteByUUID(ctx context.Context, uuidStr string) (*models.Paste, error) {
	pasteUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	paste := &models.Paste{}
	if err := storage.DBConn.WithContext(ctx).First(paste, "uuid = ?", pasteUUID).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, ErrPasteNotFound
		}

		return nil, err
	}

	return paste, nil
}

// Helper function to handle paste expiry and burn-after-read.
func handlePasteExpiryAndBurn(ctx context.Context, paste *models.Paste) (bool, error) {
	if time.Now().After(paste.ExpiryTimestamp) {
		return true, deletePasteByUUID(ctx, paste.UUID)
	}

	if paste.Burn {
		err := storage.DBConn.WithContext(ctx).Delete(paste).Error
		if err != nil {
			log.Error("Error deleting paste after reading", zap.Error(err))

			return false, err
		}

		return true, nil
	}

	return false, nil
}

// deletePasteByUUID removes a paste from the database by its UUID.
// Returns ErrPasteNotFound if no paste with the given UUID exists, or a database error if the operation fails.
func deletePasteByUUID(ctx context.Context, pasteUUID uuid.UUID) error {
	result := storage.DBConn.WithContext(ctx).Where("uuid = ?", pasteUUID).Delete(&models.Paste{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrPasteNotFound
	}

	return nil
}

// sanitizeContent validates and cleans paste content by ensuring valid UTF-8 encoding, removing null bytes and carriage returns, and logging warnings if potentially dangerous patterns are detected. Returns the sanitized content or an error if the content is not valid UTF-8.
func sanitizeContent(content string) (string, error) {
	// Validate UTF-8 encoding
	if !utf8.ValidString(content) {
		log.Warn("Invalid UTF-8 content detected")

		return "", ErrInvalidUTF8
	}

	// Remove null bytes which can cause security issues
	content = strings.ReplaceAll(content, "\x00", "")

	// Remove carriage returns that might cause CRLF injection
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Check for potentially dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(content) {
			log.Warn("Potentially dangerous pattern detected in content",
				zap.String("pattern", pattern.String()))
			// For now, we'll log but allow the content
			// In a more strict environment, you might want to reject it
		}
	}

	return content, nil
}

// validateLanguage returns true if the given language is in the set of allowed syntax highlighting languages, or if it is empty (plain text).
func validateLanguage(language string) bool {
	if language == "" {
		return true // empty language is allowed (plain text)
	}

	// Convert to lowercase for case-insensitive comparison
	language = strings.ToLower(strings.TrimSpace(language))

	return allowedLanguages[language]
}

// validateCreatePasteRequest checks the validity of a CreatePasteRequest, ensuring content is present and within size limits, the language is allowed, and the expiry time is valid and within acceptable bounds.
// Returns a specific error if any validation fails.
func validateCreatePasteRequest(req models.CreatePasteRequest) error {
	if req.Content == "" {
		return ErrEmptyContent
	}

	// Validate content size before sanitization
	if len(req.Content) > MaxPasteSize {
		return ErrContentTooLarge
	}

	// Content sanitization is done before calling this function
	// Just validate that the content is already sanitized

	// Validate language
	if !validateLanguage(req.Language) {
		log.Warn("Invalid language specified", zap.String("language", req.Language))

		return ErrInvalidLanguage
	}

	// Validate expiry time
	if req.ExpiryTime == "" {
		return ErrInvalidExpiry
	}

	expiryTimestamp, err := time.Parse(time.RFC3339, req.ExpiryTime)
	if err != nil {
		return ErrInvalidExpiry
	}

	if expiryTimestamp.Before(time.Now()) {
		return ErrExpiryInPast
	}

	// Additional security: Check if expiry is too far in the future (sanity check)
	maxFutureTime := time.Now().Add(time.Duration(MaxExpiryMinutes) * time.Minute)
	if expiryTimestamp.After(maxFutureTime) {
		return ErrExpiryTooFar
	}

	return nil
}

// Helper function to parse the expiry time.
func parseExpiryTime(expiryTimeStr string) time.Time {
	expiryTimestamp, _ := time.Parse(time.RFC3339, expiryTimeStr)

	return expiryTimestamp
}

// DatabaseHealthCheck performs a health check on the database connection.
func DatabaseHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := storage.HealthCheck(ctx)
	if err != nil {
		log.Error("Database health check failed", zap.Error(err))
		respondWithDetailedError(w, http.StatusServiceUnavailable,
			"Database health check failed", "DB_UNHEALTHY", err.Error())

		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"service":   "database",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
