package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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

// List retrieves reports with optional status filter and pagination
func (r *ReportRepository) List(ctx context.Context, status string, limit, offset int) ([]*models.Report, int, error) {
	// Count total with filter
	var total int
	countQuery := `SELECT COUNT(*) FROM reports WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if status != "" && status != "all" {
		countQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build query with filter
	query := `
		SELECT 
			id, wallpaper_id, reporter_id, wallpaper_title, wallpaper_url,
			reporter_username, reason, status, resolved_by, resolved_at, created_at
		FROM reports
		WHERE 1=1
	`

	args = []interface{}{}
	argPos = 1

	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		var report models.Report
		var wallpaperID sql.NullString
		var resolvedBy sql.NullString
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&report.ID, &wallpaperID, &report.ReporterID,
			&report.WallpaperTitle, &report.WallpaperURL,
			&report.ReporterUsername, &report.Reason, &report.Status,
			&resolvedBy, &resolvedAt, &report.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Handle nullable fields
		if wallpaperID.Valid {
			wpID, _ := uuid.Parse(wallpaperID.String)
			report.WallpaperID = wpID
		}
		if resolvedBy.Valid {
			rbID, _ := uuid.Parse(resolvedBy.String)
			report.ResolvedBy = &rbID
		}
		if resolvedAt.Valid {
			report.ResolvedAt = &resolvedAt.Time
		}

		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetByID retrieves a single report by ID
func (r *ReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	query := `
		SELECT 
			id, wallpaper_id, reporter_id, wallpaper_title, wallpaper_url,
			reporter_username, reason, status, resolved_by, resolved_at, created_at
		FROM reports
		WHERE id = $1
	`

	var report models.Report
	var wallpaperID sql.NullString
	var resolvedBy sql.NullString
	var resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &wallpaperID, &report.ReporterID,
		&report.WallpaperTitle, &report.WallpaperURL,
		&report.ReporterUsername, &report.Reason, &report.Status,
		&resolvedBy, &resolvedAt, &report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if wallpaperID.Valid {
		wpID, _ := uuid.Parse(wallpaperID.String)
		report.WallpaperID = wpID
	}
	if resolvedBy.Valid {
		rbID, _ := uuid.Parse(resolvedBy.String)
		report.ResolvedBy = &rbID
	}
	if resolvedAt.Valid {
		report.ResolvedAt = &resolvedAt.Time
	}

	return &report, nil
}

// UpdateStatus updates the status of a report (resolved/dismissed)
func (r *ReportRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, resolvedBy uuid.UUID) error {
	query := `
		UPDATE reports
		SET status = $1,
		    resolved_by = $2,
		    resolved_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, resolvedBy, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
