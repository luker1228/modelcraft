---
phase: 3
slug: end-user-schema-cleanup
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-05
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | `modelcraft-backend/` (Makefile / justfile) |
| **Quick run command** | `cd modelcraft-backend && go build ./...` |
| **Full suite command** | `cd modelcraft-backend && just build && just lint` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** `go build ./...`
- **After every plan wave:** `just build && just lint`
- **Before `/gsd-verify-work`:** Full suite must be green（含 `just generate-gql`）
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|--------|
| 03-01-01 | 01 | 1 | SCHEMA-01 | T-03-01 | 删除 endusergraphql 路由调用后编译无悬空 import | build | `cd modelcraft-backend && go build ./...` | ⬜ pending |
| 03-01-02 | 01 | 1 | SCHEMA-01 | — | routes.go 无 endusergraphql 包引用 | grep | `grep -c 'endusergraphql' modelcraft-backend/internal/interfaces/http/routes.go` 输出 0 | ⬜ pending |
| 03-02-01 | 02 | 2 | SCHEMA-01 | — | enduser/ 目录不存在 | existence | `test ! -d modelcraft-backend/internal/interfaces/graphql/enduser && echo OK` | ⬜ pending |
| 03-02-02 | 02 | 2 | SCHEMA-01 | — | api/graph/end_user/ 目录不存在 | existence | `test ! -d modelcraft-backend/api/graph/end_user && echo OK` | ⬜ pending |
| 03-02-03 | 02 | 2 | SCHEMA-01, SCHEMA-02 | — | gqlgen.end_user.yml 不存在，generate-gql 通过 | build | `cd modelcraft-backend && test ! -f gqlgen.end_user.yml && just generate-gql && echo OK` | ⬜ pending |
| 03-02-04 | 02 | 2 | SCHEMA-01 | — | 完整构建通过 | build | `cd modelcraft-backend && just build && just lint` | ⬜ pending |
| 03-03-01 | 03 | 3 | SCHEMA-01 | — | gateway 无 end-user GraphQL 路由 | grep | `grep -c 'graphql/end-user' modelcraft-gateway/cmd/gateway/main.go` 输出 0 | ⬜ pending |
| 03-03-02 | 03 | 3 | SCHEMA-01 | — | gateway Deprecated 代码已删除 | grep | `grep -c 'VerifyEndUserAccessToken\|EndUserClaims\|EndUserJWTSecret' modelcraft-gateway/internal/auth/service.go` 输出 0 | ⬜ pending |
| 03-03-03 | 03 | 3 | SCHEMA-01 | — | gateway 构建通过 | build | `cd modelcraft-gateway && go build ./... && go vet ./...` | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

此阶段为删除/清理操作，无需新建测试文件。

Existing infrastructure covers all phase requirements:
- `go build ./...` — 编译验证覆盖所有需求
- `just lint` — 代码质量验证

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| 原 end_user 路由访问返回 404 | SCHEMA-01 | 需要运行服务 | `curl -s http://localhost:8080/graphql/end-user/org/test/project/test` 应返回 404 |
| 9 条 query 处置状态核查 | SCHEMA-02 | 人工确认文档 | 对照 RESEARCH.md §3 表格，确认每条 query 已处置 |
