"""
Hardened test lab — Flask app with real defenses + intentional FP-shaped behavior.

Used to stress-test the Claude-BugHunter discipline rules:
- OOB-Or-It-Didn't-Happen Gate
- Marker Discipline
- Body-Diff Rule
- Statistical Sampling
- Server-Policy-vs-State Rule
- Pre-Severity Gate
"""

import sqlite3
import time
import random
import secrets
import html
from collections import defaultdict
from flask import Flask, request, jsonify, g, abort
from werkzeug.middleware.proxy_fix import ProxyFix

app = Flask(__name__)
app.wsgi_app = ProxyFix(app.wsgi_app)
app.config["DATABASE"] = "/tmp/hardened-lab/app.sqlite"

# -- Rate limiting (per IP) --
_rate = defaultdict(list)
RATE_WINDOW = 60
RATE_LIMIT_LOGIN = 10


def rate_limit(key, limit=RATE_LIMIT_LOGIN):
    now = time.time()
    _rate[key] = [t for t in _rate[key] if now - t < RATE_WINDOW]
    if len(_rate[key]) >= limit:
        return False
    _rate[key].append(now)
    return True


def db():
    if "db" not in g:
        g.db = sqlite3.connect(app.config["DATABASE"])
        g.db.row_factory = sqlite3.Row
    return g.db


@app.teardown_appcontext
def close_db(error):
    if "db" in g:
        g.db.close()


@app.before_request
def init_db_if_needed():
    if not hasattr(app, "_initialized"):
        with app.app_context():
            conn = db()
            conn.executescript("""
                CREATE TABLE IF NOT EXISTS users (
                  id INTEGER PRIMARY KEY, email TEXT UNIQUE, password TEXT, role TEXT
                );
                CREATE TABLE IF NOT EXISTS sessions (
                  token TEXT PRIMARY KEY, user_id INTEGER, created INTEGER
                );
                CREATE TABLE IF NOT EXISTS products (
                  id INTEGER PRIMARY KEY, name TEXT, description TEXT
                );
            """)
            existing = conn.execute("SELECT COUNT(*) FROM users").fetchone()[0]
            if existing == 0:
                # Seed users
                conn.executemany(
                    "INSERT INTO users (email, password, role) VALUES (?, ?, ?)",
                    [
                        ("admin@lab.test", "admin-secret-2026", "admin"),
                        ("alice@lab.test", "alice-pw-9090", "user"),
                        ("bob@lab.test", "bob-pw-1234", "user"),
                        ("carol@lab.test", "carol-pw-5555", "user"),
                    ],
                )
                # Seed products — note the "JavaScript Tutorial" entry intentionally
                conn.executemany(
                    "INSERT INTO products (name, description) VALUES (?, ?)",
                    [
                        ("Hardened Notebook", "A secure note-taking app."),
                        ("JavaScript Tutorial Pack", "Learn JavaScript safely."),
                        ("Python Cookbook", "Recipes for the Python programmer."),
                        ("Cloud-Native Patterns", "Patterns for modern infrastructure."),
                    ],
                )
                conn.commit()
            app._initialized = True


def get_user_by_session(token):
    if not token:
        return None
    row = db().execute(
        "SELECT u.* FROM users u JOIN sessions s ON s.user_id = u.id "
        "WHERE s.token = ? AND (strftime('%s','now') - s.created) < 3600",
        (token,),
    ).fetchone()
    return dict(row) if row else None


# ============================================================
# /login — defended SQLi (prepared statement) + body-diff user enum
# ============================================================
@app.route("/login", methods=["POST"])
def login():
    """
    Defenses:
      - Prepared statement → SQLi auth-bypass with "admin'--" returns invalid_user
      - Rate limit (10 attempts/min/IP) → brute-force throttled
    But intentional flaw:
      - Response body distinguishes "user not found" vs "wrong password" →
        username enumeration via Body-Diff Rule
    """
    ip = request.headers.get("X-Forwarded-For", request.remote_addr)
    if not rate_limit(f"login:{ip}"):
        return jsonify(ok=False, err="rate_limited"), 429

    data = request.get_json(silent=True) or {}
    email = data.get("email", "")
    password = data.get("password", "")

    # Prepared statement — SQLi-safe
    user = db().execute(
        "SELECT * FROM users WHERE email = ?", (email,)
    ).fetchone()
    if not user:
        # Note: distinct response body → user-enum oracle
        return jsonify(ok=False, err="invalid_user"), 401
    if user["password"] != password:
        # Different body → completes the oracle
        return jsonify(ok=False, err="invalid_password"), 401

    token = secrets.token_urlsafe(32)
    db().execute(
        "INSERT INTO sessions (token, user_id, created) VALUES (?, ?, strftime('%s','now'))",
        (token, user["id"]),
    )
    db().commit()
    return jsonify(ok=True, token=token)


