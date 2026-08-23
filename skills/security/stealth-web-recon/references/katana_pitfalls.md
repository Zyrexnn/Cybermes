# katana Pitfalls — Verified Reproduction & Workarounds

## 1. `-silent` routes results to stderr (v1.7.0) — THE big one
**Symptom:** `katana ... -silent > out.txt 2>/dev/null` produces a 0-byte `out.txt` even though the target is crawlable. Looked like a successful empty crawl; wasted two full crawl runs.

**Why:** With `-silent`, katana prints crawl results to **stderr**, not stdout. Shell `> out.txt` captures empty stdout; `2>/dev/null` throws away the real results.

**Proof (observed):**
```
# BROKEN — 0 lines
katana -u https://x/ -d 3 -rate-limit 8 -silent > main.txt 2>/dev/null
wc -l main.txt   # -> 0

# WORKS — 12976 lines (results on stdout, banner on stderr)
katana -u https://x/ -d 2 -c 8 -rate-limit 8 -timeout 12 2>err.txt | tee out.txt
wc -l out.txt    # -> 12976
```

**Fix (pick one):**
- katana's own output flag: `katana ... -o file.txt` (no `-silent`, no shell redirect).
- Or drop `-silent`, capture stdout: `katana ... > out.txt 2>err.txt`.

## 2. Headless mode silently empty
`-jc -kf all` requires Chromium. Missing -> exit 0, 0 endpoints. Use standard HTTP crawl unless Chromium confirmed.

## 3. Background truncation by SIGINT
Background `katana ... > file` whose parent shell receives SIGINT can be interrupted mid-crawl ("Ctrl+C pressed" in stderr, empty/partial file). Use:
`setsid bash -c 'katana ... > file 2>err.txt; echo DONE >> file'`

## 4. SPA / Cloudflare-fronted hosts return only root URL
Non-JS crawl of SPA or Bot-Managed hosts yields just `/`. Recover deeper surface from links found on OTHER crawled pages (e.g. Moodle `elena` surfaced via main-site HTML, not via its own crawl).
