package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

var binaries = []struct {
	name    string
	pkgPath string
}{
	{"smart_pipe", "./cmd/smart_pipe"},
	{"secret_scan", "./cmd/secret_scan"},
	{"search_knowledge", "./cmd/search_knowledge"},
	{"aggregate_reports", "./cmd/aggregate_reports"},
	{"cybermes-mcp", "./cmd/cybermes-mcp"},
}

func main() {
	start := time.Now()
	rootDir, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to determine root directory: %v\n", err)
		os.Exit(1)
	}

	outDir := filepath.Join(rootDir, "tools", "bin")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	fmt.Printf("🔨 Building Cybermes Go Toolchain (%s/%s)...\n", runtime.GOOS, runtime.GOARCH)

	for _, b := range binaries {
		outPath := filepath.Join(outDir, b.name+ext)
		cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", outPath, b.pkgPath)
		cmd.Dir = rootDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Build failed for %s: %v\n", b.name, err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ %s%s\n", b.name, ext)
	}

	fmt.Printf("✨ All %d binaries compiled successfully to tools/bin in %v\n", len(binaries), time.Since(start).Round(time.Millisecond))
}
