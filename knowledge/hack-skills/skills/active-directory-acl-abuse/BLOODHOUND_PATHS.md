# BloodHound Attack Paths & Cypher Queries

> **AI LOAD INSTRUCTION**: Load this for common BloodHound attack paths, custom Cypher queries for Neo4j, and chain analysis techniques. Assumes the main [SKILL.md](./SKILL.md) is already loaded for individual ACL abuse techniques.

---

## 1. BLOODHOUND DATA COLLECTION BEST PRACTICES

### Collection Methods Comparison

| Method | Speed | Noise | Data |
|---|---|---|---|
| `DCOnly` | Fast | Low | Users, groups, trusts, ACLs (from DC only) |
| `All` | Slow | High | Everything including sessions and local groups |
| `Session` | Medium | Medium | Logged-in user sessions (run multiple times) |
| `ACL` | Medium | Low | ACL data only |
| `ObjectProps` | Fast | Low | Object properties (descriptions, etc.) |

### Stealth Collection

```bash
# Minimum noise — DC only queries
bloodhound-python -d domain.com -u user -p pass -c DCOnly -dc DC01.domain.com

# Add sessions over time
bloodhound-python -d domain.com -u user -p pass -c Session -dc DC01.domain.com

# Avoid SMB enumeration (noisiest)
SharpHound.exe -c DCOnly,ACL --excludedc --stealth
```

---

## 2. ESSENTIAL CYPHER QUERIES

### 2.1 Find All Paths to Domain Admin

```cypher
MATCH p=shortestPath((n)-[*1..]->(m:Group))
WHERE m.name STARTS WITH "DOMAIN ADMINS"
AND n.owned = true
RETURN p
```

### 2.2 Find Users with DCSync Rights

```cypher
MATCH p=(n)-[:MemberOf|GetChanges*1..]->(d:Domain)
MATCH p2=(n)-[:MemberOf|GetChangesAll*1..]->(d)
WHERE n:User OR n:Group
RETURN n.name
```

### 2.3 Kerberoastable Users with Paths to DA

```cypher
MATCH (u:User {hasspn:true})
MATCH p=shortestPath((u)-[*1..]->(g:Group))
WHERE g.name STARTS WITH "DOMAIN ADMINS"
RETURN u.name, length(p) AS hops
ORDER BY hops ASC
```

### 2.4 Users with Dangerous ACLs

```cypher
MATCH p=(n:User)-[r:GenericAll|GenericWrite|WriteDacl|WriteOwner|ForceChangePassword]->(m)
WHERE NOT n.name STARTS WITH "DVTA"
RETURN n.name AS attacker, type(r) AS permission, m.name AS target
```

### 2.5 Find Computers with Unconstrained Delegation

```cypher
MATCH (c:Computer {unconstraineddelegation:true})
WHERE NOT c.name STARTS WITH "DC"
RETURN c.name
```

### 2.6 Find AS-REP Roastable Users

```cypher
MATCH (u:User {dontreqpreauth:true})
RETURN u.name, u.description
```

### 2.7 Computers Where Domain Users Are Local Admin

```cypher
MATCH p=(g:Group {name:"DOMAIN USERS@DOMAIN.COM"})-[:AdminTo]->(c:Computer)
RETURN c.name
```

### 2.8 Find All GPO Controllers

```cypher
MATCH p=(n)-[r:GenericAll|GenericWrite|WriteOwner|WriteDacl]->(g:GPO)
RETURN n.name AS controller, g.name AS gpo, type(r) AS permission
```

### 2.9 Shortest Path from Owned to High-Value Targets

```cypher
MATCH p=shortestPath((n {owned:true})-[*1..]->(m {highvalue:true}))
RETURN p
```

### 2.10 Find LAPS Readers

```cypher
MATCH p=(n)-[:ReadLAPSPassword]->(c:Computer)
RETURN n.name AS reader, c.name AS computer
```

---

## 3. COMMON ATTACK PATH PATTERNS

### Pattern 1: Nested Group Membership → DA

```
lowpriv_user
  └── MemberOf → IT-Support
        └── MemberOf → Server-Admins
              └── MemberOf → Domain Admins
```

```cypher
MATCH p=(u:User {name:"LOWPRIV@DOMAIN.COM"})-[:MemberOf*1..5]->(g:Group {name:"DOMAIN ADMINS@DOMAIN.COM"})
RETURN p
```

