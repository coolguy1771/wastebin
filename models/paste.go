package models

import (
	"time"

	"gorm.io/gorm"
)

// Paste is a model for a paste
type Paste struct {
	CreatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	PasteID       string         `json:"paste_id" gorm:"primary_key;not null"`
	PasteLanguage string         `json:"language" example:"go"`
	Paste         string         `json:"paste" example:"Paste A"`
	Author        string         `json:"author" example:"Dino"`
	Burn          bool           `json:"burn" example:"false"`
}
