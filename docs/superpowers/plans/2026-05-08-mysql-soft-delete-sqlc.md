# MySQL Soft Delete for sqlc Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `modelcraft-backend` 建立一套“源码 SQL 即真实 SQL”的软删除体系：一次性 `codemod` 原地改写现有 `.sql` 查询，持续 `lint` 拦截未显式带 `deleted_at` 条件且无注解的查询，并完成 MySQL schema / sqlc / 少量 raw SQL 的收口。

**Architecture:** 在 `cmd/sqlsoftdelete` 提供两个子命令：`codemod` 和 `lint`。两者共用 `internal/tooling/sqlsoftdelete` 里的 policy、注解解析、TiDB MySQL AST 解析、statement walker。`codemod` 只负责一次性把 `db/queries/**/*.sql` 改到规范；`lint` 负责日常守门，直接失败任何未注解且缺少 `deleted_at` 显式过滤的语句。实际物理 `DELETE` 在软删表上统一改写为显式 `UPDATE ... SET deleted_at = ..., delete_token = ...`，从而保持手写 SQL 与 `sqlc` 生成输入一致。

**Tech Stack:** Go 1.25, sqlc, MySQL, TiDB parser AST (`github.com/pingcap/tidb/pkg/parser`), just, testify

---

## File Map

| 文件 | 操作 | 说明 |
|------|------|------|
| `modelcraft-backend/db/soft_delete.yaml` | 新建 | 黑名单、`delete_token` 表清单、注解配置、lint 目标路径 |
| `modelcraft-backend/cmd/sqlsoftdelete/main.go` | 新建 | `codemod` / `lint` CLI 入口 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/config.go` | 新建 | 读取 YAML policy，判定某表是否启用 soft delete / 是否需要 `delete_token` |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/annotation.go` | 新建 | 解析 `@include_deleted` / `@only_deleted` / `@physical_delete` |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/sqlcfile.go` | 新建 | 解析 `-- name:` SQLC block，逐语句处理 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/ast.go` | 新建 | TiDB parser 封装与 AST restore helper |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/lint.go` | 新建 | 语句 lint：缺失 `deleted_at`、错误物理删、JOIN/子查询漏过滤 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/rewrite.go` | 新建 | 一次性 codemod：原地改写 `.sql` 源文件 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/app.go` | 新建 | 供 `main.go` 和测试复用的命令分发逻辑 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/*.test.go` | 新建 | policy / lint / rewrite / app / schema contract 测试 |
| `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/` | 新建 | lint / rewrite fixture |
| `modelcraft-backend/justfile` | 修改 | 增加 `soft-delete-codemod` / `soft-delete-lint` / `generate-sqlc`，并把 lint 纳入 `just lint` |
| `modelcraft-backend/db/schema/mysql/01_project.sql` | 修改 | `projects` 增加 `deleted_at` / `delete_token`，索引改造 |
| `modelcraft-backend/db/schema/mysql/02_database_cluster.sql` | 修改 | `database_clusters` 软删字段、唯一约束改造 |
| `modelcraft-backend/db/schema/mysql/03_model_domain.sql` | 修改 | `models` / `model_groups` / `logical_foreign_keys` / `model_enums` / `field_definitions` 软删字段与唯一键调整 |
| `modelcraft-backend/db/schema/mysql/05_organizations.sql` | 修改 | `organizations` 软删字段与唯一键调整 |
| `modelcraft-backend/db/schema/mysql/06_users.sql` | 修改 | `users` / `profile` 软删字段与唯一键调整；`user_organizations` 保持黑名单物理删 |
| `modelcraft-backend/db/schema/mysql/07_roles_permissions.sql` | 修改 | `roles` 软删字段与唯一键调整；`user_roles` / `role_permissions` 保持黑名单物理删 |
| `modelcraft-backend/db/schema/mysql/12_end_user_auth.sql` | 修改 | `end_user_users` / `end_user_roles` 软删字段与唯一键调整；`end_user_accounts` / `end_user_role_users` 保持黑名单物理删 |
| `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql` | 修改 | `end_user_data_permissions` / `end_user_permission_bundles` 软删字段与唯一键调整；item / relation 表保持黑名单物理删 |
| `modelcraft-backend/db/queries/project.sql` | 修改 | 原地补 `deleted_at = 0`，`ArchiveProject` 限制活跃态 |
| `modelcraft-backend/db/queries/cluster.sql` | 修改 | 原地补过滤与软删 update |
| `modelcraft-backend/db/queries/model.sql` | 修改 | 原地补过滤与软删 update |
| `modelcraft-backend/db/queries/model_group.sql` | 修改 | 原地补过滤与软删 update |
| `modelcraft-backend/db/queries/enum.sql` | 修改 | 原地补过滤与软删 update |
| `modelcraft-backend/db/queries/field.sql` | 修改 | 原地补过滤与软删 update |
| `modelcraft-backend/db/queries/org.sql` | 修改 | `organizations` / `users` 查询补过滤；`DeleteMembership` 保持黑名单物理删 |
| `modelcraft-backend/db/queries/profile.sql` | 修改 | `profile` 和 `user_organizations` / `users` JOIN 过滤补齐 |
| `modelcraft-backend/db/queries/user_auth.sql` | 修改 | `users` 查询补过滤 |
| `modelcraft-backend/db/queries/casbin.sql` | 修改 | `roles` 查询补过滤、`DeleteRole` 改软删；`user_roles` / `role_permissions` 维持物理删 |
| `modelcraft-backend/db/queries/rbac/role.sql` | 修改 | `end_user_roles` 查询补过滤、`DeleteEndUserRole` 改软删 |
| `modelcraft-backend/db/queries/rbac/bundle.sql` | 修改 | `end_user_permission_bundles` 查询补过滤、`DeleteEndUserBundle` 改软删 |
| `modelcraft-backend/db/queries/rbac/permission.sql` | 修改 | `end_user_data_permissions` 查询补过滤、`DeleteEndUserPermission` 改软删 |
| `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go` | 修改 | 手写 raw SQL 追加 `deleted_at = 0`，`DELETE` 改软删 update |
| `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository_test.go` | 修改 | 期望 SQL 从物理删改为软删 update，并覆盖 count/list 的 `deleted_at = 0` |
| `modelcraft-backend/internal/infrastructure/dbgen/*.go` | 生成 | `sqlc` 重生后的代码，不手改 |
| `modelcraft-backend/internal/infrastructure/dbgenwrap/safe_querier_gen.go` | 生成 | `gowrap` 重新生成，不手改 |

