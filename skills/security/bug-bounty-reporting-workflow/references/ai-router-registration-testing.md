# AI Router Registration Flow Testing Playbook

## Purpose
This document outlines the methodology for testing **registration flows** in AI API gateways (e.g., Xyrus Router, New API, One API) for common weaknesses like missing CAPTCHA, lack of email verification, and weak password policies. This is a **non-destructive** testing approach focused on **read-only validation** of security controls.

---

## Target Scope
- **AI API Gateways** (e.g., Xyrus Router, New API, One API)
- **Registration Endpoints** (e.g., `/create-workspace`, `/api/user/register`)
- **Authentication Flows** (e.g., auto-login post-registration, email verification)

---

## Testing Workflow

### 1. Identify Registration Endpoints
- **Manual Inspection:** Navigate the SPA to find registration links (e.g., "Create account", "Sign up", "Get started").
- **JS Bundle Mining:** Search for registration-related routes in the main JS bundle:
  ```bash
  curl -s "https://TARGET/" | grep -oE '/static/js/[a-zA-Z0-9._-]+\.js' | head -1
  curl -s "https://TARGET/static/js/<bundle>.js" -o /tmp/bundle.js
  grep -oE '"/api/[a-zA-Z0-9_/{}.-]+"' /tmp/bundle.js | grep -i "register\|signup\|create" | sort -u
  ```
- **Common Paths:**
  - `/create-workspace` (Xyrus Router)
  - `/api/user/register` (New API / One API)
  - `/register`
  - `/signup`

### 2. Test Registration Form Controls
Use **browser automation** to submit test registrations and observe the behavior. Example test cases:

| Test Case | Expected Secure Behavior | Insecure Behavior | Severity |
|-----------|--------------------------|-------------------|----------|
| **No CAPTCHA** | Form includes CAPTCHA (reCAPTCHA, hCaptcha, Turnstile) | No CAPTCHA present | Medium (Bot Abuse Risk) |
| **No Email Verification** | Registration requires email verification (code/link) | Auto-login or immediate access | Medium (Fake Account Risk) |
| **Weak Password Policy** | Password requires complexity (min 8 chars, mixed case, symbols) | Accepts simple passwords (e.g., `TestPass123!`) | Low (Brute-Force Risk) |
| **Rate Limiting** | Returns `429 Too Many Requests` after N attempts | Accepts unlimited registrations | Medium (Spam Risk) |
| **Username/Email Enumeration** | Generic error message (e.g., "Invalid credentials") | Specific error (e.g., "Email already exists") | Low (Privacy Risk) |

### 3. Test Data for Registration
Use **dummy data** to avoid impacting real users:
- **Email:** `testuserN@temp-mail.org` (replace `N` with a unique number)
- **Username:** `Test User N`
- **Password:** `TestPass123!` (or a weak password to test policy)

**Example (Xyrus Router):**
```
Full Name: Test User 1
Email: testuser1@temp-mail.org
Password: TestPass123!
```

### 4. Observe Post-Registration Behavior
After submitting the form, check:
1. **Redirect Destination:**
   - Secure: Redirects to a verification page (e.g., "Check your email").
   - Insecure: Auto-logins to the dashboard (no verification).
2. **Email Verification:**
   - Secure: Requires clicking a link or entering a code.
   - Insecure: No email sent or verification skipped.
3. **Session State:**
   - Secure: No session created until verification.
   - Insecure: Session active immediately (auto-login).

### 5. Automate Testing for Multiple Accounts
To test **rate limiting** or **bulk registration**, use a script to create multiple accounts sequentially:
```bash
# Example: Create 10 test accounts with unique emails
for i in {1..10}; do
  email="testuser${i}@temp-mail.org"
  curl -s -X POST "https://TARGET/api/user/register" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"TestPass123!\",\"name\":\"Test User $i\"}" \
    -o "response_${i}.json"
  sleep 2  # Rate-limiting delay
  echo "Account $i: $(grep -oE '"message":"[^"]+"' response_${i}.json)"
done
```
**Note:** Only run this if you have **explicit authorization** to test bulk registration.

