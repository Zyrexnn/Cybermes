#!/usr/bin/env python3
"""
Cybermes Automated PDF & HTML Security Report Generator
Transforms structured finding markdown files and metadata.json into an executive-grade PDF & HTML report.
Utilizes Playwright Chromium for pixel-perfect print layout and modern typography.
"""

import os
import re
import sys
import json
import argparse
from datetime import datetime
from pathlib import Path
import markdown

BASE_DIR = Path(__file__).resolve().parent.parent
REPORTS_DIR = BASE_DIR / "reports"

REPORT_HTML_TEMPLATE = """<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Security Assessment Report — {{ target }}</title>
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
    --accent: #6366f1;
    --code-bg: #090d16;
  }

  * { box-sizing: border-box; margin: 0; padding: 0; }
  
  body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
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

  /* Cover & Header */
  .header-card {
    border-bottom: 2px solid #e2e8f0;
    padding-bottom: 25px;
    margin-bottom: 30px;
  }

  .brand-tag {
    display: inline-block;
    background: #4f46e5;
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
  }

  .meta-item span {
    font-size: 14px;
    font-weight: 600;
    color: #0f172a;
  }

  /* Stats Pills */
  .stats-bar {
    display: flex;
    gap: 10px;
    margin: 25px 0;
  }

  .stat-pill {
    flex: 1;
    text-align: center;
    padding: 12px 10px;
    border-radius: 8px;
    color: #ffffff;
    font-weight: 700;
  }

  .stat-pill .num { font-size: 22px; display: block; }
  .stat-pill .lbl { font-size: 11px; text-transform: uppercase; opacity: 0.9; }

  .pill-crit { background: #dc2626; }
  .pill-high { background: #ea580c; }
  .pill-med { background: #d97706; }
  .pill-low { background: #2563eb; }
  .pill-info { background: #475569; }

  /* Sections */
  h2 {
    font-size: 18px;
    font-weight: 700;
    color: #0f172a;
    border-left: 4px solid #4f46e5;
    padding-left: 10px;
    margin: 35px 0 15px 0;
  }

  /* Tables */
  table {
    width: 100%;
    border-collapse: collapse;
    margin: 15px 0;
    font-size: 13px;
  }

  th, td {
    border: 1px solid #e2e8f0;
    padding: 10px 12px;
    text-align: left;
  }

  th {
    background: #f1f5f9;
    font-weight: 600;
    color: #334155;
  }

  tr:nth-child(even) td {
    background: #f8fafc;
  }

  /* Finding Item Card */
  .finding-card {
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    margin-bottom: 25px;
    overflow: hidden;
    page-break-inside: avoid;
  }

  .finding-header {
    background: #f8fafc;
    padding: 14px 18px;
    border-bottom: 1px solid #e2e8f0;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .finding-title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .badge {
    padding: 3px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
    color: #ffffff;
    text-transform: uppercase;
  }

  .badge-CRITICAL { background: #dc2626; }
  .badge-HIGH { background: #ea580c; }
  .badge-MEDIUM { background: #d97706; }
  .badge-LOW { background: #2563eb; }
  .badge-INFORMATIONAL, .badge-INFO { background: #475569; }

  .finding-body {
    padding: 18px;
  }

  .finding-body h1, .finding-body h2, .finding-body h3 {
    font-size: 14px;
    margin: 15px 0 8px 0;
    border-left: none;
    padding-left: 0;
    color: #0f172a;
  }

  .finding-body p {
    margin-bottom: 10px;
  }

  .finding-body ul, .finding-body ol {
    margin-left: 20px;
    margin-bottom: 12px;
  }

  /* Code blocks */
  pre {
    background: #0f172a;
    color: #e2e8f0;
    padding: 14px;
    border-radius: 6px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    overflow-x: auto;
    margin: 12px 0;
    page-break-inside: avoid;
    white-space: pre-wrap;
    word-break: break-all;
  }

  code {
    font-family: 'JetBrains Mono', monospace;
    font-size: 12px;
    background: #f1f5f9;
    color: #0f172a;
    padding: 2px 4px;
    border-radius: 3px;
  }

  pre code {
    background: transparent;
    color: inherit;
    padding: 0;
  }

  /* Print optimizations */
  @media print {
    body { background: #ffffff; font-size: 12pt; }
    .container { max-width: 100%; padding: 0; }
    .finding-card { page-break-inside: avoid; margin-bottom: 20px; }
    h2 { page-break-after: avoid; }
    pre { page-break-inside: avoid; }
  }
</style>
</head>
<body>
<div class="container">
  <!-- Header -->
  <div class="header-card">
    <div class="brand-tag">Cybermes Offensive Security Platform</div>
    <h1>Security Assessment & Vulnerability Report</h1>
    <p style="color: #64748b;">Target Asset: <strong>{{ target }}</strong></p>

    <div class="meta-grid">
      <div class="meta-item">
        <strong>Assessment Date</strong>
        <span>{{ scan_time }}</span>
      </div>
      <div class="meta-item">
        <strong>Total Confirmed Findings</strong>
        <span>{{ total_findings }} Flaws</span>
      </div>
      <div class="meta-item">
        <strong>Assessment Scope</strong>
        <span>Authorized Full-Surface</span>
      </div>
    </div>
  </div>

  <!-- Stats Breakdown -->
  <div class="stats-bar">
    <div class="stat-pill pill-crit">
      <span class="num">{{ stats.CRITICAL }}</span>
      <span class="lbl">Critical</span>
    </div>
    <div class="stat-pill pill-high">
      <span class="num">{{ stats.HIGH }}</span>
      <span class="lbl">High</span>
    </div>
    <div class="stat-pill pill-med">
      <span class="num">{{ stats.MEDIUM }}</span>
      <span class="lbl">Medium</span>
    </div>
    <div class="stat-pill pill-low">
      <span class="num">{{ stats.LOW }}</span>
      <span class="lbl">Low</span>
    </div>
    <div class="stat-pill pill-info">
      <span class="num">{{ stats.INFORMATIONAL }}</span>
      <span class="lbl">Info</span>
    </div>
  </div>

  <!-- Findings Matrix -->
  <h2>1. Executive Findings Matrix</h2>
  <table>
    <thead>
      <tr>
        <th>Severity</th>
        <th>Vulnerability Title</th>
        <th>CVSS</th>
        <th>CWE</th>
        <th>Affected Endpoint</th>
      </tr>
    </thead>
    <tbody>
      {% if findings %}
        {% for f in findings %}
        <tr>
          <td><span class="badge badge-{{ f.severity }}">{{ f.severity }}</span></td>
          <td><strong>{{ f.title }}</strong></td>
          <td><code>{{ f.cvss }}</code></td>
          <td><code>{{ f.cwe }}</code></td>
          <td><code>{{ f.endpoint }}</code></td>
        </tr>
        {% endfor %}
      {% else %}
        <tr><td colspan="5" style="text-align: center; color: #64748b;">No confirmed vulnerabilities reported.</td></tr>
      {% endif %}
    </tbody>
  </table>

  <!-- Detailed Vulnerability Chapters -->
  <h2>2. Detailed Technical Findings & Proof of Concepts</h2>
  {% if detailed_findings %}
    {% for df in detailed_findings %}
    <div class="finding-card">
      <div class="finding-header">
        <div class="finding-title">#{{ loop.index }} — {{ df.title }}</div>
        <span class="badge badge-{{ df.severity }}">{{ df.severity }}</span>
      </div>
      <div class="finding-body">
        {{ df.html_content | safe }}
      </div>
    </div>
    {% endfor %}
  {% else %}
    <p style="color: #64748b; font-style: italic;">No detailed vulnerability reports found.</p>
  {% endif %}

  <!-- PoC Scripts -->
  {% if pocs %}
  <h2>3. Standalone Verification PoC Scripts</h2>
  <ul>
    {% for poc in pocs %}
    <li style="margin-bottom: 6px;"><code>pocs/{{ poc }}</code></li>
    {% endfor %}
  </ul>
  {% endif %}

  <!-- Footer -->
  <div style="margin-top: 50px; border-top: 1px solid #e2e8f0; padding-top: 15px; font-size: 11px; color: #94a3b8; text-align: center;">
    Generated autonomously by <strong>Cybermes Security Intelligence</strong> · Confidential Assessment Deliverable
  </div>
</div>
</body>
</html>
"""

