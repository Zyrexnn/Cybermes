package mcp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"cybermes/pkg/scope"
	"github.com/mark3labs/mcp-go/mcp"
)

func sanitizeSlug(input string) string {
	return scope.SanitizeSlug(input)
}

func (s *Server) resolveSlug(explicitSlug, defaultInput string) string {
	explicit := strings.TrimSpace(explicitSlug)
	if explicit != "" {
		return sanitizeSlug(explicit)
	}
	return sanitizeSlug(defaultInput)
}

func (s *Server) enforceScope(target, targetSlug string) *mcp.CallToolResult {
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(target, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason))
		}
	}
	return nil
}

func parseCustomHeaders(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	headers := make(map[string]string)
	if strings.HasPrefix(raw, "{") {
		var jsonMap map[string]any
		if err := json.Unmarshal([]byte(raw), &jsonMap); err == nil {
			for k, v := range jsonMap {
				headers[k] = fmt.Sprintf("%v", v)
			}
			return headers
		}
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

func getSeverityBadge(sev string) string {
	switch strings.ToLower(sev) {
	case "critical":
		return "🔴 **CRITICAL**"
	case "high":
		return "🟠 **HIGH**"
	case "medium":
		return "🟡 **MEDIUM**"
	default:
		return "🔵 **LOW**"
	}
}

func isSafeSubpath(baseDir, targetPath string) bool {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != "."
}
