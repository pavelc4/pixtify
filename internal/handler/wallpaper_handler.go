package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/service"
)

type WallpaperHandler struct {
	wallpaperService *service.WallpaperService
}

func NewWallpaperHandler(wallpaperService *service.WallpaperService) *WallpaperHandler {
	return &WallpaperHandler{
		wallpaperService: wallpaperService,
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
