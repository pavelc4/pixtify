package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pavelc4/pixtify/internal/middleware"
)

func SetupRoutes(
	app *fiber.App,
	userHandler *UserHandler,
	oauthHandler *OAuthHandler,
	reportHandler *ReportHandler,
	wallpaperHandler *WallpaperHandler,
	collectionHandler *CollectionHandler,
	jwtMiddleware *middleware.JWTMiddleware,
	rateLimiter *middleware.RateLimiterMiddleware,
) {
	// API routes
	api := app.Group("/api")

	// Apply general API
	api.Use(rateLimiter.APILimiter())

	// Public routes
	public := api.Group("")
	{
		public.Post("/users/register", rateLimiter.RegisterLimiter(), userHandler.Register)
		public.Post("/users/login", rateLimiter.LoginLimiter(), userHandler.Login)

		// OAuth
		public.Get("/auth/github", rateLimiter.OAuthLimiter(), oauthHandler.GithubLogin)
		public.Get("/auth/google", rateLimiter.OAuthLimiter(), oauthHandler.GoogleLogin)
		public.Get("/auth/github/callback", rateLimiter.OAuthLimiter(), oauthHandler.GithubCallback)
		public.Get("/auth/google/callback", rateLimiter.OAuthLimiter(), oauthHandler.GoogleCallback)

		// Token
		public.Post("/auth/refresh", oauthHandler.RefreshToken)
		public.Post("/auth/logout", oauthHandler.Logout)

		// Public Wallpaper Routes
		public.Get("/wallpapers", wallpaperHandler.ListWallpapers)
		public.Get("/wallpapers/:id", wallpaperHandler.GetWallpaper)

		// Public Collection Routes
		public.Get("/collections/:id", collectionHandler.GetCollectionById)
		public.Get("/collections/:id/wallpapers", collectionHandler.GetCollectionWallpapers)
	}

	// Protected routes (Auth Required)
	protected := api.Group("", jwtMiddleware.Protected())
	{
		// Auth endpoints
		protected.Get("/auth/profile", oauthHandler.GetProfile)
		protected.Post("/auth/logout-all", oauthHandler.LogoutAll)

		// User endpoints
		protected.Get("/users", userHandler.ListAllUsers)
		protected.Get("/users/me", userHandler.GetCurrentUser)
		protected.Put("/users/me", userHandler.UpdateCurrentUser)
		protected.Delete("/users/me", userHandler.DeleteCurrentUser)

		// Public user profile
		protected.Get("/users/:id", userHandler.GetProfile)

		// REPORTS
		protected.Post("/reports", reportHandler.CreateReport)

		// WALLPAPERS (Protected)
		protected.Post("/wallpapers", wallpaperHandler.UploadWallpaper)

		// LIKES
		protected.Post("/wallpapers/:id/like", wallpaperHandler.LikeWallpaper)
		protected.Get("/users/me/liked-wallpapers", wallpaperHandler.GetMyLikes)

		// COLLECTIONS
		protected.Post("/collections", collectionHandler.CreateCollection)
		protected.Get("/collections/me", collectionHandler.GetMyCollections)
		protected.Post("/collections/:id/wallpapers", collectionHandler.AddWallpaperToCollection)
		protected.Delete("/collections/:id/wallpapers/:wallpaperId", collectionHandler.RemoveWallpaperFromCollection)
		protected.Delete("/collections/:id", collectionHandler.DeleteCollection)

		// Moderator/Owner routes (content moderation - no separate dashboard)
		moderator := protected.Group("", jwtMiddleware.RequireModeratorOrOwner())
		{
			moderator.Post("/users/:id/ban", userHandler.BanUser)
			moderator.Post("/users/:id/unban", userHandler.UnbanUser)
			moderator.Get("/users/:id/stats", userHandler.GetUserStats)
		}
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
