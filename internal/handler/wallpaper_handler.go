package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/service"
)

type WallpaperHandler struct {
	wallpaperService *service.WallpaperService
	likeService      *service.LikeService
}

func NewWallpaperHandler(wallpaperService *service.WallpaperService, likeService *service.LikeService) *WallpaperHandler {
	return &WallpaperHandler{
		wallpaperService: wallpaperService,
		likeService:      likeService,
	}
}

func (h *WallpaperHandler) UploadWallpaper(c *fiber.Ctx) error {
	// Get User ID (set by middleware)
	userID := c.Locals("user_id").(string)

	// Validate file present
	file, err := c.FormFile("image")
	if err != nil {
		return badRequestError(c, "Image file is required")
	}

	// Validate fields
	title := c.FormValue("title")
	if title == "" {
		return badRequestError(c, "Title is required")
	}
	description := c.FormValue("description")

	// Tags
	tagsStr := c.FormValue("tags")
	var tags []string
	if tagsStr != "" {
		tags = service.SplitTags(tagsStr)
	}

	// Read file content
	src, err := file.Open()
	if err != nil {
		return internalError(c, "Failed to open file")
	}
	defer src.Close()

	fileData := make([]byte, file.Size)
	_, err = src.Read(fileData)
	if err != nil {
		return internalError(c, "Failed to read file")
	}

	contentType := file.Header.Get("Content-Type")

	// Call Service
	input := service.CreateWallpaperInput{
		UserID:      userID,
		Title:       title,
		Description: description,
		ImageData:   fileData,
		ContentType: contentType,
		Tags:        tags,
	}

	wallpaper, err := h.wallpaperService.CreateWallpaper(c.Context(), input)
	if err != nil {
		return internalError(c, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Wallpaper uploaded successfully",
		"wallpaper": wallpaper,
	})
}

func (h *WallpaperHandler) ListWallpapers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	wallpapers, total, err := h.wallpaperService.ListWallpapers(c.Context(), page, limit)
	if err != nil {
		return internalError(c, "Failed to fetch wallpapers")
	}

	return c.JSON(fiber.Map{
		"data": wallpapers,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *WallpaperHandler) GetWallpaper(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return badRequestError(c, "Invalid wallpaper ID")
	}

	wp, err := h.wallpaperService.GetWallpaper(c.Context(), id)
	if err != nil {
		return internalError(c, "Wallpaper not found or error fetching")
	}

	return c.JSON(fiber.Map{
		"data": wp,
	})
}

func (h *WallpaperHandler) LikeWallpaper(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	wallpaperID := c.Params("id")

	if _, err := uuid.Parse(wallpaperID); err != nil {
		return badRequestError(c, "Invalid wallpaper ID")
	}

	liked, err := h.likeService.ToggleLike(c.Context(), userID, wallpaperID)
	if err != nil {
		return internalError(c, err.Error())
	}

	action := "liked"
	if !liked {
		action = "unliked"
	}

	return c.JSON(fiber.Map{
		"message": "Wallpaper " + action + " successfully",
		"liked":   liked,
	})
}

func (h *WallpaperHandler) GetMyLikes(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	wallpapers, total, err := h.likeService.GetUserLikedWallpapers(c.Context(), userID, page, limit)
	if err != nil {
		return internalError(c, "Failed to fetch liked wallpapers")
	}

	return c.JSON(fiber.Map{
		"data": wallpapers,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
