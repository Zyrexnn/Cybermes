# AI Gateway Control Plane Assessment Case Study

## Target Architecture
- **UI (Control Plane):** `gateway-ui.example.io` (Vercel / Edge)
- **API Gateway (Data Plane):** `ai-gateway-backend.example-cloud.internal` (Container backend)

## Discovery Workflow
1. **Endpoint Mining (Passive):**
   - Documentation at `/docs` revealed the actual API Gateway backend URL.
   - `/api/v1/models` was identified as a core discovery endpoint.

2. **Information Disclosure (CWE-200):**
   - **Endpoint:** `GET https://gateway-ui.example.io/api/v1/models`
   - **Finding:** Endpoint returned a 200 OK with the full model registry (models, providers, multipliers) without an `Authorization` header.
   - **Severity:** High (CVSS 7.5) due to exposure of commercial model configuration and internal registry.

3. **CORS Misconfiguration (CWE-942):**
   - **Finding:** `access-control-allow-origin: *` present on `/`, `/login`, `/register`.
   - **Mitigation Check:** `Access-Control-Allow-Credentials` was absent, limiting impact to unauthenticated data.

4. **Auth Partitioning:**
   - Keys generated on the Control Plane were rejected by the Data Plane when quota/funding state was uninitialized, showing distinct validation layers between control and data planes.

## Reproduction Payloads
```bash
# Check for unauthenticated model leak
curl -i "https://gateway-ui.example.io/api/v1/models"

# Check for CORS wildcard
curl -I -H "Origin: https://evil.com" "https://gateway-ui.example.io/"
```