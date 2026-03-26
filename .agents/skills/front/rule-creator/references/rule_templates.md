# Go Rules Templates

This reference provides templates for creating effective rules in the 5 categories.

## Template Structure

Every rule file should follow this structure:

```markdown
---
paths:
  - "**/*.go"
---

# [Category]: [Rule Name]

[Brief description of what this rule ensures]

## Requirements

- [Specific requirement 1]
- [Specific requirement 2]
- [Specific requirement 3]

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

[Why this rule matters - benefits, risks it mitigates]

---

See skill: `[related-skill]` for comprehensive guidance.
```

## Category 1: code-style

**Purpose**: Enforce consistent code formatting and naming conventions.

**Common rules**:
- Naming conventions (variables, functions, types)
- Code organization (imports, package structure)
- Comment and documentation standards
- Go idioms and patterns
- Error handling style

**Template**:

```markdown
---
paths:
  - "**/*.go"
---

# Code Style: [Specific Style Rule]

Enforce [specific style guideline] for [code element].

## Requirements

- Use [convention] for [code element]
- Follow [pattern] when [condition]
- Avoid [anti-pattern] because [reason]

## Examples

### ✅ Good Example

```go
// Example following the style guide
type UserService struct {
    repo Repository
}

func (s *UserService) CreateUser(ctx context.Context, input *CreateUserInput) (*User, error) {
    if input == nil {
        return nil, bizerrors.New(bizerrors.ParamInvalid, "input cannot be nil")
    }
    // Implementation
}
```

### ❌ Bad Example

```go
// Example violating the style guide
type userService struct {
    Repo Repository
}

func (s *userService) createUser(Input *CreateUserInput) (*User, error) {
    // Missing context, inconsistent naming
}
```

## Rationale

[Why this style matters - readability, maintainability, consistency]

---

See skill: `coding-standards` for comprehensive Go idioms and patterns.
```

## Category 2: testing

**Purpose**: Ensure comprehensive test coverage and TDD practices.

**Common rules**:
- Test-first development (TDD)
- Test coverage requirements
- Test naming conventions
- Mock and fixture patterns
- Test organization

**Template**:

```markdown
---
paths:
  - "**/*.go"
---

# Testing: [Specific Testing Rule]

Ensure [testing practice] for [code category].

## Requirements

- Write tests BEFORE implementation (TDD Red-Green-Refactor)
- Test files must be named `*_test.go` and placed next to implementation
- Cover happy path, edge cases, and error scenarios
- Use testify/assert and testify/mock for assertions and mocking
- Minimum [X]% coverage for [category]

## Examples

### ✅ Good Example (TDD Workflow)

```go
// Step 1: Write test first (Red)
func TestProjectService_CreateProject(t *testing.T) {
    t.Run("should create project with valid input", func(t *testing.T) {
        service := setupTestService()
        input := &CreateProjectInput{
            ID:    "test-project",
            Title: "Test Project",
        }

        result, err := service.CreateProject(context.Background(), input)

        assert.NoError(t, err)
        assert.NotNil(t, result)
        assert.Equal(t, "test-project", result.ID)
    })

    t.Run("should return error for duplicate ID", func(t *testing.T) {
        // Error case testing
    })
}

// Step 2: Implement to pass tests (Green)
// Step 3: Refactor while keeping tests green
```

### ❌ Bad Example

```go
// Implementation without tests
func (s *ProjectService) CreateProject(ctx context.Context, input *CreateProjectInput) (*Project, error) {
    // No tests written first - violates TDD
    return s.repo.Create(ctx, input)
}
```

## Rationale

TDD ensures code correctness, prevents regressions, and enables confident refactoring. Tests serve as documentation and catch bugs early.

---

See skill: `coding-standards` for testing best practices.
```

## Category 3: security

**Purpose**: Prevent security vulnerabilities and enforce secure coding practices.

**Common rules**:
- Input validation and sanitization
- SQL injection prevention
- Authentication and authorization
- Sensitive data handling
- Error message security

**Template**:

```markdown
---
paths:
  - "**/*.go"
---

# Security: [Specific Security Rule]

Prevent [vulnerability type] by [security practice].

## Requirements

- Validate all user input at API boundaries
- Use parameterized queries to prevent SQL injection
- Never log sensitive data (passwords, tokens, keys)
- Handle authentication errors without leaking user existence
- Use crypto/rand for security-sensitive random values

## Examples

### ✅ Good Example

```go
// Secure: Using parameterized query
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Where("email = ?", email).  // Parameterized - safe
        First(&user).Error
    if err != nil {
        return nil, bizerrors.Wrap(err, "query user by email")
    }
    return &user, nil
}

// Secure: No sensitive data in logs
logger.Infof("User login attempt for email: %s", email)  // Don't log password
```

### ❌ Bad Example

```go
// VULNERABLE: SQL injection risk
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)  // Dangerous!
    // Attacker can inject: ' OR '1'='1
}

// VULNERABLE: Logging sensitive data
logger.Infof("Login: email=%s, password=%s", email, password)  // Never log passwords!
```

## Rationale

Security vulnerabilities lead to data breaches, compliance violations, and loss of user trust. OWASP Top 10 vulnerabilities must be actively prevented.

---

See skill: `backend-patterns` for security patterns and authentication guidance.
```

