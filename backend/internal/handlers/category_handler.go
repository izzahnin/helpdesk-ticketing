package handlers

import (
	"strconv"

	"helpdesk-backend/internal/dto"
	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	Categories *repository.CategoryRepository
}

// List godoc
// @Summary List categories
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Category
// @Failure 500 {object} map[string]string
// @Router /api/categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.Categories.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch categories"})
		return
	}
	c.JSON(200, rows)
}

// Create godoc
// @Summary Create a category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CategoryRequest true "Category data"
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var in dto.CategoryRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	if in.DefaultSLAHours <= 0 {
		in.DefaultSLAHours = 24
	}
	category := models.Category{Name: in.Name, DefaultSLAHours: in.DefaultSLAHours}
	if err := h.Categories.Create(c.Request.Context(), &category); err != nil {
		c.JSON(500, gin.H{"error": "failed to create category"})
		return
	}
	c.JSON(201, category)
}

// Update godoc
// @Summary Update a category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Category ID"
// @Param request body dto.CategoryRequest true "Category data"
// @Success 200 {object} models.Category
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/categories/{id} [patch]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	var in dto.CategoryRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	category := models.Category{ID: id, Name: in.Name, DefaultSLAHours: in.DefaultSLAHours}
	if err := h.Categories.Update(c.Request.Context(), &category); err != nil {
		c.JSON(500, gin.H{"error": "failed to update category"})
		return
	}
	c.JSON(200, category)
}

// Delete godoc
// @Summary Delete a category
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "Category ID"
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Categories.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete category"})
		return
	}
	c.JSON(200, gin.H{"message": "deleted"})
}

// SetSLARule godoc
// @Summary Create or update an SLA rule
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.SLARuleRequest true "SLA rule data"
// @Success 200 {object} models.SLARule
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sla-rules [post]
func (h *CategoryHandler) SetSLARule(c *gin.Context) {
	var in dto.SLARuleRequest
	if err := c.ShouldBindJSON(&in); err != nil || in.CategoryID == 0 || in.Priority == "" || in.ResolutionHours <= 0 {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	rule := models.SLARule{CategoryID: in.CategoryID, Priority: in.Priority, ResolutionHours: in.ResolutionHours}
	if err := h.Categories.UpsertSLARule(c.Request.Context(), &rule); err != nil {
		c.JSON(500, gin.H{"error": "failed to save sla rule"})
		return
	}
	c.JSON(200, rule)
}
