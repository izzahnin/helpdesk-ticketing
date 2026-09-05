package handlers

import (
	"strconv"
	"time"

	"helpdesk-backend/internal/dto"
	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/repository"
	"helpdesk-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	Tickets *service.TicketService
	Repo    *repository.TicketRepository
}

// List godoc
// @Summary List tickets visible to the current user
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Ticket
// @Failure 500 {object} map[string]string
// @Router /api/tickets [get]
func (h *TicketHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	role := c.GetString("role")
	rows, err := h.Tickets.ListForUser(c.Request.Context(), userID, role)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch tickets"})
		return
	}
	c.JSON(200, rows)
}

// Submit godoc
// @Summary Submit a ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SubmitTicketRequest true "Ticket data"
// @Success 201 {object} models.Ticket
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tickets [post]
func (h *TicketHandler) Submit(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var in dto.SubmitTicketRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Title == "" || in.Description == "" || in.CategoryID == 0 {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	t := &models.Ticket{
		Title:       in.Title,
		Description: in.Description,
		RequesterID: userID,
		CategoryID:  in.CategoryID,
		Priority:    in.Priority,
	}
	if err := h.Tickets.Submit(c.Request.Context(), t); err != nil {
		c.JSON(500, gin.H{"error": "failed to create ticket"})
		return
	}
	c.JSON(201, t)
}

// Detail godoc
// @Summary Get ticket details and comments
// @Tags tickets
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Ticket ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tickets/{id} [get]
func (h *TicketHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	role := c.GetString("role")

	ticket, err := h.Repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "ticket not found"})
		return
	}
	if err := repository.EnsureTicketAccess(ticket, userID, role); err != nil {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	comments, err := h.Repo.FindComments(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch comments"})
		return
	}
	c.JSON(200, gin.H{"ticket": ticket, "comments": comments})
}

// AddComment godoc
// @Summary Add a comment to a ticket
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Ticket ID"
// @Param request body dto.CommentRequest true "Comment data"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/tickets/{id}/comments [post]
func (h *TicketHandler) AddComment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	userID := c.GetInt64("user_id")
	role := c.GetString("role")

	ticket, err := h.Repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "ticket not found"})
		return
	}
	if err := repository.EnsureTicketAccess(ticket, userID, role); err != nil {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	var in dto.CommentRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Message == "" {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	if err := h.Repo.AddComment(c.Request.Context(), id, userID, in.Message); err != nil {
		c.JSON(500, gin.H{"error": "failed to add comment"})
		return
	}
	c.JSON(201, gin.H{"message": "comment added"})
}

// Assign godoc
// @Summary Assign a ticket to staff
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Ticket ID"
// @Param request body dto.AssignTicketRequest true "Assignment data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tickets/{id}/assign [patch]
func (h *TicketHandler) Assign(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	actorID := c.GetInt64("user_id")
	var in dto.AssignTicketRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.AssigneeID == 0 {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	if err := h.Repo.AssignWithActivityLog(c.Request.Context(), id, in.AssigneeID, actorID); err != nil {
		c.JSON(500, gin.H{"error": "failed to assign ticket"})
		return
	}
	c.JSON(200, gin.H{"message": "assigned"})
}

// UpdateStatus godoc
// @Summary Update ticket status
// @Tags tickets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Ticket ID"
// @Param request body dto.UpdateTicketStatusRequest true "Status data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/tickets/{id}/status [patch]
func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	actorID := c.GetInt64("user_id")
	var in dto.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Status == "" {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	if err := h.Tickets.ChangeStatus(c.Request.Context(), id, actorID, in.Status); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "status updated", "updated_at": time.Now()})
}
