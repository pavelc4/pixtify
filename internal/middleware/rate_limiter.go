package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/pavelc4/pixtify/internal/config"
)

type RateLimiterMiddleware struct {
	config config.RateLimitConfig
}

func NewRateLimiterMiddleware(cfg config.RateLimitConfig) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		config: cfg,
	}
}

func (m *RateLimiterMiddleware) LoginLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        m.config.LoginMax,
		Expiration: m.config.LoginWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many login attempts. Please try again later.",
				"retry_after": m.config.LoginWindow.Seconds(),
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
	})
}

func (m *RateLimiterMiddleware) RegisterLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        m.config.RegisterMax,
		Expiration: m.config.RegisterWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many registration attempts. Please try again later.",
				"retry_after": m.config.RegisterWindow.Seconds(),
			})
		},
	})
}

func (m *RateLimiterMiddleware) OAuthLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        m.config.OAuthMax,
		Expiration: m.config.OAuthWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Too many authentication attempts. Please try again later.",
				"retry_after": m.config.OAuthWindow.Seconds(),
			})
		},
	})
}

func (m *RateLimiterMiddleware) APILimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        m.config.APIMax,
		Expiration: m.config.APIWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded. Please slow down.",
				"retry_after": m.config.APIWindow.Seconds(),
			})
		},
	})
}

func (m *RateLimiterMiddleware) AdminLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        m.config.AdminMax,
		Expiration: m.config.AdminWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Admin rate limit exceeded.",
				"retry_after": m.config.AdminWindow.Seconds(),
			})
		},
	})
}

// Custom limiter - for specific use cases
func (m *RateLimiterMiddleware) CustomLimiter(max int, window time.Duration, message string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       message,
				"retry_after": window.Seconds(),
			})
		},
	})
}