---

## Task 1: 建立 soft delete policy 与注解解析基础

**Files:**
- Create: `modelcraft-backend/db/soft_delete.yaml`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/config.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/annotation.go`
- Test: `modelcraft-backend/internal/tooling/sqlsoftdelete/config_test.go`

- [ ] **Step 1: 写失败测试，锁定 policy 语义与注解语义**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/config_test.go
package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicy_DefaultEnabledBlacklistAndDeleteToken(t *testing.T) {
	policy, err := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}

	if !policy.SoftDeleteEnabled("models") {
		t.Fatal("models should be soft-delete enabled by default")
	}
	if policy.SoftDeleteEnabled("refresh_tokens") {
		t.Fatal("refresh_tokens should stay in hard-delete blacklist")
	}
	if !policy.NeedsDeleteToken("models") {
		t.Fatal("models should require delete_token for unique-key reuse")
	}
	if policy.NeedsDeleteToken("role_permissions") {
		t.Fatal("role_permissions should not require delete_token")
	}
}

func TestParseAnnotations(t *testing.T) {
	src := []byte("-- @include_deleted\n-- name: ListAllModels :many\nSELECT * FROM models;")
	ann := ParseAnnotations(src)
	if !ann.IncludeDeleted || ann.OnlyDeleted || ann.PhysicalDelete {
		t.Fatalf("unexpected annotation parse result: %+v", ann)
	}
}

func TestPolicyFileExists(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "..", "db", "soft_delete.yaml")); err != nil {
		t.Fatalf("policy file missing: %v", err)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestLoadPolicy|TestParseAnnotations|TestPolicyFileExists' -v
```
Expected: FAIL，提示包或文件不存在。

- [ ] **Step 3: 写 policy 文件，固定第一版黑名单与 `delete_token` 表清单**

```yaml
# modelcraft-backend/db/soft_delete.yaml
default_mode: enabled
timestamp_unit: unix_milli
annotations:
  include_deleted: "@include_deleted"
  only_deleted: "@only_deleted"
  physical_delete: "@physical_delete"
lint_paths:
  - "db/queries/**/*.sql"
blacklist_tables:
  - refresh_tokens
  - api_keys
  - security_audit_logs
  - project_auth_configs
  - project_auth_schemas
  - model_rls_policies
  - model_field_enum_associations
  - user_organizations
  - user_roles
  - role_permissions
  - end_user_accounts
  - end_user_role_users
  - end_user_bundle_data_permission_items
  - end_user_role_bundles
  - end_user_user_bundles
delete_token_tables:
  - organizations
  - users
  - profile
  - projects
  - database_clusters
  - models
  - model_groups
  - logical_foreign_keys
  - model_enums
  - field_definitions
  - roles
  - end_user_users
  - end_user_roles
  - end_user_data_permissions
  - end_user_permission_bundles
```

