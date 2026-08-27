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
	ServerVersion = "3.1.1"
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

	instructions := "Cybermes MCP Server provides autonomous offensive security research capabilities, " +
		"including 200+ vulnerability playbooks, curated payload knowledge base (<50ms search), " +
		"high-speed 48-pattern credential leak detection, and executive report aggregation."

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
