---
name: stealth-web-recon
description: WAF-aware, rate-limited external recon and web asset discovery for bug bounty / authorized pentest targets. Avoids triggering Cloudflare / openresty / generic WAF blocks while still mapping subdomains, panels, and tech stack.
---

# Stealth Web Recon — WAF-Aware Asset Discovery

## Purpose
Map a target's external attack surface (subdomains, live web services, exposed panels, tech stack) WITHOUT getting blocked by Cloudflare, openresty rate limits, or other WAFs. The user explicitly requires blocks to be avoided — treat "don't trigger the WAF" as a hard constraint, not a nice-to-have.

This is a companion to the `recon` skill: use `recon` for the generic flow, this skill for the STEALTH discipline and command set that keeps you under WAF radar.

## Core principle: passive first, active last, never burst
1. **Passive enumeration** (crt.sh, subfinder, gau/Wayback) gathers hosts with zero requests to the target.
2. **DNS resolution** (dnsx) is borderline-passive — uses public resolvers, but space the queries.
3. **Active HTTP probing** (httpx) is the only step that "hits" the target — keep it slow and rate-limited.
4. **Content scanning** (katana/ffuf/arjun) is skipped entirely on WAF/Bot-Management hosts.

## Workflow (worked, zero-block sequence)
See `references/stealth-recon.md` for the exact command set and `references/katana_pitfalls.md` for katana crawl gotchas. Summary:

1. **Subdomain discovery (passive):**
   `subfinder -d <target> -silent -recursive` → then crt.sh JSON (external, may rate-limit — retry once or discard).
2. **DNS resolve:** `dnsx -l subfinder.txt -silent -r 8.8.8.8,1.1.1.1`.
3. **httpx probe (THE active step, rate-limited):**
   `httpx -l resolved.txt -silent -t 30 -rate-limit 15 -timeout 8 -ports 80,443 -title -status-code -tech-detect -server -follow-redirects`
   - `-rate-limit 15` and `-t 30` are mandatory. Never run httpx at default concurrency against a WAF host.
   - `-tech-detect` is what reveals Cloudflare Bot Management so you can exclude those hosts.
4. **Passive URL mining:** `gau --subs <target>` (Wayback, external — empty output is NORMAL, not a failure).
5. **Panel version confirmation:** ONE slow `curl --max-time 15 -A "Mozilla/5.0"` per host, sequential, `sleep 1` between — never a burst.

## WAF-exclusion rule (critical)
If httpx tech-detect shows `Cloudflare` + `Cloudflare Bot Management`, do **header-only / passive checks** on that host. Do NOT run katana/ffuf/arjun against it. Bot Management is the fastest path to a block. (In the nurulfikri session: `admisi`, `asset`, `aset`, `lms-ddp` were Bot-Managed and excluded from content scanning.)

## Abort / back-off signals
- HTTP `429` or a JS challenge page → STOP active scanning on that host, fall back to passive-only.
- External tools (crt.sh, gau) returning empty → non-critical, discard. Never compensate by hammering the target.

## Crawl pitfalls (cost real wasted cycles — verified)
- **katana `-silent` discards ALL results.** In v1.7.0, `-silent` routes crawl output to **stderr**; `katana ... -silent > file 2>/dev/null` yields a 0-byte file (exit 0, looks fine). Fix: use katana's own `-o file` (no `-silent`), OR drop `-silent` and capture stdout (`> out.txt 2>err.txt`). Full repro in `references/katana_pitfalls.md`.
- **Headless katana empty.** `-jc -kf all` needs Chromium; absent → exit 0, 0 endpoints. Use standard HTTP crawl (`-d 2 -rate-limit 8 -delay 1000`) unless Chromium is confirmed present.
- **Background katana SIGINT truncation.** A background `katana ... > file` whose parent shell gets SIGINT truncates output ("Ctrl+C pressed" in stderr). Wrap as `setsid bash -c 'katana ... > file 2>err.txt; echo DONE >> file'` so the child survives the parent signal group.
- Recover subdomains missed by subfinder from crawl links (Moodle/OJS/SaaS hosts referenced inside page HTML, e.g. `elena`, `lppm`, `journal`).

## Read-only default-credential validation (extends F-01 panel exposure)
When a panel is publicly reachable (HTTP 200 login), validate auth strength WITHOUT brute force:
- Max **3 common pairs** per panel via its REAL login endpoint (form POST for Adminer/Proxmox/NPM; JSON `{"email"/"username","password"}` for n8n/Kuma/DocuSeal).
- Inspect **body**, not just HTTP code (login failures often still return 200 with a re-rendered form). Markers: 401/400/"invalid"/"Invalid"/re-rendered login form = fail; a session/auth cookie or dashboard HTML = success.
- A `404` on the login path means the route is wrong (CSRF/session needed), NOT a success — report as inconclusive, do not keep guessing paths (that drifts into brute force).
- If default creds succeed → report Critical and STOP (no post-auth exploitation). If all fail → F-01 stays High (public exposure), not Critical.
- Never use enumerated usernames (e.g. from WP `/wp-json/wp/v2/users`) to attack other login surfaces — that is spraying, out of scope.
- Exposed admin/automation panels without visible auth: Adminer, n8n, Uptime Kuma, Nginx Proxy Manager, DocuSeal, Proxmox, Portainer, Nessus.
- `nginx/1.14.0 (Ubuntu)` default pages = server version leak + likely EOL.
- 502/500 on `*.dev-app.*` and staging hosts = dead dev infrastructure exposed publicly.
- Missing security headers (X-Frame-Options, X-Content-Type-Options, CSP, Referrer-Policy) even when HSTS is present.
- WordPress version / readme.html leaks.

## CVE version-matching (closes the audit loop)
After mapping versions, check them against known CVEs WITHOUT false positives — NVD 2.0 parsing + CPE affected-range check + sibling-plugin filtering. This is what turns "we saw a plugin version" into a credible (or cleared) finding. Full technique: `references/cve-version-matching.md`. Key traps: WPScan now needs a token; a CVE for "Theme My Login **2FA**" ≠ `theme-my-login`; "Review**X**" ≠ "Review Schema".

## Pitfalls
- Versions often DON'T leak in HTML (title-only) — don't burn requests forcing it.
- Bursting version probes (parallel curls) trips openresty rate limits fast — always sequential + sleep.
- Don't confuse "gau/crt.sh empty" with "target blocked" — those are external services.
