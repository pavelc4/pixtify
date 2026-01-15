package tag

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Tag represents a wallpaper tag
type Tag struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	WallpaperCount int       `json:"wallpaper_count"`
	CreatedAt      time.Time `json:"created_at"`
}

// Repository handles tag database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new tag repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new tag
func (r *Repository) Create(ctx context.Context, name, slug string) (*Tag, error) {
	tag := &Tag{}

	query := `
		INSERT INTO tags (name, slug, wallpaper_count, created_at)
		VALUES ($1, $2, 0, NOW())
		RETURNING id, name, slug, wallpaper_count, created_at
	`

	err := r.db.QueryRowContext(ctx, query, name, slug).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.WallpaperCount,
		&tag.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return tag, nil
}

// List retrieves all tags with pagination
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Tag, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM tags`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated tags
	query := `
		SELECT id, name, slug, wallpaper_count, created_at
		FROM tags
		ORDER BY wallpaper_count DESC, name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tags := make([]*Tag, 0)
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.Slug,
			&tag.WallpaperCount,
			&tag.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}

// GetByID retrieves a tag by its ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Tag, error) {
	tag := &Tag{}

	query := `
		SELECT id, name, slug, wallpaper_count, created_at
		FROM tags
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.WallpaperCount,
		&tag.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// GetBySlug retrieves a tag by its slug
func (r *Repository) GetBySlug(ctx context.Context, slug string) (*Tag, error) {
	tag := &Tag{}

	query := `
		SELECT id, name, slug, wallpaper_count, created_at
		FROM tags
		WHERE slug = $1
	`

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tag.ID,
		&tag.Name,
		&tag.Slug,
		&tag.WallpaperCount,
		&tag.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return tag, nil
}

// Delete deletes a tag by ID (cascade deletes wallpaper_tags via DB constraint)
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tags WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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

// IncrementCount atomically increments the wallpaper_count for a tag
func (r *Repository) IncrementCount(ctx context.Context, tagID uuid.UUID) error {
	query := `
		UPDATE tags
		SET wallpaper_count = wallpaper_count + 1
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, tagID)
	return err
}

// DecrementCount atomically decrements the wallpaper_count for a tag
func (r *Repository) DecrementCount(ctx context.Context, tagID uuid.UUID) error {
	query := `
		UPDATE tags
		SET wallpaper_count = GREATEST(wallpaper_count - 1, 0)
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, tagID)
	return err
}
