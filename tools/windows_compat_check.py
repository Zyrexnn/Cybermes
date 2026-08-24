#!/usr/bin/env python3
"""
Cybermes Windows Compatibility Check (Forwarder to doctor.py)
"""
import sys
from pathlib import Path

if __name__ == "__main__":
    doctor_path = Path(__file__).resolve().parent / "doctor.py"
    with open(doctor_path, "r", encoding="utf-8") as f:
        code = f.read()
    exec(compile(code, str(doctor_path), "exec"))
