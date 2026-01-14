package handler

import (
	"strconv"

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

// ListReports retrieves all reports with optional status filter (moderator only)
func (h *ReportHandler) ListReports(c *fiber.Ctx) error {
	status := c.Query("status", "all") // pending, resolved, dismissed, all
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	reports, total, err := h.reportService.ListReports(c.Context(), status, page, limit)
	if err != nil {
		return badRequestError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"data": reports,
		"meta": fiber.Map{
			"page":   page,
			"limit":  limit,
			"total":  total,
			"status": status,
		},
	})
}

// GetReportByID retrieves a single report by ID (moderator only)
func (h *ReportHandler) GetReportByID(c *fiber.Ctx) error {
	reportID := c.Params("id")

	if _, err := uuid.Parse(reportID); err != nil {
		return badRequestError(c, "Invalid report ID")
	}

	report, err := h.reportService.GetReportByID(c.Context(), reportID)
	if err != nil {
		return notFoundError(c, "Report not found")
	}

	return c.JSON(fiber.Map{
		"data": report,
	})
}

// UpdateReportStatus updates the status of a report (moderator only)
func (h *ReportHandler) UpdateReportStatus(c *fiber.Ctx) error {
	reportID := c.Params("id")
	moderatorID := c.Locals("user_id").(string)

	if _, err := uuid.Parse(reportID); err != nil {
		return badRequestError(c, "Invalid report ID")
	}

	var req struct {
		Status string `json:"status"` // "resolved" or "dismissed"
	}

	if err := c.BodyParser(&req); err != nil {
		return badRequestError(c, "Invalid request body")
	}

	if err := h.reportService.UpdateReportStatus(c.Context(), reportID, moderatorID, req.Status); err != nil {
		return internalError(c, err.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Report " + req.Status + " successfully",
	})
}
