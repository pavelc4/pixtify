package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pavelc4/pixtify/internal/repository"
	"github.com/pavelc4/pixtify/internal/service"
	"github.com/pavelc4/pixtify/internal/utils"
)

type OAuthHandler struct {
	oauthService     *service.OAuthService
	userService      *service.UserService
	jwtService       *service.JWTService
	refreshTokenRepo *repository.RefreshTokenRepository
	cookieSecret     string
}

func NewOAuthHandler(
	oauthService *service.OAuthService,
	userService *service.UserService,
	jwtService *service.JWTService,
	refreshTokenRepo *repository.RefreshTokenRepository,
	cookieSecret string,
) *OAuthHandler {
	return &OAuthHandler{
		oauthService:     oauthService,
		userService:      userService,
		jwtService:       jwtService,
		refreshTokenRepo: refreshTokenRepo,
		cookieSecret:     cookieSecret,
	}
}

func (h *OAuthHandler) GithubLogin(c *fiber.Ctx) error {
	state, err := utils.GenerateRandomState()
	if err != nil {
		return internalError(c, "Failed to generate state")
	}

	signedState := h.signState(state)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    signedState,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		MaxAge:   600,
	})

	url := h.oauthService.GetGithubAuthURL(state)
	return c.Redirect(url)
}

func (h *OAuthHandler) GithubCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	if state == "" {
		return badRequestError(c, "Authentication failed. Please try again.")
	}

	storedState := c.Cookies("oauth_state")
	if storedState == "" {
		return badRequestError(c, "Authentication failed. Please try logging in again.")
	}

	if !h.verifyState(state, storedState) {
		return badRequestError(c, "Unable to connect to GitHub. Please try again.")
	}

	c.Cookie(&fiber.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
	})

	code := c.Query("code")
	if code == "" {
		return badRequestError(c, "Missing authorization code")
	}

	githubUser, err := h.oauthService.HandleGithubCallback(c.Context(), code)
	if err != nil {
		return internalError(c, "Failed to authenticate with GitHub")
	}

	if githubUser.Email == "" {
		githubUser.Email = fmt.Sprintf("%s@github.local", githubUser.Login)
	}

	user, err := h.userService.GetByEmail(c.Context(), githubUser.Email)
	if err != nil {
		return internalError(c, "Database error")
	}

	if user == nil {
		registerInput := service.RegisterInput{
			Username: githubUser.Login,
			Email:    githubUser.Email,
			Password: "oauth_user",
			FullName: githubUser.Name,
		}

		user, err = h.userService.Register(c.Context(), registerInput)
		if err != nil {
			return internalError(c, "Failed to create user")
		}
	}

	accessToken, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return internalError(c, "Failed to generate access token")
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return internalError(c, "Failed to generate refresh token")
	}

	expiresAt := time.Now().Add(h.jwtService.GetRefreshExpiry())
	if err := h.refreshTokenRepo.Store(c.Context(), user.ID.String(), refreshToken, expiresAt); err != nil {
		return internalError(c, "Failed to store refresh token")
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.JSON(fiber.Map{
		"message": "GitHub authentication successful",
		"user":    newUserResponse(user),
	})
}

func (h *OAuthHandler) GoogleLogin(c *fiber.Ctx) error {
	state, err := utils.GenerateRandomState()
	if err != nil {
		return internalError(c, "Failed to generate state")
	}

	signedState := h.signState(state)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    signedState,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		MaxAge:   600,
	})

	url := h.oauthService.GetGoogleAuthURL(state)
	return c.Redirect(url)
}

