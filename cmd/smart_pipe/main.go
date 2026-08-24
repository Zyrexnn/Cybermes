package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cybermes/pkg/stream"
)

func findProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := cwd
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
	return cwd
}

func main() {
	target := flag.String("target", "default_target", "Target slug identifier")
	flag.StringVar(target, "t", "default_target", "Target slug identifier (shorthand)")
	tool := flag.String("tool", "tool", "Security tool name (katana, ffuf, httpx, etc.)")
	flag.StringVar(tool, "n", "tool", "Security tool name (shorthand)")
	limit := flag.Int("limit", 40, "Max prioritized lines to display in context")
	flag.IntVar(limit, "l", 40, "Max prioritized lines to display in context (shorthand)")
	flag.Parse()

	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: <tool_command> | smart_pipe --target <SLUG> --tool <TOOL>")
		os.Exit(1)
	}

	rootDir := findProjectRoot()
	targetReconDir := filepath.Join(rootDir, "recon", *target)
	if err := os.MkdirAll(targetReconDir, 0777); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating recon directory: %v\n", err)
		os.Exit(1)
	}

	rawLogPath := filepath.Join(targetReconDir, fmt.Sprintf("%s_raw.txt", *tool))
	rawFile, err := os.OpenFile(rawLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening raw log file: %v\n", err)
		os.Exit(1)
	}
	defer rawFile.Close()

	res, err := stream.ProcessStream(os.Stdin, os.Stdout, rawFile, *limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing stream: %v\n", err)
		os.Exit(1)
	}

	relPath, err := filepath.Rel(rootDir, rawLogPath)
	if err != nil {
		relPath = rawLogPath
	}
	fmt.Printf("💾 Full raw output preserved: %s\n", relPath)
	_ = res
}
