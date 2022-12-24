package handlers

import (
	"time"

	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var db = storage.DBConn

// GetPaste retrieves a paste by its UUID.
// If the paste has expired or is set to be deleted after reading, it is deleted from the database.
func GetPaste(c *fiber.Ctx) error {
	// Read the paste UUID from the URL parameter
	pasteUUID, err := uuid.Parse(c.Params("uuid"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Paste not found")
	}

	// Retrieve the paste from the database
	paste := models.Paste{}
	if err := db.First(&paste, "uuid = ?", pasteUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Paste not found")
	}

	// Check if the paste has expired
	if time.Now().After(paste.ExpiryTimestamp) {
		if err := db.Delete(&paste).Error; err != nil {
			log.Error("Error deleting expired paste from the database", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting paste")
		}
		return c.SendString("Paste expired and deleted")
	}

	// Check if the paste should be deleted after reading
	if paste.Burn {
		if err := db.Delete(&paste).Error; err != nil {
			log.Error("Error deleting paste after reading", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).SendString("Error deleting paste")
		}
	}

	// Return the paste content
	return c.SendString(paste.Content)
}

func CreatePaste(c *fiber.Ctx) error {
	log.Info("CreatePaste called")
	// Parse the request body
	var req models.CreatePasteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}
	log.Info("CreatePaste request", zap.Any("request", req))
	if req.ExpiryTime == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Expiry time cannot be empty"})
	}
	// Parse the expiry time in the RFC 3339 format
	expiryTimestamp, err := time.Parse(time.RFC3339, req.ExpiryTime)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Invalid expiry time format"})
	}
	if expiryTimestamp.Before(time.Now()) {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Expiry time must be in the future"})
	}

	// Validate the other fields
	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Content cannot be empty"})
	}
	if !isValidLanguage(req.Language) {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": "Invalid language"})
	}
	log.Info("CreatePaste request validated")

	// Generate a UUID for the paste
	pasteUUID, err := uuid.NewRandom()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	log.Info("CreatePaste generated UUID", zap.String("uuid", pasteUUID.String()))

	// Save the paste to the database
	paste := models.Paste{
		Content:         req.Content,
		Burn:            req.Burn,
		Language:        req.Language,
		UUID:            pasteUUID,
		ExpiryTimestamp: expiryTimestamp,
	}
	log.Info("CreatePaste created paste", zap.Any("paste", paste))

	if err := db.Create(&paste).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	log.Info("CreatePaste saved paste to database")
	// Return the UUID of the newly created paste in the response body
	response := map[string]string{
		"status":  "success",
		"message": "Paste created with UUID: " + pasteUUID.String(),
	}
	return c.JSON(response)

}

func DeletePaste(c *fiber.Ctx) error {
	// Read the paste UUID from the URL query string
	pasteUUID, err := uuid.Parse(c.Query("uuid"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	// Delete the paste from the database
	var paste models.Paste
	if err := db.First(&paste, "uuid = ?", pasteUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())

	}
	if err := db.Delete(&paste).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("Paste deleted")
}

func isValidLanguage(language string) bool {
	supportedLanguages := map[string]struct{}{
		"plaintext":  {},
		"markdown":   {},
		"python":     {},
		"go":         {},
		"javascript": {},
		"html":       {},
		"css":        {},
	}
	_, ok := supportedLanguages[language]
	return ok
}
