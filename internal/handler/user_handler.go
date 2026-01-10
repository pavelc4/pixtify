package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/middleware"
	"github.com/pavelc4/pixtify/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
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

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	input.Username = middleware.SanitizeString(input.Username)
	input.Email = middleware.SanitizeEmail(input.Email)
	input.FullName = middleware.SanitizeString(input.FullName)

	if input.Username == "" || input.Email == "" || input.Password == "" {
		return badRequestError(c, "Username, email, and password are required")
	}

	if !middleware.IsValidUsername(input.Username) {
		return badRequestError(c, "Invalid username format (3-20 chars, alphanumeric, underscore, hyphen)")
	}

	if !middleware.IsValidEmail(input.Email) {
		return badRequestError(c, "Invalid email format")
	}

	if len(input.Password) < 8 {
		return badRequestError(c, "Password must be at least 8 characters")
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

	input.Email = middleware.SanitizeEmail(input.Email)

	if input.Email == "" || input.Password == "" {
		return badRequestError(c, "Email and password are required")
	}

	if !middleware.IsValidEmail(input.Email) {
		return badRequestError(c, "Invalid email format")
	}

	user, err := h.userService.Login(c.Context(), input.Email, input.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return unauthorizedError(c, "Invalid credentials")
		}
		return internalError(c, "Failed to login")
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user":    newUserResponse(user),
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	id := c.Params("id")

	id = middleware.SanitizeString(id)

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

	if input.FullName != nil {
		sanitized := middleware.SanitizeString(*input.FullName)
		input.FullName = &sanitized

		if len(*input.FullName) > 100 {
			return badRequestError(c, "Full name too long (max 100 characters)")
		}
	}

	if input.Bio != nil {
		sanitized := middleware.SanitizeString(*input.Bio)
		input.Bio = &sanitized

		if len(*input.Bio) > 500 {
			return badRequestError(c, "Bio too long (max 500 characters)")
		}
	}

	if input.AvatarURL != nil {
		sanitized := middleware.SanitizeString(*input.AvatarURL)
		input.AvatarURL = &sanitized

		if len(*input.AvatarURL) > 500 {
			return badRequestError(c, "Avatar URL too long")
		}
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
	// userID := c.Locals("user_id").(string)

	// TODO: Implement user deletion
	// For now, just return not implemented

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "User deletion not implemented yet",
	})
}

// ListAllUsers handles listing all users (admin only)
func (h *UserHandler) ListAllUsers(c *fiber.Ctx) error {
	// TODO: Implement user listing with pagination
	// For now, just return not implemented

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "User listing not implemented yet",
	})
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// id := c.Params("id")

	// TODO: Implement admin user deletion
	// For now, just return not implemented

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Admin user deletion not implemented yet",
	})
}
