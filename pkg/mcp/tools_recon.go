package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"cybermes/pkg/crawl"
	"cybermes/pkg/probe"
	"cybermes/pkg/scope"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerReconTools() {
	// 1. cybermes_validate_scope
	validateScopeTool := mcp.NewTool(
		"cybermes_validate_scope",
		mcp.WithDescription("Verify whether a target URL, domain, or IP address is permitted within the engagement scope definition (scope.yaml)."),
		mcp.WithString(
			"target",
			mcp.Required(),
			mcp.Description("Target URL, domain, or IP (e.g. 'https://api.example.com', '192.168.1.50', 'admin.example.com')."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug identifier to locate specific reports/<target_slug>/scope.yaml (optional)."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	// 2. cybermes_http_probe
	httpProbeTool := mcp.NewTool(
		"cybermes_http_probe",
		mcp.WithDescription("Inspect an HTTP/HTTPS service, identify status codes, headers, TLS certificates, and fingerprint backend/frontend technologies."),
		mcp.WithString(
			"target_url",
			mcp.Required(),
			mcp.Description("Full HTTP/HTTPS URL to probe (e.g. 'https://example.com' or 'http://127.0.0.1:8888')."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug for scope enforcement and logging (e.g. 'example_com')."),
		),
		mcp.WithBoolean(
			"follow_redirects",
			mcp.Description("Whether to follow HTTP 3xx redirects (default: false)."),
		),
		mcp.WithNumber(
			"timeout_seconds",
			mcp.Description("Request timeout in seconds (default: 10)."),
			mcp.DefaultNumber(10),
		),
		mcp.WithBoolean(
			"prefer_httpx",
			mcp.Description("Attempt to use external httpx binary if available before falling back to native Go (default: true)."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	// 3. cybermes_recon_crawl
	reconCrawlTool := mcp.NewTool(
		"cybermes_recon_crawl",
		mcp.WithDescription("Crawl and mine web endpoints, API routes, and JavaScript bundles with Smart Pipe token budgeting."),
		mcp.WithString(
			"target_url",
			mcp.Required(),
			mcp.Description("Base URL of web application to crawl (e.g. 'https://example.com')."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug directory for saving full output to recon/<target_slug>/katana.txt."),
		),
		mcp.WithNumber(
			"depth",
			mcp.Description("Crawl depth level (default: 2)."),
			mcp.DefaultNumber(2),
		),
		mcp.WithNumber(
			"max_endpoints",
			mcp.Description("Maximum top high-signal endpoints to include in the context response (default: 25)."),
			mcp.DefaultNumber(25),
		),
		mcp.WithNumber(
			"timeout_seconds",
			mcp.Description("Crawl execution timeout in seconds (default: 30)."),
			mcp.DefaultNumber(30),
		),
		mcp.WithBoolean(
			"prefer_katana",
			mcp.Description("Attempt to use external katana binary if available before falling back to native Go crawler (default: true)."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(validateScopeTool, s.handleValidateScope)
	s.mcpServer.AddTool(httpProbeTool, s.handleHttpProbe)
	s.mcpServer.AddTool(reconCrawlTool, s.handleReconCrawl)
}

func (s *Server) handleValidateScope(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target := strings.TrimSpace(request.GetString("target", ""))
	if target == "" {
		return mcp.NewToolResultError("target parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	format := request.GetString("format", "markdown")

	cfg, _, _ := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	result := scope.ValidateTarget(target, cfg)

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString("# 🎯 Cybermes Scope Guard Validation\n\n")

	statusBadge := "🟢 **IN SCOPE / AUTHORIZED**"
	if !result.Allowed {
		statusBadge = "🔴 **OUT OF SCOPE / BLOCKED**"
	}

	sb.WriteString(fmt.Sprintf("- **Target**: `%s`\n", result.Target))
	sb.WriteString(fmt.Sprintf("- **Status**: %s\n", statusBadge))
	sb.WriteString(fmt.Sprintf("- **Host**: `%s`\n", result.Host))
	if result.Port != "" {
		sb.WriteString(fmt.Sprintf("- **Port**: `%s`\n", result.Port))
	}
	if result.Path != "" {
		sb.WriteString(fmt.Sprintf("- **Path**: `%s`\n", result.Path))
	}
	if result.MatchedBy != "" {
		sb.WriteString(fmt.Sprintf("- **Rule Matched**: `%s`\n", result.MatchedBy))
	}
	sb.WriteString(fmt.Sprintf("- **Details**: %s\n", result.Reason))

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleHttpProbe(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL := strings.TrimSpace(request.GetString("target_url", ""))
	if targetURL == "" {
		return mcp.NewToolResultError("target_url parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	followRedirects := request.GetBool("follow_redirects", false)
	timeoutSec := request.GetInt("timeout_seconds", 10)
	preferHttpx := request.GetBool("prefer_httpx", true)
	format := request.GetString("format", "markdown")

	// Scope check
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(targetURL, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason)), nil
		}
	}

	opts := probe.ProbeOptions{
		TargetURL:       targetURL,
		Timeout:         time.Duration(timeoutSec) * time.Second,
		FollowRedirects: followRedirects,
		ToolsDir:        s.cfg.ToolsDir,
		PreferHttpx:     preferHttpx,
	}

	result, err := probe.ProbeTarget(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("HTTP Probe failed: %v", err)), nil
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 🌐 HTTP Probe Inspection: `%s`\n\n", result.URL))
	sb.WriteString("| Attribute | Value |\n| :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Status Code** | `%d %s` |\n", result.StatusCode, result.StatusText))
	if result.Title != "" {
		sb.WriteString(fmt.Sprintf("| **Page Title** | %s |\n", result.Title))
	}
	if result.WebServer != "" {
		sb.WriteString(fmt.Sprintf("| **Web Server** | `%s` |\n", result.WebServer))
	}
	if result.ContentType != "" {
		sb.WriteString(fmt.Sprintf("| **Content Type** | `%s` |\n", result.ContentType))
	}
	sb.WriteString(fmt.Sprintf("| **Response Time** | `%d ms` |\n", result.ResponseTimeMs))
	sb.WriteString(fmt.Sprintf("| **Engine Used** | `%s` |\n", result.EngineUsed))

	if len(result.Technologies) > 0 {
		sb.WriteString(fmt.Sprintf("| **Detected Tech** | %s |\n", strings.Join(result.Technologies, ", ")))
	}

	if result.TLSInfo != nil {
		sb.WriteString(fmt.Sprintf("| **TLS Version** | `%s` |\n", result.TLSInfo.Version))
		if result.TLSInfo.CipherSuite != "" {
			sb.WriteString(fmt.Sprintf("| **Cipher Suite** | `%s` |\n", result.TLSInfo.CipherSuite))
		}
		if result.TLSInfo.Issuer != "" {
			sb.WriteString(fmt.Sprintf("| **Certificate Issuer** | `%s` |\n", result.TLSInfo.Issuer))
		}
		if result.TLSInfo.ExpiresAt != "" {
			sb.WriteString(fmt.Sprintf("| **Cert Expiry** | `%s` |\n", result.TLSInfo.ExpiresAt))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleReconCrawl(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL := strings.TrimSpace(request.GetString("target_url", ""))
	if targetURL == "" {
		return mcp.NewToolResultError("target_url parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	if targetSlug == "" {
		targetSlug = sanitizeSlug(targetURL)
	}

	depth := request.GetInt("depth", 2)
	maxEndpoints := request.GetInt("max_endpoints", 25)
	timeoutSec := request.GetInt("timeout_seconds", 30)
	preferKatana := request.GetBool("prefer_katana", true)
	format := request.GetString("format", "markdown")

	// Scope Guard check
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(targetURL, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason)), nil
		}
	}

	outputDir := filepath.Join(s.cfg.RootDir, "recon", targetSlug)

	opts := crawl.CrawlOptions{
		TargetURL:    targetURL,
		Depth:        depth,
		MaxEndpoints: maxEndpoints,
		Timeout:      time.Duration(timeoutSec) * time.Second,
		ToolsDir:     s.cfg.ToolsDir,
		OutputDir:    outputDir,
		PreferKatana: preferKatana,
	}

	result, err := crawl.CrawlTarget(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Recon Crawl failed: %v", err)), nil
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 🕷️ Recon Crawl Discovery: `%s`\n\n", result.TargetURL))
	sb.WriteString(fmt.Sprintf("- **Total Endpoints Found**: `%d`\n", result.TotalEndpointsFound))
	sb.WriteString(fmt.Sprintf("- **Displayed High-Signal Endpoints**: `%d`\n", len(result.TopEndpoints)))
	sb.WriteString(fmt.Sprintf("- **Engine Used**: `%s` (Completed in %d ms)\n", result.EngineUsed, result.DurationMs))
	if result.SavedFilePath != "" {
		sb.WriteString(fmt.Sprintf("- **Full Raw Output Preserved**: `%s`\n", result.SavedFilePath))
	}
	sb.WriteString("\n### 🎯 Top High-Signal Endpoints (Ranked by Smart Pipe)\n\n")
	sb.WriteString("| Score | Endpoint / Route |\n| :---: | :--- |\n")

	for _, ep := range result.TopEndpoints {
		sb.WriteString(fmt.Sprintf("| `%d` | `%s` |\n", ep.Score, ep.Text))
	}

	if result.TotalEndpointsFound > len(result.TopEndpoints) {
		sb.WriteString(fmt.Sprintf("\n> 💡 *Note: Remaining %d lower-entropy endpoints filtered to preserve token economy. Full list saved in disk.*",
			result.TotalEndpointsFound-len(result.TopEndpoints)))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
