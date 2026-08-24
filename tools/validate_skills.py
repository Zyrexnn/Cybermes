#!/usr/bin/env python3
import sys
from pathlib import Path

if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass

def validate_skills(skills_dir: Path) -> int:
    if not skills_dir.exists():
        print(f"[!] Skills directory not found: {skills_dir}")
        return 1

    total = 0
    valid = 0
    missing_skill_md = []
    missing_desc = []

    for entry in sorted(skills_dir.iterdir()):
        if entry.is_dir() and not entry.name.startswith('.'):
            total += 1
            skill_file = entry / "SKILL.md"
            if not skill_file.exists():
                missing_skill_md.append(entry.name)
                continue

            content = skill_file.read_text(encoding="utf-8", errors="ignore")
            if not content.strip():
                missing_skill_md.append(entry.name)
                continue

            if "name:" in content or "description:" in content or "#" in content:
                valid += 1
            else:
                missing_desc.append(entry.name)

    print("=" * 60)
    print("  Cybermes Skill Integrity Audit")
    print("=" * 60)
    print(f"Total Skill Folders : {total}")
    print(f"Valid & Active      : {valid}")

    if missing_skill_md:
        print(f"\n[WARN] Missing SKILL.md ({len(missing_skill_md)}):")
        for name in missing_skill_md:
            print(f"   - skills/{name}")

    if missing_desc:
        print(f"\n[WARN] Missing description / header ({len(missing_desc)}):")
        for name in missing_desc:
            print(f"   - skills/{name}")

    print("=" * 60)
    return 0

if __name__ == "__main__":
    base_dir = Path(__file__).resolve().parent.parent
    skills_path = base_dir / "skills"
    sys.exit(validate_skills(skills_path))
