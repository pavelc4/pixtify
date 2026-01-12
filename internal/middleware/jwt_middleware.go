package middleware

import (
	"strings"

	"github.com/pavelc4/pixtify/internal/service"

	"github.com/gofiber/fiber/v2"
)

type JWTMiddleware struct {
	jwtService *service.JWTService
}

func NewJWTMiddleware(jwtService *service.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

func (m *JWTMiddleware) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		token = c.Cookies("access_token")

		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" {

				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func (m *JWTMiddleware) RequireOwner() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role == nil || role != "owner" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Owner access required",
			})
		}

		return c.Next()
	}
}

func (m *JWTMiddleware) RequireModerator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role == nil || role != "moderator" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Moderator access required",
			})
		}

		return c.Next()
	}
}

func (m *JWTMiddleware) RequireModeratorOrOwner() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role == nil || (role != "moderator" && role != "owner") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Moderator or Owner access required",
			})
		}

		return c.Next()
	}
}

func (m *JWTMiddleware) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("access_token")

		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token != "" {
			claims, err := m.jwtService.ValidateAccessToken(token)
			if err == nil {
				c.Locals("user_id", claims.UserID)
				c.Locals("email", claims.Email)
				c.Locals("role", claims.Role)
			}
		}

		return c.Next()
	}
}
