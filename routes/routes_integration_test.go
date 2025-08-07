//go:build integration
// +build integration

package routes_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/routes"
	"github.com/coolguy1771/wastebin/storage"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger, err := log.New(os.Stdout, "ERROR")
	if err != nil {
		panic(err)
	}
	log.ResetDefault(logger)
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// setupTestApp initializes the Chi router with routes for testing.
func setupTestApp() http.Handler {
	return routes.AddRoutes(nil)
}

// setupTestDB initializes a mock in-memory SQLite database for testing.
func setupTestDB(t *testing.T) {
	t.Helper()
	
	// Set up test configuration
	config.Conf = config.Config{
		WebappPort:          "3000",
		DBMaxIdleConns:      5,
		DBMaxOpenConns:      10,
		DBPort:              5432,
		DBHost:              "localhost",
		DBUser:              "test",
		DBPassword:          "",
		DBName:              "test",
		Dev:                 true,
		TLSEnabled:          false,
		TLSCertFile:         "",
		TLSKeyFile:          "",
		AllowedOrigins:      "",
		RequireAuth:         false,
		AuthUsername:        "",
		AuthPassword:        "",
		CSRFKey:             "",
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		LogLevel:            "ERROR",
		LocalDB:             true,
		TracingEnabled:      false,
		MetricsEnabled:      false,
		ServiceName:         "",
		ServiceVersion:      "",
		Environment:         "",
		OTLPTraceEndpoint:   "",
		OTLPMetricsEndpoint: "",
		MetricsInterval:     0,
	}
	
	// Use an in-memory SQLite database for testing.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		SkipDefaultTransaction:                   false,
		DefaultTransactionTimeout:                0,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		PrepareStmtMaxSize:                       0,
		PrepareStmtTTL:                           0,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		IgnoreRelationshipsWhenMigrating:         false,
		DisableNestedTransaction:                 false,
		AllowGlobalUpdate:                        false,
		QueryFields:                              false,
		CreateBatchSize:                          0,
		TranslateError:                           false,
		PropagateUnscoped:                        false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	})
	require.NoError(t, err)

	//nolint:reassign // DBConn is reassigned for test setup.
	storage.DBConn = db
	// Auto-migrate the Paste model for testing purposes.
	require.NoError(t, storage.DBConn.AutoMigrate(&models.Paste{}))
}

func TestGetPaste(t *testing.T) {
	setupTestDB(t)

	app := setupTestApp()
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		UUID:            pasteUUID,
		Content:         "Test content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
	}
	require.NoError(t, storage.DBConn.Create(&testPaste).Error)
	// Make a request to the GetPaste endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/paste/"+pasteUUID.String(), http.NoBody)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}

func TestCreatePaste(t *testing.T) {
	setupTestDB(t)

	app := setupTestApp()
	form := `expires=10&text=New paste content&burn=false&extension=txt`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/paste", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status Created")
}

func TestDeletePaste(t *testing.T) {
	setupTestDB(t)

	app := setupTestApp()
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		UUID:            pasteUUID,
		Content:         "Test content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
	}
	require.NoError(t, storage.DBConn.Create(&testPaste).Error)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/paste/"+pasteUUID.String(), http.NoBody)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}

func TestGetRawPaste(t *testing.T) {
	setupTestDB(t)

	app := setupTestApp()
	pasteUUID := uuid.New()
	testPaste := models.Paste{
		Model:           gorm.Model{},
		UUID:            pasteUUID,
		Content:         "Test raw content",
		ExpiryTimestamp: time.Now().Add(10 * time.Minute),
		Burn:            false,
		Language:        "",
	}
	require.NoError(t, storage.DBConn.Create(&testPaste).Error)

	req := httptest.NewRequest(http.MethodGet, "/paste/"+pasteUUID.String()+"/raw", http.NoBody)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}

func TestServeSPA(t *testing.T) {
	setupTestDB(t)

	config.Conf.Dev = true
	app := routes.AddRoutes(nil)
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	assert.True(
		t,
		rr.Code == http.StatusOK || rr.Code == http.StatusNotFound,
		"Expected status OK or NotFound in test environment",
	)
}

func TestMiddlewareCors(t *testing.T) {
	setupTestDB(t)

	app := setupTestApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	// In dev mode, the allowed origin should be http://localhost:3000
	config.Conf.Dev = true
	assert.Equal(t, "http://localhost:3000", rr.Header().Get("Access-Control-Allow-Origin"), "Expected CORS header to match origin")
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status OK")
}