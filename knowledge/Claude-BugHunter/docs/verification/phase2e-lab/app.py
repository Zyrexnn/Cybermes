"""
Phase 2E lab — Flask app with three intentional vulns exercising different skill areas:

1. JWT alg=none + weak HMAC          → hunt-api-misconfig
2. GraphQL introspection + alias abuse + node-IDOR → hunt-graphql
3. Race-condition non-atomic coupon redeem  → hunt-race-condition
"""

import time
import json
import sqlite3
import secrets
import threading
import jwt as pyjwt
from flask import Flask, request, jsonify, g, abort

app = Flask(__name__)
app.config["DATABASE"] = "/tmp/phase2e-lab/app.sqlite"

# ============================================================
# JWT SETUP — intentionally vulnerable HMAC verification
# ============================================================
# Weak HMAC secret (commonly seen in dev configs) — defenders rotate this; some never do
JWT_SECRET = "secret-but-just-long-enough-for-pyjwt-warn-suppress"  # intentionally guessable (prefix 'secret') for brute-test, padded to satisfy PyJWT 2.12+ key-length validation

# Intentional flaw: server accepts alg=none after parsing the JWT
# This mimics the classic `decode(..., algorithms=None)` mistake or `algorithm=token.alg`
# Real PyJWT requires an explicit algorithms list, so we re-implement the broken behavior


def issue_token(claims):
    return pyjwt.encode(claims, JWT_SECRET, algorithm="HS256")


def verify_token_broken(token):
    """
    Vulnerable verifier that mimics the classic alg=none acceptance bug.
    Parses the header, reads `alg`, and if alg='none', returns the payload
    without signature verification. Otherwise uses HS256 with weak secret.
    """
    try:
        header = pyjwt.get_unverified_header(token)
        if header.get("alg", "").lower() == "none":
            # Vulnerable: trust the unsigned token
            return pyjwt.decode(token, options={"verify_signature": False})
        return pyjwt.decode(token, JWT_SECRET, algorithms=["HS256"])
    except Exception:
        return None


# ============================================================
# DB
# ============================================================
def db():
    if "db" not in g:
        g.db = sqlite3.connect(app.config["DATABASE"])
        g.db.row_factory = sqlite3.Row
        g.db.execute("PRAGMA journal_mode = MEMORY")  # speed for race tests
    return g.db


@app.teardown_appcontext
def close_db(error):
    if "db" in g:
        g.db.close()


