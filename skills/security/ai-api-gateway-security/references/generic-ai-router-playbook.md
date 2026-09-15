# Generic AI Router & API Gateway Security Playbook

## 1. Architectural Overview

Modern AI API Gateways and LLM Routers commonly employ a **split-plane architecture**:
- **Control Plane (Frontend / User Portal)**: Typically built with Next.js, React, or SvelteKit, often deployed on edge platforms (Vercel, Cloudflare Pages). Handles user authentication, dashboard analytics, payment top-up processing, and API key generation.
- **Data Plane (API Gateway & Upstream Routing Engine)**: Built with high-performance Go, FastAPI, or Node.js services (e.g. New API, One API, LiteLLM). Proxies OpenAI-compatible requests (`/v1/chat/completions`, `/v1/models`) to upstream providers (Anthropic, OpenAI, DeepSeek, Groq).

```text
[Client / LLM Consumer]
      │
      ▼  (Bearer Token: sk-...)
┌──────────────────────────────────────────────────────────────┐
│  Data Plane (API Gateway): https://gateway.example.com       │
│  - Endpoint: /v1/chat/completions                            │
│  - Model Catalog: /api/v1/models                             │
└──────────────────────────────┬───────────────────────────────┘
                               │ Upstream Routing & Fair Use
                               ▼
┌──────────────────────────────────────────────────────────────┐
│  Control Plane (Admin / Dashboard): https://app.example.com  │
│  - User Auth & Billing Top-Up                                │
│  - Key Generation & Quota Allocation                         │
└──────────────────────────────────────────────────────────────┘
```

---

## 2. Core Audit Vectors (OWASP GenAI Top 10 2026)

### Vector 1: Unauthenticated Model Catalog Enumeration (LLM02: Sensitive Info Disclosure)
Many gateways expose the internal catalog endpoint without requiring an `Authorization` header:
- **Endpoint**: `GET /api/v1/models` or `GET /v1/models`
- **Impact**: Exposes proprietary model multipliers, upstream backend routing routes, or confidential internal model identifiers.
- **Diagnostic Command**:
  ```bash
  python3 scripts/audit_ai_gateway_config.py --url https://gateway.example.com
  ```
- **Remediation**: Require valid Bearer token authentication for all catalog endpoints.

---

### Vector 2: Auth Partitioning & Token Propagation Latency
In split-plane architectures, API keys created on the Control Plane may be asynchronously replicated to the Data Plane:
1. **Timing Flaws**: Attempting prompt completions immediately after key generation may return `401 Invalid Token` or trigger unhandled fallback exceptions.
2. **Revocation Bypass**: Confirm whether revoking an API key on the Control Plane immediately invalidates cached tokens on the Data Plane or leaves a persistence window.

---

### Vector 3: Payment Flow & Webhook State Machine Integrity
Top-up subdomains often integrate regional or credit card gateways:
1. **Status Spoofing**: Verify whether payment confirmation webhooks validate HMAC cryptographic signatures (e.g., `X-Signature` or `X-Callback-Token`) before crediting account balance.
2. **Currency & Amount Tampering**: Check that product tier IDs and amounts are strictly evaluated server-side rather than accepted from client JSON bodies.

---

### Vector 4: CORS Misconfigurations on Edge Routers
- Edge proxies behind Cloudflare or reverse proxies often apply permissive CORS headers to facilitate web chat interfaces.
- Test for arbitrary origin reflection with credential support:
  ```bash
  curl -I -H "Origin: https://audit.example.org" "https://app.example.com/"
  ```
- **Remediation**: Restrict `Access-Control-Allow-Origin` to the specific origin of the authorized web interface.