- [ ] **Step 4: 实现 config / annotation 解析**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/config.go
package sqlsoftdelete

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Policy struct {
	DefaultMode      string            `yaml:"default_mode"`
	TimestampUnit    string            `yaml:"timestamp_unit"`
	Annotations      AnnotationConfig  `yaml:"annotations"`
	LintPaths        []string          `yaml:"lint_paths"`
	BlacklistTables  []string          `yaml:"blacklist_tables"`
	DeleteTokenTables []string         `yaml:"delete_token_tables"`

	blacklistSet    map[string]struct{}
	deleteTokenSet  map[string]struct{}
}

type AnnotationConfig struct {
	IncludeDeleted string `yaml:"include_deleted"`
	OnlyDeleted    string `yaml:"only_deleted"`
	PhysicalDelete string `yaml:"physical_delete"`
}

func LoadPolicy(path string) (*Policy, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var p Policy
	if err := yaml.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	p.blacklistSet = make(map[string]struct{}, len(p.BlacklistTables))
	for _, table := range p.BlacklistTables {
		p.blacklistSet[strings.ToLower(table)] = struct{}{}
	}
	p.deleteTokenSet = make(map[string]struct{}, len(p.DeleteTokenTables))
	for _, table := range p.DeleteTokenTables {
		p.deleteTokenSet[strings.ToLower(table)] = struct{}{}
	}
	if p.DefaultMode == "" {
		p.DefaultMode = "enabled"
	}
	return &p, nil
}

func (p *Policy) SoftDeleteEnabled(table string) bool {
	_, blacklisted := p.blacklistSet[strings.ToLower(table)]
	return p.DefaultMode == "enabled" && !blacklisted
}

func (p *Policy) NeedsDeleteToken(table string) bool {
	_, ok := p.deleteTokenSet[strings.ToLower(table)]
	return ok
}
```

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/annotation.go
package sqlsoftdelete

import "bytes"

type Annotations struct {
	IncludeDeleted bool
	OnlyDeleted    bool
	PhysicalDelete bool
}

func ParseAnnotations(src []byte) Annotations {
	return Annotations{
		IncludeDeleted: bytes.Contains(src, []byte("@include_deleted")),
		OnlyDeleted:    bytes.Contains(src, []byte("@only_deleted")),
		PhysicalDelete: bytes.Contains(src, []byte("@physical_delete")),
	}
}
```

- [ ] **Step 5: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestLoadPolicy|TestParseAnnotations|TestPolicyFileExists' -v
```
Expected: PASS。

- [ ] **Step 6: 提交**

```bash
cd modelcraft-backend && git add db/soft_delete.yaml internal/tooling/sqlsoftdelete/config.go internal/tooling/sqlsoftdelete/annotation.go internal/tooling/sqlsoftdelete/config_test.go
git commit -m "feat(sqlsoftdelete): add policy and annotation parsing"
```

---

## Task 2: 先把 lint 做出来，强制“SQL 源码必须显式带 soft delete 语义”

**Files:**
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/sqlcfile.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/ast.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/lint.go`
- Test: `modelcraft-backend/internal/tooling/sqlsoftdelete/lint_test.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/missing_deleted_at.sql`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/include_deleted.sql`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/join_missing_child_filter.sql`

- [ ] **Step 1: 写失败测试，覆盖最关键的 4 条 lint 规则**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/lint_test.go
package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFile_FailsWhenSelectOmitsDeletedAt(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "lint", "missing_deleted_at.sql"))
	findings, err := LintFile(policy, "missing_deleted_at.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) == 0 {
		t.Fatal("expected missing deleted_at finding")
	}
}

func TestLintFile_PassesWithIncludeDeletedAnnotation(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "lint", "include_deleted.sql"))
	findings, err := LintFile(policy, "include_deleted.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %+v", findings)
	}
}

func TestLintFile_FailsWhenJoinTableMissesDeletedAt(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "lint", "join_missing_child_filter.sql"))
	findings, err := LintFile(policy, "join_missing_child_filter.sql", src)
	if err != nil {
		t.Fatalf("LintFile() error = %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %+v", findings)
	}
}
```

```sql
-- modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/missing_deleted_at.sql
-- name: GetModelByID :one
SELECT * FROM models WHERE id = ? LIMIT 1;
```

```sql
-- modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/include_deleted.sql
-- @include_deleted
-- name: ListAllModels :many
SELECT * FROM models ORDER BY created_at DESC;
```

```sql
-- modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/lint/join_missing_child_filter.sql
-- name: ListMembershipsWithOrgDetails :many
SELECT m.id, o.display_name
FROM user_organizations m
INNER JOIN organizations o ON m.org_name = o.name
WHERE m.user_id = ?;
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestLintFile' -v
```
Expected: FAIL，`LintFile` 未定义或未返回任何 finding。

- [ ] **Step 3: 实现 SQLC block 解析、AST restore 与 lint 核心**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/sqlcfile.go
package sqlsoftdelete

import "strings"

type QueryBlock struct {
	Header string
	Body   string
}

func SplitSQLCFile(src string) []QueryBlock {
	chunks := strings.Split(src, "-- name:")
	blocks := make([]QueryBlock, 0, len(chunks)-1)
	for _, chunk := range chunks[1:] {
		chunk = strings.TrimSpace(chunk)
		parts := strings.SplitN(chunk, "\n", 2)
		if len(parts) != 2 {
			continue
		}
		blocks = append(blocks, QueryBlock{Header: "-- name: " + parts[0], Body: strings.TrimSpace(parts[1])})
	}
	return blocks
}
```

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/lint.go
package sqlsoftdelete

