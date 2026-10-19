package handlers

import "testing"

func TestAgentInputAllowed(t *testing.T) {
	tests := []struct {
		name    string
		owner   bool
		role    string
		hasRole bool
		want    bool
	}{
		{name: "owner", owner: true, want: true},
		{name: "owner ignores missing role", owner: true, hasRole: false, want: true},
		{name: "editor", role: "editor", hasRole: true, want: true},
		{name: "collab owner", role: "owner", hasRole: true, want: true},
		{name: "viewer", role: "viewer", hasRole: true, want: false},
		{name: "missing role fails closed", hasRole: false, want: false},
		{name: "empty role fails closed", role: "", hasRole: true, want: false},
		{name: "unknown role fails closed", role: "admin", hasRole: true, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := agentInputAllowed(tt.owner, tt.role, tt.hasRole); got != tt.want {
				t.Fatalf("agentInputAllowed(%v, %q, %v) = %v, want %v", tt.owner, tt.role, tt.hasRole, got, tt.want)
			}
		})
	}
}
