# AI Gateway Deep Dive: Xyrus / New API Patterns

## Target: xyrusrouter.xyz (August 2026)

### Architecture
- **Frontend**: Next.js on Vercel.
- **Backend (API Gateway)**: Express.js on Railway.app (`ai-gateway-production-df56.up.railway.app`).
- **Communication**: The frontend documentation often leaks the direct Railway URL, which may have different security controls (CORS, Auth) than the frontend proxy.

### High-Yield Checks
1. **Unauthenticated Model Enumeration**:
   - Path: `GET /api/v1/models` (on the frontend) or `GET /v1/models` (on the Railway backend).
   - Finding: Returns full JSON list of 18+ models with multipliers and providers without any `Authorization` header.
   - Severity: **High** (Information Disclosure).

2. **CORS Wildcard on Landing Pages**:
   - Headers: `access-control-allow-origin: *` is common on the Vercel landing/login/register pages.
   - Severity: **Medium** (CORS Misconfiguration).

3. **Railway Backend Discovery**:
   - Check `/.well-known/security.txt` or `/docs`.
   - The backend URL (`*.up.railway.app`) often lacks the custom WAF/Rate-limiters applied to the main domain.

4. **Account Creation Policy**:
   - If owner-authorized, create an account to test `Usage & Costs` and `API Key` management.
   - Check if the key generated for the "Control Plane" works for the "Data Plane" (Gateway). In Xyrus Router, they are often distinct.

## Optimization: Katana on Next.js
- Standard crawling (`katana -depth 2`) can take 30m+ and yield only 1 endpoint if it gets stuck on static assets or Next.js internal routing.
- **Fix**: Use `httpx -path /api/v1/models` and manual JS chunk analysis over broad crawling for SPA-heavy targets.
