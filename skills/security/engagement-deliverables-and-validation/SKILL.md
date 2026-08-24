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
- **Default = plain individual files; zips only when the operator asks.** EXCEPTION confirmed 2026-08-22: this operator's STANDING preference is ONE zip containing every report folder (`findings/`, `pocs/`, `evidence/`) plus recon JSON — check user profile/memory first and follow it; both delivery shapes are legitimate across engagements, never refuse a zip citing this rule alone.
- **Mirror the house format** — list `/workspace/reports/*/` first and copy the structure
  of prior engagements (`target_example_com/`, `127_0_0_1_8888/`): emoji severity matrix table
  in SUMMARY.md; metadata.json keys = `target`, `scan_time`, `total_findings`,
  `severity_summary{CRITICAL,HIGH,MEDIUM,LOW,INFORMATIONAL}`, `findings[]`
  (file_name, relative_path, title, severity, cvss, cwe, endpoint, last_modified),
  `pocs[]`, `evidence_files[]`.
- Fallback if `/workspace` unwritable: `~/reports/<target>/` with the SAME tree shape.
- Write ONE canonical copy; `cp` (never independently edit) if the user wants it elsewhere.

### Aggregator overwrite trap
`aggregate_reports <target>` regenerates `SUMMARY.md` +
`metadata.json` from `findings/*.md` on every run while preserving custom
analysis sections (bypass matrices, IDOR methodology, narrative). Order of operations:
run aggregator FIRST → then append analysis sections via targeted patches.

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

### 3b. PoC Validation Gate (run every deliverable PoC; verified 2026-08-22)

A PoC script is not done when written — it is done when executed and its output matches
the finding file. In one engagement BOTH new PoCs shipped broken on first run:

- `tarfile.extractfile(m).read(errors=...)` → `TypeError: BufferedReader.read() takes no
  keyword arguments`. Fix: `.read()` bare, then `.decode("utf-8", "ignore")` separately.
- Regex label logic inverted: phpinfo's `disable_functions` renders as `<i>no value</i>`,
  so a naive "non-empty = set" check reported KOSONG values as "terisi" (set). Match the
  literal page text ("no value"), don't infer from emptiness.
- Archive/branch URLs: verify the default branch from the repo page href before hardcoding
  `main.tar.gz`.

Gate: run each PoC via subprocess → require exit 0 AND ≥1 `[VULN]`/expected marker line
AND output consistent with the finding severity → only then package/deliver. A PoC that
prints nothing is a failed PoC, not a quiet pass.

### 3c. Auditor-Side Artifact Traps (verify raw data before claiming server behavior)

**Trap 1 — your own output redaction mistaken for server-side masking.** Secret-shaped
strings (`sk-…`, tokens, API keys) are often redacted by the agent toolchain's output
sanitizer BEFORE you see them. In one engagement every login response displayed the
session token as `sk-853...d8a9`; it authenticated both console and paid inference
gateway, so it looked like a ~28-bit masked credential accepted server-side (fake
Critical). Raw-file inspection (`len()` + repr of the on-disk token) showed the true
51-char `sk-`+48-hex value (~183 bits): the masking was the AUDITOR pipeline, not the
server. Rule: any claim about token format/masking/matching/entropy requires inspecting
the raw bytes on disk (repr + length), never the rendered tool output.

**Trap 2 — SPA catch-all fallback mistaken for an endpoint/SSRF.** ffuf hits on SPA
hosts can be the index-fallback route, not a feature: `/render/<url>` returned 307
(Go path normalization) then 200 with the app shell for external, internal, AND loopback
targets — looked like an open SSRF proxy, killed after diffing bodies against
index.html baseline (byte size + `<title>`). Rule: before claiming ANY discovered path
does something server-side, compare its response body to the app's index.html baseline.

## 4. Layered Rate-Limiter Response Shapes (detection technique)

429 shapes can differ by REQUEST PATH, not just identity: JSON content-type hits the
app-layer limiter (verbose quarantine body w/ `retryAfter`), while `text/plain` bodies
or dot-segment paths (`/api/auth/./login`) hit a global edge/global layer (plain
`{"error":"Too many requests"}`). When auditing login endpoints, fingerprint BOTH
layers and flag inconsistent response shapes between them as CWE-204 info-leak
(retryAfter exposure aids brute-force timing) even when no bypass exists.