# ============================================================
# /profile/<id> — IDOR-shaped but no actual leak
# ============================================================
@app.route("/profile/<int:user_id>")
def profile(user_id):
    """
    Defenses:
      - Returns 200 for ANY authenticated request (looks like IDOR)
      - BUT body just confirms the request — no other-user data leaked
        ("Looks like IDOR but proves nothing" — Pre-Severity Gate kills the claim)
    """
    auth = request.headers.get("Authorization", "").replace("Bearer ", "")
    me = get_user_by_session(auth)
    if not me:
        abort(401)

    target = db().execute("SELECT id FROM users WHERE id = ?", (user_id,)).fetchone()
    if not target:
        return jsonify(error="not_found"), 404

    # Returns 200 but no other-user data — Pre-Severity Gate Q6 should kill this
    return jsonify(ok=True, viewed_user_id=user_id, note="Profile view recorded.")


# ============================================================
# /fetch?url=... — URL-echo, no actual outbound HTTP request
# ============================================================
@app.route("/fetch")
def fetch():
    """
    Defenses:
      - Looks like SSRF: URL parameter reflected in error message
      - BUT no actual outbound HTTP — the "echo" is client-side error formatting
      - OOB callback will NEVER fire — OOB-Or-It-Didn't-Happen Gate catches the FP
    """
    url = request.args.get("url", "")
    if not url:
        return jsonify(error="url parameter required"), 400
    if not url.startswith(("http://", "https://")):
        return jsonify(error=f"Could not fetch {html.escape(url)} — protocol not allowed"), 400
    # Intentionally NO actual fetch — just echo
    return jsonify(error=f"Could not fetch {html.escape(url)} — destination unreachable"), 502


# ============================================================
# /search?q=... — reflection but server-encoded; FP via word collision
# ============================================================
@app.route("/search")
def search():
    """
    Defenses:
      - Server-side HTML-encoding via html.escape() → <script> becomes &lt;script&gt;
    But intentional FP:
      - Search hits matching product names containing the query
      - "javascript" query matches product name "JavaScript Tutorial Pack" naturally
      - Operator may see "javascript" in response and think reflection, but it's data
      - Marker Discipline (unique payload string) is needed to disprove
    """
    q = request.args.get("q", "").strip()
    if not q:
        return jsonify(results=[], echo="")
    # Server-side encoded reflection
    safe_echo = html.escape(q)
    rows = db().execute(
        "SELECT id, name, description FROM products WHERE name LIKE ? OR description LIKE ?",
        (f"%{q}%", f"%{q}%"),
    ).fetchall()
    return jsonify(
        echo=f"You searched for: {safe_echo}",
        results=[dict(r) for r in rows],
    )


# ============================================================
# /files?ext=... — Server-policy block, NOT a file-existence oracle
# ============================================================
@app.route("/files")
def files():
    """
    Defenses:
      - Extension-based blocklist returns 'blocked' for .config/.asmx/.svc/.ashx
        regardless of whether the file exists
    FP shape:
      - Operator may treat 'blocked' response as "file exists" oracle
      - Reality: it's a server-policy filter (Server-Policy-vs-State Rule)
    """
    ext = request.args.get("ext", "")
    blocked_exts = (".config", ".asmx", ".svc", ".ashx", ".cs", ".vb")
    if any(ext.endswith(b) for b in blocked_exts):
        return jsonify(error="This file type is blocked by the server administrator"), 403
    return jsonify(error="File not found"), 404


# ============================================================
# /admin/users — REAL broken function-level authorization
# ============================================================
@app.route("/admin/users")
def admin_users():
    """
    REAL Critical bug: any authenticated session reads the full user list incl admin.
    Should be: role == 'admin' check
    Is: only authenticated check
    """
    auth = request.headers.get("Authorization", "").replace("Bearer ", "")
    me = get_user_by_session(auth)
    if not me:
        abort(401)
    # Should check me['role'] == 'admin' — but doesn't
    rows = db().execute("SELECT id, email, role FROM users").fetchall()
    return jsonify(users=[dict(r) for r in rows])


# ============================================================
# /admin/timing-enum?user=... — Real timing-based enum, but noisy
# ============================================================
@app.route("/admin/timing-enum")
def timing_enum():
    """
    REAL but noisy timing differential:
      - Valid user: ~50ms ± 20ms
      - Invalid user: ~200ms ± 20ms
    But single probes have ±100ms noise from random.uniform
      → Single test could give either result
      → Statistical Sampling (n=10) needed to disprove/prove
    """
    u = request.args.get("user", "")
    user = db().execute("SELECT id FROM users WHERE email = ?", (u,)).fetchone()
    base_ms = 50 if user else 200
    noise_ms = random.uniform(-100, 100)  # ±100ms noise — single probes misleading
    delay = max(0.001, (base_ms + noise_ms) / 1000)
    time.sleep(delay)
    return jsonify(checked=u, found=False)  # always reports "false" — only timing leaks


@app.route("/")
def index():
    return jsonify(
        endpoints=[
            "/login (POST)",
            "/profile/<id>",
            "/fetch?url=",
            "/search?q=",
            "/files?ext=",
            "/admin/users",
            "/admin/timing-enum?user=",
        ],
        note="Hardened lab — discipline-rule stress test target",
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=58000, debug=False)
