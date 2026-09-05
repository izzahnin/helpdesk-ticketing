package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type SystemHandler struct {
	db *sqlx.DB
}

func NewSystemHandler(db *sqlx.DB) *SystemHandler {
	return &SystemHandler{db: db}
}

// Health godoc
// @Summary Check API health
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func (h *SystemHandler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// Ready godoc
// @Summary Check API and database readiness
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/ready [get]
func (h *SystemHandler) Ready(c *gin.Context) {
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		c.JSON(503, gin.H{"status": "not_ready"})
		return
	}
	c.JSON(200, gin.H{"status": "ready"})
}
