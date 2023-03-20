package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coolguy1771/wastebin/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var db *gorm.DB

// GetPaste retrieves a paste by its UUID.
// If the paste has expired or is set to be deleted after reading, it is deleted from the database.
func GetPaste(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get paste ID from URL parameter
		uuid := chi.URLParam(r, "uuid")

		// Query paste by ID
		var paste models.Paste
		if err := db.First(&paste, uuid).Error; err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Check if paste is expired
		if time.Now().Sub(paste.ExpiryTimestamp) > 1 {
			http.Error(w, "Paste not found", http.StatusNotFound)
			DeletePaste(paste)
			return
		}

		// Write paste as response
		if err := json.NewEncoder(w).Encode(paste); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func CreatePaste(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		paste := &models.Paste{}

		if err := json.NewDecoder(r.Body).Decode(paste); err != nil {
			http.Error(w, "Invalid paste data", http.StatusBadRequest)
			return
		}

		if err := validatePaste(paste); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Set UUID
		paste.UUID = uuid.New()

		// Save paste to database
		if err := db.Create(&paste).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write paste ID as response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(paste)
	}
}

func DeletePaste(paste models.Paste) error {
	if err := db.Where("uuid = ?", paste.UUID).First(&paste).Error; err != nil {
		return err
	}

	if err := db.Delete(&paste).Error; err != nil {
		// http.Error(w, "Failed to delete paste", http.StatusInternalServerError)
		return err
	}

	return nil
}

func validatePaste(paste *models.Paste) error {
	// Check if content is not empty
	if len(strings.TrimSpace(paste.Content)) == 0 {
		return errors.New("content cannot be empty")
	}

	// Check if expiration is not a past date
	if paste.ExpiryTimestamp.Before(time.Now()) {
		return errors.New("expiration date cannot be in the past")
	}

	return nil
}
