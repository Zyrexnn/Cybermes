package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFindingFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "report_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingPath := filepath.Join(tempDir, "critical_rce.md")
	content := `---
title: Remote Code Execution via File Upload
severity: CRITICAL
---

# Remote Code Execution via File Upload

| Field | Value |
| :--- | :--- |
| Severity | CRITICAL |
| CVSS v3.1 | 9.8 |
| CWE | CWE-434 |
| Affected Endpoint | /api/v1/upload |
`
	if err := os.WriteFile(findingPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write finding file: %v", err)
	}

	meta, err := ParseFindingFile(findingPath)
	if err != nil {
		t.Fatalf("ParseFindingFile failed: %v", err)
	}

	if meta.Title != "Remote Code Execution via File Upload" {
		t.Errorf("unexpected title: %q", meta.Title)
	}
	if meta.Severity != "CRITICAL" {
		t.Errorf("unexpected severity: %q", meta.Severity)
	}
	if meta.CVSS != "9.8" {
		t.Errorf("unexpected CVSS: %q", meta.CVSS)
	}
	if meta.CWE != "CWE-434" {
		t.Errorf("unexpected CWE: %q", meta.CWE)
	}
	if meta.Endpoint != "/api/v1/upload" {
		t.Errorf("unexpected endpoint: %q", meta.Endpoint)
	}
}

func TestAggregateTarget(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "target_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingsDir := filepath.Join(tempDir, "findings")
	os.MkdirAll(findingsDir, 0755)

	findingPath := filepath.Join(findingsDir, "high_idor.md")
	content := `# IDOR in Profile Endpoint

| Property | Details |
| :--- | :--- |
| Severity | HIGH |
| CVSS | 8.1 |
| CWE | 639 |
| Endpoint | /api/v1/users/{id} |
`
	os.WriteFile(findingPath, []byte(content), 0644)

	summary, err := AggregateTarget(tempDir)
	if err != nil {
		t.Fatalf("AggregateTarget failed: %v", err)
	}

	if summary.TotalFindings != 1 {
		t.Errorf("expected 1 finding, got %d", summary.TotalFindings)
	}
	if summary.SeveritySummary["HIGH"] != 1 {
		t.Errorf("expected 1 HIGH finding, got %d", summary.SeveritySummary["HIGH"])
	}

	summaryMDPath := filepath.Join(tempDir, "SUMMARY.md")
	summaryData, err := os.ReadFile(summaryMDPath)
	if err != nil {
		t.Fatalf("SUMMARY.md was not generated: %v", err)
	}

	if !strings.Contains(string(summaryData), "IDOR in Profile Endpoint") {
		t.Errorf("SUMMARY.md missing finding title")
	}
}
