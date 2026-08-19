# Autopilot Ledger Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the engine's autonomous hunt loop a cross-engagement memory that captures confirmed findings + dead classes, then skips provably-wasteful agent calls in token-saving modes.

**Architecture:** One new stdlib-only module `engine/memory.py` (capture + skip-decision + gc), import-isolated from siblings. `engine.py` calls capture at its existing `confirm()`/`report` transitions and consults `skip_decision` in the hunt loop only when `--use-memory` is set (default off = full coverage). `engine/state.py` unchanged.

**Tech Stack:** Python 3 stdlib only (`fcntl`, `json`, `os`, `urllib.parse`, `datetime`). No new deps. House test convention: in-file `_selftest()` under `__main__`, run via `python engine/<file>.py` (matches `state.py`/`recon.py`).

## Global Constraints

- **stdlib only** — no third-party imports in `engine/memory.py`.
- **Import isolation** — `engine/memory.py` imports nothing from sibling engine modules; `engine.py` passes state data in.
- **Storage root** — `~/.claude/bughunter/memory/`, overridable via `BUGHUNTER_MEMORY_DIR`.
- **Coverage safety** — skips are OFF by default (`--use-memory` opt-in); a class never hunted is never skipped; every skip is logged.
- **Filenames** — `findings.jsonl`, `negatives.jsonl`, `targets/<host>.json`. Dedup key for findings: `(host, url, param, vuln_class)`.
- **Constants (tuning knobs, top of module)** — `SCHEMA_VERSION=1`, `MAX_BYTES=10*1024*1024`, `KEEP=3`, `DEAD_CLASS_THRESHOLD=5`.
- **House test style** — no pytest fixtures; assert-based `_selftest()` run as a script.
- **Invocation** — script style `python3 engine/memory.py`, never `python -m` (repo has no `__init__.py` in `engine/`).

---

### Task 1: `engine/memory.py` — store, rotation, capture

**Files:**
- Create: `engine/memory.py`

**Interfaces:**
- Produces: `record_finding(engagement:str, host:str, tech_stack:list[str], finding:dict, verdict:dict) -> bool` (True if newly saved, False if dup/invalid); `record_run(engagement:str, tech_stack:list[str], tested:list[str], confirmed:list[dict]) -> None`; helpers `_host(url)->str`, `_stack_sig(list)->str`, `_root()->str`, `_append`, `_read`, `_rotate_if_needed`.

- [ ] **Step 1: Write the module with capture + a failing self-test**

Create `engine/memory.py`:

