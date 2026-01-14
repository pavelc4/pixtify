package wallpaper

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Wallpaper struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	OriginalURL   string    `json:"original_url"`
	ImageURL      string    `json:"image_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	Blurhash      *string   `json:"blurhash,omitempty"`
	DeviceType    string    `json:"device_type"`
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
			user_id, title, description, original_url, image_url,
			thumbnail_url, blurhash, device_type, width, height, file_size_bytes, mime_type,
			status, is_featured
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, view_count, download_count, like_count, created_at, updated_at
	`

	return r.db.QueryRowContext(
		ctx, query,
		w.UserID, w.Title, w.Description, w.OriginalURL, w.ImageURL,
		w.ThumbnailURL, w.Blurhash, w.DeviceType, w.Width, w.Height, w.FileSizeBytes, w.MimeType,
		w.Status, w.IsFeatured,
	).Scan(&w.ID, &w.ViewCount, &w.DownloadCount, &w.LikeCount, &w.CreatedAt, &w.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Wallpaper, error) {
	query := `
		SELECT id, user_id, title, description, original_url, image_url,
		       thumbnail_url, blurhash, device_type, width, height, file_size_bytes, mime_type,
		       view_count, download_count, like_count, status, is_featured,
		       created_at, updated_at
		FROM wallpapers
		WHERE id = $1 AND deleted_at IS NULL
	`

	w := &Wallpaper{}
	var description sql.NullString
	var blurhash sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&w.ID, &w.UserID, &w.Title, &description, &w.OriginalURL, &w.ImageURL,
		&w.ThumbnailURL, &blurhash, &w.DeviceType, &w.Width, &w.Height, &w.FileSizeBytes, &w.MimeType,
		&w.ViewCount, &w.DownloadCount, &w.LikeCount, &w.Status, &w.IsFeatured,
		&w.CreatedAt, &w.UpdatedAt,
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

	return w, nil
}

func (r *Repository) AddTags(ctx context.Context, wallpaperID uuid.UUID, tags []string) error {
	for _, tagName := range tags {
		var tagID int64
		err := r.db.QueryRowContext(ctx,
			`INSERT INTO tags (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = $1 RETURNING id`,
			tagName,
		).Scan(&tagID)

		if err != nil {
			return err
		}

		_, err = r.db.ExecContext(ctx,
			`INSERT INTO wallpaper_tags (wallpaper_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			wallpaperID, tagID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]*Wallpaper, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM wallpapers WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			w.id, w.user_id, w.title, w.thumbnail_url, w.width, w.height,
			w.view_count, w.like_count, w.created_at,
			u.username, u.avatar_url
		FROM wallpapers w
		INNER JOIN users u ON w.user_id = u.id
		WHERE w.deleted_at IS NULL
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

// Update updates wallpaper metadata
func (r *Repository) Update(ctx context.Context, id uuid.UUID, title, description *string) error {
	query := `
		UPDATE wallpapers
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := r.db.ExecContext(ctx, query, title, description, id)
	return err
}

// SoftDelete marks wallpaper as deleted
func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE wallpapers SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
func (r *Repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Wallpaper, error) {
	if len(ids) == 0 {
		return []*Wallpaper{}, nil
	}

	// Build query with ANY for efficient bulk fetch
	query := `
		SELECT 
			w.id, w.user_id, w.title, w.description, w.original_url, w.image_url,
			w.thumbnail_url, w.blurhash, w.device_type, w.width, w.height, w.file_size_bytes, w.mime_type,
			w.view_count, w.download_count, w.like_count, w.status, w.is_featured,
			w.created_at, w.updated_at,
			u.username, u.avatar_url
		FROM wallpapers w
		INNER JOIN users u ON w.user_id = u.id
		WHERE w.id = ANY($1) AND w.deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map for O(1) lookup when ordering results
	wallpaperMap := make(map[uuid.UUID]*Wallpaper, len(ids))

	for rows.Next() {
		var w Wallpaper
		var u User
		var description sql.NullString
		var blurhash sql.NullString
		var avatarURL sql.NullString

		err := rows.Scan(
			&w.ID, &w.UserID, &w.Title, &description, &w.OriginalURL, &w.ImageURL,
			&w.ThumbnailURL, &blurhash, &w.DeviceType, &w.Width, &w.Height, &w.FileSizeBytes, &w.MimeType,
			&w.ViewCount, &w.DownloadCount, &w.LikeCount, &w.Status, &w.IsFeatured,
			&w.CreatedAt, &w.UpdatedAt,
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

		wallpaperMap[w.ID] = &w
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Preserve input order (important for pagination)
	result := make([]*Wallpaper, 0, len(ids))
	for _, id := range ids {
		if wp, ok := wallpaperMap[id]; ok {
			result = append(result, wp)
		}
	}

	return result, nil
}

// SetFeaturedStatus toggles the is_featured flag for a wallpaper (moderator action)
func (r *Repository) SetFeaturedStatus(ctx context.Context, id uuid.UUID, isFeatured bool) error {
	query := `
		UPDATE wallpapers
		SET is_featured = $1,
		    updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, isFeatured, id)
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

// ListFeatured retrieves all featured wallpapers
func (r *Repository) ListFeatured(ctx context.Context, limit, offset int) ([]*Wallpaper, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM wallpapers WHERE is_featured = true AND deleted_at IS NULL`,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			w.id, w.user_id, w.title, w.thumbnail_url, w.width, w.height,
			w.view_count, w.like_count, w.created_at,
			u.username, u.avatar_url
		FROM wallpapers w
		INNER JOIN users u ON w.user_id = u.id
		WHERE w.is_featured = true AND w.deleted_at IS NULL
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

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return wallpapers, total, nil
}
