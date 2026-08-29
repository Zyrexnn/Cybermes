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
		"cybermes_generate_pdf",
		"cybermes_list_findings",
		"cybermes_record_finding",
		"cybermes_validate_scope",
		"cybermes_http_probe",
		"cybermes_recon_crawl",
		"cybermes_subdomain_discovery",
		"cybermes_fuzz_endpoints",
		"cybermes_filter_stream",
		"cybermes_nuclei_scan",
		"cybermes_check_environment",
		"cybermes_record_evidence",
	}

	for _, toolName := range expectedTools {
		if _, ok := tools[toolName]; !ok {
			t.Errorf("Expected tool %s to be registered", toolName)
		}
	}

	prompts := srv.MCPServer().ListPrompts()
	expectedPrompts := []string{
		"cybermes_hunt",
		"cybermes_triage",
		"cybermes_api_audit",
		"cybermes_idor_matrix",
		"cybermes_403_bypass",
		"cybermes_ai_prompt_injection_audit",
	}
	for _, p := range expectedPrompts {
		if _, ok := prompts[p]; !ok {
			t.Errorf("Expected prompt %s to be registered", p)
		}
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
				"target_slug":  testSlug,
				"format":       "markdown",
				"generate_pdf": false,
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

	// Verify Generate PDF tool
	pdfReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_generate_pdf",
			Arguments: map[string]any{
				"target_slug": testSlug,
				"format":      "markdown",
			},
		},
	}
	pdfRes, err := srv.handleGeneratePDF(ctx, pdfReq)
	if err != nil {
		t.Fatalf("handleGeneratePDF error: %v", err)
	}
	pdfText, _ := mcp.AsTextContent(pdfRes.Content[0])
	if !strings.Contains(pdfText.Text, testSlug) {
		t.Errorf("Expected PDF generation output to contain target slug, got: %s", pdfText.Text)
	}
}

func TestResourceAndPrompts(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// 1. Read skill resource template & static index
	resReq := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "skills://hunt-idor",
		},
	}
	resContents, err := srv.handleReadSkillResource(ctx, resReq)
	if err != nil || len(resContents) == 0 {
		t.Fatalf("handleReadSkillResource error: %v", err)
	}

	skillsIndexReq := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "skills://index",
		},
	}
	idxContents, err := srv.handleReadSkillsIndexResource(ctx, skillsIndexReq)
	if err != nil || len(idxContents) == 0 {
		t.Fatalf("handleReadSkillsIndexResource error: %v", err)
	}

	reportsIndexReq := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "reports://index",
		},
	}
	repContents, err := srv.handleReadReportsIndexResource(ctx, reportsIndexReq)
	if err != nil || len(repContents) == 0 {
		t.Fatalf("handleReadReportsIndexResource error: %v", err)
	}

	knowledgeResReq := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "knowledge://cheatsheets",
		},
	}
	knowContents, err := srv.handleReadKnowledgeCheatsheetsResource(ctx, knowledgeResReq)
	if err != nil || len(knowContents) == 0 {
		t.Fatalf("handleReadKnowledgeCheatsheetsResource error: %v", err)
	}

	// 2. Hunt prompt
	huntReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_hunt",
			Arguments: map[string]string{
				"target":     "api.target.com",
				"focus_area": "idor",
			},
		},
	}
	huntRes, err := srv.handleHuntPrompt(ctx, huntReq)
	if err != nil || len(huntRes.Messages) == 0 {
		t.Fatalf("handleHuntPrompt error: %v", err)
	}

	// 3. API Audit prompt
	apiAuditReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_api_audit",
			Arguments: map[string]string{
				"target_url": "https://api.target.com/v1",
				"api_type":   "rest",
			},
		},
	}
	apiRes, err := srv.handleApiAuditPrompt(ctx, apiAuditReq)
	if err != nil || len(apiRes.Messages) == 0 {
		t.Fatalf("handleApiAuditPrompt error: %v", err)
	}

	// 4. IDOR Matrix prompt
	idorReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_idor_matrix",
			Arguments: map[string]string{
				"target_url": "https://api.target.com",
				"endpoint":   "GET /invoices/{id}",
			},
		},
	}
	idorRes, err := srv.handleIdorMatrixPrompt(ctx, idorReq)
	if err != nil || len(idorRes.Messages) == 0 {
		t.Fatalf("handleIdorMatrixPrompt error: %v", err)
	}

	// 5. 403 Bypass prompt
	bypassReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_403_bypass",
			Arguments: map[string]string{
				"target_url":   "https://api.target.com",
				"blocked_path": "/admin/config",
			},
		},
	}
	bypassRes, err := srv.handleBypassPrompt(ctx, bypassReq)
	if err != nil || len(bypassRes.Messages) == 0 {
		t.Fatalf("handleBypassPrompt error: %v", err)
	}

	// 6. AI Prompt Injection Audit prompt
	aiAuditReq := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "cybermes_ai_prompt_injection_audit",
			Arguments: map[string]string{
				"target_url":     "https://api.target.com/v1/chat",
				"feature_type":   "chatbot",
				"injection_type": "direct_injection",
			},
		},
	}
	aiAuditRes, err := srv.handleAiPromptInjectionAuditPrompt(ctx, aiAuditReq)
	if err != nil || len(aiAuditRes.Messages) == 0 {
		t.Fatalf("handleAiPromptInjectionAuditPrompt error: %v", err)
	}
}

