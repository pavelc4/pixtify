package collection

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Collection struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	IsPublic       bool      `json:"is_public"`
	WallpaperCount int       `json:"wallpaper_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new collection
func (r *Repository) Create(ctx context.Context, c *Collection) error {
	query := `
		INSERT INTO collections (user_id, name, description, is_public)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, wallpaper_count
	`
	return r.db.QueryRowContext(ctx, query, c.UserID, c.Name, c.Description, c.IsPublic).Scan(
		&c.ID, &c.CreatedAt, &c.UpdatedAt, &c.WallpaperCount,
	)
}

// GetByID retrieves a collection by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Collection, error) {
	query := `
		SELECT id, user_id, name, description, is_public, wallpaper_count, created_at, updated_at
		FROM collections
		WHERE id = $1
	`
	var c Collection
	var description sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Name, &description, &c.IsPublic, &c.WallpaperCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		c.Description = &description.String
	}
	return &c, nil
}

// GetUserCollections retrieves all collections for a user with pagination
func (r *Repository) GetUserCollections(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Collection, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM collections WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get collections
	query := `
		SELECT id, user_id, name, description, is_public, wallpaper_count, created_at, updated_at
		FROM collections
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var collections []*Collection
	for rows.Next() {
		var c Collection
		var description sql.NullString
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &description, &c.IsPublic, &c.WallpaperCount, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if description.Valid {
			c.Description = &description.String
		}
		collections = append(collections, &c)
	}

	return collections, total, rows.Err()
}

// Update updates collection metadata
func (r *Repository) Update(ctx context.Context, c *Collection) error {
	query := `
		UPDATE collections
		SET name = $1, description = $2, is_public = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.db.QueryRowContext(ctx, query, c.Name, c.Description, c.IsPublic, c.ID).Scan(&c.UpdatedAt)
}

// Delete removes a collection
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM collections WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// AddWallpaper adds a wallpaper to a collection atomically using transaction
func (r *Repository) AddWallpaper(ctx context.Context, collectionID, wallpaperID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert collection item
	insertQuery := `
		INSERT INTO collection_items (collection_id, wallpaper_id)
		VALUES ($1, $2)
		ON CONFLICT (collection_id, wallpaper_id) DO NOTHING
	`
	result, err := tx.ExecContext(ctx, insertQuery, collectionID, wallpaperID)
	if err != nil {
		return err
	}

	// Only update count if a row was actually inserted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		// Update wallpaper count
		updateCountQuery := `
			UPDATE collections
			SET wallpaper_count = wallpaper_count + 1, updated_at = NOW()
			WHERE id = $1
		`
		_, err = tx.ExecContext(ctx, updateCountQuery, collectionID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RemoveWallpaper removes a wallpaper from a collection atomically using transaction
func (r *Repository) RemoveWallpaper(ctx context.Context, collectionID, wallpaperID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Delete collection item
	deleteQuery := `
		DELETE FROM collection_items
		WHERE collection_id = $1 AND wallpaper_id = $2
	`
	result, err := tx.ExecContext(ctx, deleteQuery, collectionID, wallpaperID)
	if err != nil {
		return err
	}

	// Only update count if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		// Update wallpaper count
		updateCountQuery := `
			UPDATE collections
			SET wallpaper_count = wallpaper_count - 1, updated_at = NOW()
			WHERE id = $1 AND wallpaper_count > 0
		`
		_, err = tx.ExecContext(ctx, updateCountQuery, collectionID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetCollectionWallpapers retrieves all wallpapers in a collection
func (r *Repository) GetCollectionWallpapers(ctx context.Context, collectionID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM collection_items WHERE collection_id = $1`
	err := r.db.QueryRowContext(ctx, countQuery, collectionID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get wallpaper IDs
	query := `
		SELECT wallpaper_id
		FROM collection_items
		WHERE collection_id = $1
		ORDER BY added_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, collectionID, limit, offset)
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
