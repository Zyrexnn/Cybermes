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

func TestScopeTargetsField(t *testing.T) {
	// Root scope.yaml program format: `targets` allowlist instead of `in_scope`.
	yamlContent := `
program: "Cybermes Security Assessment & Bug Bounty"
authorization: "AUTHORIZED"
targets:
  - "*.example.com"
  - "http://127.0.0.1:8888"
out_of_scope:
  - "evil.example.com"
`
	cfg, err := ParseScopeYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Failed to parse scope YAML: %v", err)
	}
	if len(cfg.EffectiveInScope()) != 2 {
		t.Fatalf("Expected 2 effective in-scope rules from targets, got %v", cfg.EffectiveInScope())
	}

	tests := []struct {
		target   string
		expected bool
		name     string
	}{
		{"https://app.example.com/api", true, "Targets wildcard match"},
		{"http://127.0.0.1:8888/api/documents/101", true, "Targets URL match"},
		{"https://evil.example.com", false, "Out-of-scope exclusion wins over targets"},
		{"https://unrelated.example.org", false, "Unlisted target blocked"},
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

func TestScopeWildcardAllowAll(t *testing.T) {
	// Mirrors the shipped root scope.yaml: targets ["*"] allows operator
	// targets while out_of_scope exclusions still apply.
	yamlContent := `
targets:
  - "*"
out_of_scope:
  - "blocked.example.com"
`
	cfg, err := ParseScopeYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Failed to parse scope YAML: %v", err)
	}

	if res := ValidateTarget("https://anything.example.net/path", cfg); !res.Allowed {
		t.Errorf("Expected wildcard '*' to allow arbitrary target (reason: %s)", res.Reason)
	}
	if res := ValidateTarget("https://blocked.example.com", cfg); res.Allowed {
		t.Errorf("Expected out_of_scope to win over wildcard '*'")
	}
	// out_of_scope wildcard blocks everything.
	cfgDenyAll := &ScopeConfig{OutOfScope: []string{"*"}}
	if res := ValidateTarget("https://anything.example.net", cfgDenyAll); res.Allowed {
		t.Errorf("Expected out_of_scope '*' to block all targets")
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
	if res.ScopeFound {
		t.Errorf("Expected ScopeFound to be false when cfg is nil")
	}
}

func TestScopeEmptyRulesBlocked(t *testing.T) {
	cfgEmpty := &ScopeConfig{}
	res := ValidateTarget("https://example.org", cfgEmpty)
	if res.Allowed {
		t.Errorf("Expected empty scope configuration without in_scope or out_of_scope rules to be blocked")
	}
}

func TestShippedScopeConfigCanRefuse(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("could not resolve repository root: %v", err)
	}

	cfg, path, err := FindScopeConfig(root, "")
	if err != nil {
		t.Fatalf("FindScopeConfig found no scope file at repository root: %v", err)
	}

	// Shipped scope file must NOT allow arbitrary external targets like example.org
	res := ValidateTarget("https://example.org", cfg)
	if res.Allowed {
		t.Fatalf("the shipped scope file at %s authorises an arbitrary target: "+
			"allowed=%v scope_found=%v reason=%q",
			path, res.Allowed, res.ScopeFound, res.Reason)
	}

	// Shipped scope file MUST allow the local demo/lab targets
	resLocal := ValidateTarget("http://127.0.0.1:8888", cfg)
	if !resLocal.Allowed {
		t.Fatalf("the shipped scope file at %s blocked local test target: %v", path, resLocal.Reason)
	}

	resLocalHost := ValidateTarget("http://localhost:8888/api/login", cfg)
	if !resLocalHost.Allowed {
		t.Fatalf("the shipped scope file at %s blocked localhost test target: %v", path, resLocalHost.Reason)
	}

	// Shipped scope file out_of_scope must block .gov, .mil, .edu
	resGov := ValidateTarget("https://agency.gov", cfg)
	if resGov.Allowed {
		t.Fatalf("the shipped scope file allowed out_of_scope target .gov")
	}
}

func TestScopeDynamicTargetOverride(t *testing.T) {
	yamlContent := `
program: "Test Bounty"
dynamic_target_override: true
targets:
  - "http://127.0.0.1:8888"
out_of_scope:
  - "blocked.example.com"
`
	cfg, err := ParseScopeYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("ParseScopeYAML failed: %v", err)
	}
	if !cfg.DynamicTargetOverride {
		t.Errorf("Expected DynamicTargetOverride to be true, got false")
	}
	if !cfg.IsUnrestricted() {
		t.Errorf("Expected IsUnrestricted() to be true when DynamicTargetOverride is true")
	}

	// Unlisted target should be allowed because dynamic_target_override is true
	res := ValidateTarget("https://unlisted-operator-target.com", cfg)
	if !res.Allowed {
		t.Errorf("Expected target to be allowed under dynamic target override, got blocked: %s", res.Reason)
	}
	if res.MatchedBy != "dynamic_target_override: true" {
		t.Errorf("Expected MatchedBy to be 'dynamic_target_override: true', got: %s", res.MatchedBy)
	}

	// Out of scope must STILL be blocked even with dynamic_target_override: true
	resBlocked := ValidateTarget("https://blocked.example.com", cfg)
	if resBlocked.Allowed {
		t.Errorf("Expected out_of_scope target to be blocked even with dynamic_target_override: true")
	}
}

func TestScaffoldTargetScope(t *testing.T) {
	tmpDir := t.TempDir()
	targetSlug := "test_client"
	initialTarget := "https://api.testclient.org"

	filePath, err := ScaffoldTargetScope(tmpDir, targetSlug, initialTarget)
	if err != nil {
		t.Fatalf("ScaffoldTargetScope failed: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Scaffold file was not created at %s", filePath)
	}

	cfg, err := ParseScopeFile(filePath)
	if err != nil {
		t.Fatalf("Failed to parse generated scaffold: %v", err)
	}

	val := ValidateTarget(initialTarget, cfg)
	if !val.Allowed {
		t.Errorf("Generated scaffold must allow initial target, got: %s", val.Reason)
	}

	// Idempotency: calling again should return existing path without error
	path2, err := ScaffoldTargetScope(tmpDir, targetSlug, initialTarget)
	if err != nil || path2 != filePath {
		t.Errorf("Expected idempotent return of existing scaffold path, got: %v (err: %v)", path2, err)
	}
}
