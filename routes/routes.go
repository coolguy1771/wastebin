package routes

import (
	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Add routes to the app
func AddRoutes(app *fiber.App) *fiber.App {
	app.Use(cors.New())

	api := app.Group("/api")
	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.JSON(fiber.Map{
			"message": "🐣 v1",
		})
		return c.Next()
	})

	v1.Get("/paste/:uuid", handlers.GetPaste)
	v1.Post("/paste", handlers.CreatePaste)
	v1.Delete("/paste/:uuid", handlers.DeletePaste)

	// Serve Single Page application
	if config.Conf.Dev {
		app.Static("/", "./web/build/")
	} else {
		app.Static("/", "/web/")
	}

	app.Get("/", serveSPA)
	app.Get("/paste/:uuid", serveSPA)
	app.Get("/paste/:uuid/raw", handlers.GetRawPaste)

	return app
}

func serveSPA(c *fiber.Ctx) error {
	if config.Conf.Dev {
		return c.SendFile("./web/build/index.html")
	} else {
		return c.SendFile("/web/index.html")
	}
}
