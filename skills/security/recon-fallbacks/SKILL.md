---
name: recon-fallbacks
description: Class-level playbook for external recon when the standard Go/unix toolchain is unavailable — Python-equivalent probing, DNS resolution, port scanning, CMS/git-server fingerprinting, anonymous repo enumeration, and verified finding write-ups. Use alongside or instead of recon/web2-recon pipelines when binaries are missing or probing needs custom logic.
---

# RECON FALLBACKS — no-binary recon & verified reporting

## When to load
- `dnsx` / `httpx` / `naabu` / `dig` missing but Python3 present.
- Probing needs per-host logic the CLI tools can't express (redirect chains, generator meta, cookie names).
- Writing findings from recon output where auto-detection must be verified before claiming.

## 1. Identify WHICH httpx you have
`command -v httpx` succeeding proves nothing — the Python lib ships a CLI stub that dies on PD flags (`Error: No such option '-s'`). Test: `httpx -version >/dev/null 2>&1 && echo PD || echo not-PD`. If not-PD but `python3 -c "import httpx"` works, go in-process (§2).

## 2. Python probe matrix (httpx lib)
```python
import httpx, re, json, socket
client = httpx.Client(verify=False, timeout=8, follow_redirects=False,
                      headers={"User-Agent": "Mozilla/5.0 ReconAgent/1.0"})
out = []
for h in hosts:
    ip = None
    try: ip = socket.gethostbyname(h)
    except socket.gaierror: pass
    for url in (f"https://{h}:443/", f"http://{h}:80/"):
        try:
            r = client.get(url)
            m = re.search(r"<title[^>]*>(.*?)</title>", r.text, re.S|re.I)
            out.append({"url": url, "ip": ip, "status": r.status_code,
                        "server": r.headers.get("server"),
                        "location": r.headers.get("location"),
                        "sec_headers": {k: r.headers.get(k) for k in
                          ("strict-transport-security","content-security-policy",
                           "x-frame-options","x-content-type-options") if k in r.headers},
                        "title": m.group(1).strip()[:100] if m else None})
        except Exception as e:
            out.append({"url": url, "ip": ip, "error": type(e).__name__})
json.dump(out, open("probe.json","w"), indent=2)
```
Save this JSON — it doubles as the security-header matrix evidence.

## 3. DNS without dnsx/dig
```python
name, _aliases, ips = socket.gethostbyname_ex(host)  # gaierror = NXDOMAIN
cname = name if name != host else None   # canonical name reveals CNAME targets
```

## 4. Port scan without naabu
nmap one-liner, 57 common+non-standard web ports:
```
nmap -Pn -n -sV --open -T4 -p 80,81,443,444,591,981,1311,2480,3000,3128,4443,4567,5000,5800,7000,7001,8000,8001,8008,8010,8042,8080,8081,8083,8088,8090,8091,8118,8123,8181,8222,8243,8280,8333,8403,8443,8500,8800,8834,8843,8879,8888,8983,9000,9001,9043,9060,9080,9090,9091,9200,9443,9502,9800,10000,16080,18091 -oA nmap_webports <IPs...>
```
State clearly in the report: 57 ports scanned, NOT a full sweep.

## 5. Second-pass redirect fingerprinting
Re-probe every 301/303 with `follow_redirects=True`; capture final URL, `<meta generator>`, Set-Cookie names. This pass is what unmasks Moodle (`/login/index.php`, theme boost), WordPress (generator meta), Gitea (`~gitea-x.y.z` asset version strings), Next.js (`/_next/static/chunks`). Generic titles like "wp" or empty SPA shells hide these until the second pass.

## 5b. Catch-all sites: baseline-filtered discovery + inline-script endpoint mining

Laravel/WordPress sites often answer ANY unknown path with **HTTP 200 + full homepage**, which
makes status-code fuzzing useless and buries real endpoints. Pattern proven on a government
Laravel site (2026-08):

1. **Baseline first**: GET `/xq7z-nope-123`, record byte size + Content-Type (there: ~56.7KB
   HTML). Filter discovery results by size distance (`abs(size - baseline) > ~2000B`) AND
   Content-Type (JSON/XML/plain vs baseline HTML), never by status alone. A uniform-size 403
   across artifact paths (`composer.lock`, `artisan`, `.htaccess`) = Apache block rules WORKING
   — log as a positive control, not noise.
2. **Sibling-path probing multiplies confirmed findings**: one broken route already known to
   trigger a debug page (`/<module>/login` → 500) ⇒ probe plausible siblings
   (`/<module>/{admin,dashboard,upload,export}`) — one pass turned 1 debug-page trigger into 7
   and widened an existing Medium finding. Asset manifests (`/mix-manifest.json`) may themselves
   be catch-all fakes returning homepage HTML; verify against real asset URLs before trusting.
