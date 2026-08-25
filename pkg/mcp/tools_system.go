package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type ToolStatus struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "Core" or "Optional (On-Demand)"
	Installed   bool   `json:"installed"`
	Path        string `json:"path,omitempty"`
	Description string `json:"description"`
}

type EnvironmentReport struct {
	OS              string       `json:"os"`
	Architecture    string       `json:"architecture"`
	GoVersion       string       `json:"go_version,omitempty"`
	WorkspaceRoot   string       `json:"workspace_root"`
	PackageManagers []string     `json:"package_managers"`
	Tools           []ToolStatus `json:"tools"`
}

func (s *Server) registerSystemTools() {
	// 1. cybermes_check_environment
	checkEnvTool := mcp.NewTool(
		"cybermes_check_environment",
		mcp.WithDescription("Inspect the local system environment, package managers, core toolchain readiness, and optional on-demand tools (like Nuclei) without modifying any files."),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary table) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	// 2. cybermes_record_evidence
	recordEvidenceTool := mcp.NewTool(
		"cybermes_record_evidence",
		mcp.WithDescription("Append non-vulnerability observations, missing security headers, reconnaissance notes, and negative test evidence directly to reports/<target_slug>/evidence/recon_notes.md according to Cybermes AGENTS.md directives."),
		mcp.WithString(
			"target_slug",
			mcp.Required(),
			mcp.Description("Target slug directory name (e.g. 'example_com')."),
		),
		mcp.WithString(
			"category",
			mcp.Required(),
			mcp.Description("Evidence category (e.g. 'Negative Tests', 'Missing Headers', 'Tech Stack Discovery', 'WAF Observation', 'Information Disclosure')."),
		),
		mcp.WithString(
			"note",
			mcp.Required(),
			mcp.Description("Detailed observation, tested parameter, response status code, or negative test rationale."),
		),
	)

	s.mcpServer.AddTool(checkEnvTool, s.handleCheckEnvironment)
	s.mcpServer.AddTool(recordEvidenceTool, s.handleRecordEvidence)
}

