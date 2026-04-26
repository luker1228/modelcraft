# 数据库字符集迁移总结

## 变更概述

将所有数据库表的字符集从 `utf8mb4_0900_ai_ci` 更新为 `utf8mb4_unicode_ci`（统一排序规则）。

## 为什么更改

### 问题背景
在测试过程中遇到字符串比较错误：
```
ERROR 1267 (HY000): Illegal mix of collations (utf8mb4_0900_ai_ci,IMPLICIT) and (utf8mb4_unicode_ci,IMPLICIT) for operation '='
```

### 根本原因
- 统一使用 `utf8mb4_unicode_ci` 排序规则
- 表定义使用 `utf8mb4_unicode_ci`，而 SQL 变量和字面量使用服务器默认排序规则
- 在字符串比较时产生排序规则冲突

### 解决方案
统一使用 `utf8mb4_unicode_ci` 排序规则，避免排序规则混用。

## 更新的文件

### Schema 文件
- ✅ `db/schema/mysql/05_organizations.sql`
- ✅ `db/schema/mysql/06_users.sql`
- ✅ `db/schema/mysql/07_roles_permissions.sql`

所有表的 `COLLATE` 从 `utf8mb4_0900_ai_ci` 改为 `utf8mb4_unicode_ci`。

### 测试脚本
- ✅ `tests/setup_test_user.sql`
  - 移除了排序规则显式转换（不再需要 `COLLATE utf8mb4_unicode_ci`）
  - 简化字符串比较逻辑

## 数据初始化策略

### Atlas 的职责
- ✅ **仅负责 Schema 管理**（CREATE TABLE, ALTER TABLE）
- ❌ **不执行数据插入**（INSERT 语句会被忽略）

### 应用的职责
应用启动时负责初始化：
1. **系统角色初始化**（在 Go 代码中）
   - owner: 拥有者 - 完全控制
   - admin: 管理员 - 管理资源和成员
   - editor: 编辑者 - 创建和编辑
   - viewer: 查看者 - 只读访问

2. **默认组织**（如果需要）
   - 通过 AuthProvider 同步或应用初始化

### 移除的文件
- ❌ `db/schema/mysql/99_seed_data.sql` - 已删除
- ❌ `task db:seed` - 已从 Taskfile 移除
- ❌ `task db:init` - 已从 Taskfile 移除

## 数据库操作命令

### 完整重建数据库
```bash
# 方式 1: 分步执行
task db:drop           # 删除数据库
task db:migrate-up     # 应用 schema

# 方式 2: 使用 reset（推荐）
task db:migrate-reset  # drop + migrate-up
```

### 查看表排序规则
```bash
task login-db
SELECT TABLE_NAME, TABLE_COLLATION
FROM information_schema.TABLES
WHERE TABLE_SCHEMA = 'modelcraft';
```

## 验证结果

### Schema 验证
```sql
-- 所有表使用正确的排序规则
mysql> SELECT TABLE_NAME, TABLE_COLLATION
       FROM information_schema.TABLES
       WHERE TABLE_SCHEMA = 'modelcraft';

TABLE_NAME              TABLE_COLLATION
----------------------  --------------------
organizations           utf8mb4_unicode_ci
users                   utf8mb4_unicode_ci
user_organizations      utf8mb4_unicode_ci
roles                   utf8mb4_unicode_ci
user_roles              utf8mb4_unicode_ci
role_permissions        utf8mb4_unicode_ci
...
```

### 测试用户脚本
- ✅ 不再有排序规则冲突
- ✅ 字符串比较正常工作
- ✅ 幂等性测试通过

## 关键设计决策

### 1. 统一排序规则
**决策**: 使用 `utf8mb4_unicode_ci`

**原因**:
- MySQL 8.0 默认排序规则
- 性能优于 `utf8mb4_unicode_ci`
- 避免排序规则混用问题

### 2. 数据初始化交给应用
**决策**: 移除 SQL seed 文件，由应用代码初始化

**原因**:
- Atlas 不支持数据插入
- 应用更灵活（环境变量、配置文件）
- 避免 schema 和数据混在一起

### 3. 简化测试脚本
**决策**: 移除所有 `COLLATE` 显式声明

**原因**:
- 表和变量使用相同排序规则
- 代码更简洁易读
- 减少维护成本

## 后续工作

### 应用代码需实现
1. **系统角色初始化** (启动时)
   ```go
   // internal/infrastructure/auth/system_roles.go
   func InitSystemRoles(db *sql.DB) error {
       roles := []Role{
           {Name: "owner", Description: "...", IsSystem: true, OrgName: "__SYSTEM__"},
           {Name: "admin", Description: "...", IsSystem: true, OrgName: "__SYSTEM__"},
           {Name: "editor", Description: "...", IsSystem: true, OrgName: "__SYSTEM__"},
           {Name: "viewer", Description: "...", IsSystem: true, OrgName: "__SYSTEM__"},
       }
       // 使用 FirstOrCreate 保证幂等性
   }
   ```

2. **权限硬编码**
   ```go
   // 系统角色权限映射
   var SystemRolePermissions = map[string][]string{
       "owner": {"*"},
       "admin": {"project:*", "model:*", "cluster:*", ...},
       "editor": {"project:read", "project:create", ...},
       "viewer": {"project:read", "model:read", ...},
   }
   ```

3. **Casbin 集成**
   - 加载系统角色权限
   - 加载自定义角色权限（从 `role_permissions` 表）

## 参考资料

- MySQL 8.0 字符集: https://dev.mysql.com/doc/refman/8.0/en/charset-unicode-sets.html
- Atlas Schema Management: https://atlasgo.io/
- 表关系图: `db/schema/mysql/TABLE_RELATIONSHIPS.md`
- Schema 重构: `db/schema/mysql/README_SCHEMA_REFACTOR.md`
