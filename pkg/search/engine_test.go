package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearcher(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kb_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	payloadsDir := filepath.Join(tempDir, "PayloadsAllTheThings")
	if err := os.MkdirAll(payloadsDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	testFile := filepath.Join(payloadsDir, "sqli.md")
	content := `# SQL Injection

## Authentication Bypass

Here are SQL injection payloads:

` + "```sql" + `
admin' OR '1'='1' --
` + "```" + `

## Blind SQLi
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	s := NewSearcher(tempDir, tempDir)
	snippets, err := s.Search("sql injection bypass", "all", 2, 1000)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(snippets) == 0 {
		t.Fatalf("expected at least 1 snippet, got 0")
	}

	if !strings.Contains(snippets[0].Content, "admin' OR '1'='1'") {
		t.Errorf("snippet content missing expected payload: %s", snippets[0].Content)
	}

	if snippets[0].SourceKB != "PayloadsAllTheThings" {
		t.Errorf("expected SourceKB 'PayloadsAllTheThings', got %q", snippets[0].SourceKB)
	}
}