### Rate-limit cooldowns: preserve the session, not the login (verified 2026-08-23)

When the target rate-limits login aggressively (429 after ~2 failures, minute-scale
cooldown), do NOT re-login for every test batch — you burn the session on cooldown
waits. Pattern: login ONCE → persist the bearer token to a file immediately → reuse
`Authorization: Bearer $(cat .tok)` across all later batches; re-authenticate only on
real token expiry (401 on /me). Put `sleep 120` INSIDE the probe call with a raised
timeout instead of separate sleep calls, keep batches small, and exclude the token
file from any delivered archive (skip `*.tok` when zipping).

## 5. Leaked-Credential Impact Validation (the "bisa login ga?" follow-up)

Operator follow-up pattern (verified 2026-08-22): after receiving findings, the operator
asks "itu db yang ke expos emang beneran bisa login?" and asserts "harusnya dah fix".
Findings must pre-answer these so the follow-up never requires new work:

- **Map every leaked credential to its usable surface BEFORE delivery**: which live URL,
  which header/parameter format (`Authorization: Bearer <key>` vs form login), which
  endpoint proves it. Prove it live (read-only GET pair: without-auth 401 vs with-key 200)
  rather than asserting from source code alone.
- **If the app behind leaked accounts has no live deployment, verify that too** — resolve
  candidate hostnames (`absensi.`, `presensi.` — subfinder misses Cloudflare-hosted
  records; try DNS directly), probe `/login`, `/api`, `/admin`, and report the actual
  state ("app under construction, 404s") instead of speculating about impact.
- **Re-check every leak URL at report/delivery time** and record the live status in the
  finding: a file "removed in a later commit" is NOT remediation while
  `/raw/commit/<sha>^/<path>` still returns 200. Operators assume fixes happened
  ("harusnya dah fix"); the report must state verified still-public/still-valid status.
- For DB credentials: check external exposure of the DB port (nmap) and say plainly
  whether the credential is reachable from outside or only post-foothold (localhost
  `DATABASE_URL` = not externally loginable). Distinguish "leaked" from "externally
  usable" in the finding itself.
- Zip delivery mechanics (standing preference): rebuild the zip FRESH from disk at
  delivery time (never ship a stale archive), include a `MANIFEST.txt` of the full file
  list, and verify with `zipfile.testzip()` before sending.

## 6. Post-Engagement Cleanup & Late Background Output (verified 2026-08-22)

### Cleanup (hard user preference, verbatim: "abis ngapa-ngapain cache nya di bersihkan")
After the report is delivered:
1. Kill background scanners (nuclei/ffuf/sqlmap). Their completion notifications still arrive AFTER the kill — exit 143/SIGTERM is the intentional kill, not an error. On any late notification, re-verify deliverable integrity (zip entries + `testzip()`, disk state) before reacting.
2. Delete: recon working data, ad-hoc tools/wordlists, `/tmp` scratch files, downloaded template repos (nuclei-templates), and tool caches (pip, playwright, browser artifacts). Keep ONLY the canonical reports tree + final zip.
3. If anything was edited after the zip was built (e.g. a late lead added to recon_notes), rebuild the zip FRESH from disk and re-send — the delivered archive must never drift from disk state.

### Late background output may contain new leads
A scanner finishing late can deliver results after the zip shipped — e.g. ffuf matches for `/.env`, `/dump.sql`, `/server-status` arriving post-delivery. Do NOT ignore or hand-wave them:
1. Manually verify each lead with curl (uniform 403 + identical body sizes across paths ⇒ WAF block page ⇒ negative; unique size/body ⇒ fetch content before claiming exposure).
2. Append verified results — positive OR negative — to `evidence/recon_notes.md`.
3. Rebuild + re-send the zip, then tell the operator whether the verdict changed.

## Related skills
- `bb-methodology` — 5-phase workflow + discipline rules (manually authored).
- `bug-bounty-reporting-workflow` — language/validation/sensitive-data preferences (manually authored).
- `hunt-idor`, `triage-validation` — per-class hunting and pre-severity gates.
- `fuzzing-and-content-discovery` — ffuf workflow; its WAF uniform-403 false-positive pattern pairs with §6 late-lead verification. Session detail (ffuf flag compatibility, uniform-403 decision rule, Swagger `/docs` follow-up, scanner lifecycle): `references/ffuf-waf-false-positives.md`.
