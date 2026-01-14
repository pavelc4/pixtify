package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pavelc4/pixtify/internal/models"
	"github.com/pavelc4/pixtify/internal/repository"
)

type ReportService struct {
	repo *repository.ReportRepository
}

func NewReportService(repo *repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) CreateReport(ctx context.Context, report *models.Report) error {
	return s.repo.Create(ctx, report)
}

// ListReports retrieves reports with optional status filter and pagination
func (s *ReportService) ListReports(ctx context.Context, status string, page, limit int) ([]*models.Report, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate status filter
	if status != "" && status != "all" {
		validStatuses := map[string]bool{"pending": true, "resolved": true, "dismissed": true}
		if !validStatuses[status] {
			return nil, 0, fmt.Errorf("invalid status: must be pending, resolved, dismissed, or all")
		}
	}

	offset := (page - 1) * limit
	return s.repo.List(ctx, status, limit, offset)
}

// GetReportByID retrieves a single report by ID
func (s *ReportService) GetReportByID(ctx context.Context, idStr string) (*models.Report, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid report ID")
	}

	return s.repo.GetByID(ctx, id)
}

// UpdateReportStatus updates the status of a report (moderator action)
func (s *ReportService) UpdateReportStatus(ctx context.Context, reportIDStr, moderatorIDStr, newStatus string) error {
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return fmt.Errorf("invalid report ID")
	}

	moderatorID, err := uuid.Parse(moderatorIDStr)
	if err != nil {
		return fmt.Errorf("invalid moderator ID")
	}

	// Validate status
	validStatuses := map[string]bool{"resolved": true, "dismissed": true}
	if !validStatuses[newStatus] {
		return fmt.Errorf("invalid status: must be 'resolved' or 'dismissed'")
	}

	// Check if report exists and is pending
	report, err := s.repo.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("report not found")
	}

	if report.Status != "pending" {
		return fmt.Errorf("report has already been %s", report.Status)
	}

	return s.repo.UpdateStatus(ctx, reportID, newStatus, moderatorID)
}
