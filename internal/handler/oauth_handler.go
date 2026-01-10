package handler

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pavelc4/pixtify/internal/service"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
	userService  *service.UserService
}

func NewOAuthHandler(oauthService *service.OAuthService, userService *service.UserService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		userService:  userService,
	}
}

func (h *OAuthHandler) GithubLogin(c *fiber.Ctx) error {

	state := "random_state_string" // TODO: Generate secure random state

	url := h.oauthService.GetGithubAuthURL(state)
	return c.Redirect(url)
}

func (h *OAuthHandler) GithubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing authorization code",
		})
	}

	githubUser, err := h.oauthService.HandleGithubCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authenticate with GitHub",
		})
	}

	if githubUser.Email == "" {
		githubUser.Email = fmt.Sprintf("%s@github.local", githubUser.Login)
	}

	user, err := h.userService.GetByEmail(c.Context(), githubUser.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if user == nil {
		registerInput := service.RegisterInput{
			Username: githubUser.Login,
			Email:    githubUser.Email,
			Password: "oauth_user", // Random password for OAuth users
			FullName: githubUser.Name,
		}

		user, err = h.userService.Register(c.Context(), registerInput)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}
	}

	// TODO: Generate JWT token here

	return c.JSON(fiber.Map{
		"message": "GitHub authentication successful",
		"user": UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	})
}

func (h *OAuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state := "random_state_string"
	url := h.oauthService.GetGoogleAuthURL(state)
	return c.Redirect(url)
}

func (h *OAuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing authorization code",
		})
	}

	googleUser, err := h.oauthService.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authenticate with Google",
		})
	}

	user, err := h.userService.GetByEmail(c.Context(), googleUser.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if user == nil {
		username := strings.Split(googleUser.Email, "@")[0]

		registerInput := service.RegisterInput{
			Username: username,
			Email:    googleUser.Email,
			Password: "oauth_user",
			FullName: googleUser.Name,
		}

		user, err = h.userService.Register(c.Context(), registerInput)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Google authentication successful",
		"user": UserResponse{
			ID:       user.ID.String(),
			Username: user.Username,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     user.Role,
		},
	})
}
