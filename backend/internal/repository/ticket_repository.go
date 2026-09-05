package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"helpdesk-backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type TicketRepository struct {
	db *sqlx.DB
}

func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func ticketSelect() string {
	return `
		SELECT t.id, t.title, t.description, t.requester_id, t.assignee_id,
		       t.category_id, t.priority, t.status, t.sla_deadline,
		       t.resolved_at, t.created_at, t.updated_at,
		       c.name AS category_name,
		       COALESCE(sr.resolution_hours, c.default_sla_hours) AS sla_total_hours
		FROM tickets t
		JOIN categories c ON c.id=t.category_id
		LEFT JOIN sla_rules sr ON sr.category_id=t.category_id AND sr.priority=t.priority
	`
}

func (r *TicketRepository) CreateWithLog(ctx context.Context, t *models.Ticket) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO tickets (title, description, requester_id, category_id, priority, status, sla_deadline)
		VALUES ($1,$2,$3,$4,$5,'Open',$6)
		RETURNING id, created_at, updated_at
	`, t.Title, t.Description, t.RequesterID, t.CategoryID, t.Priority, t.SLADeadline).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_status_log (ticket_id, old_status, new_status, changed_by)
		VALUES ($1,NULL,'Open',$2)
	`, t.ID, t.RequesterID)
	if err != nil {
		return err
	}

	newValue := "Open"
	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value)
		VALUES ($1,$2,'ticket_created',NULL,$3)
	`, t.ID, t.RequesterID, newValue)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TicketRepository) FindAll(ctx context.Context) ([]models.Ticket, error) {
	var rows []models.Ticket
	err := r.db.SelectContext(ctx, &rows, ticketSelect()+` ORDER BY t.created_at DESC`)
	return rows, err
}

func (r *TicketRepository) FindByRequester(ctx context.Context, userID int64) ([]models.Ticket, error) {
	var rows []models.Ticket
	err := r.db.SelectContext(ctx, &rows, ticketSelect()+` WHERE t.requester_id=$1 ORDER BY t.created_at DESC`, userID)
	return rows, err
}

func (r *TicketRepository) FindByAssignee(ctx context.Context, userID int64) ([]models.Ticket, error) {
	var rows []models.Ticket
	err := r.db.SelectContext(ctx, &rows, ticketSelect()+` WHERE t.assignee_id=$1 ORDER BY t.created_at DESC`, userID)
	return rows, err
}

func (r *TicketRepository) FindByID(ctx context.Context, id int64) (*models.Ticket, error) {
	var row models.Ticket
	err := r.db.GetContext(ctx, &row, ticketSelect()+` WHERE t.id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTicketNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *TicketRepository) FindComments(ctx context.Context, ticketID int64) ([]models.TicketComment, error) {
	var rows []models.TicketComment
	err := r.db.SelectContext(ctx, &rows, `
		SELECT tc.id, tc.ticket_id, tc.user_id, u.name AS author_name, u.role AS author_role,
		       tc.message, tc.created_at
		FROM ticket_comments tc
		JOIN users u ON u.id=tc.user_id
		WHERE tc.ticket_id=$1
		ORDER BY tc.created_at ASC
	`, ticketID)
	return rows, err
}

func (r *TicketRepository) AddComment(ctx context.Context, ticketID, userID int64, message string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existingTicketID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM tickets WHERE id=$1 FOR UPDATE`, ticketID).Scan(&existingTicketID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTicketNotFound
	}
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_comments (ticket_id, user_id, message)
		VALUES ($1,$2,$3)
	`, ticketID, userID, message)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value)
		VALUES ($1,$2,'comment_added',NULL,$3)
	`, ticketID, userID, "comment")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TicketRepository) AssignWithActivityLog(ctx context.Context, ticketID, assigneeID, actorID int64) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var oldAssignee sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT assignee_id FROM tickets WHERE id=$1 FOR UPDATE`, ticketID).Scan(&oldAssignee)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTicketNotFound
	}
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `UPDATE tickets SET assignee_id=$1, updated_at=now() WHERE id=$2`, assigneeID, ticketID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTicketNotFound
	}

	action := "assigned"
	var oldValue *string
	if oldAssignee.Valid {
		action = "reassigned"
		v := strconv.FormatInt(oldAssignee.Int64, 10)
		oldValue = &v
	}
	newValue := strconv.FormatInt(assigneeID, 10)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value)
		VALUES ($1,$2,$3,$4,$5)
	`, ticketID, actorID, action, oldValue, newValue)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *TicketRepository) UpdateStatusWithLog(ctx context.Context, ticketID, actorID int64, oldStatus, newStatus string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRowContext(ctx, `SELECT status FROM tickets WHERE id=$1 FOR UPDATE`, ticketID).Scan(&oldStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTicketNotFound
	}
	if err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE tickets
		SET status=$1,
		    updated_at=now(),
		    resolved_at=CASE WHEN $1 = 'Resolved' THEN now() ELSE resolved_at END
		WHERE id=$2
	`, newStatus, ticketID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTicketNotFound
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_status_log (ticket_id, old_status, new_status, changed_by)
		VALUES ($1,$2,$3,$4)
	`, ticketID, oldStatus, newStatus, actorID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_activity_log (ticket_id, actor_id, action, old_value, new_value)
		VALUES ($1,$2,'status_changed',$3,$4)
	`, ticketID, actorID, oldStatus, newStatus)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func EnsureTicketAccess(t *models.Ticket, userID int64, role string) error {
	if role == "admin" || role == "super_admin" {
		return nil
	}
	if role == "staff" && t.AssigneeID != nil && *t.AssigneeID == userID {
		return nil
	}
	if role == "end_user" && t.RequesterID == userID {
		return nil
	}
	return fmt.Errorf("forbidden")
}