@app.before_request
def init_db():
    if not hasattr(app, "_init"):
        with app.app_context():
            conn = db()
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS users (
                  id INTEGER PRIMARY KEY, email TEXT UNIQUE, role TEXT
                );
                CREATE TABLE IF NOT EXISTS posts (
                  id INTEGER PRIMARY KEY, owner_id INTEGER, title TEXT, body TEXT
                );
                CREATE TABLE IF NOT EXISTS coupons (
                  code TEXT PRIMARY KEY, balance INTEGER, redeemed_count INTEGER DEFAULT 0
                );
                CREATE TABLE IF NOT EXISTS redemptions (
                  id INTEGER PRIMARY KEY, user_id INTEGER, code TEXT, ts INTEGER
                );
            """)
            existing = conn.execute("SELECT COUNT(*) FROM users").fetchone()[0]
            if existing == 0:
                conn.executemany(
                    "INSERT INTO users (id, email, role) VALUES (?,?,?)",
                    [
                        (1, "admin@phase2e.test", "admin"),
                        (2, "alice@phase2e.test", "user"),
                        (3, "bob@phase2e.test", "user"),
                    ],
                )
                conn.executemany(
                    "INSERT INTO posts (id, owner_id, title, body) VALUES (?,?,?,?)",
                    [
                        (1, 1, "Admin secret notes", "INTERNAL: Q3 financials"),
                        (2, 2, "Alice's blog post", "Public hello"),
                        (3, 3, "Bob's thoughts", "Random idea"),
                    ],
                )
                # Single coupon worth $100 with cap of 1 redemption — race target
                conn.execute(
                    "INSERT INTO coupons (code, balance, redeemed_count) VALUES (?,?,?)",
                    ("PROMO100", 100, 0),
                )
                conn.commit()
            app._init = True


# ============================================================
# Auth helper
# ============================================================
def current_user():
    auth = request.headers.get("Authorization", "")
    if not auth.startswith("Bearer "):
        return None
    token = auth[7:]
    claims = verify_token_broken(token)
    return claims


# ============================================================
# /api/token — issue tokens. Login = trivial (any email)
# ============================================================
@app.route("/api/token", methods=["POST"])
def api_token():
    data = request.get_json(silent=True) or {}
    email = data.get("email", "")
    user = db().execute("SELECT * FROM users WHERE email = ?", (email,)).fetchone()
    if not user:
        return jsonify(error="unknown user"), 401
    token = issue_token({
        "sub": str(user["id"]),  # PyJWT 2.12+ requires string sub
        "email": user["email"],
        "role": user["role"],
        "iat": int(time.time()),
    })
    return jsonify(token=token)


# ============================================================
# /api/me — read claims via the vulnerable verifier
# ============================================================
@app.route("/api/me")
def api_me():
    claims = current_user()
    if not claims:
        return jsonify(error="invalid token"), 401
    return jsonify(claims=claims)


# ============================================================
# /api/admin/secrets — admin-only
# ============================================================
@app.route("/api/admin/secrets")
def api_admin_secrets():
    claims = current_user()
    if not claims:
        return jsonify(error="auth required"), 401
    if claims.get("role") != "admin":
        return jsonify(error="admin only"), 403
    return jsonify(secrets=[
        {"name": "API_KEY", "value": "sk-prod-deadbeef"},
        {"name": "DB_PASSWORD", "value": "prod-pg-pw-2026"},
    ])


# ============================================================
# /graphql — minimal hand-rolled GraphQL endpoint
# ============================================================
# Handles a subset of GraphQL: introspection (__schema), aliases, multi-root selections.
# Intentionally permissive — no depth limits, no cost analysis, no introspection block.

GRAPHQL_SCHEMA = {
    "queryType": {"name": "Query"},
    "types": [
        {
            "name": "Query",
            "kind": "OBJECT",
            "fields": [
                {"name": "post", "type": {"name": "Post"}, "args": [{"name": "id", "type": {"name": "ID"}}]},
                {"name": "me", "type": {"name": "User"}, "args": []},
                {"name": "user", "type": {"name": "User"}, "args": [{"name": "id", "type": {"name": "ID"}}]},
            ],
        },
        {
            "name": "Mutation",
            "kind": "OBJECT",
            "fields": [
                {"name": "redeemCoupon", "type": {"name": "RedeemResult"}, "args": [{"name": "code", "type": {"name": "String"}}]},
            ],
        },
        {"name": "Post", "kind": "OBJECT", "fields": [
            {"name": "id", "type": {"name": "ID"}},
            {"name": "title", "type": {"name": "String"}},
            {"name": "body", "type": {"name": "String"}},
            {"name": "ownerId", "type": {"name": "ID"}},
        ]},
        {"name": "User", "kind": "OBJECT", "fields": [
            {"name": "id", "type": {"name": "ID"}},
            {"name": "email", "type": {"name": "String"}},
            {"name": "role", "type": {"name": "String"}},
        ]},
        {"name": "RedeemResult", "kind": "OBJECT", "fields": [
            {"name": "success", "type": {"name": "Boolean"}},
            {"name": "balance", "type": {"name": "Int"}},
            {"name": "credited", "type": {"name": "Int"}},
        ]},
    ],
}


def gql_resolve(query):
    """Tiny GraphQL parser — handles introspection + aliases + the 3 query fields."""
    q = query.strip()
    out = {}

    # Introspection — return the schema
    if "__schema" in q:
        out["__schema"] = GRAPHQL_SCHEMA
        return {"data": out}

    # Extract selections by regex (production GraphQL would use a real parser)
    import re
    # Match patterns like:  alias: post(id: 1) { id title body }
    selections = re.findall(
        r"(?:(\w+)\s*:\s*)?(\w+)\s*(?:\(([^)]*)\))?\s*\{([^{}]*)\}",
        q,
    )
    for alias, field, args_str, sub in selections:
        key = alias if alias else field
        if field == "post":
            m = re.search(r"id\s*:\s*[\"']?(\d+)", args_str)
            pid = int(m.group(1)) if m else None
            if pid:
                row = db().execute("SELECT * FROM posts WHERE id = ?", (pid,)).fetchone()
                if row:
                    # Intentional flaw: no ownership check — IDOR via GraphQL
                    out[key] = {"id": row["id"], "title": row["title"], "body": row["body"], "ownerId": row["owner_id"]}
                else:
                    out[key] = None
        elif field == "user":
            m = re.search(r"id\s*:\s*[\"']?(\d+)", args_str)
            uid = int(m.group(1)) if m else None
            if uid:
                row = db().execute("SELECT * FROM users WHERE id = ?", (uid,)).fetchone()
                if row:
                    out[key] = {"id": row["id"], "email": row["email"], "role": row["role"]}
                else:
                    out[key] = None
        elif field == "redeemCoupon":
            m = re.search(r"code\s*:\s*[\"']?([A-Z0-9]+)", args_str)
            code = m.group(1) if m else None
            # Reuse the race-vulnerable redeem function (so alias-batching can pile on)
            out[key] = redeem_coupon_unsafe(code)
        elif field == "me":
            claims = current_user()
            if claims:
                out[key] = {"id": claims.get("sub"), "email": claims.get("email"), "role": claims.get("role")}
            else:
                out[key] = None
    return {"data": out}


@app.route("/graphql", methods=["POST", "GET"])
def graphql():
    if request.method == "GET":
        q = request.args.get("query", "")
    else:
        body = request.get_json(silent=True) or {}
        q = body.get("query", "")
    return jsonify(gql_resolve(q))


# ============================================================
# /coupon/redeem — non-atomic check-then-spend, race-vulnerable
# ============================================================
def redeem_coupon_unsafe(code):
    """
    Non-atomic check-then-spend:
      1. SELECT redeemed_count → check < 1
      2. UPDATE redeemed_count = redeemed_count + 1
    Between (1) and (2) another concurrent request can also pass the check.
    """
    if not code:
        return {"success": False, "balance": 0, "credited": 0, "error": "no_code"}
    conn = db()
    row = conn.execute("SELECT * FROM coupons WHERE code = ?", (code,)).fetchone()
    if not row:
        return {"success": False, "balance": 0, "credited": 0, "error": "unknown_code"}
    if row["redeemed_count"] >= 1:
        return {"success": False, "balance": row["balance"], "credited": 0, "error": "already_redeemed"}
    # Race window (the bug): nothing stops two threads from getting past the check above
    # Optional sleep here to widen the race window (lab-tunable; real bugs don't have this)
    time.sleep(0.05)
    conn.execute("UPDATE coupons SET redeemed_count = redeemed_count + 1 WHERE code = ?", (code,))
    conn.execute(
        "INSERT INTO redemptions (user_id, code, ts) VALUES (?, ?, strftime('%s','now'))",
        ((current_user() or {}).get("sub", 0), code),
    )
    conn.commit()
    return {"success": True, "balance": row["balance"], "credited": row["balance"]}


@app.route("/coupon/redeem", methods=["POST"])
def coupon_redeem():
    data = request.get_json(silent=True) or {}
    code = data.get("code", "")
    return jsonify(redeem_coupon_unsafe(code))


@app.route("/coupon/reset")
def coupon_reset():
    """Test helper: reset coupon for repeat runs."""
    db().execute("UPDATE coupons SET redeemed_count = 0 WHERE code = 'PROMO100'")
    db().execute("DELETE FROM redemptions")
    db().commit()
    return jsonify(ok=True)


@app.route("/")
def index():
    return jsonify(
        endpoints=[
            "POST /api/token  {email}",
            "GET /api/me  (Bearer ...)",
            "GET /api/admin/secrets  (Bearer admin token)",
            "POST /graphql  {query}",
            "POST /coupon/redeem  {code}",
            "GET /coupon/reset",
        ],
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=58001, debug=False, threaded=True)
