package models

import "time"

type StatusLog struct {
	ID        int64     `db:"id" json:"id"`
	TicketID  int64     `db:"ticket_id" json:"ticket_id"`
	OldStatus *string   `db:"old_status" json:"old_status"`
	NewStatus string    `db:"new_status" json:"new_status"`
	ChangedBy int64     `db:"changed_by" json:"changed_by"`
	ChangedAt time.Time `db:"changed_at" json:"changed_at"`
}

type ActivityLog struct {
	ID        int64     `db:"id" json:"id"`
	TicketID  int64     `db:"ticket_id" json:"ticket_id"`
	ActorID   int64     `db:"actor_id" json:"actor_id"`
	Action    string    `db:"action" json:"action"`
	OldValue  *string   `db:"old_value" json:"old_value"`
	NewValue  *string   `db:"new_value" json:"new_value"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
