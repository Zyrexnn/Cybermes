# AI Gateway Deep Dive: Common AI Router Architecture Patterns

## Architecture Overview
- **Frontend (Control Plane):** Next.js / React on Cloudflare or Vercel.
- **Backend (Data Plane):** Express / Go API Gateway (`gateway-backend.example-cloud.internal`).
- **Communication:** The frontend documentation often leaks the direct backend URL, which may have different security controls (CORS, Auth) than the frontend proxy.

## High-Yield Checks
1. **Unauthenticated Model Enumeration**:
   - Path: `GET /api/v1/models` (on the frontend) or `GET /v1/models` (on the backend).
   - Finding: Returns full JSON list of models with multipliers and providers without any `Authorization` header.
   - Severity: **High** (Information Disclosure).

2. **CORS Wildcard on Landing Pages**:
   - Headers: `access-control-allow-origin: *` is common on the landing/login/register pages.
   - Severity: **Medium** (CORS Misconfiguration).

3. **Backend Service Discovery**:
   - Check `/.well-known/security.txt` or `/docs`.
   - The direct backend URL often lacks the custom WAF/Rate-limiters applied to the main domain.

4. **Account Creation Policy**:
   - Test `Usage & Costs` and `API Key` management.
   - Check if the key generated for the "Control Plane" works for the "Data Plane" (Gateway). They are often distinct validation domains.

## Optimization: Crawling on Next.js
- Standard crawling (`katana -depth 2`) can take 30m+ and yield minimal endpoints if it gets stuck on static assets or Next.js internal routing.
- **Fix**: Use `httpx -path /api/v1/models` and manual JS chunk analysis over broad crawling for SPA-heavy targets.
