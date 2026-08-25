package scope

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ScopeConfig represents target scope definitions from scope.yaml.
type ScopeConfig struct {
	Name        string   `yaml:"name" json:"name"`
	TargetSlug  string   `yaml:"target_slug" json:"target_slug"`
	InScope     []string `yaml:"in_scope" json:"in_scope"`
	OutOfScope  []string `yaml:"out_of_scope" json:"out_of_scope"`
	AllowIPs    bool     `yaml:"allow_ips" json:"allow_ips"`
	MaxRequests int      `yaml:"max_requests" json:"max_requests"`
}

// ValidationResult contains the outcome of a target scope evaluation.
type ValidationResult struct {
	Allowed    bool   `json:"allowed"`
	Target     string `json:"target"`
	Host       string `json:"host"`
	Port       string `json:"port,omitempty"`
	Path       string `json:"path,omitempty"`
	MatchedBy  string `json:"matched_by,omitempty"`
	Reason     string `json:"reason"`
	ScopeFound bool   `json:"scope_found"`
}

// ParseScopeYAML parses raw YAML bytes into a ScopeConfig.
func ParseScopeYAML(data []byte) (*ScopeConfig, error) {
	var cfg ScopeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse scope YAML: %w", err)
	}
	return &cfg, nil
}

// ParseScopeFile reads and parses a scope configuration file.
func ParseScopeFile(filePath string) (*ScopeConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read scope file '%s': %w", filePath, err)
	}
	return ParseScopeYAML(data)
}

// FindScopeConfig locates the appropriate scope.yaml file for a target.
// It searches in reports/<targetSlug>/scope.yaml, then root/scope.yaml.
func FindScopeConfig(rootDir string, targetSlug string) (*ScopeConfig, string, error) {
	candidates := make([]string, 0, 4)
	if targetSlug != "" && rootDir != "" {
		candidates = append(candidates,
			filepath.Join(rootDir, "reports", targetSlug, "scope.yaml"),
			filepath.Join(rootDir, "reports", targetSlug, "scope.yml"),
		)
	}
	if rootDir != "" {
		candidates = append(candidates,
			filepath.Join(rootDir, "scope.yaml"),
			filepath.Join(rootDir, "scope.yml"),
		)
	}

	for _, path := range candidates {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			cfg, err := ParseScopeFile(path)
			if err == nil {
				return cfg, path, nil
			}
		}
	}

	return nil, "", fmt.Errorf("no scope.yaml found for target '%s'", targetSlug)
}

// NormalizeTarget parses and normalizes a target input (URL, domain, or IP:Port).
func NormalizeTarget(rawTarget string) (scheme, host, port, path string, err error) {
	target := strings.TrimSpace(rawTarget)
	if target == "" {
		return "", "", "", "", fmt.Errorf("target string is empty")
	}

	if !strings.Contains(target, "://") {
		target = "http://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return "", "", "", "", fmt.Errorf("invalid target URL: %w", err)
	}

	scheme = strings.ToLower(u.Scheme)
	host = u.Hostname()
	port = u.Port()
	path = u.Path
	if path == "" {
		path = "/"
	}

	return scheme, strings.ToLower(host), port, path, nil
}

