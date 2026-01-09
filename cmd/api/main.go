package main

import (
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/pavelc4/pixtify/internal/config"
	"github.com/pavelc4/pixtify/internal/handler"
	"github.com/pavelc4/pixtify/internal/repository/postgres"
	userRepo "github.com/pavelc4/pixtify/internal/repository/postgres/user"
	"github.com/pavelc4/pixtify/internal/service"
)

func main() {
	cfg := config.Load()

	dbConfig := postgres.DBConfig{
		DSN:             cfg.Database.GetDSN(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer postgres.CloseDatabase(db)

	userRepository := userRepo.NewRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	app := fiber.New(fiber.Config{
		AppName: "Pixtify API",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "error",
				"database": "disconnected",
			})
		}
		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
		})
	})

	handler.SetupRoutes(app, userHandler)

	log.Printf("Pixtify API starting on port %s (environment: %s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
