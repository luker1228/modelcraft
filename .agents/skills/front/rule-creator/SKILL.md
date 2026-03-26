---
name: rule-creator
description: "Create and manage rules for the .claude/rules/ directory. Rules are persistent instructions that guide Claude's behavior when working with Go code. Use this skill when you need to: (1) Create new rules for code-style, testing, security, api-design, or deployment, (2) Understand rule structure and best practices, (3) Generate rule files from templates. This skill helps enforce project-specific coding standards, testing practices, security guidelines, API conventions, and deployment requirements across all conversations."
---

# Rule Creator

Create effective rules for the `.claude/rules/` directory to enforce project-specific conventions for Go code.

## What Are Rules?

Rules are persistent instructions that Claude Code loads into every conversation. They guide Claude's behavior without requiring repeated user input. Rules are:

- **Always active**: Loaded automatically in all conversations
- **File-specific**: Target specific file patterns using `paths`
- **Focused**: One rule per file, covering a single concern
- **Actionable**: Provide specific, verifiable requirements

## Quick Start

### Option 1: Using the Helper Script

Use the `create_rule.py` script to generate a rule file with proper structure:

```bash
# From the skill's scripts directory
python create_rule.py <category> <rule-name>

# Example: Create a naming conventions rule
python create_rule.py code-style naming-conventions

# Example: Create a TDD workflow rule
python create_rule.py testing tdd-workflow
```

The script creates the file at `.claude/rules/<category>/<rule-name>.md` with a proper template.

**Categories**:
- `code-style` - Code formatting, naming, organization (references: `coding-standards`)
- `testing` - TDD, coverage, test patterns (references: `coding-standards`)
- `security` - Input validation, SQL injection, auth (references: `backend-patterns`)
- `api-design` - REST conventions, error responses (references: `backend-patterns`)
- `deployment` - Configuration, health checks, graceful shutdown (references: `backend-patterns`)

### Option 2: Manual Creation

1. Create a file in `.claude/rules/<category>/<rule-name>.md`
2. Use the structure from references/rule_templates.md
3. Fill in requirements, examples, and rationale

## Rule Structure

Every rule must include:

```markdown
---
paths:
  - "**/*.go"              # Required: file patterns
---

# [Category]: [Rule Name]

[One-sentence description]

## Requirements

- [Specific requirement 1]
- [Specific requirement 2]
- [Specific requirement 3]

## Examples

### ✅ Good Example

```go
// Code following the rule
```

### ❌ Bad Example

```go
// Code violating the rule
```

## Rationale

[Why this rule matters - 1-2 sentences]

---

See skill: `[related-skill]` for comprehensive guidance.
```

## Five Rule Categories

### 1. Code Style

**Purpose**: Enforce consistent code formatting and naming conventions

**Common topics**:
- Naming conventions (variables, functions, types, packages)
- Code organization (imports, package structure)
- Error handling patterns
- Comment and documentation standards

**Example rule**: Use `bizerrors` package for all error handling, never standard `errors` package

**Skill reference**: `coding-standards`

### 2. Testing

**Purpose**: Ensure comprehensive test coverage and TDD practices

**Common topics**:
- Test-driven development (write tests first)
- Test naming and file organization
- Coverage requirements (e.g., >80% for domain layer)
- Testing patterns (table-driven, subtests, mocking)

**Example rule**: Write tests BEFORE implementation (TDD Red-Green-Refactor)

**Skill reference**: `coding-standards`

### 3. Security

**Purpose**: Prevent security vulnerabilities and enforce secure coding

**Common topics**:
- Input validation at API boundaries
- SQL injection prevention (parameterized queries)
- Authentication and authorization
- Sensitive data handling (no logging passwords/tokens)
- OWASP Top 10 prevention

**Example rule**: Always use parameterized queries, never string concatenation in SQL

**Skill reference**: `backend-patterns`

### 4. API Design

**Purpose**: Ensure consistent, RESTful, and user-friendly API design

**Common topics**:
- REST conventions (HTTP methods and status codes)
- Request/response structure
- Error response format (typed errors)
- GraphQL resolver patterns
- Versioning strategy

