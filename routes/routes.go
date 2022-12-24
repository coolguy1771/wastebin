package routes

import (
	"github.com/coolguy1771/wastebin/config"
	"github.com/coolguy1771/wastebin/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// Add routes to the app
func AddRoutes(app *fiber.App) *fiber.App {

	app.Use(cors.New())

	app.Get("/swagger/*", swagger.HandlerDefault)
	api := app.Group("/api")
	v1 := api.Group("/v1", func(c *fiber.Ctx) error {
		c.JSON(fiber.Map{
			"message": "🐣 v1",
		})
		return c.Next()
	})

	v1.Get("/paste/:uuid", handlers.GetPaste)
	v1.Post("/paste", handlers.CreatePaste)
	v1.Delete("/paste/:iuud", handlers.DeletePaste)

	// Serve Single Page application
	app.Static("/", "/web/")

	if config.Conf.Dev {
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendFile("./web/build/index.html")
		})
		app.Get("/paste/*", func(c *fiber.Ctx) error {
			return c.SendFile("./web/build/index.html")
		})

		return app
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("/web/index.html")
	})
	app.Get("/paste/*", func(c *fiber.Ctx) error {
		return c.SendFile("/web/index.html")
	})

	return app
}
