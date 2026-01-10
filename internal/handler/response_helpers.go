package handler

import (
	"github.com/gofiber/fiber/v2"
	userRepo "github.com/pavelc4/pixtify/internal/repository/postgres/user"
)

func errorResponse(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

func badRequestError(c *fiber.Ctx, message string) error {
	return errorResponse(c, fiber.StatusBadRequest, message)
}

func unauthorizedError(c *fiber.Ctx, message string) error {
	return errorResponse(c, fiber.StatusUnauthorized, message)
}

func notFoundError(c *fiber.Ctx, message string) error {
	return errorResponse(c, fiber.StatusNotFound, message)
}

func conflictError(c *fiber.Ctx, message string) error {
	return errorResponse(c, fiber.StatusConflict, message)
}

func internalError(c *fiber.Ctx, message string) error {
	return errorResponse(c, fiber.StatusInternalServerError, message)
}

func successResponse(c *fiber.Ctx, message string, data interface{}) error {
	return c.JSON(fiber.Map{
		"message": message,
		"data":    data,
	})
}

func newUserResponse(user *userRepo.User) UserResponse {
	fullName := ""
	if user.FullName != nil {
		fullName = *user.FullName
	}

	return UserResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Email:     user.Email,
		FullName:  fullName,
		Role:      user.Role,
		Bio:       user.Bio,
		AvatarURL: user.AvatarURL,
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