3. **Inline `<script>` blocks beat downloaded bundles on server-rendered pages**: bundles are
   usually public template libraries (clean), while homepage inline scripts carry the LIVE AJAX
   surface. Extract and mine them:
   ```python
   scripts = re.findall(r'<script(?![^>]*src=)[^>]*>(.*?)</script>', html, re.S)
   ```
   Look for `$.ajax(`/`fetch(`/`axios.` calls and string-concatenated route fragments
   (`'/polling/' + id + '/question'`). Verify candidates with a bounded READ-ONLY ID sweep
   (`/{1..10}`): real endpoints return JSON bodies with distinct sizes; catch-all returns
   homepage HTML. Endpoints recovered from the site's own code beat wordlist guesses.
4. **Reflection-check false positives**: template inline SVG icons make naive "payload echoed"
   greps report XSS where none exists. Probe each dangerous char class separately (`" ' < >`)
   and inspect WHERE they land — raw quote/angle inside an attribute or script context is
   exploitable; entity-encoded (`&quot;`, `&lt;`) is Blade/AutoEscape working.

## 6. Gitea/Forgejo anonymous repo enumeration — TWO tiers, use both
`/explore/repos` scraping UNDERCOUNTS (one engagement: scrape showed 3 repos, API showed 11).
- **Tier 1 (complete):** `GET /api/v1/repos/search?q=&limit=50`, paginate `&page=N` until empty → `full_name` list. No auth needed on public instances.
- **Tier 2 (fallback):** `/explore/repos` scrape, extract owner/repo hrefs EXCLUDING nav paths:
```python
repos = [x for x in sorted(set(re.findall(r'href="/([\w.\-]+/[\w.\-]+)"', body)))
         if not x.startswith(("explore/","user/","assets/","api/"))]
```
Registration status: trust ONLY the literal string `"Registration is disabled"` in `/user/sign_up` body — see §8 pitfall.

### 6b. Source audit without a git binary
Default branch: parse the `/OWNER/REPO/archive/<branch>.tar.gz` href from the repo page HTML — the defaultBranch JSON key is often ABSENT from server-rendered pages, don't assume `main`. Then grep every tar member:
```python
raw = tf.extractfile(m).read()          # BufferedReader.read() takes NO kwargs (errors= → TypeError)
text = raw.decode("utf-8", "ignore")
```
High-yield signatures (all confirmed in one school/SMB codebase engagement):

| Signature | Where | Meaning |
|---|---|---|
| `username:` + plaintext `password:` near `Role.ADMIN` | `prisma/seed.ts`, `src/db/factories/*` | Seeded creds may equal prod (seeds run at deploy) — state both possibilities honestly |
| `if (token !== Key)` string compare vs `process.env.SECRET_KEY` | auth middleware | Static shared-secret API auth — one leak opens all endpoints |
| `http://10.x.x.x/...` asset/service URLs | services | Internal network layout disclosure |
| `POSTGRES_PASSWORD: postgres` / `DATABASE_URL=postgresql://...` | compose.yaml | Dev DB topology |
| API base URLs / sibling subdomains in FE repos | Next.js/Vue lib clients | New scan targets feeding back into recon scope |

Severity calibration: public ADMIN credentials = **High** (even with prod-reuse unconfirmed — say so explicitly); anonymous browsing alone with closed registration = **Low**. Repo names ↔ subdomain mapping (bukunilai-fe ↔ gradebook.apps.*) gives the chain story.

## 7. Report packaging convention
Layout: `reports/<slug>/{findings/*.md, pocs/*, evidence/*}` + `SUMMARY.md` + zip of everything plus `recon/<domain>_recon.json`. Severity-prefix finding files (`medium_*.md`, `low_*.md`) so an aggregator can sort them. When the `zip` binary is missing, `python3 -c` with `zipfile.ZipFile(..., "w", ZIP_DEFLATED)` walking the tree works everywhere.

