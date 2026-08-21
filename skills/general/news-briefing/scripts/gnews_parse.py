#!/usr/bin/env python3
"""Parse a Google News RSS feed into numbered headlines with outlet + date.

Usage: python3 gnews_parse.py /tmp/feed.xml [max_items]
Stdlib only. Handles HTML-escaped titles and <source url="...">Outlet</source>.
"""
import html
import re
import sys


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: gnews_parse.py <feed.xml> [max_items]", file=sys.stderr)
        sys.exit(1)
    path = sys.argv[1]
    limit = int(sys.argv[2]) if len(sys.argv) > 2 else 25

    data = open(path, encoding="utf-8", errors="ignore").read()
    items = re.findall(r"<item>(.*?)</item>", data, re.S)
    print(f"Total items: {len(items)}\n")

    for i, item in enumerate(items[:limit], 1):
        t = re.search(r"<title>(.*?)</title>", item, re.S)
        s = re.search(r"<source[^>]*>(.*?)</source>", item, re.S)
        d = re.search(r"<pubDate>(.*?)</pubDate>", item, re.S)
        title = html.unescape(t.group(1)).strip() if t else "?"
        # Strip trailing " - Outlet" duplicate when <source> already names it
        outlet = html.unescape(s.group(1)).strip() if s else ""
        if outlet:
            title = re.sub(r"\s*-\s*" + re.escape(outlet) + r"\s*$", "", title)
        date = d.group(1).strip() if d else "?"
        print(f"{i}. [{outlet or '?'}] {title}\n   {date}")


if __name__ == "__main__":
    main()