func (h *OAuthHandler) GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	if state == "" {
		return badRequestError(c, "Missing state parameter")
	}

	storedState := c.Cookies("oauth_state")
	if storedState == "" {
		return badRequestError(c, "Authentication failed. Please try logging in again.")
	}

	if !h.verifyState(state, storedState) {
		return badRequestError(c, "Authentication failed. Please try logging in again.")
	}

	c.Cookie(&fiber.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
	})

	code := c.Query("code")
	if code == "" {
		return badRequestError(c, "Missing authorization code")
	}

	googleUser, err := h.oauthService.HandleGoogleCallback(c.Context(), code)
	if err != nil {
		return internalError(c, "Failed to authenticate with Google")
	}

	user, err := h.userService.GetByEmail(c.Context(), googleUser.Email)
	if err != nil {
		return internalError(c, "Database error")
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
			return internalError(c, "Failed to create user")
		}
	}

	accessToken, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return internalError(c, "Failed to generate access token")
	}

	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		return internalError(c, "Failed to generate refresh token")
	}

	expiresAt := time.Now().Add(h.jwtService.GetRefreshExpiry())
	if err := h.refreshTokenRepo.Store(c.Context(), user.ID.String(), refreshToken, expiresAt); err != nil {
		return internalError(c, "Failed to store refresh token")
	}

	h.setAuthCookies(c, accessToken, refreshToken)

	return c.JSON(fiber.Map{
		"message": "Google authentication successful",
		"user":    newUserResponse(user),
	})
}

func (h *OAuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return unauthorizedError(c, "Missing refresh token")
	}

	claims, err := h.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return unauthorizedError(c, "Invalid refresh token")
	}

	storedToken, err := h.refreshTokenRepo.GetByToken(c.Context(), refreshToken)
	if err != nil {
		return internalError(c, "Failed to validate token")
	}
	if storedToken == nil {
		return unauthorizedError(c, "Token has been revoked or expired")
	}

	user, err := h.userService.GetByID(c.Context(), claims.UserID)
	if err != nil || user == nil {
		return unauthorizedError(c, "User not found")
	}

	newAccessToken, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.Role,
	)
	if err != nil {
		return internalError(c, "Failed to generate access token")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		MaxAge:   int(h.jwtService.GetAccessExpiry().Seconds()),
	})

	return c.JSON(fiber.Map{
		"message": "Token refreshed successfully",
	})
}

func (h *OAuthHandler) Logout(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken != "" {
		if err := h.refreshTokenRepo.Revoke(c.Context(), refreshToken); err != nil {
			// Log error but continue logout
		}
	}

	h.clearAuthCookies(c)

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

func (h *OAuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.refreshTokenRepo.RevokeAllByUserID(c.Context(), userID); err != nil {
		return internalError(c, "Failed to logout from all devices")
	}

	h.clearAuthCookies(c)

	return c.JSON(fiber.Map{
		"message": "Logged out from all devices",
	})
}

func (h *OAuthHandler) GetProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	user, err := h.userService.GetByID(c.Context(), userID)
	if err != nil || user == nil {
		return notFoundError(c, "User not found")
	}

	return c.JSON(fiber.Map{
		"user": newUserResponse(user),
	})
}

func (h *OAuthHandler) setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	isSecure := c.Protocol() == "https"

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(h.jwtService.GetAccessExpiry().Seconds()),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   isSecure,
		SameSite: "Lax",
		MaxAge:   int(h.jwtService.GetRefreshExpiry().Seconds()),
	})
}

func (h *OAuthHandler) clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		MaxAge:   -1,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		MaxAge:   -1,
	})
}

func (h *OAuthHandler) signState(state string) string {
	mac := hmac.New(sha256.New, []byte(h.cookieSecret))
	mac.Write([]byte(state))
	signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", state, signature)
}

func (h *OAuthHandler) verifyState(state, signedState string) bool {
	parts := strings.Split(signedState, ".")
	if len(parts) != 2 {
		return false
	}

	storedState := parts[0]
	storedSignature := parts[1]

	if state != storedState {
		return false
	}

	mac := hmac.New(sha256.New, []byte(h.cookieSecret))
	mac.Write([]byte(storedState))
	expectedSignature := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(storedSignature), []byte(expectedSignature))
}
