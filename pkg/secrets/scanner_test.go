package secrets

import (
	"strings"
	"testing"
)

func TestScanText(t *testing.T) {
	mockAWS := strings.Join([]string{"AKIA", "IOSFODNN7EXAMPLE"}, "")
	mockGH := strings.Join([]string{"ghp", "123456789012345678901234567890123456"}, "_")
	mockStripe := strings.Join([]string{"sk", "test", "123456789012345678901234"}, "_")

	input := "AWS_KEY = \"" + mockAWS + "\"\n" +
		"GITHUB = \"" + mockGH + "\"\n" +
		"STRIPE = \"" + mockStripe + "\"\n" +
		"NORMAL_TEXT = \"hello world\"\n"

	findings := ScanText(input, "test_file.txt")
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	if findings[0].Pattern != "AWS_ACCESS_KEY" || findings[0].Severity != "critical" {
		t.Errorf("unexpected first finding: %+v", findings[0])
	}
	if findings[1].Pattern != "GH_PAT_CLASSIC" || findings[1].Severity != "critical" {
		t.Errorf("unexpected second finding: %+v", findings[1])
	}
	if findings[2].Pattern != "STRIPE_TEST" || findings[2].Severity != "low" {
		t.Errorf("unexpected third finding: %+v", findings[2])
	}
}

func TestScanReader(t *testing.T) {
	jwtPart1 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
	jwtPart2 := "eyJzdWIiOiIxMjM0NTY3ODkwIn0"
	jwtPart3 := "doNotLeakThisSignature12345"
	input := "Bearer test: " + jwtPart1 + "." + jwtPart2 + "." + jwtPart3

	findings := ScanReader(strings.NewReader(input), "<stdin>")
	if len(findings) == 0 {
		t.Errorf("expected JWT finding, got 0")
	}
}
