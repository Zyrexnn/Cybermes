package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cybermes/pkg/fuzz"
	"cybermes/pkg/scope"
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerFuzzTools() {
	fuzzTool := mcp.NewTool(
		"cybermes_fuzz_endpoints",
		mcp.WithDescription("Fuzz web application routes, hidden endpoints, and API paths using ffuf or native concurrent Go worker pool with rate limiting and Scope Guard enforcement."),
		mcp.WithString(
			"target_url",
			mcp.Required(),
			mcp.Description("Target base URL to fuzz (e.g. 'https://api.example.com' or 'http://127.0.0.1:8888')."),
		),
		mcp.WithString(
			"target_slug",
			mcp.Description("Target slug for logging and scope enforcement (optional)."),
		),
		mcp.WithString(
			"wordlist",
			mcp.Description("Custom wordlist path or built-in filename in tools/wordlists/ (e.g. 'common.txt', 'api-endpoints.txt'). Defaults to common.txt."),
		),
		mcp.WithString(
			"extensions",
			mcp.Description("Comma-separated file extensions to append to wordlist items (e.g. '.json,.php,.env,.bak')."),
		),
		mcp.WithString(
			"status_codes",
			mcp.Description("Comma-separated HTTP status codes to match (default: '200,204,301,302,307,401,403,405')."),
			mcp.DefaultString("200,204,301,302,307,401,403,405"),
		),
		mcp.WithString(
			"headers",
			mcp.Description("Custom HTTP headers as JSON string (e.g. '{\"Authorization\": \"Bearer token\"}') or newline-separated 'Key: Value' pairs."),
		),
		mcp.WithString(
			"cookies",
			mcp.Description("Session cookie string to include in fuzzing requests (e.g. 'session=abc123')."),
		),
		mcp.WithNumber(
			"rate_limit",
			mcp.Description("Maximum requests per second (safe default: 10, max: 25)."),
			mcp.DefaultNumber(10),
		),
		mcp.WithNumber(
			"timeout_seconds",
			mcp.Description("Fuzz execution timeout in seconds (default: 30)."),
			mcp.DefaultNumber(30),
		),
		mcp.WithBoolean(
			"prefer_ffuf",
			mcp.Description("Attempt to use external ffuf binary before falling back to native Go concurrent fuzzer (default: true)."),
		),
		mcp.WithString(
			"format",
			mcp.Description("Output format: 'markdown' (default summary table) or 'json'."),
			mcp.Enum("markdown", "json"),
			mcp.DefaultString("markdown"),
		),
	)

	s.mcpServer.AddTool(fuzzTool, s.handleFuzzEndpoints)
}

func (s *Server) handleFuzzEndpoints(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL := strings.TrimSpace(request.GetString("target_url", ""))
	if targetURL == "" {
		return mcp.NewToolResultError("target_url parameter is required"), nil
	}

	targetSlug := strings.TrimSpace(request.GetString("target_slug", ""))
	if targetSlug == "" {
		targetSlug = sanitizeSlug(targetURL)
	}

	wordlist := strings.TrimSpace(request.GetString("wordlist", ""))
	extensionsRaw := strings.TrimSpace(request.GetString("extensions", ""))
	statusCodesRaw := strings.TrimSpace(request.GetString("status_codes", "200,204,301,302,307,401,403,405"))
	headersRaw := request.GetString("headers", "")
	cookies := strings.TrimSpace(request.GetString("cookies", ""))
	rateLimit := request.GetInt("rate_limit", 10)
	timeoutSec := request.GetInt("timeout_seconds", 30)
	preferFfuf := request.GetBool("prefer_ffuf", true)
	format := request.GetString("format", "markdown")

	// Scope Guard check
	cfg, _, err := scope.FindScopeConfig(s.cfg.RootDir, targetSlug)
	if err == nil && cfg != nil {
		val := scope.ValidateTarget(targetURL, cfg)
		if !val.Allowed {
			return mcp.NewToolResultError(fmt.Sprintf("Scope Guard Violation: %s", val.Reason)), nil
		}
	}

	var extensions []string
	if extensionsRaw != "" {
		for _, ext := range strings.Split(extensionsRaw, ",") {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				extensions = append(extensions, ext)
			}
		}
	}

	var statusCodes []int
	if statusCodesRaw != "" {
		for _, scStr := range strings.Split(statusCodesRaw, ",") {
			if sc, err := strconv.Atoi(strings.TrimSpace(scStr)); err == nil && sc > 0 {
				statusCodes = append(statusCodes, sc)
			}
		}
	}

	opts := fuzz.FuzzOptions{
		TargetURL:    targetURL,
		WordlistPath: wordlist,
		Extensions:   extensions,
		StatusCodes:  statusCodes,
		RateLimit:    rateLimit,
		Timeout:      time.Duration(timeoutSec) * time.Second,
		Headers:      parseCustomHeaders(headersRaw),
		Cookies:      cookies,
		ToolsDir:     s.cfg.ToolsDir,
		PreferFfuf:   preferFfuf,
		MaxResults:   30,
	}

	res, err := fuzz.FuzzEndpoints(ctx, opts)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Endpoint fuzzing failed: %v", err)), nil
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format JSON: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# ⚡ Endpoint Fuzzing Results: `%s`\n\n", res.TargetURL))
	sb.WriteString(fmt.Sprintf("- **Total Endpoints Probed**: `%d`\n", res.TotalProbed))
	sb.WriteString(fmt.Sprintf("- **Matched Endpoints**: `%d`\n", res.TotalMatched))
	sb.WriteString(fmt.Sprintf("- **Engine Used**: `%s` (Completed in %d ms)\n\n", res.EngineUsed, res.DurationMs))

	if len(res.Matches) == 0 {
		sb.WriteString("ℹ️ No matched endpoints found matching criteria.\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	sb.WriteString("| Status Code | Path | Content Length | Response Time | Redirect Location |\n")
	sb.WriteString("| :---: | :--- | :---: | :---: | :--- |\n")

	for _, m := range res.Matches {
		redirectStr := "-"
		if m.RedirectURL != "" {
			redirectStr = fmt.Sprintf("`%s`", m.RedirectURL)
		}
		sb.WriteString(fmt.Sprintf("| `%d` | `/%s` | `%d B` | `%d ms` | %s |\n",
			m.StatusCode, strings.TrimPrefix(m.Path, "/"), m.ContentLength, m.ResponseTimeMs, redirectStr))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
