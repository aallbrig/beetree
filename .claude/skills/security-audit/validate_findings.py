#!/usr/bin/env python3
"""Validate a security-audit findings file (JSON array). Exit 1 and print problems on failure."""
import json
import re
import sys

VERDICTS = {"confirmed", "needs_validation", "rejected"}
SEVERITIES = {"low", "medium", "high", "critical"}
ALWAYS = ("id", "unit", "title", "verdict")
PER_VERDICT = {
    "confirmed": ("severity", "trace", "observed", "fix"),
    "needs_validation": ("unresolved_fact",),
    "rejected": ("reason",),
}
ALL_OPTIONAL = {f for fields in PER_VERDICT.values() for f in fields}
LEAKS = (
    (re.compile(r"(?<!\d)\d{12}(?!\d)"), "looks like an AWS account ID"),
    (re.compile(r"/home/[A-Za-z0-9_.-]+"), "absolute home path"),
    (re.compile(r"[A-Za-z0-9._%+-]+@(?!example\.com)[A-Za-z0-9.-]+\.[A-Za-z]{2,}"), "email address"),
)


def check(findings: object) -> list[str]:
    if not isinstance(findings, list):
        return ["top level must be a JSON array"]
    errors: list[str] = []
    seen: set[str] = set()
    for i, f in enumerate(findings):
        name = f"finding {i}"
        if not isinstance(f, dict):
            errors.append(f"{name}: must be an object")
            continue
        name = f"finding {f.get('id', i)}"
        for field in ALWAYS:
            if not isinstance(f.get(field), str) or not f[field].strip():
                errors.append(f"{name}: missing {field}")
        if f.get("id") in seen:
            errors.append(f"{name}: duplicate id")
        seen.add(str(f.get("id")))
        verdict = f.get("verdict")
        if verdict not in VERDICTS:
            errors.append(f"{name}: verdict must be one of {sorted(VERDICTS)}")
            continue
        for field in PER_VERDICT[verdict]:
            if field not in f or not f[field]:
                errors.append(f"{name}: {verdict} requires {field}")
        for field in ALL_OPTIONAL - set(PER_VERDICT[verdict]):
            if field in f:
                errors.append(f"{name}: {field} is not allowed on {verdict}")
        if verdict == "confirmed":
            if f.get("severity") not in SEVERITIES:
                errors.append(f"{name}: severity must be one of {sorted(SEVERITIES)}")
            trace = f.get("trace")
            if not isinstance(trace, list) or not all(
                isinstance(t, str) and re.search(r":\d+$", t) for t in trace
            ):
                errors.append(f"{name}: trace must be a list of file:line strings")
        for field, value in f.items():
            for text in value if isinstance(value, list) else [value]:
                if isinstance(text, str):
                    for rx, why in LEAKS:
                        if rx.search(text):
                            errors.append(f"{name}: {field} {why}")
    return errors


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print("usage: validate_findings.py FILE", file=sys.stderr)
        return 2
    try:
        with open(argv[1], encoding="utf-8") as fh:
            data = json.load(fh)
    except (OSError, json.JSONDecodeError) as exc:
        print(f"cannot read {argv[1]}: {exc}", file=sys.stderr)
        return 1
    errors = check(data)
    for e in errors:
        print(e, file=sys.stderr)
    if not errors:
        print(f"findings ok: {len(data)}")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
