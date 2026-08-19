"""
Phase 2F lab — SSTI + OAuth (redirect_uri laxness) + file upload bypass.

Three endpoints, three skill-area verifications:

1. /render-email   → hunt-ssti  (Jinja2 render_template_string with user input)
2. /oauth/authorize  +  /oauth/callback  → hunt-oauth  (redirect_uri validation flaws + missing state)
3. /upload  → hunt-file-upload  (extension blocklist bypassable 5 ways)
"""

import os
import re
import time
import secrets
import urllib.parse
from flask import Flask, request, jsonify, render_template_string, redirect, abort, send_file

app = Flask(__name__)
UPLOAD_DIR = "/tmp/phase2f-lab/uploads"
os.makedirs(UPLOAD_DIR, exist_ok=True)

# ============================================================
# 1) SSTI — Jinja2 render_template_string on user input
# ============================================================
EMAIL_TEMPLATE_PREFIX = """
<html><body>
<h2>Hello {customer}</h2>
<p>Your order summary:</p>
"""
EMAIL_TEMPLATE_SUFFIX = """
<p>Thanks,<br>Acme Corp</p>
</body></html>
"""


@app.route("/render-email")
def render_email():
    """
    Customer-name templated into Jinja2 via render_template_string.
    Classic SSTI sink. No filter, no autoescape on the substituted bit.
    """
    customer = request.args.get("customer", "Valued Customer")
    # Vulnerable: template string is built from user input, then rendered
    template = EMAIL_TEMPLATE_PREFIX.replace("{customer}", customer) + EMAIL_TEMPLATE_SUFFIX
    return render_template_string(template)


# ============================================================
# 2) OAuth — minimal /authorize + /callback with weak redirect_uri validation
# ============================================================
# Authorized client redirect_uris (the legitimate ones)
REGISTERED_CLIENTS = {
    "acme-spa": {
        "client_secret": "acme-prod-secret-2026",
        # INTENTIONAL FLAW: prefix match instead of exact match
        "redirect_uri_prefix": "https://acme.example/",
    }
}

_authcodes = {}  # code -> (client_id, user_id, redirect_uri, expires)


@app.route("/oauth/authorize")
def oauth_authorize():
    """
    Issue an authorization code.
    Vulnerabilities:
      A) redirect_uri validated by prefix-match, not exact-match
         → https://acme.example/.attacker.com or https://acme.example/@attacker.com bypass
      B) No state parameter requirement (CSRF on the callback)
      C) Code returned in query string
    """
    client_id = request.args.get("client_id")
    redirect_uri = request.args.get("redirect_uri", "")
    state = request.args.get("state", "")  # accepted but not required

    if client_id not in REGISTERED_CLIENTS:
        return jsonify(error="unknown_client"), 400

    # Vulnerable prefix-match
    allowed_prefix = REGISTERED_CLIENTS[client_id]["redirect_uri_prefix"]
    if not redirect_uri.startswith(allowed_prefix):
        return jsonify(error="invalid_redirect_uri", expected_prefix=allowed_prefix), 400

    # Always succeed login for the demo (real provider would have a session)
    code = secrets.token_urlsafe(16)
    _authcodes[code] = {
        "client_id": client_id,
        "user_id": "user-42",
        "redirect_uri": redirect_uri,
        "expires": time.time() + 60,
    }

    # Redirect back to the (attacker-controllable) redirect_uri
    sep = "&" if "?" in redirect_uri else "?"
    return redirect(f"{redirect_uri}{sep}code={code}&state={state}", code=302)


@app.route("/oauth/token", methods=["POST"])
def oauth_token():
    data = request.form
    code = data.get("code")
    client_id = data.get("client_id")
    redirect_uri = data.get("redirect_uri", "")

    rec = _authcodes.get(code)
    if not rec or rec["expires"] < time.time():
        return jsonify(error="invalid_code"), 400
    if rec["client_id"] != client_id:
        return jsonify(error="client_mismatch"), 400
    # Note: no client_secret check — common misconfiguration
    # Real flaw: redirect_uri at /token isn't verified against the one at /authorize either
    return jsonify(
        access_token=secrets.token_urlsafe(24),
        token_type="Bearer",
        expires_in=3600,
        user_id=rec["user_id"],
    )


# ============================================================
# 3) File upload — extension blocklist with bypassable validation
# ============================================================
# Defense (intentionally weak): block exact .php / .phtml extensions
BLOCKED_EXTS = {".php", ".phtml", ".php5"}


def reject(reason):
    return jsonify(error=reason), 400


@app.route("/upload", methods=["POST"])
def upload():
    """
    Vulnerable upload endpoint with multiple bypassable defenses.
    """
    f = request.files.get("file")
    if not f:
        return reject("no_file")
    filename = f.filename or ""

    # INTENTIONAL FLAW 1: case-sensitive blocklist (rejects .php but accepts .PHP, .Php)
    # INTENTIONAL FLAW 2: only checks the LAST extension (accepts shell.php.jpg AND shell.jpg.php)
    # INTENTIONAL FLAW 3: trusts the filename's tail extension; null-byte truncation works on older servers
    # INTENTIONAL FLAW 4: never inspects file content (magic bytes)
    # INTENTIONAL FLAW 5: stores under user-supplied filename (path-traversal candidate)
    _, ext = os.path.splitext(filename)
    if ext in BLOCKED_EXTS:  # case-sensitive — first bypass right here
        return reject("extension_blocked")

    # Save under the user-supplied filename (path-traversal candidate)
    safe_name = filename.replace("..", "")  # naive defense
    dest = os.path.join(UPLOAD_DIR, safe_name)
    f.save(dest)
    return jsonify(ok=True, path=safe_name, url=f"/uploaded/{safe_name}")


@app.route("/uploaded/<path:p>")
def uploaded(p):
    """
    Serve uploaded files. INTENTIONAL: serves any file under UPLOAD_DIR with no MIME inspection.
    """
    full = os.path.join(UPLOAD_DIR, p)
    if not os.path.isfile(full):
        abort(404)
    # Naive Content-Type inference — .jpg gets image/jpeg even if content is HTML
    if p.lower().endswith((".jpg", ".jpeg")):
        return send_file(full, mimetype="image/jpeg")
    if p.lower().endswith(".png"):
        return send_file(full, mimetype="image/png")
    if p.lower().endswith((".html", ".htm")):
        return send_file(full, mimetype="text/html")
    if p.lower().endswith(".svg"):
        return send_file(full, mimetype="image/svg+xml")
    return send_file(full)


@app.route("/")
def index():
    return jsonify(
        endpoints=[
            "GET  /render-email?customer=...",
            "GET  /oauth/authorize?client_id=acme-spa&redirect_uri=...",
            "POST /oauth/token  (form: code, client_id, redirect_uri)",
            "POST /upload (multipart: file)",
            "GET  /uploaded/<path>",
        ],
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=58002, debug=False)
