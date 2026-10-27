#!/usr/bin/env python3
"""Security regression tests for tools/generate_pdf.py.

Finding text (endpoints, PoC output, server reflections) routinely carries
attacker-influenced markup. report.html is opened in a browser, so any
unsanitized markup becomes stored XSS in a security tool.

Run:  python3 tools/test_generate_pdf_security.py
"""

import json
import re
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from generate_pdf import clean_markdown_for_html, generate_report_for_target


def _real_markup(html_text: str) -> str:
    """Remove escaped spans (&lt;...&gt;) which are inert visible text,
    leaving only markup the browser will actually interpret."""
    return re.sub(r'&lt;.*?&gt;', '', html_text, flags=re.DOTALL).lower()


def _assert_no_executable_markup(rendered: str, context: str):
    real = _real_markup(rendered)
    patterns = [
        r"<script[\s>]",                       # real script tag
        r"\son\w+\s*=\s*[\"']",                # real event-handler attribute
        r"(href|src)\s*=\s*[\"']\s*javascript:",  # javascript: URL
        r"(href|src)\s*=\s*[\"']\s*vbscript:",
        r"(href|src)\s*=\s*[\"']\s*data:text/html",
        r"<(svg|iframe|body|object|embed)[\s>]",
    ]
    for pat in patterns:
        assert re.search(pat, real) is None, \
            f"{context}: executable pattern {pat!r} survived in: {real[:200]}"

XSS_PAYLOADS = [
    '<script>alert(1)</script>',
    '<ScRiPt>alert(1)</ScRiPt>',
    '<img src=x onerror=alert(1)>',
    '<svg onload=alert(1)>',
    '<body onload=alert(1)>',
    '<a href="javascript:alert(1)">click</a>',
    '[click](javascript:alert(1))',
    '[click](JaVaScRiPt:alert(1))',
    '[click](vbscript:msgbox(1))',
    '[x](data:text/html,<script>alert(1)</script>)',
    '![x](x" onerror="alert(1))',
]


def test_converter_neutralizes_xss():
    for payload in XSS_PAYLOADS:
        out = clean_markdown_for_html(f"Evidence: {payload}")
        _assert_no_executable_markup(out, f"payload {payload!r}")
    print("PASS test_converter_neutralizes_xss")


def test_legit_markdown_still_renders():
    src = (
        "# Title\n\n"
        "Some **bold** text and `inline code`.\n\n"
        "| A | B |\n|---|---|\n| 1 | 2 |\n\n"
        "```\nprint('<not a tag>')\n```\n\n"
        "[docs](https://example.com/guide?a=1&b=2)\n"
    )
    out = clean_markdown_for_html(src)
    for expected in ("<strong>bold</strong>", "<table>", "<pre><code>",
                     'href="https://example.com/guide?a=1&amp;b=2"'):
        assert expected in out, f"legit markdown broken, missing {expected!r}: {out[:300]}"
    assert "&lt;not a tag&gt;" in out, "code block content not escaped"
    print("PASS test_legit_markdown_still_renders")


def test_end_to_end_report_has_no_executable_markup():
    with tempfile.TemporaryDirectory() as tmp:
        target = Path(tmp) / "evil_target"
        findings = target / "findings"
        findings.mkdir(parents=True)
        (findings / "high_xss.md").write_text(
            "# Reflected XSS in Search\n\n"
            "- **Severity**: HIGH\n"
            "- **Endpoint**: `/search?q=<script>alert(1)</script>`\n\n"
            "## Description\n\n"
            "Response reflects <img src=x onerror=alert(2)> unsanitized. "
            "See [poc](javascript:alert(3)).\n\n"
            "| Severity | Detail |\n|---|---|\n| HIGH | confirmed |\n",
            encoding="utf-8",
        )
        (target / "metadata.json").write_text(json.dumps({
            "severity_summary": {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 0, "LOW": 0, "INFORMATIONAL": 0},
            "findings": [{"title": "<script>alert(4)</script>", "severity": "HIGH",
                          "cvss": "8.1", "cwe": "CWE-79",
                          "endpoint": "/search?q=<svg onload=alert(5)>"}],
            "pocs": [],
            "scan_time": "2026-01-01 00:00:00",
            "total_findings": 1,
        }), encoding="utf-8")

        html_file, _ = generate_report_for_target(target, output_pdf=False)
        rendered = html_file.read_text(encoding="utf-8")
        # Scan only researcher-controlled regions: everything inside <body>
        # (the <head> shell with its own <body> tag and styles is hardcoded).
        body_inner = rendered.split("<body>", 1)[1]
        _assert_no_executable_markup(body_inner, "report.html")
        assert "&lt;script&gt;" in rendered, "expected escaped payload text in report"
        assert "<table>" in rendered, "finding markdown table lost in report"
    print("PASS test_end_to_end_report_has_no_executable_markup")


if __name__ == "__main__":
    test_converter_neutralizes_xss()
    test_legit_markdown_still_renders()
    test_end_to_end_report_has_no_executable_markup()
    print("ALL SECURITY TESTS PASS")
