# 权限数据模型

> 对应总纲：`00-rbac-overview.md`
> 覆盖内容：资源定义、权限点结构、权限包组合、授权对象

---

## 资源（Resource）

当前系统中的资源统一定义为：

- **表（Model）**

不引入菜单、接口、按钮等其他资源类型。每个 Model 即一个可授权资源单元。

---

## 权限点（Permission）

权限点是系统中的**最小权限定义单元**。

一个权限点完整描述一条权限规则，包含四个维度：

| 维度 | 示例 |
|------|------|
| **资源**（表） | `orders` |
| **动作** | `select` / `insert` / `update` / `delete` / `export` |
| **列策略** | 全部列可见 / 仅 `status`、`remark` 可编辑 / 部分列脱敏 |
| **行策略（RLS）** | `ALL` / `SELF` / `DEPT` / `DEPT_AND_CHILDREN` |

### 行策略枚举

| 值 | 含义 | Model 上必须存在的字段 |
|----|------|----------------------|
| `ALL` | 全部行，不过滤 | 无 |
| `SELF` | 仅当前用户的行 | `owner`（EndUserRef 字段） |
| `DEPT` | 仅当前用户所在部门的行 | `dept_id` |
| `DEPT_AND_CHILDREN` | 当前部门及所有下级部门的行 | `dept_id` |

> **约束**：只有 Model 上存在对应字段，该 `rowScope` 的权限点才合法。
> 若 Model 没有 `owner` 字段，则不能创建 `rowScope = SELF` 的权限点；
> 若 Model 没有 `dept_id` 字段，则不能创建 `rowScope = DEPT / DEPT_AND_CHILDREN` 的权限点。
> 字段不存在时，创建权限点应在前端校验阻断，后端也需返回校验错误。

行策略的 **WHERE 注入由 RLS 引擎执行**，`rowScope` 是策略定义，RLS 是执行机制，两者是上下层关系而非并列。

### 权限点示例

```
权限点 A：订单查看-本人
  资源:   orders
  动作:   select
  列策略: 全部列可见
  行策略: SELF

权限点 B：订单查看-本部门
  资源:   orders
  动作:   select
  列策略: 全部列可见
  行策略: DEPT

权限点 C：订单修改-全部
  资源:   orders
  动作:   update
  列策略: 仅 status、remark 可编辑
  行策略: ALL
```

### 权限点的职责边界

- 权限点只承担**"定义能力"**的职责
- 权限点**不直接用于日常授权**
- 权限点不能单独授予给用户或角色

---

## 权限包（PermissionBundle）

权限包是多个权限点的**有序组合**，是系统中的**唯一正式授权单位**。

### 规则

- 用户可以关联权限包
- 角色可以关联权限包
- 权限点仅作为权限包内部的组成单元存在
- **不开放直接授权限点**的能力

### 权限包示例

```
权限包：订单查看包
  权限点 A：订单查看-本人
  权限点 B：订单查看-本部门

权限包：订单管理包
  权限点 C：订单查看-全部
  权限点 D：订单修改-全部
  权限点 E：订单导出-全部
```

### 权限包的设计意图

| 意图 | 说明 |
|------|------|
| **复用权限定义** | 同一套权限可授予多个用户/角色，改一个地方全量生效 |
| **批量授权** | 一次关联权限包，等于授予包内所有权限点 |
| **降低管理复杂度** | 避免给用户打零散权限补丁 |
| **易于审计** | 知道某用户有哪些权限包，即可知道其完整能力集 |

---

## 授权对象（Grantee）

正式授权对象只包括两类：

| 授权对象 | 说明 |
|----------|------|
| **用户（User）** | 直接与权限包关联 |
| **角色（Role）** | 与权限包关联；用户通过角色间接获得权限 |

### 授权关系图

```
用户 ←──关联──→ 权限包
                   │
用户 ←──属于──→ 角色 ←──关联──→ 权限包

（系统自动注入）
用户 ←──隐式属于──→ 内置隐式角色 ←──关联──→ 权限包
```

### 不作为正式授权对象

- **部门** 不是授权载体，不能直接关联权限包
- 部门只参与行权限数据范围计算，见 [04-department-scope.md](./04-department-scope.md)

---

## 数据模型结构（伪 DDL）

```
Permission（权限点）
├── id
├── modelId          ← 绑定资源（表）
├── action           ← select / insert / update / delete / export
├── columnPolicy     ← JSON，描述列可见/可编辑/脱敏策略
└── rowScope         ← ALL / SELF / DEPT / DEPT_AND_CHILDREN

PermissionBundle（权限包）
├── id
├── name
├── description
└── permissions[]    ← Permission[]

UserPermissionBundle（用户-权限包关联）
├── userId
└── bundleId

RolePermissionBundle（角色-权限包关联）
├── roleId
└── bundleId

UserRole（用户-角色关联）
├── userId
└── roleId

Role（角色）
├── id
├── name
└── isImplicit       ← true = 内置隐式角色（系统自动注入，不逐条落库关系）
```
