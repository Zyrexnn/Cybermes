package nmap

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Common TCP ports in approximate nmap frequency order, used by the native
// Go fallback engine for "top-N" scans without the nmap binary.
var commonPorts = []int{
	80, 443, 22, 21, 23, 25, 53, 110, 111, 135, 139, 143, 445,
	993, 995, 1723, 3306, 3389, 5900, 8080, 8443, 8000, 8888,
	1352, 1433, 1521, 2049, 2121, 3307, 5432, 6379, 27017,
	137, 138, 161, 162, 389, 636, 873, 1099, 1434, 2383,
	3128, 5060, 5061, 6000, 6666, 8008, 8081, 8444, 8889, 9000, 9090,
}

const (
	maxPortsPerScan = 1000
	nativeDialCap   = 64
)

var (
	listPortsRe = regexp.MustCompile(`^[\d,\s\-]+$`)
	topPortsRe  = regexp.MustCompile(`(?i)^top-(\d+)$`)
	nmapLineRe  = regexp.MustCompile(`^(\d+)/(tcp|udp)\s+(\S+)\s+(\S+)(?:\s+(.*))?$`)
	hostOKRe    = regexp.MustCompile(`^[A-Za-z0-9_.\-]+$`)
)

type PortResult struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
	Service  string `json:"service,omitempty"`
	Version  string `json:"version,omitempty"`
	Banner   string `json:"banner,omitempty"`
}

type ScanResult struct {
	Target        string       `json:"target"`
	Host          string       `json:"host"`
	PortsScanned  int          `json:"ports_scanned"`
	OpenPorts     []PortResult `json:"open_ports"`
	EngineUsed    string       `json:"engine_used"`
	DurationMs    int64        `json:"duration_ms"`
	SavedFilePath string       `json:"raw_dump_file,omitempty"`
}

type NmapOptions struct {
	Target     string
	Ports      string // "top-100" (default), "80,443", "1-1024"
	RateLimit  int    // max concurrent probes / nmap --max-rate
	Timeout    time.Duration
	ToolsDir   string
	OutputDir  string // recon dump dir, "" to skip
	PreferNmap bool
	// SkipServiceDetection disables -sV (port discovery only, fastest).
	// Version detection runs at --version-intensity 2: full-intensity -sV
	// can stall for tens of seconds on silent ports, so it is never used.
	SkipServiceDetection bool
}

// ExtractHost resolves a user target (host, IP, host:port, or URL) into a
// bare scan host plus an explicitly requested port (0 when absent).
func ExtractHost(raw string) (string, int, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", 0, fmt.Errorf("target is empty")
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return "", 0, fmt.Errorf("invalid target %q: contains whitespace", raw)
	}
	if strings.Contains(s, "..") {
		return "", 0, fmt.Errorf("invalid target %q", raw)
	}

	host := s
	port := 0
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil || u.Hostname() == "" {
			return "", 0, fmt.Errorf("invalid target URL %q", raw)
		}
		host = u.Hostname()
		if p := u.Port(); p != "" {
			port, _ = strconv.Atoi(p)
		}
	} else {
		host = strings.SplitN(s, "/", 2)[0]
		if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil {
			host = strings.Trim(host, "[]")
		} else if h, p, err := net.SplitHostPort(host); err == nil {
			host = h
			port, _ = strconv.Atoi(p)
		}
	}

	host = strings.Trim(host, "[]")
	if host == "" {
		return "", 0, fmt.Errorf("could not resolve scan host from %q", raw)
	}
	if ip := net.ParseIP(host); ip == nil && !hostOKRe.MatchString(host) {
		return "", 0, fmt.Errorf("invalid scan host %q", host)
	}
	if port < 0 || port > 65535 {
		return "", 0, fmt.Errorf("invalid port in target %q", raw)
	}
	return host, port, nil
}

