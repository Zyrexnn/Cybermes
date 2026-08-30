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

## Description
Unauthorized access to user profile data.
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

	htmlPath := filepath.Join(tempDir, "report.html")
	htmlData, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("report.html was not generated: %v", err)
	}

	if !strings.Contains(string(htmlData), "IDOR in Profile Endpoint") {
		t.Errorf("report.html missing finding title")
	}
}

func TestGenerateHTMLDashboard(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "html_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	data := &SummaryData{
		Target:        "example_com",
		ScanTime:      "2026-08-29 23:00:00",
		TotalFindings: 1,
		SeveritySummary: map[string]int{
			"CRITICAL":      1,
			"HIGH":          0,
			"MEDIUM":        0,
			"LOW":           0,
			"INFORMATIONAL": 0,
		},
		Findings: []*FindingMeta{
			{
				Title:    "SQL Injection in Login",
				Severity: "CRITICAL",
				CVSS:     "9.8",
				CWE:      "CWE-89",
				Endpoint: "/api/login",
				FileName: "critical_sqli.md",
			},
		},
	}

	htmlPath, err := GenerateHTMLDashboard(tempDir, data)
	if err != nil {
		t.Fatalf("GenerateHTMLDashboard failed: %v", err)
	}

	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read generated HTML: %v", err)
	}

	htmlStr := string(content)
	if !strings.Contains(htmlStr, "example_com") {
		t.Errorf("HTML missing target name")
	}
	if !strings.Contains(htmlStr, "SQL Injection in Login") {
		t.Errorf("HTML missing finding title")
	}
	if !strings.Contains(htmlStr, "CRITICAL") {
		t.Errorf("HTML missing severity badge")
	}
}

func TestFindChromiumBrowser(t *testing.T) {
	// FindChromiumBrowser should not panic and return string (either empty or valid path)
	browser := FindChromiumBrowser()
	t.Logf("Detected browser on host: %q", browser)
}

func TestAggregateTargetWithPDF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "full_report_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingsDir := filepath.Join(tempDir, "findings")
	os.MkdirAll(findingsDir, 0755)

	findingPath := filepath.Join(findingsDir, "medium_cors.md")
	content := `# Permissive CORS Configuration

- **Severity**: MEDIUM
- **Endpoint**: /api/data

## Description
Access-Control-Allow-Origin header is set to wildcard with credentials.
`
	os.WriteFile(findingPath, []byte(content), 0644)

	summary, artifacts, err := AggregateTargetWithPDF(tempDir, true)
	if err != nil {
		t.Fatalf("AggregateTargetWithPDF failed: %v", err)
	}

	if summary.TotalFindings != 1 {
		t.Errorf("expected 1 finding, got %d", summary.TotalFindings)
	}

	if artifacts.HTMLPath == "" {
		t.Errorf("expected HTMLPath to be populated")
	}

	if _, err := os.Stat(artifacts.HTMLPath); os.IsNotExist(err) {
		t.Errorf("report.html does not exist on disk: %s", artifacts.HTMLPath)
	}
}

func TestBuildFileURL(t *testing.T) {
	// Test POSIX Linux/Docker path
	posixPath := "/workspace/reports/target/report.html"
	posixURL := BuildFileURL(posixPath)
	if posixURL != "file:///workspace/reports/target/report.html" {
		t.Errorf("unexpected POSIX URL: %q", posixURL)
	}

	// Test Windows drive letter path
	windowsPath := "C:/Users/name/reports/target/report.html"
	windowsURL := BuildFileURL(windowsPath)
	if windowsURL != "file:///C:/Users/name/reports/target/report.html" {
		t.Errorf("unexpected Windows URL: %q", windowsURL)
	}
}