import (
	"fmt"
	"strings"
)

type Finding struct {
	File    string
	Query   string
	Message string
}

func LintFile(policy *Policy, file string, src []byte) ([]Finding, error) {
	anns := ParseAnnotations(src)
	blocks := SplitSQLCFile(string(src))
	findings := make([]Finding, 0)
	for _, block := range blocks {
		stmt, tables, err := ParseSQLBlock(block.Body)
		if err != nil {
			return nil, fmt.Errorf("%s %s: %w", file, block.Header, err)
		}
		if anns.IncludeDeleted || anns.OnlyDeleted {
			continue
		}
		for _, table := range tables {
			if !policy.SoftDeleteEnabled(table.Name) {
				continue
			}
			if stmt.IsDelete && !anns.PhysicalDelete {
				findings = append(findings, Finding{File: file, Query: block.Header, Message: "physical DELETE on soft-delete table"})
				continue
			}
			if !stmt.HasDeletedAtPredicate(table.AliasOrName()) {
				findings = append(findings, Finding{File: file, Query: block.Header, Message: "missing deleted_at predicate for " + table.Name})
			}
		}
	}
	return findings, nil
}

func RenderFindings(findings []Finding) string {
	var b strings.Builder
	for _, finding := range findings {
		fmt.Fprintf(&b, "%s %s: %s\n", finding.File, finding.Query, finding.Message)
	}
	return b.String()
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestLintFile' -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/tooling/sqlsoftdelete/sqlcfile.go internal/tooling/sqlsoftdelete/ast.go internal/tooling/sqlsoftdelete/lint.go internal/tooling/sqlsoftdelete/lint_test.go internal/tooling/sqlsoftdelete/testdata/lint
git commit -m "feat(sqlsoftdelete): add lint engine for explicit deleted_at rules"
```

---

## Task 3: 实现一次性 codemod，原地把 SQL 改成真实软删除 SQL

**Files:**
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/rewrite.go`
- Test: `modelcraft-backend/internal/tooling/sqlsoftdelete/rewrite_test.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/rewrite/delete_model.sql`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/rewrite/list_models.sql`

- [ ] **Step 1: 写失败测试，锁定 `SELECT` / `UPDATE` / `DELETE` 改写结果与幂等性**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/rewrite_test.go
package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteFile_AddsDeletedAtToSelect(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "rewrite", "list_models.sql"))
	out, changed, err := RewriteFile(policy, src)
	if err != nil {
		t.Fatalf("RewriteFile() error = %v", err)
	}
	if !changed {
		t.Fatal("expected rewrite to change SQL")
	}
	if !strings.Contains(string(out), "deleted_at = 0") {
		t.Fatalf("expected rewritten SQL to contain deleted_at = 0, got: %s", out)
	}
}

func TestRewriteFile_RewritesDeleteIntoSoftDeleteUpdate(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "rewrite", "delete_model.sql"))
	out, changed, err := RewriteFile(policy, src)
	if err != nil {
		t.Fatalf("RewriteFile() error = %v", err)
	}
	if !changed {
		t.Fatal("expected delete rewrite to change SQL")
	}
	got := string(out)
	if !strings.Contains(got, "UPDATE models") || !strings.Contains(got, "SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)") {
		t.Fatalf("unexpected soft-delete SQL: %s", got)
	}
	if !strings.Contains(got, "delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)") {
		t.Fatalf("expected delete_token rewrite, got: %s", got)
	}
}

func TestRewriteFile_IsIdempotent(t *testing.T) {
	policy, _ := LoadPolicy(filepath.Join("..", "..", "..", "db", "soft_delete.yaml"))
	src, _ := os.ReadFile(filepath.Join("testdata", "rewrite", "delete_model.sql"))
	once, _, _ := RewriteFile(policy, src)
	twice, changed, err := RewriteFile(policy, once)
	if err != nil {
		t.Fatalf("RewriteFile() error = %v", err)
	}
	if changed {
		t.Fatal("expected second rewrite to be no-op")
	}
	if string(once) != string(twice) {
		t.Fatal("expected idempotent output")
	}
}
```

```sql
-- modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/rewrite/list_models.sql
-- name: ListModels :many
SELECT * FROM models
WHERE org_name = ?
ORDER BY created_at DESC;
```

```sql
-- modelcraft-backend/internal/tooling/sqlsoftdelete/testdata/rewrite/delete_model.sql
-- name: DeleteModel :exec
DELETE FROM models WHERE id = ?;
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestRewriteFile' -v
```
Expected: FAIL，`RewriteFile` 未定义或没有改写 SQL。

- [ ] **Step 3: 实现 rewrite 核心，删除改成显式 `UPDATE` 且不引入新 sqlc 参数**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/rewrite.go
package sqlsoftdelete

import (
	"bytes"
	"strings"
)

const (
	deletedAtExpr  = "CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED)"
	deleteTokenExpr = "CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)"
)

func RewriteFile(policy *Policy, src []byte) ([]byte, bool, error) {
	blocks := SplitSQLCFile(string(src))
	if len(blocks) == 0 {
		return src, false, nil
	}
	var out bytes.Buffer
	changed := false
	anns := ParseAnnotations(src)
	prefix := extractLeadingComments(string(src))
	out.WriteString(prefix)
	for _, block := range blocks {
		rewritten, blockChanged, err := RewriteBlock(policy, anns, block)
		if err != nil {
			return nil, false, err
		}
		changed = changed || blockChanged
		out.WriteString(block.Header)
		out.WriteByte('\n')
		out.WriteString(rewritten)
		out.WriteString("\n\n")
	}
	return bytes.TrimSpace(out.Bytes()), changed, nil
}

func RewriteBlock(policy *Policy, anns Annotations, block QueryBlock) (string, bool, error) {
	stmt, tables, err := ParseSQLBlock(block.Body)
	if err != nil {
		return "", false, err
	}
	if anns.IncludeDeleted || anns.OnlyDeleted || anns.PhysicalDelete {
		return block.Body, false, nil
	}
	body := block.Body
	for _, table := range tables {
		if !policy.SoftDeleteEnabled(table.Name) {
			continue
		}
		body = stmt.EnsureDeletedAtPredicate(body, table.AliasOrName())
	}
	if stmt.IsDelete {
		main := tables[0]
		body = stmt.RewriteDeleteAsUpdate(body, main.AliasOrName(), deletedAtExpr, policy.NeedsDeleteToken(main.Name), deleteTokenExpr)
	}
	return strings.TrimSpace(body), strings.TrimSpace(body) != strings.TrimSpace(block.Body), nil
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestRewriteFile' -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/tooling/sqlsoftdelete/rewrite.go internal/tooling/sqlsoftdelete/rewrite_test.go internal/tooling/sqlsoftdelete/testdata/rewrite
git commit -m "feat(sqlsoftdelete): add one-time codemod for source SQL"
```

---

## Task 4: 把工具接进命令入口和日常工作流

**Files:**
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/app.go`
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/app_test.go`
- Create: `modelcraft-backend/cmd/sqlsoftdelete/main.go`
- Modify: `modelcraft-backend/justfile`

- [ ] **Step 1: 写失败测试，锁定 CLI 的 `lint` / `codemod --write` 行为**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/app_test.go
package sqlsoftdelete

import (
	"bytes"
	"testing"
)

func TestRun_LintSubcommandReturnsErrorOnMissingDeletedAt(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := Run([]string{"lint", "--config", "../../db/soft_delete.yaml", "./testdata/lint/missing_deleted_at.sql"}, stdout, stderr)
	if err == nil {
		t.Fatal("expected lint command to fail")
	}
	if stderr.Len() == 0 {
		t.Fatal("expected lint findings on stderr")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run TestRun_LintSubcommandReturnsErrorOnMissingDeletedAt -v
```
Expected: FAIL。

- [ ] **Step 3: 实现 `Run()`、`main.go` 与 just recipes**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/app.go
package sqlsoftdelete

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return errors.New("expected subcommand: lint or codemod")
	}
	switch args[0] {
	case "lint":
		fs := flag.NewFlagSet("lint", flag.ContinueOnError)
		configPath := fs.String("config", "db/soft_delete.yaml", "policy file")
		_ = fs.Parse(args[1:])
		policy, err := LoadPolicy(*configPath)
		if err != nil {
			return err
		}
		findings, err := LintPaths(policy, fs.Args())
		if err != nil {
			return err
		}
		if len(findings) > 0 {
			_, _ = io.WriteString(stderr, RenderFindings(findings))
			return fmt.Errorf("soft-delete lint failed: %d findings", len(findings))
		}
		_, _ = io.WriteString(stdout, "soft-delete lint passed\n")
		return nil
	case "codemod":
		fs := flag.NewFlagSet("codemod", flag.ContinueOnError)
		configPath := fs.String("config", "db/soft_delete.yaml", "policy file")
		write := fs.Bool("write", false, "rewrite files in place")
		_ = fs.Parse(args[1:])
		policy, err := LoadPolicy(*configPath)
		if err != nil {
			return err
		}
		return RewritePaths(policy, fs.Args(), *write)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func Main() int {
	if err := Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
```

```go
// modelcraft-backend/cmd/sqlsoftdelete/main.go
package main

import (
	"os"

	"modelcraft/internal/tooling/sqlsoftdelete"
)

func main() {
	os.Exit(sqlsoftdelete.Main())
}
```

```just
# modelcraft-backend/justfile
[doc("Run one-time source SQL codemod for soft delete")]
soft-delete-codemod:
    go run ./cmd/sqlsoftdelete codemod --config db/soft_delete.yaml --write db/queries/**/*.sql

[doc("Lint SQL files for explicit soft delete predicates")]
soft-delete-lint:
    go run ./cmd/sqlsoftdelete lint --config db/soft_delete.yaml db/queries/**/*.sql

[doc("Generate sqlc code after soft-delete lint passes")]
generate-sqlc: install-sqlc soft-delete-lint
    echo "📊 生成 sqlc 代码..."
    sqlc generate
    echo "✅ sqlc 代码生成完成"
```

再把现有 `generate-safe-querier` 里的 `sqlc generate` 替换成 `just --justfile justfile --working-directory . generate-sqlc`，并在 `lint` recipe 开头加 `just soft-delete-lint`。

- [ ] **Step 4: 运行 CLI 测试与 dry-run 验证**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run TestRun_LintSubcommandReturnsErrorOnMissingDeletedAt -v
cd modelcraft-backend && just --dry-run soft-delete-lint
cd modelcraft-backend && just --dry-run soft-delete-codemod
```
Expected: 测试 PASS；dry-run 输出正确的 `go run ./cmd/sqlsoftdelete ...` 命令。

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/tooling/sqlsoftdelete/app.go internal/tooling/sqlsoftdelete/app_test.go cmd/sqlsoftdelete/main.go justfile
git commit -m "feat(sqlsoftdelete): wire codemod and lint into CLI and just"
```

---

## Task 5: 用 schema contract test 约束真实 DDL，然后补齐 `deleted_at` / `delete_token`

**Files:**
- Create: `modelcraft-backend/internal/tooling/sqlsoftdelete/schema_contract_test.go`
- Modify: `modelcraft-backend/db/schema/mysql/01_project.sql`
- Modify: `modelcraft-backend/db/schema/mysql/02_database_cluster.sql`
- Modify: `modelcraft-backend/db/schema/mysql/03_model_domain.sql`
- Modify: `modelcraft-backend/db/schema/mysql/05_organizations.sql`
- Modify: `modelcraft-backend/db/schema/mysql/06_users.sql`
- Modify: `modelcraft-backend/db/schema/mysql/07_roles_permissions.sql`
- Modify: `modelcraft-backend/db/schema/mysql/12_end_user_auth.sql`
- Modify: `modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql`

- [ ] **Step 1: 写失败测试，直接验证真实 schema 文件是否带软删字段和唯一键改造**

```go
// modelcraft-backend/internal/tooling/sqlsoftdelete/schema_contract_test.go
package sqlsoftdelete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSchemaContract_SoftDeleteColumnsExist(t *testing.T) {
	files := []string{
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "01_project.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "02_database_cluster.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "03_model_domain.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "05_organizations.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "06_users.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "07_roles_permissions.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "12_end_user_auth.sql"),
		filepath.Join("..", "..", "..", "db", "schema", "mysql", "13_rbac_permissions.sql"),
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile(%s): %v", file, err)
		}
		text := string(body)
		if !strings.Contains(text, "`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0") {
			t.Fatalf("%s missing deleted_at column", file)
		}
	}
}

func TestSchemaContract_DeleteTokenUniqueIndexesExist(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "db", "schema", "mysql", "03_model_domain.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "`idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`, `delete_token`)") {
		t.Fatal("models unique key must include delete_token")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestSchemaContract' -v
```
Expected: FAIL，真实 schema 还没有 `deleted_at` / `delete_token`。

- [ ] **Step 3: 修改 DDL，给软删表加字段并让唯一键避让墓碑数据**

```sql
-- 以 modelcraft-backend/db/schema/mysql/03_model_domain.sql 中 models 表为例
`deleted_at` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '软删除时间戳，0 表示活跃',
`delete_token` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '唯一键避让位，0 表示活跃',
...
UNIQUE KEY `idx_models_name` (`org_name`, `project_slug`, `database_name`, `name`, `delete_token`) COMMENT '组织+项目内模型名称唯一索引',
KEY `idx_models_live_project` (`org_name`, `project_slug`, `deleted_at`) COMMENT '项目活跃模型查询索引'
```

按同一模式修改这些表：

- `projects`
- `database_clusters`
- `models`
- `model_groups`
- `logical_foreign_keys`
- `model_enums`
- `field_definitions`
- `organizations`
- `users`
- `profile`
- `roles`
- `end_user_users`
- `end_user_roles`
- `end_user_data_permissions`
- `end_user_permission_bundles`

保留黑名单表不动：

- `refresh_tokens`
- `api_keys`
- `security_audit_logs`
- `project_auth_configs`
- `project_auth_schemas`
- `model_rls_policies`
- `model_field_enum_associations`
- `user_organizations`
- `user_roles`
- `role_permissions`
- `end_user_accounts`
- `end_user_role_users`
- `end_user_bundle_data_permission_items`
- `end_user_role_bundles`
- `end_user_user_bundles`

- [ ] **Step 4: 运行 schema contract test，确认通过**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run 'TestSchemaContract' -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add internal/tooling/sqlsoftdelete/schema_contract_test.go db/schema/mysql/01_project.sql db/schema/mysql/02_database_cluster.sql db/schema/mysql/03_model_domain.sql db/schema/mysql/05_organizations.sql db/schema/mysql/06_users.sql db/schema/mysql/07_roles_permissions.sql db/schema/mysql/12_end_user_auth.sql db/schema/mysql/13_rbac_permissions.sql
git commit -m "feat(mysql): add soft-delete columns and unique-key escape hatches"
```

---

## Task 6: 对真实 `db/queries` 跑一次 codemod，并人工 review 关键 SQL

**Files:**
- Modify: `modelcraft-backend/db/queries/project.sql`
- Modify: `modelcraft-backend/db/queries/cluster.sql`
- Modify: `modelcraft-backend/db/queries/model.sql`
- Modify: `modelcraft-backend/db/queries/model_group.sql`
- Modify: `modelcraft-backend/db/queries/enum.sql`
- Modify: `modelcraft-backend/db/queries/field.sql`
- Modify: `modelcraft-backend/db/queries/org.sql`
- Modify: `modelcraft-backend/db/queries/profile.sql`
- Modify: `modelcraft-backend/db/queries/user_auth.sql`
- Modify: `modelcraft-backend/db/queries/casbin.sql`
- Modify: `modelcraft-backend/db/queries/rbac/role.sql`
- Modify: `modelcraft-backend/db/queries/rbac/bundle.sql`
- Modify: `modelcraft-backend/db/queries/rbac/permission.sql`

- [ ] **Step 1: 先写一个 repo-wide integration test，确保 codemod 产物能通过 lint**

```go
// 追加到 modelcraft-backend/internal/tooling/sqlsoftdelete/app_test.go
func TestLintPaths_RealQueryTreePassesAfterCodemod(t *testing.T) {
	policy, err := LoadPolicy("../../../db/soft_delete.yaml")
	if err != nil {
		t.Fatal(err)
	}
	findings, err := LintPaths(policy, []string{"../../../db/queries/**/*.sql"})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected repo query tree to pass lint, got:\n%s", RenderFindings(findings))
	}
}
```

- [ ] **Step 2: 运行测试，确认当前一定失败**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run TestLintPaths_RealQueryTreePassesAfterCodemod -v
```
Expected: FAIL，当前 repo 里大量查询未带 `deleted_at = 0`。

- [ ] **Step 3: 对真实 SQL 跑一次 codemod，然后逐个检查关键 diff**

```bash
cd modelcraft-backend && go run ./cmd/sqlsoftdelete codemod --config db/soft_delete.yaml --write db/queries/**/*.sql
cd modelcraft-backend && git diff -- db/queries/project.sql db/queries/model.sql db/queries/org.sql db/queries/casbin.sql db/queries/rbac/role.sql db/queries/rbac/bundle.sql db/queries/rbac/permission.sql
```
Expected: 关键 diff 应满足：

- `SELECT` / `COUNT` / `EXISTS` 显式带 `deleted_at = 0`
- `JOIN` 上的 soft-delete 表在 `ON` 或 `WHERE` 中带各自 `deleted_at = 0`
- `DELETE FROM models ...` 改成：

```sql
UPDATE models
SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED)
WHERE id = ?
  AND deleted_at = 0;
