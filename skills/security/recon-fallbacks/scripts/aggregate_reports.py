#!/usr/bin/env python3
"""aggregate_reports.py — index findings/*.md into reports/<SLUG>/SUMMARY.md

Usage: python3 aggregate_reports.py <report_slug>
Expects layout: /root/reports/<slug>/findings/*.md with severity-prefixed
filenames (critical_|high_|medium_|low_) and '**Target:**'/'**Status:**' lines
in each finding. Sorts findings by severity, emits a numbered SUMMARY.md.
"""
import sys, os, glob, re

def main(slug):
    base = f"/root/reports/{slug}"
    fdir = os.path.join(base, "findings")
    if not os.path.isdir(fdir):
        print(f"no findings dir for {slug}"); return 1

    sev_rank = {"critical": 0, "high": 1, "medium": 2, "low": 3}
    entries = []
    for fp in sorted(glob.glob(os.path.join(fdir, "*.md"))):
        text = open(fp).read()
        fname = os.path.basename(fp)
        m = re.match(r"(critical|high|medium|low)_", fname)
        sev = m.group(1) if m else "unknown"
        title = re.search(r"^# (.+)$", text, re.M)
        target = re.search(r"\*\*Target:\*\* (.+)$", text, re.M)
        status = re.search(r"\*\*Status:\*\* (.+)$", text, re.M)
        entries.append((sev, title.group(1) if title else fname,
                        target.group(1) if target else "-",
                        status.group(1) if status else "-", fname))

    entries.sort(key=lambda e: sev_rank.get(e[0], 9))
    counts = {}
    for e in entries:
        counts[e[0]] = counts.get(e[0], 0) + 1

    lines = [f"# SUMMARY — {slug}", "",
             f"Findings total: {len(entries)} | " +
             " ".join(f"{k.upper()}:{v}" for k, v in
                      sorted(counts.items(), key=lambda x: sev_rank.get(x[0], 9))), ""]
    for i, (sev, title, target, status, fname) in enumerate(entries, 1):
        lines.append(f"{i}. **[{sev.upper()}]** {title}")
        lines.append(f"   - Target: {target} | Status: {status} | File: findings/{fname}")
    out = os.path.join(base, "SUMMARY.md")
    open(out, "w").write("\n".join(lines) + "\n")
    print(f"wrote {out} ({len(entries)} findings)")
    return 0

if __name__ == "__main__":
    sys.exit(main(sys.argv[1]))