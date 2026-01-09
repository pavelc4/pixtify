package main

import (
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/pavelc4/pixtify/internal/config"
	"github.com/pavelc4/pixtify/internal/repository/postgres"
)

func main() {
	envPath := ".env"
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Could not load %s: %v", envPath, err)
	}

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

	app := fiber.New(fiber.Config{
		AppName: "Pixtify API",
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status":   "error",
				"app":      "Pixtify API",
				"database": "disconnected",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"app":      "Pixtify API",
			"env":      cfg.Env,
			"database": "connected",
		})
	})

	log.Printf("Pixtify API starting on port %s (environment: %s)", cfg.Port, cfg.Env)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
