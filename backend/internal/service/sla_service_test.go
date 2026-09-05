package service

import (
	"testing"
	"time"
)

func TestSLAState(t *testing.T) {
	now := time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC)
	deadline := now.Add(2 * time.Hour)
	breachedDeadline := now.Add(-time.Minute)
	totalHours := 10
	zeroHours := 0

	tests := []struct {
		name       string
		deadline   *time.Time
		totalHours *int
		status     string
		now        time.Time
		want       string
	}{
		{name: "nil deadline is ok", deadline: nil, totalHours: &totalHours, status: "Open", now: now, want: "ok"},
		{name: "resolved ticket is ok", deadline: &breachedDeadline, totalHours: &totalHours, status: "Resolved", now: now, want: "ok"},
		{name: "closed ticket is ok", deadline: &breachedDeadline, totalHours: &totalHours, status: "Closed", now: now, want: "ok"},
		{name: "past deadline is breached", deadline: &breachedDeadline, totalHours: &totalHours, status: "Open", now: now, want: "breached"},
		{name: "within twenty percent remaining is at risk", deadline: &deadline, totalHours: &totalHours, status: "Open", now: now, want: "at_risk"},
		{name: "more than twenty percent remaining is ok", deadline: &deadline, totalHours: &totalHours, status: "Open", now: now.Add(-time.Hour), want: "ok"},
		{name: "nil total hours cannot be at risk", deadline: &deadline, totalHours: nil, status: "Open", now: now, want: "ok"},
		{name: "zero total hours cannot be at risk", deadline: &deadline, totalHours: &zeroHours, status: "Open", now: now, want: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SLAState(tt.deadline, tt.totalHours, tt.status, tt.now)
			if got != tt.want {
				t.Fatalf("SLAState() = %q, want %q", got, tt.want)
			}
		})
	}
}
