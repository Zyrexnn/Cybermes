---
name: parameter-mining
description: Discovers hidden GET/POST/JSON parameters from web applications and categorizes them for IDOR, SSRF, SQLi, and redirect testing.
---

# Parameter Mining Skill

## Purpose
Identify hidden, undocumented, or legacy HTTP parameters that can serve as injection vectors for IDOR, SSRF, XSS, and SQL Injection using `arjun`, `gau`, and pattern extraction.

## Inputs
- Discovered URLs from recon: `/workspace/recon/<target>/endpoints_all_*.txt` or `/workspace/recon/<target>/parameter_urls_*.txt`.
- Live endpoints: `/workspace/recon/<target>/httpx_live_*.txt`.

## Workflow

### 1. Active Parameter Discovery (Arjun)
Probe live endpoints for hidden GET and POST JSON parameters:
```bash
# Discover GET query parameters
arjun -u "https://target.com/endpoint" \
      -m GET \
      --rate-limit 10 \
      -oJ /workspace/output/parameters/<target>_get_params.json

# Discover POST JSON parameters
arjun -u "https://target.com/api/v1/user" \
      -m JSON \
      --rate-limit 10 \
      -oJ /workspace/output/parameters/<target>_post_params.json
```

### 2. Passive Parameter Mining & Extraction (GAU + Grep)
Extract query parameter keys from historical and crawled URLs:
```bash
python3 - << 'EOF'
import json, re, sys, os
from urllib.parse import urlparse, parse_qs

target = sys.argv[1] if len(sys.argv) > 1 else ""
url_file = f"/workspace/recon/{target}/endpoints_all.txt"

params_by_type = {
    "ssrf": set(),       # url, dest, redirect, link, uri, fetch
    "idor": set(),       # id, user_id, account, uuid, order_id
    "sqli": set(),       # search, query, filter, sort, order
    "redirect": set(),   # return_to, next, redirect_url, callback
    "all_params": set()
}

if os.path.exists(url_file):
    with open(url_file) as f:
        for line in f:
            parsed = urlparse(line.strip())
            qs = parse_qs(parsed.query)
            for k in qs.keys():
                params_by_type["all_params"].add(k)
                k_lower = k.lower()
                if any(x in k_lower for x in ["url", "dest", "uri", "fetch", "domain"]):
                    params_by_type["ssrf"].add(line.strip())
                if any(x in k_lower for x in ["id", "user", "account", "uuid", "order"]):
                    params_by_type["idor"].add(line.strip())
                if any(x in k_lower for x in ["search", "query", "filter", "sort"]):
                    params_by_type["sqli"].add(line.strip())
                if any(x in k_lower for x in ["next", "return", "redirect", "callback"]):
                    params_by_type["redirect"].add(line.strip())

os.makedirs(f"/workspace/output/parameters/{target}", exist_ok=True)
for category, items in params_by_type.items():
    with open(f"/workspace/output/parameters/{target}/{category}.txt", "w") as out:
        out.write("\n".join(sorted(list(items))) + "\n")
EOF
```

## Output Artifacts
- `/workspace/output/parameters/<target>/ssrf.txt` - URLs with potential SSRF parameters.
- `/workspace/output/parameters/<target>/idor.txt` - URLs with object identifiers for authorization testing.
- `/workspace/output/parameters/<target>/redirect.txt` - Open redirect and OAuth callback candidates.
- `/workspace/output/parameters/<target>_get_params.json` - Arjun hidden parameter findings.

## Next Step Integration
- Feed `ssrf.txt` into the `oast-blind-testing` skill.
- Feed `idor.txt` into the `business-logic-and-idor` skill.
- Feed `redirect.txt` and `sqli.txt` into the `vulnerability-analysis` skill.