**Example rule**: Return 201 for resource creation, 200 for updates, 204 for deletions

**Skill reference**: `backend-patterns`

### 5. Deployment

**Purpose**: Ensure code is production-ready with operational best practices

**Common topics**:
- Configuration management (env vars, not hardcoded)
- Environment separation (dev/staging/prod)
- Health checks and readiness probes
- Structured logging
- Graceful shutdown (SIGTERM handling)

**Example rule**: Load all secrets from environment variables, never hardcode in code

**Skill reference**: `backend-patterns`

## Writing Effective Rules

### Be Specific and Actionable

✅ **Good**: "Use `bizerrors.New()` for business errors with error codes"

❌ **Bad**: "Follow best practices for error handling"

### Show, Don't Tell

Use code examples instead of verbose explanations:

```go
// ✅ Good Example shows the pattern clearly
logger.Infof("Operation started, modelID: %s", modelID)

// ❌ Bad Example shows what to avoid
logger.Info("Operation started", logfacade.Field("modelID", modelID))  // Wrong!
```

### Reference Skills for Depth

Keep rules concise by referencing comprehensive skills:

```markdown
Use domain-driven design patterns for business logic.

See skill: `backend-patterns` for DDD architecture guidance.
```

### Focus One Concern Per Rule

Don't combine unrelated topics:
- ❌ One file covering "error handling + logging + testing"
- ✅ Separate files: `error-handling.md`, `logging-practices.md`, `tdd-workflow.md`

## Path Patterns

Target specific files using glob patterns in frontmatter:

```yaml
---
paths:
  - "**/*.go"                    # All Go files
  - "internal/domain/**/*.go"    # Only domain layer
  - "**/*_test.go"               # Only test files
  - "internal/interfaces/http/**/*.go"  # Only HTTP interface layer
---
```

## Testing Your Rules

After creating rules:

1. **Load in Claude Code**: Rules auto-load in new conversations
2. **Test enforcement**: Ask Claude to write code that should follow the rule
3. **Verify behavior**: Check if Claude follows the rule without reminders
4. **Iterate**: Refine based on actual behavior

**Test commands**:
- "Write a function to create a new project" (tests code-style, error-handling)
- "Add tests for the ProjectService" (tests TDD workflow)
- "Create an API endpoint for user login" (tests security, api-design)

## Reference Files

For detailed guidance, see:

- **references/rule_templates.md** - Complete templates for all 5 categories with extensive examples
- **references/best_practices.md** - Claude Code rules best practices, common pitfalls, maintenance guidelines

## Common Pitfalls to Avoid

### ❌ Don't: Create vague rules

"Follow Go best practices" - Claude doesn't know what this means in your context

### ✅ Do: Be specific

"Use testify/assert for assertions, testify/mock for mocking dependencies"

### ❌ Don't: Duplicate skill content

Don't copy 50 lines from backend-patterns skill into a rule

### ✅ Do: Reference skills

"See skill: `backend-patterns` for DDD architecture patterns"

### ❌ Don't: Combine multiple concerns

One rule file covering errors, logging, testing, and database patterns

### ✅ Do: One concern per file

Separate files: error-handling.md, logging-practices.md, testing.md

## Maintenance

**When to update rules**:
- New conventions adopted by the team
- Security vulnerabilities discovered
- Rules unclear or misinterpreted
- Team feedback on effectiveness

**Update workflow**:
1. Edit rule file directly in `.claude/rules/<category>/`
2. Keep same filename and frontmatter
3. Test with Claude Code
4. Commit changes with clear message

## Examples

**Create a naming conventions rule**:
```bash
python scripts/create_rule.py code-style naming-conventions
# Edit .claude/rules/code-style/naming-conventions.md
```

**Create a TDD workflow rule**:
```bash
python scripts/create_rule.py testing tdd-workflow
# Edit .claude/rules/testing/tdd-workflow.md
```

**Create a SQL injection prevention rule**:
```bash
python scripts/create_rule.py security sql-injection
# Edit .claude/rules/security/sql-injection.md
```

---

See skill: `coding-standards` for comprehensive Go coding conventions.
See skill: `backend-patterns` for backend architecture patterns and best practices.
