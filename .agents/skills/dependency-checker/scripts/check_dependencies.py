#!/usr/bin/env python3
"""
Check layer dependencies using golangci-lint depguard
"""

import subprocess
import sys
import json
import re
from typing import List, Dict, Any
from pathlib import Path


class DependencyViolation:
    def __init__(self, file: str, line: int, message: str):
        self.file = file
        self.line = line
        self.message = message

    def __repr__(self):
        return f"{self.file}:{self.line} - {self.message}"


def run_golangci_lint(working_dir: str = ".") -> tuple[bool, List[DependencyViolation]]:
    """
    Run golangci-lint with depguard and parse violations

    Args:
        working_dir: Directory to run golangci-lint in

    Returns:
        Tuple of (success, violations)
    """
    violations = []

    try:
        # Run golangci-lint with JSON output
        result = subprocess.run(
            ["golangci-lint", "run", "--disable-all", "--enable=depguard", "--out-format=json", "./..."],
            cwd=working_dir,
            capture_output=True,
            text=True
        )

        # Parse JSON output
        if result.stdout:
            try:
                output = json.loads(result.stdout)
                issues = output.get("Issues", [])

                for issue in issues:
                    if issue.get("FromLinter") == "depguard":
                        violations.append(DependencyViolation(
                            file=issue.get("Pos", {}).get("Filename", "unknown"),
                            line=issue.get("Pos", {}).get("Line", 0),
                            message=issue.get("Text", "")
                        ))
            except json.JSONDecodeError:
                # Fallback to text parsing
                pass

        # If no JSON output, try text output
        if not violations and result.stderr:
            violations = parse_text_output(result.stderr)

        success = len(violations) == 0
        return success, violations

    except FileNotFoundError:
        print("❌ Error: golangci-lint not found. Please install it first.", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"❌ Error running golangci-lint: {e}", file=sys.stderr)
        return False, []


def parse_text_output(output: str) -> List[DependencyViolation]:
    """
    Parse text output from golangci-lint

    Args:
        output: Text output from golangci-lint

    Returns:
        List of violations
    """
    violations = []
    # Pattern: file.go:line:col: message (linter)
    pattern = r"(.+?):(\d+):\d+:\s+(.+?)\s+\(depguard\)"

    for match in re.finditer(pattern, output):
        file_path, line_num, message = match.groups()
        violations.append(DependencyViolation(
            file=file_path,
            line=int(line_num),
            message=message
        ))

    return violations


def print_violations(violations: List[DependencyViolation]) -> None:
    """
    Print violations in a readable format

    Args:
        violations: List of dependency violations
    """
    if not violations:
        print("✅ No dependency violations found!")
        return

    print(f"❌ Found {len(violations)} dependency violation(s):\n")

    # Group by layer
    by_layer = {}
    for v in violations:
        layer = extract_layer(v.file)
        if layer not in by_layer:
            by_layer[layer] = []
        by_layer[layer].append(v)

    for layer, layer_violations in by_layer.items():
        print(f"📂 {layer.upper()} LAYER:")
        for v in layer_violations:
            print(f"   {v.file}:{v.line}")
            print(f"      {v.message}")
        print()


def extract_layer(file_path: str) -> str:
    """
    Extract layer name from file path

    Args:
        file_path: Path to Go file

    Returns:
        Layer name (domain, infrastructure, app, interfaces)
    """
    if "/domain/" in file_path:
        return "domain"
    elif "/infrastructure/" in file_path:
        return "infrastructure"
    elif "/app/" in file_path:
        return "app"
    elif "/interfaces/" in file_path:
        return "interfaces"
    else:
        return "unknown"


def main():
    working_dir = sys.argv[1] if len(sys.argv) > 1 else "."

    print("🔍 Checking layer dependencies with depguard...\n")

    success, violations = run_golangci_lint(working_dir)

    print_violations(violations)

    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
