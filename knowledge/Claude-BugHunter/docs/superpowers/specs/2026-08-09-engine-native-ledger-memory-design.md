# Autopilot memory ("the ledger") — design

**Date:** 2026-08-09
**Status:** approved design (scoped to Option A), pre-implementation
**Repo:** elementalsouls/Claude-BugHunter

## Context

The repo ships three memory slash commands — `/remember`, `/memory-gc`, `/pickup` —
with **no backend**: they reference `python -m tools.memory_gc`, `PatternDB`, and
`patterns.jsonl`/`journal.jsonl` that do not exist. Dangling.

A full "cross-engagement memory that reshapes interactive ranking" was considered
and **cut on purpose** (see Scope decision). For a solo interactive hunter that
feature is marginal: an expert already carries stack→class intuition, tech-stack
overlap is a weak predictor, memory risks anchoring you on old patterns, and a
first-time user gets an empty store. The recall-into-ranking loop was speculative
complexity in the engine's hottest file.

**The one place memory earns its keep is autopilot.** The engine's autonomous hunt
loop (`engine.py --hunt --max-hunts N`) spends an agent (~real tokens) per
`(url, param, vuln_class)` worklist item, every run, on every target. An
autonomous agent has **no memory across runs**, so it will:
- re-confirm a finding it already confirmed on a prior run of the same target, and
- re-test a vuln_class that has *never* paid off on this tech stack across many targets.

Both are pure wasted spend. This design captures confirmed findings and dead
classes from the engine's own `state.json`, then lets the autonomous loop **skip
provably-wasteful agent calls** — a measurable token/time saving, not a UX feature.

## Scope decision (what this is and is NOT)

**IS:** capture (confirmed findings + negatives) + a conservative **skip filter in
the engine hunt loop only**, active only in token-saving autopilot modes.

**IS NOT:**
- No recall feeding interactive `rank`/`map` priority. The engine's normal ranking is untouched.
- No map annotations, no `/surface` changes, no interactive-hunter suggestions.
- No autopilot request-loop plumbing (rate limiter / circuit breaker / safe-method guard).
- No changes to Claude-OSINT.
- No Claude Code hooks; capture is engine code.

Original design on our engine — different module, schema, trigger, and purpose
from any external repo. No vendored code; no attribution owed.

## Why this is safe for coverage (the anti-tunnel-vision gate)

Memory that makes you *skip* work can make you miss a novel bug. So skipping is
gated hard:
- **Mode-gated.** `--paranoid` (default, new targets) **ignores memory entirely** —
  full coverage, no skips. Skips apply only in `--quick` / `--yolo` (explicitly
  token-optimizing, familiar-target modes) and behind an overridable `--no-memory` flag.
- **Never skip the unseen.** A vuln_class that has *never* been hunted on this stack
  is never skipped — only classes with strong negative evidence.
