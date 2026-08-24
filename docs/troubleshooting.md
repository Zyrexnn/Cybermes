# 🛠️ Troubleshooting & Diagnostic Guide

This document covers common diagnostic workflows, error patterns, and resolutions when operating Cybermes.

---

## 1. Telegram Bot Does Not Respond

### Symptoms:
Messages and slash commands (`/new`, `/status`) are sent in Telegram (delivered with double checkmarks), but the bot produces no reply.

### Resolutions:
1. **Container running in interactive CLI mode instead of Gateway mode:**
   * Verify that `docker-compose.yml` specifies `command: hermes gateway run`.
   * Check container logs: `docker compose logs --tail 30` and verify the presence of `✓ telegram connected`.
2. **Access Control Whitelist Mismatch:**
   * Inspect `.hermes/.env` or `.env`.
   * When `GATEWAY_ALLOW_ALL_USERS=false`, ensure `TELEGRAM_ALLOWED_USERS` includes your exact numeric Telegram ID.
3. **Session Deadlock or Process Stalling:**
   * Send `/reset` or `/new` in the Telegram chat to initialize a fresh session context.
   * Or restart the container service: `docker compose restart`.

---

## 2. Target Scope & Authorization Refusals

### Symptoms:
The agent refuses a target stating it is not in `scope.yaml` or returns an authorization error.

### Resolutions:
1. **Dynamic Target Override**:
   * `scope.yaml` has `dynamic_target_override: true` and `targets: ["*"]` enabled by default.
   * Any URL or domain requested by the operator is automatically authorized without manual configuration edits.
2. **Context Reset**:
   * Send `/reset` or `/new` in Telegram to clear previous refusal weights in conversation context.
3. **Framing Suggestions**:
   * Consult the [Prompt Engineering & Anti-Filter Guide](prompt_guide.md) for non-triggering diagnostic phrasing.

---

## 3. OmniRoute / Local LLM Connection Refused

### Symptoms:
Container logs output `ConnectionRefusedError: http://localhost:20128/v1`.

### Resolutions:
* Because `docker-compose.yml` runs in `network_mode: host`, `http://localhost:20128/v1` routes directly to the host loopback adapter.
* Ensure OmniRoute, LM Studio, or your local inference backend is active on the configured port.
* When routing to remote OpenRouter endpoints directly, configure `.hermes/.env`:
  ```ini
  OPENROUTER_BASE_URL=https://openrouter.ai/api/v1
  OPENROUTER_API_KEY=sk-or-v1-...
  ```

---

## 4. File Ownership & Permissions on Artifacts (`NoPermissions`)

### Symptoms:
VS Code or host text editor reports `Unable to open file (NoPermissions (FileSystemError))` when opening reports.

### Resolutions:
1. **Set Default POSIX ACLs on host** (one-time command so all new files created by Docker container are world-readable):
   ```bash
   setfacl -R -d -m u::rwx,g::rwx,o::rwx reports recon output logs
   setfacl -R -m u::rwx,g::rwx,o::rwx reports recon output logs
   ```
2. **Reclaim workspace directory ownership manually**:
   ```bash
   sudo chown -R $USER:$USER recon/ output/ reports/ logs/
   ```

---

## 5. High-Performance Go Core Tools (`tools/bin/*`) Missing or Architecture Mismatch

### Symptoms:
Running `smart_pipe`, `search_knowledge`, `secret_scan`, or `aggregate_reports` returns `command not found` or `cannot execute binary file: Exec format error`.

### Resolutions:
1. **Rebuild all Go tools locally**:
   Ensure you have Go installed (`go version`), then recompile directly from the root workspace:
   ```bash
   go build -ldflags="-s -w" -o tools/bin/smart_pipe ./cmd/smart_pipe
   go build -ldflags="-s -w" -o tools/bin/secret_scan ./cmd/secret_scan
   go build -ldflags="-s -w" -o tools/bin/search_knowledge ./cmd/search_knowledge
   go build -ldflags="-s -w" -o tools/bin/aggregate_reports ./cmd/aggregate_reports
   chmod +x tools/bin/*
   ```
2. **Ensure `tools/bin` is in your `$PATH`**:
   Source the environment loader:
   ```bash
   source env.sh
   # Or on Windows PowerShell:
   . .\env.ps1
   ```
3. **Running inside Docker**:
   When building Docker images on ARM64 (e.g. Apple Silicon M-series) or AMD64, Docker will automatically recompile Go binaries to match your CPU architecture natively during `docker compose build`.