```python
#!/usr/bin/env python3
"""
memory.py — autopilot ledger: cross-engagement capture + skip decisions.

Lets the autonomous hunt loop skip provably-wasteful agent calls: re-confirming a
known finding, or re-testing a vuln_class that never pays off on this tech stack.
Capture is harvested from the engine's own state; skip is gated (off unless the
engine passes use_memory, i.e. token-saving modes only).

Storage (~/.claude/bughunter/memory/, override BUGHUNTER_MEMORY_DIR):
  findings.jsonl       one row per confirmed finding (cross-target)
  negatives.jsonl      (host, stack_sig, vuln_class) hunted -> not confirmed
  targets/<host>.json  per-target rollup for /pickup

stdlib-only; imports nothing from sibling engine modules.
"""
import fcntl
import json
import os
import sys
from datetime import datetime, timezone
from urllib.parse import urlparse

SCHEMA_VERSION = 1
MAX_BYTES = 10 * 1024 * 1024        # rotate a JSONL past 10 MB
KEEP = 3                            # backups kept: .1 (newest) .. .3
DEAD_CLASS_THRESHOLD = 5           # negatives before a (stack,class) is "dead"


def _root():
    return os.path.expanduser(os.environ.get("BUGHUNTER_MEMORY_DIR", "~/.claude/bughunter/memory"))


def _paths():
    root = _root()
    os.makedirs(os.path.join(root, "targets"), exist_ok=True)
    return root


def _now():
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def _host(url):
    u = url or ""
    if "://" not in u:
        u = "//" + u
    return (urlparse(u).hostname or "").lower().rstrip(".")


def _stack_sig(tech_stack):
    return "|".join(sorted({t.lower() for t in (tech_stack or []) if t}))


def _rotate_if_needed(path):
    try:
        if os.path.getsize(path) < MAX_BYTES:
            return
    except FileNotFoundError:
        return
    oldest = f"{path}.{KEEP}"
    if os.path.exists(oldest):
        os.remove(oldest)
    for i in range(KEEP - 1, 0, -1):
        s, d = f"{path}.{i}", f"{path}.{i + 1}"
        if os.path.exists(s):
            os.replace(s, d)
    os.replace(path, f"{path}.1")


def _append(path, obj):
    line = json.dumps(obj, separators=(",", ":")) + "\n"
    _rotate_if_needed(path)
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_APPEND, 0o644)
    try:
        fcntl.flock(fd, fcntl.LOCK_EX)
        os.write(fd, line.encode("utf-8"))
    finally:
        fcntl.flock(fd, fcntl.LOCK_UN)
        os.close(fd)


def _read(path):
    out = []
    if not os.path.exists(path):
        return out
    with open(path, encoding="utf-8") as f:
        for ln in f:
            ln = ln.strip()
            if not ln:
                continue
            try:
                out.append(json.loads(ln))
            except json.JSONDecodeError:
                continue
    return out


def _valid_finding(e):
    if not (e.get("host") and e.get("url") and e.get("vuln_class")):
        return False
    try:
        datetime.fromisoformat(e["ts"].replace("Z", "+00:00"))
    except Exception:
        return False
    return isinstance(e.get("tech_stack", []), list)


def record_finding(engagement, host, tech_stack, finding, verdict):
    """Capture one confirmed finding. Dedup on (host,url,param,vuln_class)."""
    root = _paths()
    fp = os.path.join(root, "findings.jsonl")
    entry = {
        "schema_version": SCHEMA_VERSION, "ts": _now(), "engagement": engagement,
        "host": host or _host(finding.get("url", "")), "url": finding.get("url", ""),
        "param": finding.get("param", ""), "vuln_class": finding.get("vuln_class", ""),
        "technique": finding.get("technique", ""),
        "severity": (verdict or {}).get("severity") or finding.get("severity", ""),
        "verdict_reason": (verdict or {}).get("reason", ""),
        "tech_stack": list(tech_stack or []),
        "evidence_ref": finding.get("evidence_ref", ""),
    }
    if not _valid_finding(entry):
        print(f"memory: skipping invalid finding {entry.get('url')!r}", file=sys.stderr)
        return False
    key = (entry["host"], entry["url"], entry["param"], entry["vuln_class"])
    for e in _read(fp):
        if (e.get("host"), e.get("url"), e.get("param"), e.get("vuln_class")) == key:
            return False
    _append(fp, entry)
    return True


def _parse_key(k):
    """state 'url|param|class' -> (host, vuln_class)."""
    url, _, rest = k.partition("|")
    _param, _, cls = rest.partition("|")
    return _host(url), cls


def _write_rollup(host, tech_stack, tested, confirmed):
    root = _paths()
    p = os.path.join(root, "targets", f"{host}.json")
    prev = {}
    if os.path.exists(p):
        try:
            prev = json.load(open(p))
        except Exception:
            prev = {}
    hc_conf = [f for f in confirmed if _host(f.get("url", "")) == host]
    doc = {
        "schema_version": SCHEMA_VERSION, "host": host,
        "first_seen": prev.get("first_seen", _now()), "last_seen": _now(),
        "sessions": prev.get("sessions", 0) + 1, "tech_stack": list(tech_stack or []),
        "tested": [k for k in tested if _parse_key(k)[0] == host],
        "confirmed": [{"url": f.get("url"), "param": f.get("param", ""),
                       "vuln_class": f.get("vuln_class"), "severity": f.get("severity", "")}
                      for f in hc_conf],
    }
    tmp = p + ".tmp"
    json.dump(doc, open(tmp, "w"), indent=2)
    os.replace(tmp, p)


def record_run(engagement, tech_stack, tested, confirmed):
    """End-of-run capture: negatives for tested-but-unconfirmed classes + per-host rollup."""
    root = _paths()
    sig = _stack_sig(tech_stack)
    tested_hc = {_parse_key(k) for k in tested}
    confirmed_hc = {(_host(f.get("url", "")), f.get("vuln_class", "")) for f in confirmed}
    neg = os.path.join(root, "negatives.jsonl")
    for host, cls in tested_hc:
        if cls and (host, cls) not in confirmed_hc:
            _append(neg, {"schema_version": SCHEMA_VERSION, "ts": _now(),
                          "host": host, "stack_sig": sig, "vuln_class": cls})
    for host in {h for h, _ in tested_hc} | {h for h, _ in confirmed_hc}:
        if host:
            _write_rollup(host, tech_stack, tested, confirmed)


def _selftest():
    import tempfile, shutil
    d = tempfile.mkdtemp()
    os.environ["BUGHUNTER_MEMORY_DIR"] = d
    try:
        f = {"url": "http://a.com/x", "param": "id", "vuln_class": "idor", "severity": "high"}
        assert record_finding("e1", "a.com", ["nextjs"], f, {"reason": "pii", "severity": "high"}) is True
        assert record_finding("e1", "a.com", ["nextjs"], f, {"reason": "pii"}) is False   # dedup
        record_run("e1", ["nextjs"], ["http://a.com/y|q|xss", "http://a.com/x|id|idor"],
                   [{"url": "http://a.com/x", "vuln_class": "idor"}])
        negs = _read(os.path.join(d, "negatives.jsonl"))
        assert any(n["vuln_class"] == "xss" for n in negs)          # xss tested, not confirmed -> negative
        assert not any(n["vuln_class"] == "idor" for n in negs)     # idor confirmed -> no negative
        assert json.load(open(os.path.join(d, "targets", "a.com.json")))["sessions"] == 1
        print("memory.py self-test: PASS")
    finally:
        shutil.rmtree(d)
        os.environ.pop("BUGHUNTER_MEMORY_DIR", None)


if __name__ == "__main__":
    _selftest()
```