### Pattern 2: ACL Chain → GenericAll → Password Reset → DA

```
lowpriv_user
  └── GenericWrite → helpdesk_user
        └── GenericAll → svc_admin (DA member)
              └── ForceChangePassword → reset password → DA
```

### Pattern 3: WriteDACL → DCSync

```
lowpriv_user
  └── WriteDACL on Domain Object
        └── Grant self GetChanges + GetChangesAll
              └── DCSync → all domain hashes
```

### Pattern 4: GPO Abuse → Local Admin on DC

```
lowpriv_user
  └── GenericWrite on GPO linked to "Domain Controllers" OU
        └── Add scheduled task via GPO
              └── Task runs on DCs → SYSTEM on DC
```

### Pattern 5: LAPS + Local Admin → Session Hijack

```
lowpriv_user
  └── ReadLAPSPassword on TARGET_SERVER
        └── Local admin on TARGET_SERVER
              └── Domain Admin session on TARGET_SERVER
                    └── Credential dump → DA hash
```

---

## 4. CUSTOM BLOODHOUND QUERIES FOR SPECIFIC SCENARIOS

### Find Users with Passwords in Description

```cypher
MATCH (u:User)
WHERE u.description =~ '(?i).*(pass|pwd|cred|secret).*'
RETURN u.name, u.description
```

### Find All ACL Paths Between Two Nodes

```cypher
MATCH p=allShortestPaths((a)-[r*1..7]->(b))
WHERE a.name = "LOWPRIV@DOMAIN.COM"
AND b.name = "DOMAIN ADMINS@DOMAIN.COM"
AND ALL(rel IN relationships(p) WHERE type(rel) IN
  ["GenericAll","GenericWrite","WriteDacl","WriteOwner","ForceChangePassword",
   "AddMember","MemberOf","AdminTo","HasSession","CanRDP","CanPSRemote"])
RETURN p
```

### High-Value Target Identification

```cypher
MATCH (n)
WHERE n.highvalue = true
RETURN labels(n)[0] AS type, n.name AS target
ORDER BY type
```

### Find Kerberoastable Service Accounts with Admin Rights

```cypher
MATCH (u:User {hasspn:true})-[:AdminTo]->(c:Computer)
RETURN u.name AS service_account, collect(c.name) AS admin_on
```

### Computers Trusting Machine Accounts for Delegation

```cypher
MATCH (c1:Computer)-[:AllowedToDelegate]->(c2:Computer)
RETURN c1.name AS delegator, c2.name AS target
```

---

## 5. BLOODHOUND CE (COMMUNITY EDITION) TIPS

### API Queries (BloodHound CE)

```bash
# List available attack paths via API
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://bloodhound.local/api/v2/attack-paths" | jq

# Get path findings for a specific domain
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://bloodhound.local/api/v2/domains/$DOMAIN_ID/attack-path-findings" | jq
```

### Mark Nodes as Owned

```cypher
# Mark compromised user
MATCH (u:User {name:"COMPROMISED@DOMAIN.COM"})
SET u.owned = true
RETURN u.name
```

### Mark High-Value Targets

```cypher
# Mark custom high-value targets
MATCH (c:Computer {name:"SQLSERVER01.DOMAIN.COM"})
SET c.highvalue = true
RETURN c.name
```

---

## 6. ATTACK PATH DECISION FLOW

```
BloodHound data collected and imported
│
├── Mark owned principals (compromised users/computers)
│
├── Run "Shortest Paths from Owned to DA"
│   ├── Direct path found? → follow the chain
│   │   ├── ACL edge → exploit per SKILL.md §3
│   │   ├── Session edge → credential dump on that host
│   │   └── AdminTo edge → lateral movement to host
│   └── No path found?
│       ├── Run custom queries (§4) for non-obvious paths
│       ├── Check for Kerberoastable → high-value paths (§2.3)
│       └── Look for LAPS readers → local admin chains (§2.10)
│
├── No path to DA at all?
│   ├── Check for paths to other high-value targets
│   ├── Expand attack surface: compromise more users/hosts
│   ├── Re-collect session data (sessions change over time)
│   └── Look for cross-domain trust paths
│
└── Multiple paths available?
    ├── Prefer: ForceChangePassword (cleanest)
    ├── Then: Shadow Credentials (reversible)
    ├── Then: Targeted Kerberoast (if crackable)
    └── Avoid: Adding to Domain Admins (noisy)
```
