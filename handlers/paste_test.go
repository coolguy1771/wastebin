//go:build !integration
// +build !integration

package handlers_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coolguy1771/wastebin/handlers"
	"github.com/coolguy1771/wastebin/models"
)

// Unit tests for pure functions without external dependencies

func TestValidateCreatePasteRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     models.CreatePasteRequest
		expectError bool
		errorType   error
	}{
		{
			name: "Valid request",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				Burn:       false,
			},
			expectError: false,
		},
		{
			name: "Empty content",
			request: models.CreatePasteRequest{
				Content:    "",
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				Burn:       false,
			},
			expectError: true,
			errorType:   handlers.ErrEmptyContent,
		},
		{
			name: "Content too large",
			request: models.CreatePasteRequest{
				Content:    string(make([]byte, handlers.MaxPasteSize+1)),
				Language:   "txt",
				ExpiryTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				Burn:       false,
			},
			expectError: true,
			errorType:   handlers.ErrContentTooLarge,
		},
		{
			name: "Empty expiry time",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: "",
				Burn:       false,
			},
			expectError: true,
			errorType:   handlers.ErrInvalidExpiry,
		},
		{
			name: "Past expiry time",
			request: models.CreatePasteRequest{
				Content:    "Test content",
				Language:   "txt",
				ExpiryTime: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				Burn:       false,
			},
			expectError: true,
			errorType:   handlers.ErrExpiryInPast,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handlers.ValidateCreatePasteRequest(tt.request)

			if tt.expectError {
				require.Error(t, err)

				if tt.errorType != nil {
					assert.Equal(t, tt.errorType, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseExpiryTime(t *testing.T) {
	// Test valid RFC3339 timestamp
	validTime := time.Now().Add(1 * time.Hour)
	validTimeStr := validTime.Format(time.RFC3339)

	parsed := handlers.ParseExpiryTime(validTimeStr)
	assert.True(t, parsed.Equal(validTime.Truncate(time.Second)))

	// Test invalid timestamp - should return zero time
	parsed = handlers.ParseExpiryTime("invalid-time")
	assert.True(t, parsed.IsZero())
}

// Add more unit tests for other pure functions as they are extracted from handlers
