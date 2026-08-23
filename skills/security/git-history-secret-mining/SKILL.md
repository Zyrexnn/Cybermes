---
name: git-history-secret-mining
description: >-
  Class-level playbook for mining secrets from git history on self-hosted forges
  (Gitea/Gogs/GitLab) during authorized assessments — anonymous repo browsing
  enumeration, deleted-file recovery via deletion-commit parents, raw-blob URLs with
  `^` parent syntax, live read-only validation of leaked static keys, and the
  "removed later != fixed" remediation trap. Use when a target runs its own git
  server or when leaked credentials need impact proof.
---

# GIT-HISTORY SECRET MINING — Self-Hosted Forges (Gitea/Gogs/GitLab)

A self-hosted git forge at an obscure subdomain (`guthib.<target>`, `git.<target>`,
`code.<target>`) with **anonymous repo browsing** is a full source + full history
disclosure without any webserver misconfiguration. In one authorized engagement this
single surface yielded both CRITICAL findings: a `.env` with a production API secret key,
and a 678-account SQL user dump. No `.git/` directory exposure needed.

> Authorization note: use only in scoped engagements. Treat recovered dumps as sensitive;
> verify claims from git objects yourself, never from scanner/subagent summaries alone.

## Phase A — Enumerate the Forge

1. Fingerprint: Gitea serves `<meta name="author" content="Gitea">` and a custom tagline;
   Gogs similar; check `/api/version` (Gitea, often unauthenticated).
2. List repos: `GET /explore/repos` — includes personal namespaces
   (`<firstname>/<project>`) that appear in NO DNS record, no subfinder output, no Google
   index. Student/intern-named repos are the highest-yield.
3. Check registration status (`/user/sign_up`) separately — "registration disabled" does
   NOT prevent anonymous browsing/expansion of public repos.
4. Clone everything visible in one pass:
   ```bash
   export PATH=/workspace/tools/bin:$PATH   # git may live here; apt install git otherwise
   git clone --mirror https://host/<owner>/<repo>.git   # mirror = all refs incl. deleted branches' objects
   ```
   Note working trees can be EMPTY while history is full — always inspect via
   `git log --all`, never judge by file count.

## Phase B — Mine History for Deleted Secrets

Deletion removes files from HEAD only; blobs stay reachable by every ancestor commit:

```bash
git log --all --oneline --diff-filter=D                     # deletion commits = treasure map
git log --all --name-status --diff-filter=AM -- .env '*.sql' '*.pem' '*.key'
git log -S 'x-api-key' --oneline --all                      # pickaxe search across history
git show <deletion-sha>^:<path>                             # blob exactly BEFORE deletion
```

- Commit messages like `rem: stop tracking .env`, or a data-dump commit immediately
  followed by `add api siswa`, point straight at payloads.
- Watch for dumps disguised as code: `resources/views/anjay.sql`, seeders
  (`UserSeeder.php`), fixture SQL inside app directories.
- Verify content before reporting (anti-hallucination): extract the blob, COUNT records,
  confirm value distributions (e.g. password histogram `{123456: 678}`), identify account
  naming patterns (per-region prefixes) — then write the finding with real numbers.

### No-clone shortcut (works while forge is up)

Gitea serves ANY historical blob over plain HTTP, including parent revisions — the `^`
suffix works inside the URL:

```
https://host/<owner>/<repo>/raw/commit/<sha>^/<path>
https://host/<owner>/<repo>/commit/<sha>          # human-readable diff view
```

## Phase C — Validate Leaked Secrets LIVE (read-only)

Source-code presence is not impact proof. Prove usability without mutating anything:

```python
# Static shared-secret middleware pattern (common in small-org backends):
r0 = client.get(API + "/items")                                   # expect 401
r1 = client.get(API + "/items", headers={"Authorization": f"Bearer {leaked_key}"})
# r0==401 AND r1==200 → auth bypass PROVEN, read-only
```

- Read route definitions from the cloned source to find valid GET endpoints before probing.
- Also mine third-party keys from history (`x-api-key:` headers, Basic-auth strings to
  partner APIs) and REPORT them even when out of scope — say "reported as-is, not tested".
- Keep a reusable PoC that re-derives the secret FROM THE FORGE URL itself
  (fetch raw blob → parse key → test pair). Exit 0 + expected marker lines = validated.
- Re-run the leak URL + bypass pair AT DELIVERY TIME; paste fresh status into the finding.

## Phase D — Impact Mapping & the "harusnya dah fix" Trap

Operators assume leaks were already fixed ("harusnya dah fix"). Pre-answer their follow-ups:

| Leaked item | What to verify before delivery |
|---|---|
| DB URL | Is the port externally open (nmap)? localhost-only DATABASE_URL = NOT externally loginable — say so plainly |
| Account dump | Does the app even have a live deployment? Resolve candidate hostnames directly via DNS (subfinder misses Cloudflare-hosted records like `absensi.`); probe `/login`, `/api`, `/admin`; report actual state ("under construction, 404s") |
| API key | Live 401-vs-200 pair on a GET endpoint |
| Leak URL itself | Fresh status of `/raw/commit/<sha>^/<path>` — "file removed in later commit" is NOT remediation while this returns 200 |

Remediation language that must appear in the finding: history rewrite (BFG /
git-filter-repo) or repo privatization, PLUS rotation of every rotated-value credential —
pulling the file from HEAD fixes nothing because clones already exist.

## Pitfalls

- **Empty working tree ≠ empty repo** — history lives in `.git`; check `git log --all`.
- **Trust nothing but git objects**: extract blobs yourself (`git show`) instead of
  trusting subagent logs; one session's subagent log omitted record counts that changed
  the finding's precision.
- **Forge subdomains are unlisted**: brute candidate prefixes (`guthib.`, `git.`,
  `code.`) and grep site HTML for links; DNS-based recon will not find them.
- **`zipfile.ZipFile(...)` reopened in `'r'` mode cannot `writestr()`** — build manifest
  inside the same write-mode context when assembling report zips.
- execute_code scripts do not persist imports between calls — import modules at the top
  of EVERY script.

## References

- `scripts/validate_leaked_key.py` — runnable end-to-end validator (forge raw blob →
  parse secret → 401/200 bypass pair). Use this instead of hand-writing the PoC each time.
- Worked example with exact commands: see section above; evidence patterns belong under
  the engagement's `evidence/git_history/` folder per house conventions
  (`engagement-deliverables-and-validation`).