```

- 黑名单表保持物理删除，例如：

```sql
DELETE FROM role_permissions WHERE role_id = ?;
DELETE FROM user_roles WHERE role_id = ?;
DELETE FROM end_user_bundle_data_permission_items WHERE bundle_id = ?;
```

- [ ] **Step 4: 跑 repo-wide lint 测试，确认真实 SQL 树现在通过**

```bash
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -run TestLintPaths_RealQueryTreePassesAfterCodemod -v
cd modelcraft-backend && just soft-delete-lint
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
cd modelcraft-backend && git add db/queries/project.sql db/queries/cluster.sql db/queries/model.sql db/queries/model_group.sql db/queries/enum.sql db/queries/field.sql db/queries/org.sql db/queries/profile.sql db/queries/user_auth.sql db/queries/casbin.sql db/queries/rbac/role.sql db/queries/rbac/bundle.sql db/queries/rbac/permission.sql internal/tooling/sqlsoftdelete/app_test.go
git commit -m "refactor(sql): codemod query sources for explicit soft delete"
```

---

## Task 7: 补齐不在 `db/queries` 下的 raw SQL，重新生成 sqlc 并做最终验证

**Files:**
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go`
- Modify: `modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository_test.go`
- Generate: `modelcraft-backend/internal/infrastructure/dbgen/*.go`
- Generate: `modelcraft-backend/internal/infrastructure/dbgenwrap/safe_querier_gen.go`

