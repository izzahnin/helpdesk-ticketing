package service

import (
	"context"
	"time"

	"helpdesk-backend/internal/repository"
)

type SLAService struct {
	Categories *repository.CategoryRepository
}

func NewSLAService(categories *repository.CategoryRepository) *SLAService {
	return &SLAService{Categories: categories}
}

func (s *SLAService) ResolveHours(ctx context.Context, categoryID int64, priority string) (int, error) {
	return s.Categories.ResolveSLAHours(ctx, categoryID, priority)
}

func (s *SLAService) CalculateDeadline(ctx context.Context, categoryID int64, priority string, createdAt time.Time) (time.Time, int, error) {
	hours, err := s.ResolveHours(ctx, categoryID, priority)
	if err != nil {
		return time.Time{}, 0, err
	}

	return createdAt.Add(time.Duration(hours) * time.Hour), hours, nil
}

func SLAState(deadline *time.Time, totalHours *int, status string, now time.Time) string {
	if deadline == nil || status == "Resolved" || status == "Closed" {
		return "ok"
	}
	if now.After(*deadline) {
		return "breached"
	}
	if totalHours == nil || *totalHours <= 0 {
		return "ok"
	}

	remaining := deadline.Sub(now)
	total := time.Duration(*totalHours) * time.Hour
	if remaining <= total/5 {
		return "at_risk"
	}

	return "ok"
}
