package probe

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
	"regexp"
	"strings"
	"time"
)

var titleRegex = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)

// TLSInfo contains extracted SSL/TLS certificate details.
type TLSInfo struct {
	Version     string   `json:"version,omitempty"`
	CipherSuite string   `json:"cipher_suite,omitempty"`
	Issuer      string   `json:"issuer,omitempty"`
	Subject     string   `json:"subject,omitempty"`
	DNSNames    []string `json:"dns_names,omitempty"`
	ExpiresAt   string   `json:"expires_at,omitempty"`
}

// ProbeResult encapsulates all HTTP inspection and tech detection findings.
type ProbeResult struct {
	URL            string            `json:"url"`
	Scheme         string            `json:"scheme"`
	Host           string            `json:"host"`
	Port           string            `json:"port,omitempty"`
	StatusCode     int               `json:"status_code"`
	StatusText     string            `json:"status_text"`
	Title          string            `json:"title,omitempty"`
	WebServer      string            `json:"web_server,omitempty"`
	ContentType    string            `json:"content_type,omitempty"`
	ContentLength  int64             `json:"content_length,omitempty"`
	ResponseTimeMs int64             `json:"response_time_ms"`
	RedirectURL    string            `json:"redirect_url,omitempty"`
	Technologies   []string          `json:"technologies"`
	TLSInfo        *TLSInfo          `json:"tls_info,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	EngineUsed     string            `json:"engine_used"`
}

// ProbeOptions defines configuration parameters for probing.
type ProbeOptions struct {
	TargetURL       string
	Timeout         time.Duration
	FollowRedirects bool
	ToolsDir        string
	UserAgent       string
	PreferHttpx     bool
	Headers         map[string]string
	Cookies         string
}

// ExtractTitle finds the <title> tag text inside an HTML body.
func ExtractTitle(htmlBody string) string {
	matches := titleRegex.FindStringSubmatch(htmlBody)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// ProbeTarget executes an active HTTP probe, attempting httpx first if requested/available, or native Go.
func ProbeTarget(ctx context.Context, opts ProbeOptions) (*ProbeResult, error) {
	if opts.TargetURL == "" {
		return nil, fmt.Errorf("target URL cannot be empty")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Cybermes/2.1 (Security Assessment)"
	}

	if opts.PreferHttpx {
		res, err := ProbeHttpx(ctx, opts)
		if err == nil && res != nil {
			return res, nil
		}
		// Fallback cleanly to native Go
	}

	return ProbeNative(ctx, opts)
}

// ProbeNative performs pure Go net/http probing with TLS analysis and technology detection.
func ProbeNative(ctx context.Context, opts ProbeOptions) (*ProbeResult, error) {
	rawTarget := strings.TrimSpace(opts.TargetURL)
	if !strings.Contains(rawTarget, "://") {
		rawTarget = "http://" + rawTarget
	}

	parsedURL, err := url.Parse(rawTarget)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var tlsState *tls.ConnectionState
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Assessment mode allows self-signed/internal certs
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
	}

	var redirectURL string
	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectURL = req.URL.String()
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawTarget, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept", "*/*")

	if opts.Cookies != "" {
		req.Header.Set("Cookie", opts.Cookies)
	}
	for k, v := range opts.Headers {
		if strings.TrimSpace(k) != "" {
			req.Header.Set(strings.TrimSpace(k), strings.TrimSpace(v))
		}
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.TLS != nil {
		tlsState = resp.TLS
	}

	// Read first 1MB of body for fingerprinting & title
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	bodyStr := string(bodyBytes)

	// Extract Headers
	headerMap := make(map[string]string)
	for k, v := range resp.Header {
		headerMap[k] = strings.Join(v, ", ")
	}

	title := ExtractTitle(bodyStr)
	serverHeader := resp.Header.Get("Server")
	contentType := resp.Header.Get("Content-Type")

	// Detect Technologies
	techs := DetectTechnologies(resp.Header, resp.Cookies(), bodyStr)

	// TLS details
	var tlsInfo *TLSInfo
	if tlsState != nil {
		tlsInfo = &TLSInfo{
			Version:     tlsVersionToString(tlsState.Version),
			CipherSuite: tls.CipherSuiteName(tlsState.CipherSuite),
		}
		if len(tlsState.PeerCertificates) > 0 {
			cert := tlsState.PeerCertificates[0]
			tlsInfo.Issuer = cert.Issuer.CommonName
			tlsInfo.Subject = cert.Subject.CommonName
			tlsInfo.DNSNames = cert.DNSNames
			tlsInfo.ExpiresAt = cert.NotAfter.Format(time.RFC3339)
		}
	}

	return &ProbeResult{
		URL:            rawTarget,
		Scheme:         parsedURL.Scheme,
		Host:           parsedURL.Hostname(),
		Port:           parsedURL.Port(),
		StatusCode:     resp.StatusCode,
		StatusText:     http.StatusText(resp.StatusCode),
		Title:          title,
		WebServer:      serverHeader,
		ContentType:    contentType,
		ContentLength:  resp.ContentLength,
		ResponseTimeMs: duration.Milliseconds(),
		RedirectURL:    redirectURL,
		Technologies:   techs,
		TLSInfo:        tlsInfo,
		Headers:        headerMap,
		EngineUsed:     "native_go",
	}, nil
}

// FindHttpxBinary checks for httpx executable in toolsDir or system PATH.
func FindHttpxBinary(toolsDir string) string {
	if toolsDir != "" {
		for _, name := range []string{"httpx.exe", "httpx"} {
			binPath := filepath.Join(toolsDir, "bin", name)
			if path, err := exec.LookPath(binPath); err == nil {
				return path
			}
		}
	}
	if path, err := exec.LookPath("httpx"); err == nil {
		return path
	}
	return ""
}

// ProbeHttpx executes external httpx binary if present and parses JSON output.
func ProbeHttpx(ctx context.Context, opts ProbeOptions) (*ProbeResult, error) {
	httpxPath := FindHttpxBinary(opts.ToolsDir)
	if httpxPath == "" {
		return nil, fmt.Errorf("httpx binary not found")
	}

	timeoutSec := int(opts.Timeout.Seconds())
	if timeoutSec <= 0 {
		timeoutSec = 10
	}

	args := []string{
		"-u", opts.TargetURL,
		"-silent",
		"-status-code",
		"-title",
		"-tech-detect",
		"-json",
		"-timeout", fmt.Sprintf("%d", timeoutSec),
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

	cmd := exec.CommandContext(ctx, httpxPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	startTime := time.Now()
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("httpx execution failed: %w (stderr: %s)", err, stderr.String())
	}
	duration := time.Since(startTime)

	var httpxOut struct {
		URL         string   `json:"url"`
		StatusCode  int      `json:"status_code"`
		Title       string   `json:"title"`
		WebServer   string   `json:"webserver"`
		Tech        []string `json:"tech"`
		Host        string   `json:"host"`
		Port        string   `json:"port"`
		Scheme      string   `json:"scheme"`
		ContentType string   `json:"content_type"`
		TLS         *struct {
			Version string `json:"version"`
			Subject string `json:"subject_dn"`
			Issuer  string `json:"issuer_dn"`
		} `json:"tls"`
	}

	parsed := false
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) > 0 && line[0] == '{' {
			if err := json.Unmarshal(line, &httpxOut); err == nil && httpxOut.URL != "" {
				parsed = true
				break
			}
		}
	}

	if !parsed {
		return nil, fmt.Errorf("failed to parse valid JSON from httpx output: %s", stdout.String())
	}

	var tlsInfo *TLSInfo
	if httpxOut.TLS != nil {
		tlsInfo = &TLSInfo{
			Version: httpxOut.TLS.Version,
			Subject: httpxOut.TLS.Subject,
			Issuer:  httpxOut.TLS.Issuer,
		}
	}

	return &ProbeResult{
		URL:            httpxOut.URL,
		Scheme:         httpxOut.Scheme,
		Host:           httpxOut.Host,
		Port:           httpxOut.Port,
		StatusCode:     httpxOut.StatusCode,
		StatusText:     http.StatusText(httpxOut.StatusCode),
		Title:          httpxOut.Title,
		WebServer:      httpxOut.WebServer,
		ContentType:    httpxOut.ContentType,
		ResponseTimeMs: duration.Milliseconds(),
		Technologies:   httpxOut.Tech,
		TLSInfo:        tlsInfo,
		EngineUsed:     "httpx",
	}, nil
}

func tlsVersionToString(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("TLS 0x%04x", v)
	}
}
