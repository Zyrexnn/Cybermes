---
name: reporting
description: Generates standardized, professional Bug Bounty vulnerability reports in markdown format following HackerOne and Bugcrowd standards.
---

# Reporting Skill — Bug Bounty Vulnerability Report

## Purpose
Synthesize validated vulnerabilities and evidence into industry-standard vulnerability reports ready for submission.

## Standard Report Template

```markdown
# [VULN-TYPE] on [AFFECTED_ASSET] leads to [IMPACT]

## 1. Summary
A concise 2-3 sentence overview explaining what the vulnerability is, where it is located, and the highest possible business impact.

## 2. Severity & Classification
- **Severity Rating**: [Critical | High | Medium | Low | Informational]
- **CVSS v3.1 Score**: [e.g. 7.5 (CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N)]
- **CWE**: [e.g. CWE-89, CWE-284, CWE-79]

## 3. Affected Asset(s)
- **URL/Host**: `https://api.target.com/v1/endpoint`
- **Parameter**: `user_id`
- **HTTP Method**: `GET` / `POST`

## 4. Technical Description
Detailed explanation of why the vulnerability occurs:
- Root cause (e.g. missing server-side authorization check, unescaped user input, insecure JWT signing).
- Architecture diagram or request flow if complex.

## 5. Step-by-Step Reproduction
1. Send the following baseline request:
   ```bash
   curl -i -s -k -X GET 'https://api.target.com/v1/endpoint' \
        -H 'Authorization: Bearer <TEST_TOKEN_A>'
   ```
2. Modify parameter `id` to reference user B's account:
   ```bash
   curl -i -s -k -X GET 'https://api.target.com/v1/endpoint?id=999' \
        -H 'Authorization: Bearer <TEST_TOKEN_A>'
   ```
3. Observe that User B's private profile information is returned in the response.

## 6. Evidence & Proof of Concept
### Request:
```http
GET /v1/endpoint?id=999 HTTP/1.1
Host: api.target.com
Authorization: Bearer [REDACTED_TEST_TOKEN]
```

### Response:
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "user_id": 999,
  "email": "victim@example.com",
  "secret_api_key": "redacted_key"
}
```

## 7. Business Impact
Explain the realistic impact to the company, users, and compliance (e.g., unauthorized disclosure of PII, account takeover, full database access).

## 8. Suggested Remediation
Concrete actionable guidance for engineers:
- Code fix suggestions or pseudo-code.
- Configuration adjustments.
- Framework-native defenses.

## 9. References
- [OWASP Reference]
- [CWE Reference]
- [Relevant RFC or vendor documentation]
```

## Strict Output Location & Organization Rules
1. **Target Slugification**: Replace dots, colons, slashes with underscores (e.g., `http://127.0.0.1:8888` -> `127_0_0_1_8888`, `api.target.com` -> `api_target_com`).
2. **Directory Structure Setup**:
   ```bash
   mkdir -p reports/<TARGET_SLUG>/findings reports/<TARGET_SLUG>/pocs reports/<TARGET_SLUG>/evidence
   ```
3. **File Path Conventions**:
   - **Finding Report**: `reports/<TARGET_SLUG>/findings/[<SEVERITY>]_<vuln_name>.md`
   - **Executable PoC**: `reports/<TARGET_SLUG>/pocs/poc_<vuln_name>.py`
   - **Raw Evidence / Screenshots**: `reports/<TARGET_SLUG>/evidence/<filename>`
4. **Summary Aggregation**:
   After writing findings, always trigger the report aggregator to refresh `SUMMARY.md` and `metadata.json`:
   ```bash
   python3 tools/aggregate_reports.py <TARGET_SLUG>
   ```
5. **Anti-Hallucination Gate**:
   Only generate a finding report if a real HTTP request/response or command execution log successfully confirmed the vulnerability. Never create speculative reports.
