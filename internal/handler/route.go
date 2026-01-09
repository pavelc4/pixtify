package handler

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App, userHandler *UserHandler) {
	api := app.Group("/api")

	users := api.Group("/users")
	users.Post("/register", userHandler.Register)
	users.Post("/login", userHandler.Login)
	users.Get("/:id", userHandler.GetProfile)
}
