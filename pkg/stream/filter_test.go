package stream

import (
	"bytes"
	"strings"
	"testing"
)

func TestCleanLine(t *testing.T) {
	raw := "\x1b[31m[CRITICAL]\x1b[0m SQL injection at /api/users"
	expected := "[CRITICAL] SQL injection at /api/users"
	got := CleanLine(raw)
	if got != expected {
		t.Errorf("CleanLine() = %q, want %q", got, expected)
	}
}

func TestCalculateEntropy(t *testing.T) {
	lowEntropy := "AAAAAAAAAAAAAAAA"
	if e := CalculateEntropy(lowEntropy); e != 0.0 {
		t.Errorf("CalculateEntropy(%q) = %f, want 0.0", lowEntropy, e)
	}

	highEntropy := "8f9a2b4c1d6e3f5a7b9c0d1e2f3a4b5c"
	if e := CalculateEntropy(highEntropy); e < 3.0 {
		t.Errorf("CalculateEntropy(%q) = %f, want >= 3.0", highEntropy, e)
	}
}

func TestScoreLine(t *testing.T) {
	tests := []struct {
		line     string
		wantZero bool
		minScore int
	}{
		{"https://example.com/logo.png", true, 0},
		{"https://example.com/font.woff2", true, 0},
		{"[CRITICAL] CVE-2024-1234 Remote Code Execution", false, 90},
		{"https://example.com/.env", false, 70},
		{"[200] https://example.com/api/v1/users?id=123", false, 75},
		{"https://example.com/items?id=1", false, 30},
	}

	for _, tt := range tests {
		score := ScoreLine(tt.line)
		if tt.wantZero && score != 0 {
			t.Errorf("ScoreLine(%q) = %d, want 0", tt.line, score)
		}
		if !tt.wantZero && score < tt.minScore {
			t.Errorf("ScoreLine(%q) = %d, want >= %d", tt.line, score, tt.minScore)
		}
	}
}

func TestProcessStream(t *testing.T) {
	input := strings.Join([]string{
		"https://example.com/style.css",
		"https://example.com/api/v1/login",
		"[CRITICAL] RCE found on /exec",
		"https://example.com/about",
	}, "\n")

	var stdout bytes.Buffer
	var rawLog bytes.Buffer

	res, err := ProcessStream(strings.NewReader(input), &stdout, &rawLog, 2)
	if err != nil {
		t.Fatalf("ProcessStream failed: %v", err)
	}

	if res.TotalRaw != 4 {
		t.Errorf("res.TotalRaw = %d, want 4", res.TotalRaw)
	}
	if res.ShownCount != 2 {
		t.Errorf("res.ShownCount = %d, want 2", res.ShownCount)
	}

	rawContent := rawLog.String()
	if !strings.Contains(rawContent, "https://example.com/style.css") {
		t.Errorf("rawLog missing static asset")
	}

	outContent := stdout.String()
	if strings.Contains(outContent, "style.css") {
		t.Errorf("stdout should filter out static asset")
	}
	if !strings.Contains(outContent, "[CRITICAL] RCE found on /exec") {
		t.Errorf("stdout should prioritize critical finding")
	}
}
