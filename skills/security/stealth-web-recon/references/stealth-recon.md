# Stealth Recon — Worked Command Set

Condensed from a live bug-bounty recon of `nurulfikri.ac.id` (76 live subdomains, 146 httpx probes, zero blocks). Default low-noise sequence.

## 1. Passive subdomain discovery
```bash
timeout 90 subfinder -d <target> -silent -recursive 2>/dev/null | sort -u | tee /tmp/subfinder.txt
curl -sS --max-time 30 -A "Mozilla/5.0" "https://crt.sh/?q=%25.<target>&from=...&output=json" > /tmp/crt.json
python3 -c "import json; d=json.load(open('/tmp/crt.json')); [print(n['name_value']) for n in d]"
```

## 2. DNS resolution
```bash
timeout 120 dnsx -l /tmp/subfinder.txt -silent -r 8.8.8.8,1.1.1.1 2>/dev/null | tee /tmp/resolved.txt
```

## 3. httpx web probe — RATE-LIMITED (the only active step)
```bash
timeout 180 httpx -l /tmp/resolved.txt -silent -t 30 -rate-limit 15 -timeout 8 \
  -ports 80,443 -title -status-code -tech-detect -server -follow-redirects 2>/dev/null | tee /tmp/httpx.txt
```
Mandatory flags: -rate-limit 15, -t 30, -follow-redirects, -tech-detect (reveals Cloudflare Bot Management).

## 4. Passive URL mining (Wayback, external)
```bash
timeout 120 gau --subs <target> 2>/dev/null | sort -u | tee /tmp/gau.txt
grep -iE '\.(env|json|xml|sql|bak|zip|log|git|yml|config|backup)' /tmp/gau.txt | head
```

## 5. Confirm exposed panel versions — ONE slow GET each, sequential
```bash
for u in "https://admin.dev-app.<target>/" "https://n8n.<target>/" "https://kuma.<target>/"; do
  echo ">>> $u"
  curl -sS --max-time 15 -A "Mozilla/5.0" "$u" | grep -ioE "(Adminer [0-9.]+|Uptime Kuma|Nginx Proxy Manager|version [0-9.]+)" | head -2
  sleep 1
done
```

## WAF-exclusion example
admisi, asset, aset, lms-ddp showed Cloudflare / Cloudflare Bot Management — header-only, no content scanning.

## Gotchas
- Versions often DON'T leak in HTML (title-only) — don't burn requests forcing it.
- nginx/1.14.0 (Ubuntu) default pages = version leak + likely EOL.
- 502/500 on staging hosts = dead dev infra exposed publicly.
- gau/crt.sh empty output is NORMAL (external rate-limits), not a failure.
