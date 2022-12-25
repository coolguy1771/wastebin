package handlers

import (
	"strconv"
	"time"

	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func GetRawPaste(c *fiber.Ctx) error {
	pasteUUID, err := uuid.Parse(c.Params("uuid"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(map[string]string{"error": err.Error()})
	}

	// Retrieve the paste from the database
	paste := models.Paste{}
	if err := storage.DBConn.First(&paste, "uuid = ?", pasteUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(map[string]string{"error": err.Error()})
	}

	// Check if the paste has expired
	if time.Now().After(paste.ExpiryTimestamp) {
		if err := storage.DBConn.Delete(&paste).Error; err != nil {
			log.Error("Error deleting expired paste from the database", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Error deleting expired paste from the database"})
		}
		return c.JSON(map[string]string{"message": "Paste expired and deleted"})
	}

	// Check if the paste should be deleted after reading
	if paste.Burn {
		if err := storage.DBConn.Delete(&paste).Error; err != nil {
			log.Error("Error deleting paste after reading", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Error deleting paste after reading"})
		}
	}

	// Set the Content-Type header to the appropriate MIME type for the paste's file extension
	c.Type("text/plain")

	// Send the raw paste as the response
	return c.SendString(paste.Content)
}

// GetPaste retrieves a paste by its UUID.
// If the paste has expired or is set to be deleted after reading, it is deleted from the database.
func GetPaste(c *fiber.Ctx) error {
	// Read the paste UUID from the URL parameter
	pasteUUID, err := uuid.Parse(c.Params("uuid"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(map[string]string{"error": err.Error()})
	}

	// Retrieve the paste from the database
	paste := models.Paste{}
	if err := storage.DBConn.First(&paste, "uuid = ?", pasteUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(map[string]string{"error": err.Error()})
	}

	// Check if the paste has expired
	if time.Now().After(paste.ExpiryTimestamp) {
		if err := storage.DBConn.Delete(&paste).Error; err != nil {
			log.Error("Error deleting expired paste from the database", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Error deleting expired paste from the database"})
		}
		return c.JSON(map[string]string{"message": "Paste expired and deleted"})
	}

	// Check if the paste should be deleted after reading
	if paste.Burn {
		if err := storage.DBConn.Delete(&paste).Error; err != nil {
			log.Error("Error deleting paste after reading", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": "Error deleting paste after reading"})
		}
	}

	// Return the paste content
	return c.JSON(paste)
}

func CreatePaste(c *fiber.Ctx) error {
	log.Info("CreatePaste called")
	// Parse the request body
	expireTime, err := strconv.ParseInt(c.FormValue("expires"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}
	req := models.CreatePasteRequest{
		Content:  c.FormValue("text"),
		Burn:     c.FormValue("burn") == "true",
		Language: c.FormValue("extension"),
		// Convert the expires value to an int64 and add it to the current time
		ExpiryTime: time.Now().Add(time.Duration(expireTime) * time.Minute).Format(time.RFC3339),
	}

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

	log.Debug("Paste request body has been validated", zap.Any("request", req))

	// Generate a UUID for the paste
	pasteUUID, err := uuid.NewRandom()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	log.Info("Generated UUID", zap.String("uuid", pasteUUID.String()))

	// Save the paste to the database
	paste := models.Paste{
		Content:         req.Content,
		Burn:            req.Burn,
		Language:        req.Language,
		UUID:            pasteUUID,
		ExpiryTimestamp: expiryTimestamp,
	}
	log.Debug("created paste object", zap.Any("paste", paste))

	if err := storage.DBConn.Create(&paste).Error; err != nil {
		log.Error("Error saving paste to database", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	log.Info("Paste saved to database")
	// Return the UUID of the newly created paste in the response body
	response := map[string]string{
		"message": "Paste created",
		"uuid":    pasteUUID.String(),
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
	if err := storage.DBConn.First(&paste, "uuid = ?", pasteUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	if err := storage.DBConn.Delete(&paste).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("Paste deleted")
}
