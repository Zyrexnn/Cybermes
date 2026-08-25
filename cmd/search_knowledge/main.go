package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cybermes/pkg/search"
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

func formatCLIOutput(results []search.Snippet, query string) {
	if len(results) == 0 {
		fmt.Printf("🔍 [Knowledge Search] No relevant knowledge found for query: '%s'\n", query)
		return
	}

	fmt.Println("\n📚 ══════════════════════════════════════════════════════════════════")
	fmt.Printf("   CYBERMES KNOWLEDGE BASE SEARCH: '%s'\n", query)
	fmt.Printf("   Found %d high-signal snippets (Ranked by relevance)\n", len(results))
	fmt.Println("══════════════════════════════════════════════════════════════════════")

	for i, res := range results {
		fmt.Printf("─── [Result #%d | Score: %d] ──────────────────────────────────────────\n", i+1, res.Score)
		fmt.Printf("📂 KB Source : [%s]\n", res.SourceKB)
		fmt.Printf("📄 Location  : %s:%d\n", res.File, res.StartLine)
		fmt.Printf("🏷️ Section   : %s\n\n", res.Heading)
		fmt.Println(res.Content)
		fmt.Println()
	}

	fmt.Println("💡 Tip: Use '--limit N' or '--source [payloads|hacktricks|claude|strix]' to filter.")
}

func main() {
	source := "all"
	limit := 3
	maxLen := 1400
	jsonOutput := false

	var queryParts []string
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--source" || arg == "-s" {
			if i+1 < len(args) {
				source = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--source=") {
			source = strings.TrimPrefix(arg, "--source=")
		} else if arg == "--limit" || arg == "-n" || arg == "-l" {
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					limit = v
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--limit=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(arg, "--limit=")); err == nil {
				limit = v
			}
		} else if arg == "--max-len" || arg == "-m" {
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					maxLen = v
				}
				i++
			}
		} else if strings.HasPrefix(arg, "--max-len=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(arg, "--max-len=")); err == nil {
				maxLen = v
			}
		} else if arg == "--json" || arg == "-j" {
			jsonOutput = true
		} else if arg == "--help" || arg == "-h" {
			flag.Usage()
			return
		} else {
			queryParts = append(queryParts, arg)
		}
	}

	query := strings.Join(queryParts, " ")
	if strings.TrimSpace(query) == "" {
		fmt.Println("Usage: search_knowledge <query> [--source <source>] [--limit <limit>] [--json]")
		os.Exit(1)
	}

	rootDir := findProjectRoot()
	kbDir := filepath.Join(rootDir, "knowledge")

	searcher := search.NewSearcher(kbDir, rootDir)
	results, err := searcher.Search(query, source, limit, maxLen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search error: %v\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		out := map[string]interface{}{
			"query":         query,
			"source":        source,
			"total_results": len(results),
			"results":       results,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
	} else {
		formatCLIOutput(results, query)
	}
}
