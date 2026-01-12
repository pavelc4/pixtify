package wallpaper

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Wallpaper struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	OriginalURL   string    `json:"original_url"`
	LargeURL      string    `json:"large_url"`
	MediumURL     string    `json:"medium_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	Blurhash      *string   `json:"blurhash,omitempty"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	FileSizeBytes int64     `json:"file_size_bytes"`
	MimeType      string    `json:"mime_type"`
	ViewCount     int       `json:"view_count"`
	DownloadCount int       `json:"download_count"`
	LikeCount     int       `json:"like_count"`
	Status        string    `json:"status"`
	IsFeatured    bool      `json:"is_featured"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relations (fetched separately or joined)
	User *User    `json:"user,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, w *Wallpaper) error {
	query := `
		INSERT INTO wallpapers (
			id, user_id, title, description, 
			original_url, large_url, medium_url, thumbnail_url, blurhash,
			width, height, file_size_bytes, mime_type,
			status, is_featured, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, 
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, NOW(), NOW()
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		w.ID, w.UserID, w.Title, w.Description,
		w.OriginalURL, w.LargeURL, w.MediumURL, w.ThumbnailURL, w.Blurhash,
		w.Width, w.Height, w.FileSizeBytes, w.MimeType,
		w.Status, w.IsFeatured,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Wallpaper, error) {
	query := `
		SELECT 
			w.id, w.user_id, w.title, w.description, 
			w.original_url, w.large_url, w.medium_url, w.thumbnail_url, w.blurhash,
			w.width, w.height, w.file_size_bytes, w.mime_type,
			w.view_count, w.download_count, w.like_count,
			w.status, w.is_featured, w.created_at, w.updated_at,
			u.username, u.avatar_url
		FROM wallpapers w
		JOIN users u ON w.user_id = u.id
		WHERE w.id = $1
	`

	var w Wallpaper
	var u User
	var description sql.NullString
	var blurhash sql.NullString
	var avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.UserID, &w.Title, &description,
		&w.OriginalURL, &w.LargeURL, &w.MediumURL, &w.ThumbnailURL, &blurhash,
		&w.Width, &w.Height, &w.FileSizeBytes, &w.MimeType,
		&w.ViewCount, &w.DownloadCount, &w.LikeCount,
		&w.Status, &w.IsFeatured, &w.CreatedAt, &w.UpdatedAt,
		&u.Username, &avatarURL,
	)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		w.Description = &description.String
	}
	if blurhash.Valid {
		w.Blurhash = &blurhash.String
	}
	if avatarURL.Valid {
		u.AvatarURL = &avatarURL.String
	}
	u.ID = w.UserID
	w.User = &u

	// TODO: Fetch tags if needed

	return &w, nil
}

func (r *Repository) AddTags(ctx context.Context, wallpaperID uuid.UUID, tags []string) error {
	// First ensure tags exist
	for _, tagName := range tags {
		// Insert ignore if exists - simplest way for now
		// Need to optimize this for bulk inserts later
		var tagID uuid.UUID
		err := r.db.QueryRowContext(ctx, `
			INSERT INTO tags (name, slug) 
			VALUES ($1, $1) 
			ON CONFLICT (slug) DO UPDATE SET slug = EXCLUDED.slug 
			RETURNING id`, tagName).Scan(&tagID)
		if err != nil {
			return err
		}

		// Link tag
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO wallpaper_tags (wallpaper_id, tag_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, wallpaperID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Wallpaper, int, error) {
	// Basic list for now
	countQuery := `SELECT COUNT(*) FROM wallpapers WHERE status = 'active'`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
        SELECT 
            w.id, w.user_id, w.title, w.thumbnail_url, w.width, w.height,
            w.view_count, w.like_count, w.created_at,
            u.username, u.avatar_url
        FROM wallpapers w
        JOIN users u ON w.user_id = u.id
        WHERE w.status = 'active'
        ORDER BY w.created_at DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var wallpapers []*Wallpaper
	for rows.Next() {
		var w Wallpaper
		var u User
		var avatarURL sql.NullString

		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Title, &w.ThumbnailURL, &w.Width, &w.Height,
			&w.ViewCount, &w.LikeCount, &w.CreatedAt,
			&u.Username, &avatarURL,
		); err != nil {
			return nil, 0, err
		}

		if avatarURL.Valid {
			u.AvatarURL = &avatarURL.String
		}
		u.ID = w.UserID
		w.User = &u
		wallpapers = append(wallpapers, &w)
	}

	return wallpapers, total, nil
}
