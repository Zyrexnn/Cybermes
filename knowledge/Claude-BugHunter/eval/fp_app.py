#!/usr/bin/env python3
"""
fp_app.py — false-positive trap app for the discipline benchmark.

Endpoints that LOOK vulnerable to a naive scanner but are SAFE (behavioral
illusions — the app has no real injection/fetch/SQL surface), plus a few
genuinely-vulnerable controls. Ground truth lives in eval/fp_cases.json.

Stdlib only.  Run:  python3 eval/fp_app.py [port]   (default 3002)
"""
import html
import json
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import urlparse, parse_qs


class H(BaseHTTPRequestHandler):
    def log_message(self, *a):
        pass

    def _send(self, code, body, ctype="text/html", extra=None):
        b = body.encode() if isinstance(body, str) else body
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Server", "Apache/2.2.8")  # ambient version-header trap
        for k, v in (extra or {}).items():
            self.send_header(k, v)
        self.send_header("Content-Length", str(len(b)))
        self.end_headers()
        self.wfile.write(b)

    def do_GET(self):
        u = urlparse(self.path)
        p = u.path
        q = parse_qs(u.query)
        g = lambda k: q.get(k, [""])[0]

        # ---------- SAFE TRAPS (look vulnerable, are not) ----------
        if p == "/search":            # reflected but HTML-ENCODED -> not XSS
            return self._send(200, f"<html><body>Results for: <div>{html.escape(g('q'))}</div></body></html>")
        if p == "/api/search":        # reflected in JSON, not executed -> not XSS
            return self._send(200, json.dumps({"query": g("q"), "results": []}), "application/json")
        if p == "/go":                # same-origin-validated redirect -> not open redirect
            url = g("url")
            # Reject backslashes (browsers normalize \ -> /) and control chars — otherwise
            # `/\evil.com` becomes protocol-relative `//evil.com`, an EXTERNAL redirect, which
            # would make this "safe" ground-truth trap actually vulnerable and mis-score agents.
            if (url.startswith("/") and not url.startswith("//")
                    and "\\" not in url and all(ord(c) >= 0x20 for c in url)):
                return self._send(302, "", extra={"Location": url})
            return self._send(400, "External redirects are not allowed")
        if p == "/api/product":       # canned: no injection, no boolean/error diff
            idv = g("id")
            try:
                int(idv)
                return self._send(200, json.dumps({"id": idv, "name": "Widget", "status": "ok"}), "application/json")
            except ValueError:
                return self._send(200, json.dumps({"error": "product not found"}), "application/json")
        if p == "/api/public-config":  # ACAO:* but only public data and NO credentials allowed
            return self._send(200, json.dumps({"theme": "dark", "version": "1.0", "public": True}),
                              "application/json", {"Access-Control-Allow-Origin": "*"})
        if p == "/preview":           # url reflected as a link, NEVER fetched server-side -> not SSRF
            return self._send(200, f"<html><body>Preview: <a href=\"{html.escape(g('url'))}\">open</a></body></html>")
        if p.startswith("/api/user/"):  # public fields only, same shape for any id -> not IDOR
            uid = p.rsplit("/", 1)[-1]
            return self._send(200, json.dumps({"id": uid, "name": f"User{uid}", "avatar": f"/img/{uid}.png"}),
                              "application/json")

        # ---------- REAL VULNS (controls) ----------
        if p == "/echo":              # raw reflection in text/html -> REAL reflected XSS
            return self._send(200, f"<html><body>You said: {g('msg')}</body></html>")
        if p == "/redirect":          # unvalidated Location -> REAL open redirect
            return self._send(302, "", extra={"Location": g("next")})
        if p == "/api/account":       # full private data for ANY id, no auth -> REAL IDOR
            idv = g("id")
            n = int(idv) if idv.isdigit() else 0
            return self._send(200, json.dumps({"id": idv, "email": f"victim{idv}@corp.com",
                                               "ssn": f"123-45-{n:04d}", "balance": 9999}), "application/json")

        if p == "/":
            return self._send(200, "<html><body>fp-trap app: /search /api/search /go /api/product "
                                   "/api/public-config /preview /api/user/&lt;id&gt; /echo /redirect /api/account</body></html>")
        return self._send(404, "not found")


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 3002
    print(f"fp_app on http://127.0.0.1:{port}", flush=True)
    ThreadingHTTPServer(("127.0.0.1", port), H).serve_forever()