- [ ] **Step 2: Run the self-test, expect PASS**

Run: `python3 engine/memory.py`
Expected: `memory.py self-test: PASS`. (If it errors, the module is broken — fix before committing.)

- [ ] **Step 3: Commit**

```bash
git add engine/memory.py
git commit -m "feat(engine): add autopilot ledger store + capture (record_finding/record_run)"
```

---

### Task 2: `skip_decision`, `rollup`, and the `--gc` CLI

**Files:**
- Modify: `engine/memory.py`

**Interfaces:**
- Consumes: everything from Task 1.
- Produces: `skip_decision(host:str, tech_stack:list[str], item:dict) -> {"skip":bool,"reason":str,"carry":dict|None}`; `rollup(host:str) -> dict|None`; `gc(action:str, max_mb:int, keep:int, root:str|None) -> list[dict]`; a `__main__` that runs `--gc` when args are present, else the self-test.

- [ ] **Step 1: Add skip_decision + rollup + gc + extend the self-test**

Insert these functions **before** `_selftest()` in `engine/memory.py`:

```python
def skip_decision(host, tech_stack, item):
    """Should the hunt loop skip this worklist item? Off-path returns skip=False."""
    root = _paths()
    url, param, cls = item.get("url", ""), item.get("param", ""), item.get("vuln_class", "")
    # 1) known-confirmed -> carry the finding forward, no agent
    for e in _read(os.path.join(root, "findings.jsonl")):
        if (e.get("host"), e.get("url"), e.get("param"), e.get("vuln_class")) == (host, url, param, cls):
            carry = {"url": url, "param": param, "vuln_class": cls,
                     "severity": e.get("severity", ""), "evidence": e.get("verdict_reason", ""),
                     "verdict_reason": e.get("verdict_reason", ""),
                     "verdict": {"real": True, "severity": e.get("severity", ""),
                                 "reason": e.get("verdict_reason", ""), "from_memory": True}}
            return {"skip": True, "reason": "known-confirmed", "carry": carry}
    # 2) dead-class: >= threshold negatives for (stack_sig, cls) AND zero confirmations for it
    sig = _stack_sig(tech_stack)
    negs = sum(1 for e in _read(os.path.join(root, "negatives.jsonl"))
               if e.get("stack_sig") == sig and e.get("vuln_class") == cls)
    if negs >= DEAD_CLASS_THRESHOLD:
        confirmed_here = any(_stack_sig(e.get("tech_stack", [])) == sig and e.get("vuln_class") == cls
                             for e in _read(os.path.join(root, "findings.jsonl")))
        if not confirmed_here:
            return {"skip": True, "reason": f"dead-class ({negs} negatives, 0 confirmed)", "carry": None}
    return {"skip": False, "reason": "", "carry": None}


def rollup(host):
    p = os.path.join(_root(), "targets", f"{host}.json")
    return json.load(open(p)) if os.path.exists(p) else None


def gc(action="report", max_mb=10, keep=KEEP, root=None):
    """report | rotate | purge-backups over every *.jsonl in the ledger dir."""
    base = root or _root()
    rows = []
    for dirpath, _dirs, names in os.walk(base):
        for n in sorted(names):
            if not n.endswith(".jsonl"):
                continue
            live = os.path.join(dirpath, n)
            backups = [f"{live}.{i}" for i in range(1, keep + 1) if os.path.exists(f"{live}.{i}")]
            live_sz = os.path.getsize(live) if os.path.exists(live) else 0
            total = live_sz + sum(os.path.getsize(b) for b in backups)
            if action == "rotate" and live_sz >= max_mb * 1024 * 1024:
                _rotate_if_needed(live)
            if action == "purge-backups":
                for b in backups:
                    os.remove(b)
            rows.append({"file": live, "live_bytes": live_sz, "total_bytes": total,
                         "backups": len(backups)})
    return rows
```

