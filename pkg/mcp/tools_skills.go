package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// SkillMetadata holds summary information parsed from a skill's SKILL.md.
type SkillMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Sources     string `json:"sources,omitempty"`
	ReportCount int    `json:"report_count,omitempty"`
	Path        string `json:"path"`
}

var (
	skillsIndexCache []SkillMetadata
	skillsIndexMu    sync.RWMutex
)

// ParseSkillMetadata extracts frontmatter or fallback info from a SKILL.md file.
func ParseSkillMetadata(filePath string) (SkillMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return SkillMetadata{}, err
	}
	defer file.Close()

	meta := SkillMetadata{
		Name: filepath.Base(filepath.Dir(filePath)),
		Path: filePath,
	}

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterChecked := false
	var bodySnippet strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !frontmatterChecked {
			if line == "---" {
				inFrontmatter = true
				frontmatterChecked = true
				continue
			} else if line != "" {
				frontmatterChecked = true
			}
		}

		if inFrontmatter {
			if line == "---" {
				inFrontmatter = false
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)

				switch strings.ToLower(k) {
				case "name":
					meta.Name = v
				case "description":
					meta.Description = v
				case "sources":
					meta.Sources = v
				case "report_count":
					if count, err := strconv.Atoi(v); err == nil {
						meta.ReportCount = count
					}
				}
			}
		} else {
			if meta.Description == "" && line != "" && !strings.HasPrefix(line, "#") {
				if bodySnippet.Len() < 200 {
					bodySnippet.WriteString(line + " ")
				}
			}
		}
	}

	if meta.Description == "" && bodySnippet.Len() > 0 {
		meta.Description = strings.TrimSpace(bodySnippet.String())
	}

	return meta, nil
}

// GetSkillsIndex scans the skills directory and caches all available playbooks.
func (s *Server) GetSkillsIndex(forceRefresh bool) ([]SkillMetadata, error) {
	skillsIndexMu.RLock()
	if !forceRefresh && len(skillsIndexCache) > 0 {
		defer skillsIndexMu.RUnlock()
		return skillsIndexCache, nil
	}
	skillsIndexMu.RUnlock()

	skillsIndexMu.Lock()
	defer skillsIndexMu.Unlock()

	if !forceRefresh && len(skillsIndexCache) > 0 {
		return skillsIndexCache, nil
	}

	var results []SkillMetadata

	entries, err := os.ReadDir(s.cfg.SkillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory %s: %w", s.cfg.SkillsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(s.cfg.SkillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillPath); err == nil {
			meta, err := ParseSkillMetadata(skillPath)
			if err == nil {
				results = append(results, meta)
			}
		}
	}

	skillsIndexCache = results
	return results, nil
}

func (s *Server) registerSkillsTools() {
	listTool := mcp.NewTool(
		"cybermes_list_skills",
		mcp.WithDescription("List and filter the 200+ specialized offensive security playbooks in Cybermes (covering IDOR, BOLA, SSRF, JWT, Cloud, Active Directory, Race Conditions, Mobile, Smart Contracts, etc.)."),
		mcp.WithString(
			"filter",
			mcp.Description("Optional keyword or vulnerability type to filter skills (e.g. 'idor', 'jwt', 'auth', 'cloud', 'race', 'api', 'nextjs')."),
		),
		mcp.WithInteger(
			"limit",
			mcp.Description("Maximum number of skills to list (default: 30, max: 200)."),
			mcp.DefaultNumber(30),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default readable table) or 'json' (structured array)."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	getTool := mcp.NewTool(
		"cybermes_get_skill",
		mcp.WithDescription("Retrieve the complete Markdown playbook and step-by-step SOP for a specific Cybermes security skill."),
		mcp.WithString(
			"skill_name",
			mcp.Required(),
			mcp.Description("Exact name or directory name of the skill (e.g. 'hunt-idor', 'jwt-oauth-token-attacks', '401-403-bypass-techniques', 'hunt-ssrf', 'api-recon-and-docs')."),
		),
		mcp.WithString(
			"section",
			mcp.Description("Optional heading filter to extract only a specific section (e.g. 'Crown Jewel Targets', 'Attack Vectors', 'Checklist')."),
		),
	)

	s.mcpServer.AddTool(listTool, s.handleListSkills)
	s.mcpServer.AddTool(getTool, s.handleGetSkill)
}

