package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

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
		if err == ErrPasteNotFound {
			respondWithError(w, http.StatusNotFound, "Paste not found")
		} else if err == ErrInvalidUUID {
			respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		} else {
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
		if err == ErrPasteNotFound {
			respondWithError(w, http.StatusNotFound, "Paste not found")
		} else if err == ErrInvalidUUID {
			respondWithError(w, http.StatusBadRequest, "Invalid UUID format")
		} else {
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

// Constants for paste constraints
const (
	MaxPasteSize     = 10 * 1024 * 1024 // 10MB
	MinExpiryMinutes = 1
	MaxExpiryMinutes = 60 * 24 * 365 // 1 year
)

// CreatePaste handles the creation of a new paste.
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

	if err := validateCreatePasteRequest(req); err != nil {
		log.Error("Error validating create paste request", zap.Error(err))
		if err == ErrEmptyContent {
			respondWithError(w, http.StatusBadRequest, "Content cannot be empty")
		} else if err == ErrContentTooLarge {
			respondWithError(w, http.StatusRequestEntityTooLarge, "Content exceeds maximum size")
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
		if err := storage.DBConn.WithContext(ctx).Delete(paste).Error; err != nil {
			log.Error("Error deleting paste after reading", zap.Error(err))
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// Helper function to delete a paste by its UUID.
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

// Helper function to validate the create paste request.
func validateCreatePasteRequest(req models.CreatePasteRequest) error {
	if req.Content == "" {
		return ErrEmptyContent
	}

	if len(req.Content) > MaxPasteSize {
		return ErrContentTooLarge
	}

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

	return nil
}

// Helper function to parse the expiry time.
func parseExpiryTime(expiryTimeStr string) time.Time {
	expiryTimestamp, _ := time.Parse(time.RFC3339, expiryTimeStr)
	return expiryTimestamp
}

// DatabaseHealthCheck performs a health check on the database connection
func DatabaseHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := storage.HealthCheck(ctx); err != nil {
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
