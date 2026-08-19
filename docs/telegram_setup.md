# 🤖 Telegram Bot Integration & Usage Guide

Cybermes features an autonomous messaging gateway powered by the **Hermes Agent Gateway**, allowing operators to initiate security assessments, execute reconnaissance pipelines, query status, and receive real-time vulnerability alerts directly via Telegram.

---

## 📋 1. Prerequisites & Bot Creation

### A. Create a Telegram Bot via @BotFather
1. Open Telegram and search for **[@BotFather](https://t.me/BotFather)**.
2. Send `/newbot`.
3. Enter a display name (e.g., `Cybermes Security Agent`).
4. Choose a unique bot username ending in `bot` (e.g., `my_cybermes_sec_bot`).
5. Securely save the generated **HTTP API Token** (format: `1234567890:ABCdefGhIJKlmNoPQRsTUVwxyZ`).

### B. Retrieve Your Telegram Numeric User ID (Access Control)
To restrict bot access strictly to authorized operators:
1. Open Telegram and message **[@userinfobot](https://t.me/userinfobot)**.
2. Send `/start`.
3. Copy your numeric `Id` (e.g., `123456789`).

---

## ⚙️ 2. Environment Configuration (`.env` / `.hermes/.env`)

Configure your credentials in `.env` or `.hermes/.env`:

```ini
# ─────────────────────────────────────────────────────────────────────────────
# TELEGRAM BOT INTEGRATION (Hermes Gateway)
# ─────────────────────────────────────────────────────────────────────────────
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGhIJKlmNoPQRsTUVwxyZ
TELEGRAM_ALLOWED_USERS=123456789    # Comma-separated numeric IDs for authorization whitelist
GATEWAY_ALLOW_ALL_USERS=false       # Keep false to restrict access strictly to whitelisted users
HERMES_YOLO_MODE=1                 # Enables autonomous execution of security toolchains
```

---

## 🚀 3. Running the Gateway in Docker

### Start Gateway Container
```bash
docker compose up -d
```

### Inspect Connection Logs
```bash
docker compose logs -f
```

Upon successful connection, the log stream will indicate:
```text
✓ telegram connected
[Telegram] Connected to Telegram (polling mode)
[Telegram] 60 commands registered
```

### Restart Gateway
```bash
docker compose restart
```

---

## 📱 4. Essential Telegram Commands

| Command | Function | Description |
| :--- | :--- | :--- |
| `/new` or `/reset` | **New Session** | Clears conversation context and initializes a fresh engagement memory |
| `/status` | **Agent Status** | Inspects active inference model, token consumption, and uptime |
| `/skills` | **Available Skills** | Lists 200+ offensive security modules and specialized playbooks |
| `/help` | **Command Reference** | Displays full listing of interactive bot commands |
| `/model` | **Model Selector** | Views or switches active LLM provider and inference model |
| `/stop` | **Interrupt Execution** | Halts currently running security scan or tool execution |

---

## 💡 5. Operational Best Practices

1. **Explicit Target Scope Definition**:
   ```text
   Perform passive reconnaissance against http://localhost:8888 in accordance with scope.yaml and enumerate live endpoints.
   ```
2. **Context Recovery on Safety Refusal**:
   * If an LLM safety filter misclassifies authorized security testing, invoke `/reset` to flush the conversational history.
   * Frame instructions explicitly referencing the authorized engagement parameters defined in `scope.yaml`.
