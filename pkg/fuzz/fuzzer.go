package fuzz

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FuzzMatch represents an endpoint discovered through fuzzing.
type FuzzMatch struct {
	URL            string `json:"url"`
	Path           string `json:"path"`
	StatusCode     int    `json:"status_code"`
	ContentLength  int64  `json:"content_length"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	RedirectURL    string `json:"redirect_url,omitempty"`
}

// FuzzOptions configures fuzzing parameters.
type FuzzOptions struct {
	TargetURL    string
	WordlistPath string
	Extensions   []string
	StatusCodes  []int
	RateLimit    int
	Timeout      time.Duration
	Headers      map[string]string
	Cookies      string
	ToolsDir     string
	PreferFfuf   bool
	MaxResults   int
}

// FuzzResult encapsulates discovery findings and execution metadata.
type FuzzResult struct {
	TargetURL    string      `json:"target_url"`
	TotalProbed  int         `json:"total_probed"`
	TotalMatched int         `json:"total_matched"`
	Matches      []FuzzMatch `json:"matches"`
	EngineUsed   string      `json:"engine_used"`
	DurationMs   int64       `json:"duration_ms"`
}

// Built-in core wordlist fallback if no wordlist file exists on disk
var defaultCommonPaths = []string{
	"api", "api/v1", "api/v2", "admin", "administrator", "login", "auth",
	"swagger", "swagger-ui.html", "openapi.json", "docs", "v1", "v2",
	"dashboard", "portal", "config", "env", ".env", ".git", ".git/config",
	"graphql", "health", "metrics", "debug", "actuator", "status",
	"users", "user", "accounts", "profile", "orders", "invoices", "documents",
	"upload", "uploads", "files", "static", "assets", "backup", "dump",
}

// FindFfufBinary checks if ffuf executable is in toolsDir or system PATH.
func FindFfufBinary(toolsDir string) string {
	if toolsDir != "" {
		for _, name := range []string{"ffuf.exe", "ffuf"} {
			binPath := filepath.Join(toolsDir, "bin", name)
			if path, err := exec.LookPath(binPath); err == nil {
				return path
			}
		}
	}
	if path, err := exec.LookPath("ffuf"); err == nil {
		return path
	}
	return ""
}

// ResolveWordlist locates a valid wordlist file or returns a temporary wordlist path.
func ResolveWordlist(customPath, toolsDir string) (string, []string, error) {
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			lines, err := readLines(customPath)
			return customPath, lines, err
		}
	}

	if toolsDir != "" {
		candidates := []string{
			filepath.Join(toolsDir, "wordlists", "common.txt"),
			filepath.Join(toolsDir, "wordlists", "api-endpoints.txt"),
			filepath.Join(toolsDir, "wordlists", "raft-medium-directories.txt"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				lines, err := readLines(c)
				return c, lines, err
			}
		}
	}

	return "", defaultCommonPaths, nil
}

func readLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// FuzzEndpoints executes rate-controlled directory and endpoint fuzzing.
func FuzzEndpoints(ctx context.Context, opts FuzzOptions) (*FuzzResult, error) {
	if opts.TargetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}

	if !strings.Contains(opts.TargetURL, "://") {
		opts.TargetURL = "https://" + opts.TargetURL
	}

	baseURL, err := url.Parse(opts.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	if opts.RateLimit <= 0 || opts.RateLimit > 25 {
		opts.RateLimit = 10
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 30
	}

	if len(opts.StatusCodes) == 0 {
		opts.StatusCodes = []int{200, 204, 301, 302, 307, 401, 403, 405}
	}

	wordlistFile, wordlistLines, _ := ResolveWordlist(opts.WordlistPath, opts.ToolsDir)

	startTime := time.Now()
	var matches []FuzzMatch
	var totalProbed int
	var engineUsed string

	if opts.PreferFfuf && wordlistFile != "" {
		ffufBin := FindFfufBinary(opts.ToolsDir)
		if ffufBin != "" {
			var err error
			matches, totalProbed, err = runFfuf(ctx, ffufBin, baseURL.String(), wordlistFile, opts)
			if err == nil && len(matches) > 0 {
				engineUsed = "ffuf"
			}
		}
	}

	if engineUsed == "" {
		var err error
		matches, totalProbed, err = runNativeFuzzer(ctx, baseURL, wordlistLines, opts)
		if err != nil {
			return nil, fmt.Errorf("native fuzzing failed: %w", err)
		}
		engineUsed = "native_go"
	}

	// Sort matches: 200/204 first, then 301/302, then 401/403
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].StatusCode == matches[j].StatusCode {
			return matches[i].Path < matches[j].Path
		}
		return matches[i].StatusCode < matches[j].StatusCode
	})

	if len(matches) > opts.MaxResults {
		matches = matches[:opts.MaxResults]
	}

	duration := time.Since(startTime)

	return &FuzzResult{
		TargetURL:    opts.TargetURL,
		TotalProbed:  totalProbed,
		TotalMatched: len(matches),
		Matches:      matches,
		EngineUsed:   engineUsed,
		DurationMs:   duration.Milliseconds(),
	}, nil
}

type ffufJSON struct {
	Results []struct {
		Input    map[string]string `json:"input"`
		URL      string            `json:"url"`
		Status   int               `json:"status"`
		Length   int64             `json:"length"`
		Duration int64             `json:"duration"`
		Redirect string            `json:"redirectlocation"`
	} `json:"results"`
}

func runFfuf(ctx context.Context, ffufPath, targetURL, wordlistFile string, opts FuzzOptions) ([]FuzzMatch, int, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	statusStrings := make([]string, len(opts.StatusCodes))
	for i, sc := range opts.StatusCodes {
		statusStrings[i] = strconv.Itoa(sc)
	}

	fuzzURL := strings.TrimRight(targetURL, "/") + "/FUZZ"
	args := []string{
		"-u", fuzzURL,
		"-w", wordlistFile,
		"-mc", strings.Join(statusStrings, ","),
		"-rate", fmt.Sprintf("%d", opts.RateLimit),
		"-timeout", "5",
		"-s",
		"-json",
	}

	if opts.Cookies != "" {
		hasCookieHeader := false
		for k := range opts.Headers {
			if strings.EqualFold(k, "cookie") {
				hasCookieHeader = true
				break
			}
		}
		if !hasCookieHeader {
			args = append(args, "-H", fmt.Sprintf("Cookie: %s", opts.Cookies))
		}
	}

	for k, v := range opts.Headers {
		if strings.TrimSpace(k) != "" {
			args = append(args, "-H", fmt.Sprintf("%s: %s", strings.TrimSpace(k), strings.TrimSpace(v)))
		}
	}

	cmd := exec.CommandContext(cmdCtx, ffufPath, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil && stdout.Len() == 0 {
		return nil, 0, err
	}

	var parsed ffufJSON
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		// Try parsing newline delimited JSON
		scanner := bufio.NewScanner(&stdout)
		var matches []FuzzMatch
		for scanner.Scan() {
			line := scanner.Bytes()
			var single struct {
				URL      string `json:"url"`
				Input    map[string]string `json:"input"`
				Status   int    `json:"status"`
				Length   int64  `json:"length"`
				Redirect string `json:"redirectlocation"`
			}
			if err := json.Unmarshal(line, &single); err == nil && single.Status > 0 {
				path := single.Input["FUZZ"]
				if path == "" {
					path = single.URL
				}
				matches = append(matches, FuzzMatch{
					URL:           single.URL,
					Path:          path,
					StatusCode:    single.Status,
					ContentLength: single.Length,
					RedirectURL:   single.Redirect,
				})
			}
		}
		return matches, len(matches), nil
	}

	var matches []FuzzMatch
	for _, r := range parsed.Results {
		path := r.Input["FUZZ"]
		if path == "" {
			path = r.URL
		}
		matches = append(matches, FuzzMatch{
			URL:            r.URL,
			Path:           path,
			StatusCode:     r.Status,
			ContentLength:  r.Length,
			ResponseTimeMs: r.Duration / 1000000,
			RedirectURL:    r.Redirect,
		})
	}

	return matches, len(matches), nil
}

func runNativeFuzzer(ctx context.Context, baseURL *url.URL, words []string, opts FuzzOptions) ([]FuzzMatch, int, error) {
	statusSet := make(map[int]bool)
	for _, sc := range opts.StatusCodes {
		statusSet[sc] = true
	}

	// Prepare paths with extensions
	var candidatePaths []string
	for _, w := range words {
		w = strings.TrimLeft(w, "/")
		if w == "" {
			continue
		}
		candidatePaths = append(candidatePaths, w)
		for _, ext := range opts.Extensions {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			candidatePaths = append(candidatePaths, w+ext)
		}
	}

	client := &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	rateDelay := time.Duration(1000/opts.RateLimit) * time.Millisecond
	throttle := time.NewTicker(rateDelay)
	defer throttle.Stop()

	var matches []FuzzMatch
	var matchesMu sync.Mutex
	totalProbed := 0

	jobs := make(chan string, len(candidatePaths))
	for _, p := range candidatePaths {
		jobs <- p
	}
	close(jobs)

	numWorkers := 5
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				select {
				case <-ctx.Done():
					return
				case <-throttle.C:
				}

				targetEndpoint := strings.TrimRight(baseURL.String(), "/") + "/" + path
				req, err := http.NewRequestWithContext(ctx, "GET", targetEndpoint, nil)
				if err != nil {
					continue
				}

				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Cybermes/3.2 (API Fuzzer)")
				req.Header.Set("Accept", "*/*")

				if opts.Cookies != "" {
					req.Header.Set("Cookie", opts.Cookies)
				}
				for k, v := range opts.Headers {
					if strings.TrimSpace(k) != "" {
						req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
					}
				}

				start := time.Now()
				resp, err := client.Do(req)
				dur := time.Since(start)

				matchesMu.Lock()
				totalProbed++
				matchesMu.Unlock()

				if err != nil {
					continue
				}

				statusCode := resp.StatusCode
				contentLen := resp.ContentLength
				redirectLoc := resp.Header.Get("Location")
				_ = resp.Body.Close()

				if statusSet[statusCode] {
					matchesMu.Lock()
					matches = append(matches, FuzzMatch{
						URL:            targetEndpoint,
						Path:           path,
						StatusCode:     statusCode,
						ContentLength:  contentLen,
						ResponseTimeMs: dur.Milliseconds(),
						RedirectURL:    redirectLoc,
					})
					matchesMu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return matches, totalProbed, nil
}