def clean_markdown_for_html(content: str) -> str:
    """Pre-process markdown content for clean HTML conversion."""
    # Convert markdown tables, codeblocks and lists
    md = markdown.Markdown(extensions=['extra', 'tables', 'fenced_code', 'nl2br'])
    return md.convert(content)

def generate_report_for_target(target_dir: Path, output_pdf: bool = True) -> tuple:
    """Render HTML and export PDF via Playwright for a specific target directory."""
    target_name = target_dir.name
    metadata_file = target_dir / "metadata.json"
    findings_dir = target_dir / "findings"
    
    metadata = {}
    if metadata_file.exists():
        try:
            metadata = json.loads(metadata_file.read_text(encoding="utf-8"))
        except Exception:
            pass

    stats = metadata.get("severity_summary", {"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0, "INFORMATIONAL": 0})
    findings = metadata.get("findings", [])
    pocs = metadata.get("pocs", [])
    scan_time = metadata.get("scan_time", datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
    total_findings = metadata.get("total_findings", len(findings))

    detailed_findings = []
    if findings_dir.exists():
        for f_path in sorted(findings_dir.glob("*.md")):
            if f_path.name.lower() in ("summary.md", "readme.md"):
                continue
            raw_text = f_path.read_text(encoding="utf-8", errors="replace")
            # Extract title & severity
            title_match = re.search(r"^#\s+(?:Vulnerability Report:\s*)?(.+)$", raw_text, re.MULTILINE)
            title = title_match.group(1).strip() if title_match else f_path.stem.replace("_", " ")
            
            sev_match = re.search(r"(?:Severity|Severity Rating)\s*[:|]\s*\*?`?([A-Za-z]+)`?\*?", raw_text, re.IGNORECASE)
            severity = sev_match.group(1).upper() if sev_match else "INFORMATIONAL"
            if severity == "INFO":
                severity = "INFORMATIONAL"

            # Remove first H1 to prevent duplicated title in card
            cleaned_text = re.sub(r"^#\s+.*$", "", raw_text, count=1, flags=re.MULTILINE).strip()
            html_body = clean_markdown_for_html(cleaned_text)

            detailed_findings.append({
                "title": title,
                "severity": severity,
                "html_content": html_body
            })

    final_report_file = target_dir / "FINAL_REPORT.md"
    if not detailed_findings and final_report_file.exists():
        raw_text = final_report_file.read_text(encoding="utf-8", errors="replace")
        html_body = clean_markdown_for_html(raw_text)
        detailed_findings.append({
            "title": "Full Assessment Narrative & Findings",
            "severity": "LOW" if stats.get("LOW", 0) > 0 else "INFORMATIONAL",
            "html_content": html_body
        })

    # Render template using Jinja2
    from jinja2 import Template
    template = Template(REPORT_HTML_TEMPLATE)
    html_rendered = template.render(
        target=target_name,
        scan_time=scan_time,
        total_findings=total_findings,
        stats=stats,
        findings=findings,
        detailed_findings=detailed_findings,
        pocs=pocs
    )

    html_file = target_dir / "report.html"
    html_file.write_text(html_rendered, encoding="utf-8")
    try:
        os.chmod(html_file, 0o666)
    except Exception:
        pass

    pdf_file = target_dir / "REPORT.pdf"

    if output_pdf:
        try:
            from playwright.sync_api import sync_playwright
            with sync_playwright() as p:
                browser = p.chromium.launch(
                    headless=True,
                    args=['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage']
                )
                page = browser.new_page()
                try:
                    page.set_content(html_rendered, wait_until="networkidle", timeout=15000)
                except Exception:
                    page.set_content(html_rendered, wait_until="load", timeout=10000)
                page.pdf(
                    path=str(pdf_file),
                    format="A4",
                    print_background=True,
                    margin={"top": "15mm", "bottom": "15mm", "left": "15mm", "right": "15mm"}
                )
                browser.close()
            try:
                os.chmod(pdf_file, 0o666)
            except Exception:
                pass
        except Exception as e:
            print(f"⚠️  PDF export via Playwright failed for [{target_name}]: {e}")

    return html_file, pdf_file

def main():
    parser = argparse.ArgumentParser(description="Cybermes PDF & HTML Report Generator")
    parser.add_argument("target", nargs="?", help="Target slug name (e.g. target_example_com)")
    parser.add_argument("--all", action="store_true", help="Generate reports for all targets")
    parser.add_argument("--no-pdf", action="store_true", help="Generate HTML only without rendering PDF")
    args = parser.parse_args()

    targets_to_process = []
    if args.all:
        for p in REPORTS_DIR.iterdir():
            if p.is_dir() and not p.name.startswith("."):
                targets_to_process.append(p)
    elif args.target:
        target_dir = REPORTS_DIR / args.target
        if target_dir.exists():
            targets_to_process.append(target_dir)
        else:
            print(f"❌ Target directory not found: {target_dir}")
            sys.exit(1)
    else:
        subdirs = [p for p in REPORTS_DIR.iterdir() if p.is_dir() and not p.name.startswith(".")]
        if subdirs:
            targets_to_process.extend(subdirs)
        else:
            print("No target directory found in reports/.")
            sys.exit(0)

    for t_dir in targets_to_process:
        html_f, pdf_f = generate_report_for_target(t_dir, output_pdf=not args.no_pdf)
        print(f"✓ Generated reports for [{t_dir.name}]:")
        print(f"   📄 HTML Dashboard: {html_f}")
        if not args.no_pdf and pdf_f.exists():
            print(f"   📑 Executive PDF : {pdf_f} ({pdf_f.stat().st_size / 1024:.1f} KB)")

if __name__ == "__main__":
    main()
