package dto

type UpdateUserRequest struct {
	Role     string `json:"role" binding:"required"`
	IsActive bool   `json:"is_active"`
}
