package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/pavelc4/pixtify/internal/config"
	"github.com/pavelc4/pixtify/internal/handler"
	"github.com/pavelc4/pixtify/internal/middleware"
	"github.com/pavelc4/pixtify/internal/processor"
	"github.com/pavelc4/pixtify/internal/repository"
	"github.com/pavelc4/pixtify/internal/repository/postgres"
	"github.com/pavelc4/pixtify/internal/repository/postgres/user"
	"github.com/pavelc4/pixtify/internal/repository/postgres/wallpaper"
	"github.com/pavelc4/pixtify/internal/service"
	"github.com/pavelc4/pixtify/internal/storage"
)

func main() {
	cfg := config.Load()
	log.Println("Configuration loaded")

	db, err := postgres.NewPostgresDB(cfg.Database.GetDSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	log.Println("Database connected")

	accessExpiry, err := time.ParseDuration(cfg.JWT.AccessExpiry)
	if err != nil {
		log.Fatal("Invalid JWT_ACCESS_EXPIRY:", err)
	}

	refreshExpiry, err := time.ParseDuration(cfg.JWT.RefreshExpiry)
	if err != nil {
		log.Fatal("Invalid JWT_REFRESH_EXPIRY:", err)
	}

	userRepo := user.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	log.Println("Repositories initialized")

	userService := service.NewUserService(userRepo)
	oauthService := service.NewOAuthService(cfg.OAuth)
	jwtService := service.NewJWTService(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		accessExpiry,
		refreshExpiry,
	)
	log.Println("Services initialized")

	userHandler := handler.NewUserHandler(userService, jwtService)
	oauthHandler := handler.NewOAuthHandler(
		oauthService,
		userService,
		jwtService,
		refreshTokenRepo,
		cfg.CookieSecret,
	)
	log.Println("Handlers initialized")

	reportRepo := repository.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService, userService)
	log.Println("Report handlers initialized")

	// Wallpaper System
	imageProcessor := processor.NewImageProcessor()

	// MinIO
	minioStorage, err := storage.NewMinIOStorage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.CDNURL,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize MinIO storage: %v. Wallpaper uploads will fail.", err)
	} else {
		// buckets
		buckets := []string{cfg.Storage.BucketOriginals, cfg.Storage.BucketThumbnails}
		err = minioStorage.InitializeBuckets(context.Background(), buckets)
		if err != nil {
			log.Printf("Warning: Failed to initialize MinIO buckets: %v", err)
		} else {
			log.Println("MinIO buckets initialized and policies set")
		}
	}

	wallpaperRepo := wallpaper.NewRepository(db)
	wallpaperService := service.NewWallpaperService(wallpaperRepo, minioStorage, imageProcessor)
	wallpaperHandler := handler.NewWallpaperHandler(wallpaperService)
	log.Println("Wallpaper system initialized")

	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)
	rateLimitConfig := config.DefaultRateLimitConfig()
	rateLimiter := middleware.NewRateLimiterMiddleware(rateLimitConfig)
	log.Println("Middleware initialized")

	app := fiber.New(fiber.Config{
		AppName:      "Pixtify API",
		ServerHeader: "Pixtify",
		ErrorHandler: customErrorHandler,
		BodyLimit:    100 * 1024 * 1024,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	handler.SetupRoutes(app, userHandler, oauthHandler, reportHandler, wallpaperHandler, jwtMiddleware, rateLimiter)
	log.Println("Routes configured")

	port := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Env)
	log.Printf("Access token expiry: %s", accessExpiry)
	log.Printf("Refresh token expiry: %s", refreshExpiry)

	log.Printf("Rate limits configured:")
	log.Printf("  - Login: %d req/%s", rateLimitConfig.LoginMax, rateLimitConfig.LoginWindow)
	log.Printf("  - Register: %d req/%s", rateLimitConfig.RegisterMax, rateLimitConfig.RegisterWindow)
	log.Printf("  - OAuth: %d req/%s", rateLimitConfig.OAuthMax, rateLimitConfig.OAuthWindow)
	log.Printf("  - API: %d req/%s", rateLimitConfig.APIMax, rateLimitConfig.APIWindow)

	if err := app.Listen(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
