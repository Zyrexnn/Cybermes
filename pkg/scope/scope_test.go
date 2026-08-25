package scope

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScopeValidation(t *testing.T) {
	yamlContent := `
name: "Acme Bug Bounty Scope"
target_slug: "acme_corp"
in_scope:
  - "*.acme.com"
  - "api.payment.internal"
  - "192.168.1.0/24"
  - "10.0.0.5"
  - "http://staging.partner.org:8080"
out_of_scope:
  - "admin.acme.com"
  - "*.dev.acme.com"
  - "/logout"
  - "/auth/reset-password"
  - "192.168.1.100"
`

	cfg, err := ParseScopeYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Failed to parse scope YAML: %v", err)
	}

	tests := []struct {
		target   string
		expected bool
		name     string
	}{
		{"https://acme.com", true, "Root domain match"},
		{"https://sub.acme.com/api/v1", true, "Subdomain wildcard match"},
		{"http://nested.deep.acme.com", true, "Multi-level subdomain match"},
		{"https://admin.acme.com", false, "Explicit out_of_scope domain"},
		{"https://test.dev.acme.com", false, "Wildcard out_of_scope domain"},
		{"https://acme.com/logout", false, "Excluded path /logout"},
		{"https://acme.com/auth/reset-password", false, "Excluded path /auth/reset-password"},
		{"http://192.168.1.50:8080/app", true, "In-scope CIDR range"},
		{"http://192.168.1.100:8080", false, "Out-of-scope specific IP inside CIDR"},
		{"http://10.0.0.5", true, "Exact IP match"},
		{"http://10.0.0.6", false, "Unlisted IP"},
		{"http://staging.partner.org:8080/dashboard", true, "Host and port match"},
		{"http://staging.partner.org:9090", false, "Port mismatch"},
		{"https://google.com", false, "Unrelated domain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ValidateTarget(tt.target, cfg)
			if res.Allowed != tt.expected {
				t.Errorf("Target '%s' expected allowed=%v, got allowed=%v (reason: %s)",
					tt.target, tt.expected, res.Allowed, res.Reason)
			}
		})
	}
}

func TestScopeFindConfig(t *testing.T) {
	tmpDir := t.TempDir()
	reportsDir := filepath.Join(tmpDir, "reports", "test_target")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		t.Fatalf("Failed to create temp reports dir: %v", err)
	}

	scopeFile := filepath.Join(reportsDir, "scope.yaml")
	content := `
in_scope:
  - "target.test"
out_of_scope:
  - "secret.target.test"
`
	if err := os.WriteFile(scopeFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write scope file: %v", err)
	}

	cfg, foundPath, err := FindScopeConfig(tmpDir, "test_target")
	if err != nil {
		t.Fatalf("FindScopeConfig failed: %v", err)
	}

	if cfg == nil || len(cfg.InScope) == 0 {
		t.Fatal("Expected non-nil cfg with in_scope rules")
	}

	if foundPath != scopeFile {
		t.Errorf("Expected foundPath=%s, got=%s", scopeFile, foundPath)
	}

	res := ValidateTarget("https://target.test/login", cfg)
	if !res.Allowed {
		t.Errorf("Expected target.test to be allowed")
	}

	resSecret := ValidateTarget("https://secret.target.test", cfg)
	if resSecret.Allowed {
		t.Errorf("Expected secret.target.test to be blocked")
	}
}

func TestScopeDirectOperatorAuthFallback(t *testing.T) {
	// If cfg is nil, Direct Operator Authorization applies
	res := ValidateTarget("https://custom-target.com:8443/api", nil)
	if !res.Allowed {
		t.Errorf("Expected target to be allowed under Direct Operator Authorization when no scope file is loaded")
	}
	if res.Host != "custom-target.com" || res.Port != "8443" {
		t.Errorf("Expected parsed host and port, got host=%s port=%s", res.Host, res.Port)
	}
}
