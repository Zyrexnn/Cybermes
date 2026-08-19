"""
Phase 2G lab — SAML + MFA-bypass + XXE.

Three endpoints, three skill-area verifications:

1. /saml/acs       — minimal SAML SP with weak validation (hunt-saml)
                     - accepts unsigned responses (signature stripping)
                     - reads NameID textContent (comment injection vulnerable parser)
                     - missing audience restriction check

2. /2fa/verify     — MFA endpoint with NO rate limit (hunt-mfa-bypass)
                     - brute force all 10^6 OTPs

3. /parse-xml      — lxml parser with resolve_entities=True (hunt-xxe)
                     - file read via SYSTEM entity
                     - OOB via external HTTP URL
"""

import secrets
import re
import time
from collections import defaultdict
from flask import Flask, request, jsonify, abort
from lxml import etree

app = Flask(__name__)

# Tracked: bind active sessions to authenticated user
_sessions = {}  # token -> {user, mfa_ok}

# Fixed test secrets — visible to attacker via XXE
TEST_USERS = {
    "admin@phase2g.test": {"role": "admin", "secret": "admin-mfa-secret"},
    "alice@phase2g.test": {"role": "user", "secret": "alice-mfa-secret"},
}

# Fixed OTP per user (in real life, this is generated dynamically)
# We use a fixed secret 6-digit code so the lab is reproducible
TRUE_OTP = "847291"  # the actual 6-digit code admin would receive

# ============================================================
# 1) SAML — minimal SP with three INTENTIONAL flaws
# ============================================================
# Trusted IdP "public key" (for the demo, we trivially check the Signature element
# is *present* — not that it actually verifies. This mimics an SP that calls a
# library but ignores the return value, or uses a parse-but-don't-verify path.)


def parse_saml_response(xml_bytes):
    """
    Vulnerable SAML SP parsing:
      - signature presence checked but NOT cryptographically verified
      - reads NameID via textContent → vulnerable to comment-injection
      - no audience restriction check
      - no IssueInstant / NotOnOrAfter time check
    """
    try:
        # NOTE: resolve_entities=False here — XXE protection on SAML parsing
        # (separate from /parse-xml endpoint which deliberately allows entities)
        parser = etree.XMLParser(resolve_entities=False, no_network=True)
        root = etree.fromstring(xml_bytes, parser)
    except Exception as e:
        return None, f"parse_error: {e}"

    ns = {"saml": "urn:oasis:names:tc:SAML:2.0:assertion",
          "samlp": "urn:oasis:names:tc:SAML:2.0:protocol",
          "ds": "http://www.w3.org/2000/09/xmldsig#"}

    # Naive signature check: just ensure ds:Signature is present
    # Real-world flaw: parsers that don't actually validate, or skip on bad sig
    sig = root.find(".//ds:Signature", ns)
    if sig is None:
        # FLAW 1: signature stripping — server accepts unsigned response anyway
        # (we ACCEPT the unsigned response — only print a warning)
        pass

    # FLAW 2: textContent-based NameID extraction (vulnerable to comment injection)
    # Some parsers truncate at XML comment, some don't — leading to identity confusion
    nameid_elem = root.find(".//saml:Subject/saml:NameID", ns)
    if nameid_elem is None:
        return None, "no_nameid"

    # Pythonic textContent: lxml's .text returns up to first child node.
    # If we use itertext() or full tostring/text, behavior differs across libs.
    # We use .text — which IS truncate-at-comment vulnerable:
    nameid = nameid_elem.text

    # FLAW 3: no audience restriction check, no NotOnOrAfter check
    # Just trust the NameID

    return nameid, None


@app.route("/saml/acs", methods=["POST"])
def saml_acs():
    """SAML SP Assertion Consumer Service."""
    raw = request.form.get("SAMLResponse") or request.get_data()
    if isinstance(raw, str):
        xml_bytes = raw.encode()
    else:
        xml_bytes = raw

    # Some clients base64-encode the SAML response; try both
    import base64
    try:
        decoded = base64.b64decode(raw if isinstance(raw, str) else raw.decode(), validate=True)
        if decoded.startswith(b"<"):
            xml_bytes = decoded
    except Exception:
        pass

    nameid, err = parse_saml_response(xml_bytes)
    if err:
        return jsonify(error=err), 400

    # Authenticate the user identified in NameID
    if nameid in TEST_USERS:
        token = secrets.token_urlsafe(16)
        _sessions[token] = {"user": nameid, "mfa_ok": False, "role": TEST_USERS[nameid]["role"]}
        return jsonify(ok=True, token=token, user=nameid, role=TEST_USERS[nameid]["role"])
    return jsonify(error="unknown user", nameid=nameid), 401


# ============================================================
# 2) MFA brute force — no rate limit on /2fa/verify
# ============================================================
@app.route("/2fa/verify", methods=["POST"])
def mfa_verify():
    """
    Verify a 6-digit OTP. Lab flaw: no rate limit, no lockout.
    Caller submits {token, otp}. Server checks token + OTP digit string.
    """
    data = request.get_json(silent=True) or {}
    token = data.get("token", "")
    otp = data.get("otp", "")
    sess = _sessions.get(token)
    if not sess:
        return jsonify(error="invalid_session"), 401

    # FLAW: no rate limit, no per-IP cap, no incremental delay
    # Real-world: should be N attempts then lockout
    if not re.fullmatch(r"\d{6}", otp):
        return jsonify(error="malformed_otp"), 400

    if otp == TRUE_OTP:
        sess["mfa_ok"] = True
        return jsonify(ok=True, mfa_completed=True)
    return jsonify(ok=False, error="incorrect_otp"), 401


@app.route("/2fa/status")
def mfa_status():
    token = (request.headers.get("Authorization", "") or "").replace("Bearer ", "")
    sess = _sessions.get(token)
    if not sess:
        return jsonify(error="invalid_session"), 401
    return jsonify(mfa_ok=sess.get("mfa_ok"), user=sess["user"], role=sess["role"])


# ============================================================
# 3) XXE — lxml with resolve_entities=True
# ============================================================
@app.route("/parse-xml", methods=["POST"])
def parse_xml():
    """
    Parse arbitrary XML with resolve_entities=True. Vulnerable to XXE.
    Returns the textContent of the root element. File-read and OOB attacks both work.
    """
    raw = request.get_data()
    try:
        # INTENTIONAL FLAW: resolve_entities=True + no_network=False = XXE-ready
        parser = etree.XMLParser(
            resolve_entities=True,
            no_network=False,  # allow HTTP entity URLs (OOB)
            load_dtd=True,
        )
        root = etree.fromstring(raw, parser)
        # Echo the textContent so the entity expansion is visible
        text = "".join(root.itertext()).strip()
        return jsonify(echo=text[:5000], root_tag=root.tag)
    except etree.XMLSyntaxError as e:
        return jsonify(error=f"xml_error: {e}"), 400
    except Exception as e:
        return jsonify(error=f"parse_error: {e}"), 400


@app.route("/")
def index():
    return jsonify(
        endpoints=[
            "POST /saml/acs        (form: SAMLResponse=<xml> or raw XML body)",
            "POST /2fa/verify      (json: {token, otp})",
            "GET  /2fa/status      (Bearer ...)",
            "POST /parse-xml       (raw XML body)",
        ],
        test_users=list(TEST_USERS.keys()),
        note="hunt-saml + hunt-mfa-bypass + hunt-xxe verification",
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=58003, debug=False, threaded=True)
