package report

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
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

func TestParseFindingFile_RecordFindingFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "report_test_record_finding")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingPath := filepath.Join(tempDir, "high_unauthenticated_device.md")
	content := `# Unauthenticated Device Configuration Access

- **Severity**: HIGH
- **Endpoint**: ` + "`" + `GET http://target:8000/api/v1/device_config` + "`" + `
- **Date**: 2026-09-05

## Description

The API endpoint exposes sensitive device configuration without requiring authentication.
Note: Target parameter or URL in description should not override the endpoint.

## Steps to Reproduce

1. Send GET request to target endpoint.
`
	if err := os.WriteFile(findingPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write finding file: %v", err)
	}

	meta, err := ParseFindingFile(findingPath)
	if err != nil {
		t.Fatalf("ParseFindingFile failed: %v", err)
	}

	if meta.Title != "Unauthenticated Device Configuration Access" {
		t.Errorf("unexpected title: %q", meta.Title)
	}
	if meta.Severity != "HIGH" {
		t.Errorf("expected severity HIGH, got %q", meta.Severity)
	}
	if meta.Endpoint != "GET http://target:8000/api/v1/device_config" {
		t.Errorf("expected endpoint 'GET http://target:8000/api/v1/device_config', got %q", meta.Endpoint)
	}
}

func TestParseFindingFile_LowercasePrefixFallback(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "report_test_prefix")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingPath := filepath.Join(tempDir, "medium_cors_misconfiguration.md")
	content := `# CORS Misconfiguration

No explicit severity table or key-value list here.
`
	if err := os.WriteFile(findingPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write finding file: %v", err)
	}

	meta, err := ParseFindingFile(findingPath)
	if err != nil {
		t.Fatalf("ParseFindingFile failed: %v", err)
	}

	if meta.Severity != "MEDIUM" {
		t.Errorf("expected fallback severity MEDIUM from filename prefix, got %q", meta.Severity)
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

func TestAggregateArtifactPermissions(t *testing.T) {
	// Regression test: report artifacts must never be world-writable.
	// Pentest evidence tampered by another local user must be impossible.
	//
	// NOTE: file mode arguments to os.WriteFile are filtered by the process
	// umask, so the test neutralizes it first. Otherwise a typical umask of
	// 022 would mask the bug (0666 silently landing as 0644) and the test
	// would pass on unfixed code.
	oldMask := syscall.Umask(0)
	defer syscall.Umask(oldMask)

	tempDir, err := os.MkdirTemp("", "perms_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingsDir := filepath.Join(tempDir, "findings")
	if err := os.MkdirAll(findingsDir, 0755); err != nil {
		t.Fatalf("failed to create findings dir: %v", err)
	}
	findingPath := filepath.Join(findingsDir, "high_idor.md")
	content := `# IDOR in Profile Endpoint

| Property | Details |
| :--- | :--- |
| Severity | HIGH |
| CVSS | 8.1 |
| CWE | 639 |
| Endpoint | /api/v1/users/{id} |
`
	if err := os.WriteFile(findingPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write finding file: %v", err)
	}

	if _, err := AggregateTarget(tempDir); err != nil {
		t.Fatalf("AggregateTarget failed: %v", err)
	}

	for _, name := range []string{"SUMMARY.md", "metadata.json", "report.html"} {
		p := filepath.Join(tempDir, name)
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("%s was not generated: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm&0o022 != 0 {
			t.Errorf("%s is group/other-writable (mode %04o), want no write bits outside owner", name, perm)
		}
	}
}

func TestValidateCVSSScore(t *testing.T) {
	valid := map[string]string{
		"9.8": "9.8", "8": "8.0", " 7.5 ": "7.5", "10": "10.0",
		"10.0": "10.0", "0": "0.0", "9.8 (Critical)": "9.8",
		"": "", "N/A": "", "-": "",
	}
	for in, want := range valid {
		got, err := ValidateCVSSScore(in)
		if err != nil {
			t.Errorf("ValidateCVSSScore(%q) unexpected error: %v", in, err)
		} else if got != want {
			t.Errorf("ValidateCVSSScore(%q) = %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"10.1", "-1", "high", "9.8.1", "abc", "NaN"} {
		if _, err := ValidateCVSSScore(in); err == nil {
			t.Errorf("ValidateCVSSScore(%q) expected error, got nil", in)
		}
	}
}

func TestNormalizeCWE(t *testing.T) {
	valid := map[string]string{
		"79": "CWE-79", "CWE-79": "CWE-79", "cwe-89": "CWE-89",
		" CWE-639 ": "CWE-639", "007": "CWE-7", "": "", "N/A": "",
	}
	for in, want := range valid {
		got, err := NormalizeCWE(in)
		if err != nil {
			t.Errorf("NormalizeCWE(%q) unexpected error: %v", in, err)
		} else if got != want {
			t.Errorf("NormalizeCWE(%q) = %q, want %q", in, got, want)
		}
	}
	for _, in := range []string{"XSS", "CWE-", "CWE-abc", "79a", "--1"} {
		if _, err := NormalizeCWE(in); err == nil {
			t.Errorf("NormalizeCWE(%q) expected error, got nil", in)
		}
	}
}

func TestFindDuplicateFinding(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dedup_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	findingsDir := filepath.Join(tempDir, "findings")
	if err := os.MkdirAll(findingsDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `# IDOR in Profile Endpoint

- **Severity**: HIGH
- **Endpoint**: ` + "`GET /api/v1/users/{id}`" + `
`
	if err := os.WriteFile(filepath.Join(findingsDir, "high_idor.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Same finding despite case/whitespace differences.
	dup, err := FindDuplicateFinding(findingsDir, "  idor IN profile endpoint ", "get  /API/v1/users/{id}")
	if err != nil {
		t.Fatalf("FindDuplicateFinding: %v", err)
	}
	if dup == "" {
		t.Error("expected duplicate to be detected")
	}

	// Same title but different endpoint is a distinct finding.
	if dup, _ := FindDuplicateFinding(findingsDir, "IDOR in Profile Endpoint", "GET /api/v1/orders/{id}"); dup != "" {
		t.Errorf("false positive duplicate: %s", dup)
	}

	// Missing directory is not an error.
	if dup, err := FindDuplicateFinding(filepath.Join(tempDir, "nope"), "X", "Y"); err != nil || dup != "" {
		t.Errorf("expected empty result for missing dir, got %q, %v", dup, err)
	}
}