- [ ] **Step 1: 先写失败测试，要求 raw SQL 仓储也遵守 soft delete 规则**

```go
// 追加到 modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository_test.go
func TestSqlEndUserRepository_Delete_SoftDeletesInsteadOfPhysicalDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewSqlEndUserRepository(db, "org-a", "project-a")
	mock.ExpectExec("UPDATE end_user_users").
		WithArgs("user-404", "org-a").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), "org-a", "user-404"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}
```

同时把现有查询期望改成显式活跃态：

```go
mock.ExpectQuery("SELECT COUNT\(\*\) FROM end_user_users WHERE org_name = \? AND deleted_at = 0")
mock.ExpectQuery("FROM end_user_users\s+WHERE org_name = \?\s+AND deleted_at = 0")
```

- [ ] **Step 2: 运行针对性测试，确认失败**

```bash
cd modelcraft-backend && go test ./internal/infrastructure/repository -run 'TestSqlEndUserRepository_(Delete|GetByUsername|ListWithTotal)' -v
```
Expected: FAIL，现有 raw SQL 还是物理删除且没带 `deleted_at = 0`。

- [ ] **Step 3: 修改 raw SQL 仓储，和 schema/query 规则保持一致**

```go
// modelcraft-backend/internal/infrastructure/repository/sql_enduser_repository.go
const deleteQuery = `
	UPDATE end_user_users
	SET deleted_at = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000 AS UNSIGNED),
	    delete_token = CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(6)) * 1000000 AS UNSIGNED),
	    updated_at = NOW()
	WHERE id = ? AND org_name = ? AND deleted_at = 0
