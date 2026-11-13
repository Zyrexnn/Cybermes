## What's Changed

Cybermes `v3.5.1` introduces an automated update system for the NPM package, native Go Nmap service fingerprinting, finding deduplication in report aggregators, hardened scope guards, and security fixes across MCP client injection and PDF generation.

---

### 🚀 Features & Enhancements

- **MCP NPM Auto-Updater**:
  - Added `cybermes-mcp update` command to verify and update the CLI package across user environments ([29b9314](https://github.com/Zyrexnn/Cybermes/commit/29b9314)).
  - Implemented background check notifications with cache-aware throttling.
- **Nmap Fingerprinting & Port Scanning**:
  - Built-in native Go Nmap wrapper supporting targeted service discovery and NSE execution (`pkg/nmap`, `pkg/mcp/tools_nmap.go`).
  - Added automatic URL scheme stripping for robust target IP/hostname extraction ([1752c0a](https://github.com/Zyrexnn/Cybermes/commit/1752c0a), [9cde24d](https://github.com/Zyrexnn/Cybermes/commit/9cde24d)).
- **Finding Deduplication & Quality Gates**:
  - Automatic finding deduplication engine in report aggregator to eliminate redundant vulnerability entries (`pkg/report/dedup.go`, [#49](https://github.com/Zyrexnn/Cybermes/pull/49)).
  - Strict finding filename validation (snake_case, severity prefixes, no square brackets) to prevent shell globbing issues ([1e6445c](https://github.com/Zyrexnn/Cybermes/commit/1e6445c)).
  - Automated suppression of informational findings spam in report summaries.
- **Scope Guard Hardening**:
  - Sensitive domain blacklist and expanded path exclusions to prevent unintended scans ([74539a2](https://github.com/Zyrexnn/Cybermes/commit/74539a2)).
  - Eliminated fail-open wildcard defaults in scope matching ([d337e4a](https://github.com/Zyrexnn/Cybermes/commit/d337e4a)).
  - Added support for dynamic target overrides and automated per-target workspace scaffolding ([fdf5c16](https://github.com/Zyrexnn/Cybermes/commit/fdf5c16), [a68fdda](https://github.com/Zyrexnn/Cybermes/commit/a68fdda)).
- **Skill Engine & Diagnostics**:
  - Added nested skill directory resolution in MCP tools ([9cde24d](https://github.com/Zyrexnn/Cybermes/commit/9cde24d)).
  - Strict YAML frontmatter validator for skill definitions ([2a30944](https://github.com/Zyrexnn/Cybermes/commit/2a30944)).
  - Extended `tools/doctor.py` to verify Go native toolchain compilation and MCP binary status ([887f0ac](https://github.com/Zyrexnn/Cybermes/commit/887f0ac)).

---

### 🛡️ Security & Bug Fixes

- **MCP Server Hardening**:
  - Unified target slug sanitization using `scope.SanitizeSlug` ([57e006d](https://github.com/Zyrexnn/Cybermes/commit/57e006d), [#77](https://github.com/Zyrexnn/Cybermes/pull/77)).
  - Hardened client injection adapters and configuration paths across Claude, Codex, and Cursor integrations ([da087f5](https://github.com/Zyrexnn/Cybermes/commit/da087f5)).
  - Sanitized PDF report generation parameters and added automated security regression tests ([tools/generate_pdf.py](file:///home/ikhsan/Documents/Cybermes/tools/generate_pdf.py), [tools/test_generate_pdf_security.py](file:///home/ikhsan/Documents/Cybermes/tools/test_generate_pdf_security.py)).
- **CI / Build Pipeline**:
  - Added PyYAML zero-dependency fallback in CI skill validation ([2cce573](https://github.com/Zyrexnn/Cybermes/commit/2cce573)).
  - Reconfigured Dependabot targeting `dev` branch with pinned major upgrades ([308985d](https://github.com/Zyrexnn/Cybermes/commit/308985d)).

---

### 📖 Documentation & Sponsorship

- **Infistar AI Gateway Integration**:
  - Added official sponsorship placement, banner, and QR assets ([README.md](file:///home/ikhsan/Documents/Cybermes/README.md), [assets/sponsors/](file:///home/ikhsan/Documents/Cybermes/assets/sponsors/)).
  - Published integration guide and environment templates for 200+ upstream LLM models ([docs/INFISTAR_SETUP.md](file:///home/ikhsan/Documents/Cybermes/docs/INFISTAR_SETUP.md)).
  - Published release announcement on [GitHub Discussions #81](https://github.com/Zyrexnn/Cybermes/discussions/81).
- **Attribution & Code Hygiene**:
  - Overhauled upstream toolchain attribution and verified repository links ([ATTRIBUTION.md](file:///home/ikhsan/Documents/Cybermes/ATTRIBUTION.md)).
  - Cleaned redundant comment slop across Go server implementations.

---

### 📦 Dependency Updates

- Upgraded `github.com/mark3labs/mcp-go` to latest stable release ([3dd9bfa](https://github.com/Zyrexnn/Cybermes/commit/3dd9bfa), [c4f968d](https://github.com/Zyrexnn/Cybermes/commit/c4f968d)).
- Upgraded Python dependencies: `mcp` 1.30.0, `pypdf` 6.21.4, `pyasn1` 0.6.5, `aiohttp` 3.14.4, `pillow` 13.0.0, `virtualenv` 21.14.9, `filelock` 4.2.1, `graphql-core` 3.2.12.
- Upgraded Node dependencies: `hono` 4.13.13, `proxy-addr` 2.0.8, `ip-address` 10.7.2.

---

### 👥 Contributors

Thank you to everyone who contributed code, documentation, and reviews to this release:

- **@Zyrexnn** (Lead Maintainer) — MCP server hardening, scope guards, auto-updater, release orchestration
- **@ikhsannov** — Scope security hardening, client injector refactor, CI pipelines
- **@reyzzzl** — Nmap scanning integration, report finding deduplication engine, report test suites ([#49](https://github.com/Zyrexnn/Cybermes/pull/49))
- **@mgkvibez** (Michael Oogwu) — Documentation updates and release badge tracking ([#75](https://github.com/Zyrexnn/Cybermes/pull/75))
- **@dependabot[bot]** — Automated dependency maintenance

---

### ⚡ Installation & Verification

**Via NPM (Universal Launcher):**
```bash
# Run directly without global installation
npx cybermes-mcp serve

# Or update existing global installation
npm install -g cybermes-mcp@latest
cybermes-mcp update
```

**Verify package version:**
```bash
npm view cybermes-mcp version
# 3.5.1
```

**Full Changelog**: https://github.com/Zyrexnn/Cybermes/compare/v3.5.0...v3.5.1