func TestToolFilterStream(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	mockLog := `
		[INFO] 200 OK https://example.com/index.html
		[CRITICAL] SQL Injection detected at https://example.com/api/v1/users?id=1'--
		[DEBUG] Loading static asset https://example.com/images/logo.png
		[HIGH] Auth Bypass found at https://example.com/admin/dashboard
		[INFO] 404 Not Found https://example.com/nonexistent
	`

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_filter_stream",
			Arguments: map[string]any{
				"content":   mockLog,
				"limit":     float64(5),
				"min_score": float64(10),
				"format":    "markdown",
			},
		},
	}

	res, err := srv.handleFilterStream(ctx, req)
	if err != nil {
		t.Fatalf("handleFilterStream error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Cybermes Smart Stream Filter") || !strings.Contains(text.Text, "SQL Injection") {
		t.Errorf("Expected filtered stream summary with SQL Injection, got: %s", text.Text)
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

func TestToolCheckEnvironment(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_check_environment",
			Arguments: map[string]any{
				"format": "markdown",
			},
		},
	}

	res, err := srv.handleCheckEnvironment(ctx, req)
	if err != nil {
		t.Fatalf("handleCheckEnvironment error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Cybermes Environment & Toolchain Diagnostic") {
		t.Errorf("Expected Diagnostic header in output, got: %s", text.Text)
	}
	if !strings.Contains(text.Text, "nuclei") {
		t.Errorf("Expected nuclei tool in status matrix, got: %s", text.Text)
	}
}

func TestToolRecordEvidence(t *testing.T) {
	srv, rootDir := setupTestServer(t)
	ctx := context.Background()

	testSlug := "test_evidence_mcp"
	targetDir := filepath.Join(rootDir, "reports", testSlug)
	defer os.RemoveAll(targetDir)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_record_evidence",
			Arguments: map[string]any{
				"target_slug": testSlug,
				"category":    "Missing Headers",
				"note":        "Strict-Transport-Security and Content-Security-Policy headers missing on /api/login.",
			},
		},
	}

	res, err := srv.handleRecordEvidence(ctx, req)
	if err != nil {
		t.Fatalf("handleRecordEvidence error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Evidence logged successfully") {
		t.Errorf("Expected success confirmation, got: %s", text.Text)
	}

	reconNotesPath := filepath.Join(targetDir, "evidence", "recon_notes.md")
	if _, err := os.Stat(reconNotesPath); os.IsNotExist(err) {
		t.Errorf("Expected recon_notes.md to be created at %s", reconNotesPath)
	}
}

