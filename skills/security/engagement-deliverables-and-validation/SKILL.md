---
name: engagement-deliverables-and-validation
description: Workspace conventions for authorized security-engagement deliverables — canonical report location (/workspace/reports/<target>/), plain files not zip, follow-the-existing-format rule, report-aggregator overwrite trap, bypass-before-clear validation gate for negative claims, and Cloudflare UA pitfalls in PoC reproduction. Use at report-writing time or before declaring any control "secure".
---

# Engagement Deliverables & Validation Conventions

Workspace-specific conventions for authorized bug bounty / pentest engagements.
Complements (does not replace) `bb-methodology` and `bug-bounty-reporting-workflow`,
which are manually authored and hold the general methodology. This skill captures
the operator's standing corrections that those skills cannot be auto-updated with.

## 1. Deliverable Location & Format (hard user preference)

User corrections, verbatim: "taro di folder report bro kan udah ada contohnya" and
"kenapa bentuknya zip".

- **Canonical tree:** `/workspace/reports/<target>/` with subdirs `findings/`,
  `pocs/`, `evidence/`, plus `SUMMARY.md` and `metadata.json`.
- **Plain individual files ONLY — never zip.** Zips are for transport on explicit request only.
- **Mirror the house format** — list `/workspace/reports/*/` first and copy the structure
  of prior engagements (`xyrusrouter_xyz/`, `127_0_0_1_8888/`): emoji severity matrix table
  in SUMMARY.md; metadata.json keys = `target`, `scan_time`, `total_findings`,
  `severity_summary{CRITICAL,HIGH,MEDIUM,LOW,INFORMATIONAL}`, `findings[]`
  (file_name, relative_path, title, severity, cvss, cwe, endpoint, last_modified),
  `pocs[]`, `evidence_files[]`.
- Fallback if `/workspace` unwritable: `~/reports/<target>/` with the SAME tree shape.
- Write ONE canonical copy; `cp` (never independently edit) if the user wants it elsewhere.

### Aggregator overwrite trap
`/workspace/tools/aggregate_reports.py <target>` regenerates `SUMMARY.md` +
`metadata.json` from `findings/*.md` on every run and silently discards hand-written
analysis sections (bypass matrices, IDOR methodology, narrative). Order of operations:
run aggregator FIRST → then append analysis sections via targeted patches → do NOT
re-run the aggregator afterward without re-checking your sections survived.

### Complete-disclosure reporting (hard user requirement, comprehensive audit pass)
User corrections, verbatim: "jangan ada yang di sembunyikan ini untuk evaluasi" and
"kirim laporan final nya ya lengkap dengan endpoin dan lainnya".

- EVERY confirmed observation goes into the final report — no silent triage-pruning of
  weak/informational items. If something feels minor, grade it honestly and keep it;
  evaluation-grade means exhaustive.
- Exposed swagger/OpenAPI specs: reproduce the FULL endpoint table (method, path,
  parameters, request-body schema) into SUMMARY.md itself — "swagger UI exposed" alone
  is not a deliverable; the endpoint inventory is the value.
- Verbatim leaked secrets/app-keys belong in the report body AND a dedicated evidence
  file (redact only on explicit request).
- Deliver the final report as a `MEDIA:` attachment pointing at SUMMARY.md plus an
  inline condensed version in the chat message — never filesystem-only.
- Findings may be re-graded between draft and final (Medium→High after digging is fine),
  but never dropped: replace the old file; finding count only moves upward.

## 2. Bypass-Before-Clear Gate (for negative claims)

Operator expectation: "jangan ambigu menganggap hal itu tervalidasi" — never declare a
control secure from plain unauth probes alone. A negative claim ("rate-limit works",
"middleware protects /api/*") is report-grade only after its known bypass battery was
ATTEMPTED and recorded failing:

| Control | Minimum bypass battery |
|---|---|
| Middleware/auth gate | CVE-2025-29927 full variant set (`x-middleware-subrequest`: `middleware`, `src/middleware`, `pages/_middleware`, each ×5-chain); XFF internal IPs; `X-Original-URL`/`X-Rewrite-URL`; `X-Forwarded-Host`; path normalization (`//`, `/./`, `%2f`, case variants, matrix params) |
| Rate limiter | `X-Forwarded-For`/`X-Real-IP` spoofing. Spoofed `CF-Connecting-IP` → Cloudflare itself rejects (403 error code 1000): that is the WAF working, NOT a bypass success — classify correctly |
| Parser-level auth | duplicate JSON keys, array-typed values, NoSQL `$gt`, null/missing fields, unicode homoglyphs, form-encoded vs JSON vs text/plain content-type differentials |
| Cache isolation | static-extension probes on sensitive API paths + inspect `cf-cache-status` (BYPASS/DYNAMIC ⇒ not cacheable ⇒ WCD dead) |

Save every failed vector to `evidence/evidence_roundN_bypass_validation.txt` and surface
the matrix in SUMMARY.md — documented failed bypasses are coverage proof, not wasted requests.

## 3. Cloudflare UA Pitfall in Multi-Tool Reproduction

When reproducing/validating findings against Cloudflare-fronted targets:
- `python3 urllib/requests` with default UA often gets **403 (CF block)** while `curl`
  gets 200 on the same URL. This fabricates false tool-differentials — exactly what
  cross-tool reproduction is supposed to rule out.
- Fix: send a real browser UA in every scripted probe
  (`User-Agent: Mozilla/5.0 ... Chrome/126.0 Safari/537.36`) and keep it identical
  across curl and Python runs so responses are comparable.
- Also remember urllib string-slicing truncation breaks JSON parsing — read full bodies
  before `json.loads` in PoC scripts.

## 4. Layered Rate-Limiter Response Shapes (detection technique)

429 shapes can differ by REQUEST PATH, not just identity: JSON content-type hits the
app-layer limiter (verbose quarantine body w/ `retryAfter`), while `text/plain` bodies
or dot-segment paths (`/api/auth/./login`) hit a global edge/global layer (plain
`{"error":"Too many requests"}`). When auditing login endpoints, fingerprint BOTH
layers and flag inconsistent response shapes between them as CWE-204 info-leak
(retryAfter exposure aids brute-force timing) even when no bypass exists.

## Related skills
- `bb-methodology` — 5-phase workflow + discipline rules (manually authored).
- `bug-bounty-reporting-workflow` — language/validation/sensitive-data preferences (manually authored).
- `hunt-idor`, `triage-validation` — per-class hunting and pre-severity gates.
