package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"helpdesk-backend/internal/dto"
	"helpdesk-backend/internal/repository"
)

type UserHandler struct{ Users *repository.UserRepository }

// List godoc
// @Summary List all users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /api/users [get]
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.Users.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch users"})
		return
	}
	c.JSON(200, users)
}

// Staff godoc
// @Summary List active staff users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /api/users/staff [get]
func (h *UserHandler) Staff(c *gin.Context) {
	users, err := h.Users.FindStaff(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch staff"})
		return
	}
	c.JSON(200, users)
}

// Update godoc
// @Summary Update a user's role and active status
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int64 true "User ID"
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/{id} [patch]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	var in dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	if !validUserRole(in.Role) {
		c.JSON(400, gin.H{"error": "invalid role"})
		return
	}
	if !canManageRole(c.GetString("role"), in.Role) {
		c.JSON(403, gin.H{"error": "forbidden"})
		return
	}
	if err := h.Users.UpdateRoleStatus(c.Request.Context(), id, in.Role, in.IsActive); err != nil {
		c.JSON(500, gin.H{"error": "failed to update user"})
		return
	}
	c.JSON(200, gin.H{"message": "updated"})
}

func validUserRole(role string) bool {
	switch role {
	case "super_admin", "admin", "staff", "end_user":
		return true
	default:
		return false
	}
}

func canManageRole(actorRole, targetRole string) bool {
	if actorRole == "super_admin" {
		return true
	}
	if actorRole == "admin" {
		return targetRole == "staff" || targetRole == "end_user"
	}
	return false
}
