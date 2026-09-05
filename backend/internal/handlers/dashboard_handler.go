package handlers

import (
	"errors"
	"strconv"
	"time"

	"helpdesk-backend/internal/dto"
	"helpdesk-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	Repo *repository.DashboardRepository
}

func parseDashboardFilter(c *gin.Context, allowStaffID bool) (repository.DashboardFilter, error) {
	var f repository.DashboardFilter
	if v := c.Query("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, err
		}
		f.From = &t
	}
	if v := c.Query("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, err
		}
		f.To = &t
	}
	if v := c.Query("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return f, err
		}
		f.CategoryID = &id
	}
	if v := c.Query("staff_id"); v != "" {
		if !allowStaffID {
			return f, errors.New("staff_id is not allowed")
		}
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return f, err
		}
		f.StaffID = &id
	}
	return f, nil
}

// Summary godoc
// @Summary Get dashboard summary
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Param staff_id query int64 false "Staff ID"
// @Success 200 {object} dto.DashboardSummaryResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/summary [get]
func (h *DashboardHandler) Summary(c *gin.Context) {
	f, err := parseDashboardFilter(c, true)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	statuses, err := h.Repo.StatusSummary(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch summary"})
		return
	}
	compliance, err := h.Repo.SLACompliance(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch sla compliance"})
		return
	}
	total, open, resolved := summarizeTicketStatus(statuses)
	c.JSON(200, dto.DashboardSummaryResponse{
		TotalTickets:     total,
		OpenTickets:      open,
		ResolvedTickets:  resolved,
		SLACompliancePct: compliance,
		StatusSummary:    toStatusSummaryResponse(statuses),
	})
}

// Trend godoc
// @Summary Get ticket trend
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Param staff_id query int64 false "Staff ID"
// @Success 200 {array} dto.TrendPointResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/trend [get]
func (h *DashboardHandler) Trend(c *gin.Context) {
	f, err := parseDashboardFilter(c, true)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.Repo.Trend(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch trend"})
		return
	}
	c.JSON(200, rows)
}

// StaffPerformance godoc
// @Summary Get staff performance
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Param staff_id query int64 false "Staff ID"
// @Success 200 {array} dto.StaffPerformanceResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/staff-performance [get]
func (h *DashboardHandler) StaffPerformance(c *gin.Context) {
	f, err := parseDashboardFilter(c, true)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.Repo.StaffPerformance(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch staff performance"})
		return
	}
	c.JSON(200, rows)
}

// SLABreaches godoc
// @Summary List breached tickets
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Param staff_id query int64 false "Staff ID"
// @Success 200 {array} dto.SLABreachResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/dashboard/sla-breaches [get]
func (h *DashboardHandler) SLABreaches(c *gin.Context) {
	f, err := parseDashboardFilter(c, true)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.Repo.SLABreaches(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch sla breaches"})
		return
	}
	c.JSON(200, rows)
}

// MyPerformance godoc
// @Summary Get current staff performance
// @Tags dashboard
// @Produce json
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Success 200 {object} dto.MyPerformanceResponse
// @Failure 400 {object} map[string]string
// @Router /api/dashboard/my-performance [get]
func (h *DashboardHandler) MyPerformance(c *gin.Context) {
	f, err := parseDashboardFilter(c, false)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	staffID := c.GetInt64("user_id")
	f.StaffID = &staffID
	statuses, err := h.Repo.StatusSummary(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch summary"})
		return
	}
	compliance, err := h.Repo.SLACompliance(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch sla compliance"})
		return
	}
	breaches, err := h.Repo.SLABreaches(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch sla breaches"})
		return
	}
	perf, err := h.Repo.StaffPerformance(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch staff performance"})
		return
	}
	avg := 0.0
	resolved := 0
	if len(perf) > 0 {
		avg = perf[0].AvgHours
		resolved = perf[0].Resolved
	}
	total := 0
	for _, s := range statuses {
		total += s.Count
	}
	c.JSON(200, dto.MyPerformanceResponse{
		AssignedTotal:      total,
		ResolvedTotal:      resolved,
		AvgResolutionHours: avg,
		SLACompliancePct:   compliance,
		StatusSummary:      toStatusSummaryResponse(statuses),
		SLABreaches:        toSLABreachResponse(breaches),
	})
}

func summarizeTicketStatus(statuses []repository.StatusSummary) (total, open, resolved int) {
	for _, status := range statuses {
		total += status.Count
		switch status.Status {
		case "Open":
			open += status.Count
		case "Resolved", "Closed":
			resolved += status.Count
		}
	}
	return total, open, resolved
}

func toStatusSummaryResponse(rows []repository.StatusSummary) []dto.StatusSummaryResponse {
	out := make([]dto.StatusSummaryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.StatusSummaryResponse{
			Status: row.Status,
			Count:  row.Count,
		})
	}
	return out
}

func toSLABreachResponse(rows []repository.BreachedTicket) []dto.SLABreachResponse {
	out := make([]dto.SLABreachResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.SLABreachResponse{
			ID:          row.ID,
			Title:       row.Title,
			Priority:    row.Priority,
			SLADeadline: row.SLADeadline,
		})
	}
	return out
}
