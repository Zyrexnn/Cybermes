package crawl

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"cybermes/pkg/stream"
)

var (
	hrefRegex   = regexp.MustCompile(`(?i)(?:href|src|action)=["']([^"'#\s>]+)["']`)
	apiEndpoint = regexp.MustCompile(`(?i)["'](/(?:api|v[0-9]|rest|graphql|admin|auth|oauth|users|invoices|orders|documents|internal)[^"'#\s]*)["']`)
)

// CrawlOptions configures the crawler parameters.
type CrawlOptions struct {
	TargetURL    string
	Depth        int
	MaxEndpoints int
	Timeout      time.Duration
	ToolsDir     string
	OutputDir    string
	PreferKatana bool
	UserAgent    string
}

// CrawlResult holds the summarized and filtered crawl discovery data.
type CrawlResult struct {
	TargetURL           string              `json:"target_url"`
	TotalEndpointsFound int                 `json:"total_endpoints_found"`
	TopEndpoints        []stream.ScoredLine `json:"top_endpoints"`
	SavedFilePath       string              `json:"saved_file_path,omitempty"`
	EngineUsed          string              `json:"engine_used"`
	DurationMs          int64               `json:"duration_ms"`
}

// FindKatanaBinary checks if katana binary is present in toolsDir or system PATH.
func FindKatanaBinary(toolsDir string) string {
	if toolsDir != "" {
		for _, name := range []string{"katana.exe", "katana"} {
			binPath := filepath.Join(toolsDir, "bin", name)
			if path, err := exec.LookPath(binPath); err == nil {
				return path
			}
		}
	}
	if path, err := exec.LookPath("katana"); err == nil {
		return path
	}
	return ""
}

// CrawlTarget executes endpoint crawling using Katana if available, or native Go crawler.
func CrawlTarget(ctx context.Context, opts CrawlOptions) (*CrawlResult, error) {
	if opts.TargetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}
	if opts.Depth <= 0 {
		opts.Depth = 2
	}
	if opts.MaxEndpoints <= 0 {
		opts.MaxEndpoints = 25
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Cybermes/2.1 (Recon Crawler)"
	}

	startTime := time.Now()

	var rawEndpoints []string
	var engineUsed string

	if opts.PreferKatana {
		katanaPath := FindKatanaBinary(opts.ToolsDir)
		if katanaPath != "" {
			var err error
			rawEndpoints, err = runKatana(ctx, katanaPath, opts)
			if err == nil && len(rawEndpoints) > 0 {
				engineUsed = "katana"
			}
		}
	}

	if engineUsed == "" {
		var err error
		rawEndpoints, err = runNativeCrawler(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("native crawler failed: %w", err)
		}
		engineUsed = "native_go"
	}

	duration := time.Since(startTime)

	// Deduplicate & Score using pkg/stream (Smart Pipe)
	uniqueMap := make(map[string]bool)
	scoredList := make([]stream.ScoredLine, 0, len(rawEndpoints))

	for _, ep := range rawEndpoints {
		ep = strings.TrimSpace(ep)
		if ep == "" || uniqueMap[ep] {
			continue
		}
		uniqueMap[ep] = true

		score := stream.ScoreLine(ep)
		scoredList = append(scoredList, stream.ScoredLine{
			Score: score,
			Text:  ep,
		})
	}

	// Sort high-signal endpoints first
	sort.Slice(scoredList, func(i, j int) bool {
		if scoredList[i].Score == scoredList[j].Score {
			return len(scoredList[i].Text) < len(scoredList[j].Text)
		}
		return scoredList[i].Score > scoredList[j].Score
	})

	// Save complete findings to disk (Preserve full raw dump)
	var savedPath string
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err == nil {
			savedPath = filepath.Join(opts.OutputDir, "katana.txt")
			var buf bytes.Buffer
			for _, item := range scoredList {
				buf.WriteString(fmt.Sprintf("%s\n", item.Text))
			}
			_ = os.WriteFile(savedPath, buf.Bytes(), 0644)
		}
	}

	// Slice top N for LLM token budget
	topCount := opts.MaxEndpoints
	if len(scoredList) < topCount {
		topCount = len(scoredList)
	}
	topEndpoints := scoredList[:topCount]

	return &CrawlResult{
		TargetURL:           opts.TargetURL,
		TotalEndpointsFound: len(scoredList),
		TopEndpoints:        topEndpoints,
		SavedFilePath:       savedPath,
		EngineUsed:          engineUsed,
		DurationMs:          duration.Milliseconds(),
	}, nil
}