Then **replace** the `__main__` block at the bottom with:

```python
if __name__ == "__main__":
    import argparse
    ap = argparse.ArgumentParser(description="autopilot ledger")
    ap.add_argument("--gc", action="store_true", help="run garbage-collection instead of the self-test")
    ap.add_argument("--rotate", action="store_true", help="with --gc: rotate files over the cap")
    ap.add_argument("--purge-backups", action="store_true", help="with --gc: delete all .1/.2/.3 backups")
    ap.add_argument("--dir", default=None, help="ledger dir (default ~/.claude/bughunter/memory)")
    ap.add_argument("--max-mb", type=int, default=10)
    a = ap.parse_args()
    if a.gc:
        act = "purge-backups" if a.purge_backups else ("rotate" if a.rotate else "report")
        for r in gc(act, a.max_mb, KEEP, a.dir):
            print(f"  {r['file']}: live={r['live_bytes']}B total={r['total_bytes']}B backups={r['backups']}")
    else:
        _selftest()
```

And **extend** `_selftest()` — add before its final `print(...)`:

```python
        # skip_decision: known-confirmed carries the finding
        d = skip_decision("a.com", ["nextjs"], {"url": "http://a.com/x", "param": "id", "vuln_class": "idor"})
        assert d["skip"] and d["reason"] == "known-confirmed" and d["carry"]["verdict"]["real"] is True
        # unseen class is never skipped
        assert skip_decision("a.com", ["nextjs"], {"url": "http://a.com/x", "param": "id", "vuln_class": "rce"})["skip"] is False
        # dead-class only after >= threshold negatives with zero confirmations
        for _ in range(DEAD_CLASS_THRESHOLD):
            record_run("e", ["django"], ["http://b.com/z|q|csrf"], [])
        assert skip_decision("b.com", ["django"], {"url": "http://b.com/z", "param": "q", "vuln_class": "csrf"})["skip"] is True
        # a confirmation on that (stack,class) revives it (not dead)
        record_finding("e", "b.com", ["django"], {"url": "http://b.com/z", "param": "q", "vuln_class": "csrf", "severity": "low"}, {"reason": "x"})
        assert skip_decision("b.com", ["django"], {"url": "http://b.com/z", "param": "q", "vuln_class": "csrf"})["reason"] == "known-confirmed"
```

