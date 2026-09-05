package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ReportRepository struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

type ReportRow struct {
	TicketID    int64   `db:"ticket_id"`
	Title       string  `db:"title"`
	Category    string  `db:"category"`
	Priority    string  `db:"priority"`
	Status      string  `db:"status"`
	Requester   string  `db:"requester"`
	Assignee    *string `db:"assignee"`
	CreatedAt   string  `db:"created_at"`
	ResolvedAt  *string `db:"resolved_at"`
	SLADeadline *string `db:"sla_deadline"`
	SLAState    string  `db:"sla_state"`
}

func (r *ReportRepository) ExportRows(ctx context.Context, f DashboardFilter) ([]ReportRow, error) {
	args := []interface{}{}
	where := filterSQL(f, &args)
	var rows []ReportRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT t.id AS ticket_id, t.title, c.name AS category, t.priority, t.status,
		       req.name AS requester, ass.name AS assignee,
		       t.created_at::text AS created_at,
		       t.resolved_at::text AS resolved_at,
		       t.sla_deadline::text AS sla_deadline,
		       CASE
		         WHEN t.sla_deadline IS NULL OR t.status IN ('Resolved','Closed') THEN 'ok'
		         WHEN now() > t.sla_deadline THEN 'breached'
		         WHEN t.sla_deadline - now() <= (COALESCE(sr.resolution_hours, c.default_sla_hours) * interval '1 hour') / 5 THEN 'at_risk'
		         ELSE 'ok'
		       END AS sla_state
		FROM tickets t
		JOIN categories c ON c.id=t.category_id
		LEFT JOIN sla_rules sr ON sr.category_id=t.category_id AND sr.priority=t.priority
		JOIN users req ON req.id=t.requester_id
		LEFT JOIN users ass ON ass.id=t.assignee_id
		WHERE `+where+`
		ORDER BY t.created_at DESC
	`, args...)
	return rows, err
}
