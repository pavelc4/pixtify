package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/models"
	"github.com/pavelc4/pixtify/internal/service"
)

type ReportHandler struct {
	reportService    *service.ReportService
	userService      *service.UserService
	wallpaperService *service.WallpaperService
}

func NewReportHandler(reportService *service.ReportService, userService *service.UserService, wallpaperService *service.WallpaperService) *ReportHandler {
	return &ReportHandler{
		reportService:    reportService,
		userService:      userService,
		wallpaperService: wallpaperService,
	}
}

func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	var req models.ReportRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	// Get user ID from JWT context
	reporterID := c.Locals("user_id").(string)
	userID, err := uuid.Parse(reporterID)
	if err != nil {
		return badRequestError(c, "Invalid user ID")
	}

	// Fetch user from database to get username
	user, err := h.userService.GetByID(c.Context(), reporterID)
	if err != nil || user == nil {
		return unauthorizedError(c, "User not found")
	}

	// Fetch actual wallpaper data from database
	wallpaper, err := h.wallpaperService.GetWallpaper(c.Context(), req.WallpaperID.String())
	if err != nil {
		return notFoundError(c, "Wallpaper not found")
	}

	report := models.Report{
		ID:               uuid.New(),
		WallpaperID:      req.WallpaperID,
		ReporterID:       userID,
		WallpaperTitle:   wallpaper.Title,
		WallpaperURL:     wallpaper.ThumbnailURL,
		ReporterUsername: user.Username,
		Reason:           req.Reason,
		Status:           "pending",
	}

	if err := h.reportService.CreateReport(c.Context(), &report); err != nil {
		return internalError(c, "Failed to create report")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Report created successfully",
		"report_id": report.ID,
	})
}
