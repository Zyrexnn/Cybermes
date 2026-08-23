#!/usr/bin/env python3
"""
Cybermes Intelligent Knowledge Search Engine (search_knowledge.py)
Performs fast, token-efficient, context-aware semantic retrieval across local
security knowledge bases (PayloadsAllTheThings, HackTricks, Claude-BugHunter, Strix).

Extracts precise payload blocks and bypass techniques without cluttering LLM context.
"""

import os
import sys
import re
import json
import shutil
import argparse
import subprocess
from pathlib import Path
from typing import List, Dict, Any, Tuple

BASE_DIR = Path(__file__).resolve().parent.parent
KNOWLEDGE_DIR = BASE_DIR / "knowledge"

KB_MAPPING = {
    "payloads": "PayloadsAllTheThings",
    "hacktricks": "hacktricks",
    "claude": "Claude-BugHunter",
    "strix": "strix-skills",
    "hack": "hack-skills",
    "all": None
}

class KnowledgeSearcher:
    def __init__(self, base_dir: Path = KNOWLEDGE_DIR):
        self.base_dir = base_dir
        self.has_ripgrep = shutil.which("rg") is not None

    def search(self, query: str, source: str = "all", limit: int = 4, max_chars: int = 1500) -> List[Dict[str, Any]]:
        """Search knowledge bases and return scored, structured snippets."""
        search_path = self.base_dir
        if source in KB_MAPPING and KB_MAPPING[source]:
            target_sub = self.base_dir / KB_MAPPING[source]
            if target_sub.exists():
                search_path = target_sub

        keywords = [k.strip() for k in query.split() if len(k.strip()) > 1]
        if not keywords:
            return []

        # Find matching candidate files
        candidate_files = self._find_candidate_files(keywords, search_path)
        
        # Parse and score relevant sections from candidates
        results = []
        for file_path, match_count in candidate_files[:25]:
            snippets = self._extract_snippets(file_path, keywords, max_chars)
            for snip in snippets:
                snip["file"] = str(file_path.relative_to(BASE_DIR))
                snip["source_kb"] = self._detect_kb(file_path)
                results.append(snip)

        # Sort by relevance score descending
        results.sort(key=lambda x: x["score"], reverse=True)
        return results[:limit]

    def _detect_kb(self, file_path: Path) -> str:
        try:
            rel = file_path.relative_to(self.base_dir)
            return rel.parts[0] if rel.parts else "knowledge"
        except Exception:
            return "knowledge"

    def _find_candidate_files(self, keywords: List[str], search_path: Path) -> List[Tuple[Path, int]]:
        """Use ripgrep (fast) or python walk to find files containing query terms."""
        file_matches: Dict[Path, int] = {}
        primary_term = keywords[0]

        if self.has_ripgrep:
            try:
                cmd = ["rg", "-i", "-l", "--max-count", "50", "-g", "*.md", "-g", "*.txt", primary_term, str(search_path)]
                proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=10)
                if proc.returncode == 0:
                    for line in proc.stdout.splitlines():
                        p = Path(line.strip())
                        if p.is_file() and p.suffix.lower() in (".md", ".txt") and "site/data" not in p.parts:
                            if p.name.lower() not in ("summary.md", "_sidebar.md", "toc.md"):
                                file_matches[p] = 1
            except Exception:
                pass

        if not file_matches:
            # Fallback pure python search
            for root, _, files in os.walk(search_path):
                if "site/data" in root or "node_modules" in root:
                    continue
                for f in files:
                    if f.lower() in ("summary.md", "_sidebar.md", "toc.md"):
                        continue
                    if f.endswith((".md", ".txt")):
                        p = Path(root) / f
                        try:
                            content = p.read_text(encoding="utf-8", errors="ignore")
                            if any(k.lower() in content.lower() for k in keywords):
                                file_matches[p] = sum(content.lower().count(k.lower()) for k in keywords)
                        except Exception:
                            continue

        # Score files by keyword density
        scored = []
        for p, count in file_matches.items():
            try:
                text = p.read_text(encoding="utf-8", errors="ignore").lower()
                total_hits = sum(text.count(k.lower()) for k in keywords)
                # Boost if multiple query keywords match
                unique_keyword_hits = sum(1 for k in keywords if k.lower() in text)
                score = total_hits + (unique_keyword_hits * 15)
                scored.append((p, score))
            except Exception:
                continue

        scored.sort(key=lambda x: x[1], reverse=True)
        return scored

    def _extract_snippets(self, file_path: Path, keywords: List[str], max_chars: int) -> List[Dict[str, Any]]:
        """Extract markdown sections / code blocks matching the keywords."""
        try:
            content = file_path.read_text(encoding="utf-8", errors="replace")
        except Exception:
            return []

        lines = content.splitlines()
        snippets = []

        # Split content into logical Markdown sections by headings (#, ##, ###)
        sections = []
        curr_heading = "General"
        curr_lines = []
        start_line = 1

        for i, line in enumerate(lines, 1):
            if re.match(r"^#{1,4}\s+", line):
                if curr_lines:
                    sections.append((curr_heading, start_line, "\n".join(curr_lines)))
                curr_heading = line.strip()
                curr_lines = [line]
                start_line = i
            else:
                curr_lines.append(line)
        if curr_lines:
            sections.append((curr_heading, start_line, "\n".join(curr_lines)))

        # Score and filter sections
        for heading, s_line, text in sections:
            lower_text = text.lower()
            lower_heading = heading.lower()

            hit_count = sum(lower_text.count(k.lower()) for k in keywords)
            if hit_count == 0:
                continue

            score = hit_count * 5
            # Header match gets strong boost
            if any(k.lower() in lower_heading for k in keywords):
                score += 40
            # Code block presence boosts utility
            if "```" in text:
                score += 35
            # Payloads / PoC indicators boost score
            if any(term in lower_text for term in ("payload", "bypass", "exploit", "poc", "syntax", "example")):
                score += 25
            # Penalize pure index / TOC summary files
            if file_path.name.lower() in ("summary.md", "_sidebar.md"):
                score -= 60

            # Trim text cleanly to max_chars without cutting code blocks abruptly
            if len(text) > max_chars:
                # Prioritize keeping code blocks if present
                code_blocks = re.findall(r"```[\s\S]*?```", text)
                if code_blocks and len(code_blocks[0]) < max_chars:
                    trimmed_text = f"{heading}\n\n{code_blocks[0]}\n\n*(Truncated for context efficiency)*"
                else:
                    trimmed_text = text[:max_chars].rstrip() + "\n\n*(Truncated for context efficiency)*"
            else:
                trimmed_text = text

            snippets.append({
                "heading": heading,
                "start_line": s_line,
                "score": score,
                "content": trimmed_text
            })

        snippets.sort(key=lambda x: x["score"], reverse=True)
        return snippets[:2]


