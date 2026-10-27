package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestHelpers_SanitizeAndResolveSlug(t *testing.T) {
	s := &Server{}

	if got := sanitizeSlug("Example.com/API-v1! test"); got != "example_com_api_v1_test" {
		t.Errorf("sanitizeSlug failed, got: %s", got)
	}

	if got := s.resolveSlug("", "https://api.target.com"); got != "https_api_target_com" {
		t.Errorf("resolveSlug with default failed, got: %s", got)
	}

	if got := s.resolveSlug("my_target_slug", "https://api.target.com"); got != "my_target_slug" {
		t.Errorf("resolveSlug with explicit failed, got: %s", got)
	}
}

func TestHelpers_ParseCustomHeaders(t *testing.T) {
	// JSON format
	rawJSON := `{"Authorization": "Bearer test123", "X-Custom-ID": 999}`
	headers := parseCustomHeaders(rawJSON)
	if headers["Authorization"] != "Bearer test123" || headers["X-Custom-ID"] != "999" {
		t.Errorf("parseCustomHeaders JSON failed, got: %v", headers)
	}

	// Line by line format
	rawLines := "Authorization: Bearer line_token\nX-Tenant: test-corp\n# comment line\n\n"
	headers2 := parseCustomHeaders(rawLines)
	if headers2["Authorization"] != "Bearer line_token" || headers2["X-Tenant"] != "test-corp" {
		t.Errorf("parseCustomHeaders lines failed, got: %v", headers2)
	}

	// Empty
	if got := parseCustomHeaders(""); got != nil {
		t.Errorf("expected nil for empty headers, got: %v", got)
	}
}

func TestHelpers_SeverityBadges(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"critical", "🔴 **CRITICAL**"},
		{"CRITICAL", "🔴 **CRITICAL**"},
		{"high", "🟠 **HIGH**"},
		{"medium", "🟡 **MEDIUM**"},
		{"low", "🔵 **LOW**"},
		{"info", "🔵 **LOW**"},
	}

	for _, tc := range tests {
		if got := getSeverityBadge(tc.input); got != tc.expected {
			t.Errorf("getSeverityBadge(%s) = %s, expected %s", tc.input, got, tc.expected)
		}
	}
}

func TestHelpers_IsSafeSubpath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cybermes-path-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "subfolder")
	_ = os.MkdirAll(subDir, 0755)
	safeFile := filepath.Join(subDir, "file.txt")

	if !isSafeSubpath(tmpDir, safeFile) {
		t.Errorf("expected %s to be safe inside %s", safeFile, tmpDir)
	}

	// Traversal attempt
	traversalPath := filepath.Join(tmpDir, "..", "outside.txt")
	if isSafeSubpath(tmpDir, traversalPath) {
		t.Errorf("expected traversal path %s to be rejected", traversalPath)
	}

	// Same dir
	if isSafeSubpath(tmpDir, tmpDir) {
		t.Errorf("base dir itself should not be considered a subpath")
	}
}

func TestResource_PathTraversalProtection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cybermes-res-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	skillsDir := filepath.Join(tmpDir, "skills")
	reportsDir := filepath.Join(tmpDir, "reports")
	_ = os.MkdirAll(skillsDir, 0755)
	_ = os.MkdirAll(reportsDir, 0755)

	cfg := Config{
		RootDir:    tmpDir,
		SkillsDir:  skillsDir,
		ReportsDir: reportsDir,
	}
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	// Attempt path traversal via skills://
	_, err = s.handleReadSkillResource(context.Background(), mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "skills://../../etc/passwd",
		},
	})
	if err == nil {
		t.Errorf("expected error for path traversal in skills resource, got nil")
	}

	// Attempt path traversal via reports://
	_, err = s.handleReadReportResource(context.Background(), mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "reports://../../etc/passwd/summary",
		},
	})
	if err == nil {
		t.Errorf("expected error for path traversal in reports resource, got nil")
	}
}
