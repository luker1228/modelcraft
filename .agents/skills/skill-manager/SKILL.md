---
name: skill-manager
description: >
  Manage shared skills across CodeBuddy and Claude Code.
  Use this skill when creating, installing, or managing skills that should be
  available in both CodeBuddy and Claude Code. Also use when the user mentions
  "sync skills", "shared skills", "install skill", or asks about the skill
  management mechanism. This skill ensures skills live in a single source of
  truth and are symlinked to both agent tool directories.
---

# Skill Manager

This skill manages the unified skill storage for CodeBuddy and Claude Code,
keeping a single source of truth that both tools reference via symlinks.

## Directory Convention

```
.agents/skills/<skill-name>/SKILL.md    # Source of truth (edit here)
.codebuddy/skills/<skill-name>           # Symlink → ../../.agents/skills/<skill-name>
.claude/skills/<skill-name>              # Symlink → ../../.agents/skills/<skill-name>
```

## Creating a New Skill

When creating a new skill, follow these steps:

### 1. Create the skill in the source directory

```
.agents/skills/<skill-name>/
└── SKILL.md
```

Write the SKILL.md with YAML frontmatter and Markdown instructions.

### 2. Create symlinks for both tools

```bash
ln -s ../../.agents/skills/<skill-name> .codebuddy/skills/<skill-name>
ln -s ../../.agents/skills/<skill-name> .claude/skills/<skill-name>
```

### 3. Verify the links work

```bash
ls .codebuddy/skills/<skill-name>/SKILL.md
ls .claude/skills/<skill-name>/SKILL.md
```

Both should resolve to the same file under `.agents/skills/`.

## Installing an External Skill

When the user wants to install a skill they found elsewhere:

1. Copy the skill directory into `.agents/skills/`
2. Create symlinks in both `.codebuddy/skills/` and `.claude/skills/`
3. Verify access

## Listing Skills

To see all shared skills:

```bash
ls -la .agents/skills/
```

To verify symlinks are intact:

```bash
ls -la .codebuddy/skills/
ls -la .claude/skills/
```

All entries should be symlinks pointing to `../../.agents/skills/`.

## Important Rules

- **Always edit skills in `.agents/skills/`** — that's the source of truth
- **Never edit through a symlink** in `.codebuddy/skills/` or `.claude/skills/`
- **Both tools share the same SKILL.md** — the format is compatible between CodeBuddy and Claude Code
- **Symlink pattern**: always use `../../.agents/skills/<name>` (two levels up from `.codebuddy/skills/` or `.claude/skills/`)

## Troubleshooting

### Broken symlink

If a symlink is broken (skill was deleted from `.agents/skills/`), remove the dead symlink:

```bash
rm .codebuddy/skills/<skill-name>
rm .claude/skills/<skill-name>
```

### Skill only needed by one tool

If a skill is specific to one tool, it can still live in `.agents/skills/` for organizational purposes.
Only create the symlink for the tool that needs it. This keeps everything in one place for discoverability.

### Checking for unsynced skills

To find skills that exist in one tool but not the other:

```bash
diff <(ls .codebuddy/skills/ | sed 's/ //g') <(ls .claude/skills/ | sed 's/ //g')
```
