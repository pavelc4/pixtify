package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Store(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, id, userID, token, expiresAt, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1
		  AND revoked = FALSE
		  AND expires_at > NOW()
	`

	rt := &RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.Revoked,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Token not found or expired
	}
	if err != nil {
		return nil, err
	}

	return rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token = $1
	`

	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE user_id = $1 AND revoked = FALSE
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *RefreshTokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR revoked = TRUE
	`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (r *RefreshTokenRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM refresh_tokens
		WHERE user_id = $1
		  AND revoked = FALSE
		  AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