- [ ] **Step 2: Run the self-test, expect PASS**

Run: `python3 engine/memory.py`
Expected: `memory.py self-test: PASS`

- [ ] **Step 3: Smoke the gc CLI**

Run: `python3 engine/memory.py --gc`
Expected: prints a per-file line for any `*.jsonl` under the ledger dir (or nothing if empty). Exit 0.

- [ ] **Step 4: Commit**

```bash
git add engine/memory.py
git commit -m "feat(engine): ledger skip_decision (known-confirmed/dead-class) + rollup + gc CLI"
```

---

### Task 3: Wire capture into `engine.py`

**Files:**
- Modify: `engine/engine.py` (import; `_tech()` helper; capture after `confirm()`; capture in `report()`)

**Interfaces:**
- Consumes: `memory.record_finding`, `memory.record_run`, `memory._host` from Tasks 1-2.
- Produces: `Engine._tech() -> list[str]` (sorted tech union from state targets).

- [ ] **Step 1: Import memory + add the tech-stack helper**

In `engine/engine.py`, after line 27 (`import agent as A  # noqa: E402`), add:

```python
import memory                     # noqa: E402  (autopilot ledger; import-isolated)
```

Add this method to the `Engine` class (place it just above `def hunt(self):`):

```python
    def _tech(self):
        """Sorted union of tech fingerprints across recon targets (the stack signature source)."""
        return sorted({t for tg in self.eng.state.get("targets", []) for t in tg.get("tech", [])})
```

- [ ] **Step 2: Capture each confirmed finding (validate phase)**

In `validate()`, the confirm call is:

```python
                self.eng.confirm(c, v)
```

Immediately **after** that line, add:

```python
                memory.record_finding(self.scope.name, memory._host(c["url"]), self._tech(), c, v)
```

- [ ] **Step 3: Capture the run at report time**

In `report()`, after the line:

```python
        self.eng.log(f"report: wrote {len(c)} confirmed finding(s) -> {path}")
```

and before `return path`, add:

```python
        memory.record_run(self.scope.name, self._tech(), self.eng.state["tested"], self.eng.state["confirmed"])
```

- [ ] **Step 4: Regression + capture check via the mock engine**

Run:
```bash
python3 engine/state.py
BUGHUNTER_MEMORY_DIR=/tmp/led1 python3 engine/engine.py --scope engine/engagement.example.json --base /tmp/eng1 --mock --hunt
```
Expected: `state.py self-test: PASS`; the mock run finishes; `/tmp/led1/findings.jsonl` exists with ≥1 row and `/tmp/led1/targets/` has a `<host>.json`.

Verify: `cat /tmp/led1/findings.jsonl` shows a confirmed finding; `ls /tmp/led1/targets`.

- [ ] **Step 5: Commit**

```bash
git add engine/engine.py
git commit -m "feat(engine): capture confirmed findings + run rollup into the ledger"
```

---

### Task 4: Skip filter in the hunt loop + `--use-memory` flag

**Files:**
- Modify: `engine/engine.py` (`Engine.__init__` signature; hunt loop; argparse; `Engine(...)` construction)