func (s *Server) handleListSkills(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filter := strings.ToLower(strings.TrimSpace(request.GetString("filter", "")))
	limit := request.GetInt("limit", 30)
	if limit <= 0 {
		limit = 30
	}
	if limit > 200 {
		limit = 200
	}
	format := request.GetString("format", "markdown")

	skills, err := s.GetSkillsIndex(false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load skills: %v", err)), nil
	}

	var matched []SkillMetadata
	for _, sk := range skills {
		if filter == "" ||
			strings.Contains(strings.ToLower(sk.Name), filter) ||
			strings.Contains(strings.ToLower(sk.Description), filter) ||
			strings.Contains(strings.ToLower(sk.Sources), filter) {
			matched = append(matched, sk)
			if len(matched) >= limit {
				break
			}
		}
	}

	if len(matched) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("🔍 No skills found matching filter '%s'. Total available skills: %d.", filter, len(skills))), nil
	}

	if format == "json" {
		out := map[string]any{
			"filter":        filter,
			"total_skills":  len(skills),
			"matched_count": len(matched),
			"skills":        matched,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 🛡️ Cybermes Offensive Skills Library (Showing %d of %d)\n\n", len(matched), len(skills)))
	sb.WriteString("| Skill Name | Reports / Evidence | Description |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	for _, m := range matched {
		repInfo := "-"
		if m.ReportCount > 0 {
			repInfo = fmt.Sprintf("📊 %d BB reports", m.ReportCount)
		}
		desc := m.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		desc = strings.ReplaceAll(desc, "|", "\\|")
		sb.WriteString(fmt.Sprintf("| **`%s`** | %s | %s |\n", m.Name, repInfo, desc))
	}

	sb.WriteString("\n💡 *Tip: Call `cybermes_get_skill(skill_name=\"<name>\")` or read `skills://<name>` to view the full playbook.*")
	return mcp.NewToolResultText(sb.String()), nil
}

func (s *Server) handleGetSkill(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	skillName, err := request.RequireString("skill_name")
	if err != nil || strings.TrimSpace(skillName) == "" {
		return mcp.NewToolResultError("Missing required parameter: 'skill_name'"), nil
	}

	skillName = strings.TrimSpace(skillName)
	sectionFilter := strings.ToLower(strings.TrimSpace(request.GetString("section", "")))

	skillPath := filepath.Join(s.cfg.SkillsDir, skillName, "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		// Try matching skill without strict case or prefix
		skills, _ := s.GetSkillsIndex(false)
		found := false
		for _, sk := range skills {
			if strings.EqualFold(sk.Name, skillName) {
				skillPath = sk.Path
				found = true
				break
			}
		}
		if !found {
			return mcp.NewToolResultError(fmt.Sprintf("Skill '%s' not found. Use `cybermes_list_skills` to search available playbooks.", skillName)), nil
		}
	}

	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read skill file: %v", err)), nil
	}

	fullContent := string(contentBytes)

	if sectionFilter != "" {
		extracted := extractMarkdownSection(fullContent, sectionFilter)
		if extracted != "" {
			return mcp.NewToolResultText(fmt.Sprintf("### 📖 Skill: `%s` > `%s`\n\n%s", skillName, sectionFilter, extracted)), nil
		}
	}

	return mcp.NewToolResultText(fullContent), nil
}

func extractMarkdownSection(content, sectionQuery string) string {
	lines := strings.Split(content, "\n")
	var sectionContent strings.Builder
	capturing := false
	captureLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, r := range trimmed {
				if r == '#' {
					level++
				} else {
					break
				}
			}
			headingTitle := strings.TrimSpace(trimmed[level:])

			if capturing {
				if level <= captureLevel {
					break
				}
			} else if strings.Contains(strings.ToLower(headingTitle), sectionQuery) {
				capturing = true
				captureLevel = level
			}
		}

		if capturing {
			sectionContent.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(sectionContent.String())
}
