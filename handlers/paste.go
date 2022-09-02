package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/gofiber/fiber/v2"
)

// ResponseHTTP represents response body of this API
type ResponseHTTP struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// GetPasteByID is a function to get a paste by ID
// @Summary Get paste by ID
// @Description Get paste by ID
// @Tags pastes
// @Accept json
// @Produce json
// @Param id path int true "Paste ID"
// @Success 200 {object} ResponseHTTP{data=[]models.Paste}
// @Failure 404 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/paste/{id} [get]
func GetPasteByID(c *fiber.Ctx) error {
	id := c.Params("id")
	db := storage.DBConn

	paste := new(models.Paste)
	if err := db.Where("paste_ID = ?", id).First(&paste).Error; err != nil {
		switch err.Error() {
		case "record not found":
			return c.Status(http.StatusNotFound).JSON(ResponseHTTP{
				Success: false,
				Message: fmt.Sprintf("Paste with ID %v not found.", id),
				Data:    nil,
			})
		default:
			return c.Status(http.StatusServiceUnavailable).JSON(ResponseHTTP{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			})

		}
	}

	return c.JSON(ResponseHTTP{
		Success: true,
		Message: "Success get paste by ID.",
		Data:    *paste,
	})
}

// CreatePaste registers a new paste data
// @Summary Creates a new paste
// @Description Crate paste
// @Tags pastes
// @Accept json
// @Produce json
// @Param paste body models.Paste true "Create paste"
// @Success 200 {object} ResponseHTTP{data=models.Paste}
// @Failure 400 {object} ResponseHTTP{}
// @Router /v1/paste [post]
func CreatePaste(c *fiber.Ctx) error {
	db := storage.DBConn

	paste := new(models.Paste)
	if err := c.BodyParser(&paste); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ResponseHTTP{
			Success: false,
			Message: err.Error(),
			Data:    nil,
		})
	}
	paste.PasteID = RandomString(8)

	db.Create(paste)

	return c.JSON(ResponseHTTP{
		Success: true,
		Message: "Success registered a paste.",
		Data:    *paste,
	})
}

// DeletePaste function removes a paste by ID
// @Summary Remove paste by ID
// @Description Remove paste by ID
// @Tags pastes
// @Accept json
// @Produce json
// @Param id path string true "Paste ID"
// @Success 200 {object} ResponseHTTP{}
// @Failure 404 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/paste/{id} [delete]
func DeletePaste(c *fiber.Ctx) error {
	id := c.Params("id")
	db := storage.DBConn

	paste := new(models.Paste)
	if err := db.First(&paste, id).Error; err != nil {
		switch err.Error() {
		case "record not found":
			return c.Status(http.StatusNotFound).JSON(ResponseHTTP{
				Success: false,
				Message: fmt.Sprintf("Paste with ID %v not found.", id),
				Data:    nil,
			})
		default:
			return c.Status(http.StatusServiceUnavailable).JSON(ResponseHTTP{
				Success: false,
				Message: err.Error(),
				Data:    nil,
			})

		}
	}

	db.Delete(&paste)

	return c.JSON(ResponseHTTP{
		Success: true,
		Message: "Success deleted paste.",
		Data:    nil,
	})
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

