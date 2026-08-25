package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func setupTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	rootDir := FindProjectRoot("")
	if rootDir == "" {
		t.Fatal("Could not locate project root directory")
	}

	cfg := Config{
		RootDir: rootDir,
		Version: "2.1.0-test",
	}

	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}

	return srv, rootDir
}

func TestNewServer(t *testing.T) {
	srv, _ := setupTestServer(t)

	if srv.MCPServer() == nil {
		t.Fatal("MCPServer instance should not be nil")
	}

	tools := srv.MCPServer().ListTools()
	expectedTools := []string{
		"cybermes_search_knowledge",
		"cybermes_list_skills",
		"cybermes_get_skill",
		"cybermes_scan_secrets",
		"cybermes_aggregate_report",
		"cybermes_list_findings",
		"cybermes_record_finding",
		"cybermes_validate_scope",
		"cybermes_http_probe",
		"cybermes_recon_crawl",
	}

	for _, toolName := range expectedTools {
		if _, ok := tools[toolName]; !ok {
			t.Errorf("Expected tool %s to be registered", toolName)
		}
	}

	prompts := srv.MCPServer().ListPrompts()
	if _, ok := prompts["cybermes_hunt"]; !ok {
		t.Error("Expected prompt cybermes_hunt to be registered")
	}
	if _, ok := prompts["cybermes_triage"]; !ok {
		t.Error("Expected prompt cybermes_triage to be registered")
	}
}

func TestToolSearchKnowledge(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_search_knowledge",
			Arguments: map[string]any{
				"query":  "jwt bypass",
				"source": "all",
				"limit":  float64(2),
				"format": "markdown",
			},
		},
	}

	res, err := srv.handleSearchKnowledge(ctx, req)
	if err != nil {
		t.Fatalf("handleSearchKnowledge returned unexpected error: %v", err)
	}
	if res == nil || len(res.Content) == 0 {
		t.Fatal("handleSearchKnowledge returned empty result")
	}

	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}

	if !strings.Contains(text.Text, "Cybermes Knowledge Base") {
		t.Errorf("Expected output to mention Cybermes Knowledge Base, got: %s", text.Text)
	}
}

func TestToolSkills(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// 1. List skills
	listReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_list_skills",
			Arguments: map[string]any{
				"filter": "idor",
				"limit":  float64(10),
			},
		},
	}

	listRes, err := srv.handleListSkills(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListSkills error: %v", err)
	}

	listText, ok := mcp.AsTextContent(listRes.Content[0])
	if !ok || !strings.Contains(listText.Text, "hunt-idor") {
		t.Errorf("Expected skill listing to contain 'hunt-idor', got: %s", listText.Text)
	}

	// 2. Get specific skill
	getReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_get_skill",
			Arguments: map[string]any{
				"skill_name": "hunt-idor",
			},
		},
	}

	getRes, err := srv.handleGetSkill(ctx, getReq)
	if err != nil {
		t.Fatalf("handleGetSkill error: %v", err)
	}
	getText, ok := mcp.AsTextContent(getRes.Content[0])
	if !ok || !strings.Contains(getText.Text, "hunt-idor") {
		t.Errorf("Expected skill content to contain 'hunt-idor', got: %s", getText.Text)
	}
}

func TestToolScanSecrets(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	mockSnippet := `
		// Database config
		const awsKey = "AKIAIOSFODNN7EXAMPLE";
		const ghToken = "ghp_111122223333444455556666777788889999";
	`

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_scan_secrets",
			Arguments: map[string]any{
				"content":      mockSnippet,
				"mask_secrets": true,
				"format":       "json",
			},
		},
	}

	res, err := srv.handleScanSecrets(ctx, req)
	if err != nil {
		t.Fatalf("handleScanSecrets error: %v", err)
	}

	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok || !strings.Contains(text.Text, "AWS_ACCESS_KEY") || !strings.Contains(text.Text, "GH_PAT_CLASSIC") {
		t.Errorf("Expected secret scan output to contain AWS_ACCESS_KEY and GH_PAT_CLASSIC, got: %s", text.Text)
	}
}

