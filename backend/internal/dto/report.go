package dto

import "time"

type ReportExportRow struct {
	TicketID    int64      `json:"ticket_id" db:"ticket_id"`
	Title       string     `json:"title" db:"title"`
	Category    string     `json:"category" db:"category"`
	Priority    string     `json:"priority" db:"priority"`
	Status      string     `json:"status" db:"status"`
	Requester   string     `json:"requester" db:"requester"`
	Assignee    *string    `json:"assignee" db:"assignee"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at" db:"resolved_at"`
	SLADeadline *time.Time `json:"sla_deadline" db:"sla_deadline"`
	SLAState    string     `json:"sla_state" db:"sla_state"`
}
