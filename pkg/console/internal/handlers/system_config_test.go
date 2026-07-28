package handlers

import "testing"

func TestIsSystemConfigurationAdminRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{name: "main account", role: "main", want: true},
		{name: "built-in admin", role: "admin", want: true},
		{name: "ordinary account", role: "sub", want: false},
		{name: "missing role", role: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSystemConfigurationAdminRole(tt.role); got != tt.want {
				t.Fatalf("isSystemConfigurationAdminRole(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}
