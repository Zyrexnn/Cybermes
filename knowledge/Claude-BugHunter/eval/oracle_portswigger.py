#!/usr/bin/env python3
"""
PortSwigger Web Security Academy lab solved-status oracle.

A running lab instance (https://<32-hex>.web-security-academy.net/) renders a
status widget into its page chrome:

  SOLVED      -> <div class='widgetcontainer-lab-status is-solved'> ... <p>Solved</p>
  NOT SOLVED  -> <div class='widgetcontainer-lab-status is-notsolved'> ... <p>Not solved</p>

So the oracle just GETs the instance root and matches the widget class. (Reference
implementation: canhieu2412/BSCPscan pwnbscp/recon.py _detect_lab_solved.)

  python3 eval/oracle_portswigger.py                       # parser self-test
  python3 eval/oracle_portswigger.py https://XXXX.web-security-academy.net   # live check
"""
import sys
import urllib.request


def ps_solved(instance_url, timeout=12):
    """Return (solved, ok). solved in {True, False, None}; ok=False means the widget
    could not be read (expired instance / wrong URL / network)."""
    url = instance_url.rstrip("/") + "/"
    try:
        req = urllib.request.Request(url, headers={"User-Agent": "claude-bughunter-eval-oracle"})
        html = urllib.request.urlopen(req, timeout=timeout).read().decode("utf-8", "replace")
    except Exception:
        return None, False
    return _classify(html)


def _classify(html):
    # primary: the widget class is authoritative
    if "widgetcontainer-lab-status is-solved" in html:
        return True, True
    if "widgetcontainer-lab-status is-notsolved" in html:
        return False, True
    # secondary: bare class / status text (handles minor markup variation)
    if "is-notsolved" in html or "Not solved" in html:
        return False, True
    if "is-solved" in html or ">Solved<" in html:
        return True, True
    return None, False  # widget not present — probably not a live lab instance


def _selftest():
    SOLVED = "<div class='widgetcontainer-lab-status is-solved'><span>LAB</span><p>Solved</p></div>"
    NOTSOLVED = "<div class='widgetcontainer-lab-status is-notsolved'><span>LAB</span><p>Not solved</p></div>"
    DQ_SOLVED = '<div class="widgetcontainer-lab-status is-solved"><p>Solved</p></div>'
    NEITHER = "<html><body>some random page</body></html>"
    assert _classify(SOLVED) == (True, True), "solved widget"
    assert _classify(NOTSOLVED) == (False, True), "not-solved widget"
    assert _classify(DQ_SOLVED) == (True, True), "double-quoted solved"
    assert _classify(NEITHER) == (None, False), "no widget"
    print("oracle_portswigger parser self-test: PASS (4/4)")


if __name__ == "__main__":
    _selftest()
    if len(sys.argv) > 1:
        print(sys.argv[1], "->", ps_solved(sys.argv[1]))
