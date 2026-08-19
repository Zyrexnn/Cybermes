# ENGAGEMENTS

This file records the **authorized engagements** that drove development of the bundle's skills. It exists so that readers can calibrate the README's "battle-tested" framing against real targets, rather than against the deliberately-vulnerable public training platforms (DVWA, OWASP Juice Shop, testphp.vulnweb, Hacker101) that the bundle is also exercised against.

Specifics are intentionally redacted per signed SoWs / NDAs. What remains is enough to let a future reader judge whether the bundle's content was tested against real, defended, internet-exposed enterprise infrastructure — versus only against lab targets.

---

## Redaction Rules

**Present here:**
- Engagement **type** (bug-bounty / WAPT / red-team / pentest)
- **Quarter-year** only (no specific date)
- **Continent-level region** (Europe / Asia / North America / etc.)
- **Tech-stack class** at the level of abstraction that calibrates impact without naming a vendor + version chain that uniquely identifies a target
- Skills **produced or extended** as a result (these skills are themselves public in this repo)
- **Finding count by severity tier** (no per-finding specifics)
- **Lesson narrative** describing what gap in the bundle the engagement exposed

**Intentionally absent:**
- Client / company / brand / product / employee names
- Specific URLs, IP addresses, hostnames, internal AD domain names
- Mobile-app package identifiers
- Per-finding payload, request body, screenshot, CVE-pending number, advisory ID
- Specific months within the quarter (intentional abstraction)
- Anything an attacker could correlate against Wayback / breach corpora / OSINT to identify the target post-publication

---

## Engagement Index

| # | Period | Type | Region | Tech-stack class | Findings (C/H/M/L) | Skills produced/extended |
|---|---|---|---|---|---|---|
| 01 | 2026 Q2 | WAPT (bug-bounty mode) | Europe | Internet-exposed on-prem CMS farm, EoL since 2023 | 3 / 0 / 2 / 6 | 3 new + 2 extended |
| 02 | 2026 Q2 | External red-team | Asia | Mobile-heavy external surface + cloud identity fabric | 2 / 4 / 5 / 3 | 4 new + 1 extended |

*Future engagements will be added as authorizations permit; some prior engagements are not listed because the SoW / NDA does not permit even abstracted disclosure.*

---

## Engagement 01 — 2026 Q2 · WAPT (bug-bounty mode)

**Region:** Europe

**Tech-stack class:**
- Internet-exposed Microsoft SharePoint Server farm on a version line that reached end-of-extended-support in early 2023
- Hosted behind a US-based public-cloud Layer-7 load balancer in an EU region, fronting on-prem IIS back-ends
- Dual-auth configuration (custom Forms-auth UI + NTLM challenge available on every API endpoint)
- Internal Active Directory was a child of a parent corporate AD forest

**Findings:** **3 Critical** · **0 High** · **2 Medium** · **6 Low / Informational** (11 total)

