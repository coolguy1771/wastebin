package routes

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestApp initializes the Fiber app with routes for testing.
func setupTestApp() *fiber.App {
	app := fiber.New()
	AddRoutes(app)
	return app
}

// setupTestDB initializes a mock in-memory SQLite database for testing.
func setupTestDB() {
	// Use an in-memory SQLite database for testing.
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	storage.DBConn = db

	// Auto-migrate the Paste model for testing purposes.
	storage.DBConn.AutoMigrate(&models.Paste{})
}

// TestGetPaste tests the GET /api/v1/paste/:uuid endpoint.
func TestGetPaste(t *testing.T) {
	setupTestDB()
	app := setupTestApp()

	// Create a test paste in the database
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		UUID:            pasteUUID,
		Content:         "Test content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
	}
	storage.DBConn.Create(&testPaste)

	// Make a request to the GetPaste endpoint
	req := httptest.NewRequest("GET", "/api/v1/paste/"+pasteUUID.String(), nil)
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected status OK")
}

// TestCreatePaste tests the POST /api/v1/paste endpoint.
func TestCreatePaste(t *testing.T) {
	setupTestDB()
	app := setupTestApp()

	// Prepare form data
	form := `expires=10&text=New paste content&burn=false&extension=text`

	req := httptest.NewRequest("POST", "/api/v1/paste", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make a request to the CreatePaste endpoint
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected status OK")
}

// TestDeletePaste tests the DELETE /api/v1/paste/:uuid endpoint.
func TestDeletePaste(t *testing.T) {
	setupTestDB()
	app := setupTestApp()

	// Create a test paste in the database
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		UUID:            pasteUUID,
		Content:         "Test content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
	}
	storage.DBConn.Create(&testPaste)

	// Make a request to the DeletePaste endpoint
	req := httptest.NewRequest("DELETE", "/api/v1/paste/"+pasteUUID.String(), nil)
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected status OK")
}

// TestGetRawPaste tests the GET /paste/:uuid/raw endpoint.
func TestGetRawPaste(t *testing.T) {
	setupTestDB()
	app := setupTestApp()

	// Create a test paste in the database
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		UUID:            pasteUUID,
		Content:         "Test raw content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
	}
	storage.DBConn.Create(&testPaste)

	// Make a request to the GetRawPaste endpoint
	req := httptest.NewRequest("GET", "/paste/"+pasteUUID.String()+"/raw", nil)
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected status OK")
}

// TestServeSPA tests the static file serving for Single Page Application.
func TestServeSPA(t *testing.T) {
	app := fiber.New()

	// Simulate development mode
	config.Conf.Dev = true

	// Initialize routes
	AddRoutes(app)

	// Make a request to the root endpoint
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)

	// Assert the response
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Expected status OK in development mode")
}

// TestMiddlewareCors tests the CORS middleware.
func TestMiddlewareCors(t *testing.T) {
	app := setupTestApp()

	// Make an OPTIONS request to test CORS
	req := httptest.NewRequest("OPTIONS", "/api/v1/paste/test-uuid", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, _ := app.Test(req)

	// Assert CORS headers
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"), "Expected CORS header to be '*'")
	assert.Equal(t, "OPTIONS, GET, POST, DELETE", resp.Header.Get("Access-Control-Allow-Methods"), "Expected CORS methods")
}