// ParsePorts validates a port specification and expands it to a sorted,
// deduplicated port list. Accepted forms: "top-N", "80,443", "1-1024".
// Anything else (notably flag-shaped input like "-oN") is rejected so the
// spec can never smuggle extra nmap CLI flags.
func ParsePorts(spec string, explicitPort int) (ports []int, useTop bool, topN int, err error) {
	s := strings.ToLower(strings.TrimSpace(spec))
	if s == "" || s == "top-100" {
		if explicitPort > 0 {
			return []int{explicitPort}, false, 0, nil
		}
		s = "top-100"
	}
	if m := topPortsRe.FindStringSubmatch(s); len(m) == 2 {
		n, _ := strconv.Atoi(m[1])
		if n < 1 || n > maxPortsPerScan {
			return nil, false, 0, fmt.Errorf("top-N out of range 1-%d: %q", maxPortsPerScan, spec)
		}
		// The nmap binary receives the full requested N via --top-ports.
		// The native fallback engine covers the embedded commonPorts slice.
		capped := n
		if capped > len(commonPorts) {
			capped = len(commonPorts)
		}
		out := make([]int, capped)
		copy(out, commonPorts[:capped])
		return out, true, n, nil
	}
	if !listPortsRe.MatchString(s) {
		return nil, false, 0, fmt.Errorf("invalid ports specification %q: use 'top-N', '80,443' or '1-1024'", spec)
	}
	seen := map[int]struct{}{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			lo, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			hi, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err1 != nil || err2 != nil || lo < 1 || hi > 65535 || lo > hi {
				return nil, false, 0, fmt.Errorf("invalid port range %q", part)
			}
			for p := lo; p <= hi; p++ {
				seen[p] = struct{}{}
			}
		} else {
			p, convErr := strconv.Atoi(part)
			if convErr != nil || p < 1 || p > 65535 {
				return nil, false, 0, fmt.Errorf("invalid port %q", part)
			}
			seen[p] = struct{}{}
		}
		if len(seen) > maxPortsPerScan {
			return nil, false, 0, fmt.Errorf("port count exceeds limit of %d (narrow the range)", maxPortsPerScan)
		}
	}
	if len(seen) == 0 {
		return nil, false, 0, fmt.Errorf("no ports parsed from %q", spec)
	}
	ports = make([]int, 0, len(seen))
	for p := range seen {
		ports = append(ports, p)
	}
	sort.Ints(ports)
	return ports, false, 0, nil
}

func findNmapBinary(toolsDir string) string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	if toolsDir != "" {
		for _, name := range []string{"nmap" + ext, "nmap"} {
			if candidate := filepath.Join(toolsDir, "bin", name); isExecutable(candidate) {
				return candidate
			}
		}
	}
	if p, err := exec.LookPath("nmap" + ext); err == nil {
		return p
	}
	if p, err := exec.LookPath("nmap"); err == nil {
		return p
	}
	return ""
}

func isExecutable(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return fi.Mode().Perm()&0111 != 0
}

func clampRate(rate int) int {
	if rate < 1 {
		return 1
	}
	if rate > 100 {
		return 100
	}
	return rate
}

// ScanTarget runs a bounded, non-destructive TCP port scan: external nmap
// binary when available (and preferred), otherwise a native Go connect scan.
func ScanTarget(ctx context.Context, opts NmapOptions) (*ScanResult, error) {
	host, explicitPort, err := ExtractHost(opts.Target)
	if err != nil {
		return nil, err
	}
	ports, useTop, topN, err := ParsePorts(opts.Ports, explicitPort)
	if err != nil {
		return nil, err
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 60 * time.Second
	}
	rate := clampRate(opts.RateLimit)

	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	start := time.Now()

	res := &ScanResult{Target: opts.Target, Host: host, PortsScanned: len(ports)}
	if opts.PreferNmap {
		if bin := findNmapBinary(opts.ToolsDir); bin != "" {
			out, runErr := runNmapBinary(ctx, bin, host, useTop, topN, ports, rate, !opts.SkipServiceDetection)
			res.EngineUsed = "nmap"
			res.DurationMs = time.Since(start).Milliseconds()
			if runErr != nil {
				return res, runErr
			}
			res.OpenPorts = parseNmapNormal(out)
			saveDump(opts.OutputDir, out, res)
			return res, nil
		}
	}

	open := nativeConnectScan(ctx, host, ports, rate)
	res.EngineUsed = "native-go"
	res.OpenPorts = open
	res.DurationMs = time.Since(start).Milliseconds()
	saveDump(opts.OutputDir, renderNativeDump(res), res)
	return res, ctx.Err()
}

