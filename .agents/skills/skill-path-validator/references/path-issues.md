# Path Validation Reference

## Common Path Issues in Skills

### 1. Absolute Paths in SKILL.md

**Bad:**
```markdown
Use the Python script at `/home/alice/modelcraft/scripts/process.py`
```

**Good:**
```markdown
Use the included Python script at `scripts/process.py`
```

### 2. Missing File References

**Bad:**
```markdown
See [Advanced Guide](references/advanced.md)  # File doesn't exist
```

**Good:**
```markdown
- Basic usage: See [Getting Started](references/getting-started.md)
- Advanced: See [Advanced Guide](references/advanced.md)
```
Then create both files.

### 3. Cross-Directory Dependencies

**Bad** (skill assumes shared code exists in parent):
```markdown
Import utility from `../../shared/utils.py`
```

**Good** (either copy or document):
```markdown
# Option A: Copy shared code into this skill
cp ../shared-skill/scripts/utils.py scripts/

# Option B: Document requirement
This skill requires the `shared-skill` to be installed first.
```

### 4. Unresolved Glob Patterns

**Bad** (pattern in description without clarity):
```markdown
description: Process all `.txt` files in the data directory
```

**Good** (clarify what globs mean):
```markdown
description: >
  Process text files (*.txt) in the input directory.
  Supports patterns like src/**/*.txt for nested directories.
globs:
  - "*.txt"
  - "data/**/*.txt"
```

## Testing Paths

### Quick Check

```bash
# List all SKILL.md files
find .agents/skills -name SKILL.md

# Test a single skill
bash .agents/skills/skill-path-validator/scripts/validate_skill.sh \
  .agents/skills/my-skill

# Test all skills
bash .agents/skills/skill-path-validator/scripts/validate_all_skills.sh \
  .agents/skills
```

### Using ripgrep

```bash
# Find absolute paths
cd .agents/skills/my-skill
rg '(/Users/|/home/|^/\w+/)' SKILL.md

# Find unclosed brackets (broken markdown links)
rg '\[([^\]]*)\](?!\()' SKILL.md

# Find external references (http/https)
rg '\[.*\]\(http' SKILL.md
```

## Best Practices

1. **Keep paths relative**: All file references should be relative to the skill directory
2. **Use consistent naming**: `scripts/`, `references/`, `assets/` are the standard directories
3. **Document external dependencies**: If a skill relies on another skill, state this clearly
4. **Validate before sharing**: Run validation scripts before packaging or distributing a skill
5. **Use globs for patterns**: In SKILL.md frontmatter, use `globs:` array to declare file patterns
