package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"cybermes/pkg/report"
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
	allFlag := flag.Bool("all", false, "Aggregate reports for all targets in reports/ directory")
	flag.BoolVar(allFlag, "a", false, "Aggregate reports for all targets (shorthand)")
	pdfFlag := flag.Bool("pdf", true, "Generate executive PDF report via native Chrome DevTools Protocol")
	noPDFFlag := flag.Bool("no-pdf", false, "Skip PDF generation (generate HTML & Markdown only)")
	flag.Parse()

	generatePDF := *pdfFlag && !*noPDFFlag

	rootDir := findProjectRoot()
	reportsDir := filepath.Join(rootDir, "reports")

	if *allFlag {
		results, err := report.AggregateAll(reportsDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error aggregating all targets: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully aggregated reports for %d targets.\n", len(results))
		return
	}

	target := flag.Arg(0)
	if target == "" {
		fmt.Println("Usage: aggregate_reports <TARGET_SLUG> [--no-pdf] OR aggregate_reports --all")
		os.Exit(1)
	}

	targetDir := filepath.Join(reportsDir, target)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0777); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating target report directory: %v\n", err)
			os.Exit(1)
		}
	}

	summary, artifacts, err := report.AggregateTargetWithPDF(targetDir, generatePDF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error aggregating target %s: %v\n", target, err)
		os.Exit(1)
	}

	fmt.Printf("[Aggregate Reports] %s | Confirmed: %d (Crit: %d, High: %d, Med: %d, Low: %d, Info: %d)\n",
		summary.Target,
		summary.TotalFindings,
		summary.SeveritySummary["CRITICAL"],
		summary.SeveritySummary["HIGH"],
		summary.SeveritySummary["MEDIUM"],
		summary.SeveritySummary["LOW"],
		summary.SeveritySummary["INFORMATIONAL"],
	)
	fmt.Printf("  Updated: reports/%s/SUMMARY.md\n", summary.Target)
	fmt.Printf("  Updated: reports/%s/metadata.json\n", summary.Target)
	if artifacts != nil && artifacts.HTMLPath != "" {
		fmt.Printf("  Dashboard: reports/%s/report.html\n", summary.Target)
	}
	if artifacts != nil && artifacts.PDFGenerated {
		fmt.Printf("  Executive PDF: reports/%s/REPORT.pdf\n", summary.Target)
	}
}
