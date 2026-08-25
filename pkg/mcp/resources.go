package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
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

	s.mcpServer.AddResourceTemplate(skillTemplate, s.handleReadSkillResource)
	s.mcpServer.AddResourceTemplate(reportTemplate, s.handleReadReportResource)
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
