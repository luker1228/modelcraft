## 1. 数据库与数据迁移

- [x] 1.1 新增 `end_user_project_access` 表、外键与联合索引（含唯一约束 `end_user_id + org_name + project_slug`）
- [x] 1.2 调整 `end_user_users` 唯一键为 `UNIQUE(org_name, username)` 并移除账号层 `project_slug` 强绑定字段
- [x] 1.3 调整 `end_user_accounts`、`end_user_roles`、`end_user_role_users` 的项目绑定字段与约束到 Org 作用域
- [x] 1.4 编写数据回填脚本：合并同组织重复用户名并生成项目访问关系
- [x] 1.5 产出迁移校验脚本（迁移前后记录数、一致性与冲突清单）

## 2. 领域模型与仓储改造

- [x] 2.1 重构 EndUser 领域模型：移除 Project 归属字段，增加 Org 作用域语义
- [x] 2.2 改造 EndUserRepository 查询接口为 Org 作用域（含列表分页与状态更新）
- [x] 2.3 新增 EndUserProjectAccess 领域模型与 Repository（授予、撤销、列表、更新权限包）
- [x] 2.4 实现删除 EndUser 时访问关系级联清理逻辑

## 3. GraphQL 契约与接口实现

- [x] 3.1 在 Org GraphQL 增加 EndUser 生命周期接口（create/list/updateStatus/delete）
- [x] 3.2 在 Project GraphQL 增加访问授权接口（grant/revoke/list/update）
- [x] 3.3 执行 `just generate-gql` 并补全新增 resolver 实现
- [x] 3.4 统一错误码与输入校验（重复授权、无权限、资源不存在等）

## 4. 认证流程改造（一次签发 + 项目自由选择）

- [x] 4.1 改造登录接口：校验凭据后一次签发 JWT，并返回可访问项目全集
- [x] 4.2 实现登录后项目上下文选择机制：仅允许选择已授权项目，且无需二次签发令牌
- [x] 4.3 增加无项目访问 `NO_PROJECT_ACCESS` 错误分支与项目选择权限校验
- [x] 4.4 调整 refresh/logout 的 Org 作用域路径与会话隔离策略

## 5. 测试与发布收口

- [x] 5.1 增加 Repository/Adapter 测试覆盖（唯一约束、级联删除、权限包更新）
- [x] 5.2 增加 GraphQL 集成测试覆盖 Org 与 Project 两套 EndUser 接口
- [x] 5.3 增加认证流程测试覆盖（一次签发、项目自由选择、无项目访问）
- [x] 5.4 发布前执行全链路回归并输出迁移与回滚操作清单
  - 已执行（2026-04-25）：`npm --prefix ./tests-bdd run test:end-user-auth -- --name "被禁用的终端用户无法刷新 token" --name "终端用户选择已授权项目上下文" --name "被禁用的终端用户无法选择项目上下文"`
  - 阻塞已解除：
    - 修复 `Org GraphQL end-user management API` 路由注入，管理接口与认证接口统一走 org-scope 服务；
    - BDD 客户端与 step 已对齐当前后端契约（header/body/错误码匹配）。
  - 回归结果（2026-04-25）：`API_BASE_URL=http://localhost:18080 npm --prefix ./tests-bdd run test:end-user-auth -- --name "被禁用的终端用户无法刷新 token" --name "终端用户选择已授权项目上下文" --name "被禁用的终端用户无法选择项目上下文"` → `3 scenarios (3 passed), 19 steps (19 passed)`。
  - 迁移操作清单：
    1. `cd ./modelcraft-backend && just db up`（应用 schema 变更）
    2. `cd ./modelcraft-backend && just enduser-v2-backfill .env <run_id>`（执行 enduser-v2 数据回填）
    3. `cd ./modelcraft-backend && just enduser-v2-validate .env <run_id> <validate_run_id>`（执行一致性校验并导出冲突）
    4. 记录审计表：`migration_audit_enduser_v2_counts`、`migration_audit_enduser_v2_merge_log`、`migration_audit_enduser_v2_validation_checks`、`migration_audit_enduser_v2_validation_conflicts`。
  - 回滚操作清单：
    1. `cd ./modelcraft-backend && just db down`（回滚最近一次 Atlas migration）
    2. 根据 `<run_id>` 清理回填审计数据（`migration_audit_enduser_v2_*`），并恢复变更前数据库备份。
    3. 重新执行 `just db status` 与 `just enduser-v2-validate .env <run_id>` 确认回滚后无残留冲突。