package service

import "testing"

func TestTransitionAllowed(t *testing.T) {
	tests := []struct {
		name      string
		oldStatus string
		newStatus string
		want      bool
	}{
		{name: "open to in progress", oldStatus: "Open", newStatus: "In Progress", want: true},
		{name: "open to resolved is invalid", oldStatus: "Open", newStatus: "Resolved", want: false},
		{name: "in progress to pending", oldStatus: "In Progress", newStatus: "Pending", want: true},
		{name: "in progress to resolved", oldStatus: "In Progress", newStatus: "Resolved", want: true},
		{name: "pending to in progress", oldStatus: "Pending", newStatus: "In Progress", want: true},
		{name: "pending to resolved", oldStatus: "Pending", newStatus: "Resolved", want: true},
		{name: "resolved to closed", oldStatus: "Resolved", newStatus: "Closed", want: true},
		{name: "resolved to open", oldStatus: "Resolved", newStatus: "Open", want: true},
		{name: "closed to open", oldStatus: "Closed", newStatus: "Open", want: true},
		{name: "closed to resolved is invalid", oldStatus: "Closed", newStatus: "Resolved", want: false},
		{name: "unknown old status is invalid", oldStatus: "Unknown", newStatus: "Open", want: false},
		{name: "empty statuses are invalid", oldStatus: "", newStatus: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransitionAllowed(tt.oldStatus, tt.newStatus)
			if got != tt.want {
				t.Fatalf("TransitionAllowed(%q, %q) = %v, want %v", tt.oldStatus, tt.newStatus, got, tt.want)
			}
		})
	}
}
