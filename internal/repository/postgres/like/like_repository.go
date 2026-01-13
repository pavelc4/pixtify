package like

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	WallpaperID uuid.UUID `json:"wallpaper_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ToggleLike(ctx context.Context, userID, wallpaperID uuid.UUID) (bool, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND wallpaper_id = $2)`
	err := r.db.QueryRowContext(ctx, checkQuery, userID, wallpaperID).Scan(&exists)
	if err != nil {
		return false, err
	}

	if exists {
		// Remove like
		deleteQuery := `DELETE FROM likes WHERE user_id = $1 AND wallpaper_id = $2`
		_, err = r.db.ExecContext(ctx, deleteQuery, userID, wallpaperID)
		return false, err
	}

	// Add like
	insertQuery := `INSERT INTO likes (user_id, wallpaper_id) VALUES ($1, $2)`
	_, err = r.db.ExecContext(ctx, insertQuery, userID, wallpaperID)
	return true, err
}

func (r *Repository) IsLiked(ctx context.Context, userID, wallpaperID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = $1 AND wallpaper_id = $2)`
	err := r.db.QueryRowContext(ctx, query, userID, wallpaperID).Scan(&exists)
	return exists, err
}

func (r *Repository) GetWallpaperLikeCount(ctx context.Context, wallpaperID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM likes WHERE wallpaper_id = $1`
	err := r.db.QueryRowContext(ctx, query, wallpaperID).Scan(&count)
	return count, err
}

func (r *Repository) GetUserLikes(ctx context.Context, userID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM likes WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get wallpaper IDs
	query := `
		SELECT wallpaper_id 
		FROM likes 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var wallpaperIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		wallpaperIDs = append(wallpaperIDs, id)
	}

	return wallpaperIDs, total, rows.Err()
}

func (r *Repository) IncrementWallpaperLikeCount(ctx context.Context, wallpaperID uuid.UUID, delta int) error {
	query := `UPDATE wallpapers SET like_count = like_count + $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, delta, wallpaperID)
	return err
}
