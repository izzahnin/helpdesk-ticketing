package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type DashboardRepository struct {
	db *sqlx.DB
}

func NewDashboardRepository(db *sqlx.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

type DashboardFilter struct {
	From       *time.Time
	To         *time.Time
	CategoryID *int64
	StaffID    *int64
}

func filterSQL(f DashboardFilter, args *[]interface{}) string {
	parts := []string{"1=1"}
	if f.From != nil {
		*args = append(*args, *f.From)
		parts = append(parts, fmt.Sprintf("t.created_at >= $%d", len(*args)))
	}
	if f.To != nil {
		*args = append(*args, f.To.Add(24*time.Hour))
		parts = append(parts, fmt.Sprintf("t.created_at < $%d", len(*args)))
	}
	if f.CategoryID != nil {
		*args = append(*args, *f.CategoryID)
		parts = append(parts, fmt.Sprintf("t.category_id = $%d", len(*args)))
	}
	if f.StaffID != nil {
		*args = append(*args, *f.StaffID)
		parts = append(parts, fmt.Sprintf("t.assignee_id = $%d", len(*args)))
	}
	return strings.Join(parts, " AND ")
}

type StatusSummary struct {
	Status string `db:"status" json:"status"`
	Count  int    `db:"count" json:"count"`
}

func (r *DashboardRepository) StatusSummary(ctx context.Context, f DashboardFilter) ([]StatusSummary, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var rows []StatusSummary
	err := r.db.SelectContext(ctx, &rows, `SELECT t.status, COUNT(*) AS count FROM tickets t WHERE `+where+` GROUP BY t.status`, args...)
	return rows, err
}

func (r *DashboardRepository) SLACompliance(ctx context.Context, f DashboardFilter) (float64, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var pct float64
	err := r.db.GetContext(ctx, &pct, `
		SELECT COALESCE(COUNT(*) FILTER (WHERE t.resolved_at <= t.sla_deadline) * 100.0 / NULLIF(COUNT(*),0), 0)
		FROM tickets t
		WHERE t.status IN ('Resolved','Closed') AND `+where, args...)
	return pct, err
}

type TrendPoint struct {
	Week  string `db:"week" json:"week"`
	Count int    `db:"count" json:"count"`
}

func (r *DashboardRepository) Trend(ctx context.Context, f DashboardFilter) ([]TrendPoint, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var rows []TrendPoint
	err := r.db.SelectContext(ctx, &rows, `
		SELECT TO_CHAR(DATE_TRUNC('week', t.created_at), 'YYYY-MM-DD') AS week, COUNT(*) AS count
		FROM tickets t
		WHERE `+where+`
		GROUP BY 1 ORDER BY 1
	`, args...)
	return rows, err
}

type StaffPerformance struct {
	StaffName string  `db:"staff_name" json:"staff_name"`
	StaffID   int64   `db:"staff_id" json:"staff_id"`
	Resolved  int     `db:"resolved" json:"resolved"`
	AvgHours  float64 `db:"avg_resolution_hours" json:"avg_resolution_hours"`
	Rank      int     `db:"rank" json:"rank"`
}

func (r *DashboardRepository) StaffPerformance(ctx context.Context, f DashboardFilter) ([]StaffPerformance, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var rows []StaffPerformance
	err := r.db.SelectContext(ctx, &rows, `
		SELECT u.name AS staff_name,
		       u.id AS staff_id,
		       COUNT(*) AS resolved,
		       COALESCE(ROUND(AVG(EXTRACT(EPOCH FROM (t.resolved_at - t.created_at))/3600)::numeric, 2), 0) AS avg_resolution_hours,
		       RANK() OVER (ORDER BY COUNT(*) DESC) AS rank
		FROM tickets t
		JOIN users u ON u.id=t.assignee_id
		WHERE t.status IN ('Resolved','Closed') AND `+where+`
		GROUP BY u.id, u.name
	`, args...)
	return rows, err
}

type BreachedTicket struct {
	ID          int64     `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Priority    string    `db:"priority" json:"priority"`
	SLADeadline time.Time `db:"sla_deadline" json:"sla_deadline"`
}

func (r *DashboardRepository) SLABreaches(ctx context.Context, f DashboardFilter) ([]BreachedTicket, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var rows []BreachedTicket
	err := r.db.SelectContext(ctx, &rows, `
		SELECT t.id, t.title, t.priority, t.sla_deadline
		FROM tickets t
		WHERE t.status NOT IN ('Resolved','Closed') AND t.sla_deadline < now() AND `+where+`
		ORDER BY t.sla_deadline ASC
	`, args...)
	return rows, err
}
