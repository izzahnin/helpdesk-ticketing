package repository

import (
	"context"
	"database/sql"
	"errors"

	"helpdesk-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `
	id,
	name,
	email,
	password_hash,
	role,
	is_active,
	created_at,
	updated_at
`

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash, role, is_active)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, updated_at
	`, u.Name, u.Email, u.PasswordHash, u.Role, u.IsActive).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) UpsertBootstrapSuperAdmin(ctx context.Context, u *models.User) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO users (name, email, password_hash, role, is_active)
		VALUES ($1,$2,$3,'super_admin',true)
		ON CONFLICT (email) DO UPDATE
		SET name=EXCLUDED.name,
		    password_hash=EXCLUDED.password_hash,
		    role='super_admin',
		    is_active=true,
		    updated_at=now()
		RETURNING id, created_at, updated_at
	`, u.Name, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	query := `SELECT ` + userColumns + ` FROM users WHERE email=$1`

	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	var u models.User
	query := `SELECT ` + userColumns + ` FROM users WHERE id=$1`

	if err := r.db.GetContext(ctx, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var rows []models.User
	query := `SELECT ` + userColumns + ` FROM users ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &rows, query)
	return rows, err
}

func (r *UserRepository) FindStaff(ctx context.Context) ([]models.User, error) {
	var rows []models.User
	query := `SELECT ` + userColumns + ` FROM users WHERE role='staff' AND is_active=true ORDER BY name`

	err := r.db.SelectContext(ctx, &rows, query)
	return rows, err
}

func (r *UserRepository) UpdateRoleStatus(ctx context.Context, id int64, role string, isActive bool) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE users SET role=$1, is_active=$2, updated_at=now() WHERE id=$3
	`, role, isActive, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
