---
name: dependency-checker
description: Enforce layered architecture dependency rules using golangci-lint depguard. Use when: (1) User requests to check, validate, or verify layer dependencies, (2) User wants to detect or fix architectural violations, (3) User mentions "depguard", "layer dependencies", "dependency rules", or "architecture validation", (4) Before committing code changes that touch multiple layers, (5) Setting up CI/CD dependency checks.
---

# Dependency Checker

Automated enforcement of DDD layered architecture dependency rules using golangci-lint's depguard linter.

## Overview

This skill helps maintain clean architecture by detecting and fixing violations of unidirectional layer dependencies:

- **Domain** → No dependencies (only `pkg/`)
- **Infrastructure** → Domain only
- **Application** → Domain + Infrastructure
- **Interfaces** → Application only (not Infrastructure directly)

## Quick Start

### Check Dependencies

Run dependency check using the helper script:

```bash
python .claude/skills/dependency-checker/scripts/check_dependencies.py
```

Or directly with golangci-lint:

```bash
golangci-lint run --disable-all --enable=depguard ./...
```

### Setup Depguard (First Time)

If depguard is not configured in `.golangci.yml`, run:

```bash
python .claude/skills/dependency-checker/scripts/setup_depguard.py \
    .golangci.yml \
    .claude/skills/dependency-checker/assets/depguard_config.yaml
```

This adds depguard rules to your golangci-lint configuration.

## Workflow

### 1. Detect Violations

Run the dependency checker before committing code:

```bash
# Check all packages
python scripts/check_dependencies.py

# Check specific directory
python scripts/check_dependencies.py internal/domain
```

Output shows violations grouped by layer:

```
❌ Found 3 dependency violation(s):

📂 DOMAIN LAYER:
   internal/domain/modeldesign/model.go:5
      Domain layer must not depend on Infrastructure layer

📂 INFRASTRUCTURE LAYER:
   internal/infrastructure/persistence/repo.go:8
      Infrastructure layer must not depend on Application layer
```

### 2. Understand the Violation

For each violation, identify:

1. **What layer** has the problem
2. **What import** is forbidden
3. **Why** it violates the rule (see references/dependency_rules.md)

### 3. Fix the Violation

Common fixes:

**Domain → Infrastructure violation:**
- Move interface definition to domain layer
- Implement interface in infrastructure
- Use dependency injection

**Infrastructure → App violation:**
- Remove the app layer import
- Infrastructure should only know about domain

**App → Interfaces violation:**
- Use domain types in application layer
- Convert to interface types at the interface layer

**Interfaces → Infrastructure violation:**
- Add application service as intermediary
- Never bypass application layer

See references/dependency_rules.md for detailed examples.

### 4. Verify Fix

Re-run dependency check:

```bash
python scripts/check_dependencies.py
```

Expected output when clean:

```
✅ No dependency violations found!
```

## Integration with CI/CD

Add to your CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Check layer dependencies
  run: golangci-lint run --disable-all --enable=depguard ./...
```

Or using the Taskfile:

```yaml
# Taskfile.yml
check-deps:
  desc: Check layer dependencies
  cmds:
    - python .claude/skills/dependency-checker/scripts/check_dependencies.py
```

## Common Violations

### Violation 1: Domain depending on Infrastructure

```go
// ❌ BAD
package modeldesign
import "modelcraft/internal/infrastructure/persistence"
```

**Fix:** Define interface in domain, implement in infrastructure

### Violation 2: Infrastructure depending on Application

```go
// ❌ BAD
package persistence
import "modelcraft/internal/app/modeldesign"
```

**Fix:** Remove app import, use only domain types

### Violation 3: Application depending on Interfaces

```go
// ❌ BAD
package modeldesign
import "modelcraft/internal/interfaces/graphql/model"
```

**Fix:** Return domain types, convert at interface layer

### Violation 4: Interfaces depending on Infrastructure

```go
// ❌ BAD
package http
import "modelcraft/internal/infrastructure/persistence"
```

**Fix:** Depend on application layer instead

## Resources

### scripts/setup_depguard.py
Configures depguard rules in `.golangci.yml`. Run once to enable enforcement.

### scripts/check_dependencies.py
Runs golangci-lint with depguard and formats output by layer. Use for quick checks.

### assets/depguard_config.yaml
Template configuration for depguard with all layer rules defined.

### references/dependency_rules.md
Comprehensive documentation of layer dependency rules with examples and fixes. Load when you need detailed guidance on specific violations.

## Tips

- Run dependency check before every commit
- Fix violations immediately to prevent technical debt
- Use the check script for readable output grouped by layer
- Reference dependency_rules.md for detailed violation examples
- Configure depguard in CI to prevent violations from being merged

---

See skill: `backend-patterns` for DDD architecture patterns.
See rule: `.claude/rules/code-style/layered-dependency.md` for enforcement guidelines.
