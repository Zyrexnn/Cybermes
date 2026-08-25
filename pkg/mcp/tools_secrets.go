package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cybermes/pkg/secrets"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerSecretsTools() {
	tool := mcp.NewTool(
		"cybermes_scan_secrets",
		mcp.WithDescription("Scan raw text, code snippets, HTTP responses, or local files/directories for credentials and API keys using Cybermes' 48-pattern high-precision secret detector (AWS, GCP, GitHub, Slack, Stripe, Private Keys, JWTs, etc.)."),
		mcp.WithString(
			"content",
			mcp.Description("Raw text, source code, HTTP payload, or config file content to scan for leaks directly in memory."),
		),
		mcp.WithString(
			"path",
			mcp.Description("Path to a local file or directory to scan recursively. Can be relative to workspace or absolute."),
		),
		mcp.WithString(
			"min_severity",
			mcp.Description("Minimum severity threshold to report: 'low', 'medium', 'high', 'critical'."),
			mcp.Enum("low", "medium", "high", "critical"),
			mcp.DefaultString("low"),
		),
		mcp.WithBoolean(
			"mask_secrets",
			mcp.Description("Mask the sensitive inner characters of detected secrets (e.g. 'AKIA...X9Z') to prevent context leak (default: true)."),
			mcp.DefaultBool(true),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary list) or 'json' (structured findings)."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleScanSecrets)
}

func (s *Server) handleScanSecrets(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := request.GetString("content", "")
	targetPath := request.GetString("path", "")
	minSev := strings.ToLower(request.GetString("min_severity", "low"))
	mask := request.GetBool("mask_secrets", true)
	format := request.GetString("format", "markdown")

	if content == "" && targetPath == "" {
		return mcp.NewToolResultError("Either 'content' (raw text) or 'path' (file/directory) must be provided."), nil
	}

	var findings []secrets.Finding

	if content != "" {
		findings = secrets.ScanText(content, "raw_content")
	} else {
		// Resolve path relative to RootDir if not absolute
		absPath := targetPath
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(s.cfg.RootDir, absPath)
		}

		info, err := os.Stat(absPath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Target path not found: %s", targetPath)), nil
		}

		if info.IsDir() {
			findings, err = secrets.ScanDirectory(absPath, 8)
		} else {
			findings, err = secrets.ScanFile(absPath)
		}
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Scanning error: %v", err)), nil
		}
	}

	// Filter by minimum severity
	filtered := filterFindingsBySeverity(findings, minSev)

	if mask {
		for i := range filtered {
			filtered[i].Match = maskSecret(filtered[i].Match)
		}
	}

	if len(filtered) == 0 {
		return mcp.NewToolResultText("✅ No secrets or credential leaks detected matching criteria."), nil
	}

	if format == "json" {
		out := map[string]any{
			"total_detected": len(findings),
			"reported_count": len(filtered),
			"min_severity":   minSev,
			"findings":       filtered,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 🚨 Cybermes Secret Scan: Found %d Leaked Credential(s)\n\n", len(filtered)))
	sb.WriteString("| Severity | Pattern Type | Category | Location | Masked Match |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

	for _, f := range filtered {
		sevBadge := getSeverityBadge(f.Severity)
		loc := f.Source
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.Source, f.Line)
		}
		sb.WriteString(fmt.Sprintf("| %s | `%s` | `%s` | `%s` | `%s` |\n",
			sevBadge, f.Pattern, f.Category, loc, f.Match))
	}

	sb.WriteString("\n⚠️ *Action Item: Revoke and rotate exposed credentials immediately.*")
	return mcp.NewToolResultText(sb.String()), nil
}

func filterFindingsBySeverity(findings []secrets.Finding, minSev string) []secrets.Finding {
	rank := map[string]int{
		"critical": 4,
		"high":     3,
		"medium":   2,
		"low":      1,
	}

	minRank := rank[minSev]
	if minRank == 0 {
		minRank = 1
	}

	var out []secrets.Finding
	for _, f := range findings {
		r := rank[strings.ToLower(f.Severity)]
		if r >= minRank {
			out = append(out, f)
		}
	}
	return out
}

func maskSecret(raw string) string {
	raw = strings.TrimSpace(raw)
	n := len(raw)
	if n <= 6 {
		return "******"
	}
	if n <= 12 {
		return raw[:2] + strings.Repeat("*", n-4) + raw[n-2:]
	}
	return raw[:4] + "..." + raw[n-4:]
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
