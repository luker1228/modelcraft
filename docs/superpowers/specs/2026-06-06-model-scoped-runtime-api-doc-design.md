# Model-Scoped Runtime API Doc Design

**Date:** 2026-06-06  
**Status:** Approved  
**Scope:** End-user record page model-scoped API documentation modal for Runtime GraphQL

---

## 1. Summary

Replace the current generic end-user API-docs guidance with a model-scoped API documentation modal opened from the end-user record page.

The modal exists for one narrow goal: help a customer understand how to call the current model's Runtime GraphQL API without learning GraphQL from scratch.

This modal does **not** try to be a full Playground, schema explorer, or GraphQL tutorial.

---

## 2. Problem

The current global API docs page has two structural limits:

1. Runtime GraphQL is dynamic, so a global static example cannot stay high-quality for every model.
2. Customers who are not familiar with GraphQL need a concrete entry point, not a generic query reference.

Because the user is already inside a concrete record workspace, the frontend already knows:

- `orgName`
- `projectSlug`
- `databaseName`
- `modelName`

That context should be used to generate documentation that is directly tied to the current model.

---

## 3. Goals

The modal must do exactly four things:

1. Teach the customer what the Runtime API URL means
2. Teach the customer how to pass the API token
3. Provide one working `curl` example for the current model
4. Show the customer how to use AI to continue from that example

---

## 4. Non-Goals

This change does **not** include:

- A standalone GraphiQL page
- A generic GraphQL playground inside the docs
- Python examples
- Field reference tables
- Auto-generated `create` / `update` / `delete` examples
- A complete GraphQL tutorial
- Replacing the existing token management flow

The documentation is intentionally minimal. It should optimize for first success, not exhaustiveness.

---

## 5. Entry Point

### 5.1 Placement

Add a new `API 文档` action inside the end-user record workspace for a concrete model.

The action opens a modal or sheet-style overlay from the current page. It must not navigate away from the record workspace.

### 5.2 Why record-page scoped

The record page is the right entry point because it already carries the full runtime context.

This avoids forcing the customer to manually reconstruct:

- which org they are in
- which project they are in
- which database they are targeting
- which model this API belongs to

---

## 6. Content Structure

The modal contains exactly five sections in this order.

### 6.1 Server URL

Show the environment base URL explicitly:

```text
http://lukemxjia.devcloud.woa.com:9080
```

Accompany it with one short explanation:

> This is the server address. The Runtime GraphQL path is appended after it.

### 6.2 Runtime URL Meaning

Show the fully qualified endpoint for the current model:

```text
http://lukemxjia.devcloud.woa.com:9080/end-user/graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
```

In the UI, the primary display must use the **real current values**, not placeholders.

The explanation below it must teach the meaning of each path segment:

- `org` = organization
- `project` = project
- `db` = database
- `model` = model

This section exists to teach structure, not just expose a string.

### 6.3 Token Usage

Show only one supported header pattern:

```text
Authorization: Bearer <API_TOKEN>
```

Accompany it with one short explanation:

> Replace `<API_TOKEN>` with the token created on the API Token page.

Do not explain session tokens, access tokens, JWT internals, or browser auth in this modal.

### 6.4 Single Working curl Example

Provide exactly one example, and make it the canonical first example.

The example uses:

- the fixed server URL
- the current model endpoint
- one minimal runtime query
- the runtime operation name `findMany`

The example shape is:

```bash
SERVER_URL="http://lukemxjia.devcloud.woa.com:9080"
TOKEN="replace-with-your-api-token"

curl -X POST "${SERVER_URL}/end-user/graphql/org/<orgName>/project/<projectSlug>/db/<db>/model/<model>" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"query":"query { findMany(take: 5, skip: 0) { items { id } } }"}'
```

Rules:

- Only one example is shown in the first version
- The query must use `findMany`
- The selection set must stay minimal: `items { id }`
- The customer only needs to replace the token before use
- The displayed endpoint should use real runtime values for the current model

We intentionally do not generate richer field selections in v1 because a guessed example is worse than a minimal correct one.

### 6.5 AI Prompt

Provide a copyable prompt template that helps the customer continue with AI.

The prompt must include:

- server URL
- full current endpoint
- token header format
- the working `findMany` example
- a placeholder for the customer's business goal

The expected AI output should be constrained to:

1. a GraphQL query or mutation
2. a corresponding `curl` command
3. explicit notes about which fields need replacement

This section exists because the runtime API is a complete GraphQL interface and AI can build on a known-good starting example.

---

## 7. Data Sources

The modal content is generated from existing runtime context already available on the record page.

Required inputs:

- `orgName`
- `projectSlug`
- `databaseName`
- `modelName`

The implementation must not require a new backend endpoint for v1.

---

## 8. UI Behavior

### 8.1 Presentation

Use the existing frontend dialog or sheet primitives already used by end-user workspace flows.

The tone should be documentation-first, not tool-first:

- readable blocks
- short explanations
- copy-friendly code blocks
- no playground editor

### 8.2 Copy UX

The modal should support copying:

- the full endpoint
- the `Authorization` header example
- the `curl` command
- the AI prompt

Each copy target should be independent so the customer can copy only what they need.

### 8.3 Empty or missing runtime context

If the current model context is incomplete, the action should not open a broken modal.

Fallback behavior:

- disable the action, or
- open with a clear message that the current model runtime context is unavailable

v1 should prefer disabling the action until all required values are known.

---

## 9. Relationship With Existing Global API Docs

The existing end-user global API docs page should no longer be treated as the primary onboarding surface for runtime usage.

Its role becomes secondary:

- explain API Token creation
- explain the general runtime URL pattern
- direct users to model-specific API docs from a concrete model page

The model-scoped modal becomes the primary place for high-quality runtime examples.

---

## 10. Implementation Shape

Expected frontend scope:

1. Add an `API 文档` trigger inside the end-user record workspace
2. Add a new modal component dedicated to model-scoped runtime API docs
3. Generate the endpoint, header example, `curl`, and AI prompt from the current model context
4. Keep the current global docs page unchanged or lightly de-emphasized in this first step

This should be implemented entirely in the frontend unless a hidden dependency is discovered during coding.

---

## 11. Risks and Constraints

### 11.1 Minimal example vs. rich example

The chosen tradeoff is correctness over completeness.

Using only:

```graphql
query { findMany(take: 5, skip: 0) { items { id } } }
```

reduces the chance that generated docs fail because of model-specific field assumptions.

### 11.2 Environment specificity

The server URL is intentionally exposed as:

```text
http://lukemxjia.devcloud.woa.com:9080
```

This is acceptable for the current target environment, but the implementation should keep the value isolated so it can be changed later if deployment topology changes.

---

## 12. Acceptance Criteria

The change is complete when:

1. A customer can open model-specific API docs from the end-user record page
2. The modal shows the fixed server URL and the current model's full endpoint
3. The modal explains the `Authorization: Bearer <API_TOKEN>` header
4. The modal shows exactly one model-specific `findMany` `curl` example
5. The customer can copy that `curl` and only replace the token before use
6. The modal includes a copyable AI prompt based on the current endpoint and working example
7. No GraphiQL or generic playground is introduced in this change

