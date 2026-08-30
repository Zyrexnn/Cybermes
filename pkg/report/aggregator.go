package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SummaryData struct {
	Target          string            `json:"target"`
	ScanTime        string            `json:"scan_time"`
	TotalFindings   int               `json:"total_findings"`
	SeveritySummary map[string]int    `json:"severity_summary"`
	Findings        []*FindingMeta    `json:"findings"`
	PoCs            []string          `json:"pocs"`
	EvidenceFiles   []string          `json:"evidence_files"`
	ReconNotes      *string           `json:"recon_notes"`
}

func AggregateTarget(targetDir string) (*SummaryData, error) {
	findingsDir := filepath.Join(targetDir, "findings")
	pocsDir := filepath.Join(targetDir, "pocs")
	evidenceDir := filepath.Join(targetDir, "evidence")
	targetName := filepath.Base(targetDir)

	var findings []*FindingMeta
	if entries, err := os.ReadDir(findingsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if !strings.HasSuffix(name, ".md") || name == "summary.md" || name == "readme.md" {
				continue
			}
			meta, err := ParseFindingFile(filepath.Join(findingsDir, e.Name()))
			if err == nil {
				findings = append(findings, meta)
			}
		}
	}

	sort.SliceStable(findings, func(i, j int) bool {
		wI := severityWeights[findings[i].Severity]
		wJ := severityWeights[findings[j].Severity]
		if wI != wJ {
			return wI < wJ
		}
		return findings[i].Title < findings[j].Title
	})

	var pocs []string
	if entries, err := os.ReadDir(pocsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				pocs = append(pocs, e.Name())
			}
		}
	}
	sort.Strings(pocs)

	var evidence []string
	if entries, err := os.ReadDir(evidenceDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				evidence = append(evidence, e.Name())
			}
		}
	}
	sort.Strings(evidence)

	sevSummary := map[string]int{
		"CRITICAL":      0,
		"HIGH":          0,
		"MEDIUM":        0,
		"LOW":           0,
		"INFORMATIONAL": 0,
	}
	for _, f := range findings {
		if _, ok := sevSummary[f.Severity]; ok {
			sevSummary[f.Severity]++
		}
	}

	var reconNotes *string
	reconPath := filepath.Join(evidenceDir, "recon_notes.md")
	if _, err := os.Stat(reconPath); err == nil {
		rn := "evidence/recon_notes.md"
		reconNotes = &rn
	}

	summaryData := &SummaryData{
		Target:          targetName,
		ScanTime:        time.Now().Format("2006-01-02 15:04:05"),
		TotalFindings:   len(findings),
		SeveritySummary: sevSummary,
		Findings:        findings,
		PoCs:            pocs,
		EvidenceFiles:   evidence,
		ReconNotes:      reconNotes,
	}

	metaFile := filepath.Join(targetDir, "metadata.json")
	metaBytes, err := json.MarshalIndent(summaryData, "", "  ")
	if err == nil {
		os.WriteFile(metaFile, metaBytes, 0666)
	}

	summaryFile := filepath.Join(targetDir, "SUMMARY.md")
	existingBytes, _ := os.ReadFile(summaryFile)
	customContent := ExtractCustomSections(string(existingBytes))

	if err := GenerateSummaryMD(targetDir, summaryData, customContent); err != nil {
		return nil, err
	}

	// Always generate interactive report.html dashboard
	_, _ = GenerateHTMLDashboard(targetDir, summaryData)

	return summaryData, nil
}

// AggregateTargetWithPDF aggregates findings, generates SUMMARY.md, metadata.json, report.html, and optionally REPORT.pdf
func AggregateTargetWithPDF(targetDir string, generatePDF bool) (*SummaryData, *ReportArtifacts, error) {
	summaryData, err := AggregateTarget(targetDir)
	if err != nil {
		return nil, nil, err
	}
	artifacts, err := GenerateFullReport(targetDir, summaryData, generatePDF)
	return summaryData, artifacts, err
}

func GenerateSummaryMD(targetDir string, data *SummaryData, customContent string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# 🛡️ Security Assessment Summary: `%s`\n\n", data.Target))
	sb.WriteString(fmt.Sprintf("- **Generated At**: %s\n", data.ScanTime))
	sb.WriteString(fmt.Sprintf("- **Total Confirmed Findings**: %d\n", data.TotalFindings))
	sb.WriteString(fmt.Sprintf("- **Severity Breakdown**: 🔴 Critical: %d | 🟠 High: %d | 🟡 Medium: %d | 🔵 Low: %d | ⚪ Info: %d\n\n",
		data.SeveritySummary["CRITICAL"],
		data.SeveritySummary["HIGH"],
		data.SeveritySummary["MEDIUM"],
		data.SeveritySummary["LOW"],
		data.SeveritySummary["INFORMATIONAL"],
	))
	sb.WriteString("---\n\n## 📑 Findings Matrix\n\n")
	sb.WriteString("| Severity | Title / Vulnerability | CVSS v3.1 | CWE | Affected Endpoint | Report Link |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")

	if len(data.Findings) == 0 {
		sb.WriteString("| - | *No confirmed vulnerabilities reported* | - | - | - | - |\n")
	} else {
		for _, f := range data.Findings {
			badge := f.Severity
			switch f.Severity {
			case "CRITICAL":
				badge = "🔴 `CRITICAL`"
			case "HIGH":
				badge = "🟠 `HIGH`"
			case "MEDIUM":
				badge = "🟡 `MEDIUM`"
			case "LOW":
				badge = "🔵 `LOW`"
			case "INFORMATIONAL":
				badge = "⚪ `INFO`"
			default:
				badge = fmt.Sprintf("`%s`", f.Severity)
			}
			link := fmt.Sprintf("[%s](%s)", f.FileName, f.RelativePath)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | `%s` | %s |\n",
				badge, f.Title, f.CVSS, f.CWE, f.Endpoint, link))
		}
	}

	sb.WriteString("\n---\n\n## 🧪 Proof of Concept Scripts (`pocs/`)\n\n")
	if len(data.PoCs) > 0 {
		for _, poc := range data.PoCs {
			sb.WriteString(fmt.Sprintf("- [`pocs/%s`](pocs/%s)\n", poc, poc))
		}
	} else {
		sb.WriteString("- *No standalone PoC scripts attached.*\n")
	}

	sb.WriteString("\n## 📁 Evidence & Recon Notes (`evidence/`)\n\n")
	if data.ReconNotes != nil {
		sb.WriteString(fmt.Sprintf("- 📝 **[Reconnaissance & Informational Notes](%s)**\n", *data.ReconNotes))
	}
	if len(data.EvidenceFiles) > 0 {
		for _, ev := range data.EvidenceFiles {
			sb.WriteString(fmt.Sprintf("- [`evidence/%s`](evidence/%s)\n", ev, ev))
		}
	} else {
		sb.WriteString("- *No visual or trace evidence attached.*\n")
	}

	if customContent != "" {
		sb.WriteString("\n---\n\n")
		sb.WriteString(customContent)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	summaryPath := filepath.Join(targetDir, "SUMMARY.md")
	return os.WriteFile(summaryPath, []byte(sb.String()), 0666)
}

func AggregateAll(reportsDir string) ([]*SummaryData, error) {
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, err
	}
	var results []*SummaryData
	for _, e := range entries {
		if e.IsDir() {
			targetDir := filepath.Join(reportsDir, e.Name())
			summary, err := AggregateTarget(targetDir)
			if err == nil {
				results = append(results, summary)
			}
		}
	}
	return results, nil
}
