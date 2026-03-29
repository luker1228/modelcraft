---
name: skill-path-validator
description: >
  Validate and maintain correct paths in local project skills. Use when creating, updating, or auditing skills in .agents/skills/ to ensure all file references (in SKILL.md body, scripts/, references/, assets/) point to valid locations and follow project conventions. Checks for: (1) hardcoded absolute paths that should be relative, (2) missing or incorrect asset/reference file references, (3) broken globs or file patterns, (4) inconsistent directory structure. Triggers on phrases like "validate paths in skill", "fix broken references", "audit skill structure", "path check for skill", or when implementing new skills.
---

# Skill Path Validator

Maintain correct file paths across skill resources. This skill ensures skills in `.agents/skills/` follow consistent path conventions and all references resolve correctly.

## Quick Workflow

1. **Identify the skill directory** to validate (e.g., `/root/modelcraft_project/.agents/skills/my-skill`)
2. **Scan for path issues** using ripgrep (preferred) or fallback to basic tools
3. **Fix discovered problems** (broken references, hardcoded paths, missing files)
4. **Verify completeness** before confirming

## Path Validation Rules

### Acceptable Patterns

- **Relative paths in SKILL.md**: `[See docs](references/path-issues.md)`, `scripts/process.py`
- **Globs in SKILL.md**: `- **API docs**: See [references/](references/)` or patterns like `*.md`, `src/**/*.ts`
- **Absolute paths in scripts/references**: Only when documented as environment-specific (e.g., `/root/modelcraft_project` for dev)
- **Asset file references**: Direct paths like `assets/template.html` are acceptable for bundled files

### Anti-Patterns to Fix

- ❌ Hardcoded absolute paths in SKILL.md body (e.g., `/Users/alice/...` or `/home/bob/...`)
- ❌ Cross-directory assumptions (e.g., `../../scripts/foo.py` when referencing from a skill)
- ❌ References to non-existent files without explicit documentation
- ❌ Unresolved globs or file patterns in SKILL.md descriptions

## Detection & Repair

Use ripgrep (via `rg`) to search for path issues:

```bash
# Find potential absolute paths in SKILL.md files
rg '(/Users/|/home/|/tmp/|^/\w+/)' --glob='SKILL.md' .agents/skills/

# Find likely file references (looking for unclosed brackets or missing files)
rg '\[.*\]\(' --glob='SKILL.md' .agents/skills/ | grep -v '(http'

# Check for hardcoded paths in scripts
rg 'SKILL_DIR|project_root|/root/modelcraft' .agents/skills/*/scripts/ || true
```

If ripgrep is unavailable, install via system package manager:

```bash
# Ubuntu/Debian
sudo apt-get install -y ripgrep

# Alpine (in container)
apk add ripgrep

# macOS
brew install ripgrep
```

## Typical Issues & Fixes

### Issue 1: Broken Reference in SKILL.md

**Problem**: `See [references/path-issues.md](references/path-issues.md)` but file path is wrong or file was moved.

**Fix**:
1. Create `schema.md` in `references/` subdirectory, or
2. Update the reference to point to correct file, or
3. Move file to match reference

### Issue 2: Hardcoded Absolute Path

**Problem**: SKILL.md contains `Use the template at /Users/alice/templates/base.html`

**Fix**: Replace with relative reference:
- Move template to `assets/base.html`
- Update text: "Use the template at `assets/base.html`"

### Issue 3: Cross-Directory Reference

**Problem**: SKILL.md references `../../other-skill/scripts/util.py`

**Fix**: Either:
- Copy `util.py` into this skill's `scripts/`, or
- Add documentation that this skill requires the other skill, or
- Refactor shared logic into a shared references file

## Validation Checklist

When validating a skill:

- [ ] All `.md` file references in SKILL.md exist in the skill directory
- [ ] All paths are relative or properly documented as environment-specific
- [ ] Scripts in `scripts/` are executable and have shebangs where needed
- [ ] Reference files in `references/` have table of contents if >100 lines
- [ ] Asset files in `assets/` are not loaded into context (no references in SKILL.md body)
- [ ] No hardcoded user home dirs (`/Users/`, `/home/`, etc.)
- [ ] Glob patterns in descriptions resolve to actual files (test with `find` or `rg`)
- [ ] No circular dependencies or cross-skill imports

## Usage Examples

### Validate a specific skill

```bash
# Run validation on skill-path-validator itself
cd /root/modelcraft_project/.agents/skills/skill-path-validator
rg '/Users/|/home/' SKILL.md  # Check for absolute paths
ls -la references/ scripts/ assets/  # Verify structure
```

### Find all skills with potential path issues

```bash
cd /root/modelcraft_project/.agents/skills
for skill_dir in */; do
  echo "Checking $skill_dir..."
  rg '/Users/|/home/|^\s*/' "$skill_dir/SKILL.md" | grep -v 'http' || echo "  ✓ Clean"
done
```

### Fix multiple skills in parallel (using agents)

When many skills need validation, use **concurrent subagents**:
- Launch 2-3 parallel agents to scan different skill directories
- Each agent reports issues found
- Apply fixes sequentially after gathering all reports

This avoids redundant scanning and speeds up comprehensive audits.

## When to Use This Skill

- Creating new skills: validate initial structure before release
- Updating existing skills: check references after moving files
- CI/CD integration: periodic audits to catch drift
- Code review: validate path correctness in skill PRs
- Troubleshooting: when Claude can't find bundled resources or getting "file not found" errors in skills