func TestToolReportAndRecord(t *testing.T) {
	srv, rootDir := setupTestServer(t)
	ctx := context.Background()

	testSlug := "test_mcp_validation"
	targetDir := filepath.Join(rootDir, "reports", testSlug)
	defer os.RemoveAll(targetDir)

	// Record a test finding
	recordReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_record_finding",
			Arguments: map[string]any{
				"target_slug":        testSlug,
				"severity":           "high",
				"title":              "Test IDOR Invoices",
				"endpoint":           "GET /api/v1/invoices/99",
				"description":        "Unit test description for IDOR vulnerability.",
				"reproduction_steps": "1. Send request with user B cookie\n2. Observe 200 OK",
				"poc_script":         "import requests\nprint('poc')",
				"remediation":        "Validate object ownership.",
			},
		},
	}

	recRes, err := srv.handleRecordFinding(ctx, recordReq)
	if err != nil {
		t.Fatalf("handleRecordFinding error: %v", err)
	}
	recText, _ := mcp.AsTextContent(recRes.Content[0])
	if !strings.Contains(recText.Text, "high_test_idor_invoices.md") {
		t.Errorf("Expected recorded finding filename, got: %s", recText.Text)
	}

	// Verify Aggregate Report
	aggReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_aggregate_report",
			Arguments: map[string]any{
				"target_slug": testSlug,
				"format":      "markdown",
			},
		},
	}

	aggRes, err := srv.handleAggregateReport(ctx, aggReq)
	if err != nil {
		t.Fatalf("handleAggregateReport error: %v", err)
	}
	aggText, _ := mcp.AsTextContent(aggRes.Content[0])
	if !strings.Contains(aggText.Text, testSlug) {
		t.Errorf("Expected aggregate report to contain target slug, got: %s", aggText.Text)
	}
}

func TestResourceAndPrompts(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// 1. Read skill resource
	resReq := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "skills://hunt-idor",
		},
	}

	resContents, err := srv.handleReadSkillResource(ctx, resReq)
	if err != nil {
		t.Fatalf("handleReadSkillResource error: %v", err)
	}
	if len(resContents) == 0 {
		t.Fatal("Expected non-empty resource contents")
	}

	// 2. Hunt prompt
	promptReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_hunt",
			Arguments: map[string]string{
				"target":     "api.target.com",
				"focus_area": "idor",
			},
		},
	}

	promptRes, err := srv.handleHuntPrompt(ctx, promptReq)
	if err != nil {
		t.Fatalf("handleHuntPrompt error: %v", err)
	}
	if len(promptRes.Messages) == 0 {
		t.Fatal("Expected messages in prompt result")
	}
}

func TestToolValidateScope(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_validate_scope",
			Arguments: map[string]any{
				"target": "https://api.example.com",
				"format": "markdown",
			},
		},
	}

	res, err := srv.handleValidateScope(ctx, req)
	if err != nil {
		t.Fatalf("handleValidateScope error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Scope Guard") {
		t.Errorf("Expected Scope Guard output, got: %s", text.Text)
	}
}

func TestToolHttpProbe(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx/1.22.0")
		w.Header().Set("X-Powered-By", "Next.js")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>Test App</title></head><body><div id="__NEXT_DATA__"></div></body></html>`))
	}))
	defer mockServer.Close()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_http_probe",
			Arguments: map[string]any{
				"target_url":   mockServer.URL,
				"prefer_httpx": false,
				"format":       "markdown",
			},
		},
	}

	res, err := srv.handleHttpProbe(ctx, req)
	if err != nil {
		t.Fatalf("handleHttpProbe error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "200 OK") || !strings.Contains(text.Text, "Test App") {
		t.Errorf("Expected probe details in output, got: %s", text.Text)
	}
}

func TestToolReconCrawl(t *testing.T) {
	srv, rootDir := setupTestServer(t)
	ctx := context.Background()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><body><a href="/api/v1/auth">Auth</a><a href="/dashboard">Dash</a></body></html>`))
	}))
	defer mockServer.Close()

	testSlug := "mcp_test_crawl"
	defer func() {
		_ = os.RemoveAll(filepath.Join(rootDir, "recon", testSlug))
	}()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_recon_crawl",
			Arguments: map[string]any{
				"target_url":    mockServer.URL,
				"target_slug":   testSlug,
				"prefer_katana": false,
				"format":        "markdown",
			},
		},
	}

	res, err := srv.handleReconCrawl(ctx, req)
	if err != nil {
		t.Fatalf("handleReconCrawl error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Recon Crawl Discovery") || !strings.Contains(text.Text, "/api/v1/auth") {
		t.Errorf("Expected crawl summary in output, got: %s", text.Text)
	}
}

