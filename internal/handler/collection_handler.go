package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/service"
)

type CollectionHandler struct {
	collectionService *service.CollectionService
}

func NewCollectionHandler(collectionService *service.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// CreateCollection creates a new collection
func (h *CollectionHandler) CreateCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	if req.Name == "" {
		return badRequestError(c, "Collection name is required")
	}

	collection, err := h.collectionService.CreateCollection(c.Context(), userID, req.Name, req.Description, req.IsPublic)
	if err != nil {
		return internalError(c, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Collection created successfully",
		"collection": collection,
	})
}

// GetMyCollections retrieves all collections for the authenticated user
func (h *CollectionHandler) GetMyCollections(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	collections, total, err := h.collectionService.GetUserCollections(c.Context(), userID, page, limit)
	if err != nil {
		return internalError(c, "Failed to fetch collections")
	}

	return c.JSON(fiber.Map{
		"data": collections,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetCollectionById retrieves a specific collection by ID
func (h *CollectionHandler) GetCollectionById(c *fiber.Ctx) error {
	collectionID := c.Params("id")
	if _, err := uuid.Parse(collectionID); err != nil {
		return badRequestError(c, "Invalid collection ID")
	}

	// Get user ID if authenticated, empty string if not
	userID := ""
	if uid := c.Locals("user_id"); uid != nil {
		userID = uid.(string)
	}

	collection, err := h.collectionService.GetCollectionDetails(c.Context(), collectionID, userID)
	if err != nil {
		return internalError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": collection,
	})
}

// AddWallpaperToCollection adds a wallpaper to a collection
func (h *CollectionHandler) AddWallpaperToCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	if _, err := uuid.Parse(collectionID); err != nil {
		return badRequestError(c, "Invalid collection ID")
	}

	var req struct {
		WallpaperID string `json:"wallpaper_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	if _, err := uuid.Parse(req.WallpaperID); err != nil {
		return badRequestError(c, "Invalid wallpaper ID")
	}

	if err := h.collectionService.AddWallpaperToCollection(c.Context(), userID, collectionID, req.WallpaperID); err != nil {
		return internalError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Wallpaper added to collection successfully",
	})
}

// RemoveWallpaperFromCollection removes a wallpaper from a collection
func (h *CollectionHandler) RemoveWallpaperFromCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")
	wallpaperID := c.Params("wallpaperId")

	if _, err := uuid.Parse(collectionID); err != nil {
		return badRequestError(c, "Invalid collection ID")
	}

	if _, err := uuid.Parse(wallpaperID); err != nil {
		return badRequestError(c, "Invalid wallpaper ID")
	}

	if err := h.collectionService.RemoveWallpaperFromCollection(c.Context(), userID, collectionID, wallpaperID); err != nil {
		return internalError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Wallpaper removed from collection successfully",
	})
}

// GetCollectionWallpapers retrieves all wallpapers in a collection
func (h *CollectionHandler) GetCollectionWallpapers(c *fiber.Ctx) error {
	collectionID := c.Params("id")
	if _, err := uuid.Parse(collectionID); err != nil {
		return badRequestError(c, "Invalid collection ID")
	}

	// Get user ID if authenticated
	userID := ""
	if uid := c.Locals("user_id"); uid != nil {
		userID = uid.(string)
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	wallpapers, total, err := h.collectionService.GetCollectionWallpapers(c.Context(), collectionID, userID, page, limit)
	if err != nil {
		return internalError(c, err.Error())
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

// DeleteCollection deletes a collection
func (h *CollectionHandler) DeleteCollection(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	collectionID := c.Params("id")

	if _, err := uuid.Parse(collectionID); err != nil {
		return badRequestError(c, "Invalid collection ID")
	}

	if err := h.collectionService.DeleteCollection(c.Context(), userID, collectionID); err != nil {
		return internalError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Collection deleted successfully",
	})
}