func runKatana(ctx context.Context, katanaPath string, opts CrawlOptions) ([]string, error) {
	depthStr := fmt.Sprintf("%d", opts.Depth)
	timeoutSec := int(opts.Timeout.Seconds())
	if timeoutSec <= 0 {
		timeoutSec = 30
	}

	args := []string{
		"-u", opts.TargetURL,
		"-d", depthStr,
		"-jc",
		"-silent",
		"-ct", fmt.Sprintf("%ds", timeoutSec),
	}

	cmd := exec.CommandContext(ctx, katanaPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("katana command failed: %w", err)
	}

	var results []string
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			results = append(results, line)
		}
	}

	return results, nil
}

func runNativeCrawler(ctx context.Context, opts CrawlOptions) ([]string, error) {
	baseURL, err := url.Parse(opts.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 5 * time.Second,
	}

	visited := make(map[string]bool)
	var visitedMu sync.Mutex

	endpoints := make(map[string]bool)
	var endpointsMu sync.Mutex

	queue := []string{opts.TargetURL}
	maxPages := 30

	crawlLoop:
	for depth := 0; depth < opts.Depth && len(queue) > 0; depth++ {
		var nextQueue []string
		for _, curURL := range queue {
			if ctx.Err() != nil {
				break crawlLoop
			}

			visitedMu.Lock()
			if visited[curURL] || len(visited) >= maxPages {
				visitedMu.Unlock()
				continue
			}
			visited[curURL] = true
			visitedMu.Unlock()

			req, err := http.NewRequestWithContext(ctx, "GET", curURL, nil)
			if err != nil {
				continue
			}
			req.Header.Set("User-Agent", opts.UserAgent)

			resp, err := client.Do(req)
			if err != nil {
				continue
			}

			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
			_ = resp.Body.Close()
			bodyStr := string(bodyBytes)

			// 1. Extract HTML links & script tags
			matches := hrefRegex.FindAllStringSubmatch(bodyStr, -1)
			for _, m := range matches {
				if len(m) > 1 {
					extracted := strings.TrimSpace(m[1])
					resolved := resolveURL(baseURL, extracted)
					if resolved != "" {
						endpointsMu.Lock()
						endpoints[resolved] = true
						endpointsMu.Unlock()

						if isSameHost(baseURL, resolved) {
							nextQueue = append(nextQueue, resolved)
						}
					}
				}
			}

			// 2. Extract API patterns from JS / HTML content
			apiMatches := apiEndpoint.FindAllStringSubmatch(bodyStr, -1)
			for _, am := range apiMatches {
				if len(am) > 1 {
					extracted := strings.TrimSpace(am[1])
					resolved := resolveURL(baseURL, extracted)
					if resolved != "" {
						endpointsMu.Lock()
						endpoints[resolved] = true
						endpointsMu.Unlock()
					}
				}
			}
		}
		queue = nextQueue
	}

	res := make([]string, 0, len(endpoints))
	for ep := range endpoints {
		res = append(res, ep)
	}

	return res, nil
}

func resolveURL(base *url.URL, ref string) string {
	if ref == "" || strings.HasPrefix(ref, "javascript:") || strings.HasPrefix(ref, "mailto:") || strings.HasPrefix(ref, "data:") {
		return ""
	}

	if strings.Contains(ref, "://") {
		return ref
	}

	u, err := url.Parse(ref)
	if err != nil {
		return ""
	}

	return base.ResolveReference(u).String()
}

func isSameHost(base *url.URL, rawTarget string) bool {
	u, err := url.Parse(rawTarget)
	if err != nil {
		return false
	}
	return strings.EqualFold(base.Hostname(), u.Hostname())
}
