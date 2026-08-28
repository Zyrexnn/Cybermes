package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cybermes/pkg/report"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	// Static Browseable Resources
	skillsIndexRes := mcp.NewResource(
		"skills://index",
		"Cybermes Skills Library Index",
		mcp.WithResourceDescription("Browseable catalog of 200+ specialized offensive security playbooks, vulnerability categories, and bug bounty methodologies."),
		mcp.WithMIMEType("text/markdown"),
	)

	reportsIndexRes := mcp.NewResource(
		"reports://index",
		"Target Engagement Index",
		mcp.WithResourceDescription("Executive overview matrix of all targets, confirmed vulnerability counts, and PoC status across the workspace."),
		mcp.WithMIMEType("text/markdown"),
	)

	knowledgeCheatsheetsRes := mcp.NewResource(
		"knowledge://cheatsheets",
		"Curated Exploitation Cheatsheets",
		mcp.WithResourceDescription("Overview of offline payload repositories, bypass cheatsheets, and injection techniques (PayloadsAllTheThings, HackTricks, Strix, Claude-BugHunter)."),
		mcp.WithMIMEType("text/markdown"),
	)

	// Dynamic Resource Templates
	skillTemplate := mcp.NewResourceTemplate(
		"skills://{skill_name}",
		"Cybermes Offensive Skill Playbook",
		mcp.WithTemplateDescription("Read-only access to Cybermes vulnerability playbooks and testing methodologies (e.g. skills://hunt-idor, skills://jwt-oauth-token-attacks)."),
		mcp.WithTemplateMIMEType("text/markdown"),
	)

	reportTemplate := mcp.NewResourceTemplate(
		"reports://{target_slug}/summary",
		"Target Executive Summary Report",
		mcp.WithTemplateDescription("Read-only access to the executive SUMMARY.md report for a specific engagement target (e.g. reports://example_com/summary)."),
		mcp.WithTemplateMIMEType("text/markdown"),
	)

	s.mcpServer.AddResource(skillsIndexRes, s.handleReadSkillsIndexResource)
	s.mcpServer.AddResource(reportsIndexRes, s.handleReadReportsIndexResource)
	s.mcpServer.AddResource(knowledgeCheatsheetsRes, s.handleReadKnowledgeCheatsheetsResource)
	s.mcpServer.AddResourceTemplate(skillTemplate, s.handleReadSkillResource)
	s.mcpServer.AddResourceTemplate(reportTemplate, s.handleReadReportResource)
}

func (s *Server) handleReadSkillsIndexResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	skills, err := s.GetSkillsIndex(false)
	if err != nil {
		return nil, fmt.Errorf("failed to load skills index: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# 🛡️ Cybermes Offensive Skills Library Catalog\n\n")
	sb.WriteString(fmt.Sprintf("Total Indexed Playbooks: **%d playbooks**\n\n", len(skills)))
	sb.WriteString("| Skill Name | Reports / Evidence | Description |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	for _, sk := range skills {
		repInfo := "-"
		if sk.ReportCount > 0 {
			repInfo = fmt.Sprintf("📊 %d reports", sk.ReportCount)
		}
		desc := strings.ReplaceAll(sk.Description, "|", "\\|")
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		sb.WriteString(fmt.Sprintf("| [`%s`](skills://%s) | %s | %s |\n", sk.Name, sk.Name, repInfo, desc))
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "skills://index",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		},
	}, nil
}

func (s *Server) handleReadReportsIndexResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	results, err := report.AggregateAll(s.cfg.ReportsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate reports: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# 📊 Cybermes Target Engagements Overview\n\n")
	sb.WriteString(fmt.Sprintf("Active Engagement Targets: **%d targets**\n\n", len(results)))

	if len(results) == 0 {
		sb.WriteString("ℹ️ No target assessments recorded yet. Run `cybermes_record_finding` to begin logging findings.\n")
	} else {
		sb.WriteString("| Target Slug | Total Findings | Critical | High | Medium | Low | Summary Link |\n")
		sb.WriteString("| :--- | :---: | :---: | :---: | :---: | :---: | :--- |\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("| **`%s`** | %d | %d | %d | %d | %d | [`reports://%s/summary`](reports://%s/summary) |\n",
				r.Target, r.TotalFindings,
				r.SeveritySummary["CRITICAL"], r.SeveritySummary["HIGH"],
				r.SeveritySummary["MEDIUM"], r.SeveritySummary["LOW"],
				r.Target, r.Target))
		}
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "reports://index",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		},
	}, nil
}

func (s *Server) handleReadKnowledgeCheatsheetsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	var sb strings.Builder
	sb.WriteString("# 📚 Cybermes Curated Exploitation & Payload Knowledge Base\n\n")
	sb.WriteString("Cybermes indexes offline security playbooks and cheatsheets with sub-50ms deterministic query response.\n\n")
	sb.WriteString("## Indexed Knowledge Sources:\n\n")
	sb.WriteString("1. **PayloadsAllTheThings (`source: 'payloads'`)**: Comprehensive injection payloads, webshells, bypass strings, and methodology trees.\n")
	sb.WriteString("2. **HackTricks (`source: 'hacktricks'`)**: Attack surface playbooks for Web, Cloud (AWS/GCP/Azure), Networks, and Pentesting methodologies.\n")
	sb.WriteString("3. **Strix & Claude-BugHunter (`source: 'claude'`)**: Verified modern bug bounty recipes, AI router exploitation, API gateways, and token abuse.\n\n")
	sb.WriteString("### Quick Usage:\n")
	sb.WriteString("Search any payload with: `cybermes_search_knowledge(query=\"...\")`\n")

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "knowledge://cheatsheets",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		},
	}, nil
}

func (s *Server) handleReadSkillResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	prefix := "skills://"
	if !strings.HasPrefix(uri, prefix) {
		return nil, fmt.Errorf("invalid skill URI: %s", uri)
	}

	skillName := strings.TrimPrefix(uri, prefix)
	skillPath := filepath.Join(s.cfg.SkillsDir, skillName, "SKILL.md")

	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		// Fallback fuzzy search
		skills, _ := s.GetSkillsIndex(false)
		for _, sk := range skills {
			if strings.EqualFold(sk.Name, skillName) {
				skillPath = sk.Path
				break
			}
		}
	}

	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}

func (s *Server) handleReadReportResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	prefix := "reports://"
	suffix := "/summary"

	if !strings.HasPrefix(uri, prefix) || !strings.HasSuffix(uri, suffix) {
		return nil, fmt.Errorf("invalid report URI: %s (expected reports://<target_slug>/summary)", uri)
	}

	trimmed := strings.TrimPrefix(uri, prefix)
	targetSlug := strings.TrimSuffix(trimmed, suffix)
	summaryPath := filepath.Join(s.cfg.ReportsDir, targetSlug, "SUMMARY.md")

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil, fmt.Errorf("summary report not found for target: %s", targetSlug)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     string(data),
		},
	}, nil
}
