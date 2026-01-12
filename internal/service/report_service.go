package service

import (
	"context"

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
