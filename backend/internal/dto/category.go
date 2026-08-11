package dto

type CategoryRequest struct {
	Name            string `json:"name" binding:"required"`
	DefaultSLAHours int    `json:"default_sla_hours"`
}

type SLARuleRequest struct {
	CategoryID      int64  `json:"category_id" binding:"required"`
	Priority        string `json:"priority" binding:"required"`
	ResolutionHours int    `json:"resolution_hours" binding:"required,min=1"`
}
