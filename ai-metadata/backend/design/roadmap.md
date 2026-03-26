# Roadmap

> ModelCraft 里程碑规划。以实际代码能力为准，如实标注完成状态。

---

## v1.0 — 核心骨架（当前版本）

> 状态：基础可用，各模块较为粗糙，待持续完善。

### ✅ 多租户（基础可用）

- Organization 创建与隔离，以 `Name` 为租户键
- Org 状态管理：active / suspended / deleted（软删除）
- URL 路由级 Org 隔离

### ✅ 项目管理（基础可用）

- Project CRUD，`(OrgName, Slug)` 复合主键
- Project 与 DatabaseCluster 1:1 关联/解绑
- Project 归档 / 激活

### ✅ 用户与成员管理（基础可用）

- User 自动创建（首次登录时通过 ExternalID 关联 Casdoor 用户）
- Membership：用户加入 Org，邀请流程（invited → active）
- 成员挂起 / 激活

### ✅ 权限系统 RBAC（基础可用）

- Role 定义，支持系统角色（不可删除）
- Permission 值对象，格式 `{resource}:{action}`，支持通配符
- UserRole 分配，Casbin enforcer 集成
- 权限缓存

### ✅ 认证（部分完成）

- Casdoor JWT 集成：token 验证、Claims 提取
- 项目级 ProjectAuthConfig，支持配置多种 provider
- ⚠️ Keycloak / OIDC：结构预留，未完整实现

### ✅ 数据库集群管理（基础可用）

- DatabaseCluster CRUD，MySQL 连接参数管理
- 密码加密存储（不明文持久化）
- 连接超时配置（5-15 秒）
- 连接池由 Infrastructure 层管理

### ✅ 模型设计（基础可用）

- DataModel CRUD，ModelLocator 定位机制
- FieldDefinition：20+ 字段类型和格式
- EnumDefinition：单选 / 多选枚举
- LogicalForeignKey：逻辑外键，成对创建
- ModelGroup：模型分组与排序
- Schema Compare：比对设计态与目标库实际 Schema 的差异
- Schema Sync：将差异 DDL Apply 到目标 MySQL
- Reverse Engineering：从现有数据库表反向导入模型

### ✅ 运行态 GraphQL（基础可用）

- 动态生成 GraphQL Schema（根据已同步的模型结构）
- 完整 CRUD：findUnique / findFirst / findMany / createOne / updateOne / deleteOne
- 批量操作：createMany / updateMany / deleteMany
- 聚合：count / aggregate
- 过滤条件（Prisma 风格）：equals / in / lt / gt / contains / AND / OR / NOT
- 分页支持

---

## Future Milestones

> 以下为规划中的能力，尚未开始实现。

### SQL 编辑器

- 在 ModelCraft 界面直接对目标 MySQL 执行 SQL 查询
- 复用 DatabaseCluster 连接，需独立权限控制
- 仅支持 SQL 系数据库（不支持 NoSQL）

### 认证补全

- Keycloak 完整集成
- 通用 OIDC 支持

### 模型版本管理

- 模型变更历史记录
- Schema 变更回滚

### 多 SQL 方言支持

- PostgreSQL 支持（在 SQL 系范围内扩展）
- 字段类型映射扩展

### 可视化设计器

- 前端图形化模型设计界面
- ER 图展示

---

## 状态说明

| 标记 | 含义 |
|------|------|
| ✅ 基础可用 | 核心功能已实现，但较为粗糙，待持续完善 |
| ⚠️ 部分完成 | 主体实现，但有明确的未完成项 |
| 📋 规划中 | 已识别需求，尚未实现 |
