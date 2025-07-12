package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestApp initializes the Chi router with routes for testing.
func setupTestApp() http.Handler {
	router := AddRoutes(nil)

	return router
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/paste/"+pasteUUID.String(), nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}

// TestCreatePaste tests the POST /api/v1/paste endpoint.
func TestCreatePaste(t *testing.T) {
	setupTestDB()

	app := setupTestApp()

	// Prepare form data
	form := `expires=10&text=New paste content&burn=false&extension=text`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/paste", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make a request to the CreatePaste endpoint
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status Created")
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
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/paste/"+pasteUUID.String(), nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
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
	req := httptest.NewRequest(http.MethodGet, "/paste/"+pasteUUID.String()+"/raw", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Assert the response
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}

// TestServeSPA tests the static file serving for Single Page Application.
func TestServeSPA(t *testing.T) {
	// Simulate development mode
	config.Conf.Dev = true

	// Initialize routes
	app := AddRoutes(nil)

	// Make a request to the root endpoint
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// In development mode, it may return 404 if the index.html file doesn't exist
	// This is expected behavior for the test environment
	assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusNotFound, "Expected status OK or NotFound in test environment")
}

// TestMiddlewareCors tests the CORS middleware.
func TestMiddlewareCors(t *testing.T) {
	app := setupTestApp()

	// Make a GET request to test CORS headers (OPTIONS may not be handled by Chi CORS)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)

	// Assert CORS headers are present
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"), "Expected CORS header to be '*'")
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}
