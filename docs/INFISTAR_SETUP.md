# 🌐 Infistar.cc Integration Guide for Cybermes

**[Infistar.cc](https://www.infistar.cc/register?aff=AI2NALGX&ref_source=link)** is an official AI infrastructure partner for Cybermes. It provides a unified, high-availability API gateway that exposes 200+ frontier large language and multimodal models through a standard OpenAI-compatible protocol.

> **Cybermes × Infistar.cc** | Claude, GPT, Gemini, Grok, DeepSeek & more, one key, a fraction of official prices.

---

## 🎁 Exclusive Partner Perk for Cybermes Users

Sign up through our exclusive link and get **CNY 5 free API credit**:

👉 **[Register on Infistar.cc](https://www.infistar.cc/register?aff=AI2NALGX&ref_source=link)**  
**Referral Code:** `AI2NALGX` *(enter during registration if prompted)*

---

## 🔑 Key Platform Highlights

- 💥 **Up to 99% off**: from 0.1x official pricing.
- ✅ **Official upstreams, 100% genuine models**: no downgrades, no dilution, every model verifiable.
- ⚡ **1,000+ redundant routes with automatic failover**: one route stumbles, the next takes over.
- 🧩 **200+ models from 16 brands**: text, image, and video with one key.
- 🏆 **Recommended by 100+ top open-source projects**, the largest with 100k+ stars.
- 🎯 **Run Cybermes’ multi-model agent workflows with one API key** instead of managing separate accounts.

---

## 🛠️ Step-by-Step Setup

### 1. Register & Generate an API Key
1. Visit **[Infistar Registration](https://www.infistar.cc/register?aff=AI2NALGX&ref_source=link)** (Referral Code: `AI2NALGX`) and complete sign-up.
2. Navigate to your **Dashboard / API Keys** section.
3. Generate a new API key (e.g. `sk-infistar-...`).

### 2. Configure Cybermes Environment (`.env`)
Copy `.env.example` to `.env` if you have not already done so:
```bash
cp .env.example .env
```

Add your Infistar credentials:
```bash
# Infistar.cc Configuration (Referral Code: AI2NALGX)
OPENAI_API_BASE="https://api.infistar.cc/v1"
OPENAI_API_KEY="sk-infistar-your-actual-api-key"
```

### 3. Recommended Models for Offensive Workflows

| Task | Recommended Model | Advantage |
| :--- | :--- | :--- |
| **Deep Reasoning & Vulnerability Analysis** | `deepseek-reasoner` (R1) | High-precision logical deduction and root cause analysis at minimal token cost. |
| **Exploit & PoC Generation** | `claude-3-5-sonnet` | Strict instruction-following and reliable Python exploit scripting. |
| **High-Volume Recon & Triage** | `deepseek-chat` (V3) or `gpt-4o-mini` | Low latency and high throughput for parsing massive endpoint lists. |

---

## 💻 Python Code Example (Standard Client)

If you are running custom automation, security scripts, or verification tools within Cybermes:

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

- **Partner Registration**: [https://www.infistar.cc/register?aff=AI2NALGX&ref_source=link](https://www.infistar.cc/register?aff=AI2NALGX&ref_source=link) *(Code: `AI2NALGX`)*
- **Infistar Platform Workflow Docs**: [https://doc.infistar.ai/en/getting-started/platform-workflow](https://doc.infistar.ai/en/getting-started/platform-workflow)