**Interfaces:**
- Consumes: `memory.skip_decision`, `Engine._tech` from Task 3.
- Produces: `Engine(..., use_memory=False)` constructor param; `--use-memory` CLI flag.

- [ ] **Step 1: Add `use_memory` to the constructor**

Change the `Engine.__init__` signature from:

```python
    def __init__(self, scope_path, base, model, max_hunts, max_turns, timeout, mock=False,
                 allow_intrusive=False, parallel=3, expand=False):
```

to (add `use_memory=False` at the end):

```python
    def __init__(self, scope_path, base, model, max_hunts, max_turns, timeout, mock=False,
                 allow_intrusive=False, parallel=3, expand=False, use_memory=False):
```

And after the line `self.parallel = max(1, parallel)` add:

```python
        self.use_memory = use_memory
```

- [ ] **Step 2: Insert the gated skip filter in the hunt worklist loop**

The current loop in `hunt()` is:

```python
        for it in wl:
            if not self.scope.in_scope_host(it["url"]):     # belt-and-suspenders scope gate
                self.eng.log(f"hunt: REFUSING out-of-scope {it['url']} ({self.scope.reject_reason(it['url'])})")
                self.eng.mark_tested(it)
            else:
                todo.append(it)
```

Replace it with (adds one `elif` branch; scope gate unchanged, default path unchanged):

```python
        tech = self._tech()
        for it in wl:
            if not self.scope.in_scope_host(it["url"]):     # belt-and-suspenders scope gate
                self.eng.log(f"hunt: REFUSING out-of-scope {it['url']} ({self.scope.reject_reason(it['url'])})")
                self.eng.mark_tested(it)
                continue
            if self.use_memory:
                d = memory.skip_decision(memory._host(it["url"]), tech, it)
                if d["skip"]:
                    self.eng.log(f"hunt: skipped via ledger ({d['reason']}): "
                                 f"{it.get('vuln_class')}@{it['url']}")
                    if d["carry"]:
                        self.eng.confirm(d["carry"], d["carry"]["verdict"])
                    self.eng.mark_tested(it)
                    continue
            todo.append(it)
```

- [ ] **Step 3: Add the `--use-memory` CLI flag**

In `main()`, after the `--expand` argument block, add:

```python
    ap.add_argument("--use-memory", action="store_true",
                    help="OPT-IN: skip provably-wasteful agent calls (known-confirmed / dead-class) "
                         "using the autopilot ledger. Default OFF = full coverage.")
```

- [ ] **Step 4: Pass it into the Engine constructor**

Change:

```python
    eng = Engine(a.scope, a.base, a.model, a.max_hunts, a.max_turns, a.timeout, a.mock,
                 a.allow_intrusive, a.parallel, a.expand)
```

to:

```python
    eng = Engine(a.scope, a.base, a.model, a.max_hunts, a.max_turns, a.timeout, a.mock,
                 a.allow_intrusive, a.parallel, a.expand, a.use_memory)
```

- [ ] **Step 5: Verify skip behavior on a second mock run**

Run:
```bash
# Run 1: populate the ledger (no memory used)
BUGHUNTER_MEMORY_DIR=/tmp/led2 python3 engine/engine.py --scope engine/engagement.example.json --base /tmp/eng2 --mock --hunt
# Run 2: same target WITH memory -> known-confirmed items are skipped + logged
BUGHUNTER_MEMORY_DIR=/tmp/led2 python3 engine/engine.py --scope engine/engagement.example.json --base /tmp/eng2b --mock --hunt --use-memory 2>&1 | grep "skipped via ledger"
```
Expected: run 2 prints at least one `hunt: skipped via ledger (known-confirmed): ...` line.

Run the coverage-safety check (default = no skips):
```bash
BUGHUNTER_MEMORY_DIR=/tmp/led2 python3 engine/engine.py --scope engine/engagement.example.json --base /tmp/eng2c --mock --hunt 2>&1 | grep -c "skipped via ledger" || true
```
Expected: `0` (no `--use-memory` = full coverage, no skips).

