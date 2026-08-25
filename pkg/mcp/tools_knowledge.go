package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerKnowledgeTools() {
	tool := mcp.NewTool(
		"cybermes_search_knowledge",
		mcp.WithDescription("Search the Cybermes curated offensive security knowledge base (PayloadsAllTheThings, HackTricks, Strix, Claude-BugHunter) with sub-50ms deterministic ranking for exploit payloads, bypasses, and methodology snippets."),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("The security search query, attack technique, parameter type, or vulnerability pattern (e.g. 'jwt none algorithm bypass', 'idor tenant isolation', 'sqli time based blind')."),
		),
		mcp.WithString(
			"source",
			mcp.Description("Knowledge base source filter. Options: 'all' (default), 'payloads' (PayloadsAllTheThings), 'hacktricks', 'claude' (Claude-BugHunter), 'strix', 'hack'."),
			mcp.Enum("all", "payloads", "hacktricks", "claude", "strix", "hack"),
			mcp.DefaultString("all"),
		),
		mcp.WithInteger(
			"limit",
			mcp.Description("Maximum number of top-ranked snippets to return (default: 3, max: 10)."),
			mcp.DefaultNumber(3),
		),
		mcp.WithInteger(
			"max_len",
			mcp.Description("Maximum character length per snippet before safe chunking (default: 1400)."),
			mcp.DefaultNumber(1400),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output formatting. Options: 'markdown' (default readable block) or 'json' (structured data)."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSearchKnowledge)
}

func (s *Server) handleSearchKnowledge(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil || strings.TrimSpace(query) == "" {
		return mcp.NewToolResultError("Missing required parameter: 'query'"), nil
	}

	source := request.GetString("source", "all")
	limit := request.GetInt("limit", 3)
	if limit <= 0 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}

	maxLen := request.GetInt("max_len", 1400)
	if maxLen <= 0 {
		maxLen = 1400
	}

	format := request.GetString("format", "markdown")

	results, err := s.searcher.Search(query, source, limit, maxLen)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Knowledge search error: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("🔍 No high-signal knowledge snippets found for query: '%s' (Source: %s).", query, source)), nil
	}

	if format == "json" {
		out := map[string]any{
			"query":         query,
			"source":        source,
			"total_results": len(results),
			"snippets":      results,
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("JSON marshal error: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	// Format as high-signal Markdown
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 📚 Cybermes Knowledge Base: `%s` (Found %d snippets)\n\n", query, len(results)))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("#### Snippet #%d | Score: %d | Source: `[%s]`\n", i+1, r.Score, r.SourceKB))
		sb.WriteString(fmt.Sprintf("- **Location**: `%s:%d`\n", r.File, r.StartLine))
		if r.Heading != "" {
			sb.WriteString(fmt.Sprintf("- **Section**: %s\n\n", r.Heading))
		} else {
			sb.WriteString("\n")
		}
		sb.WriteString("```markdown\n")
		sb.WriteString(strings.TrimSpace(r.Content))
		sb.WriteString("\n```\n\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