func (s *Server) handleCheckEnvironment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := request.GetString("format", "markdown")

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	report := EnvironmentReport{
		OS:            fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
		Architecture:  runtime.GOARCH,
		GoVersion:     runtime.Version(),
		WorkspaceRoot: s.cfg.RootDir,
	}

	// Detect Package Managers
	pms := []string{"winget", "brew", "apt", "pdtm", "go", "python", "pip"}
	for _, pm := range pms {
		if _, err := exec.LookPath(pm); err == nil {
			report.PackageManagers = append(report.PackageManagers, pm)
		}
	}

	// Tool definitions
	toolsToCheck := []struct {
		name     string
		toolType string
		desc     string
	}{
		{"search_knowledge", "Core", "Offline Sub-50ms Payload & Methodology Search"},
		{"secret_scan", "Core", "48-Pattern Deterministic Credential Detector"},
		{"smart_pipe", "Core", "Token Budgeting & Output Entropy Stream Filter"},
		{"aggregate_reports", "Core", "Executive Report & Finding Indexer"},
		{"httpx", "Core (Go fallback available)", "HTTP Prober & Tech Fingerprinter"},
		{"katana", "Core (Go fallback available)", "SPA & Endpoint Crawler"},
		{"subfinder", "Core", "Passive Subdomain Discovery"},
		{"nuclei", "Optional (On-Demand)", "Vulnerability Template & CVE Scanner (~150MB)"},
		{"sqlmap", "Optional (On-Demand)", "Automated SQL Injection Auditor"},
	}

	for _, t := range toolsToCheck {
		st := ToolStatus{
			Name:        t.name,
			Type:        t.toolType,
			Description: t.desc,
			Installed:   false,
		}

		// Check local tools/bin
		localCandidate := filepath.Join(s.cfg.ToolsDir, "bin", t.name+ext)
		if _, err := os.Stat(localCandidate); err == nil {
			st.Installed = true
			st.Path = localCandidate
		} else if p, err := exec.LookPath(t.name + ext); err == nil {
			st.Installed = true
			st.Path = p
		} else if p, err := exec.LookPath(t.name); err == nil {
			st.Installed = true
			st.Path = p
		}

		report.Tools = append(report.Tools, st)
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to encode JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString("### 🖥️ Cybermes Environment & Toolchain Diagnostic\n\n")
	sb.WriteString(fmt.Sprintf("- **OS & Architecture**: `%s`\n", report.OS))
	sb.WriteString(fmt.Sprintf("- **Go Runtime**: `%s`\n", report.GoVersion))
	sb.WriteString(fmt.Sprintf("- **Workspace Root**: `%s`\n", report.WorkspaceRoot))
	if len(report.PackageManagers) > 0 {
		sb.WriteString(fmt.Sprintf("- **Detected Package Managers**: `%s`\n", strings.Join(report.PackageManagers, ", ")))
	}
	sb.WriteString("\n#### 🧰 Toolchain Status Matrix\n\n")
	sb.WriteString("| Tool Name | Type | Status | Location / Notes |\n")
	sb.WriteString("| :--- | :--- | :---: | :--- |\n")

	for _, st := range report.Tools {
		badge := "🟢 **READY**"
		loc := fmt.Sprintf("`%s`", st.Path)
		if !st.Installed {
			if strings.Contains(st.Type, "Optional") {
				badge = "⚪ **OPTIONAL (NOT INSTALLED)**"
				loc = fmt.Sprintf("*Available on-demand: %s*", getOnDemandInstallInstructions())
			} else {
				badge = "🟡 **FALLBACK ACTIVE**"
				loc = "*Using native Go internal engine*"
			}
		}
		sb.WriteString(fmt.Sprintf("| **`%s`** | %s | %s | %s |\n",
			st.Name, st.Type, badge, loc))
	}

	sb.WriteString("\n💡 *Tip: Core tools run 100% standalone. Optional tools like Nuclei can be installed on-demand only when needed.*")
	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleRecordEvidence(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetSlug, err := request.RequireString("target_slug")
	if err != nil || strings.TrimSpace(targetSlug) == "" {
		return mcp.NewToolResultError("Missing required parameter 'target_slug'"), nil
	}
	category, err := request.RequireString("category")
	if err != nil || strings.TrimSpace(category) == "" {
		return mcp.NewToolResultError("Missing required parameter 'category'"), nil
	}
	note, err := request.RequireString("note")
	if err != nil || strings.TrimSpace(note) == "" {
		return mcp.NewToolResultError("Missing required parameter 'note'"), nil
	}

	targetSlug = sanitizeSlug(targetSlug)
	evidenceDir := filepath.Join(s.cfg.ReportsDir, targetSlug, "evidence")
	if err := os.MkdirAll(evidenceDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create evidence directory: %v", err)), nil
	}

	reconNotesPath := filepath.Join(evidenceDir, "recon_notes.md")

	// Initialize file with header if not existing
	if _, err := os.Stat(reconNotesPath); os.IsNotExist(err) {
		initialHeader := fmt.Sprintf("# 📝 Reconnaissance Notes & Evidence: `%s`\n\n"+
			"*This file aggregates all non-critical reconnaissance findings, negative authorization tests, missing security headers, and tech observations per Cybermes AGENTS.md directives.*\n\n"+
			"---\n\n", targetSlug)
		_ = os.WriteFile(reconNotesPath, []byte(initialHeader), 0644)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("### [%s] %s\n- **Timestamp**: `%s`\n- **Details**:\n%s\n\n",
		strings.ToUpper(strings.TrimSpace(category)), strings.TrimSpace(category), timestamp, strings.TrimSpace(note))

	f, err := os.OpenFile(reconNotesPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open recon_notes.md: %v", err)), nil
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write entry: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Evidence logged successfully to `reports/%s/evidence/recon_notes.md` under category `[%s]`.", targetSlug, category)), nil
}
