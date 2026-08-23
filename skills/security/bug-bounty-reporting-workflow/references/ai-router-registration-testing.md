# AI Router Registration Flow Testing Playbook

## Purpose
This document outlines the methodology for testing **registration flows** in AI API gateways for common weaknesses like missing CAPTCHA, lack of email verification, and weak password policies. This is a **non-destructive** testing approach focused on **read-only validation** of security controls.

---

## Target Scope
- **AI API Gateways** (e.g., New API, One API, Next.js AI Routers)
- **Registration Endpoints** (e.g., `/create-workspace`, `/api/user/register`, `/register`)
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
  - `/create-workspace`
  - `/api/user/register`
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
- **Email:** `testuserN@example.com` (replace `N` with a unique number)
- **Username:** `Test User N`
- **Password:** `TestPass123!` (or a weak password to test policy)

**Example:**
```
Full Name: Test User 1
Email: testuser1@example.com
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
# Example: Create test accounts with unique emails
for i in {1..5}; do
  email="testuser${i}@example.com"
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
  [Step-by-step reproduction command]
```