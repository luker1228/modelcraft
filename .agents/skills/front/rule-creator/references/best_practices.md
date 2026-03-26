# Rule Creator Best Practices

This reference provides guidance on creating effective rules following Claude Code's memory system best practices.

## Core Principles

### 1. Rules Are Persistent Instructions

Rules are persistent instructions that Claude Code loads into every conversation. They guide Claude's behavior without requiring repeated user input.

**Key characteristics**:
- **Always active**: Rules are always in context, unlike skills that trigger conditionally
- **Persistent guidance**: Rules persist across all conversations
- **File-specific**: Use `paths` to target specific file types or directories
- **Project-specific**: Rules live in `.claude/rules/` and apply to that project only

### 2. Progressive Disclosure

Rules contribute to context window usage. Follow these principles:

- **Keep rules concise**: Each rule should be focused and under 200 lines
- **One rule per file**: Don't combine unrelated concerns
- **Reference skills for depth**: Use `See skill: ...` to link to comprehensive guidance
- **Use examples over prose**: Code examples are more effective than lengthy explanations

### 3. Specificity Matters

**Good rules are actionable**:
```markdown
✅ Use `bizerrors.New()` for business errors, never `errors.New()`
✅ Test files must be named `*_test.go` and placed next to implementation
✅ Log using `logger.Infof()` format strings, not structured fields
```

**Bad rules are vague**:
```markdown
❌ Follow best practices for error handling
❌ Write good tests
❌ Use proper logging
```

## Rule Structure

### Required Frontmatter

Every rule must start with YAML frontmatter:

```yaml
---
paths:
  - "**/*.go"              # Required: file patterns this rule applies to
---
```

**Path patterns**:
- `**/*.go` - All Go files in any directory
- `internal/domain/**/*.go` - Only domain layer Go files
- `*.md` - Only markdown files in root
- `tests/**/*_test.go` - Only test files in tests directory

### Recommended Content Structure

```markdown
---
paths:
  - "**/*.go"
---

# [Category]: [Rule Name]

[One-sentence description of what this rule enforces]

## Requirements

- [Actionable requirement 1]
- [Actionable requirement 2]
- [Actionable requirement 3]

## Examples

### ✅ Good Example

```go
// Code that follows the rule
```

### ❌ Bad Example

```go
// Code that violates the rule
```

## Rationale

[1-2 sentences explaining why this rule matters]

---

See skill: `[related-skill]` for comprehensive guidance.
```

### Content Guidelines

**Title**: Use format `[Category]: [Specific Rule]`
- Code Style: Naming Conventions
- Testing: TDD Workflow
- Security: SQL Injection Prevention
- API Design: REST Conventions
- Deployment: Configuration Management

**Requirements**: 3-7 specific, actionable bullet points
- Use imperative language ("Use...", "Avoid...", "Ensure...")
- Each requirement should be verifiable
- Order by importance (most critical first)

**Examples**: Always include both good and bad examples
- Use ✅ for good examples, ❌ for bad examples
- Show real code, not pseudocode
- Include comments explaining why it's good/bad
- Keep examples short (10-20 lines max)

**Rationale**: Brief explanation (1-3 sentences)
- Why this rule exists
- What problems it prevents
- What benefits it provides

**Skill Reference**: Link to related skills for depth
- Only reference skills that exist in your project
- Use exact skill name (check `.claude/skills/`)
- Place at the end of the rule file

## Category-Specific Guidelines

### Code Style Rules

**Focus on**:
- Naming conventions (variables, functions, types, packages)
- Code organization (imports, package structure, file layout)
- Comments and documentation (godoc, inline comments)
- Go idioms (error handling, defer, interface design)
- Formatting (handled by gofmt, but clarify project-specific choices)

**Example topics**:
- Use `bizerrors` instead of standard `errors` package
- Struct field visibility (public vs private)
- Error message formatting
- Interface naming (avoid -er suffix abuse)

### Testing Rules

**Focus on**:
- TDD workflow (write tests first)
- Test naming and organization
- Coverage requirements
- Test patterns (table-driven, subtests)
- Mocking strategies