- [ ] **Step 6: Commit**

```bash
git add engine/engine.py
git commit -m "feat(engine): gated ledger skip filter in hunt loop + --use-memory flag"
```

---

### Task 5: Rewrite the memory command docs

**Files:**
- Modify: `commands/memory-gc.md`, `commands/pickup.md`, `commands/remember.md`, `commands/autopilot.md`

**Interfaces:** none (documentation). Goal: no dangling references to `tools.memory_gc`, `PatternDB`, `AuditLog`, `.claude/settings.json`, `patterns.jsonl`, `journal.jsonl` remain.

- [ ] **Step 1: Rewrite `commands/memory-gc.md`**

Replace the whole file with:

```markdown
---
name: memory-gc
description: Inspect or rotate the autopilot ledger JSONL files (findings.jsonl, negatives.jsonl). Caps file size and keeps N rotated backups so memory does not grow unbounded.
---

# /memory-gc

Garbage-collect the autopilot ledger (`~/.claude/bughunter/memory/`). Reports current sizes, rotates oversized files past a configurable cap, or purges old backups.

## Usage

```
/memory-gc                       # report only
/memory-gc --rotate              # rotate files above 10 MB (default cap)
/memory-gc --rotate --max-mb 5   # custom cap
/memory-gc --purge-backups       # delete all .1/.2/.3 backups
/memory-gc --dir <path>          # scan a non-default ledger dir
```

## Implementation

The agent shells out to, from the repo root:

```bash
python3 engine/memory.py --gc [--rotate] [--purge-backups] [--dir PATH] [--max-mb N]
```

## Defaults

- **Rotation cap:** 10 MB per file · **Backups kept:** 3 (`<file>.1` newest → `<file>.3` oldest)
- **Scope:** every `*.jsonl` under the ledger dir (`findings.jsonl`, `negatives.jsonl`)

Rotation also fires automatically on append inside the ledger writer, so files stay bounded without any session-end hook.
```

- [ ] **Step 2: Rewrite `commands/pickup.md`**

Replace the whole file with:

```markdown
---
name: pickup
description: Pick up a previous hunt on a target — shows hunt history and untested surface from the autopilot ledger. Usage: /pickup target.com
---

# /pickup

Pick up where you left off on a target.

> **Renamed from `/resume`** — `/resume` is a reserved Claude Code command.

## What This Does

1. Reads the target rollup from `~/.claude/bughunter/memory/targets/<host>.json`.
2. Shows hunt history (sessions, last seen, tech stack).
3. Lists confirmed findings and the endpoints already tested.

## Usage

```
/pickup target.com
```

## Implementation

The agent reads the rollup directly:

```bash
python3 -c "import sys; sys.path.insert(0,'engine'); import memory, json; print(json.dumps(memory.rollup('target.com'), indent=2))"
```

## If No Previous Hunt

```
No ledger data for target.com. Run /recon then /autopilot target.com first.
```
```

- [ ] **Step 3: Rewrite `commands/remember.md`**

Replace the whole file with:

```markdown
---
name: remember
description: Optional manual note on a target or the last confirmed finding. Capture is automatic during autopilot; this is for extra context. Usage: /remember
---

# /remember

> **Capture is automatic.** During an engine/autopilot run, every confirmed finding is
> written to the ledger (`~/.claude/bughunter/memory/findings.jsonl`) with no action from you.
> `/remember` is only for adding an **optional manual note** — a technique detail or
> follow-up idea — that the automatic capture would not include.

## Usage

```
/remember   # then describe the note; it is appended to the target's rollup
```

## What It Is Not

This is no longer the primary way findings enter memory — the engine captures those
deterministically. Use it sparingly for human context.
```

- [ ] **Step 4: Update `commands/autopilot.md` — memory section**

