#!/usr/bin/env python3
"""
Helper script to create a new rule file with proper structure.

Usage:
    python create_rule.py <category> <rule-name>

Example:
    python create_rule.py code-style naming-conventions
    python create_rule.py testing tdd-workflow

Categories: code-style, testing, security, api-design, deployment
"""

import os
import sys
from pathlib import Path


TEMPLATE = """---
paths:
  - "**/*.go"
---

# {category_title}: {rule_title}

[TODO: One-sentence description of what this rule enforces]

## Requirements

- [TODO: Actionable requirement 1]
- [TODO: Actionable requirement 2]
- [TODO: Actionable requirement 3]

## Examples

### ✅ Good Example

```go
// TODO: Code that follows the rule
```

### ❌ Bad Example

```go
// TODO: Code that violates the rule
```

## Rationale

[TODO: 1-2 sentences explaining why this rule matters - benefits and risks mitigated]

---

See skill: `{skill_reference}` for comprehensive guidance.
"""

CATEGORIES = {
    "code-style": {
        "title": "Code Style",
        "skill": "coding-standards"
    },
    "testing": {
        "title": "Testing",
        "skill": "coding-standards"
    },
    "security": {
        "title": "Security",
        "skill": "backend-patterns"
    },
    "api-design": {
        "title": "API Design",
        "skill": "backend-patterns"
    },
    "deployment": {
        "title": "Deployment",
        "skill": "backend-patterns"
    }
}


def title_case(text: str) -> str:
    """Convert kebab-case or snake_case to Title Case."""
    return text.replace("-", " ").replace("_", " ").title()


def create_rule(category: str, rule_name: str) -> None:
    """Create a new rule file with proper structure."""
    if category not in CATEGORIES:
        print(f"❌ Invalid category: {category}")
        print(f"   Valid categories: {', '.join(CATEGORIES.keys())}")
        sys.exit(1)

    # Get project root (.claude/rules/)
    # Script is in .claude/skills/rule-creator/scripts/
    # So we go up 3 levels to .claude/, then into rules/
    rules_dir = Path.cwd().parent.parent.parent / "rules" / category

    # Also support being run from project root
    if not rules_dir.parent.parent.exists() or rules_dir.parent.parent.name != ".claude":
        # Try finding .claude/rules from project root
        project_root = Path.cwd()
        while project_root != project_root.parent:
            claude_rules = project_root / ".claude" / "rules" / category
            if (project_root / ".claude").exists():
                rules_dir = claude_rules
                break
            project_root = project_root.parent

    rules_dir.mkdir(parents=True, exist_ok=True)

    # Generate file path
    rule_file = rules_dir / f"{rule_name}.md"

    if rule_file.exists():
        print(f"⚠️  Rule file already exists: {rule_file}")
        response = input("   Overwrite? (y/N): ")
        if response.lower() != "y":
            print("   Cancelled.")
            sys.exit(0)

    # Generate content from template
    category_info = CATEGORIES[category]
    content = TEMPLATE.format(
        category_title=category_info["title"],
        rule_title=title_case(rule_name),
        skill_reference=category_info["skill"]
    )

    # Write file
    rule_file.write_text(content)
    print(f"✅ Created rule: {rule_file}")
    print(f"\nNext steps:")
    print(f"1. Edit {rule_file}")
    print(f"2. Replace [TODO] placeholders with actual content")
    print(f"3. Add specific requirements and examples")
    print(f"4. Test the rule by asking Claude Code to follow it")


def main():
    if len(sys.argv) != 3:
        print("Usage: python create_rule.py <category> <rule-name>")
        print("\nCategories:")
        for cat, info in CATEGORIES.items():
            print(f"  - {cat:<15} (references skill: {info['skill']})")
        print("\nExample:")
        print("  python create_rule.py code-style naming-conventions")
        sys.exit(1)

    category = sys.argv[1]
    rule_name = sys.argv[2]

    create_rule(category, rule_name)


if __name__ == "__main__":
    main()