---

## Findings Template
Use this template to document registration flow weaknesses:

```markdown
### Finding: [Title]
- **Severity:** [Low/Medium/High]
- **CVSS:** [e.g., CVSS 3.1: AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:L/A:N (5.3)]
- **CWE:** [e.g., CWE-287 (Improper Authentication), CWE-799 (Missing Rate Limiting)]
- **Description:** [Brief description of the issue]
- **Evidence:**
  ```
  [Request/Response snippets or screenshots]
  ```
- **Impact:** [What an attacker could do]
- **Remediation:** [How to fix the issue]
- **Reproduction Steps:**
  1. [Step 1]
  2. [Step 2]
  3. [Step 3]
```

---

## Example Findings

### 1. Missing CAPTCHA in Registration
- **Severity:** Medium
- **CVSS:** CVSS 3.1: AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:L/A:N (4.3)
- **CWE:** CWE-799 (Missing Rate Limiting)
- **Description:** The registration form does not include CAPTCHA, allowing automated bots to create accounts at scale.
- **Evidence:**
  - Screenshot of registration form (no CAPTCHA field).
  - Successful registration of 10 test accounts without CAPTCHA.
- **Impact:** Attackers can create thousands of fake accounts for spam, abuse, or credential stuffing.
- **Remediation:** Add CAPTCHA (e.g., reCAPTCHA, hCaptcha, or Turnstile) to the registration form.
- **Reproduction Steps:**
  1. Navigate to `/create-workspace`.
  2. Fill in the form with dummy data.
  3. Submit the form without solving CAPTCHA.
  4. Observe successful account creation.

### 2. No Email Verification
- **Severity:** Medium
- **CVSS:** CVSS 3.1: AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:L/A:N (5.3)
- **CWE:** CWE-287 (Improper Authentication)
- **Description:** Registration completes without email verification, allowing attackers to create accounts with fake or victim email addresses.
- **Evidence:**
  - Auto-login to dashboard immediately after registration.
  - No email received at the provided address.
- **Impact:** Attackers can impersonate users, create fake accounts, or abuse the system.
- **Remediation:** Enable email verification and require users to confirm their email before accessing the dashboard.
- **Reproduction Steps:**
  1. Register with a dummy email (e.g., `testuser1@temp-mail.org`).
  2. Observe immediate redirect to the dashboard.
  3. Confirm no verification email was sent.

### 3. Weak Password Policy
- **Severity:** Low
- **CVSS:** CVSS 3.1: AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:L/A:N (3.7)
- **CWE:** CWE-521 (Weak Password Requirements)
- **Description:** The registration form accepts weak passwords (e.g., `TestPass123!`), increasing the risk of brute-force attacks.
- **Evidence:**
  - Successful registration with password `TestPass123!`.
  - No error message about password complexity.
- **Impact:** Attackers can guess or brute-force passwords more easily.
- **Remediation:** Enforce a strong password policy (min 12 chars, mixed case, numbers, symbols).
- **Reproduction Steps:**
  1. Register with a weak password (e.g., `TestPass123!`).
  2. Observe successful account creation.

---

## Tools
- **Browser Automation:** Use Hermes `browser_*` tools to interact with registration forms.
- **cURL:** For API-based registration testing.
- **Rate Limiting:** Always include delays (`sleep 2`) between requests to avoid triggering WAF/Cloudflare.

---

## Safety Notes
1. **Non-Destructive Testing:** Only test registration flows with **dummy data** (e.g., `testuserN@temp-mail.org`).
2. **Cleanup:** Delete test accounts after testing if possible.
3. **Authorization:** Ensure you have **explicit permission** to test registration flows, especially for bulk testing.
4. **Stealth:** Avoid triggering rate limits or WAF blocks. Use delays between requests.

---

## References
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [CWE-287: Improper Authentication](https://cwe.mitre.org/data/definitions/287.html)
- [CWE-799: Missing Rate Limiting](https://cwe.mitre.org/data/definitions/799.html)
- [CWE-521: Weak Password Requirements](https://cwe.mitre.org/data/definitions/521.html)