package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pavelc4/pixtify/internal/middleware"
)

func SetupRoutes(
	app *fiber.App,
	userHandler *UserHandler,
	oauthHandler *OAuthHandler,
	jwtMiddleware *middleware.JWTMiddleware,
) {
	// API routes
	api := app.Group("/api")

	// Public routes (No Auth Required)
	public := api.Group("")
	{
		// User registration & login
		public.Post("/users/register", userHandler.Register)
		public.Post("/users/login", userHandler.Login)

		// OAuth login redirects
		public.Get("/auth/github", oauthHandler.GithubLogin)
		public.Get("/auth/google", oauthHandler.GoogleLogin)

		// OAuth callbacks
		public.Get("/auth/github/callback", oauthHandler.GithubCallback)
		public.Get("/auth/google/callback", oauthHandler.GoogleCallback)

		// Refresh token endpoint
		public.Post("/auth/refresh", oauthHandler.RefreshToken)
		public.Post("/auth/logout", oauthHandler.Logout)
	}

	// Protected routes (Auth Required)
	protected := api.Group("", jwtMiddleware.Protected())
	{
		// Auth endpoints
		protected.Get("/auth/profile", oauthHandler.GetProfile)
		protected.Post("/auth/logout-all", oauthHandler.LogoutAll)

		// User endpoints
		protected.Get("/users/me", userHandler.GetCurrentUser)
		protected.Put("/users/me", userHandler.UpdateCurrentUser)
		protected.Delete("/users/me", userHandler.DeleteCurrentUser)

		// Public user profile
		protected.Get("/users/:id", userHandler.GetProfile)
	}

	// Admin routes (Admin Role Required)
	admin := api.Group("/admin", jwtMiddleware.Protected(), jwtMiddleware.RequireAdmin())
	{
		admin.Get("/users", userHandler.ListAllUsers)
		admin.Delete("/users/:id", userHandler.DeleteUser)
	}

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "pixtify-api",
		})
	})

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Route not found",
		})
	})
}
