package handler

import "github.com/gofiber/fiber/v2"

func SetupRoutes(app *fiber.App, userHandler *UserHandler, oauthHandler *OAuthHandler) {
	// API routes
	api := app.Group("/api")

	// User routes
	users := api.Group("/users")
	users.Post("/register", userHandler.Register)
	users.Post("/login", userHandler.Login)
	users.Get("/:id", userHandler.GetProfile)

	// OAuth Github
	auth := api.Group("/auth")
	auth.Get("/github", oauthHandler.GithubLogin)
	auth.Get("/github/callback", oauthHandler.GithubCallback)

	// OAuth Google
	auth.Get("/google", oauthHandler.GoogleLogin)
	auth.Get("/google/callback", oauthHandler.GoogleCallback)
}
