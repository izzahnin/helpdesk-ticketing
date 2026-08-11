package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"helpdesk-backend/internal/models"
)

type RefreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

const refreshTokenColumns = `
	id,
	user_id,
	token_hash,
	expires_at,
	revoked_at,
	created_at
`

func (r *RefreshTokenRepository) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1,$2,$3)
	`, userID, tokenHash, expiresAt)

	return err
}

func (r *RefreshTokenRepository) FindValidByHash(ctx context.Context, tokenHash string, now time.Time) (*models.RefreshToken, error) {
	var row models.RefreshToken
	query := `SELECT ` + refreshTokenColumns + `
		FROM refresh_tokens
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > $2
	`

	if err := r.db.GetContext(ctx, &row, query, tokenHash, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, err
	}

	return &row, nil
}

func (r *RefreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at=now()
		WHERE token_hash=$1 AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}

	return nil
}
