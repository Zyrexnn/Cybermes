#!/usr/bin/env python3
"""Skill pack frontmatter schema and integrity validator."""

from __future__ import annotations

import sys
from pathlib import Path

try:
    import yaml
except ImportError:
    yaml = None

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass


def extract_frontmatter(content: str) -> tuple[str | None, str | None]:
    lines = content.splitlines()
    if not lines or lines[0].strip() != "---":
        return None, "File does not start with '---' frontmatter fence"

    closing_idx = -1
    for idx in range(1, len(lines)):
        if lines[idx].strip() == "---":
            closing_idx = idx
            break

    if closing_idx == -1:
        return None, "Missing closing '---' frontmatter delimiter"

    return "\n".join(lines[1:closing_idx]), None


def validate_frontmatter_data(frontmatter_text: str) -> tuple[bool, str | None]:
    if yaml is not None:
        try:
            data = yaml.safe_load(frontmatter_text)
        except yaml.YAMLError as exc:
            return False, f"YAML parse error: {exc}"
        except Exception as exc:
            return False, f"Unexpected YAML load error: {exc}"
    else:
        # Fallback key-value parser when PyYAML is not installed in the environment
        data = {}
        for line in frontmatter_text.splitlines():
            trimmed = line.strip()
            if not trimmed or trimmed.startswith("#"):
                continue
            if ":" in trimmed:
                key, val = trimmed.split(":", 1)
                data[key.strip()] = val.strip().strip("'\"")

    if not isinstance(data, dict):
        return False, "YAML frontmatter root must be a mapping"

    name = data.get("name")
    if not isinstance(name, str) or not name.strip():
        return False, "Field 'name' must be a non-empty string"

    description = data.get("description")
    if not isinstance(description, str) or not description.strip():
        return False, "Field 'description' must be a non-empty string"

    return True, None


def validate_skills(skills_dir: Path) -> int:
    if not skills_dir.exists():
        print(f"[!] Skills directory not found: {skills_dir}")
        return 1

    total = 0
    valid = 0
    missing_skill_md: list[str] = []
    invalid_frontmatter: list[tuple[str, str]] = []

    for entry in sorted(skills_dir.iterdir()):
        if entry.is_dir() and not entry.name.startswith("."):
            total += 1
            skill_file = entry / "SKILL.md"
            if not skill_file.exists():
                missing_skill_md.append(entry.name)
                continue

            content = skill_file.read_text(encoding="utf-8", errors="ignore").lstrip("\ufeff")
            if not content.strip():
                missing_skill_md.append(entry.name)
                continue

            frontmatter_text, extract_err = extract_frontmatter(content)
            if extract_err is not None or frontmatter_text is None:
                invalid_frontmatter.append((entry.name, extract_err or "Invalid frontmatter structure"))
                continue

            is_valid, val_err = validate_frontmatter_data(frontmatter_text)
            if not is_valid:
                invalid_frontmatter.append((entry.name, val_err or "Invalid schema"))
                continue

            valid += 1

    print("=" * 60)
    print("  Cybermes Skill Integrity Audit")
    print("=" * 60)
    print(f"Total Skill Folders        : {total}")
    print(f"Valid & Active             : {valid}")
    print(f"Missing SKILL.md           : {len(missing_skill_md)}")
    print(f"Invalid Frontmatter Schema : {len(invalid_frontmatter)}")

    if missing_skill_md:
        print(f"\n[WARN] Missing SKILL.md ({len(missing_skill_md)}):")
        for name in missing_skill_md:
            print(f"   - skills/{name}")

    if invalid_frontmatter:
        print(f"\n[ERROR] Invalid frontmatter schema ({len(invalid_frontmatter)}):")
        for name, err in invalid_frontmatter:
            print(f"   - skills/{name}: {err}")

    print("=" * 60)

    if invalid_frontmatter:
        return 1
    return 0


if __name__ == "__main__":
    base_dir = Path(__file__).resolve().parent.parent
    skills_path = Path(sys.argv[1]) if len(sys.argv) > 1 else base_dir / "skills"
    sys.exit(validate_skills(skills_path))
