---
name: authentication
description: Assesses authentication flows, OAuth2 / OIDC implementations, session lifecycle management, and JWT security.
---

# Authentication & Session Testing Skill

## Purpose
Inspect authentication mechanics, token handling, session invalidation, and privilege separation in web and mobile backends.

## Workflow
1. **Identify Auth Mechanisms**:
   - Cookie-based sessions (inspect `HttpOnly`, `Secure`, `SameSite` flags)
   - JWT tokens (Header: `alg`, Payload claims: `sub`, `exp`, `role`)
   - OAuth2 / OIDC flows (Redirect URI validation, `state` parameter presence, PKCE enforcement)
2. **Evaluate Common Flaws**:
   - **JWT Vulnerabilities**: Weak secret keys, `alg: none` handling, missing signature validation, token expiry validation.
   - **Session Lifecycle**: Does logging out invalidate the server-side session or JWT? Can old sessions be reused after password change?
   - **OAuth Flaws**: Open redirect on `redirect_uri`, CSRF due to missing or static `state`, token leakage via Referer headers.
   - **Password Reset Flows**: Predictable tokens, host header injection in reset links, account enumeration via timing/response differences.
3. **Document Flow & Observations**:
   Log sequence diagrams and HTTP traces for any identified authorization discrepancy.

## Safety Constraints
- Never lock out real accounts through brute force.
- Never alter production user passwords without explicit consent.
