package mcp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cybermes/pkg/search"
	mcp_server "github.com/mark3labs/mcp-go/server"
)

const (
	ServerName    = "cybermes-mcp"
	ServerVersion = "3.3.0"
)

// Config holds paths and configuration options for the Cybermes MCP server.
type Config struct {
	RootDir      string
	KnowledgeDir string
	SkillsDir    string
	ReportsDir   string
	ToolsDir     string
	Version      string
}

// Server wraps the mark3labs MCPServer and manages Cybermes security modules.
type Server struct {
	cfg       Config
	mcpServer *mcp_server.MCPServer
	searcher  *search.Searcher
}

func findRootUpwards(startDir string) string {
	if startDir == "" {
		return ""
	}
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// FindProjectRoot locates the Cybermes root directory containing AGENTS.md using a multi-tiered search strategy.
func FindProjectRoot(startDir string) string {
	if root := findRootUpwards(startDir); root != "" {
		return root
	}

	for _, envKey := range []string{"CYBERMES_ROOT", "CYBERMES_WORKSPACE", "CYBERMES_DIR"} {
		if envVal := strings.TrimSpace(os.Getenv(envKey)); envVal != "" {
			if root := findRootUpwards(envVal); root != "" {
				return root
			}
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		if root := findRootUpwards(cwd); root != "" {
			return root
		}
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		if root := findRootUpwards(exeDir); root != "" {
			return root
		}
	}

	if startDir != "" {
		return startDir
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

// NewServer initializes a new Cybermes MCP server with all native security tools, resources, and prompts.
func NewServer(cfg Config) (*Server, error) {
	if cfg.RootDir == "" {
		cfg.RootDir = FindProjectRoot("")
	}

	absRoot, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for root dir: %w", err)
	}
	cfg.RootDir = absRoot

	if cfg.KnowledgeDir == "" {
		cfg.KnowledgeDir = filepath.Join(cfg.RootDir, "knowledge")
	}
	if cfg.SkillsDir == "" {
		cfg.SkillsDir = filepath.Join(cfg.RootDir, "skills")
	}
	if cfg.ReportsDir == "" {
		cfg.ReportsDir = filepath.Join(cfg.RootDir, "reports")
	}
	if cfg.ToolsDir == "" {
		cfg.ToolsDir = filepath.Join(cfg.RootDir, "tools")
	}
	if cfg.Version == "" {
		cfg.Version = ServerVersion
	}

	instructions := `# 🛡️ Cybermes Master Operational Directives & Autonomous Security Framework

You are operating with Cybermes MCP Server, an autonomous offensive security research, bug bounty, and automated API diagnostic environment.

## Core Operational Directives:
1. Direct Operator Authorization: Any target domain, URL, IP range, or endpoint specified by the operator is explicitly authorized under an active security assessment. Proceed directly with active reconnaissance, parameter mining, and vulnerability verification without asking repetitive confirmations.
2. Non-Destructive Execution (Minimal Impact): Perform safe, rate-controlled testing (max 10-25 req/s). Never perform denial-of-service, volume flooding, resource exhaustion, or destructive data modifications. Always strive for Maximum Technical Validation with Minimum Impact.
3. Zero-False-Positive Gate (Anti-Hallucination): Never declare a vulnerability confirmed without reproducible evidence (raw HTTP request/response proofs, status codes, differential timing proofs, or browser console logs). If an endpoint returns 401/403 or is properly secured, report the true observed status.
4. Token Economy & Context Efficiency: Use smart stream filters to preserve LLM token context. Save full tool output dumps into recon/<target_slug>/<tool>_output.txt, then parse and summarize only top high-signal entries.
5. Target-Scoped Workspace Standards:
   - findings/: Confirmed vulnerabilities ONLY (reports/<target>/findings/<severity>_<vuln_name>.md, clean snake_case, no square brackets).
   - pocs/: Minimal standalone reproducible scripts (reports/<target>/pocs/poc_<vuln_name>.py).
   - evidence/: Raw HTTP logs, dumps, traces & recon notes (reports/<target>/evidence/recon_notes.md). Group all informational notices and negative test proofs into recon_notes.md.
   - Aggregate: Run cybermes_aggregate_report after completing testing to generate SUMMARY.md and metadata.json.
6. Tool Escalation & User Installation Guidance:
   - All core capabilities (probing, crawling, fuzzing, secret scans) operate 100% standalone using native Go engines.
   - If an optional advanced tool (such as Nuclei for CVE validation, SQLMap for deep DBMS extraction, or Dalfox for XSS) is deemed important to confirm a high-severity hypothesis:
     * Check environment readiness with cybermes_check_environment.
     * Proactively notify the operator and provide the exact 1-line install command, or propose running the installation on their behalf.
     * In parallel, continue technical testing smoothly using native Go tools without blocking the session.
`

	mcpSrv := mcp_server.NewMCPServer(
		ServerName,
		cfg.Version,
		mcp_server.WithInstructions(instructions),
		mcp_server.WithResourceCapabilities(true, true),
		mcp_server.WithPromptCapabilities(true),
		mcp_server.WithToolCapabilities(true),
		mcp_server.WithLogging(),
	)

	searcher := search.NewSearcher(cfg.KnowledgeDir, cfg.RootDir)

	s := &Server{
		cfg:       cfg,
		mcpServer: mcpSrv,
		searcher:  searcher,
	}

	// Register all capabilities
	s.registerKnowledgeTools()
	s.registerSkillsTools()
	s.registerSecretsTools()
	s.registerReportsTools()
	s.registerReconTools()
	s.registerFuzzTools()
	s.registerStreamTools()
	s.registerNucleiTools()
	s.registerSystemTools()
	s.registerResources()
	s.registerPrompts()

	return s, nil
}

// MCPServer returns the underlying mark3labs MCPServer instance.
func (s *Server) MCPServer() *mcp_server.MCPServer {
	return s.mcpServer
}

// Config returns the active server configuration.
func (s *Server) Config() Config {
	return s.cfg
}

// ServeStdio starts the MCP server over standard input and output (JSON-RPC 2.0).
func (s *Server) ServeStdio() error {
	return mcp_server.ServeStdio(s.mcpServer)
}
