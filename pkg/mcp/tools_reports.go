package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cybermes/pkg/report"
	"github.com/mark3labs/mcp-go/mcp"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func (s *Server) registerReportsTools() {
	aggTool := mcp.NewTool(
		"cybermes_aggregate_report",
		mcp.WithDescription("Aggregate vulnerability findings, PoC scripts, and evidence files for a target (or all targets) into executive SUMMARY.md, metadata.json, report.html, and REPORT.pdf."),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug/directory name in reports/ (e.g. 'example_com', '127_0_0_1_8888'). If empty or 'all', aggregates all targets."),
		),
		mcp.WithBoolean(
			"generate_pdf",
			mcp.Description("Whether to automatically render optional REPORT.pdf via Chrome DevTools Protocol (default: false)."),
			mcp.DefaultBool(false),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary table) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	pdfTool := mcp.NewTool(
		"cybermes_generate_pdf",
		mcp.WithDescription("Render pixel-perfect executive PDF and interactive HTML dashboard reports for a target using native Go Chrome DevTools Protocol."),
		mcp.WithString(
			"target_slug",
			mcp.Required(),
			mcp.Description("Target slug directory name in reports/ (e.g. 'example_com')."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	listFindingsTool := mcp.NewTool(
		"cybermes_list_findings",
		mcp.WithDescription("List all confirmed vulnerability findings, severity breakdown, and PoC status for a specific engagement target."),
		mcp.WithString(
			"target_slug",
			mcp.Required(),
			mcp.Description("Target slug in reports/ (e.g. 'example_com')."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	createFindingTool := mcp.NewTool(
		"cybermes_record_finding",
		mcp.WithDescription("Record a verified, reproducible security vulnerability finding directly into the target's structured reports workspace following strict Cybermes AGENTS.md standards."),
		mcp.WithString(
			"target_slug",
			mcp.Required(),
			mcp.Description("Target slug directory (e.g. 'example_com')."),
		),
		mcp.WithString(
			"severity",
			mcp.Required(),
			mcp.Description("Vulnerability severity level: 'critical', 'high', 'medium', 'low', 'informational'."),
			mcp.Enum("critical", "high", "medium", "low", "informational"),
		),
		mcp.WithString(
			"title",
			mcp.Required(),
			mcp.Description("Concise vulnerability title (e.g. 'IDOR in Invoice API allows unauthorized document access')."),
		),
		mcp.WithString(
			"endpoint",
			mcp.Required(),
			mcp.Description("Vulnerable URL or endpoint path (e.g. 'GET /api/v1/invoices/{id}')."),
		),
		mcp.WithString(
			"description",
			mcp.Required(),
			mcp.Description("Detailed vulnerability description and technical root cause analysis."),
		),
		mcp.WithString(
			"reproduction_steps",
			mcp.Required(),
			mcp.Description("Step-by-step reproduction guide with raw HTTP requests and responses."),
		),
		mcp.WithString(
			"poc_script",
			mcp.Description("Optional standalone Python/Bash PoC script content."),
		),
		mcp.WithString(
			"remediation",
			mcp.Description("Specific code or architecture remediation advice."),
		),
	)

	s.mcpServer.AddTool(aggTool, s.handleAggregateReport)
	s.mcpServer.AddTool(pdfTool, s.handleGeneratePDF)
	s.mcpServer.AddTool(listFindingsTool, s.handleListFindings)
	s.mcpServer.AddTool(createFindingTool, s.handleRecordFinding)
}

func (s *Server) handleAggregateReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	format := request.GetString("format", "markdown")
	generatePDF := request.GetBool("generate_pdf", false)

	if targetSlug == "" || strings.EqualFold(targetSlug, "all") {
		results, err := report.AggregateAll(s.cfg.ReportsDir)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to aggregate reports: %v", err)), nil
		}

		if format == "json" {
			data, _ := json.MarshalIndent(results, "", "  ")
			return mcp.NewToolResultText(string(data)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("### Cybermes Executive Report Aggregator (All Targets: %d)\n\n", len(results)))
		sb.WriteString("| Target | Total Findings | Critical | High | Medium | Low | Info |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, sm := range results {
			sb.WriteString(fmt.Sprintf("| **`%s`** | %d | %d | %d | %d | %d | %d |\n",
				sm.Target, sm.TotalFindings,
				sm.SeveritySummary["CRITICAL"], sm.SeveritySummary["HIGH"],
				sm.SeveritySummary["MEDIUM"], sm.SeveritySummary["LOW"],
				sm.SeveritySummary["INFORMATIONAL"],
			))
		}
		return mcp.NewToolResultText(sb.String()), nil
	}

	targetDir := filepath.Join(s.cfg.ReportsDir, targetSlug)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create target report directory: %v", err)), nil
		}
	}

	summary, artifacts, err := report.AggregateTargetWithPDF(targetDir, generatePDF)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to aggregate target '%s': %v", targetSlug, err)), nil
	}

	if format == "json" {
		responseObj := map[string]interface{}{
			"summary":   summary,
			"artifacts": artifacts,
		}
		data, _ := json.MarshalIndent(responseObj, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### Report Summary: `%s`\n\n", summary.Target))
	sb.WriteString(fmt.Sprintf("- **Total Findings**: `%d`\n", summary.TotalFindings))
	sb.WriteString(fmt.Sprintf("- **Critical**: `%d` | **High**: `%d` | **Medium**: `%d` | **Low**: `%d` | **Info**: `%d`\n",
		summary.SeveritySummary["CRITICAL"], summary.SeveritySummary["HIGH"],
		summary.SeveritySummary["MEDIUM"], summary.SeveritySummary["LOW"],
		summary.SeveritySummary["INFORMATIONAL"],
	))
	sb.WriteString(fmt.Sprintf("- **PoC Scripts**: `%d` standalone script(s)\n\n", len(summary.PoCs)))
	sb.WriteString("#### Generated Artifacts:\n")
	sb.WriteString(fmt.Sprintf("- **Executive Summary**: `reports/%s/SUMMARY.md`\n", summary.Target))
	sb.WriteString(fmt.Sprintf("- **Metadata JSON**: `reports/%s/metadata.json`\n", summary.Target))
	if artifacts != nil && artifacts.HTMLPath != "" {
		sb.WriteString(fmt.Sprintf("- **Interactive Dashboard**: `reports/%s/report.html`\n", summary.Target))
	}
	if artifacts != nil && artifacts.PDFGenerated {
		sb.WriteString(fmt.Sprintf("- **Executive PDF**: `reports/%s/REPORT.pdf`\n", summary.Target))
	} else if artifacts != nil && artifacts.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("- **PDF Status**: Not generated (%s)\n", artifacts.ErrorMessage))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleGeneratePDF(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetSlug, err := request.RequireString("target_slug")
	if err != nil || strings.TrimSpace(targetSlug) == "" {
		return mcp.NewToolResultError("Missing required parameter: 'target_slug'"), nil
	}

	targetSlug = sanitizeSlug(targetSlug)
	format := request.GetString("format", "markdown")
	targetDir := filepath.Join(s.cfg.ReportsDir, targetSlug)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Report directory for target '%s' does not exist.", targetSlug)), nil
	}

	summary, artifacts, err := report.AggregateTargetWithPDF(targetDir, true)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to generate PDF for '%s': %v", targetSlug, err)), nil
	}

	if format == "json" {
		data, _ := json.MarshalIndent(artifacts, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### PDF & HTML Report Generator: `%s`\n\n", targetSlug))
	sb.WriteString(fmt.Sprintf("- **Total Findings**: `%d`\n", summary.TotalFindings))
	sb.WriteString(fmt.Sprintf("- **HTML Dashboard**: `reports/%s/report.html`\n", targetSlug))

	if artifacts.PDFGenerated {
		sb.WriteString(fmt.Sprintf("- **Executive PDF**: `reports/%s/REPORT.pdf`\n", targetSlug))
		if artifacts.BrowserUsed != "" {
			sb.WriteString(fmt.Sprintf("- **Browser Engine**: `%s`\n", artifacts.BrowserUsed))
		}
		sb.WriteString("\nExecutive PDF and Interactive HTML reports have been compiled successfully.")
	} else {
		sb.WriteString(fmt.Sprintf("- **PDF Status**: Failed (%s)\n", artifacts.ErrorMessage))
		sb.WriteString("\nInteractive HTML dashboard (`report.html`) is available and can be viewed directly.")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleListFindings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetSlug, err := request.RequireString("target_slug")
	if err != nil || strings.TrimSpace(targetSlug) == "" {
		return mcp.NewToolResultError("Missing required parameter: 'target_slug'"), nil
	}

	targetSlug = strings.TrimSpace(targetSlug)
	format := request.GetString("format", "markdown")
	targetDir := filepath.Join(s.cfg.ReportsDir, targetSlug)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Report directory for target '%s' does not exist.", targetSlug)), nil
	}

	summary, err := report.AggregateTarget(targetDir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse findings for '%s': %v", targetSlug, err)), nil
	}

	if format == "json" {
		data, _ := json.MarshalIndent(summary.Findings, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	if len(summary.Findings) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No confirmed findings recorded yet for target `%s`.", targetSlug)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### Confirmed Findings for Target: `%s` (%d total)\n\n", targetSlug, len(summary.Findings)))
	sb.WriteString("| Severity | Title | Vulnerable Endpoint | File |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")

	for _, f := range summary.Findings {
		sevBadge := getSeverityBadge(f.Severity)
		sb.WriteString(fmt.Sprintf("| %s | **%s** | `%s` | `%s` |\n",
			sevBadge, f.Title, f.Endpoint, f.FileName))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleRecordFinding(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetSlug, err := request.RequireString("target_slug")
	if err != nil {
		return mcp.NewToolResultError("Missing 'target_slug'"), nil
	}
	severity, err := request.RequireString("severity")
	if err != nil {
		return mcp.NewToolResultError("Missing 'severity'"), nil
	}
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError("Missing 'title'"), nil
	}
	endpoint, err := request.RequireString("endpoint")
	if err != nil {
		return mcp.NewToolResultError("Missing 'endpoint'"), nil
	}
	description, err := request.RequireString("description")
	if err != nil {
		return mcp.NewToolResultError("Missing 'description'"), nil
	}
	reproSteps, err := request.RequireString("reproduction_steps")
	if err != nil {
		return mcp.NewToolResultError("Missing 'reproduction_steps'"), nil
	}

	pocScript := request.GetString("poc_script", "")
	remediation := request.GetString("remediation", "Implement strict authorization checks and validate user access permissions.")

	targetSlug = sanitizeSlug(targetSlug)
	severity = strings.ToLower(strings.TrimSpace(severity))

	// Ensure directories exist
	targetDir := filepath.Join(s.cfg.ReportsDir, targetSlug)
	findingsDir := filepath.Join(targetDir, "findings")
	pocsDir := filepath.Join(targetDir, "pocs")
	evidenceDir := filepath.Join(targetDir, "evidence")

	for _, dir := range []string{findingsDir, pocsDir, evidenceDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory %s: %v", dir, err)), nil
		}
	}

	vulnSlug := sanitizeSlug(title)
	if len(vulnSlug) > 40 {
		vulnSlug = vulnSlug[:40]
	}
	findingFileName := fmt.Sprintf("%s_%s.md", severity, vulnSlug)
	findingFilePath := filepath.Join(findingsDir, findingFileName)

	var doc strings.Builder
	doc.WriteString(fmt.Sprintf("# %s\n\n", title))
	doc.WriteString(fmt.Sprintf("- **Severity**: %s\n", strings.ToUpper(severity)))
	doc.WriteString(fmt.Sprintf("- **Endpoint**: `%s`\n", endpoint))
	doc.WriteString(fmt.Sprintf("- **Date**: %s\n\n", time.Now().Format("2006-01-02")))
	doc.WriteString("## Description\n\n")
	doc.WriteString(description + "\n\n")
	doc.WriteString("## Steps to Reproduce\n\n")
	doc.WriteString(reproSteps + "\n\n")

	if pocScript != "" {
		pocFileName := fmt.Sprintf("poc_%s.py", vulnSlug)
		pocFilePath := filepath.Join(pocsDir, pocFileName)
		_ = os.WriteFile(pocFilePath, []byte(pocScript), 0644)
		doc.WriteString("## Proof of Concept\n\n")
		doc.WriteString(fmt.Sprintf("Standalone script: [`pocs/%s`](../pocs/%s)\n\n", pocFileName, pocFileName))
	}

	doc.WriteString("## Remediation\n\n")
	doc.WriteString(remediation + "\n")

	if err := os.WriteFile(findingFilePath, []byte(doc.String()), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write finding file: %v", err)), nil
	}

	// Auto-aggregate markdown, metadata, and HTML dashboard (zero PDF overhead)
	_, _ = report.AggregateTarget(targetDir)

	return mcp.NewToolResultText(fmt.Sprintf("Finding successfully recorded and aggregated:\n- **File**: `reports/%s/findings/%s`\n- **Target**: `%s`\n- **Severity**: `%s`\n- **Summary**: `reports/%s/SUMMARY.md`\n- **Dashboard**: `reports/%s/report.html`\n- **PDF Deliverable**: Call cybermes_generate_pdf when ready",
		targetSlug, findingFileName, targetSlug, strings.ToUpper(severity), targetSlug, targetSlug)), nil
}

func sanitizeSlug(input string) string {
	lower := strings.ToLower(input)
	clean := slugRegex.ReplaceAllString(lower, "_")
	return strings.Trim(clean, "_")
}
