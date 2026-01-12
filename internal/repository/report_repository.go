package repository

import (
	"context"
	"database/sql"

	"github.com/pavelc4/pixtify/internal/models"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *models.Report) error {
	query := `
        INSERT INTO reports (
            id, wallpaper_id, reporter_id, wallpaper_title,
            wallpaper_url, reporter_username, reason, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.WallpaperID, report.ReporterID,
		report.WallpaperTitle, report.WallpaperURL,
		report.ReporterUsername, report.Reason, report.Status,
	)
	return err
}
