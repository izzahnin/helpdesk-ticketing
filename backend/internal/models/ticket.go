package models

import "time"

type Ticket struct {
	ID            int64      `db:"id" json:"id"`
	Title         string     `db:"title" json:"title"`
	Description   string     `db:"description" json:"description"`
	RequesterID   int64      `db:"requester_id" json:"requester_id"`
	AssigneeID    *int64     `db:"assignee_id" json:"assignee_id"`
	CategoryID    int64      `db:"category_id" json:"category_id"`
	CategoryName  string     `db:"category_name" json:"category_name"`
	Priority      string     `db:"priority" json:"priority"`
	Status        string     `db:"status" json:"status"`
	SLADeadline   *time.Time `db:"sla_deadline" json:"sla_deadline"`
	SLATotalHours *int       `db:"sla_total_hours" json:"sla_total_hours"`
	ResolvedAt    *time.Time `db:"resolved_at" json:"resolved_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

type TicketComment struct {
	ID         int64     `db:"id" json:"id"`
	TicketID   int64     `db:"ticket_id" json:"ticket_id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	AuthorName string    `db:"author_name" json:"author_name"`
	AuthorRole string    `db:"author_role" json:"author_role"`
	Message    string    `db:"message" json:"message"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