def format_cli_output(results: List[Dict[str, Any]], query: str) -> None:
    """Format and print results cleanly to terminal."""
    if not results:
        print(f"🔍 [Knowledge Search] No relevant knowledge found for query: '{query}'")
        return

    print(f"\n📚 ══════════════════════════════════════════════════════════════════")
    print(f"   CYBERMES KNOWLEDGE BASE SEARCH: '{query}'")
    print(f"   Found {len(results)} high-signal snippets (Ranked by relevance)")
    print(f"══════════════════════════════════════════════════════════════════════\n")

    for i, res in enumerate(results, 1):
        kb = res.get("source_kb", "knowledge")
        file_path = res.get("file", "")
        heading = res.get("heading", "")
        line_no = res.get("start_line", 1)
        score = res.get("score", 0)

        print(f"─── [Result #{i} | Score: {score}] ──────────────────────────────────────────")
        print(f"📂 KB Source : [{kb}]")
        print(f"📄 Location  : {file_path}:{line_no}")
        print(f"🏷️ Section   : {heading}\n")
        print(res["content"])
        print("\n")

    print(f"💡 Tip: Use '--limit N' or '--source [payloads|hacktricks|claude|strix]' to filter.\n")


def main():
    parser = argparse.ArgumentParser(
        description="Cybermes Local Knowledge Base & Payload Search Engine",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python3 tools/search_knowledge.py "ssti jinja2 filter bypass"
  python3 tools/search_knowledge.py "jwt none algorithm" --source payloads
  python3 tools/search_knowledge.py "idor broken object level authorization" --limit 2
  python3 tools/search_knowledge.py "nosql blind injection" --json
        """
    )
    parser.add_argument("query", nargs="?", help="Search query (keywords, vulnerability type, or bypass technique)")
    parser.add_argument("--source", "-s", default="all", choices=list(KB_MAPPING.keys()), help="Target knowledge base")
    parser.add_argument("--limit", "-n", type=int, default=3, help="Max number of ranked results to return (default: 3)")
    parser.add_argument("--max-len", "-m", type=int, default=1400, help="Max character limit per snippet (default: 1400)")
    parser.add_argument("--json", "-j", action="store_true", help="Output results in structured JSON format")
    args = parser.parse_args()

    if not args.query:
        parser.print_help()
        sys.exit(1)

    searcher = KnowledgeSearcher()
    results = searcher.search(
        query=args.query,
        source=args.source,
        limit=args.limit,
        max_chars=args.max_len
    )

    if args.json:
        print(json.dumps({
            "query": args.query,
            "source": args.source,
            "total_results": len(results),
            "results": results
        }, indent=2))
    else:
        format_cli_output(results, args.query)


if __name__ == "__main__":
    main()
