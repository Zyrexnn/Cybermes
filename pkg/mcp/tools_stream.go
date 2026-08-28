package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"cybermes/pkg/stream"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerStreamTools() {
	filterTool := mcp.NewTool(
		"cybermes_filter_stream",
		mcp.WithDescription("Filter raw command dumps, large crawling logs, or HTTP responses using Cybermes Smart Pipe entropy and keyword scoring to extract top high-signal lines and preserve token budget."),
		mcp.WithString(
			"content",
			mcp.Required(),
			mcp.Description("Raw text, tool dump, log stream, or HTTP output to filter."),
		),
		mcp.WithNumber(
			"limit",
			mcp.Description("Maximum top high-signal lines to return (default: 25, max: 100)."),
			mcp.DefaultNumber(25),
		),
		mcp.WithNumber(
			"min_score",
			mcp.Description("Minimum entropy/priority score threshold (default: 10)."),
			mcp.DefaultNumber(10),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(filterTool, s.handleFilterStream)
}

func (s *Server) handleFilterStream(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := request.RequireString("content")
	if err != nil || strings.TrimSpace(content) == "" {
		return mcp.NewToolResultError("content parameter is required and cannot be empty"), nil
	}

	limit := request.GetInt("limit", 25)
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	minScore := request.GetInt("min_score", 10)
	format := request.GetString("format", "markdown")

	scanner := bufio.NewScanner(strings.NewReader(content))
	var totalLines int
	var scoredList []stream.ScoredLine
	seen := make(map[string]bool)

	for scanner.Scan() {
		line := stream.CleanLine(scanner.Text())
		if line == "" {
			continue
		}
		totalLines++

		if !seen[line] {
			seen[line] = true
			score := stream.ScoreLine(line)
			if score >= minScore {
				scoredList = append(scoredList, stream.ScoredLine{
					Score: score,
					Text:  line,
				})
			}
		}
	}

	sort.SliceStable(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	displayCount := limit
	if displayCount > len(scoredList) {
		displayCount = len(scoredList)
	}

	topList := scoredList[:displayCount]

	if strings.ToLower(format) == "json" {
		out := map[string]any{
			"total_raw_lines": totalLines,
			"unique_scored":   len(scoredList),
			"returned_count":  len(topList),
			"top_entries":     topList,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### 📊 Cybermes Smart Stream Filter\n\n"))
	sb.WriteString(fmt.Sprintf("- **Total Input Lines**: `%d`\n", totalLines))
	sb.WriteString(fmt.Sprintf("- **Prioritized High-Signal Entries**: `%d` (from `%d` unique candidates)\n\n", len(topList), len(scoredList)))

	if len(topList) == 0 {
		sb.WriteString("ℹ️ No entries exceeded the minimum score threshold.\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	sb.WriteString("| Score | High-Signal Entry |\n| :---: | :--- |\n")
	for _, item := range topList {
		sb.WriteString(fmt.Sprintf("| `%d` | `%s` |\n", item.Score, item.Text))
	}

	if len(scoredList) > displayCount {
		sb.WriteString(fmt.Sprintf("\n> 💡 *Note: %d lower-priority entries filtered out to protect LLM context window.*", len(scoredList)-displayCount))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
