# AI Router Architecture & Admin Endpoint Assessment Case Study

## Target Overview
- **Domain**: `ai-router.example.org`
- **IP Addresses**: `198.51.100.10`, `198.51.100.11` (Cloudflare Anycast / Edge)
- **CDN/WAF**: Cloudflare (with Bot Management)
- **Technology Stack**: Cloudflare, HSTS, HTTP/3
- **Application Type**: AI Gateway Service (Next.js/Vercel-like frontend)
- **Backend**: Microservice hosting (`ai-gateway-backend.example-cloud.internal`)
- **Scope**: Authorized Security Assessment

## Architecture Analysis

### Infrastructure Split
- **Frontend**: `ai-router.example.org` (Vercel/Cloudflare)
- **Backend**: `ai-gateway-backend.example-cloud.internal` (Cloud container)
- **Pattern**: Typical AI Router split architecture (Control Plane + Data Plane)

### Service Type
- **AI Gateway**: OpenAI-compatible API gateway
- **Features**:
  - Multi-model support
  - Integrated payment flow
  - Instant activation after payment
  - Usage tracking and quota policy

## Reconnaissance Results

### Subdomain Discovery
- **Primary Domain**: `ai-router.example.org`
- **DNS Records**:
  - A: `198.51.100.10`, `198.51.100.11`

### Live Endpoints Discovered
```
https://ai-router.example.org/
https://ai-router.example.org/account/claim-link/
https://ai-router.example.org/account/admin/vouchers
https://ai-router.example.org/account/admin/claim-links/
https://ai-router.example.org/shop/packages
https://ai-router.example.org/token/
https://ai-router.example.org/v1/usage/detail
https://ai-router.example.org/buyer/fup-policy
https://ai-router.example.org/buyer/announcement
https://ai-router.example.org/admin/
```

### Admin Dashboard Paths
- ✅ **Found**: `/account/admin/vouchers` (POST endpoint)
- ✅ **Found**: `/account/admin/claim-links/` (POST endpoint)
- ✅ **Found**: `/admin/` (returns 404, but exists in route crawl)

## Testing Methodology Applied

### 1. Passive Reconnaissance
- DNS resolution with `dnsx`
- HTTP probing with `httpx` (rate-limited)
- Endpoint crawling with `katana` (depth 3, rate-limit 5)

### 2. Active Testing (Stealth Mode)
- **Rate Limiting**: All requests capped at 5 req/s
- **User Agent**: Mozilla/5.0 (to avoid bot detection)
- **Method Testing**: GET, POST, OPTIONS on all endpoints
- **Header Analysis**: Security headers, CORS, content-type

### 3. Admin Endpoint Testing
- Tested with/without authentication
- Tested with/without trailing slashes
- Tested different HTTP methods (GET, POST, OPTIONS)
- Analyzed error messages for information disclosure

## Key Findings

### Admin Endpoint Behavior

#### `/account/admin/vouchers`
- **GET**: 405 Method Not Allowed
- **POST**: 401 Unauthorized (login_required)
- **Validation**: Returns field validation errors when tested with invalid data
- **Required Fields**: `days` (integer)

#### `/account/admin/claim-links/`
- **GET**: 405 Method Not Allowed
- **POST**: 401 Unauthorized (login_required)
- **Validation**: Returns field validation errors
- **Required Fields**: `share_token_id`

#### `/admin/`
- **GET**: 404 Not Found (detected in JS crawl)
- **Note**: May be a redirect or hidden client-side route

### Error Message Analysis

**Detailed Error Responses:**
```json
// Authentication required
{"detail":{"error":{"code":"login_required","message":"Authentication required."}}}

// Field validation errors
{"detail":[{"type":"missing","loc":["body","days"],"msg":"Field required"}]}
```

## Remediation Recommendations
1. Implement unified rate-limiting across all `/account/admin/*` and public endpoints.
2. Standardize error message formats to avoid exposing internal field names and schemas to unauthenticated callers.
3. Ensure backend proxy verifies incoming requests and prevents direct access to cloud backend hostnames.