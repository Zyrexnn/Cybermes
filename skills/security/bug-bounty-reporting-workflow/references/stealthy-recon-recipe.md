# Stealthy Recon Recipe (Validated Clean Execution)

Validated against large institutional web targets: 80 subdomains → 76 live → 146 httpx probes, plus katana crawl. **No WAF/Cloudflare blocks, no 429s.** Standard rule: low-noise, rate-controlled scanning that avoids tripping automated rate limits and WAF blocks.

## Golden rules
1. Passive-first. No HTTP to target until DNS is resolved.
2. Cap rate. Treat the values below as ceilings.
3. Skip `Cloudflare` + `Cloudflare Bot Management` hosts for any active fuzzing.
4. Long crawls → background + poll (foreground tool caps at ~180s).

## Command sequence
```bash
# 1) Passive subdomain discovery (no hits to target)
subfinder -d <target> -silent -recursive > subfinder.txt
# (crt.sh may be rate-limited/empty — retry with backoff if needed)
# curl -sS "https://crt.sh/?q=%25.<target>&output=json"

# 2) DNS resolve only (safe, no HTTP)
dnsx -l subfinder.txt -silent -r 8.8.8.8,1.1.1.1 > resolved.txt

# 3) Low-rate HTTP probe (RATE-LIMIT 15 = ceiling)
httpx -l resolved.txt -silent -t 30 -rate-limit 15 -timeout 8 \
  -ports 80,443 -title -status-code -tech-detect -server -follow-redirects > httpx.txt

# 4) Passive historical URLs (wayback — no direct hits)
gau --subs <target> 2>/dev/null | sort -u > gau.txt   # may be empty if rate-limited

# 5) Crawl — background (depth>=2 exceeds 180s foreground cap)
cat > targets.txt <<'EOF'
https://www.<target>/
https://portal.<target>/
EOF
timeout 540 katana -list targets.txt -jc -kf all -d 2 -c 10 -delay 1200 \
  -rate-limit 8 -timeout 10 -silent > crawl.txt 2>/dev/null
# run via terminal(background=true, notify_on_complete=true); poll for completion
```

## Findings-to-flag from probe output
- `Cloudflare` + `Cloudflare Bot Management` → mark "do not fuzz", passive only.
- Admin/internal panels on public IP (Adminer, n8n, Uptime Kuma, Nginx Proxy Manager, DocuSeal, Proxmox, Portainer, Nessus) → exposure finding (verify auth before severity).
- Outdated server banner (e.g. `nginx/1.14.0 (Ubuntu)`) → version disclosure + EOL.
- Status `502`/`500`/`default page` on staging hosts → large attack surface.
- Missing security headers on main web properties (HSTS, CSP, X-Frame-Options).

## Rate-limit cheat-sheet (ceilings, not targets)
| Tool  | Flag                          | Value |
|-------|-------------------------------|-------|
| httpx | `-rate-limit`                 | 15    |
| katana| `-rate-limit` + `-delay`      | 8 + 1500ms |
| dnsx  | (default, public resolvers)   | safe  |
