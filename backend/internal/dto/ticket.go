package dto

type SubmitTicketRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	CategoryID  int64  `json:"category_id" binding:"required"`
	Priority    string `json:"priority"`
}

type CommentRequest struct {
	Message string `json:"message" binding:"required"`
}

type AssignTicketRequest struct {
	AssigneeID int64 `json:"assignee_id" binding:"required"`
}

type UpdateTicketStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