`

const getByIDQuery = `
	SELECT id, username, password, is_forbidden, created_by, created_at, updated_at
	FROM end_user_users
	WHERE id = ? AND org_name = ? AND deleted_at = 0
`
```

同样把：

- `GetByUsername`
- `UpdateStatus`
- `ListWithTotal` 的 count/list SQL
- `ListAccessibleProjectsByRoleAssignment` 里涉及 `end_user_roles` 的 JOIN

都补上 `deleted_at = 0`。

- [ ] **Step 4: 重跑 sqlc 生成与 targeted verification**

```bash
cd modelcraft-backend && just generate-sqlc
cd modelcraft-backend && just generate-safe-querier
cd modelcraft-backend && go test ./internal/tooling/sqlsoftdelete -v
cd modelcraft-backend && go test ./internal/infrastructure/repository -run 'TestSqlEndUserRepository_(Delete|GetByUsername|ListWithTotal)' -v
cd modelcraft-backend && go test ./internal/infrastructure/repository -run 'TestProjectTo|TestModelGroup|TestSqlCasbin' -v
cd modelcraft-backend && just soft-delete-lint
```
Expected:

- 工具包测试 PASS
- raw SQL 仓储测试 PASS
- 受 sqlc 影响的仓储测试 PASS
- `soft-delete-lint` PASS

