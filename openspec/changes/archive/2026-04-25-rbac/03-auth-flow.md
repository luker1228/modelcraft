# 鉴权流程

> 对应总纲：`00-rbac-overview.md`
> 覆盖内容：权限来源合并、权限点展开、判定规则

---

## 权限来源

用户最终有效权限来源于三个通道，**三者取并集**：

```
有效权限集 =
    用户直接关联的权限包（UserPermissionBundle）
  ∪ 用户显式角色关联的权限包（UserRole → RolePermissionBundle）
  ∪ 系统自动注入的隐式角色关联的权限包（Role.isImplicit = true）
```

---

## 鉴权流程（5 步）

```
Step 1: 查询用户直接绑定的权限包
        SELECT * FROM user_permission_bundles WHERE user_id = ?

Step 2: 查询用户显式角色绑定的权限包
        SELECT role_id FROM user_roles WHERE user_id = ?
        → SELECT * FROM role_permission_bundles WHERE role_id IN (...)

Step 3: 自动加入隐式角色绑定的权限包
        SELECT id FROM roles WHERE is_implicit = true
        → SELECT * FROM role_permission_bundles WHERE role_id IN (...)

Step 4: 展开全部权限点
        从以上三个来源的权限包中，展开所有 Permission（权限点）

Step 5: 按目标资源 × 动作 × 列策略 × 行策略进行判定
```

---

## 判定规则

### 核心原则

> **默认拒绝，显式允许。**

```
命中 allow → 通过
未命中    → 拒绝
```

### 无 deny 机制

系统第一版只保留 allow 机制，不引入 deny：

- 有授权就有权限
- 没授权就没有权限
- 不支持"显式拒绝"
- 不处理"允许和拒绝冲突"的优先级问题

这避免了以下复杂性：

- user allow 与 role deny 冲突
- 多来源授权优先级难以解释
- 鉴权逻辑复杂化
- 排障和审计困难

### 多来源合并（取并集）

当用户同时拥有多个权限包时，有效权限是所有来源的并集。

示例：

```
用户直接关联：[权限包 A（订单查看-本人）]
显式角色：    [权限包 B（订单查看-全部）]
隐式角色：    [权限包 C（基础自服务）]

有效权限集 = A ∪ B ∪ C
→ 对 orders 的 select 行策略 = ALL（因 B 包含 ALL，取最大范围）
```

> ⚠️ **行策略取并集**：当同一资源同一动作出现多个行策略时，取范围最广的。
> 例如同时有 `SELF` 和 `ALL`，有效结果为 `ALL`。

---

## 与 RLS 的关系

RBAC 的 `rowScope` 与 RLS 是**上下层关系**，不是两套独立机制：

- **RBAC 权限点**定义策略：该用户在这张 Model 上能看哪些行（`rowScope`）
- **RLS 引擎**负责执行：将 `rowScope` 编译为参数化 SQL WHERE 子句并注入查询

| 层 | 职责 | 触发条件 |
|----|------|----------|
| **RBAC 权限点 `rowScope`** | 定义行过滤策略（SELF/DEPT/DEPT_AND_CHILDREN/ALL） | 授权配置时 |
| **RLS 引擎执行** | 将策略编译为 WHERE 注入 | 每次查询时 |

### 行策略的字段前提

`rowScope` 的执行依赖 Model 上特定字段的存在：

- `SELF` → 要求 Model 有 `owner` 字段（EndUserRef 类型）
- `DEPT` / `DEPT_AND_CHILDREN` → 要求 Model 有 `dept_id` 字段
- `ALL` → 无字段要求

**Model 不满足字段前提时，该 `rowScope` 的权限点不允许创建。**
