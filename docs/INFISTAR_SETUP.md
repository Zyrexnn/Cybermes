# 🌐 Infistar.cc Integration Guide for Cybermes

**[Infistar.cc](https://www.infistar.cc/register?aff=C92NALGX&ref_source=link)** is an official AI infrastructure partner for Cybermes. It provides a unified, high-availability API gateway that exposes 200+ large language and multimodal models through a standard OpenAI-compatible protocol.

This guide outlines how to configure Infistar with Cybermes, autonomous agents (Hermes), and MCP client environments.

---

## 🎁 Cybermes Partner Benefit

Every Cybermes user who creates an account via our official referral link receives **¥5 free API credit** at sign-up:

👉 **[Register on Infistar & Claim Your Free ¥5 Credit](https://www.infistar.cc/register?aff=C92NALGX&ref_source=link)**

---

## 🔑 Why Use Infistar with Cybermes?

1. **Unified Routing**: Access models across 16 providers (Claude 3.5 Sonnet, GPT-4o, DeepSeek-R1/V3, Gemini 2.0, Grok, Qwen) using a single API key.
2. **Cost-Effective**: Pricing starting from 0.1x official rates with genuine, verifiable model routing.
3. **High Availability**: 1,000+ redundant failover routes prevent scan halts when upstream providers experience rate limits.
4. **Standard Protocol**: Drop-in compatible with any tool supporting `openai`-compatible base URLs (`/v1`).

---

## 🛠️ Step-by-Step Setup

### 1. Register & Generate an API Key
1. Visit **[Infistar Registration](https://www.infistar.cc/register?aff=C92NALGX&ref_source=link)** and complete account creation.
2. Navigate to your **Dashboard / API Keys** section.
3. Generate a new API key (e.g. `sk-infistar-...`).

### 2. Configure Cybermes Environment (`.env`)
Copy `.env.example` to `.env` if you have not already done so:
```bash
cp .env.example .env
```

Add your Infistar credentials:
```bash
# Infistar.cc Configuration
OPENAI_API_BASE="https://api.infistar.cc/v1"
OPENAI_API_KEY="sk-infistar-your-actual-api-key"
```

### 3. Recommended Models for Security Workflows

| Task | Recommended Model | Advantage |
| :--- | :--- | :--- |
| **Deep Reasoning & Bug Analysis** | `deepseek-reasoner` (R1) | High-precision logical deduction and root cause analysis at minimal token cost. |
| **Exploit & PoC Generation** | `claude-3-5-sonnet` | Strict instruction-following and reliable Python exploit scripting. |
| **High-Volume Recon Triage** | `deepseek-chat` (V3) or `gpt-4o-mini` | Low latency and high throughput for parsing massive endpoint lists. |

---

## 💻 Python Code Example (Standard Client)

If you are running custom automation or evaluation scripts within Cybermes:

```python
from openai import OpenAI

client = OpenAI(
    base_url="https://api.infistar.cc/v1",
    api_key="your_infistar_api_key_here",
)

response = client.chat.completions.create(
    model="deepseek-reasoner",
    messages=[
        {"role": "system", "content": "You are a cyber security expert."},
        {"role": "user", "content": "Analyze the potential impact of an unauthenticated IDOR vulnerability on an order management endpoint."}
    ]
)

print(response.choices[0].message.content)
```

---

## 📚 Official Resources

- **Partner Registration**: [https://www.infistar.cc/register?aff=C92NALGX&ref_source=link](https://www.infistar.cc/register?aff=C92NALGX&ref_source=link)
- **Infistar Platform Workflow Docs**: [https://doc.infistar.ai/en/getting-started/platform-workflow](https://doc.infistar.ai/en/getting-started/platform-workflow)
