#!/usr/bin/env python3
"""
Cybermes Report Aggregator & Indexer
Parses findings in reports/<target>/findings/ and generates a unified SUMMARY.md and metadata.json.
Features resilient multi-format parsing (tables, lists, key-values, frontmatter) and custom section preservation.
"""

import os
import re
import sys
import json
import argparse
from datetime import datetime
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
REPORTS_DIR = BASE_DIR / "reports"

SEVERITY_ORDER = {
    "CRITICAL": 1,
    "HIGH": 2,
    "MEDIUM": 3,
    "LOW": 4,
    "INFORMATIONAL": 5,
    "UNKNOWN": 6,
}

def clean_extracted_value(val: str) -> str:
    """Clean markdown backticks, asterisks, and whitespace."""
    if not val:
        return ""
    return re.sub(r'[`\*_]', '', val).strip()

def parse_finding_file(filepath: Path) -> dict:
    """Parse a single finding markdown file to extract structured metadata."""
    content = filepath.read_text(encoding="utf-8", errors="replace")
    filename = filepath.name

    # 1. Extract Title
    title = None
    fm_match = re.search(r"^---\s*\n(.*?)\n---", content, re.DOTALL)
    if fm_match:
        fm_title = re.search(r"^title:\s*['\"]?(.+?)['\"]?$", fm_match.group(1), re.MULTILINE)
        if fm_title:
            title = fm_title.group(1).strip()

    if not title:
        title_match = re.search(r"^#\s+(?:(?:Vulnerability Report|Finding|Vuln):\s*)?(?:\[[A-Z]+\]\s*[-:]?\s*)?(?:[0-9]+\.\s*)?(.+)$", content, re.MULTILINE)
        if title_match:
            title = title_match.group(1).strip()
        else:
            title = filename.replace(".md", "").replace("_", " ").replace("-", " ")

    # 2. Extract Severity
    severity = None
    if fm_match:
        fm_sev = re.search(r"^severity:\s*['\"]?([A-Za-z]+)['\"]?$", fm_match.group(1), re.MULTILINE | re.IGNORECASE)
        if fm_sev:
            severity = fm_sev.group(1).strip().upper()

    if not severity:
        # Table format
        sev_match = re.search(r"\|\s*\*{0,2}(?:Severity|Severity Rating|Risk Level)\*{0,2}\s*\|\s*[`\*]?([A-Za-z]+)", content, re.IGNORECASE)
        if not sev_match:
            # Key-value format
            sev_match = re.search(r"(?:Severity|Severity Rating|Risk Level)\s*[:=]\s*[`\*]?([A-Za-z]+)", content, re.IGNORECASE)
        if sev_match:
            severity = sev_match.group(1).strip().upper()
        else:
            # Check filename prefix
            prefix_match = re.search(r"^(?:\[)?(CRITICAL|HIGH|MEDIUM|LOW|INFO|INFORMATIONAL)(?:\])?[-_]", filename, re.IGNORECASE)
            severity = prefix_match.group(1).upper() if prefix_match else "UNKNOWN"

    if severity in ("INFO", "INFORMATIONAL", "NOTE"):
        severity = "INFORMATIONAL"
    elif severity not in SEVERITY_ORDER:
        severity = "UNKNOWN"

    # 3. Extract CVSS
    cvss_match = re.search(r"\|\s*\*{0,2}CVSS(?:\s*v?3(?:\.1)?)?(?:\s*Score)?\*{0,2}\s*\|\s*[`\*]?([0-9\.]+(?:\s*\([^\)\|\n]+\))?)", content, re.IGNORECASE)
    if not cvss_match:
        cvss_match = re.search(r"CVSS(?:\s*v?3(?:\.1)?)?(?:\s*Score)?\s*[:=]\s*[`\*]?([0-9\.]+(?:\s*\([^\)\|\n]+\))?)", content, re.IGNORECASE)
    cvss = clean_extracted_value(cvss_match.group(1)) if cvss_match else "N/A"

    # 4. Extract CWE
    cwe_match = re.search(r"\|\s*\*{0,2}CWE\*{0,2}\s*\|\s*[`\*]?((?:CWE-)?\d+[^|*\n`]*)", content, re.IGNORECASE)
    if not cwe_match:
        cwe_match = re.search(r"CWE\s*[:=]\s*[`\*]?((?:CWE-)?\d+[^|*\n`]*)", content, re.IGNORECASE)
    if cwe_match:
        raw_cwe = clean_extracted_value(cwe_match.group(1))
        cwe = raw_cwe if raw_cwe.upper().startswith("CWE-") else f"CWE-{raw_cwe}"
    else:
        cwe = "N/A"

    # 5. Extract Affected Endpoint/Asset
    ep_match = re.search(r"\|\s*\*{0,2}(?:Affected Endpoint|Affected Asset|Target|Endpoint|URL/Host|URL)\*{0,2}\s*\|\s*[`\*]?([^|*\n`]+)", content, re.IGNORECASE)
    if not ep_match:
        ep_match = re.search(r"(?:Affected Endpoint|Affected Asset|Target|Endpoint|URL/Host|URL)\s*[:=]\s*[`\*]?([^|*\n`]+)", content, re.IGNORECASE)
    endpoint = clean_extracted_value(ep_match.group(1)) if ep_match else "N/A"

    return {
        "file_name": filename,
        "relative_path": f"findings/{filename}",
        "title": title,
        "severity": severity,
        "cvss": cvss,
        "cwe": cwe,
        "endpoint": endpoint,
        "last_modified": datetime.fromtimestamp(filepath.stat().st_mtime).strftime("%Y-%m-%d %H:%M:%S")
    }

