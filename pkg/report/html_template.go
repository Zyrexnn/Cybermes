package report

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type DetailedFinding struct {
	Title        string
	Severity     string
	CVSS         string
	CWE          string
	Endpoint     string
	FileName     string
	HTMLContent  template.HTML
}

type HTMLReportData struct {
	Target           string
	ScanTime         string
	TotalFindings    int
	Stats            map[string]int
	Findings         []*FindingMeta
	DetailedFindings []DetailedFinding
	PoCs             []string
	EvidenceFiles    []string
	ReconNotes       string
}

const reportHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Security Assessment Report — {{ .Target }}</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap');

  :root {
    --bg: #0f172a;
    --card-bg: #1e293b;
    --text-main: #f8fafc;
    --text-muted: #94a3b8;
    --border: #334155;
    --crit: #ef4444;
    --high: #f97316;
    --med: #eab308;
    --low: #3b82f6;
    --info: #64748b;
    --accent: #4f46e5;
    --code-bg: #090d16;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }
  
  body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background-color: #ffffff;
    color: #1e293b;
    line-height: 1.6;
    font-size: 14px;
    padding: 0;
  }

  .container {
    max-width: 900px;
    margin: 0 auto;
    padding: 30px;
  }

  /* Header & Metadata */
  .header-card {
    border-bottom: 2px solid #e2e8f0;
    padding-bottom: 25px;
    margin-bottom: 30px;
  }

  .brand-tag {
    display: inline-block;
    background: var(--accent);
    color: #ffffff;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    padding: 4px 10px;
    border-radius: 4px;
    margin-bottom: 12px;
  }

  h1 {
    font-size: 26px;
    font-weight: 800;
    color: #0f172a;
    margin-bottom: 8px;
  }

  .meta-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 15px;
    margin-top: 20px;
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    padding: 15px;
    border-radius: 8px;
  }

  .meta-item strong {
    display: block;
    font-size: 11px;
    text-transform: uppercase;
    color: #64748b;
    letter-spacing: 0.5px;
    margin-bottom: 4px;
  }

  .meta-item span {
    font-size: 14px;
    font-weight: 600;
    color: #1e293b;
  }

  /* Severity Metrics Grid */
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 12px;
    margin-bottom: 30px;
  }

  .stat-box {
    padding: 16px;
    border-radius: 8px;
    text-align: center;
    border: 1px solid transparent;
  }

  .stat-box .count {
    font-size: 28px;
    font-weight: 800;
    line-height: 1;
    margin-bottom: 6px;
  }

  .stat-box .label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .stat-crit { background: #fef2f2; border-color: #fecaca; color: var(--crit); }
  .stat-high { background: #fff7ed; border-color: #ffedd5; color: var(--high); }
  .stat-med  { background: #fefce8; border-color: #fef08a; color: #ca8a04; }
  .stat-low  { background: #eff6ff; border-color: #dbeafe; color: var(--low); }
  .stat-info { background: #f8fafc; border-color: #e2e8f0; color: var(--info); }

  /* Sections */
  .section-title {
    font-size: 18px;
    font-weight: 700;
    color: #0f172a;
    margin: 30px 0 15px 0;
    padding-bottom: 8px;
    border-bottom: 1px solid #e2e8f0;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  /* Table */
  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 30px;
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    overflow: hidden;
  }

  th {
    background: #f8fafc;
    text-align: left;
    padding: 12px 16px;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    color: #475569;
    border-bottom: 1px solid #e2e8f0;
  }

  td {
    padding: 12px 16px;
    border-bottom: 1px solid #f1f5f9;
    font-size: 13px;
    vertical-align: middle;
  }

  tr:last-child td { border-bottom: none; }
  tr:hover td { background: #f8fafc; }

  /* Badges */
  .badge {
    display: inline-block;
    padding: 3px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .badge-CRITICAL { background: #fee2e2; color: #991b1b; }
  .badge-HIGH     { background: #ffedd5; color: #9a3412; }
  .badge-MEDIUM   { background: #fef9c3; color: #854d0e; }
  .badge-LOW      { background: #e0f2fe; color: #075985; }
  .badge-INFORMATIONAL { background: #f1f5f9; color: #475569; }
  .badge-UNKNOWN  { background: #f1f5f9; color: #475569; }

  /* Detailed Finding Cards */
  .finding-card {
    background: #ffffff;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    margin-bottom: 24px;
    page-break-inside: avoid;
    box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  }

  .finding-header {
    background: #f8fafc;
    padding: 16px 20px;
    border-bottom: 1px solid #e2e8f0;
    border-top-left-radius: 8px;
    border-top-right-radius: 8px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .finding-title {
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
  }

  .finding-body {
    padding: 20px;
  }

  .finding-body h2 { font-size: 15px; margin: 16px 0 8px 0; color: #1e293b; }
  .finding-body h3 { font-size: 14px; margin: 14px 0 6px 0; color: #334155; }
  .finding-body p { margin-bottom: 12px; }
  .finding-body ul, .finding-body ol { margin: 0 0 14px 20px; }
  .finding-body li { margin-bottom: 4px; }

  /* Code Blocks */
  pre {
    background: var(--code-bg);
    color: #f8fafc;
    padding: 14px;
    border-radius: 6px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    overflow-x: auto;
    margin: 12px 0;
    line-height: 1.5;
  }

  code {
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    background: #f1f5f9;
    color: #0f172a;
    padding: 2px 5px;
    border-radius: 4px;
  }

  pre code {
    background: transparent;
    color: inherit;
    padding: 0;
  }

  .endpoint-box {
    background: #f1f5f9;
    border-left: 3px solid var(--accent);
    padding: 8px 12px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    margin-bottom: 15px;
    border-radius: 0 4px 4px 0;
  }

  /* Footer */
  .footer {
    margin-top: 40px;
    padding-top: 20px;
    border-top: 1px solid #e2e8f0;
    text-align: center;
    font-size: 12px;
    color: #94a3b8;
  }

  @media print {
    body { background-color: #ffffff; font-size: 12px; }
    .container { max-width: 100%; padding: 0; }
    .finding-card { border: 1px solid #cbd5e1; margin-bottom: 20px; }
    pre { background: #1e293b !important; color: #f8fafc !important; -webkit-print-color-adjust: exact; }
    .badge { -webkit-print-color-adjust: exact; }
    .stat-box { -webkit-print-color-adjust: exact; }
  }
</style>
</head>
<body>

<div class="container">
  <!-- Header Card -->
  <div class="header-card">
    <div class="brand-tag">Cybermes Security Assessment</div>
    <h1>Executive Security Report</h1>
    
    <div class="meta-grid">
      <div class="meta-item">
        <strong>Target</strong>
        <span>{{ .Target }}</span>
      </div>
      <div class="meta-item">
        <strong>Assessment Date</strong>
        <span>{{ .ScanTime }}</span>
      </div>
      <div class="meta-item">
        <strong>Total Findings</strong>
        <span>{{ .TotalFindings }} Confirmed</span>
      </div>
    </div>
  </div>

  <!-- Severity Matrix Box -->
  <div class="stats-grid">
    <div class="stat-box stat-crit">
      <div class="count">{{ index .Stats "CRITICAL" }}</div>
      <div class="label">Critical</div>
    </div>
    <div class="stat-box stat-high">
      <div class="count">{{ index .Stats "HIGH" }}</div>
      <div class="label">High</div>
    </div>
    <div class="stat-box stat-med">
      <div class="count">{{ index .Stats "MEDIUM" }}</div>
      <div class="label">Medium</div>
    </div>
    <div class="stat-box stat-low">
      <div class="count">{{ index .Stats "LOW" }}</div>
      <div class="label">Low</div>
    </div>
    <div class="stat-box stat-info">
      <div class="count">{{ index .Stats "INFORMATIONAL" }}</div>
      <div class="label">Info</div>
    </div>
  </div>

  <!-- Findings Table -->
  <div class="section-title">
    <span>Confirmed Vulnerabilities</span>
    <span style="font-size: 13px; font-weight: 500; color: #64748b;">{{ .TotalFindings }} Total</span>
  </div>

  <table>
    <thead>
      <tr>
        <th style="width: 110px;">Severity</th>
        <th>Title / Vulnerability</th>
        <th style="width: 90px;">CVSS</th>
        <th style="width: 100px;">CWE</th>
        <th>Vulnerable Endpoint</th>
      </tr>
    </thead>
    <tbody>
      {{ if eq (len .Findings) 0 }}
      <tr>
        <td colspan="5" style="text-align: center; color: #64748b; padding: 20px;">No confirmed vulnerabilities reported for this target.</td>
      </tr>
      {{ else }}
      {{ range .Findings }}
      <tr>
        <td><span class="badge badge-{{ .Severity }}">{{ .Severity }}</span></td>
        <td><strong>{{ .Title }}</strong></td>
        <td><code>{{ if .CVSS }}{{ .CVSS }}{{ else }}-{{ end }}</code></td>
        <td><code>{{ if .CWE }}{{ .CWE }}{{ else }}-{{ end }}</code></td>
        <td><code style="word-break: break-all;">{{ if .Endpoint }}{{ .Endpoint }}{{ else }}-{{ end }}</code></td>
      </tr>
      {{ end }}
      {{ end }}
    </tbody>
  </table>

  <!-- Standalone PoC Scripts -->
  {{ if gt (len .PoCs) 0 }}
  <div class="section-title">
    <span>Standalone Proof-of-Concept Scripts</span>
    <span style="font-size: 13px; font-weight: 500; color: #64748b;">{{ len .PoCs }} Attached</span>
  </div>
  <table>
    <thead>
      <tr>
        <th>Script Name</th>
        <th style="width: 120px;">Type</th>
        <th>Location</th>
      </tr>
    </thead>
    <tbody>
      {{ range .PoCs }}
      <tr>
        <td><strong>{{ . }}</strong></td>
        <td><code>PoC Script</code></td>
        <td><code>reports/{{ $.Target }}/pocs/{{ . }}</code></td>
      </tr>
      {{ end }}
    </tbody>
  </table>
  {{ end }}

  <!-- Detailed Finding Writeups -->
  {{ if gt (len .DetailedFindings) 0 }}
  <div class="section-title" style="margin-top: 40px;">
    <span>Detailed Finding Writeups & Technical Evidence</span>
  </div>

  {{ range .DetailedFindings }}
  <div class="finding-card">
    <div class="finding-header">
      <span class="finding-title">{{ .Title }}</span>
      <span class="badge badge-{{ .Severity }}">{{ .Severity }}</span>
    </div>
    <div class="finding-body">
      {{ if .Endpoint }}
      <div class="endpoint-box"><strong>Endpoint:</strong> {{ .Endpoint }}</div>
      {{ end }}
      {{ .HTMLContent }}
    </div>
  </div>
  {{ end }}
  {{ end }}

  <!-- Footer -->
  <div class="footer">
    <p>Generated by Cybermes Offensive Security & Bug Bounty Framework</p>
    <p style="font-size: 11px; margin-top: 4px;">Confidential Assessment Report &bull; Authorized Security Research Only</p>
  </div>
</div>

</body>
</html>
`

// RenderMarkdownToHTML converts basic markdown into clean HTML safe for the report
func RenderMarkdownToHTML(md string) template.HTML {
	lines := strings.Split(md, "\n")
	var sb strings.Builder
	inCodeBlock := false
	inList := false
	inTable := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block handling
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				sb.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				if inList {
					sb.WriteString("</ul>\n")
					inList = false
				}
				if inTable {
					sb.WriteString("</table>\n")
					inTable = false
				}
				sb.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			sb.WriteString(html.EscapeString(line) + "\n")
			continue
		}

		// Table handling
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			if strings.Contains(trimmed, "---") {
				continue // skip header separator
			}
			if !inTable {
				if inList {
					sb.WriteString("</ul>\n")
					inList = false
				}
				sb.WriteString("<table>\n")
				inTable = true
			}
			cells := strings.Split(strings.Trim(trimmed, "|"), "|")
			sb.WriteString("<tr>")
			for _, cell := range cells {
				sb.WriteString("<td>" + formatInlineMarkdown(strings.TrimSpace(cell)) + "</td>")
			}
			sb.WriteString("</tr>\n")
			continue
		} else if inTable {
			sb.WriteString("</table>\n")
			inTable = false
		}

		// Empty line
		if trimmed == "" {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			continue
		}

		// Headings
		if strings.HasPrefix(trimmed, "### ") {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			sb.WriteString("<h3>" + formatInlineMarkdown(trimmed[4:]) + "</h3>\n")
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			sb.WriteString("<h2>" + formatInlineMarkdown(trimmed[3:]) + "</h2>\n")
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			if inList {
				sb.WriteString("</ul>\n")
				inList = false
			}
			continue // skip primary H1 as it is rendered in card header
		}

		// List items
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList {
				sb.WriteString("<ul>\n")
				inList = true
			}
			sb.WriteString("<li>" + formatInlineMarkdown(trimmed[2:]) + "</li>\n")
			continue
		}

		// Numbered list
		numberedMatch := regexp.MustCompile(`^\d+\.\s+`).FindString(trimmed)
		if numberedMatch != "" {
			if !inList {
				sb.WriteString("<ol>\n")
				inList = true
			}
			sb.WriteString("<li>" + formatInlineMarkdown(trimmed[len(numberedMatch):]) + "</li>\n")
			continue
		}

		if inList {
			sb.WriteString("</ul>\n")
			inList = false
		}

		// Regular paragraph
		sb.WriteString("<p>" + formatInlineMarkdown(trimmed) + "</p>\n")
	}

	if inCodeBlock {
		sb.WriteString("</code></pre>\n")
	}
	if inList {
		sb.WriteString("</ul>\n")
	}
	if inTable {
		sb.WriteString("</table>\n")
	}

	return template.HTML(sb.String())
}

var (
	boldRe   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe = regexp.MustCompile(`\*(.+?)\*`)
	codeRe   = regexp.MustCompile("`([^`]+)`")
	linkRe   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

func formatInlineMarkdown(text string) string {
	escaped := html.EscapeString(text)
	// Replace backticks
	escaped = codeRe.ReplaceAllString(escaped, `<code>$1</code>`)
	// Replace bold
	escaped = boldRe.ReplaceAllString(escaped, `<strong>$1</strong>`)
	// Replace italic
	escaped = italicRe.ReplaceAllString(escaped, `<em>$1</em>`)
	// Replace links
	escaped = linkRe.ReplaceAllString(escaped, `<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>`)
	return escaped
}

// GenerateHTMLDashboard builds the standalone report.html from SummaryData and finding files
func GenerateHTMLDashboard(targetDir string, data *SummaryData) (string, error) {
	findingsDir := filepath.Join(targetDir, "findings")
	var detailed []DetailedFinding

	if entries, err := os.ReadDir(findingsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				continue
			}
			filePath := filepath.Join(findingsDir, e.Name())
			rawBytes, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			meta, err := ParseFindingFile(filePath)
			if err != nil {
				continue
			}

			detailed = append(detailed, DetailedFinding{
				Title:       meta.Title,
				Severity:    meta.Severity,
				CVSS:        meta.CVSS,
				CWE:         meta.CWE,
				Endpoint:    meta.Endpoint,
				FileName:    meta.FileName,
				HTMLContent: RenderMarkdownToHTML(string(rawBytes)),
			})
		}
	}

	sort.SliceStable(detailed, func(i, j int) bool {
		wI := severityWeights[detailed[i].Severity]
		wJ := severityWeights[detailed[j].Severity]
		if wI != wJ {
			return wI < wJ
		}
		return detailed[i].Title < detailed[j].Title
	})

	htmlData := HTMLReportData{
		Target:           data.Target,
		ScanTime:         data.ScanTime,
		TotalFindings:    data.TotalFindings,
		Stats:            data.SeveritySummary,
		Findings:         data.Findings,
		DetailedFindings: detailed,
		PoCs:             data.PoCs,
		EvidenceFiles:    data.EvidenceFiles,
	}
	if data.ReconNotes != nil {
		htmlData.ReconNotes = *data.ReconNotes
	}

	tmpl, err := template.New("report").Parse(reportHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML report template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, htmlData); err != nil {
		return "", fmt.Errorf("failed to execute HTML report template: %w", err)
	}

	htmlPath := filepath.Join(targetDir, "report.html")
	if err := os.WriteFile(htmlPath, buf.Bytes(), 0666); err != nil {
		return "", fmt.Errorf("failed to write report.html: %w", err)
	}

	return htmlPath, nil
}
