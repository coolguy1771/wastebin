package main

import (
	"github.com/coolguy1771/wastebin/config"
	_ "github.com/coolguy1771/wastebin/docs"
	"github.com/coolguy1771/wastebin/models"
	"github.com/coolguy1771/wastebin/storage"

	"github.com/coolguy1771/wastebin/routes"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// @title Wastebin
// @version 1.0
// @description This is an API for Wastebin

// @contact.name Tyler Witlin

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @BasePath /api
func main() {

	config.Load()

	//env.SetupEnvFile()
	//// Create a new Zap logger
	logger, err := zap.NewProduction()

	if err != nil {
		logger.Sugar().Fatalf("can't initialize zap logger: %v", err)
	}
	//
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	// Create new fiber instance
	app := fiber.New(fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: false,
		ServerHeader:  "Fiber",
		AppName:       "Wastebin",
	})

	if err := storage.Connect(); err != nil {
		logger.Sugar().Fatal("Can't connect database:", err.Error())
	}

	storage.DBConn.AutoMigrate(&models.Paste{})

	// Load routes
	routes.AddRoutes(app)

	// Listen on the user specified port defaulting to 3000
	sugar.Fatal(app.Listen(":" + config.Conf.WebappPort))

}
