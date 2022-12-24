package models

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CreatePasteRequest struct {
	Content    string
	Burn       bool
	Language   string
	ExpiryTime string
}

type Paste struct {
	DeletedAt       gorm.DeletedAt `gorm:"index"`
	Content         string         `json:"content" example:"Paste A"`
	Burn            bool           `json:"burn" example:"false"`
	Language        string         `json:"language" example:"go"`
	UUID            uuid.UUID      `json:"paste_id"`
	ExpiryTimestamp time.Time      `json:"expiry_timestamp" example:"2021-01-01T00:00:00Z"`
}

type DB struct {
	*gorm.DB
	Logger  *zap.Logger
	DBName  string
	Retries int
}
