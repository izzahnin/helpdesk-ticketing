package models

import "time"

type Category struct {
	ID              int64     `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	DefaultSLAHours int       `db:"default_sla_hours" json:"default_sla_hours"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type SLARule struct {
	ID              int64     `db:"id" json:"id"`
	CategoryID      int64     `db:"category_id" json:"category_id"`
	Priority        string    `db:"priority" json:"priority"`
	ResolutionHours int       `db:"resolution_hours" json:"resolution_hours"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
