# RLS CEL Policy Expression Design

## Context

ModelCraft currently stores RLS policy expressions as `usingExpr` and `withCheckExpr`. The current editor and backend behavior treat both as JSON predicates, but the intended semantics are different:

- `using` filters existing rows and is compiled into a MySQL `WHERE` condition.
- `check` validates the write input before execution. If validation fails, the insert or update is not executed.

MySQL does not provide PostgreSQL-style native RLS `USING` / `WITH CHECK` policies. ModelCraft must enforce these semantics in the application/runtime layer.

## Decision

Use CEL (`google/cel-go`) as the unified policy expression language for both `using` and `check`.

The expression syntax is shared, but the execution target differs:

- `using`: CEL expression over `row` and `auth`, compiled to MySQL `WHERE` SQL plus parameters.
- `check`: CEL expression over `input` and `auth`, evaluated in memory before the write.

Example `using` expression:

```cel
row.owner_id == auth.user_id && row.status in ["draft", "pending"]
```

Example `check` expression:

```cel
input.owner_id == auth.user_id && input.status in ["draft", "pending"]
```

## Expression Context

`using` expressions may reference:

- `row.<field>`: a field on the existing database row.
- `auth.<name>`: a value from the authenticated user context.

`check` expressions may reference:

- `input.<field>`: a field in the write input.
- `auth.<name>`: a value from the authenticated user context.

References outside the allowed context are invalid. For example, `input.*` is invalid in `using`, and `row.*` is invalid in `check`.

## Supported Language Subset

The first implementation should support a deliberately small CEL subset:

- Boolean operators: `&&`, `||`, `!`
- Comparisons: `==`, `!=`, `>`, `>=`, `<`, `<=`
- Membership: `in`
- Literals: string, number, boolean, null, arrays
- Field selection from `row`, `input`, and `auth`
- Parentheses for grouping

The first implementation should not support:

- Function calls
- Macros such as `all`, `exists`, `map`, or `filter`
- Arithmetic
- Regex
- Cross-model queries
- Subqueries
- Dynamic field names

This keeps the SQL compiler and in-memory evaluator small, auditable, and predictable.

## Execution Model

### Read

For `read`, ModelCraft uses `using`.

```txt
using(row, auth) -> MySQL WHERE
```

Rows that do not match the compiled condition are not returned.

### Create

For `create`, ModelCraft uses `check`.

```txt
check(input, auth) -> boolean
```

If the expression returns `false` or evaluation fails, ModelCraft rejects the request and does not execute the insert.

### Update

For `update`, ModelCraft uses both expressions:

```txt
using(row, auth) -> MySQL WHERE
check(input, auth) -> boolean
```

`using` limits which existing rows can be updated.

`check` evaluates only the submitted patch/input fields. It does not load the old row and does not evaluate the merged post-update row.

If either check fails, ModelCraft rejects the request and does not execute the update.

### Delete

For `delete`, ModelCraft uses `using`.

```txt
using(row, auth) -> MySQL WHERE
```

Rows that do not match the compiled condition cannot be deleted.

## Backend Components

### PolicyExpressionCompiler

Responsible for parsing and validating CEL expressions.

Responsibilities:

- Parse CEL expressions using `cel-go`.
- Check that the expression returns boolean.
- Enforce mode-specific allowed variables.
- Validate referenced model fields against the model schema.
- Validate referenced auth fields against the project auth schema.
- Reject unsupported CEL constructs.

### UsingSQLCompiler

Responsible for converting a validated `using` CEL AST into MySQL SQL.

Example:

```cel
row.owner_id == auth.user_id && row.status in ["draft", "pending"]
```

Compiles to:

```sql
owner_id = ? AND status IN (?, ?)
```

With params:

```json
["u_123", "draft", "pending"]
```

The compiler must use bound parameters for all dynamic values.

### InputCheckEvaluator

Responsible for evaluating a validated `check` expression against write input.

Example:

```cel
input.owner_id == auth.user_id && input.status in ["draft", "pending"]
```

With:

```json
{
  "input": { "owner_id": "u_123", "status": "draft" },
  "auth": { "user_id": "u_123" }
}
```

Returns `true`, so the write may proceed.

## Frontend Changes

The editor should reflect the two semantics:

- Rename `Using Expr` to `Using Filter`.
- Rename `Check Expr` to `Input Check`.
- Use CEL examples instead of JSON examples.

Action-specific fields:

- `read`: show `Using Filter` only.
- `create`: show `Input Check` only.
- `update`: show both.
- `delete`: show `Using Filter` only.

Dry run behavior:

- `Using Filter`: validate and show compiled SQL `WHERE` plus params.
- `Input Check`: validate against sample input and show `true` or `false`.

## Compatibility

The first phase may keep existing storage and GraphQL field names:

- `usingExpr`
- `withCheckExpr`

The content changes from JSON predicate to CEL expression for newly created or edited policies.

For existing JSON policies, the backend should support legacy compatibility by detecting JSON expressions and routing them through the old compiler/evaluator path. A later migration can rename `withCheckExpr` to `inputCheckExpr` and migrate stored expressions if needed.

## Error Handling

Validation errors should include:

- Expression syntax errors.
- Unsupported CEL construct errors.
- Unknown model field errors.
- Unknown auth field errors.
- Invalid context errors, such as `input` used in `using`.
- Non-boolean expression result errors.

Runtime errors should reject the operation and return a clear policy validation error. They must not fall back to allow.

## Testing

Backend tests should cover:

- CEL parse and boolean result validation.
- Variable scoping for `using` and `check`.
- Field and auth reference validation.
- `using` AST to SQL compilation for comparisons, boolean operators, `in`, null, and params.
- `check` evaluation for create input.
- `check` evaluation for update patch input only.
- Deny behavior on false, syntax error, unsupported constructs, and runtime errors.
- Legacy JSON expression compatibility.

Frontend tests should cover:

- Action-specific field visibility.
- CEL placeholder examples.
- Save behavior for visible fields only.
- Dry run display for SQL compilation and input check results.