The three Criticals were each verified with two independent reproduction tools (`curl` + Python raw socket + Burp Repeater per the bundle's Multi-Tool Reproduction Bar). The two Mediums included the EoL-software exposure and an Active-Directory-topology disclosure via the NTLM Type-2 challenge. The six Low/Info findings covered stack-trace disclosure surfaces, permissive CSP, anonymous API metadata enumeration, and load-balanced ViewState handoff failures.

**Skills produced:**
- `hunt-sharepoint` — new (SharePoint Server attack surface)
- `hunt-aspnet` — new (ASP.NET-specific surface: ViewState, machineKey, request validator, dual-parser anti-pattern)
- `hunt-ntlm-info` — new (NTLM Type-2 challenge AD-topology disclosure)

**Skills extended:**
- `bb-methodology` — added PART 0 (Mode-Confirmation Gate), PART 4 (Methodology Discipline: Marker / Body-Diff / Statistical-Sample / Shell-Loop rules), Pushback Protocol, Multi-Tool Reproduction Bar
- `triage-validation` — added Pre-Severity Gate and Retraction Discipline
- `hunt-auth-bypass` — added Legacy-Protocol Matrix (mapping target tech to legacy login endpoints, e.g. SharePoint's `/_vti_bin/Authentication.asmx` as the WordPress XMLRPC equivalent)

**Gap exposed → bundle change:** The engagement's highest-impact finding came from probing a legacy SOAP login endpoint that the branded login UI made everyone forget about — the WordPress-XMLRPC-bypass pattern applied to SharePoint. The bundle had `hunt-auth-bypass` loaded during the engagement, but the skill described only WordPress / XMLRPC. The Legacy-Protocol Matrix was added afterwards so that future engagements against any custom-branded login UI prompt an immediate probe of the platform's legacy login endpoint.

A first-pass on the engagement produced five hygiene-tier findings (EoL software, permissive CSP, stack traces) framed as a red-team report. After pushback, the second pass — using bug-bounty discipline (impact-only) and walking 3 more `hunt-*` skill checklists end-to-end — produced the 11 findings above. This experience drove the Mode-Confirmation Gate (PART 0 of `bb-methodology`) and the Pushback Protocol.

Four false positives also surfaced during the engagement, each retracted after verification: a marker-collision reflection bug, a status-code-only Host-header "bypass" with byte-identical body, a single-sample timing leak that collapsed under n=80 reproduction, and a server-policy blocklist mistaken for a file-existence oracle. All four became the `feedback-false-positive-discipline` content in the discipline-rule additions to `bb-methodology` PART 4.

---

## Engagement 02 — 2026 Q2 · External red-team

**Region:** Asia

**Tech-stack class:**
- Public-internet attack surface for an enterprise organisation with a heavy mobile-app catalogue (multiple Android applications distributed via Google Play)
- Cloud identity fabric in active use (Entra / M365 / OAuth) including OAuth flows where claim fields are scanned with substring matching
- On-prem and cloud assets co-existing; identity provider federated with an internal AD
- Conventional perimeter (SSL VPN, web portal, ERP-adjacent endpoints) plus the mobile attack surface

**Findings:** **2 Critical** · **4 High** · **5 Medium** · **3 Low / Informational** (14 total)

Findings were packaged into a client-facing red-team deliverable with embedded screenshots (per the format documented in `redteam-report-template`). Severity distribution skews toward High and Medium because the engagement was scoped as external red-team (which accepts hygiene + recon findings as deliverables), distinct from the bug-bounty mode that drove Engagement 01.

**Skills produced:**
- `apk-redteam-pipeline` — new (Android APK acquisition, jadx decompile, secret-grep, pinned-cert extraction, exported-component enum, Frida instrumentation templates)
- `mid-engagement-ir-detection` — new (methodology for detecting client SOC patches, attacker activity, and security-state changes that occur DURING a red-team engagement)
- `redteam-mindset` — new (operator-discipline corrections that separate offensive testing from defensive WAPT)
- `redteam-report-template` — new (client-facing red-team deliverable format: Subject / Observations / Description / Impact / Recommendation / PoC structure with DOCX rendering pipeline)

**Skills extended:**
- `bb-methodology` — added the Pushback Protocol's worked example (the lesson that "if a user authorizes full engagement, no mid-run permission gates — discipline rules govern finding-correctness, not effort-throttling")

**Gaps exposed → bundle change:**

1. **Mid-engagement security-state changes** — during testing, the client's SOC detected and **patched a confirmed SQLi within ~30 minutes** of the first probe, AND a separate external attacker was observed locking accounts of legitimate users during the test session (mid-engagement IR activity). Initially this was treated as "test invalidated, retract." The corrected handling — keep the finding with timestamped pre-patch evidence + document the IR detection as itself a deliverable observation — became `mid-engagement-ir-detection`. The "don't retract confirmed findings just because they stop reproducing — assume client patched" rule is core to the skill.

2. **Operator capability assumption** — the engagement's authorization gave broader scope than initial conservative defaults assumed, causing multiple findings to be missed during the first pass. Lesson: when the user says "I have credentials," default to LEAST capability (creds only — no MFA device, no endpoint compromise, no browser session) and ask before describing operator-assisted flows. Became `feedback_operator_capability_assumption` and shaped `redteam-mindset`.

3. **OAuth claim-field substring traps** — Microsoft's MFA-required (AADSTS50076) and Conditional-Access-claims-challenge response bodies contain the literal text `access_token` as a substring inside the claims-challenge JSON, NOT as an actual issued token. A naive substring-match in tooling treats this as "auth bypass succeeded." Lesson: always JSON-parse, never substring-match. Became `feedback_oauth_substring_trap` in the memory layer and informs the OAuth and identity-fabric coverage in `hunt-oauth` and `m365-entra-attack`.

4. **APK acquisition pipeline gaps** — multiple targeted APKs had truncated downloads via standard tooling (apkpure / apkmirror), requiring a fallback chain (Play Store extractor → apkpure → apkmirror) and manual verification. The lesson became `apk-redteam-pipeline`'s acquisition section, which now codifies the fallback ordering and the truncation-detection step that triggers a retry.

5. **The 8.5/10 honest-revalidation experience** — after this engagement, an internal "is the bundle as strong as we claim" revalidation downgraded the bundle's self-assessment from 10/10 to 8.5/10. That triggered the public-critique response cycle (Workstreams A through F in this repo's commit history) that produced this file.

---

## Skill-to-Engagement Map (reverse index)

| Skill (in `skills/`) | Authored / Extended By |
|---|---|
| `hunt-sharepoint` | Engagement 01 |
| `hunt-aspnet` | Engagement 01 |
| `hunt-ntlm-info` | Engagement 01 |
| `bb-methodology` (PART 0, PART 4, Pushback Protocol, Multi-Tool Reproduction Bar) | Engagement 01 (+ Engagement 02 worked example) |
| `triage-validation` (Pre-Severity Gate, Retraction Discipline) | Engagement 01 |
| `hunt-auth-bypass` (Legacy-Protocol Matrix) | Engagement 01 |
| `apk-redteam-pipeline` | Engagement 02 |
| `mid-engagement-ir-detection` | Engagement 02 |
| `redteam-mindset` | Engagement 02 |
| `redteam-report-template` | Engagement 02 |

Every other skill in `skills/` is **report-curated** (built from public disclosed bug-bounty / coordinated-disclosure / CVE corpus — see each skill's `## Disclosed Report Citations` section and frontmatter `sources:`) rather than engagement-derived. Both kinds of skill are valuable; the distinction matters because they're calibrated differently. Engagement-derived skills carry one real PoC's worth of evidence and the lessons of one live target. Report-curated skills carry the pattern depth of many real targets without first-hand reproduction.

---

## Calibration Notes for the README's "Battle-Tested" Claim

The README states the bundle is "battle-tested across authorized red-team and bug-hunting engagements, plus public training platforms (DVWA, OWASP Juice Shop, Hacker101, testphp.vulnweb.com)."

Concretely, "battle-tested" means:
- **Two authorized engagements documented here** (more not listed for SoW reasons).
- **One pre-publication revalidation cycle** (the 8.5/10 honest re-grade after Engagement 02) that drove substantive content additions in Workstreams A-F (report-curation backfill across 11 skills, 5 missing 2024-2026 surfaces added, 3 zero-report skills moved to ≥6 citations each, 5 chain-composition sections added to the high-volume skills, HTTP/2 single-packet deep reference added to `hunt-race-condition`).
- **The training-platform exercises** are separately useful — primarily for new operators learning to use the bundle — but should NOT be conflated with the authorized engagements. The training platforms are deliberately-vulnerable; finding bugs in them validates the operator, not the bundle.

What "battle-tested" does NOT mean (intentional honesty):
- It does not mean every `hunt-*` skill in the bundle was used in every engagement. The engagement count is small; the bundle is broad. Most skills here have not seen a live engagement reproduction yet, and that fact is visible in each skill's `sources:` frontmatter (`engagement_*` vs `hackerone_public` / `github_security_advisories`).
- It does not mean the engagements were comprehensive penetration tests of those clients. They were scoped engagements with specific deliverables; what they tested is reflected in the skills produced, not in the totality of what the clients' attack surfaces contain.

This file is the calibration source. Readers who want to know "is the bundle's content backed by real engagement evidence" can read here and make their own determination at the abstraction level the SoWs permit.

---

## How to Read the Frontmatter `sources:` Field on Individual Skills

Each `skills/*/SKILL.md` declares its evidence basis. The convention used in this repo:

- `sources: hackerone_public, github_security_advisories, ...` — skill is **report-curated** from public disclosed corpus. Citations live in the `## Disclosed Report Citations` section of that skill.
- `sources: authorized-engagement` — skill is **engagement-derived** from one of the engagements listed above. The lessons codified in the skill came from first-hand reproduction in a real target rather than from public reports.
- `sources: <mixed>` — skill has both report-curated patterns AND engagement-derived sections (e.g. `cloud-iam-deep` after Workstream B added the Cognito Identity Pool chain).

The honest claim for any skill in the bundle is what its `sources:` field says — not the README's headline. The README is the marketing; the skill frontmatter is the contract.
