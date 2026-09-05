package handlers

import (
	"encoding/csv"
	"strconv"

	"helpdesk-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct{ Repo *repository.ReportRepository }

// ExportCSV godoc
// @Summary Export tickets as CSV
// @Tags reports
// @Produce text/csv
// @Security BearerAuth
// @Param from query string false "Start date (YYYY-MM-DD)"
// @Param to query string false "End date (YYYY-MM-DD)"
// @Param category_id query int64 false "Category ID"
// @Param staff_id query int64 false "Staff ID"
// @Success 200 {string} string "CSV report"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/reports/export [get]
func (h *ReportHandler) ExportCSV(c *gin.Context) {
	f, err := parseDashboardFilter(c, true)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	rows, err := h.Repo.ExportRows(c.Request.Context(), f)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to export report"})
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="helpdesk-report.csv"`)
	w := csv.NewWriter(c.Writer)
	if err := w.Write([]string{"ticket_id", "title", "category", "priority", "status", "requester", "assignee", "created_at", "resolved_at", "sla_deadline", "sla_state"}); err != nil {
		c.JSON(500, gin.H{"error": "failed to write csv"})
		return
	}
	for _, r := range rows {
		assignee, resolved, deadline := "", "", ""
		if r.Assignee != nil {
			assignee = *r.Assignee
		}
		if r.ResolvedAt != nil {
			resolved = *r.ResolvedAt
		}
		if r.SLADeadline != nil {
			deadline = *r.SLADeadline
		}
		if err := w.Write([]string{
			strconv.FormatInt(r.TicketID, 10), r.Title, r.Category, r.Priority, r.Status,
			r.Requester, assignee, r.CreatedAt, resolved, deadline, r.SLAState,
		}); err != nil {
			c.JSON(500, gin.H{"error": "failed to write csv"})
			return
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return
	}
}
