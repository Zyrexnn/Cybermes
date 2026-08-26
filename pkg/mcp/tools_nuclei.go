package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cybermes/pkg/scope"
	"github.com/mark3labs/mcp-go/mcp"
)

type NucleiFinding struct {
	TemplateID  string   `json:"template-id"`
	Info        struct {
		Name        string   `json:"name"`
		Severity    string   `json:"severity"`
		Description string   `json:"description,omitempty"`
		Tags        []string `json:"tags,omitempty"`
		Reference   []string `json:"reference,omitempty"`
	} `json:"info"`
	MatchedAt   string   `json:"matched-at"`
	Type        string   `json:"type"`
	Timestamp   string   `json:"timestamp"`
	CurlCommand string   `json:"curl-command,omitempty"`
	Extracted   []string `json:"extracted-results,omitempty"`
}

func (s *Server) registerNucleiTools() {
	nucleiTool := mcp.NewTool(
		"cybermes_nuclei_scan",
		mcp.WithDescription("Run targeted, non-destructive vulnerability template scans using Nuclei (CVE verification, misconfigurations, exposed panels, auth-bypasses). Includes on-demand dependency detection and rate limiting."),
		mcp.WithString(
			"target_url",
			mcp.Required(),
			mcp.Description("Target URL or endpoint to scan (e.g. 'https://api.example.com' or 'http://127.0.0.1:8888')."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug for logging and scope enforcement (e.g. 'example_com')."),
		),
		mcp.WithString(
			"tags",
			mcp.Description("Comma-separated template tags to execute (e.g. 'cve,auth-bypass,misconfig,exposure,panel')."),
			mcp.DefaultString("cve,auth-bypass,misconfig"),
		),
		mcp.WithString(
			"severity",
			mcp.Description("Filter template severity levels (e.g. 'critical,high,medium' or 'critical,high')."),
			mcp.DefaultString("critical,high,medium"),
		),
		mcp.WithString(
			"template_id",
			mcp.Description("Specific single template ID or CVE ID to verify (e.g. 'cve-2023-46604', 'exposed-panels')."),
		),
		mcp.WithNumber(
			"rate_limit",
			mcp.Description("Maximum requests per second (safe default: 10, max: 20)."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber(
			"timeout_seconds",
			mcp.Description("Scan execution timeout in seconds (default: 45)."),
			mcp.DefaultNumber(45),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary table) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(nucleiTool, s.handleNucleiScan)
}

func findNucleiBinary(toolsDir string) string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	// 1. Check tools/bin/nuclei
	candidate := filepath.Join(toolsDir, "bin", "nuclei"+ext)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	// 2. Check system PATH
	if p, err := exec.LookPath("nuclei" + ext); err == nil {
		return p
	}
	if p, err := exec.LookPath("nuclei"); err == nil {
		return p
	}

	return ""
}

func getOnDemandInstallInstructions() string {
	switch runtime.GOOS {
	case "windows":
		return "`winget install projectdiscovery.nuclei` atau `pdtm -i nuclei` atau jalankan `.\\tools\\update_tools.ps1 -IncludeNuclei`"
	case "darwin":
		return "`brew install nuclei` atau `pdtm -i nuclei`"
	default:
		return "`pdtm -i nuclei` atau `go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest`"
	}
}

func (s *Server) handleNucleiScan(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL := strings.TrimSpace(request.GetString("target_url", ""))
	if targetURL == "" {
		return mcp.NewToolResultError("target_url parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	if targetSlug == "" {
		targetSlug = sanitizeSlug(targetURL)
	}

	tags := strings.TrimSpace(request.GetString("tags", "cve,auth-bypass,misconfig"))
	severity := strings.TrimSpace(request.GetString("severity", "critical,high,medium"))
	templateID := strings.TrimSpace(request.GetString("template_id", ""))
	rateLimit := request.GetInt("rate_limit", 10)
	if rateLimit <= 0 || rateLimit > 20 {
		rateLimit = 10
	}
	timeoutSec := request.GetInt("timeout_seconds", 45)
	if timeoutSec <= 0 {
		timeoutSec = 45
	}
	format := request.GetString("format", "markdown")

	// Scope Guard check
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(targetURL, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason)), nil
		}
	}

	// Locate binary
	nucleiBin := findNucleiBinary(s.cfg.ToolsDir)
	if nucleiBin == "" {
		// Return friendly On-Demand missing dependency guide instead of crashing
		instructions := getOnDemandInstallInstructions()
		msg := fmt.Sprintf("ℹ️ **Nuclei is an Optional / On-Demand Tool** and is not currently installed on this system.\n\n"+
			"To enable automatic CVE and template scanning with Nuclei, install it using:\n"+
			"- **Install Command**: %s\n\n"+
			"💡 *Tip: Cybermes core testing (IDOR, API auth, parameter mining, secrets, and crawling) continues to function normally without Nuclei.*",
			instructions)
		return mcp.NewToolResultText(msg), nil
	}

	// Prepare safe execution arguments
	args := []string{
		"-u", targetURL,
		"-rate-limit", fmt.Sprintf("%d", rateLimit),
		"-silent",
		"-jsonl",
		"-timeout", "10",
		"-duc", // disable update check during active scans
	}

	if templateID != "" {
		args = append(args, "-t", templateID)
	} else {
		if tags != "" {
			args = append(args, "-tags", tags)
		}
		if severity != "" {
			args = append(args, "-severity", severity)
		}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, nucleiBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()
	if execErr != nil && cmdCtx.Err() == context.DeadlineExceeded {
		return mcp.NewToolResultError(fmt.Sprintf("Nuclei scan timed out after %d seconds.", timeoutSec)), nil
	}

	// Parse JSONL results
	var findings []NucleiFinding
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var f NucleiFinding
		if err := json.Unmarshal(line, &f); err == nil && f.TemplateID != "" {
			findings = append(findings, f)
		}
	}

	// Save raw output to recon directory
	reconDir := filepath.Join(s.cfg.RootDir, "recon", targetSlug)
	_ = os.MkdirAll(reconDir, 0755)
	dumpFile := filepath.Join(reconDir, "nuclei_output.txt")
	_ = os.WriteFile(dumpFile, stdout.Bytes(), 0644)

	if strings.ToLower(format) == "json" {
		out := map[string]any{
			"target_url":     targetURL,
			"target_slug":    targetSlug,
			"total_findings": len(findings),
			"findings":       findings,
			"raw_dump_file":  dumpFile,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 🎯 Nuclei Vulnerability Scan Results: `%s`\n\n", targetURL))
	sb.WriteString(fmt.Sprintf("- **Total Templates Triggered**: `%d`\n", len(findings)))
	sb.WriteString(fmt.Sprintf("- **Rate Limit Enforced**: `%d req/s`\n", rateLimit))
	sb.WriteString(fmt.Sprintf("- **Full Output Dump**: `%s`\n\n", dumpFile))

	if len(findings) == 0 {
		sb.WriteString("✅ **No template vulnerabilities detected.** Target is secured against specified tags/CVE checks.\n")
		sb.WriteString(fmt.Sprintf("\n> 💡 *Note: Evaluated tags: `%s` | Severity: `%s`*", tags, severity))
		return mcp.NewToolResultText(sb.String()), nil
	}

	sb.WriteString("| Severity | Template ID | Vulnerability Name | Matched URL |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")

	for _, f := range findings {
		sevBadge := getSeverityBadge(f.Info.Severity)
		sb.WriteString(fmt.Sprintf("| %s | `%s` | **%s** | `%s` |\n",
			sevBadge, f.TemplateID, f.Info.Name, f.MatchedAt))
	}

	sb.WriteString("\n⚠️ *Action Item: Validate high-signal findings with reproducible evidence before recording to findings/.*")
	return mcp.NewToolResultText(sb.String()), nil
}