- **Never silent.** Every skip is logged (`engine.log` + run summary: "skipped 6 agent
  calls via ledger: 4 known-confirmed, 2 dead-class"). Coverage reduction is always visible.
- **Known-confirmed is carry-forward, not deletion.** A skipped known-confirmed item is
  re-attached to the run's findings from the ledger, not dropped.

## Architecture

One new module, one purpose.

```
engine/memory.py     autopilot ledger: capture + skip-decision (stdlib-only, import-isolated)
   ├── record_finding(engagement_name, host, tech_stack, finding, verdict)   # capture (from confirm)
   ├── record_run(engagement_name, host, tech_stack, tested, confirmed)      # capture (from report/done): negatives + rollup
   ├── skip_decision(host, tech_stack, item) -> {"skip":bool,"reason":str,"carry":finding|None}
   ├── rollup(host) -> dict            # for /pickup
   └── gc(action, max_mb, keep)        # report/rotate/purge; also CLI __main__
```

`engine.py` imports it (`import memory` resolves via the existing
`sys.path.insert(0, dirname(__file__))`), calls capture at transitions it already
passes through, and calls `skip_decision` inside the hunt loop **only when memory
mode is on**. `engine/state.py` is unchanged (read-only consumer). Conventions
match the engine: stdlib-only, no `__init__.py`, in-file `_selftest()` under
`__main__` (like `state.py`/`recon.py`), deterministic. `memory.py` imports
nothing from siblings — `engine.py` passes state data in.

### Storage

Root: `~/.claude/bughunter/memory/` (override via `BUGHUNTER_MEMORY_DIR`).

| File | Purpose | Written by |
|---|---|---|
| `findings.jsonl` | one row per confirmed finding, cross-target | `record_finding` |
| `negatives.jsonl` | per-run `(host, stack_sig, vuln_class)` hunted → not confirmed | `record_run` |
| `targets/<host>.json` | per-target rollup for `/pickup` | `record_run` |

Append serialized under an `fcntl` advisory lock (correctness boundary — parallel
engine runs must not interleave a line). Files size-rotate at 10 MB, keep 3 backups,
rotation fires on append (bounded without a session-end hook) and via `/memory-gc`.
`tech_stack` is normalized to a sorted **stack signature** (lowercased token set) so
negatives aggregate across targets sharing a stack.

### Schemas

`findings.jsonl`:
```json
{"schema_version":1,"ts":"...Z","engagement":"acme-bb","host":"app.acme.com","url":"https://app.acme.com/api/users",
 "param":"id","vuln_class":"idor","technique":"numeric_id_swap","severity":"high",
 "verdict_reason":"confirmed PII exposure for arbitrary id","tech_stack":["graphql","nextjs","postgres"],
 "evidence_ref":"evidence/idor-1.txt"}
```
Dedup key: `(host, url, param, vuln_class)`.

`negatives.jsonl` (one row per (run, host, class) that was tested and not confirmed):
```json
{"schema_version":1,"ts":"...","host":"app.acme.com","stack_sig":"graphql|nextjs|postgres","vuln_class":"xss"}
```

`targets/<host>.json`: `{schema_version, host, first_seen, last_seen, sessions, tech_stack,
tested[], confirmed[<finding subset>], untested[<worklist subset>]}`.

Light validation on write (required fields, ISO `ts`, `tech_stack` list-of-str); invalid →
skip with stderr warning, never crash a run.

## Capture (deterministic, engine-native)

Additive calls in `engine.py` at existing transitions:

1. **After `self.eng.confirm(c, v)`** (validate phase, engine.py:301): call
   `memory.record_finding(name, host_of(c), tech_stack_of(host), c, v)`. `tech_stack`
   from `state["targets"]` fingerprint. Every confirmed finding is captured automatically.
2. **In the `report` phase** (engine.py:328): `memory.record_run(name, host, tech_stack, state["tested"], state["confirmed"])`:
   - Update `targets/<host>.json` (sessions++, tested, confirmed, untested worklist, tech_stack).
   - For each `vuln_class` present in `tested` but absent from `confirmed` for the host,
     append a `negatives.jsonl` row keyed by the host's stack signature.

Wiring capture into `confirm()`/`report` in code = fires on 100% of engine runs. No
manual `/remember`, no agent reliance.

## Skip filter (the autopilot saving — hunt loop only)

In the hunt phase, the loop is `wl = self.eng.worklist()[:self.max_hunts]` then per item
spawn an agent (engine.py:225–244). Add, **only when `self.use_memory` is true**, before
spawning:

```
d = memory.skip_decision(host, tech_stack, item)
if d["skip"]:
    self.eng.log(f"hunt: skipped via ledger ({d['reason']}): {item_key}")
    if d["carry"]: self.eng.confirm(d["carry"], d["carry"]["verdict"])   # carry-forward known finding
    self.eng.mark_tested(item)
    continue    # no agent spent
```

`skip_decision` returns skip=True in exactly two cases:
- **known-confirmed:** `(host, url, param, vuln_class)` already in `findings.jsonl` →
  `reason="known-confirmed"`, `carry=<that finding>`.
- **dead-class:** `(stack_sig, vuln_class)` has ≥ `DEAD_CLASS_THRESHOLD` negative rows
  across targets **and zero** confirmations anywhere for that `(stack_sig, vuln_class)` →
  `reason="dead-class"`, `carry=None`.

Otherwise skip=False. Thresholds/weights are named constants at the top of `memory.py`
(a calibration knob — which classes actually recur needs tuning a fixed model can't guess);
`DEAD_CLASS_THRESHOLD` defaults conservative (e.g. 5).

`engine.py` gets a **single `--use-memory` flag, default OFF** (so a bare engine run keeps
full coverage — safe by default). The mode→flag mapping lives where the modes live, in
`commands/autopilot.md`: `--paranoid`/default → do **not** pass `--use-memory` (full
coverage); `--quick`/`--yolo` → pass `--use-memory` (token-saving skips on). This keeps
`engine.py` unaware of the slash-command mode names — it only knows the one flag.

## Commands (rewritten to this system)

- **`/pickup <host>`** — read `targets/<host>.json`: history (sessions, findings), untested
  surface, and the host's confirmed findings. No interactive recall/suggestions.
- **`/memory-gc`** — report / `--rotate` / `--purge-backups` / `--dir` / `--max-mb` via
  `python3 engine/memory.py --gc ...` (script invocation, matching `python3 engine/engine.py`).
  Rewrite doc to our filenames + entrypoint; drop `AuditLog`/`tools.memory_gc`/Stop-hook prose.
- **`/remember`** — optional manual annotation (note/tag on a host or the last confirmed
  finding). No longer the primary path; capture is automatic.
- **`commands/autopilot.md`** — document the memory-skip behavior, the mode→`--use-memory`
  mapping (`--quick`/`--yolo` on; `--paranoid`/default off), and that skips are logged, never silent.

## Testing

`engine/memory.py` `_selftest()` under `__main__`, asserting:
- `record_finding` dedup on `(host, url, param, vuln_class)`.
- `skip_decision` = **known-confirmed** (returns carry) for a captured finding's exact key.
- `skip_decision` = **dead-class** once negatives ≥ threshold with zero confirmations; and
  **NOT dead-class** if a confirmation exists for that `(stack_sig, class)`, or if never hunted.
- rotation at the size cap; `gc --purge-backups` removes `.1`..`.3`.

Round-trip (integration): run the engine mock (`--mock --hunt`, mock verdicts at engine.py:381)
→ confirm a finding → assert a `findings.jsonl` row → run the same target again in `--yolo`
→ assert that item is **skipped as known-confirmed** (agent not spawned, finding carried) and
the skip is logged.

## Verification (from repo root)

1. `python engine/memory.py` → `memory.py self-test: PASS`.
2. `python engine/state.py` still passes (unchanged; regression guard).
3. `python3 engine/engine.py --scope engine/engagement.example.json --base /tmp/eng --mock --hunt`
   twice on the same target: run 1 populates `findings.jsonl`; run 2 in `--yolo` logs
   "skipped N via ledger" and spawns fewer agents.
4. `--paranoid` run performs **no** skips (full coverage regression).
5. `python3 engine/memory.py --gc` reports; `--gc --rotate` exits 0.
6. Grep clean: no dangling `tools.memory_gc`/`PatternDB` references remain in commands/.

## Files

**New:** `engine/memory.py`; this spec.
**Modified:** `engine/engine.py` (capture in confirm/report; gated skip filter + a single
`--use-memory` flag, default off); `commands/{pickup,memory-gc,remember,autopilot}.md`
(autopilot.md maps `--quick`/`--yolo` → `--use-memory`).
**Untouched:** `engine/state.py`, `pyproject.toml` (script-invoked, no packaging change), Claude-OSINT.

## Value framing (honest)

Value is **measurable agent-spend reduction on repeat/autonomous hunts** — not an
interactive-hunter feature. Expect ~zero benefit on a first run or a never-before-seen
stack (cold start), and real benefit only as the ledger accumulates repeat targets and
recurring stacks. Coverage is never silently reduced; `--paranoid` always runs full.

## Delivery

Branch `feat/ledger-memory`; one PR:
`feat(engine): autopilot ledger — capture findings + skip provably-wasteful agent calls`.
Squash-merge.