def extract_custom_sections(existing_summary_content: str) -> str:
    """Preserve custom handwritten sections from an existing SUMMARY.md."""
    if not existing_summary_content:
        return ""
    
    custom_patterns = [
        r"(##\s+(?:Verified Working Controls|Executive Narrative|Recommendations|Priority Action Items|Attack Path Narrative).*)$"
    ]
    
    preserved = []
    for pat in custom_patterns:
        match = re.search(pat, existing_summary_content, re.DOTALL | re.IGNORECASE)
        if match:
            preserved.append(match.group(1).strip())
            
    return "\n\n---\n\n".join(preserved)

def aggregate_target(target_dir: Path) -> dict:
    """Aggregate findings for a given target directory."""
    findings_dir = target_dir / "findings"
    pocs_dir = target_dir / "pocs"
    evidence_dir = target_dir / "evidence"

    target_name = target_dir.name
    findings = []

    if findings_dir.exists():
        for f in sorted(findings_dir.glob("*.md")):
            if f.name.lower() in ("summary.md", "readme.md"):
                continue
            findings.append(parse_finding_file(f))

    # Sort findings by severity
    findings.sort(key=lambda x: (SEVERITY_ORDER.get(x["severity"], 99), x["title"]))

    # List pocs
    pocs = []
    if pocs_dir.exists():
        pocs = [p.name for p in sorted(pocs_dir.glob("*")) if p.is_file()]

    # List evidence
    evidence = []
    if evidence_dir.exists():
        evidence = [e.name for e in sorted(evidence_dir.glob("*")) if e.is_file()]

    # Count stats
    severity_counts = {"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0, "INFORMATIONAL": 0}
    for item in findings:
        sev = item["severity"]
        if sev in severity_counts:
            severity_counts[sev] += 1

    recon_notes = None
    recon_notes_file = evidence_dir / "recon_notes.md"
    if recon_notes_file.exists():
        recon_notes = "evidence/recon_notes.md"

    summary_data = {
        "target": target_name,
        "scan_time": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "total_findings": len(findings),
        "severity_summary": severity_counts,
        "findings": findings,
        "pocs": pocs,
        "evidence_files": evidence,
        "recon_notes": recon_notes
    }

    # Write metadata.json
    metadata_file = target_dir / "metadata.json"
    with open(metadata_file, "w", encoding="utf-8") as f:
        json.dump(summary_data, f, indent=2)
    try:
        os.chmod(metadata_file, 0o666)
    except Exception:
        pass

    # Check for existing SUMMARY.md to preserve custom notes
    summary_file = target_dir / "SUMMARY.md"
    existing_text = summary_file.read_text(encoding="utf-8", errors="replace") if summary_file.exists() else ""
    custom_content = extract_custom_sections(existing_text)

    # Generate SUMMARY.md
    generate_summary_md(target_dir, summary_data, custom_content)

    return summary_data

def generate_summary_md(target_dir: Path, data: dict, custom_content: str = ""):
    """Write a clean, professional SUMMARY.md document."""
    lines = [
        f"# 🛡️ Security Assessment Summary: `{data['target']}`",
        "",
        f"- **Generated At**: {data['scan_time']}",
        f"- **Total Confirmed Findings**: {data['total_findings']}",
        f"- **Severity Breakdown**: "
        f"🔴 Critical: {data['severity_summary']['CRITICAL']} | "
        f"🟠 High: {data['severity_summary']['HIGH']} | "
        f"🟡 Medium: {data['severity_summary']['MEDIUM']} | "
        f"🔵 Low: {data['severity_summary']['LOW']} | "
        f"⚪ Info: {data['severity_summary']['INFORMATIONAL']}",
        "",
        "---",
        "",
        "## 📑 Findings Matrix",
        "",
        "| Severity | Title / Vulnerability | CVSS v3.1 | CWE | Affected Endpoint | Report Link |",
        "| :--- | :--- | :--- | :--- | :--- | :--- |"
    ]

    if not data["findings"]:
        lines.append("| - | *No confirmed vulnerabilities reported* | - | - | - | - |")
    else:
        for f in data["findings"]:
            sev_badge = {
                "CRITICAL": "🔴 `CRITICAL`",
                "HIGH": "🟠 `HIGH`",
                "MEDIUM": "🟡 `MEDIUM`",
                "LOW": "🔵 `LOW`",
                "INFORMATIONAL": "⚪ `INFO`"
            }.get(f["severity"], f"`{f['severity']}`")

            link = f"[{f['file_name']}]({f['relative_path']})"
            lines.append(f"| {sev_badge} | {f['title']} | {f['cvss']} | {f['cwe']} | `{f['endpoint']}` | {link} |")

    lines.extend([
        "",
        "---",
        "",
        "## 🧪 Proof of Concept Scripts (`pocs/`)",
        ""
    ])

    if data["pocs"]:
        for poc in data["pocs"]:
            lines.append(f"- [`pocs/{poc}`](pocs/{poc})")
    else:
        lines.append("- *No standalone PoC scripts attached.*")

    lines.extend([
        "",
        "## 📁 Evidence & Recon Notes (`evidence/`)",
        ""
    ])

    if data.get("recon_notes"):
        lines.append(f"- 📝 **[Reconnaissance & Informational Notes]({data['recon_notes']})**")

    if data["evidence_files"]:
        for ev in data["evidence_files"]:
            lines.append(f"- [`evidence/{ev}`](evidence/{ev})")
    else:
        lines.append("- *No visual or trace evidence attached.*")

    if custom_content:
        lines.extend([
            "",
            "---",
            "",
            custom_content
        ])

    lines.append("")

    summary_file = target_dir / "SUMMARY.md"
    summary_file.write_text("\n".join(lines), encoding="utf-8")
    try:
        os.chmod(summary_file, 0o666)
    except Exception:
        pass

def main():
    parser = argparse.ArgumentParser(description="Cybermes Report Aggregator & Indexer")
    parser.add_argument("target", nargs="?", help="Target slug name (e.g. 127_0_0_1_8888 or example_com)")
    parser.add_argument("--all", action="store_true", help="Process all target directories under reports/")
    parser.add_argument("--no-pdf", action="store_true", help="Skip rendering HTML and PDF reports")
    args = parser.parse_args()

    if not REPORTS_DIR.exists():
        REPORTS_DIR.mkdir(parents=True, exist_ok=True)

    targets_to_process = []

    if args.all:
        for p in REPORTS_DIR.iterdir():
            if p.is_dir() and not p.name.startswith("."):
                targets_to_process.append(p)
    elif args.target:
        target_dir = REPORTS_DIR / args.target
        if not target_dir.exists():
            target_dir.mkdir(parents=True, exist_ok=True)
            (target_dir / "findings").mkdir(exist_ok=True)
            (target_dir / "pocs").mkdir(exist_ok=True)
            (target_dir / "evidence").mkdir(exist_ok=True)
        targets_to_process.append(target_dir)
    else:
        subdirs = [p for p in REPORTS_DIR.iterdir() if p.is_dir() and not p.name.startswith(".")]
        if subdirs:
            targets_to_process.extend(subdirs)
        else:
            print("No target directory specified and no existing target directories found.")
            sys.exit(0)

    pdf_generator = None
    if not args.no_pdf:
        try:
            sys.path.insert(0, str(BASE_DIR / "tools"))
            from generate_pdf import generate_report_for_target
            pdf_generator = generate_report_for_target
        except Exception as e:
            print(f"⚠️  PDF generator module not available: {e}")

    for t_dir in targets_to_process:
        res = aggregate_target(t_dir)
        print(f"✓ Aggregated reports for target [{t_dir.name}]: {res['total_findings']} findings indexed -> {t_dir / 'SUMMARY.md'}")
        if pdf_generator:
            try:
                html_f, pdf_f = pdf_generator(t_dir, output_pdf=True)
                if pdf_f.exists():
                    print(f"   📑 Rendered PDF: {pdf_f.name} ({pdf_f.stat().st_size / 1024:.1f} KB)")
            except Exception as pe:
                print(f"   ⚠️  Failed to render PDF: {pe}")

if __name__ == "__main__":
    main()