// ValidateTarget checks whether a given target is permitted under the ScopeConfig rules.
func ValidateTarget(rawTarget string, cfg *ScopeConfig) ValidationResult {
	if cfg == nil {
		// If no scope is defined, by default we consider operator targets authorized under AGENTS.md rule 2.1
		_, host, port, path, err := NormalizeTarget(rawTarget)
		if err != nil {
			return ValidationResult{
				Allowed:    false,
				Target:     rawTarget,
				Reason:     fmt.Sprintf("Invalid target format: %v", err),
				ScopeFound: false,
			}
		}
		return ValidationResult{
			Allowed:    true,
			Target:     rawTarget,
			Host:       host,
			Port:       port,
			Path:       path,
			Reason:     "No scope file loaded; target permitted under Direct Operator Authorization (AGENTS.md).",
			ScopeFound: false,
		}
	}

	_, host, port, path, err := NormalizeTarget(rawTarget)
	if err != nil {
		return ValidationResult{
			Allowed:    false,
			Target:     rawTarget,
			Reason:     fmt.Sprintf("Invalid target format: %v", err),
			ScopeFound: true,
		}
	}

	// 1. Check Out-of-Scope rules first (Strict Exclusion)
	for _, outRule := range cfg.OutOfScope {
		outRule = strings.TrimSpace(outRule)
		if outRule == "" {
			continue
		}
		if matchesRule(host, port, path, rawTarget, outRule) {
			return ValidationResult{
				Allowed:    false,
				Target:     rawTarget,
				Host:       host,
				Port:       port,
				Path:       path,
				MatchedBy:  "out_of_scope: " + outRule,
				Reason:     fmt.Sprintf("Target '%s' is explicitly OUT OF SCOPE (matched rule '%s')", rawTarget, outRule),
				ScopeFound: true,
			}
		}
	}

	// 2. Check In-Scope rules
	if len(cfg.InScope) == 0 {
		return ValidationResult{
			Allowed:    true,
			Target:     rawTarget,
			Host:       host,
			Port:       port,
			Path:       path,
			Reason:     "In-scope list is empty; target passed out-of-scope check.",
			ScopeFound: true,
		}
	}

	for _, inRule := range cfg.InScope {
		inRule = strings.TrimSpace(inRule)
		if inRule == "" {
			continue
		}
		if matchesRule(host, port, path, rawTarget, inRule) {
			return ValidationResult{
				Allowed:    true,
				Target:     rawTarget,
				Host:       host,
				Port:       port,
				Path:       path,
				MatchedBy:  "in_scope: " + inRule,
				Reason:     fmt.Sprintf("Target '%s' is IN SCOPE (matched rule '%s')", rawTarget, inRule),
				ScopeFound: true,
			}
		}
	}

	// 3. Not matched in in-scope list
	return ValidationResult{
		Allowed:    false,
		Target:     rawTarget,
		Host:       host,
		Port:       port,
		Path:       path,
		Reason:     fmt.Sprintf("Target '%s' (host: %s) does not match any allowed in-scope pattern.", rawTarget, host),
		ScopeFound: true,
	}
}

// matchesRule evaluates host, port, path, or rawTarget against a single rule pattern.
func matchesRule(host, port, path, rawTarget, rule string) bool {
	rule = strings.ToLower(rule)
	host = strings.ToLower(host)

	// CIDR check (e.g. 192.168.1.0/24)
	if strings.Contains(rule, "/") && !strings.Contains(rule, "://") {
		if _, ipNet, err := net.ParseCIDR(rule); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				return ipNet.Contains(ip)
			}
		}
	}

	// Direct IP check
	if ip := net.ParseIP(host); ip != nil && ip.String() == rule {
		return true
	}

	// Full URL or Path match
	if strings.Contains(rule, "://") {
		_, rHost, rPort, rPath, err := NormalizeTarget(rule)
		if err == nil {
			if host == rHost && (rPort == "" || port == rPort) {
				if rPath == "/" || strings.HasPrefix(path, rPath) {
					return true
				}
			}
		}
	} else if strings.HasPrefix(rule, "/") {
		if strings.HasPrefix(path, rule) {
			return true
		}
	}

	// Wildcard subdomain matching (e.g. *.example.com)
	if strings.HasPrefix(rule, "*.") {
		rootDomain := strings.TrimPrefix(rule, "*.")
		if host == rootDomain || strings.HasSuffix(host, "."+rootDomain) {
			return true
		}
	}

	// Exact host or domain match
	if host == rule {
		return true
	}

	// Host + Port match (e.g. example.com:8080 or 127.0.0.1:8888)
	if port != "" && fmt.Sprintf("%s:%s", host, port) == rule {
		return true
	}

	// Regex fallback if rule starts and ends with /
	if strings.HasPrefix(rule, "^") || strings.HasSuffix(rule, "$") {
		if re, err := regexp.Compile(rule); err == nil {
			if re.MatchString(rawTarget) || re.MatchString(host) {
				return true
			}
		}
	}

	return false
}
