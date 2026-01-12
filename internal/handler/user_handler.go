package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	jwtService  *service.JWTService
}

func NewUserHandler(userService *service.UserService, jwtService *service.JWTService) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtService:  jwtService,
	}
}

type UserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	FullName  string  `json:"full_name,omitempty"`
	Role      string  `json:"role"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Register handles user registration
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	user, err := h.userService.Register(c.Context(), input)
	if err != nil {
		if err == service.ErrUserExists {
			return conflictError(c, "User already exists")
		}
		return internalError(c, "Failed to register user")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    newUserResponse(user),
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	user, err := h.userService.Login(c.Context(), input.Email, input.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return unauthorizedError(c, "Invalid credentials")
		}
		return internalError(c, "Failed to login")
	}

	// Generate JWT access token
	accessToken, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return internalError(c, "Failed to generate access token")
	}

	return c.JSON(fiber.Map{
		"message":      "Login successful",
		"user":         newUserResponse(user),
		"access_token": accessToken,
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	id := c.Params("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		return badRequestError(c, "Invalid user ID")
	}

	user, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			return notFoundError(c, "User not found")
		}
		return internalError(c, "Failed to get user profile")
	}

	return c.JSON(fiber.Map{
		"user": newUserResponse(user),
	})
}

func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil || user == nil {
		return notFoundError(c, "User not found")
	}

	return c.JSON(fiber.Map{
		"user": newUserResponse(user),
	})
}

func (h *UserHandler) UpdateCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var input struct {
		FullName  *string `json:"full_name"`
		Bio       *string `json:"bio"`
		AvatarURL *string `json:"avatar_url"`
	}

	if err := c.BodyParser(&input); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return badRequestError(c, "Invalid user ID")
	}

	user, err := h.userService.UpdateProfile(c.Context(), uid, input.FullName, input.Bio, input.AvatarURL)
	if err != nil {
		if err == service.ErrUserNotFound {
			return notFoundError(c, "User not found")
		}
		return internalError(c, "Failed to update profile")
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user":    newUserResponse(user),
	})
}

func (h *UserHandler) DeleteCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	uid, err := uuid.Parse(userID)
	if err != nil {
		return badRequestError(c, "Invalid user ID")
	}

	if err := h.userService.DeleteUser(c.Context(), uid); err != nil {
		if err == service.ErrUserNotFound {
			return notFoundError(c, "User not found")
		}
		return internalError(c, "Failed to delete user")
	}

	return c.JSON(fiber.Map{
		"message": "User account deleted successfully",
	})
}

// ListAllUsers handles listing all users (admin only)
func (h *UserHandler) ListAllUsers(c *fiber.Ctx) error {
	// Get pagination params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := h.userService.ListUsers(c.Context(), page, limit)
	if err != nil {
		return internalError(c, "Failed to list users")
	}

	// Convert to response format
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = newUserResponse(user)
	}

	return c.JSON(fiber.Map{
		"users": userResponses,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		return badRequestError(c, "Invalid user ID")
	}

	if err := h.userService.DeleteUser(c.Context(), userID); err != nil {
		if err == service.ErrUserNotFound {
			return notFoundError(c, "User not found")
		}
		return internalError(c, "Failed to delete user")
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
