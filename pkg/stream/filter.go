package stream

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"
	"strings"
)

var (
	ansiRegex     = regexp.MustCompile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
	uuidRegex     = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	staticSuffixes = []string{
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp",
		".woff", ".woff2", ".ttf", ".eot", ".otf",
		".css", ".mp4", ".mp3", ".webm", ".avi", ".mov",
	}
	criticalMarkers = []string{
		"[critical]", "[high]", "cve-", "rce", "sql injection",
		"sqli", "idor", "ssrf", "xxe", "auth bypass",
	}
	secretMarkers = []string{
		".env", ".git", "swagger", "openapi", "graphql",
		"id_rsa", "password", "secret_key", "bearer ", "token=", "jwt",
	}
)

type ScoredLine struct {
	Score int
	Text  string
}

type ProcessResult struct {
	TotalRaw       int
	UniqueScored   int
	ShownCount     int
	PreservedCount int
}

func CleanLine(line string) string {
	cleaned := ansiRegex.ReplaceAllString(line, "")
	return strings.TrimSpace(cleaned)
}

func CalculateEntropy(text string) float64 {
	if len(text) < 16 {
		return 0.0
	}
	var counts [256]int
	length := float64(len(text))
	for i := 0; i < len(text); i++ {
		counts[text[i]]++
	}
	var entropy float64
	for _, count := range counts {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

func isStaticAsset(lower string) bool {
	for _, ext := range staticSuffixes {
		if strings.HasSuffix(lower, ext) || strings.Contains(lower, ext+"?") || strings.Contains(lower, ext+"#") {
			return true
		}
	}
	return false
}

func ScoreLine(line string) int {
	lower := strings.ToLower(line)
	if isStaticAsset(lower) {
		return 0
	}

	score := 10

	for _, m := range criticalMarkers {
		if strings.Contains(lower, m) {
			score += 80
			break
		}
	}

	for _, m := range secretMarkers {
		if strings.Contains(lower, m) {
			score += 60
			break
		}
	}

	if strings.Contains(lower, "200 ok") || strings.Contains(lower, "[200]") {
		score += 25
		if strings.Contains(lower, "/api/") || strings.Contains(lower, "/v1/") || strings.Contains(lower, "/v2/") {
			score += 25
		}
	} else if strings.Contains(lower, "[401]") || strings.Contains(lower, "[403]") ||
		strings.Contains(lower, "401 unauthorized") || strings.Contains(lower, "403 forbidden") {
		score += 20
		if strings.Contains(lower, "/admin") || strings.Contains(lower, "/api/") || strings.Contains(lower, "/internal") {
			score += 25
		}
	} else if strings.Contains(lower, "[500]") || strings.Contains(lower, "500 internal server error") {
		score += 15
	}

	if strings.Contains(line, "?") && strings.Contains(line, "=") {
		score += 20
	}
	if uuidRegex.MatchString(line) {
		score += 20
	}

	if strings.Contains(lower, "key") || strings.Contains(lower, "secret") ||
		strings.Contains(lower, "tok") || strings.Contains(lower, "pass") {
		if CalculateEntropy(line) > 3.8 {
			score += 30
		}
	}

	return score
}

func ProcessStream(r io.Reader, stdout io.Writer, rawOut io.Writer, limit int) (ProcessResult, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	rawBuf := bufio.NewWriter(rawOut)
	defer rawBuf.Flush()

	var totalRaw int
	var scored []ScoredLine
	seen := make(map[string]struct{})

	for scanner.Scan() {
		cleaned := CleanLine(scanner.Text())
		if cleaned == "" {
			continue
		}
		totalRaw++
		rawBuf.WriteString(cleaned)
		rawBuf.WriteByte('\n')

		if _, exists := seen[cleaned]; !exists {
			seen[cleaned] = struct{}{}
			score := ScoreLine(cleaned)
			if score > 0 {
				scored = append(scored, ScoredLine{Score: score, Text: cleaned})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return ProcessResult{}, err
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	displayCount := limit
	if displayCount > len(scored) {
		displayCount = len(scored)
	}

	fmt.Fprintf(stdout, "📊 [Smart Filter] %d high-signal findings prioritized (from %d total raw lines).\n\n", displayCount, totalRaw)

	for i := 0; i < displayCount; i++ {
		fmt.Fprintln(stdout, scored[i].Text)
	}

	if len(scored) > displayCount {
		fmt.Fprintf(stdout, "\n... (+%d more filtered entries archived in raw log)\n", len(scored)-displayCount)
	}

	return ProcessResult{
		TotalRaw:       totalRaw,
		UniqueScored:   len(scored),
		ShownCount:     displayCount,
		PreservedCount: len(scored) - displayCount,
	}, nil
}
