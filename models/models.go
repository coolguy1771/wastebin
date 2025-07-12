package models

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreatePasteRequest represents the request payload for creating a new paste.
type CreatePasteRequest struct {
	Content    string `json:"content" validate:"required"`     // Content of the paste (required)
	Burn       bool   `json:"burn"`                            // Whether the paste should be deleted after reading
	Language   string `json:"language" validate:"required"`    // Language of the paste content (required)
	ExpiryTime string `json:"expiry_time" validate:"required"` // Expiry time in RFC3339 format (required)
}

// Paste represents a paste entity stored in the database.
type Paste struct {
	UUID            uuid.UUID `json:"paste_id" gorm:"type:uuid;primaryKey"`            // Unique identifier for the paste
	Content         string    `json:"content" example:"Paste A"`                       // Content of the paste
	Burn            bool      `json:"burn" example:"false"`                            // Burn-after-read flag
	Language        string    `json:"language" example:"go"`                           // Programming language or content type
	ExpiryTimestamp time.Time `json:"expiry_timestamp" example:"2021-01-01T00:00:00Z"` // Expiry timestamp for the paste
}

// DB represents the database connection with optional logging and retry support.
type DB struct {
	*gorm.DB             // Embedded GORM DB instance
	Logger   *zap.Logger // Logger instance for logging database operations
	DBName   string      // Name of the database
	Retries  int         // Number of retries for database operations
}
