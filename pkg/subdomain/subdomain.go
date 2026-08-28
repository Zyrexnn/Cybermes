package subdomain

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SubdomainOptions configures subdomain discovery execution.
type SubdomainOptions struct {
	Domain          string
	Timeout         time.Duration
	ToolsDir        string
	PreferSubfinder bool
}

// SubdomainResult encapsulates discovered subdomains and metadata.
type SubdomainResult struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	TotalFound int      `json:"total_found"`
	EngineUsed string   `json:"engine_used"`
	DurationMs int64    `json:"duration_ms"`
}

// FindSubfinderBinary checks if subfinder executable is in toolsDir or system PATH.
func FindSubfinderBinary(toolsDir string) string {
	if toolsDir != "" {
		for _, name := range []string{"subfinder.exe", "subfinder"} {
			binPath := filepath.Join(toolsDir, "bin", name)
			if path, err := exec.LookPath(binPath); err == nil {
				return path
			}
		}
	}
	if path, err := exec.LookPath("subfinder"); err == nil {
		return path
	}
	return ""
}

// CleanDomain normalizes target inputs into a pure base domain.
func CleanDomain(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "://") {
		if u, err := url.Parse(raw); err == nil {
			raw = u.Hostname()
		}
	}
	raw = strings.TrimPrefix(raw, "*.")
	raw = strings.TrimPrefix(raw, ".")
	if idx := strings.Index(raw, "/"); idx != -1 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, ":"); idx != -1 {
		raw = raw[:idx]
	}
	return strings.ToLower(raw)
}

// DiscoverSubdomains executes passive discovery using subfinder if available, or native crt.sh CT logs.
func DiscoverSubdomains(ctx context.Context, opts SubdomainOptions) (*SubdomainResult, error) {
	domain := CleanDomain(opts.Domain)
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}

	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	startTime := time.Now()
	var subs []string
	var engineUsed string

	if opts.PreferSubfinder {
		subfinderBin := FindSubfinderBinary(opts.ToolsDir)
		if subfinderBin != "" {
			var err error
			subs, err = runSubfinder(ctx, subfinderBin, domain, opts.Timeout)
			if err == nil && len(subs) > 0 {
				engineUsed = "subfinder"
			}
		}
	}

	if engineUsed == "" {
		var err error
		subs, err = runCrtSh(ctx, domain, opts.Timeout)
		if err != nil && len(subs) == 0 {
			return &SubdomainResult{
				Domain:     domain,
				Subdomains: []string{},
				TotalFound: 0,
				EngineUsed: "crt_sh_offline",
				DurationMs: time.Since(startTime).Milliseconds(),
			}, nil
		}
		engineUsed = "crt_sh_native"
	}

	// Deduplicate & normalize
	uniqueMap := make(map[string]bool)
	var cleanedSubs []string

	for _, s := range subs {
		s = CleanDomain(s)
		if s == "" || !strings.HasSuffix(s, domain) {
			continue
		}
		if !uniqueMap[s] {
			uniqueMap[s] = true
			cleanedSubs = append(cleanedSubs, s)
		}
	}

	sort.Strings(cleanedSubs)
	duration := time.Since(startTime)

	return &SubdomainResult{
		Domain:     domain,
		Subdomains: cleanedSubs,
		TotalFound: len(cleanedSubs),
		EngineUsed: engineUsed,
		DurationMs: duration.Milliseconds(),
	}, nil
}

func runSubfinder(ctx context.Context, subfinderPath, domain string, timeout time.Duration) ([]string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, subfinderPath, "-d", domain, "-silent")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil && stdout.Len() == 0 {
		return nil, fmt.Errorf("subfinder failed: %w (stderr: %s)", err, stderr.String())
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

type crtShEntry struct {
	NameValue string `json:"name_value"`
}

func runCrtSh(ctx context.Context, domain string, timeout time.Duration) ([]string, error) {
	apiURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", url.QueryEscape(domain))

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Cybermes/3.2 (Subdomain Recon)")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("crt.sh returned HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}

	var entries []crtShEntry
	if err := json.Unmarshal(bodyBytes, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse crt.sh JSON: %w", err)
	}

	var results []string
	for _, entry := range entries {
		names := strings.Split(entry.NameValue, "\n")
		for _, n := range names {
			n = strings.TrimSpace(n)
			if n != "" {
				results = append(results, n)
			}
		}
	}

	return results, nil
}
