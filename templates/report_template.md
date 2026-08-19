# Vulnerability Report: [VULNERABILITY TITLE]

## 1. Executive Summary
- **Vulnerability Title**: [Concise Title]
- **Severity**: [Critical / High / Medium / Low / Informational]
- **CVSS v3.1 Score**: [e.g. 8.6 (CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:L/A:N)]
- **CWE**: [CWE-ID - Name, e.g. CWE-89: SQL Injection / CWE-639: IDOR]
- **Affected Asset**: `[Domain / IP / Service]`
- **Affected Endpoint**: `[Method] [Endpoint Path / URL]`

---

## 2. Technical Description & Root Cause
[Provide a clear technical explanation of the flaw, how the backend handles the input, and why the security control fails.]

---

## 3. Prerequisites & Environment
- **Authentication State**: [Unauthenticated / Authenticated as Low-Privilege User]
- **Special Headers / Parameters**: [List required headers or cookies]
- **Target Scope Reference**: [Confirm asset is authorized in scope.yaml]

---

## 4. Step-by-Step Reproduction
1. [Step 1 with precise URL and parameters]
2. [Step 2 with payload injection or token replay]
3. [Step 3 observing the vulnerable server response]

---

## 5. Raw HTTP Request & Response Proofs

### Request:
```http
[RAW_REQUEST_HERE]
```

### Response:
```http
[RAW_RESPONSE_HERE]
```

---

## 6. Controlled Minimal-Impact Proof of Concept (PoC)
```python
# Minimal-impact verification script
import requests

target = "http://target.com/api/endpoint"
# Proof of concept logic here
```

---

## 7. Impact Assessment
[Explain the real-world business and technical impact if exploited by an adversary.]

---

## 8. Remediation Guidance
- **Short-term Mitigation**: [Immediate patch or parameter validation rule]
- **Long-term Fix**: [Architectural fix, parameterized queries, authorization middleware]
- **Reference**: [OWASP / CWE Guidance link]
