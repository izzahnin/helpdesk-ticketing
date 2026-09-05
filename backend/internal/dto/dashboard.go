package dto

import "time"

type DashboardFilter struct {
	From       string `form:"from"`
	To         string `form:"to"`
	CategoryID *int64 `form:"category_id"`
	StaffID    *int64 `form:"staff_id"`
}

type StatusSummaryResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type DashboardSummaryResponse struct {
	TotalTickets     int                     `json:"total_tickets"`
	OpenTickets      int                     `json:"open_tickets"`
	ResolvedTickets  int                     `json:"resolved_tickets"`
	SLACompliancePct float64                 `json:"sla_compliance_pct"`
	StatusSummary    []StatusSummaryResponse `json:"status_summary"`
}

type TrendPointResponse struct {
	Week  string `json:"week"`
	Count int    `json:"count"`
}

type StaffPerformanceResponse struct {
	Rank               int     `json:"rank"`
	StaffID            int64   `json:"staff_id"`
	StaffName          string  `json:"staff_name"`
	Resolved           int     `json:"resolved"`
	AvgResolutionHours float64 `json:"avg_resolution_hours"`
}

type SLABreachResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Priority    string    `json:"priority"`
	SLADeadline time.Time `json:"sla_deadline"`
}

type MyPerformanceResponse struct {
	AssignedTotal      int                     `json:"assigned_total"`
	ResolvedTotal      int                     `json:"resolved_total"`
	AvgResolutionHours float64                 `json:"avg_resolution_hours"`
	SLACompliancePct   float64                 `json:"sla_compliance_pct"`
	StatusSummary      []StatusSummaryResponse `json:"status_summary"`
	SLABreaches        []SLABreachResponse     `json:"sla_breaches"`
}
