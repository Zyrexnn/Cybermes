---
name: js-recon-secret-hunting
description: Downloads client-side JavaScript bundles, searches for leaked API keys, tokens, hardcoded credentials, and hidden internal endpoints.
---

# JavaScript Recon & Secret Hunting Skill

## Purpose
Inspect client-side single page application (SPA) bundles (React, Vue, Next.js, Angular, Webpack) to discover hardcoded API credentials, private cloud buckets, unindexed API routes, and hidden features using `trufflehog` and regex pattern matching.

## Inputs
- JS URL list: `/workspace/recon/<target>/js_files_*.txt`

## Workflow

### 1. Download JS Assets Locally
Download identified JavaScript files for fast local static analysis:
```bash
TARGET_DIR="/workspace/recon/<target>/js_downloads"
mkdir -p "$TARGET_DIR"

if [ -f "/workspace/recon/<target>/js_files.txt" ]; then
    while read -r url; do
        filename=$(echo "$url" | md5sum | cut -d' ' -f1).js
        curl -sSL -m 10 "$url" -o "$TARGET_DIR/$filename" 2>/dev/null || true
    done < "/workspace/recon/<target>/js_files.txt"
fi
```

### 2. Secret & Credential Scanning (TruffleHog)
Scan the downloaded JS directory for high-entropy secrets and verified API keys:
```bash
if command -v trufflehog >/dev/null 2>&1; then
    trufflehog filesystem "$TARGET_DIR" \
               --json \
               --no-verification=false \
               > /workspace/output/secrets/<target>_trufflehog.json 2>/dev/null || true
fi
```

### 3. Route & Endpoint Regex Extraction
Extract hidden API endpoints, GraphQL queries, and Cloud storage URLs from JavaScript source code:
```bash
mkdir -p /workspace/output/secrets/<target>

# Extract internal API paths
grep -rhoEI '["'"'"']/(api|v1|v2|v3|admin|internal|graphql|auth)/[a-zA-Z0-9_\-\/]+["'"'"']' "$TARGET_DIR" \
    | tr -d '"'"'" | sort -u > "/workspace/output/secrets/<target>/hidden_routes.txt" || true

# Extract Cloud Buckets (AWS S3, GCP Storage, Azure Blob)
grep -rhoEI '([a-zA-Z0-9.\-_]+\.s3(\.|\-)[a-zA-Z0-9.\-_]+\.amazonaws\.com|storage\.googleapis\.com/[a-zA-Z0-9.\-_]+|[a-zA-Z0-9.\-_]+\.blob\.core\.windows\.net)' "$TARGET_DIR" \
    | sort -u > "/workspace/output/secrets/<target>/cloud_buckets.txt" || true
```

## Output Artifacts
- `/workspace/output/secrets/<target>_trufflehog.json` - High-confidence leaked API tokens (AWS, Stripe, Firebase, JWT).
- `/workspace/output/secrets/<target>/hidden_routes.txt` - Hidden backend API routes discovered inside JS bundles.
- `/workspace/output/secrets/<target>/cloud_buckets.txt` - S3/GCP/Azure storage bucket identifiers.

## Safety & Responsible Disclosure Guidelines
- If live production credentials (AWS Secret Key, Stripe Live Key, Database password) are found, **halt testing immediately** and document the finding for immediate responsible disclosure.
- Never use discovered credentials to read or exfiltrate customer PII data.
