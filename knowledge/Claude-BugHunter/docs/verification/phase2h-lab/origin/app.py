"""
Phase 2H — origin app behind Nginx.
Intentional flaws for cache poisoning + HTTP smuggling tests:

1. Reflects X-Forwarded-Host in a Link header used for asset URLs
   → cache poisoning when X-Forwarded-Host is not part of cache key

2. Echoes Transfer-Encoding header presence — used as smuggling marker
   (the actual smuggling happens at the front-end / back-end disagree-on-length level)
"""
from flask import Flask, request, jsonify, make_response

app = Flask(__name__)


@app.route("/")
def home():
    """
    Reflects X-Forwarded-Host into an HTML <link> tag.
    If the front-end cache stores this response keyed only by URL
    (not by X-Forwarded-Host), a single attacker request can poison
    the cache for all subsequent visitors.
    """
    xfh = request.headers.get("X-Forwarded-Host", request.headers.get("Host", "localhost"))
    body = f"""<!doctype html>
<html>
<head>
  <title>Phase 2H origin</title>
  <link rel="stylesheet" href="https://{xfh}/static/style.css">
  <script src="https://{xfh}/static/app.js"></script>
</head>
<body>
  <h1>Welcome — assets load from {xfh}</h1>
  <p>This page is cached by the front-end proxy.</p>
</body>
</html>"""
    resp = make_response(body)
    resp.headers["Cache-Control"] = "public, max-age=300"
    resp.headers["Content-Type"] = "text/html"
    return resp


@app.route("/api/who")
def who():
    """
    Echoes all request headers — useful to see what the back-end actually received.
    Helps confirm HTTP smuggling (does the smuggled request reach back-end?).
    """
    return jsonify(
        method=request.method,
        path=request.path,
        headers={k: v for k, v in request.headers.items()},
        body=request.get_data().decode("utf-8", "replace")[:1000],
    )


@app.route("/api/log", methods=["GET", "POST"])
def log():
    """
    Test endpoint that just records what arrived.
    Hit by smuggled request → confirms smuggling worked.
    """
    return jsonify(method=request.method, path=request.path, headers=dict(request.headers), body=request.get_data().decode("utf-8", "replace")[:500])


@app.route("/static/<path:p>")
def static_files(p):
    return f"/* static asset {p} */", 200, {"Content-Type": "text/css"}


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000, debug=False, threaded=True)