func runNmapBinary(ctx context.Context, bin, host string, useTop bool, topN int, ports []int, rate int, lightVersion bool) ([]byte, error) {
	args := []string{"-sT", "--open", "-T3", "--max-rate", strconv.Itoa(rate)}
	if lightVersion {
		args = append(args, "-sV", "--version-intensity", "2")
	}
	args = append(args, "-oN", "-")
	if useTop {
		args = append(args, "--top-ports", strconv.Itoa(topN))
	} else {
		strs := make([]string, len(ports))
		for i, p := range ports {
			strs[i] = strconv.Itoa(p)
		}
		args = append(args, "-p", strings.Join(strs, ","))
	}
	args = append(args, host)

	cmd := exec.CommandContext(ctx, bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("nmap scan timed out or cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("nmap execution failed: %v (%s)", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func parseNmapNormal(out []byte) []PortResult {
	var open []PortResult
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		m := nmapLineRe.FindStringSubmatch(strings.TrimSpace(scanner.Text()))
		if len(m) == 0 || !strings.EqualFold(m[3], "open") {
			continue
		}
		port, _ := strconv.Atoi(m[1])
		open = append(open, PortResult{
			Port:     port,
			Protocol: strings.ToLower(m[2]),
			State:    "open",
			Service:  m[4],
			Version:  strings.TrimSpace(m[5]),
		})
	}
	sort.Slice(open, func(i, j int) bool { return open[i].Port < open[j].Port })
	return open
}

// nativeConnectScan performs a concurrent TCP connect scan with lightweight
// banner grabbing. Safe for unprivileged execution (no raw sockets).
func nativeConnectScan(ctx context.Context, host string, ports []int, rate int) []PortResult {
	workers := rate
	if workers < 1 {
		workers = 1
	}
	if workers > nativeDialCap {
		workers = nativeDialCap
	}

	jobs := make(chan int)
	results := make(chan PortResult, len(ports))
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				if ctx.Err() != nil {
					return
				}
				if pr, ok := probePort(ctx, host, p); ok {
					results <- pr
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for _, p := range ports {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()
	wg.Wait()
	close(results)

	var open []PortResult
	for pr := range results {
		open = append(open, pr)
	}
	sort.Slice(open, func(i, j int) bool { return open[i].Port < open[j].Port })
	return open
}

func probePort(ctx context.Context, host string, port int) (PortResult, bool) {
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return PortResult{}, false
	}
	defer conn.Close()

	pr := PortResult{Port: port, Protocol: "tcp", State: "open"}
	_ = conn.SetReadDeadline(time.Now().Add(1500 * time.Millisecond))
	buf := make([]byte, 512)
	if n, err := conn.Read(buf); err == nil && n > 0 {
		pr.Banner = strings.TrimSpace(strings.Map(func(r rune) rune {
			if r < 32 || r == 127 {
				return ' '
			}
			return r
		}, string(buf[:n])))
		if len(pr.Banner) > 120 {
			pr.Banner = pr.Banner[:120]
		}
	}
	return pr, true
}

func renderNativeDump(res *ScanResult) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Cybermes native-go port scan: %s\n", res.Host))
	for _, p := range res.OpenPorts {
		line := fmt.Sprintf("%d/%s open %s", p.Port, p.Protocol, p.Service)
		if p.Banner != "" {
			line += " " + p.Banner
		}
		sb.WriteString(line + "\n")
	}
	return []byte(sb.String())
}

func saveDump(outputDir string, data []byte, res *ScanResult) {
	if outputDir == "" || len(data) == 0 {
		return
	}
	_ = os.MkdirAll(outputDir, 0755)
	dumpFile := filepath.Join(outputDir, "nmap_output.txt")
	if err := os.WriteFile(dumpFile, data, 0644); err == nil {
		res.SavedFilePath = dumpFile
	}
}
