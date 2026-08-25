package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"cybermes/pkg/mcp"
)

func main() {
	var (
		workspaceFlag = flag.String("workspace", "", "Path to the Cybermes repository root directory (default: auto-detect)")
		versionFlag   = flag.Bool("version", false, "Print Cybermes MCP server version and exit")
	)
	flag.StringVar(workspaceFlag, "w", "", "Path to Cybermes repository root (shorthand)")
	flag.BoolVar(versionFlag, "v", false, "Print version (shorthand)")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Cybermes MCP Server v%s\n", mcp.ServerVersion)
		os.Exit(0)
	}

	// Direct all diagnostics to stderr; stdout is reserved for MCP JSON-RPC protocol
	logger := log.New(os.Stderr, "[cybermes-mcp] ", log.LstdFlags)

	rootDir := *workspaceFlag
	if rootDir == "" {
		rootDir = mcp.FindProjectRoot("")
	}

	cfg := mcp.Config{
		RootDir: rootDir,
		Version: mcp.ServerVersion,
	}

	server, err := mcp.NewServer(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize Cybermes MCP server: %v", err)
	}

	logger.Printf("Cybermes MCP Server v%s initialized (Workspace: %s)", cfg.Version, server.Config().RootDir)
	logger.Printf("Listening for JSON-RPC 2.0 requests over stdio...")

	if err := server.ServeStdio(); err != nil {
		logger.Fatalf("Server terminated with error: %v", err)
	}
}
