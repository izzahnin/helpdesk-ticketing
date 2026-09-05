package handlers

import "testing"

func TestValidUserRole(t *testing.T) {
	tests := []struct {
		role string
		want bool
	}{
		{role: "super_admin", want: true},
		{role: "admin", want: true},
		{role: "staff", want: true},
		{role: "end_user", want: true},
		{role: "", want: false},
		{role: "manager", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			got := validUserRole(tt.role)
			if got != tt.want {
				t.Fatalf("validUserRole(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestCanManageRole(t *testing.T) {
	tests := []struct {
		name       string
		actorRole  string
		targetRole string
		want       bool
	}{
		{name: "super admin can assign super admin", actorRole: "super_admin", targetRole: "super_admin", want: true},
		{name: "super admin can assign admin", actorRole: "super_admin", targetRole: "admin", want: true},
		{name: "super admin can assign staff", actorRole: "super_admin", targetRole: "staff", want: true},
		{name: "super admin can assign end user", actorRole: "super_admin", targetRole: "end_user", want: true},
		{name: "admin cannot assign super admin", actorRole: "admin", targetRole: "super_admin", want: false},
		{name: "admin cannot assign admin", actorRole: "admin", targetRole: "admin", want: false},
		{name: "admin can assign staff", actorRole: "admin", targetRole: "staff", want: true},
		{name: "admin can assign end user", actorRole: "admin", targetRole: "end_user", want: true},
		{name: "staff cannot assign role", actorRole: "staff", targetRole: "end_user", want: false},
		{name: "end user cannot assign role", actorRole: "end_user", targetRole: "staff", want: false},
		{name: "unknown actor cannot assign role", actorRole: "unknown", targetRole: "staff", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canManageRole(tt.actorRole, tt.targetRole)
			if got != tt.want {
				t.Fatalf("canManageRole(%q, %q) = %v, want %v", tt.actorRole, tt.targetRole, got, tt.want)
			}
		})
	}
}
