# Independent Re-Validation Standard (anti-inflation)

Companion to the `bug-bounty-reporting-workflow` skill. Use this mode when the user says
"validate ulang", "re-validate", "jangan percaya laporan sebelumnya", "independent check", or
asks to re-test all prior findings with fresh evidence and no trust in prior severities.

## Mindset
The default build-and-report flow *accumulates* findings. Re-validation is the opposite: it
*tears down* inflated claims. Your job is to produce FEWER, fully-defensible findings. The user
explicitly prefers "sedikit finding tapi valid daripada banyak false positive".

## Verdict vocabulary (use exactly)
| Label | Meaning |
|-------|---------|
| CONFIRMED | Reproduced with a fresh request + concrete response indicator. |
| PROBABLE | Reproduced, but impact needs an unproven condition. |
| POTENTIAL | Behavior observed, exploitability unproven (e.g. outdated SW, no PoC). |
| INFORMATIONAL | Real behavior, zero security impact alone (endpoint names, public client_id, public dev host). |
| FALSE POSITIVE | Artifact mistaken for vuln (SPA HTML fallback ≠ auth bypass; CSRF token ≠ vuln). |
| NOT REPRODUCIBLE | Cannot reproduce now (WAF/000/rate-limit/patched). Do NOT submit. |
| STALE | Was true earlier, no longer. |
| DUPLICATE | Same root cause as another finding (e.g. CORS on sibling hosts = one finding). |

## Anti-inflation rules (the explicit round-3 corrections)
1. **CORS `*` without `Allow-Credentials` = Medium, NOT Critical.** No cookie/auth theft.
   Always grep `access-control-allow-credentials` after any CORS header — its presence/absence
   is the #1 severity driver.
2. **CSRF token = protection, never a vuln.** Django admin "exposed" = public reachable login
   (brute-force risk if no rate-limit/MFA). Don't score the token as the issue.
3. **Outdated SW ≠ Critical.** Verify which CVEs are applicable AND unauthenticated-exploitable.
   If most need auth / no public RCE → Medium (outdated software), not High.
4. **API discovery with all-401 = INFORMATIONAL.** Endpoint names are not a vulnerability.
5. **Public dev/uat host = INFORMATIONAL alone.** Needs demonstrated impact (weaker auth, debug
   traces, real PII, prod-DB reachability).
6. **OAuth `client_id` is public by spec.** Only a finding if `redirect_uri` is open-redirectable
   or `state`/`nonce` missing. PKCE present = good.
7. **SPA 200 vs API 200.** `/api/*` returning `<!DOCTYPE html>` is a routing fallback, not an
   auth bypass. Check `content-type` + first bytes before claiming missing-authZ.

## Per-finding re-test procedure
1. Fresh HTTP probe from scratch — never cite the old report as evidence.
2. Capture exact URL, request (headers/origin), response (status, headers, body sample).
3. Assign verdict from vocabulary above.
4. Recompute CVSS honestly (AC:H / UI:R / C:L/I:L when exploitation is conditional).
5. Separate CONFIRMED FACTS vs HYPOTHETICAL IMPACT explicitly.

## Deliverable format
File: `reports/FINAL-REVALIDATION-REPORT.md`
- Top: summary table `| # | Original Finding | Re-Test Status | New Severity | Evidence | Submit? | Reason |`
- Severity-Inflation-Corrections block (old → new per finding)
- Final split: SUBMIT (kept) vs DON'T SUBMIT (removed/downgraded) with reasons
- Evidence files saved to a `revalidation/` dir, one `.txt` per finding

## Worked example: monash.edu round-3 (2026-08-17)
11 prior findings → 5 kept (all Medium/Low), 0 Critical/High survived.

| # | Finding | Verdict | New Sev | Submit? |
|---|---------|---------|---------|---------|
| 1 | CORS static origin (admin-forms-dev) | CONFIRMED | Medium 5.3 | YES |
| 2 | Django admin exposed | CONFIRMED | Medium 5.9 | YES |
| 3 | WordPress 5.0.1 | CONFIRMED (httpx tech-detect) | Medium 5.3 | YES |
| 4 | Directory listing /static/ | CONFIRMED | Low 3.1 | YES |
| 5 | .env.example | NOT REPRODUCIBLE (HTTP 000, was 466b) | - | NO |
| 6 | 28 API endpoints | INFORMATIONAL (all 401) | Info | NO |
| 7 | 30 dev/uat hosts | INFORMATIONAL | Info | NO |
| 8 | CORS wildcard `*` Django | CONFIRMED (no credentials flag) | Medium 5.3 | YES |
| 9 | CORS on 4 hosts | DUPLICATE of #1 | - | NO (merge) |
| 10 | Okta client_id | INFORMATIONAL (public by spec) | Info | NO |
| 11 | Banner authZ | FALSE POSITIVE (SPA fallback HTML) | - | NO |

Severity corrections: #1 High→Medium, #2 High→Medium, #3 High→Medium, #8 Critical→Medium.
Evidence that drove corrections:
- CORS #1 & #8: header present but NO `access-control-allow-credentials: true` on any response.
- #3: httpx tech-detect returned `WordPress:5.0.1, PHP:7.2.24`; CVEs (CVE-2019-8942/8943/9787/16217)
  mostly require auth; no public unauthenticated RCE → Medium, not High.
- #10: `redirect_uri` fixed (not open-redirectable), PKCE S256 enabled → no exploit.
- #11: `content-type: text/html`, body `<!DOCTYPE html><div id="root">` → SPA shell, not JSON.
