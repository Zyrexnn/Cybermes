package report

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	wsCollapseRe = regexp.MustCompile(`\s+`)
	cweDigitsRe  = regexp.MustCompile(`^(\d{1,5})$`)
	cweFullRe    = regexp.MustCompile(`(?i)^cwe-(\d{1,5})$`)
)

// normalizeKeyPart lowercases, trims, and collapses inner whitespace so that
// semantically identical titles/endpoints map to the same dedup key despite
// trivial agent formatting differences.
func normalizeKeyPart(s string) string {
	return wsCollapseRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), " ")
}

// FindingKey builds a stable deduplication identity for a finding from its
// title and endpoint.
func FindingKey(title, endpoint string) string {
	return normalizeKeyPart(title) + "|" + normalizeKeyPart(endpoint)
}

// FindDuplicateFinding scans findingsDir for an already-recorded finding with
// the same title+endpoint key. It returns the existing file path, or "" when
// no duplicate exists.
func FindDuplicateFinding(findingsDir, title, endpoint string) (string, error) {
	want := FindingKey(title, endpoint)
	entries, err := os.ReadDir(findingsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.ToLower(e.Name())
		if !strings.HasSuffix(name, ".md") || name == "summary.md" || name == "readme.md" {
			continue
		}
		meta, err := ParseFindingFile(filepath.Join(findingsDir, e.Name()))
		if err != nil {
			continue
		}
		if FindingKey(meta.Title, meta.Endpoint) == want {
			return filepath.Join(findingsDir, e.Name()), nil
		}
	}
	return "", nil
}

// ValidateCVSSScore checks a CVSS v3.x base score string ("0.0"–"10.0") and
// returns it in canonical one-decimal form. Empty input yields "" (omitted).
func ValidateCVSSScore(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" || strings.EqualFold(s, "n/a") || s == "-" {
		return "", nil
	}
	// Accept an optional trailing "(Label)" suffix as produced by some templates.
	if i := strings.Index(s, "("); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 10 {
		return "", fmt.Errorf("invalid CVSS score %q: must be a number between 0.0 and 10.0", raw)
	}
	return strconv.FormatFloat(v, 'f', 1, 64), nil
}

// NormalizeCWE canonicalizes a CWE reference ("79", "cwe-79", "CWE-79") to
// "CWE-79". Empty input yields "" (omitted).
func NormalizeCWE(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" || strings.EqualFold(s, "n/a") || s == "-" {
		return "", nil
	}
	if m := cweFullRe.FindStringSubmatch(s); len(m) == 2 {
		return "CWE-" + stripLeadingZeros(m[1]), nil
	}
	if m := cweDigitsRe.FindStringSubmatch(s); len(m) == 2 {
		return "CWE-" + stripLeadingZeros(m[1]), nil
	}
	return "", fmt.Errorf("invalid CWE reference %q: expected 'CWE-<digits>' (e.g. 'CWE-79')", raw)
}

func stripLeadingZeros(digits string) string {
	stripped := strings.TrimLeft(digits, "0")
	if stripped == "" {
		return "0"
	}
	return stripped
}
