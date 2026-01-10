package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

// Register handles user registration
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userService.Register(c.Context(), input)
	if err != nil {
		if err == service.ErrUserExists {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "User already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := h.userService.Login(c.Context(), input.Email, input.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to login",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user": UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	})
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	id := c.Params("id")

	userID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user profile",
		})
	}

	return c.JSON(fiber.Map{
		"user": UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  getStringValue(user.FullName),
			Role:      user.Role,
			Bio:       user.Bio,
			AvatarURL: user.AvatarURL,
		},
	})
}

func (h *UserHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(fiber.Map{
		"user": UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  getStringValue(user.FullName),
			Role:      user.Role,
			Bio:       user.Bio,
			AvatarURL: user.AvatarURL,
		},
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := h.userService.UpdateProfile(c.Context(), uid, input.FullName, input.Bio, input.AvatarURL)
	if err != nil {
		if err == service.ErrUserNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
		"user": UserResponse{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  getStringValue(user.FullName),
			Role:      user.Role,
			Bio:       user.Bio,
			AvatarURL: user.AvatarURL,
		},
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

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