**Example topics**:
- Test files must use `*_test.go` suffix
- Minimum 80% coverage for domain layer
- Use testify/assert for assertions
- Test happy path, edge cases, and errors

### Security Rules

**Focus on**:
- Input validation
- SQL injection prevention
- Authentication/authorization
- Sensitive data handling
- Cryptography usage

**Example topics**:
- Validate all user input at boundaries
- Use parameterized queries
- Never log passwords or tokens
- Use crypto/rand for security-sensitive randomness
- Prevent OWASP Top 10 vulnerabilities

### API Design Rules

**Focus on**:
- REST conventions
- HTTP methods and status codes
- Request/response structure
- Error handling and formatting
- Versioning strategy

**Example topics**:
- Use proper HTTP verbs (GET/POST/PUT/DELETE)
- Return 201 for resource creation
- Consistent error response format
- GraphQL resolver patterns

### Deployment Rules

**Focus on**:
- Configuration management
- Environment separation
- Health checks
- Logging and monitoring
- Graceful shutdown

**Example topics**:
- Load secrets from environment variables
- Implement /health endpoint
- Use structured logging
- Handle SIGTERM for graceful shutdown
- Database migrations must be reversible

## Common Pitfalls

### ❌ Don't: Create overly broad rules

```markdown
# Bad: Too vague
Follow Go best practices and write clean code.
```

**Why**: Claude doesn't know what "best practices" or "clean code" means in your context.

### ✅ Do: Create specific, actionable rules

```markdown
# Good: Specific and actionable
Use `bizerrors.New()` for business errors with error codes.
Never use standard `errors.New()` in business logic.
```

### ❌ Don't: Duplicate skill content in rules

```markdown
# Bad: Duplicating skill content
[50 lines explaining DDD architecture...]
```

**Why**: Skills already contain this information. Rules should reference skills, not duplicate them.

### ✅ Do: Reference skills for comprehensive guidance

```markdown
# Good: Reference skill for details
Use domain-driven design patterns for business logic.

See skill: `backend-patterns` for DDD architecture guidance.
```

### ❌ Don't: Create mega-rules covering multiple concerns

```markdown
# Bad: Multiple unrelated rules in one file
1. Error handling guidelines
2. Logging practices
3. Testing requirements
4. Database patterns
```

**Why**: Hard to update, maintain, and target specific files.

### ✅ Do: One focused rule per file

```markdown
# error-handling.md
Focus on error handling patterns only

# logging-practices.md
Focus on logging patterns only
```

## Testing Your Rules

After creating rules, test them:

1. **Load them in Claude Code**: Rules auto-load in conversations
2. **Check they trigger**: Ask Claude to write code that should follow the rule
3. **Verify enforcement**: See if Claude follows the rule without being reminded
4. **Iterate**: Refine based on Claude's actual behavior

**Test scenarios**:
- Ask Claude to write new code in scope of the rule
- Ask Claude to review existing code for rule violations
- Ask Claude to refactor code to follow the rule

## Maintenance

**When to update rules**:
- Project patterns evolve (new conventions adopted)
- Rules are unclear or misinterpreted
- New security vulnerabilities discovered
- Team feedback on rule effectiveness

**Update workflow**:
1. Edit rule file directly in `.claude/rules/`
2. Keep same filename and frontmatter paths
3. Test updated rule with Claude Code
4. Commit changes with clear message
5. Communicate changes to team

## Integration with Skills

**When to use rules vs skills**:

| Use Rules When | Use Skills When |
|----------------|-----------------|
| Enforcing coding standards | Teaching complex workflows |
| Persistent project conventions | Conditional, specialized tasks |
| Always-on guidance | Task-specific knowledge |
| Short, targeted instructions | Multi-step procedures |
| File-specific requirements | Bundling scripts/references |

**Best practice**: Rules should reference skills for depth

```markdown
# In rule file:
Use TDD workflow for all new features.

See skill: `coding-standards` for comprehensive testing patterns.
```

This keeps rules concise while providing access to detailed guidance when needed.
