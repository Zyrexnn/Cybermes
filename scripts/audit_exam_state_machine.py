#!/usr/bin/env python3
"""
Cybermes Examination Platform State Machine & Schema Auditor
===========================================================
Defensive diagnostic utility to inspect online examination / learning
management API responses or JSON state files for sensitive answer leakage,
un-rendered question exposure, or improper client-side timer controls.

Usage:
  python3 scripts/audit_exam_state_machine.py --file exam_response.json
  python3 scripts/audit_exam_state_machine.py --url https://exam.example.edu/api/v1/session/101 --token <SESSION_TOKEN>
"""

import argparse
import json
import ssl
import sys
import urllib.error
import urllib.request
from typing import Any, Dict, List, Tuple


SENSITIVE_KEY_PATTERNS = [
    "correct_answer",
    "correct_option",
    "correct_index",
    "is_correct",
    "answer_key",
    "kunci_jawaban",
    "pembahasan",
    "explanation",
    "solution",
    "passing_grade",
]


def analyze_data(data: Any, path: str = "") -> List[Dict[str, Any]]:
    """Recursively traverse payload to detect premature answer key disclosure."""
    findings = []
    if isinstance(data, dict):
        for k, v in data.items():
            current_path = f"{path}.{k}" if path else k
            k_lower = k.lower()
            for pattern in SENSITIVE_KEY_PATTERNS:
                if pattern in k_lower:
                    findings.append({
                        "path": current_path,
                        "key": k,
                        "value_preview": str(v)[:40] if v is not None else "null",
                        "severity": "HIGH" if "correct" in k_lower or "kunci" in k_lower else "MEDIUM",
                        "description": f"Sensitive assessment field '{k}' found in client-delivered payload."
                    })
            findings.extend(analyze_data(v, current_path))
    elif isinstance(data, list):
        for i, item in enumerate(data):
            current_path = f"{path}[{i}]"
            findings.extend(analyze_data(item, current_path))
    return findings


def audit_exam_payload(payload: Any) -> Dict[str, Any]:
    """Audit parsed JSON structure for assessment logic flaws."""
    issues = analyze_data(payload)
    
    # Check for timer/state machine fields
    has_server_timestamp = False
    has_client_duration = False
    if isinstance(payload, dict):
        keys_lower = [k.lower() for k in payload.keys()]
        if any(k in keys_lower for k in ["server_time", "server_timestamp", "expires_at", "deadline"]):
            has_server_timestamp = True
        if any(k in keys_lower for k in ["remaining_seconds", "duration", "client_time"]):
            has_client_duration = True

    timer_warning = None
    if has_client_duration and not has_server_timestamp:
        timer_warning = "Client duration field detected without corresponding server-side epoch timestamp or expiration."

    return {
        "total_sensitive_fields": len(issues),
        "sensitive_fields": issues,
        "timer_check": {
            "has_server_timestamp": has_server_timestamp,
            "has_client_duration": has_client_duration,
            "warning": timer_warning
        }
    }


def main():
    parser = argparse.ArgumentParser(description="Cybermes Academic Platform Schema Auditor")
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--file", help="Path to local JSON response file from exam API")
    group.add_argument("--url", help="Live exam session endpoint (e.g. https://exam.example.edu/api/v1/session/123)")
    parser.add_argument("--token", help="Bearer session token for live endpoint access")
    parser.add_argument("--json", action="store_true", help="Output summary in JSON format")

    args = parser.parse_args()

    if args.file:
        try:
            with open(args.file, "r", encoding="utf-8") as f:
                data = json.load(f)
        except Exception as e:
            print(f"❌ Error reading JSON file: {e}", file=sys.stderr)
            sys.exit(1)
    else:
        req = urllib.request.Request(args.url)
        req.add_header("User-Agent", "Cybermes-ExamAuditor/1.0")
        if args.token:
            req.add_header("Authorization", f"Bearer {args.token}")
        ctx = ssl.create_default_context()
        try:
            with urllib.request.urlopen(req, timeout=10, context=ctx) as resp:
                data = json.loads(resp.read().decode("utf-8"))
        except Exception as e:
            print(f"❌ Error requesting endpoint: {e}", file=sys.stderr)
            sys.exit(1)

    result = audit_exam_payload(data)

    if args.json:
        print(json.dumps(result, indent=2))
        return

    print("=" * 65)
    print("  🎓 Cybermes Examination Platform Schema Audit Report")
    print("=" * 65)
    print(f"Premature Key Exposures: {result['total_sensitive_fields']}")
    if result["timer_check"]["warning"]:
        print(f"Timer Warning:           ⚠️  {result['timer_check']['warning']}")
    else:
        print(f"Server Timer Validation:  ✅ Verified server timestamp indicators")
    print()

    if not result["sensitive_fields"]:
        print("✅ No premature answer keys or solution data exposed in payload.")
    else:
        for idx, f in enumerate(result["sensitive_fields"], 1):
            print(f"[{idx}] {f['severity']}: {f['path']}")
            print(f"    Key:     {f['key']}")
            print(f"    Preview: {f['value_preview']}")
            print(f"    Note:    {f['description']}")
            print()


if __name__ == "__main__":
    main()
