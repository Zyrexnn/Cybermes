# ffuf Compatibility & WAF False-Positive Handling
Verified during the 2026-08-22 semutssh.com engagement. Companion to
`engagement-deliverables-and-validation` §6 (late-lead verification) — the
`fuzzing-and-content-discovery` skill is manually authored and could not absorb these.

## 1. Version compatibility (probe before scripting)
- `ffuf -V` first. Older builds reject `-rl` (rate limit) → use `-rate <n>`, accepted across versions.
- An unknown flag makes ffuf print its help text and exit WITHOUT producing the output file — a silent empty-result failure. If the JSON output is missing, suspect flag incompatibility before suspecting zero matches.
- Always pair `-o file` with `-of json`; bare `-o` writes a text dump that breaks downstream `json.load()` parsing.

## 2. Uniform-403 WAF block pages = false positives
Signature observed on a Cloudflare-fronted CodeIgniter host:
- ffuf matched `/.env`, `/.git/config`, `/dump.sql`, `/server-status`,
  `/application/logs/log-2026-08-22.php`, `/composer.json`, `/.htaccess` — all 403.
- Manual curl re-check: every path returned 403 with an IDENTICAL 285-byte body = the
  generic WAF block page, not the files.

Decision rule:
1. All matches same status + same body size ⇒ block page ⇒ NEGATIVE result.
2. Record the negative explicitly in `evidence/recon_notes.md` (paths tested, sizes,
   conclusion) — it is coverage proof, not a finding.
3. A path with a UNIQUE size/body is the only candidate for a real hit: fetch and read
   its content before raising a finding.

## 3. Exposed Swagger UI follow-up (payment gateway case)
ffuf hit on `/docs` (200, ~950B, title "Semut Payment Gateway API Docs"):
1. Grep the docs HTML for the spec location — the Swagger UI init page carries a
   `url: 'https://host/docs/spec'` config line.
2. Fetch the spec with `Accept: application/json` AND a browser UA (Cloudflare may
   serve differently); also try `docs/spec.json`, `docs/openapi.json`, `docs/spec.yaml`,
   `index.php/docs/spec` path variants.
3. If spec = 404 while UI = 200, that is still a LOW finding: it confirms the internal
   API surface exists. Pair it with any error-message differential observed on the API
   (e.g. login endpoint distinguishing "field kosong" vs "salah kredensial" = username
   enumeration aid).
4. Do not burn time on broken spec endpoints — document and move on.

## 4. Background scanner lifecycle during engagements
- Long scans (nuclei over many hosts × template sets) can run 20+ min; run them with
  `background=true, notify_on_complete=true` and keep doing manual validation meanwhile.
- When the engagement ends before they finish, kill them; expect exit 143 notifications
  later. Re-verify deliverable integrity on each late notification before reacting
  (see umbrella §6).
- A late-finishing ffuf can still contribute verified leads (positive or negative) —
  verify manually, append to recon_notes, rebuild + re-send the zip.