In `commands/autopilot.md`, replace the block under `## Safety Guarantees` line
`- **Every request** is logged to \`hunt-memory/audit.jsonl\`` — delete that single stale
line. Then add a new section immediately before `## After Autopilot`:

```markdown
## Ledger memory (token saving)

The autonomous loop can consult the **autopilot ledger** to skip provably-wasteful agent
calls — an item already **confirmed** on a prior run (carried forward, not re-tested) or a
vuln_class with strong **negative** history on this tech stack (`dead-class`).

- **`--paranoid` (default) / `--normal`:** memory is **off** — full coverage, no skips.
- **`--quick` / `--yolo`:** the command passes `--use-memory` to the engine — skips on.
- A vuln_class never hunted on this stack is **never** skipped, and every skip is logged
  (`hunt: skipped via ledger (...)`). Coverage is never silently reduced.
- Capture is automatic; the ledger fills itself as you run. Manage size with `/memory-gc`.
```

- [ ] **Step 5: Verify no dangling references remain**

Run:
```bash
grep -rniE "tools\.memory_gc|PatternDB|AuditLog|patterns\.jsonl|journal\.jsonl|\.claude/settings\.json" commands/ || echo "CLEAN"
```
Expected: `CLEAN`.

- [ ] **Step 6: Commit**

```bash
git add commands/memory-gc.md commands/pickup.md commands/remember.md commands/autopilot.md
git commit -m "docs(commands): rewrite memory commands to the autopilot ledger"
```

---

### Task 6: Update credits + docs pointer, final verification

**Files:**
- Modify: `docs/credits.md`

**Interfaces:** none.

- [ ] **Step 1: Trim the vendored-commands claim in `docs/credits.md`**

`docs/credits.md` currently lists `/remember`, `/memory-gc`, `/pickup` among "Vendored slash
commands | shuvonsec/claude-bug-bounty (MIT)". Those three now describe an **original**
ledger backend, not the vendored interface. Edit the credits so they are no longer listed as
vendored: move `/remember`, `/memory-gc`, `/pickup` out of the vendored list, and add one line
under the author's original work, e.g.:

```markdown
- Autopilot ledger (`engine/memory.py`) + the `/remember`, `/memory-gc`, `/pickup` commands —
  original design and implementation (cross-engagement capture + skip-decision for the engine
  hunt loop). Not derived from any external memory implementation.
```

Leave the remaining genuinely-vendored entries unchanged.

- [ ] **Step 2: Full verification pass**

Run each and confirm:
```bash
python3 engine/memory.py                       # -> memory.py self-test: PASS
python3 engine/state.py                         # -> state.py self-test: PASS (regression)
python3 engine/memory.py --gc                   # -> exits 0
grep -rniE "tools\.memory_gc|PatternDB|AuditLog|patterns\.jsonl|journal\.jsonl" commands/ || echo CLEAN   # -> CLEAN
```
Then the two-run skip proof from Task 4 Step 5 (known-confirmed skip appears with `--use-memory`, absent without).

- [ ] **Step 3: Commit**

```bash
git add docs/credits.md
git commit -m "docs(credits): ledger is original work, not vendored"
```

---

## Self-Review Notes

- **Spec coverage:** capture (Task 3) ✔; negatives + rollup (Tasks 1,3) ✔; skip filter known-confirmed + dead-class (Tasks 2,4) ✔; coverage gate `--use-memory` default off (Task 4) ✔; skips logged (Task 4) ✔; command rewrites (Task 5) ✔; gc CLI (Task 2) ✔; `state.py` untouched / no pyproject change ✔; credits (Task 6) ✔.
- **Type consistency:** `record_finding`/`record_run`/`skip_decision`/`rollup`/`gc` signatures identical across tasks; `skip_decision` return dict shape (`skip`/`reason`/`carry`) consistent; `carry["verdict"]` shape matches `Engagement.confirm(finding, verdict)`.
- **Coverage safety:** `--use-memory` default off (bare run = full coverage); unseen class never skipped (self-test asserts); every skip logged.