## Category 4: api-design

**Purpose**: Ensure consistent, RESTful, and user-friendly API design.

**Common rules**:
- REST conventions and HTTP methods
- Request/response patterns
- Error response format
- Versioning strategy
- GraphQL conventions

**Template**:

```markdown
---
paths:
  - "**/*.go"
---

# API Design: [Specific API Rule]

Ensure [API pattern] for [endpoint category].

## Requirements

- Follow REST conventions (GET for read, POST for create, PUT/PATCH for update, DELETE for remove)
- Use proper HTTP status codes (200, 201, 400, 401, 403, 404, 500)
- Return consistent error format (BusinessError with code and message)
- Include request IDs for traceability
- Document all endpoints in OpenAPI/GraphQL schema

## Examples

### ✅ Good Example

```go
// RESTful endpoint with proper error handling
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    var input models.CreateProjectInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        respondError(w, bizerrors.New(bizerrors.ParamInvalid, "invalid request body"))
        return
    }

    project, err := h.service.CreateProject(r.Context(), &input)
    if err != nil {
        respondError(w, err)  // Handles business errors with proper status codes
        return
    }

    respondJSON(w, http.StatusCreated, project)  // 201 for successful creation
}

// Consistent error response
type ErrorResponse struct {
    RequestID string `json:"requestId"`
    Code      string `json:"code"`      // e.g., "CONFLICT.PROJECT"
    Message   string `json:"message"`
}
```

### ❌ Bad Example

```go
// Poor API design
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
    // Wrong: No input validation
    // Wrong: Generic error response
    // Wrong: Using 200 instead of 201
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"error": "something went wrong"})
}
```

## Rationale

Consistent API design improves developer experience, enables automatic client generation, and reduces integration errors. RESTful conventions and typed errors make APIs predictable.

---

See skill: `backend-patterns` for comprehensive API design patterns.
```

## Category 5: deployment

**Purpose**: Ensure code is production-ready and follows operational best practices.

**Common rules**:
- Configuration management
- Environment separation
- Database migrations
- Logging and monitoring
- Health checks and graceful shutdown

**Template**:

```markdown
---
paths:
  - "**/*.go"
---

# Deployment: [Specific Deployment Rule]

Ensure [operational practice] for [deployment aspect].

## Requirements

- Load sensitive config from environment variables, not hardcoded
- Support multiple environments (dev, staging, production)
- Implement health check endpoints
- Use structured logging with appropriate levels
- Handle graceful shutdown (SIGTERM)
- Database migrations must be reversible

## Examples

### ✅ Good Example

```go
// Secure configuration
type Config struct {
    DBPassword  string `env:"DB_PASSWORD"`       // From env, not hardcoded
    JWTSecret   string `env:"JWT_SECRET"`
    Environment string `env:"APP_ENV" default:"development"`
}

// Health check endpoint
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now(),
        "checks": map[string]bool{
            "database": h.checkDatabase(),
            "redis": h.checkRedis(),
        },
    }
    respondJSON(w, http.StatusOK, status)
}

// Graceful shutdown
func (s *Server) Shutdown() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := s.httpServer.Shutdown(ctx); err != nil {
        logger.Errorf("Server shutdown error: %v", err)
    }
    s.db.Close()
}
```

### ❌ Bad Example

```go
// DANGEROUS: Hardcoded secrets
const (
    DBPassword = "my-secret-password"  // Never hardcode secrets!
    JWTSecret  = "jwt-secret-key"
)

// Poor: No health check
// Poor: No graceful shutdown
```

## Rationale

Production-ready code requires proper configuration management, observability, and graceful degradation. Secrets in code lead to security breaches. Lack of health checks causes downtime.

---

See skill: `backend-patterns` for deployment and operational patterns.
```

## Best Practices

### Writing Effective Rules

1. **Be specific**: Rules should be concrete and actionable, not vague principles
2. **Show examples**: Good/bad examples are more effective than prose
3. **Explain why**: Include rationale to help Claude understand importance
4. **Reference skills**: Link to comprehensive skills for deeper guidance
5. **Keep focused**: One rule per file, avoid combining unrelated concerns

### Rule Organization

```
.claude/rules/
├── code-style/
│   ├── naming-conventions.md
│   ├── error-handling.md
│   └── imports-organization.md
├── testing/
│   ├── tdd-workflow.md
│   ├── test-coverage.md
│   └── test-patterns.md
├── security/
│   ├── input-validation.md
│   ├── sql-injection.md
│   └── sensitive-data.md
├── api-design/
│   ├── rest-conventions.md
│   ├── error-responses.md
│   └── graphql-patterns.md
└── deployment/
    ├── configuration.md
    ├── health-checks.md
    └── graceful-shutdown.md
```

### Updating Existing Rules

When updating rules:
1. Keep the same file structure and YAML frontmatter
2. Add new requirements incrementally
3. Update examples to reflect current patterns
4. Document changes in commit messages
5. Test rules by running Claude Code on sample code