func TestToolNucleiScan(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// Test with a mock / dummy target
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_nuclei_scan",
			Arguments: map[string]any{
				"target_url":      "https://example.com",
				"target_slug":     "test_nuclei_slug",
				"template_id":     "non-existent-dummy-cve-1234",
				"timeout_seconds": float64(5),
				"format":          "markdown",
			},
		},
	}

	res, err := srv.handleNucleiScan(ctx, req)
	if err != nil {
		t.Fatalf("handleNucleiScan error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	// Either executes or returns on-demand install instructions
	if !strings.Contains(text.Text, "Nuclei") {
		t.Errorf("Expected output to mention Nuclei, got: %s", text.Text)
	}
}

func TestFindProjectRoot(t *testing.T) {
	root := FindProjectRoot("")
	if root == "" || root == "." {
		t.Errorf("Expected FindProjectRoot to discover project root, got: %s", root)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err != nil {
		t.Errorf("Expected AGENTS.md in discovered root %s, but stat returned: %v", root, err)
	}

	// Test with environment variable
	t.Setenv("CYBERMES_ROOT", root)
	envRoot := FindProjectRoot("")
	if envRoot != root {
		t.Errorf("Expected CYBERMES_ROOT to return %s, got %s", root, envRoot)
	}
}

func TestToolSubdomainDiscovery(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_subdomain_discovery",
			Arguments: map[string]any{
				"domain":           "example.com",
				"prefer_subfinder": false,
				"timeout_seconds":  float64(5),
				"format":           "markdown",
			},
		},
	}

	res, err := srv.handleSubdomainDiscovery(ctx, req)
	if err != nil {
		t.Fatalf("handleSubdomainDiscovery error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Subdomain Discovery") {
		t.Errorf("Expected Subdomain Discovery output, got: %s", text.Text)
	}
}

func TestToolFuzzEndpoints(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "test_words.txt")
	_ = os.WriteFile(wlFile, []byte("api/v1/health\nadmin\nlogin\n"), 0644)

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_fuzz_endpoints",
			Arguments: map[string]any{
				"target_url":      mockServer.URL,
				"wordlist":        wlFile,
				"prefer_ffuf":     false,
				"rate_limit":      float64(25),
				"timeout_seconds": float64(3),
				"format":          "markdown",
			},
		},
	}

	res, err := srv.handleFuzzEndpoints(ctx, req)
	if err != nil {
		t.Fatalf("handleFuzzEndpoints error: %v", err)
	}
	text, ok := mcp.AsTextContent(res.Content[0])
	if !ok {
		t.Fatal("Expected TextContent in result")
	}
	if !strings.Contains(text.Text, "Endpoint Fuzzing Results") {
		t.Errorf("Expected Endpoint Fuzzing Results output, got: %s", text.Text)
	}
}

func TestToolFilterStream_JSONAndErrors(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// 1. Error on empty content
	emptyReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_filter_stream",
			Arguments: map[string]any{
				"content": "",
			},
		},
	}
	res, err := srv.handleFilterStream(ctx, emptyReq)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !res.IsError {
		t.Error("Expected error result on empty content")
	}

	// 2. JSON output format
	mockLog := "200 OK https://example.com/api/v1/users\n403 Forbidden /admin\n[DEBUG] static asset"
	jsonReq := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "cybermes_filter_stream",
			Arguments: map[string]any{
				"content":   mockLog,
				"format":    "json",
				"limit":     float64(10),
				"min_score": float64(5),
			},
		},
	}
	jsonRes, err := srv.handleFilterStream(ctx, jsonReq)
	if err != nil {
		t.Fatalf("handleFilterStream json error: %v", err)
	}
	jsonText, ok := mcp.AsTextContent(jsonRes.Content[0])
	if !ok || !strings.Contains(jsonText.Text, "total_raw_lines") {
		t.Errorf("Expected JSON stream filter result, got: %s", jsonText.Text)
	}
}

func TestPrompts_MissingArguments(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	// Missing target in cybermes_hunt
	_, err := srv.handleHuntPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{Name: "cybermes_hunt", Arguments: map[string]string{}},
	})
	if err == nil {
		t.Error("Expected error when target is missing in hunt prompt")
	}

	// Missing target_url in cybermes_api_audit
	_, err = srv.handleApiAuditPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{Name: "cybermes_api_audit", Arguments: map[string]string{}},
	})
	if err == nil {
		t.Error("Expected error when target_url is missing in api audit prompt")
	}

	// Missing endpoint in cybermes_idor_matrix
	_, err = srv.handleIdorMatrixPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{Name: "cybermes_idor_matrix", Arguments: map[string]string{"target_url": "https://example.com"}},
	})
	if err == nil {
		t.Error("Expected error when endpoint is missing in idor matrix prompt")
	}

	// Missing blocked_path in cybermes_403_bypass
	_, err = srv.handleBypassPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{Name: "cybermes_403_bypass", Arguments: map[string]string{"target_url": "https://example.com"}},
	})
	if err == nil {
		t.Error("Expected error when blocked_path is missing in 403 bypass prompt")
	}

	// Missing target_url in cybermes_ai_prompt_injection_audit
	_, err = srv.handleAiPromptInjectionAuditPrompt(ctx, mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{Name: "cybermes_ai_prompt_injection_audit", Arguments: map[string]string{}},
	})
	if err == nil {
		t.Error("Expected error when target_url is missing in ai prompt injection audit prompt")
	}
}


