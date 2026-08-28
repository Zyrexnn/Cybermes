package subdomain

import (
	"strings"
	"testing"
)

func TestCleanDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://api.example.com/v1", "api.example.com"},
		{"http://*.admin.example.com:8080", "admin.example.com"},
		{"*.dev.example.com", "dev.example.com"},
		{"Example.COM", "example.com"},
		{"sub.target.co.id/path/test", "sub.target.co.id"},
	}

	for _, tt := range tests {
		got := CleanDomain(tt.input)
		if got != tt.expected {
			t.Errorf("CleanDomain(%q) = %q; expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestSubdomainFiltering(t *testing.T) {
	rawInputs := []string{
		"api.example.com",
		"*.example.com",
		"auth.example.com",
		"staging.example.com",
		"unrelated.otherdomain.com",
		"https://portal.example.com/login",
	}

	domain := "example.com"
	uniqueMap := make(map[string]bool)
	var cleanedSubs []string

	for _, raw := range rawInputs {
		s := CleanDomain(raw)
		if s == "" || !strings.HasSuffix(s, domain) {
			continue
		}
		if !uniqueMap[s] {
			uniqueMap[s] = true
			cleanedSubs = append(cleanedSubs, s)
		}
	}

	if len(cleanedSubs) != 5 {
		t.Errorf("Expected 5 unique subdomains (api, example.com, auth, staging, portal), got %d: %v", len(cleanedSubs), cleanedSubs)
	}
}
