package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"cybermes/pkg/nmap"
	"cybermes/pkg/scope"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerNmapTools() {
	nmapTool := mcp.NewTool(
		"cybermes_nmap_scan",
		mcp.WithDescription("Discover open TCP ports and fingerprint services with nmap (unprivileged connect scan) or the native Go fallback engine. Rate-limited, Scope Guard enforced, raw output preserved to recon/."),
		mcp.WithString(
			"target",
			mcp.Required(),
			mcp.Description("Target host, IP, host:port, or URL (e.g. '192.168.1.10', 'api.example.com', 'http://127.0.0.1:8888'). Port ranges come from 'ports'."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug for logging and scope enforcement (e.g. 'example_com')."),
		),
		mcp.WithString(
			"ports",
			mcp.Description("Ports to scan: 'top-100' (default), comma list '80,443', or range '1-1024' (max 1000 ports)."),
			mcp.DefaultString("top-100"),
		),
		mcp.WithNumber(
			"rate_limit",
			mcp.Description("Max concurrent probes / nmap --max-rate (default: 25, max: 100)."),
			mcp.DefaultNumber(25),
		),
		mcp.WithNumber(
			"timeout_seconds",
			mcp.Description("Scan execution timeout in seconds (default: 60)."),
			mcp.DefaultNumber(60),
		),
		mcp.WithBoolean(
			"service_detection",
			mcp.Description("Enable light service/version fingerprinting (-sV, default: true). Disable for fastest port discovery."),
			mcp.DefaultBool(true),
		),
		mcp.WithBoolean(
			"prefer_nmap",
			mcp.Description("Attempt to use the external nmap binary before falling back to the native Go connect scanner (default: true)."),
			mcp.DefaultBool(true),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary table) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(nmapTool, s.handleNmapScan)
}

func (s *Server) handleNmapScan(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target := strings.TrimSpace(request.GetString("target", ""))
	if target == "" {
		return mcp.NewToolResultError("target parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	if targetSlug == "" {
		targetSlug = sanitizeSlug(target)
	} else {
		targetSlug = sanitizeSlug(targetSlug)
	}

	ports := strings.TrimSpace(request.GetString("ports", "top-100"))
	rateLimit := request.GetInt("rate_limit", 25)
	if rateLimit <= 0 || rateLimit > 100 {
		rateLimit = 25
	}
	timeoutSec := request.GetInt("timeout_seconds", 60)
	if timeoutSec <= 0 {
		timeoutSec = 60
	}
	serviceDetection := request.GetBool("service_detection", true)
	preferNmap := request.GetBool("prefer_nmap", true)
	format := request.GetString("format", "markdown")

	// Scope Guard check
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(target, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason)), nil
		}
	}

	reconDir := filepath.Join(s.cfg.RootDir, "recon", targetSlug)

	opts := nmap.NmapOptions{
		Target:               target,
		Ports:                ports,
		RateLimit:            rateLimit,
		Timeout:              time.Duration(timeoutSec) * time.Second,
		ToolsDir:             s.cfg.ToolsDir,
		OutputDir:            reconDir,
		PreferNmap:           preferNmap,
		SkipServiceDetection: !serviceDetection,
	}

	res, err := nmap.ScanTarget(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Port scan failed: %v", err)), nil
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 🛰️ Nmap Port Scan Results: `%s`\n\n", res.Host))
	sb.WriteString(fmt.Sprintf("- **Ports Scanned**: `%d`\n", res.PortsScanned))
	sb.WriteString(fmt.Sprintf("- **Open Ports**: `%d`\n", len(res.OpenPorts)))
	sb.WriteString(fmt.Sprintf("- **Engine Used**: `%s` (Completed in %d ms)\n", res.EngineUsed, res.DurationMs))
	if res.SavedFilePath != "" {
		if rel, relErr := filepath.Rel(s.cfg.RootDir, res.SavedFilePath); relErr == nil {
			sb.WriteString(fmt.Sprintf("- **Full Output Dump**: `%s`\n", rel))
		} else {
			sb.WriteString(fmt.Sprintf("- **Full Output Dump**: `%s`\n", res.SavedFilePath))
		}
	}
	sb.WriteString("\n")

	if len(res.OpenPorts) == 0 {
		sb.WriteString("✅ **No open TCP ports detected** in the scanned range.\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	sb.WriteString("| Port | State | Service | Version / Banner |\n")
	sb.WriteString("| :---: | :---: | :--- | :--- |\n")
	for _, p := range res.OpenPorts {
		detail := strings.TrimSpace(strings.Trim(p.Version+" "+p.Banner, " "))
		if detail == "" {
			detail = "-"
		}
		svc := strings.TrimSpace(p.Service)
		if svc == "" {
			svc = "-"
		}
		sb.WriteString(fmt.Sprintf("| `%d/%s` | `%s` | `%s` | `%s` |\n",
			p.Port, p.Protocol, p.State, svc, detail))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