- [ ] **Step 5: 运行最后一组全局检查并提交**

```bash
cd modelcraft-backend && just db up
cd modelcraft-backend && just lint
cd modelcraft-backend && go test ./... 
git add internal/infrastructure/repository/sql_enduser_repository.go internal/infrastructure/repository/sql_enduser_repository_test.go internal/infrastructure/dbgen internal/infrastructure/dbgenwrap/safe_querier_gen.go
git commit -m "feat(sqlc): enforce soft delete across schema queries and repositories"
```
Expected: `just lint` 与 `go test ./...` 全绿。如果 `just db up` 因本地数据库未启动失败，先启动本地 MySQL，再重跑，不要跳过 schema 验证。

---

## Self-Review

- Spec coverage: 覆盖了用户确认的 4 个核心点：`deleted_at BIGINT`、默认黑名单启用、一次性 `codemod`、持续 `lint`。额外补上了真实仓库里未走 `sqlc` 的 raw SQL 收口，否则 lint 只能覆盖 `.sql` 文件，方案有漏洞。
- Placeholder scan: 没有 `TODO/TBD`；每个任务都给了确切文件、命令、预期结果和示例代码。
- Type consistency: 统一使用 `Policy`、`Annotations`、`LintFile`、`RewriteFile`、`Run` 这些名字；删除改写统一使用 SQL 内联 timestamp 表达式，避免把 `sqlc` 方法签名改坏。

