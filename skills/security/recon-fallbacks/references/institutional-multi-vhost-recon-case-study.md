# Case Study: Multi-Vhost Reconnaissance & Self-Hosted Forge Audit

Concrete instances of the patterns in `recon-fallbacks/SKILL.md` §8. Maintained as reference detail; general methodology lives in `SKILL.md`.

## False-positive: Gitea "open registration"
- First-pass auto-detection matched the word "Register" in the page `<title>` of `/user/sign_up` → flagged MEDIUM "open registration".
- Manual body inspection showed the literal string: `Registration is disabled. Please contact your site administrator.`
- Correction chain that worked: grep claim-bearing text in raw HTML → rewrite finding to LOW (anonymous repo browsing only) → document the correction inside the finding file ("Catatan koreksi metodologi") → recapture evidence HTML so archives match final claims.

## Undercounted extraction: /explore/repos
- First regex (`href="/x/y"[^>]*class="...name..."`) → 3 repos.
- Corrected generic regex excluding nav prefixes → **11 repos**, including production source repositories (`inventory-be`, `inventory-fe`, `attendance-mobile`, `school-middleware`, `gradebook-fe`).
- Lesson: extraction regexes must be validated against the saved raw page before numbers enter a report.

## Environment quirks observed (fixes, not permanent constraints)
- `httpx` CLI present was actually the Python-lib stub (`Error: No such option '-s'`) → used python-httpx in-process prober.
- crt.sh returned HTTP 502 nginx error page twice (~40s apart) — retry once, then log as limitation.
- `zip` binary missing → `python3 -m zipfile` / `zipfile.ZipFile` walk produced the deliverable archive.
- `git` binary missing → Gitea `/archive/<branch>.tar.gz` download + tarfile in-memory grep fully replaced clone+grep for source audits.

## Non-standard ports found by the 57-port nmap sweep (why alt-port scanning pays)
- `203.0.113.45:8081` — PowerDNS Authoritative Server Monitor 4.5.3, unauthenticated: query stats, log-message ring with GET-based `?resetring=logmessages` action (CSRF/reset issue), remote client IPs. Fingerprint via `Server: PowerDNS/4.5.3`.
- `203.0.113.45:8888` — nmap labeled it "uvicorn http"; Phase 2 proved TLS 1.3 + FastAPI (`/docs`, `/redoc`, `/openapi.json` all 200; openapi `paths:{}` = empty skeleton app). See SKILL.md §8.7 — probe BOTH schemes on non-standard ports before labeling.

## Stack fingerprints confirmed this session (second redirect pass)
- Moodle: redirect to `/login/index.php`, theme dir `/theme/boost/`, `yui_combo.php` assets, `data-basename="boost"` on `<html>`.
- WordPress versions leak via `<meta name="generator" content="WordPress X.Y.Z">`; Elementor version likewise. User enumeration: `GET /wp-json/wp/v2/users` returns slug list when not hardened.
- Gitea asset version string pattern: `?v=13.0.4~gitea-1.22.0`.

---

# Phase 2: Source Code Audit & Configuration Dumps

## Gitea repo archive audit → High finding
- Downloaded all 11 repos via `/archive/main.tar.gz` (default branch confirmed from repo-page href), grepped every tar member in memory.
- Confirmed signatures (see SKILL.md §6b table): `prisma/seed.ts` with plaintext admin creds + `Role.ADMIN` (`admin_seed/SeedPass2026!`); static shared-secret auth (`if (token !== Key)` vs `process.env.SECRET_KEY`); internal IP `http://10.0.12.30/images/...` in upload service; dev `postgres/postgres` in compose.yaml.
- Frontend repository revealed candidate internal scan targets: `new-portal.apps.internal`, `access.apps.internal` (logged as internal scope).

## phpinfo exposure (Medium on multiple servers)
- `GET /info.php` 200 on apex (via Cloudflare) AND LMS origin — PHP 8.2 / 8.4, full config dumps archived as evidence HTML.
- Parser pitfall: `disable_functions` renders `<i>no value</i>` — a non-empty-string test mislabels it as "set". Match the literal page text.
- LMS phpinfo extras worth grabbing when present: `session.save_path=tcp://127.0.0.1:6379` (Redis session store), loaded module list (pdo_pgsql ⇒ PostgreSQL backend), `System:` line (hostname `lms-internal`, Linux kernel build date).

## PoC gate failures caught by running-before-shipping
- `tarfile.extractfile(m).read(errors=...)` → TypeError (BufferedReader.read takes no kwargs) — both §6b and §8.6.
- PoC regex for default branch was fine but initially hardcoded `main.tar.gz` without checking the repo page href — verify branch first.
- After fixes, subprocess-ran each PoC: exit 0 AND expected `[VULN]` markers present → then packaged deliverable.

## Negative results recorded (completeness discipline)
No `.env`/`.git`/backup/debug.log/server-status across 9 hosts × 22 common paths; WP debug.log absent; wpforms `.htaccess` 403; Moodle admin/install redirect to login; xmlrpc on apex blocked by Cloudflare (520) while origin serves methods — edge-only protection is itself a finding (origin reachable once real IP known).

---

# Phase 3: LMS Deep-Dive (see SKILL.md §9 for method)

## Version triangulation outcome
upgrade.txt floor ≥4.5 · UPGRADING.md top section `## 5.0.1+` (no 5.1 section) · environment.xml max MOODLE block **5.1** (first regex attempt matched `<LIBRARY version="19">` → garbage "19"; anchor `<MOODLE\s`) ⇒ deployed code era 5.0.x–5.1.x, pre-5.0.8/5.1.5 ⇒ relevant CVE batches apply (headline: report-builder capability gap = student-data leak path). Reported as CONFIRMED(version)/POTENTIAL(exploit-needs-account) — honest split.

## Anonymous matrix outcome
All content login-gated ✓ · signup.php 404 ✓ · forgot_password + token.php generic (hash-diff showed ONLY sesskey/random-token differences — no enum oracle) ✓ · admin/cron.php admin-blocked ✓ · admin/settings.php app-error-reachable (31 KB app body vs 146 B nginx 404 — size tells you which layer answered) ⚠ · config-dist.php executed → runtime Fatal leak ⚠ · composer.lock/package.json/environment.xml readable ⚠ · PHP sources served as 0-byte executions (safe).

## Rate-limit test outcome
8× POST login, fake user, 0.3 s interval → `[303,'invalid',303,'invalid',...]` alternating, zero throttle keywords → Medium "no rate-limit observed up to 8 attempts". Deliberately stopped at 8 (non-destructive bound); token.php noted as parallel surface. PoC script shipped after fixing the environment.xml regex; validated exit 0 + correct output before delivery.

## Deliverable flow that worked
findings .md per severity-prefix file → recon JSON findings[] updated with phase tags → aggregator regenerated SUMMARY.md → zipfile walk rebuilt single zip (integrity-checked via testzip) → deliverable output.