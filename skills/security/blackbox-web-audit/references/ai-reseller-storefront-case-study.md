# Case Study: AI-Credit Reseller & Inference Gateway Assessment

Full-surface authorized audit of a Cloudflare-fronted SaaS (AI credit reseller:
local currency topup → Claude/GPT inference via custom proxy gateway). 5 findings (2 High, 2 Medium,
1 Low), several large false positives eliminated. Techniques below are reusable reference patterns.

## What worked, in order of yield

1. **SPF → origin IP → multi-vhost bypass** (High). SPF `v=spf1 ip4:198.51.100.25 ~all`
   leaked the origin server. That ONE server served: marketing apex (LiteSpeed behind CF),
   `my.` console SPA, `api.` = Flowise instance, the production inference gateway
   (`inference-gateway.internal.example.net`), mail, and 7 sibling products on odd ports.
   Every Cloudflare control was bypassable by pinning hosts to the origin (`--resolve` or `/etc/hosts`).
2. **Public config endpoint as policy oracle + attack map.**
   `GET /api/auth/providers` returned `{registration:true, min_password:5,
   password_reset:false, google:true, github:true, free_balance_usd:0.20}` —
   revealed exactly which auth tests to prioritize before sending payloads.
3. **Register→login→session-preservation loop** for authenticated testing under an
   aggressive login rate-limit (429 after ~2 fails): register a disposable account
   (instant activation, no email verification — itself a Medium finding), save the
   bearer token to a file, reuse it across all batches; re-login only on real expiry.
4. **Gateway-vs-console rate-limit differential**: console login was throttled hard
   but the paid inference gateway `/v1/models` had ZERO throttle (25 sequential + 15
   parallel random-bearer requests, no 429/503) AND accepted session tokens as
   credentials (401→402 "insufficient balance" = auth passed, balance empty). No-rate-limit-on-
   paid-gateway + weak-token-hypothesis would have been an attack chain.
5. **nmap -p- port clusters** on the exposed origin revealed secondary services
   (SMS gateway :2029, payment gateway :2037, messaging API :2051, Go microservices :8795/:8796
   with custom `X-Custom-API-Key`/`X-Custom-Gateway-Secret` headers leaking their
   internal auth-header naming; one leaked its model list without auth).
6. **Flowise fingerprinting** (api. host): `/api/v1/version` → 200 unauth (version
   disclosure), `/ping` open, admin endpoints correctly 401, setup/signup locked,
   `get-upload-file?path=` traversal variants all rejected (patched), reflected CORS.

## False positives killed (and how each was caught)

1. **"Masked weak session token" (would-be Critical)** — every login/register response
   showed tokens like `sk-853...d8a9`, and that literal string AUTHENTICATED against
   both console and gateway. Conclusion "server accepts ~28-bit masked tokens" was
   wrong: raw-file inspection (`len()`/repr) showed the true 51-char token (~183 bits);
   the masking was OUR OWN toolchain's secret-redaction pipeline. Never reason about
   credential format from rendered tool output — read the bytes on disk.
2. **"SSRF render proxy"** — ffuf hit `/render/<url>` (307 then 200). Body-diff vs
   index.html baseline showed identical SPA shell for external/internal/loopback
   targets = catch-all fallback, not server-side fetching.
3. **"API chat endpoint"** — `grep api/chat` matched example code inside a docs JS
   chunk, not a live route. Context-check every endpoint match before probing claims.
4. **NSIS installer dead end** — desktop app .exe yielded only a PE stub to `strings`
   and 7z (no embedded secrets worth reporting); don't over-invest in installers when
   the SPA bundle is available. Also: backgrounded long downloads poll silently;
   poll with process tool instead of assuming completion.

## Engagement mechanics that mattered

- Deliverables convention: findings/pocs/evidence/recon + SUMMARY.md index + MANIFEST.txt,
  ALL findings verified with proofs, cleanup of caches/tmp after delivery, exclude the
  session-token file (*.tok) from deliverable archives.
- nuclei/ffuf flag verification: check `tool -h` output once up front for the exact version
  installed instead of trusting remembered flags.
- Login rate-limit cooldowns are minute-scale: batch tests around ONE live session,
  put sleeps inside the probe call, never re-login per batch.

## Findings inventory (final)

| Sev | Finding |
|---|---|
| High | Origin server exposure → full CF bypass (all vhosts incl. payment gateway) |
| High | Inference gateway: zero rate-limit + session-token acceptance (stuffing surface) |
| Med | Registration without email verification + min_password=5 + enum oracle (409) |
| Med | Missing DMARC on apex; gateway domain has NO SPF/DMARC at all |
| Low | Multi-service info disclosure on public origin ports (7+ products) |

Negative results documented as coverage: Flowise traversal/setup takeover blocked,
IDOR none (keys/invoices/topups properly scoped), OAuth redirect_uri locked + state OK,
CORS reflection non-credentialed only, SSRF none.
