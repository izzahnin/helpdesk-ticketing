package dto

type UserResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Role     string `json:"role" binding:"required"`
	IsActive bool   `json:"is_active"`
}
