# Code Review Checklist: model_enums SQL Query Bug

## Bug Pattern

**Category**: SQL Query Scope / Multi-tenant Safety

**Rule**: All SQL queries on the `model_enums` table MUST include both `org_name` AND `project_slug` as filter conditions. Using only `project_slug` is insufficient.

## Why

`project_slug` alone does not uniquely identify a project across the entire system. The same `project_slug` value can exist in multiple organizations (`org_name`). Querying `model_enums` with only `project_slug` can return rows belonging to a different org, causing:

- Data leakage across organizations (tenant isolation violation)
- Incorrect enum results being associated with the wrong project
- Silent data corruption that is hard to detect in tests

The combination of `(org_name, project_slug)` is the true composite key that scopes a project to a specific organization.

## Rule to Apply During Code Review

When reviewing any SQL query (raw SQL, sqlc-generated, or ORM-based) that touches the `model_enums` table:

1. Check that the `WHERE` clause includes **both** `org_name` and `project_slug`.
2. If only `project_slug` is present, flag it as a **tenant isolation bug**.
3. This same rule applies to any table that stores project-scoped data with an `org_name` column.

## Correct vs. Incorrect Examples

```sql
-- ❌ WRONG: Only project_slug, missing org_name
SELECT * FROM model_enums WHERE project_slug = $1;

-- ✅ CORRECT: Both org_name and project_slug
SELECT * FROM model_enums WHERE org_name = $1 AND project_slug = $2;
```

```go
// ❌ WRONG in Go/sqlc params
params := db.ListModelEnumsParams{
    ProjectSlug: projectSlug,
}

// ✅ CORRECT
params := db.ListModelEnumsParams{
    OrgName:     orgName,
    ProjectSlug: projectSlug,
}
```

## Scope

This rule applies to all queries on `model_enums` and should be extended as a general principle to any project-scoped table that has an `org_name` column:

- `model_enums`
- Any future tables following the same multi-tenant schema pattern

## Summary

> **Checklist item**: Every `model_enums` query must filter by **both** `org_name` and `project_slug`. Missing `org_name` is a tenant isolation bug.
