package cmd

import (
	"strings"
	"testing"
)

func TestAddAccount(t *testing.T) {
	tests := []struct {
		username string
		password string
		want     string
	}{
		{"", "", "Username must not be empty"},
		{"abc", "abc", "Password must be at least"},
		{"abc", "fooBar123", ""},
	}
	for _, tt := range tests {
		err := addAccount(tt.username, tt.password)
		if err != nil {
			if tt.want == "" {
				t.Errorf("got unexpected error: %v", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("expected error containing '%s', got error '%v'", err, tt.want)
			}
			continue
		}
		if tt.want != "" {
			t.Errorf("expected error '%s', got no error", tt.want)
		}
	}
}

func TestPrintAccounts(t *testing.T) {
	if err := addAccount("foo", "fooBar123"); err != nil {
		t.Errorf("failed to add test account: %v", err)
		return
	}
	var b strings.Builder
	err := printAccounts("", &b)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	got := b.String()
	if !strings.Contains(got, "foo") {
		t.Errorf("expected string containing 'foo', got '%s'", got)
	}
}
