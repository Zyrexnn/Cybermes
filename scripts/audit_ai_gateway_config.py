#!/usr/bin/env python3
"""
Cybermes AI Gateway Security Configuration Auditor
==================================================
Non-destructive diagnostic utility to audit AI Router and LLM Gateway
endpoints against OWASP GenAI LLM Top 10 (2026) recommendations.

Usage:
  python3 scripts/audit_ai_gateway_config.py --url https://gateway.example.com [--token <TOKEN>] [--json]
"""

import argparse
import json
import ssl
import sys
import urllib.error
import urllib.request
from typing import Any, Dict, List, Optional, Tuple


def make_request(
    url: str,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Tuple[int, Dict[str, str], bytes]:
    """Execute a safe, rate-controlled single HTTP request with timeout."""
    req_headers = {
        "User-Agent": "Cybermes-AIGateway-Auditor/1.0",
        "Accept": "application/json, text/plain, */*",
    }
    if headers:
        req_headers.update(headers)

    ctx = ssl.create_default_context()
    req = urllib.request.Request(url, headers=req_headers, method="GET")

    try:
        with urllib.request.urlopen(req, timeout=timeout, context=ctx) as resp:
            status = resp.status
            resp_headers = {k.lower(): v for k, v in resp.headers.items()}
            body = resp.read()
            return status, resp_headers, body
    except urllib.error.HTTPError as e:
        resp_headers = {k.lower(): v for k, v in e.headers.items()} if e.headers else {}
        body = e.read() if hasattr(e, "read") else b""
        return e.code, resp_headers, body
    except Exception as err:
        return 0, {}, str(err).encode("utf-8")


def audit_ai_gateway(base_url: str, auth_token: Optional[str] = None) -> Dict[str, Any]:
    """Audit gateway endpoints for common security posture weaknesses."""
    base_url = base_url.rstrip("/")
    findings: List[Dict[str, Any]] = []

    # 1. Audit CORS Headers on Root / Control Plane
    root_status, root_headers, _ = make_request(
        base_url + "/",
        headers={"Origin": "https://audit-canary.example.com"}
    )
    cors_origin = root_headers.get("access-control-allow-origin", "")
    cors_creds = root_headers.get("access-control-allow-credentials", "").lower()

    if cors_origin == "*":
        findings.append({
            "check": "CORS_WILDCARD",
            "severity": "LOW",
            "issue": "Access-Control-Allow-Origin is set to wildcard (*)",
            "detail": "Allows unauthenticated client-side scripts from any origin to read public response headers."
        })
    elif "audit-canary.example.com" in cors_origin:
        sev = "HIGH" if cors_creds == "true" else "MEDIUM"
        findings.append({
            "check": "CORS_ARBITRARY_ORIGIN_REFLECTION",
            "severity": sev,
            "issue": "Origin reflection observed in Access-Control-Allow-Origin",
            "detail": f"Server reflected arbitrary origin with Allow-Credentials: {cors_creds}"
        })

    # 2. Audit Model Catalog Endpoint (OWASP GenAI LLM02: Sensitive Information Disclosure)
    models_url = f"{base_url}/api/v1/models"
    status_no_auth, _, body_no_auth = make_request(models_url)

    if status_no_auth == 200:
        model_count = 0
        has_internal_keys = False
        try:
            parsed = json.loads(body_no_auth.decode("utf-8", errors="ignore"))
            data = parsed.get("data", []) if isinstance(parsed, dict) else parsed
            if isinstance(data, list):
                model_count = len(data)
            raw_str = body_no_auth.decode("utf-8", errors="ignore").lower()
            if any(k in raw_str for k in ["multiplier", "internal_url", "upstream_key", "secret"]):
                has_internal_keys = True
        except Exception:
            pass

        findings.append({
            "check": "UNAUTHENTICATED_MODEL_CATALOG_EXPOSURE",
            "severity": "HIGH" if has_internal_keys else "MEDIUM",
            "issue": f"Model catalog exposed without authentication ({model_count} models discovered)",
            "detail": (
                "The endpoint /api/v1/models returned 200 OK to unauthenticated callers. "
                + ("Sensitive upstream metadata or routing multipliers detected." if has_internal_keys else "Publicly lists available AI models.")
            ),
            "endpoint": models_url
        })

    # 3. Audit Auth Token Validation (if token supplied)
    if auth_token:
        status_auth, _, _ = make_request(
            models_url,
            headers={"Authorization": f"Bearer {auth_token}"}
        )
        if status_auth == 200:
            auth_status = "AUTHENTICATED_ACCESS_VERIFIED"
        else:
            auth_status = f"TOKEN_REJECTED_STATUS_{status_auth}"
    else:
        auth_status = "NO_TOKEN_SUPPLIED"

    return {
        "target": base_url,
        "auth_probe_status": auth_status,
        "total_findings": len(findings),
        "findings": findings,
    }


def main():
    parser = argparse.ArgumentParser(description="Cybermes AI Gateway Security Auditor")
    parser.add_argument("--url", required=True, help="Target AI Gateway base URL (e.g. https://gateway.example.com)")
    parser.add_argument("--token", default=None, help="Optional Bearer token to test valid credential propagation")
    parser.add_argument("--json", action="store_true", help="Output results in JSON format")

    args = parser.parse_args()
    results = audit_ai_gateway(args.url, args.token)

    if args.json:
        print(json.dumps(results, indent=2))
        return

    print("=" * 65)
    print("  🛡️  Cybermes AI Gateway Configuration Audit Report")
    print("=" * 65)
    print(f"Target:       {results['target']}")
    print(f"Auth Check:   {results['auth_probe_status']}")
    print(f"Total Issues: {results['total_findings']}\n")

    if not results["findings"]:
        print("✅ No high-risk configuration exposures identified.")
    else:
        for idx, f in enumerate(results["findings"], 1):
            print(f"[{idx}] Severity: {f['severity']}")
            print(f"    Check:    {f['check']}")
            print(f"    Issue:    {f['issue']}")
            print(f"    Details:  {f['detail']}")
            if "endpoint" in f:
                print(f"    Endpoint: {f['endpoint']}")
            print()


if __name__ == "__main__":
    main()
