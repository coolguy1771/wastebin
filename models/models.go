package models

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreatePasteRequest represents the request payload for creating a new paste.
type CreatePasteRequest struct {
	Content    string `json:"content"    validate:"required"` // Content of the paste (required)
	Burn       bool   `json:"burn"`                           // Whether the paste should be deleted after reading
	Language   string `json:"language"   validate:"required"` // Language of the paste content (required)
	ExpiryTime string `json:"expiryTime" validate:"required"` // Expiry time in RFC3339 format (required)
}

// Paste represents a paste entity stored in the database.
type Paste struct {
	gorm.Model

	UUID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"pasteId"`
	Content         string    `                            json:"content"`
	Burn            bool      `                            json:"burn"`
	Language        string    `                            json:"language"`
	ExpiryTimestamp time.Time `                            json:"expiryTimestamp"`
}

// DB represents the database connection with optional logging and retry support.
type DB struct {
	*gorm.DB // Embedded GORM DB instance

	Logger  *zap.Logger // Logger instance for logging database operations
	DBName  string      // Name of the database
	Retries int         // Number of retries for database operations
}
