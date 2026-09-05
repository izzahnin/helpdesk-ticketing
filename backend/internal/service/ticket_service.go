package service

import (
	"context"
	"errors"
	"time"

	"helpdesk-backend/internal/models"
	"helpdesk-backend/internal/repository"
)

type TicketService struct {
	Repo *repository.TicketRepository
	SLA  *SLAService
}

func NewTicketService(repo *repository.TicketRepository, sla *SLAService) *TicketService {
	return &TicketService{Repo: repo, SLA: sla}
}

func (s *TicketService) ListForUser(ctx context.Context, userID int64, role string) ([]models.Ticket, error) {
	switch role {
	case "admin", "super_admin":
		return s.Repo.FindAll(ctx)
	case "staff":
		return s.Repo.FindByAssignee(ctx, userID)
	default:
		return s.Repo.FindByRequester(ctx, userID)
	}
}

func (s *TicketService) Submit(ctx context.Context, t *models.Ticket) error {
	if t.Priority == "" {
		t.Priority = "Medium"
	}
	t.Status = "Open"
	deadline, _, err := s.SLA.CalculateDeadline(ctx, t.CategoryID, t.Priority, time.Now())
	if err != nil {
		return err
	}
	t.SLADeadline = &deadline
	return s.Repo.CreateWithLog(ctx, t)
}

func (s *TicketService) ChangeStatus(ctx context.Context, ticketID, actorID int64, newStatus string) error {
	t, err := s.Repo.FindByID(ctx, ticketID)
	if err != nil {
		return err
	}
	if !TransitionAllowed(t.Status, newStatus) {
		return errors.New("invalid status transition")
	}
	return s.Repo.UpdateStatusWithLog(ctx, ticketID, actorID, t.Status, newStatus)
}

func TransitionAllowed(oldStatus, newStatus string) bool {
	allowed := map[string][]string{
		"Open":        {"In Progress"},
		"In Progress": {"Pending", "Resolved"},
		"Pending":     {"In Progress", "Resolved"},
		"Resolved":    {"Closed", "Open"},
		"Closed":      {"Open"},
	}
	for _, s := range allowed[oldStatus] {
		if s == newStatus {
			return true
		}
	}
	return false
}
