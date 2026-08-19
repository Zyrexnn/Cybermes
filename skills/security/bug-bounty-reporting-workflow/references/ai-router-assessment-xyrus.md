# AI Router Assessment Case Study: xyrusrouter.xyz

## Target Overview
- **UI (Control Plane):** `xyrusrouter.xyz` (Vercel)
- **API Gateway (Data Plane):** `ai-gateway-production-df56.up.railway.app` (Railway)

## Discovery Workflow
1. **Endpoint Mining (Passive):**
   - Documentation at `/docs` revealed the actual API Gateway URL on Railway.
   - `/api/v1/models` was identified as a core discovery endpoint.

2. **Information Disclosure (CWE-200):**
   - **Endpoint:** `GET https://xyrusrouter.xyz/api/v1/models`
   - **Finding:** Endpoint returned a 200 OK with the full model registry (18 models, providers, multipliers) without any `Authorization` header.
   - **Severity:** High (7.5) due to exposure of commercial model configuration and registry.

3. **CORS Misconfiguration (CWE-942):**
   - **Finding:** `access-control-allow-origin: *` present on `/`, `/login`, `/register`.
   - **Mitigation Check:** `Access-Control-Allow-Credentials` was absent, limiting impact to unauthenticated data.

4. **Auth Partitioning:**
   - Keys generated on the Control Plane (Vercel) were initially rejected by the Data Plane (Railway) with "AI API key tidak valid", suggesting separate auth databases or a "funding required" state before keys propagate.

## Reproduction Payloads
```bash
# Check for unauthenticated model leak
curl -i "https://xyrusrouter.xyz/api/v1/models"

# Check for CORS wildcard
curl -I -H "Origin: https://evil.com" "https://xyrusrouter.xyz/"
```