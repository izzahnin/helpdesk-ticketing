package repository

import (
	"context"

	"helpdesk-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

const categoryColumns = `
	id,
	name,
	default_sla_hours,
	created_at,
	updated_at
`

func (r *CategoryRepository) FindAll(ctx context.Context) ([]models.Category, error) {
	var rows []models.Category
	query := `SELECT ` + categoryColumns + ` FROM categories ORDER BY name`

	err := r.db.SelectContext(ctx, &rows, query)
	return rows, err
}

func (r *CategoryRepository) Create(ctx context.Context, c *models.Category) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO categories (name, default_sla_hours)
		VALUES ($1,$2)
		RETURNING id, created_at, updated_at
	`, c.Name, c.DefaultSLAHours).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *CategoryRepository) Update(ctx context.Context, c *models.Category) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE categories SET name=$1, default_sla_hours=$2, updated_at=now()
		WHERE id=$3
	`, c.Name, c.DefaultSLAHours, c.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func (r *CategoryRepository) UpsertSLARule(ctx context.Context, rule *models.SLARule) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sla_rules (category_id, priority, resolution_hours)
		VALUES ($1,$2,$3)
		ON CONFLICT (category_id, priority)
		DO UPDATE SET resolution_hours=EXCLUDED.resolution_hours, updated_at=now()
	`, rule.CategoryID, rule.Priority, rule.ResolutionHours)
	return err
}

func (r *CategoryRepository) ResolveSLAHours(ctx context.Context, categoryID int64, priority string) (int, error) {
	var hours int
	err := r.db.GetContext(ctx, &hours, `
		SELECT COALESCE(sr.resolution_hours, c.default_sla_hours) AS resolution_hours
		FROM categories c
		LEFT JOIN sla_rules sr ON sr.category_id=c.id AND sr.priority=$2
		WHERE c.id=$1
	`, categoryID, priority)
	return hours, err
}
