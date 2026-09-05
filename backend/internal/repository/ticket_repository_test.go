package repository

import (
	"testing"

	"helpdesk-backend/internal/models"
)

func TestEnsureTicketAccess(t *testing.T) {
	assigneeID := int64(20)
	ticket := &models.Ticket{
		RequesterID: 10,
		AssigneeID:  &assigneeID,
	}

	tests := []struct {
		name    string
		userID  int64
		role    string
		wantErr bool
	}{
		{name: "super admin can access any ticket", userID: 99, role: "super_admin", wantErr: false},
		{name: "admin can access any ticket", userID: 99, role: "admin", wantErr: false},
		{name: "assigned staff can access ticket", userID: assigneeID, role: "staff", wantErr: false},
		{name: "unassigned staff cannot access ticket", userID: 21, role: "staff", wantErr: true},
		{name: "requester can access own ticket", userID: 10, role: "end_user", wantErr: false},
		{name: "end user cannot access another user's ticket", userID: 11, role: "end_user", wantErr: true},
		{name: "unknown role cannot access ticket", userID: 10, role: "unknown", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureTicketAccess(ticket, tt.userID, tt.role)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EnsureTicketAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsureTicketAccessForStaffWithoutAssignee(t *testing.T) {
	ticket := &models.Ticket{RequesterID: 10}

	err := EnsureTicketAccess(ticket, 20, "staff")
	if err == nil {
		t.Fatal("EnsureTicketAccess() should reject staff when ticket has no assignee")
	}
}
