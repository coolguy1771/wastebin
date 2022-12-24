package main

import (
	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/log"
	"github.com/coolguy1771/wastebin/storage"

	"github.com/coolguy1771/wastebin/routes"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {

	config.Load()

	_, err := storage.Connect()
	if err != nil {
		log.Fatal("Error connecting to the database", zap.Error(err))
	}

	//TODO fix this
	//defer db.Close()

	// Create new fiber instance
	app := fiber.New(fiber.Config{
		Prefork:               false,
		CaseSensitive:         true,
		StrictRouting:         false,
		ServerHeader:          "Fiber",
		AppName:               "Wastebin",
		DisableStartupMessage: true,
	})

	// Load routes
	routes.AddRoutes(app)

	log.Info("Starting the server", zap.String("port", config.Conf.WebappPort))

	// Listen on the user specified port defaulting to 3000
	if err := app.Listen(":" + config.Conf.WebappPort); err != nil {
		log.Fatal("Error starting the server", zap.Error(err))
	}
}
