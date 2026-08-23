# Laravel App Audit — Response Signatures & Probe Patterns

Condensed from an enterprise Laravel web audit: Laravel + Apache on Linux, public CMS frontend, custom business module, authorized engagement.

## Response-signature triage table

| Probe | Observed | Meaning |
|---|---|---|
| `/.env`, `/.git/HEAD`, `/composer.json`, `/phpinfo.php`, `/server-status`, `/adminer.php` | 403, ~282 bytes, `charset=iso-8859-1` | Real Apache-level block (`<FilesMatch>` deny) — control working |
| `/storage/logs/laravel.log`, `/vendor/composer/installed.json` | 404, ~1.5KB | File absent — clean |
| `/telescope`, `/horizon`, `/.env.backup`, any unknown path | 200, ±56KB text/html, title = site homepage | Catch-all fallback route — NOT exposure |
| `/members/{invalid-id}` | 200, empty profile shell ("Detail Profil -") | Soft-404; data public by design, not IDOR |
| `/portal/login` | 500, 334KB debug page "View [frontend.portal.index.login] not found" | Broken route + APP_DEBUG=true → Medium finding |
| `/portal/section/99999` (or 0, abc) | 500, 321KB debug page "Undefined offset: 0" | Unvalidated controller array access + debug page |

Distinguish catch-all hits from real ones: compare byte sizes and check whether the body is the homepage shell with only canonical/og:url changed.

## Debug-page identification

- Stock Laravel error page (APP_DEBUG=true): full HTML error layout, stack frames under `vendor/laravel/framework/src/Illuminate/...`, absolute server paths (`/var/www/html/<project>/app/...`), custom controllers/middleware named in frames. May NOT include env/request sections — absence of `.env` secrets does not downgrade the finding; path disclosure alone = Medium.
- **Always re-request the failing URL with `Accept: application/json`.** Debug mode typically also returns the FULL STACK TRACE as JSON (~15KB vs ~330KB HTML): `message`, `exception` class, exact `file`+`line` (e.g. names the app controller line: `PortalController.php:67`, custom middleware `HSTSMiddleware.php:13`), and an N-frame `trace` array. This is stronger evidence (precise code locations) and proves the leak applies to non-browser/API clients too. Same verdict logic: env/request sections usually absent → no secret leak, still Medium.
- NOT Ignition if zero `ignition`/`whoops` markers → CVE-2021-3129 irrelevant; say so explicitly as a verified negative.
- Extract evidence with Python regex over the saved dump: `r"/var/www/html/[a-zA-Z0-9_./-]+"` for paths, `r"app/Http/(?:Controllers|Middleware)/[A-Za-z0-9_.]+\.php"` for app code names.
- Root cause to report alongside: missing view file / unguarded array access in controller (`PortalController`), not just the debug flag.

## Validation-round patterns (follow-up to confirm/refute round-1 findings)

- Username-enumeration timing: 3× baseline POSTs with ONE repeated fake user → take median ms; then N candidate usernames (admin, secretary, staff, …) each with fresh CSRF token, ~1s apart. Spread < ~50ms ≈ noise = no side-channel (record per-user ms table). Combine with uniform error-body check ("Invalid credentials" for everything).
- ID-sweep for hidden entities: enumerate `/module/{type}/{id}` GET over ranges bracketing known sparse IDs (e.g. section {5,6,7,169} → sweep 1–60 + 150–180), ~350ms interval. Classify by response signature (valid = 200 large page; invalid = 500 debug page). Verdict "hidden/draft entities leak" ONLY if a valid-looking ID appears that is absent from public nav menus. ~135 requests, zero hidden — negative recorded as coverage proof.
- SQLi canary battery on query params: baseline canary vs `'` vs `" OR 1=1-- -` vs `' UNION SELECT NULL-- -`; uniform byte-size (<35B delta) + zero DB-error markers = parameterized, negative.
- Save every battery's raw output as machine-readable `evidence/<topic>.json` from the same script that runs it — the script doubles as the deliverable PoC in `pocs/val_*.py`.

## Login throttle PoC pattern (non-destructive)

- Scrape CSRF `_token` from GET /login with a cookie jar before each POST; without it every attempt fails as token-mismatch, which false-negatives a throttle test (same trap as Moodle logintoken).
- 6 attempts, SAME fake username (`fakeuser_probe_same` / `WrongPass123!`), ~1s sleep. Verdict keys: HTTP 429 vs 200/302, message text ("too many|throttl|locked"), uniform error body ("Invalid credentials") also proves no username enumeration via response diffing.
- Uniform failures across attempts = no-throttle finding CONFIRMED (CWE-307); if throttled, record lockout window as verified control.
- Never run wordlists; stop at negative verification.

## Misc verified-negatives worth re-testing per target

- CORS: foreign + null Origin, expect no ACAO reflection.
- Open redirect params (`?next=`, `?redirect=`, `?url=`): expect no external Location.
- Canary reflection: reflected only inside `<link rel="canonical">`, HTML-encoded → informational, not XSS.
- TLS: legacy 1.0/1.1 rejected, 1.2+1.3 negotiated (python ssl explicit-version test).
- Public polling forms (`POST /polling`) are state-changing → GET-only note, flag potential vote-stuffing for future authorized testing instead of probing.
