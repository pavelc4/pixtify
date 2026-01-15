package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/service"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagService *service.TagService
}

// NewTagHandler creates a new tag handler
func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{
		tagService: tagService,
	}
}

// CreateTag handles POST /api/tags (moderator only)
func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	tag, err := h.tagService.CreateTag(c.Context(), req.Name)
	if err != nil {
		// Handle specific errors
		switch err {
		case service.ErrTagNameTooShort, service.ErrTagNameTooLong:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case service.ErrTagAlreadyExists:
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create tag",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(tag)
}

// ListTags handles GET /api/tags (public)
func (h *TagHandler) ListTags(c *fiber.Ctx) error {
	// Parse pagination parameters
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	tags, total, err := h.tagService.ListTags(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve tags",
		})
	}

	return c.JSON(fiber.Map{
		"tags":   tags,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// DeleteTag handles DELETE /api/tags/:id (moderator only)
func (h *TagHandler) DeleteTag(c *fiber.Ctx) error {
	// Parse tag ID
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tag ID",
		})
	}

	err = h.tagService.DeleteTag(c.Context(), id)
	if err != nil {
		if err == service.ErrTagNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Tag not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete tag",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Tag deleted successfully",
	})
}