## 8. VERIFICATION DISCIPLINE (pitfalls that produced real corrections)
1. **Title/keyword false positives.** Auto-detect flagged Gitea "open registration" because the page title contained "Register" — the body actually said "Registration is disabled". Before writing a finding, grep the actual claim-bearing text, not navigation/title keywords.
2. **First extraction undercounts.** Initial regex over `/explore/repos` found 3 repos; the corrected PoC regex found 11 (incl. production source). Re-run detection with the final pattern against saved raw HTML and UPDATE all artifacts (finding .md, recon JSON, notes) to the larger true set.
3. **PoC > recon for ground truth.** Run the PoC script before finalizing severity; let its output rewrite the finding. If PoC contradicts earlier recon, the PoC wins and the correction gets documented in the finding file itself.
4. **Recapture evidence after corrections** so archived HTML/JSON matches the final claims — stale evidence contradicts your own report.
5. crt.sh intermittently returns HTTP 502 with an nginx error page (not valid JSON): retry once, then note the CT-enum gap as a limitation instead of stalling.
6. **Run every deliverable PoC before shipping (Phase 2 lesson).** In one engagement BOTH new PoCs failed on first run: `tarfile.extractfile(m).read(errors=...)` → TypeError (see §6b), and a phpinfo parser labeled `disable_functions` as "terisi" because it tested non-emptiness while the page renders `<i>no value</i>` — match literal page text, don't infer from emptiness. Gate: subprocess each PoC → exit 0 AND ≥1 expected marker line AND output consistent with the finding file → only then package/deliver.
7. **Port-scan service labels lie about scheme on non-standard ports.** nmap `-sV` called a TLS 1.3 FastAPI service "uvicorn http :8888"; the finding file had to be rewritten post-hoc. Before reporting "plain HTTP API", handshake-test HTTPS explicitly (`ssl.wrap_socket` or curl -k) and probe BOTH schemes; when an earlier phase mislabeled, rewrite the finding and note the correction rather than stacking new claims on the wrong premise.

## 9. Moodle LMS anonymous deep-dive (no credentials; verified on a school deployment 2026-08)

### Version triangulation WITHOUT admin access — report a RANGE, never a pinned number
- `/lib/upgrade.txt` header → branch floor ("=== 4.5 Onwards === ... replaced by UPGRADING.md" ⇒ ≥4.5).
- `/UPGRADING.md` TOP section → deployed code era (top `## 5.0.1+` with NO `## 5.1` section = code written before 5.1 release notes existed, even if the site now runs 5.1).
- `/admin/environment.xml` max `<MOODLE version="X.Y">` block → branch ceiling (anchor regex to `<MOODLE\s` — a bare `version="([\d.]+)"` matches unrelated `<LIBRARY version="19">` elements and yields garbage like "19").
- `/composer.json` name + `/composer.lock` → sanity only (no version key present).
- Combine floor ≤ deployed ≤ ceiling. For CVE matching the operative claim is "pre-patch X": verify the site predates the fixed release (UPGRADING.md lacks the fixed-version section) instead of claiming an exact build you cannot see.

### Anonymous access-control matrix (record positives AND negatives)
- Content pages (profile/course/blog/search/calendar) all redirect to login = control WORKING; record as confirmed non-issue, not skipped coverage.
- `/login/signup.php` 404 = self-registration closed; `/admin/cron.php` "disabled by administrator" = protected.
- `/login/forgot_password.php` + `/login/token.php`: hash-diff FULL bodies for valid vs INVALID usernames — Moodle renders them GENERIC (only sesskey/random-token bytes differ) ⇒ NO user enumeration; don't claim enum from status-code equality alone.
- PHP files (`/version.php`, `/config.php`) → 200 with **0 bytes** = executed server-side, safe; but `config-dist.php` may EXECUTE and leak a runtime Fatal (`$CFG->dataroot not configured`) — template config shipped to prod.
- Data files read in full by default: `lib/upgrade.txt`, `UPGRADING.md`, `composer.lock` (full dependency tree), `package.json`, `admin/environment.xml`.
- `/admin/settings.php` without params → app-level error "A required parameter (section) was missing" (NOT an nginx 404 — compare body sizes) = admin surface reachable from the internet.

### Login brute-force exposure test (non-destructive, bounded)
8 failed POSTs to `/login/index.php` with a FAKE username, 0.3 s interval, then STOP: statuses alternate 303/200 with "Invalid login", zero throttle/lockout/captcha keywords = "no rate-limit observed up to N attempts" (a lower bound — never claim unbounded). Save the sequence as evidence; also check `/login/token.php` (mobile WS) as the parallel stuffing surface.

### CVE batch matching
Fetch moodle.org/security, keep advisories whose affected-range covers the triangulated range, and split status honestly: version exposure = CONFIRMED; exploit impact usually needs a low-priv account = POTENTIAL. CVE-2026-58348 (report-builder fragment capability check) is the student-data leak path.

## Related
- `recon`, `web2-recon` (primary pipelines — this skill is their no-binary/fallback companion)
- `web-auth-defenses-testing` — login rate-limit/lockout scoring methodology (§9 above is its Moodle-flavored instance)
