# ModelCraft 模型设计规范

## 1. 核心理念

### 1.1 模型的本质

ModelCraft 的核心产物是 **HTTP API**。模型设计的一切都围绕"生成高质量的 HTTP API"展开。

```
输入: 模型定义 (Design-Time Schema)
输出: HTTP API Endpoint + CRUD 能力
  ├── GraphQL API (当前实现)
  └── RESTful API (未来扩展)
```

### 1.2 设计原则

1. **统一的关联机制**: 所有关联都使用 `RELATION` 类型
2. **语义优先**: 字段定义是语义类型，不关心物理存储
3. **产物驱动**: 以 HTTP API 生成为中心目标
4. **分层解耦**: 语义层 → 产物层 (API + Database + Runtime)
5. **枚举为系统配置**: 枚举是静态配置，不允许运行时动态增删

### 1.3 架构分层

```
┌─────────────────────────────────────────────┐
│   语义层 (Semantic Layer)                   │
│   - 用户定义模型和字段                       │
│   - 关注"是什么"                            │
├─────────────────────────────────────────────┤
                  ↓ 编译/生成
┌─────────────────────────────────────────────┐
│   产物层 (Artifact Layer)                   │
│   - HTTP API (GraphQL/RESTful)             │
│   - Database Schema                        │
│   - Runtime Code                           │
│   - 关注"怎么做"                            │
└─────────────────────────────────────────────┘
```

---

## 2. 模型定义

### 2.1 模型结构

```go
type DataModel struct {
    ID          string
    ModelName   string
    Fields      []*Field        // 字段列表
    // ... 其他属性
}
```

### 2.2 普通模型

**用途**: 业务实体建模

**特点**:
- 提供完整的 CRUD API
- 支持所有字段类型（包括关联、枚举）
- 支持复杂查询（过滤、排序、分页、聚合）
- 数据存储在用户数据库

**示例**:
```yaml
model:
  name: order
  
fields:
  - name: id
    type: STRING
    format: UUID
  - name: order_number
    type: STRING
  - name: customer
    type: RELATION
    config:
      target_model: customer
```

**生成的 API**:
```graphql
type Order {
  id: ID!
  order_number: String!
  customer: Customer!
}

type Query {
  order(id: ID!): Order
  orders(where: OrderWhereInput, orderBy: [OrderOrderByInput!]): [Order!]!
  ordersAggregate(where: OrderWhereInput): OrderAggregateResult!
}

type Mutation {
  createOrder(input: CreateOrderInput!): Order!
  updateOrder(id: ID!, input: UpdateOrderInput!): Order!
  deleteOrder(id: ID!): Boolean!
}
```

---

## 3. 字段类型体系

### 3.1 字段结构

```go
type Field struct {
    Name        string
    Type        FieldType       // 语义类型（大类）
    Format      *FieldFormat    // 可选：需要系统行为的格式
    
    // 通用属性
    Required    bool            // 是否必填
    Unique      bool            // 是否唯一
    Indexed     bool            // 是否索引
    DisplayOrder string         // 设计态展示排序（lexicographic fractional index）
    
    // 字段特定配置
    Config      FieldConfig     // 验证配置
}
```

### 3.2 字段类型定义

```go
type FieldType string

const (
    // 基础类型 (Scalar Types)
    TypeString      FieldType = "STRING"
    TypeNumber      FieldType = "NUMBER"
    TypeBoolean     FieldType = "BOOLEAN"
    TypeDateTime    FieldType = "DATETIME"
    TypeJSON        FieldType = "JSON"
    
    // 枚举类型 (Enum Type)
    TypeEnum        FieldType = "ENUM"
    
    // 关联类型 (Relation Type)
    TypeRelation    FieldType = "RELATION"
)
```

**关键设计**：
- 字段类型是**语义类型**，描述"是什么"
- 不关心物理存储（VARCHAR/INT/TIMESTAMP）
- 物理存储由产物生成层决定
- 枚举字段通过 `EnumName` 引用系统配置的枚举定义

### 3.3 Format 的定义

**Format 判断标准**：需要**系统行为**的特殊处理，不只是验证。

```go
type FieldFormat string

const (
    // STRING 的 Format
    FormatUUID      FieldFormat = "UUID"        // 自动生成、不可修改
    
    // NUMBER 的 Format
    FormatInteger   FieldFormat = "INTEGER"     // 影响存储和 GraphQL 类型
    FormatDecimal   FieldFormat = "DECIMAL"     // 精确小数存储
    
    // DATETIME 的 Format
    FormatDate      FieldFormat = "DATE"        // 日期存储
    FormatTime      FieldFormat = "TIME"        // 时间存储
    FormatDateTime  FieldFormat = "DATETIME"    // 日期时间存储
)
```

| 字段 | 有 Format？| 原因 |
|------|-----------|------|
| UUID | ✅ `UUID` | 系统自动生成、不可修改、不可手动填入 |
| Email | ❌ 无 Format | 只是 STRING + pattern 验证 |
| URL | ❌ 无 Format | 只是 STRING + pattern 验证 |
| Phone | ❌ 无 Format | 只是 STRING + pattern 验证 |
| INTEGER | ✅ `INTEGER` | 影响存储类型、GraphQL 类型 (Int vs Float) |
| DECIMAL | ✅ `DECIMAL` | 特殊存储（精度）、计算行为 |
| DATE | ✅ `DATE` | 特殊存储、时区处理、日期函数 |
| DATETIME | ✅ `DATETIME` | 自动时间戳（auto_now）、时区 |

### 3.4 基础类型详解

#### 3.4.1 STRING

```go
type StringConfig struct {
    MinLength   *int            // 最小长度
    MaxLength   *int            // 最大长度
    Pattern     *string         // 正则表达式
}
```

**示例**:
```yaml
# 普通字符串
- name: name
  type: STRING
  config:
    min_length: 1
    max_length: 100

# UUID 主键
- name: id
  type: STRING
  format: UUID
  required: true
  unique: true

# 邮箱（无 Format，只用验证）
- name: email
  type: STRING
  required: true
  config:
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
```

**产物生成**:
```sql
-- MySQL
name VARCHAR(100) NOT NULL
id VARCHAR(26) PRIMARY KEY    -- UUID v7 长度固定 26
email VARCHAR(255) NOT NULL
```

```graphql
type User {
  name: String!
  id: ID!
  email: String!
}
```

#### 3.4.2 NUMBER

```go
type NumberConfig struct {
    Min         *float64        // 最小值
    Max         *float64        // 最大值
    Precision   *int            // 精度（DECIMAL）
    Scale       *int            // 小数位数（DECIMAL）
}
```

**示例**:
```yaml
# 整数
- name: quantity
  type: NUMBER
  format: INTEGER
  config:
    min: 0

# 浮点数
- name: rating
  type: NUMBER
  config:
    min: 0.0
    max: 5.0

# 精确小数（金额）
- name: price
  type: NUMBER
  format: DECIMAL
  config:
    precision: 10
    scale: 2
```

**产物生成**:
```sql
quantity INT UNSIGNED
rating DOUBLE
price DECIMAL(10,2)
```

```graphql
type Product {
  quantity: Int!
  rating: Float!
  price: Decimal!  # Custom scalar
}
```

#### 3.4.3 BOOLEAN

```yaml
- name: is_active
  type: BOOLEAN
```

**产物生成**:
```sql
is_active BOOLEAN DEFAULT TRUE
```

```graphql
type User {
  is_active: Boolean!
}
```

#### 3.4.4 DATETIME

```go
type DateTimeConfig struct {
    AutoNow     bool            // 自动更新为当前时间
    AutoNowAdd  bool            // 创建时自动设置
}
```

**示例**:
```yaml
# 日期
- name: birth_date
  type: DATETIME
  format: DATE

# 时间
- name: open_time
  type: DATETIME
  format: TIME

# 日期时间（自动时间戳）
- name: created_at
  type: DATETIME
  format: DATETIME
  config:
    auto_now_add: true

- name: updated_at
  type: DATETIME
  format: DATETIME
  config:
    auto_now: true
```

**产物生成**:
```sql
birth_date DATE
open_time TIME
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
```

```graphql
scalar Date
scalar Time
scalar DateTime

type User {
  birth_date: Date
  open_time: Time
  created_at: DateTime!
  updated_at: DateTime!
}
```

#### 3.4.5 JSON

```go
type JSONConfig struct {
    Schema      *JSONSchema     // JSON Schema 定义（可选）
}

type JSONSchema struct {
    Type        string                  // object | array
    Properties  map[string]*JSONSchema  // 对象属性
    Items       *JSONSchema             // 数组元素
    Required    []string                // 必填字段
}
```

**示例**:
```yaml
# 简单 JSON
- name: metadata
  type: JSON

# 带 Schema 的 JSON
- name: settings
  type: JSON
  config:
    schema:
      type: object
      properties:
        theme:
          type: string
          enum: ["light", "dark"]
        notifications:
          type: object
          properties:
            email:
              type: boolean
            sms:
              type: boolean
      required: ["theme"]
```

**产物生成**:
```sql
metadata JSON
settings JSON
```

```graphql
scalar JSON

type User {
  metadata: JSON
  settings: JSON
}
```

#### 3.4.6 ENUM

**用途**: 字段值限定为预定义的枚举选项

**特点**:
- 引用系统配置的枚举定义（EnumDefinition）
- 枚举定义存储在系统库，作为项目级配置
- 不允许运行时动态增删枚举值（需通过管理界面配置）
- 支持单选和多选模式

**枚举定义结构**:
```go
type EnumDefinition struct {
    ID            string       // 枚举定义ID
    OrgName       string       // 所属组织
    ProjectSlug   string       // 所属项目
    Name          string       // 枚举名称（英文标识）
    Title         string       // 显示名称
    Description   string       // 描述
    Options       []EnumOption // 枚举选项列表
    IsMultiSelect bool         // 是否支持多选
}

type EnumOption struct {
    Code        string // 枚举值（存储在数据库中）
    Label       string // 显示标签
    Order       int32  // 排序
    Description string // 选项描述
}
```

**字段配置**:
```go
type Field struct {
    Type     FieldType  // "ENUM"
    EnumName string     // 引用的枚举名称
}
```

**示例**:
```yaml
# 1. 定义枚举（项目级配置）
enum_definition:
  name: order_status
  title: 订单状态
  is_multi_select: false
  options:
    - code: PENDING
      label: 待支付
      order: 1
    - code: PAID
      label: 已支付
      order: 2
    - code: SHIPPED
      label: 已发货
      order: 3
    - code: DELIVERED
      label: 已送达
      order: 4

# 2. 字段引用枚举
fields:
  - name: status
    type: ENUM
    enum_name: order_status
    required: true
```

**产物生成**:

```sql
-- MySQL（使用 VARCHAR 存储 code 值）
CREATE TABLE orders (
  status VARCHAR(50) NOT NULL,
  INDEX idx_status (status)
);
-- 注意：不使用 ENUM 类型，保持灵活性
```

```graphql
# GraphQL Schema（生成原生 enum 类型）
enum OrderStatus {
  PENDING
  PAID
  SHIPPED
  DELIVERED
}

type Order {
  id: ID!
  status: OrderStatus!
}

input OrderWhereInput {
  status: OrderStatusFilter
}

input OrderStatusFilter {
  equals: OrderStatus
  not: OrderStatus
  in: [OrderStatus!]
  notIn: [OrderStatus!]
}
```

**关联查询**:
```yaml
# 如果需要查询枚举的扩展信息（label、description）
# 使用虚拟字段 statusLabel
fields:
  - name: status
    type: ENUM
    enum_name: order_status
    
  - name: statusLabel
    type: ENUM_LABEL
    enum_label_config:
      source_field: status
```

```graphql
type Order {
  id: ID!
  status: OrderStatus!
  statusLabel: String!    # 自动从枚举定义中查询 label
}
```

**验证行为**:
- 创建/更新时自动校验值是否在枚举选项中
- 多选模式下，值存储为 JSON 数组
- 枚举定义变更后，现有数据不受影响（使用 code 值存储）

**与 RELATION 的区别**:

| 维度 | ENUM | RELATION |
|------|------|----------|
| 数据存储 | 直接存储 code 值 | 存储外键 ID |
| 查询性能 | 高（无需 JOIN） | 需要 JOIN |
| 数据完整性 | 应用层校验 | 数据库外键约束（可选） |
| 扩展性 | 固定结构（code/label） | 任意字段 |
| 适用场景 | 状态、类型、分类 | 关联实体 |

### 3.5 类型总结

```
┌─────────────────────────────────────────────┐
│          字段类型体系                        │
├─────────────────────────────────────────────┤
│                                             │
│  基础类型 (Scalar Types)                    │
│  ├── STRING                                 │
│  │   └── Format: UUID                      │
│  ├── NUMBER                                 │
│  │   └── Format: INTEGER | DECIMAL         │
│  ├── BOOLEAN                                │
│  ├── DATETIME                               │
│  │   └── Format: DATE | TIME | DATETIME    │
│  └── JSON                                   │
│      └── 配置: JSON Schema (可选)          │
│                                             │
│  枚举类型 (Enum Type)                       │
│  └── ENUM (引用 EnumDefinition)            │
│                                             │
│  关联类型 (Relation Type)                   │
│  └── RELATION (下一节详述)                  │
│                                             │
│  特点:                                      │
│  - 所有类型都是语义类型                      │
│  - 不关心物理存储                            │
│  - 产物层自动推导存储类型                     │
│  - Format 只用于需要系统行为的场景           │
│  - 枚举为系统配置，不允许运行时动态增删       │
│                                             │
└─────────────────────────────────────────────┘
```

---

## 4. 关联类型详解

### 4.1 关联的本质

**核心洞察**：
1. **关联是字段间的映射**：明确指定 `源模型.源字段 → 目标模型.目标字段`
2. **用户只定义关联字段**：系统自动创建关联关系并推导外键列
3. **外键列是实现细节**：不暴露在 API 中，列名与关联字段名相同
4. **枚举不是关联**：枚举字段直接存储 code 值，不需要外键

```
关联关系的三层结构：

┌─────────────────────────────────────────────┐
│  语义层（用户定义）                          │
│  - 关联字段: customer (RELATION)            │
│    config:                                  │
│      target_model: customer                 │
│      target_field: id        ← 必须指定！   │
├─────────────────────────────────────────────┤
│  关联关系层（系统创建）                      │
│  - ModelRelation:                           │
│    source: Order.customer                   │
│    target: Customer.id       ← 完整映射     │
│    type: MANY_TO_ONE                        │
│    foreign_key_mode: LOGICAL                │
├─────────────────────────────────────────────┤
│  物理存储层（产物生成）                      │
│  - Column: customer VARCHAR(26)             │
│  - Index: idx_customer                      │
│  - FK Constraint (可选)                     │
└─────────────────────────────────────────────┘

查询时的数据流：
  customer: "abc123"     →  customer: {id: "abc123", name: "张三"}
       ↑                            ↑
  存储外键值                  通过 target_field 加载完整对象

枚举字段 ≠ 关联：
  status: "PENDING"      →  statusLabel: "待支付"
       ↑                            ↑
  直接存储 code              查询枚举定义（无外键）
```

### 4.2 RELATION 类型

**关联字段配置**：

```go
type RelationConfig struct {
    // 关联目标（必填）
    TargetModel     string          // 目标模型名（如 "customer"）
    TargetField     string          // 目标字段名（如 "id"）← 必须明确指定
    
    // 关联类型
    Type            RelationType    // MANY_TO_ONE | ONE_TO_MANY | MANY_TO_MANY
    
    // 外键配置
    ForeignKeyMode  ForeignKeyMode  // PHYSICAL | LOGICAL | NONE
    
    // 物理外键的级联行为（仅 PHYSICAL 模式）
    OnDelete        *OnDeleteAction // CASCADE | RESTRICT | SET_NULL
    OnUpdate        *OnUpdateAction // CASCADE | RESTRICT | SET_NULL
    
    // 双向关联配置（ONE_TO_MANY / MANY_TO_MANY）
    Reverse         *ReverseConfig  // 反向字段配置
    
    // 多对多
    JoinTable       *JoinTableConfig
}

type ReverseConfig struct {
    FieldName string        // 反向字段名（如 "items"）
    Exposed   bool          // 是否在 API 中暴露
}
```

**系统内部的关联关系**：

```go
// ModelRelation 由系统自动创建，用户不直接操作
type ModelRelation struct {
    ID              string
    
    // 源端（定义关联字段的模型）
    SourceModelID   string        // Order 的 ID
    SourceFieldName string        // "customer"
    
    // 目标端（被关联的模型）
    TargetModelID   string        // Customer 的 ID
    TargetFieldName string        // "id" ← 从 RelationConfig.TargetField 获取
    
    // 关联类型
    Type            RelationType
    
    // 外键配置
    ForeignKeyMode  ForeignKeyMode
    OnDelete        *OnDeleteAction
    OnUpdate        *OnUpdateAction
}
```

**类型定义**：

```go
type RelationType string

const (
    RelationManyToOne  RelationType = "MANY_TO_ONE"   // 多对一
    RelationOneToMany  RelationType = "ONE_TO_MANY"   // 一对多
    RelationManyToMany RelationType = "MANY_TO_MANY"  // 多对多
)

type ForeignKeyMode string

const (
    FKPhysical  ForeignKeyMode = "PHYSICAL"  // 数据库外键约束
    FKLogical   ForeignKeyMode = "LOGICAL"   // 应用层校验存在性
    FKNone      ForeignKeyMode = "NONE"      // 无约束（特殊场景）
)

type OnDeleteAction string

const (
    OnDeleteCascade  OnDeleteAction = "CASCADE"   // 级联删除
    OnDeleteRestrict OnDeleteAction = "RESTRICT"  // 禁止删除
    OnDeleteSetNull  OnDeleteAction = "SET_NULL"  // 设为 NULL
)

type OnUpdateAction string

const (
    OnUpdateCascade  OnUpdateAction = "CASCADE"   // 级联更新
    OnUpdateRestrict OnUpdateAction = "RESTRICT"  // 禁止更新
    OnUpdateSetNull  OnUpdateAction = "SET_NULL"  // 设为 NULL
)
```

### 4.3 TargetField 的作用

**为什么必须指定 target_field？**

1. **明确关联映射**：`Order.customer → Customer.id`
2. **支持非主键关联**：可关联到唯一字段（如 code, sku）
3. **决定输入类型**：target_field 的类型决定 GraphQL 输入类型
4. **生成正确的外键**：数据库外键约束需要知道目标列

**TargetField 的要求**：
- ✅ 必须是目标模型的字段
- ✅ 必须是唯一字段（Unique）或主键（Primary）
- ✅ 不能是关联字段（不支持嵌套关联）

**常见场景**：

| 场景 | target_field | 类型 | 说明 |
|------|-------------|------|------|
| 关联到主键 | `id` | UUID | 最常见，关联到自动生成的 ID |
| 关联到唯一标识 | `code` | String | 枚举、配置表常用 |
| 关联到业务唯一键 | `sku` | String | 商品 SKU、订单号等 |
| 关联到邮箱 | `email` | String | 用户邮箱（如果唯一） |

### 4.4 外键模式对比

| 维度 | PHYSICAL | LOGICAL | NONE |
|------|----------|---------|------|
| **定义位置** | 数据库 Schema | 应用层元数据 | 无 |
| **完整性保证** | 数据库引擎 | 应用代码 | 无 |
| **关联查询** | ✅ 支持 | ✅ 支持 | ✅ 支持 |
| **级联删除** | ✅ DB 自动 | ❌ 不提供 | ❌ 不提供 |
| **级联更新** | ✅ DB 自动 | ❌ 不提供 | ❌ 不提供 |
| **存在性校验** | ✅ DB 自动 | ✅ 应用层 | ❌ 不校验 |
| **跨库支持** | ❌ 不支持 | ✅ 支持 | ✅ 支持 |
| **性能** | 写入较慢 | 写入较快 | 最快 |
| **适用场景** | 单库、强一致性 | 微服务、灵活部署 | 日志、审计 |

**关键设计**：
- 逻辑外键**只提供关联查询能力和存在性校验**，不提供级联删除/更新
- 物理外键由数据库保证完整性和级联行为
- NONE 模式用于特殊场景（如审计日志，允许引用已删除的记录）
- 所有模式都支持关联查询，区别在于数据完整性保证方式

### 4.5 MANY_TO_ONE（多对一）

**场景**：多个订单属于一个客户

#### 4.5.1 关联到主键（最常见）

**用户定义**：

```yaml
model:
  name: order
  
fields:
  - name: id
    type: STRING
    format: UUID
    required: true
    unique: true
    
  # 关联字段：订单属于哪个客户
  - name: customer
    type: RELATION
    config:
      target_model: customer
      target_field: id          # ← 关联到 Customer.id
      type: MANY_TO_ONE
      foreign_key_mode: LOGICAL
    required: true
```

**系统创建的关联关系**：

```go
// 自动创建 ModelRelation
relation := &ModelRelation{
    ID:              "rel_xxx",
    SourceModelID:   "order_model_id",
    SourceFieldName: "customer",
    TargetModelID:   "customer_model_id",
    TargetFieldName: "id",         // ← 从 config.target_field 获取
    Type:            MANY_TO_ONE,
    ForeignKeyMode:  LOGICAL,
}

// Field 引用 Relation
field.RelationID = relation.ID
field.Relation = relation
```

**产物生成**：

```sql
-- 数据库表
CREATE TABLE orders (
  id VARCHAR(26) PRIMARY KEY,
  customer VARCHAR(26) NOT NULL,    -- 列名 = 字段名，类型从 target_field 推导
  INDEX idx_customer (customer)      -- 自动创建索引
);
-- 注意：LOGICAL 模式不生成 FOREIGN KEY 约束
```

**列类型推导规则**：
- Target 字段是 `STRING/UUID` → `VARCHAR(26)`
- Target 字段是 `STRING` → `VARCHAR(255)`
- Target 字段是 `NUMBER/INTEGER` → `INT`

```graphql
# GraphQL Schema
type Order {
  id: ID!
  customer: Customer!    # 关联对象（完整类型）
}

input OrderWhereInput {
  customer: CustomerWhereInput    # 支持嵌套过滤
}

input CreateOrderInput {
  customer: ID!          # 输入类型从 target_field 推导（UUID → ID）
}
```

**GraphQL 输入类型推导规则**：
- Target 字段是 `UUID` → `ID!`
- Target 字段是 `STRING` → `String!`
- Target 字段是 `NUMBER/INTEGER` → `Int!`

**应用层行为**：

```go
// 创建订单时校验 customer 存在（逻辑外键）
func CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, error) {
    // 1. 校验目标记录存在
    // 查询 Customer 模型，WHERE id = input.Customer
    exists, err := customerRepo.ExistsByField(ctx, "id", input.Customer)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, bizerrors.Errorf(
            "PARAM_INVALID.ORDER",
            "customer %s not found",
            input.Customer,
        )
    }
    
    // 2. 创建订单（存储外键值）
    return orderRepo.Create(ctx, input)
}

// 查询订单时加载关联对象
func GetOrder(ctx context.Context, id string) (*Order, error) {
    order, err := orderRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // 通过 target_field 加载关联对象
    // SELECT * FROM customers WHERE id = order.customer
    customer, err := customerRepo.FindByField(ctx, "id", order.Customer)
    if err != nil {
        return nil, err
    }
    
    order.CustomerObject = customer
    return order, nil
}
```// 删除客户时检查依赖
func DeleteCustomer(ctx context.Context, id string) (bool, error) {
    // 检查是否有订单引用
    hasOrders, err := orderRepo.ExistsByCustomer(ctx, id)
    if err != nil {
        return false, err
    }
    if hasOrders {
        return false, bizerrors.Errorf(
            "OPERATION_DENIED.CUSTOMER",
            "cannot delete customer: has orders",
        )
    }
    
    return customerRepo.Delete(ctx, id)
}
```

#### 4.5.2 使用物理外键

```yaml
fields:
  - name: customer
    type: RELATION
    config:
      target_model: customer
      target_field: id
      type: MANY_TO_ONE
      foreign_key_mode: PHYSICAL  # 物理外键
      on_delete: RESTRICT          # 禁止删除被引用的客户
      on_update: CASCADE           # 客户 ID 更新时同步更新
    required: true
```

**产物生成**：

```sql
CREATE TABLE orders (
  id VARCHAR(26) PRIMARY KEY,
  customer VARCHAR(26) NOT NULL,
  INDEX idx_customer (customer),
  
  -- 物理外键约束，引用 target_field 指定的列
  CONSTRAINT fk_order_customer
    FOREIGN KEY (customer) REFERENCES customers(id)
    --          ↑ 源列                        ↑ target_field
    ON DELETE RESTRICT
    ON UPDATE CASCADE
);
```

**行为差异**：
```go
// 创建订单 - 数据库自动校验
func CreateOrder(ctx context.Context, input CreateOrderInput) (*Order, error) {
    // 不需要手动校验，数据库会自动检查目标记录是否存在
    return orderRepo.Create(ctx, input)
    // 如果 customer 不存在，数据库返回外键约束错误
}

// 删除客户 - 数据库自动阻止
func DeleteCustomer(ctx context.Context, id string) (bool, error) {
    // 不需要手动检查，数据库会自动阻止
    return customerRepo.Delete(ctx, id)
    // 如果有订单引用，数据库返回外键约束错误
}
```

#### 4.5.3 关联到唯一字段（非主键）

**场景**：关联到商品 SKU（业务唯一标识）

```yaml
# Product 模型
model:
  name: product
fields:
  - name: id
    type: STRING
    format: UUID
    unique: true
    
  - name: sku
    type: STRING
    unique: true    # ← 唯一字段，可作为关联目标
    required: true
    
  - name: name
    type: STRING

# OrderItem 模型
model:
  name: order_item
fields:
  - name: product
    type: RELATION
    config:
      target_model: product
      target_field: sku      # ← 关联到 sku，而非 id
      type: MANY_TO_ONE
      foreign_key_mode: LOGICAL
```

**产物生成**：

```sql
-- Product 表
CREATE TABLE products (
  id VARCHAR(26) PRIMARY KEY,
  sku VARCHAR(255) NOT NULL,
  name VARCHAR(255),
  UNIQUE KEY uk_sku (sku)  -- sku 必须唯一
);

-- OrderItem 表
CREATE TABLE order_items (
  id VARCHAR(26) PRIMARY KEY,
  product VARCHAR(255) NOT NULL,  -- 列类型从 Product.sku 推导（VARCHAR）
  INDEX idx_product (product)
);
-- 注意：LOGICAL 模式不生成外键约束
```

```graphql
# GraphQL Schema
type OrderItem {
  id: ID!
  product: Product!    # 关联对象
}

input CreateOrderItemInput {
  product: String!     # 输入类型从 target_field 推导（sku 是 String）
}

# 查询示例
{
  orderItems {
    id
    product {      # 通过 sku 加载 Product 对象
      sku
      name
    }
  }
}
```

**应用层行为**：

```go
// 创建订单项时校验 product.sku 存在
func CreateOrderItem(ctx context.Context, input CreateOrderItemInput) (*OrderItem, error) {
    // 查询 Product 模型，WHERE sku = input.Product
    exists, err := productRepo.ExistsByField(ctx, "sku", input.Product)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, bizerrors.Errorf(
            "PARAM_INVALID.ORDER_ITEM",
            "product sku %s not found",
            input.Product,
        )
    }
    
    return orderItemRepo.Create(ctx, input)
}

// 查询时通过 sku 加载关联对象
func GetOrderItem(ctx context.Context, id string) (*OrderItem, error) {
    item, err := orderItemRepo.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // SELECT * FROM products WHERE sku = item.product
    product, err := productRepo.FindByField(ctx, "sku", item.Product)
    if err != nil {
        return nil, err
    }
    
    item.ProductObject = product
    return item, nil
}
```

### 4.6 ONE_TO_MANY 反向关联（推荐方案）

**推荐方案**：单向定义 + 自动推导反向字段

**场景**：Order 有多个 OrderItem

**只在外键方定义关联**：

```yaml
# OrderItem 模型（外键方）
model:
  name: order_item
fields:
  - name: id
    type: STRING
    format: UUID
    
  - name: order
    type: RELATION
    config:
      target_model: order
      target_field: id
      type: MANY_TO_ONE
      foreign_key_mode: LOGICAL
      # 配置反向字段
      reverse:
        field_name: items      # Order 模型的反向字段名
        exposed: true          # 是否在 GraphQL API 中暴露
```

**系统行为**：

1. **在 OrderItem 表中生成外键列**：
   ```sql
   CREATE TABLE order_items (
     id VARCHAR(26) PRIMARY KEY,
     order VARCHAR(26) NOT NULL,  -- 外键列
     INDEX idx_order (order)
   );
   ```

2. **在 Order 的 GraphQL 类型中自动生成虚拟字段**：
   ```graphql
   type Order {
     id: ID!
     items: [OrderItem!]!    # 虚拟字段（反向查询）
   }
   
   type OrderItem {
     id: ID!
     order: Order!           # 真实关联字段
   }
   ```

3. **不需要在 Order 模型中定义 items 字段**：
   - 用户不需要在 Order 模型中添加 items 字段
   - 系统根据 `reverse` 配置自动生成
   - 查询 `order.items` 时，系统自动执行 `WHERE order = order.id`

**查询示例**：

```graphql
# 正向查询（通过外键）
{
  orderItems {
    id
    order {        # 通过 order 字段加载 Order 对象
      id
      order_number
    }
  }
}

# 反向查询（自动生成的虚拟字段）
{
  orders {
    id
    order_number
    items {        # 自动执行 SELECT * FROM order_items WHERE order = ?
      id
      product
    }
  }
}
```

**优点**：
- ✅ 单一数据源（只在外键方定义）
- ✅ 符合物理真相（数据库只有外键）
- ✅ API 层自动生成双向关联
- ✅ 不会出现两边配置不一致的问题

### 4.7 字段推导规则

**关键设计**：用户只定义关联字段，系统自动推导外键列。

| 关联配置 | 数据库层 | GraphQL 层 | 说明 |
|---------|---------|-----------|------|
| `name: customer` | `customer` 列 | `customer: Customer!` | 列名 = 字段名 |
| `target_field: id` | `FOREIGN KEY REFERENCES customers(id)` | `input: ID!` | 引用目标列，输入类型从目标字段推导 |
| `target_field: sku` | `FOREIGN KEY REFERENCES products(sku)` | `input: String!` | 引用 sku 列，输入类型为 String |
| `foreign_key_mode: LOGICAL` | 无 FK 约束 | 应用层校验 | 不生成数据库约束 |
| `foreign_key_mode: PHYSICAL` | FK 约束 + 级联 | 数据库保证 | 生成完整外键约束 |
| `reverse: {field_name: items}` | 无物理字段 | `items: [OrderItem!]!` | 虚拟字段，反向查询 |

**外键列推导规则**：

```
字段名 → 列名 → 列类型

Order.customer (RELATION)
  ↓
customer (列名 = 字段名)
  ↓
VARCHAR(26) (从 Customer.id 的类型推导)
  ↓
REFERENCES customers(id) (target_field 指定的列)
```

**GraphQL 输入类型推导规则**：

```
target_field 类型 → GraphQL 输入类型

Customer.id (STRING/UUID) → ID!
Product.sku (STRING)      → String!
Category.code (STRING)    → String!
Status.value (NUMBER)     → Int!
```

**示例**：

```yaml
# 用户定义
- name: customer
  type: RELATION
  config:
    target_model: customer
    target_field: id          # ← 关键
    type: MANY_TO_ONE
    foreign_key_mode: LOGICAL
```

**系统推导**：

```sql
-- 数据库：推导外键列
CREATE TABLE orders (
  customer VARCHAR(26) NOT NULL,    -- 列名 = 字段名，类型从 target_field 推导
  INDEX idx_customer (customer)
);
-- LOGICAL 模式：不生成 FOREIGN KEY 约束
```

```graphql
# GraphQL：推导输入类型
input CreateOrderInput {
  customer: ID!    # 从 Customer.id (UUID) 推导为 ID
}

type Order {
  customer: Customer!    # 关联对象类型
}
```

**复杂示例**：

```yaml
# 关联到非主键唯一字段
- name: product
  type: RELATION
  config:
    target_model: product
    target_field: sku         # ← 关联到 sku（String, Unique）
    type: MANY_TO_ONE
    foreign_key_mode: PHYSICAL
    on_delete: RESTRICT
```

**推导结果**：

```sql
CREATE TABLE order_items (
  product VARCHAR(255) NOT NULL,    -- 从 Product.sku (String) 推导
  
  -- 物理外键：引用 sku 列
  CONSTRAINT fk_order_item_product
    FOREIGN KEY (product) REFERENCES products(sku)
    ON DELETE RESTRICT
);
```

```graphql
input CreateOrderItemInput {
  product: String!    # 从 Product.sku (String) 推导
}
```

---

## 5. 产物生成规则

### 5.1 产物分层

```
┌─────────────────────────────────────────────┐
│   语义层 (Design-Time Schema)               │
│   - 模型定义                                │
│   - 字段定义                                │
│   - 关联配置                                │
├─────────────────────────────────────────────┤
                  ↓ 编译/生成
┌─────────────────────────────────────────────┐
│   产物层 (Runtime Artifacts)                │
│   ├── HTTP API Layer                       │
│   │   └── GraphQL Schema                   │
│   ├── Storage Layer                        │
│   │   └── Database Schema (SQL)            │
│   └── Runtime Layer                        │
│       ├── Resolver Code                    │
│       ├── Validator Code                   │
│       └── Cache Strategy                   │
└─────────────────────────────────────────────┘
```

### 5.2 GraphQL Schema 生成

#### 5.2.1 类型生成规则

**Regular Model** → `type`

```yaml
# 输入
model:
  name: order
  type: REGULAR
fields:
  - name: id
    type: STRING
    format: UUID
  - name: order_number
    type: STRING
  - name: total
    type: NUMBER
    format: DECIMAL
```

```graphql
# 输出
type Order {
  id: ID!
  order_number: String!
  total: Decimal!
}
```

**Enum Model** → `type` (不是 `enum`)

```yaml
# 输入
model:
  name: order_status
  type: ENUM
  enum_config:
    code_field: code
    label_field: label
fields:
  - name: code
    type: STRING
  - name: label
    type: STRING
  - name: color
    type: STRING
```

```graphql
# 输出
type OrderStatus {
  code: String!
  label: String!
  color: String
}
```

**关键**：Enum Model 生成的是 `type`，不是 GraphQL 原生 `enum`，因为：
- 支持扩展字段（color, icon, description）
- 运行时可配置
- 统一的 CRUD API

#### 5.2.2 字段类型映射

| 语义类型 | Format | GraphQL 类型 | 说明 |
|---------|--------|-------------|------|
| STRING | - | String | 普通字符串 |
| STRING | UUID | ID | UUID 映射为 ID |
| NUMBER | INTEGER | Int | 整数 |
| NUMBER | DECIMAL | Decimal | 自定义 Scalar |
| NUMBER | - (FLOAT) | Float | 浮点数 |
| BOOLEAN | - | Boolean | 布尔值 |
| DATETIME | DATE | Date | 自定义 Scalar |
| DATETIME | TIME | Time | 自定义 Scalar |
| DATETIME | DATETIME | DateTime | 自定义 Scalar |
| JSON | - | JSON | 自定义 Scalar |
| ENUM | - | Enum | GraphQL 原生 enum 类型 |
| RELATION | - | TargetType | 关联对象类型 |

**自定义 Scalar 定义**：

```graphql
scalar Decimal
scalar Date
scalar Time
scalar DateTime
scalar JSON
```

**枚举类型生成**：

```yaml
# 输入：枚举字段
- name: status
  type: ENUM
  enum_name: order_status

# 枚举定义（系统配置）
enum_definition:
  name: order_status
  options:
    - code: PENDING, label: 待支付
    - code: PAID, label: 已支付
```

```graphql
# 输出：GraphQL 原生 enum
enum OrderStatus {
  PENDING
  PAID
  SHIPPED
  DELIVERED
}

type Order {
  status: OrderStatus!
}
```

#### 5.2.3 Query 入口生成

**单条查询**：
```graphql
type Query {
  order(id: ID!): Order
}
```

**列表查询**：
```graphql
type Query {
  orders(
    where: OrderWhereInput
    orderBy: [OrderOrderByInput!]
    skip: Int
    take: Int
  ): [Order!]!
}
```

**聚合查询**：
```graphql
type Query {
  ordersAggregate(where: OrderWhereInput): OrderAggregateResult!
}

type OrderAggregateResult {
  count: Int!
  sum: OrderSumAggregate
  avg: OrderAvgAggregate
  min: OrderMinAggregate
  max: OrderMaxAggregate
}
```

#### 5.2.4 Mutation 入口生成

**创建**：
```graphql
type Mutation {
  createOrder(input: CreateOrderInput!): Order!
}

input CreateOrderInput {
  order_number: String!
  total: Decimal!
  status: OrderStatus!  # 枚举字段 → 枚举值
  customer: ID!         # 关联字段 → 外键 ID
  # 注意：UUID 字段（如 id）不包含在 Input 中
}
```

**更新**：
```graphql
type Mutation {
  updateOrder(id: ID!, input: UpdateOrderInput!): Order!
}

input UpdateOrderInput {
  order_number: String
  total: Decimal
  customer: ID
  # 所有字段可选
  # UUID 字段不可更新
}
```

**删除**：
```graphql
type Mutation {
  deleteOrder(id: ID!): Boolean!
}
```

#### 5.2.5 过滤输入生成

```graphql
input OrderWhereInput {
  AND: [OrderWhereInput!]
  OR: [OrderWhereInput!]
  NOT: OrderWhereInput
  
  id: IDFilter
  order_number: StringFilter
  total: DecimalFilter
  status: OrderStatusFilter     # 枚举过滤
  customer: CustomerWhereInput  # 关联嵌套过滤
}

input OrderStatusFilter {
  equals: OrderStatus
  not: OrderStatus
  in: [OrderStatus!]
  notIn: [OrderStatus!]
}

input StringFilter {
  equals: String
  not: String
  in: [String!]
  notIn: [String!]
  contains: String
  startsWith: String
  endsWith: String
}

input DecimalFilter {
  equals: Decimal
  not: Decimal
  in: [Decimal!]
  notIn: [Decimal!]
  lt: Decimal
  lte: Decimal
  gt: Decimal
  gte: Decimal
}
```

### 5.3 Database Schema 生成

#### 5.3.1 表名规则

```
模型名 → 表名（复数形式 + snake_case）

order       → orders
order_item  → order_items
user        → users
```

#### 5.3.2 列类型映射

| 语义类型 | Format | MySQL 类型 | PostgreSQL 类型 |
|---------|--------|-----------|----------------|
| STRING | - | VARCHAR(255) | VARCHAR(255) |
| STRING | UUID | VARCHAR(26) | VARCHAR(26) |
| NUMBER | INTEGER | INT | INTEGER |
| NUMBER | DECIMAL | DECIMAL(p,s) | NUMERIC(p,s) |
| NUMBER | - (FLOAT) | DOUBLE | DOUBLE PRECISION |
| BOOLEAN | - | BOOLEAN | BOOLEAN |
| DATETIME | DATE | DATE | DATE |
| DATETIME | TIME | TIME | TIME |
| DATETIME | DATETIME | DATETIME | TIMESTAMP |
| JSON | - | JSON | JSONB |
| ENUM | - | VARCHAR(50) | VARCHAR(50) |
| RELATION | MANY_TO_ONE | VARCHAR(26) | VARCHAR(26) |

**配置参数**：
- STRING 默认长度：255（可配置）
- DECIMAL 默认：precision=10, scale=2（可配置）
- UUID 固定：26 字符（UUID v7）
- ENUM 默认长度：50（存储 code 值）

#### 5.3.3 表结构生成

```yaml
# 输入
model:
  name: order
fields:
  - name: id
    type: STRING
    format: UUID
    required: true
    unique: true
    
  - name: order_number
    type: STRING
    required: true
    unique: true
    indexed: true
    
  - name: total
    type: NUMBER
    format: DECIMAL
    required: true
    config:
      precision: 10
      scale: 2
  
  - name: status
    type: ENUM
    enum_name: order_status
    required: true
      
  - name: customer
    type: RELATION
    config:
      target_model: customer
      type: MANY_TO_ONE
      foreign_key: LOGICAL
    required: true
    
  - name: created_at
    type: DATETIME
    format: DATETIME
    config:
      auto_now_add: true
```

```sql
-- 输出（MySQL）
CREATE TABLE orders (
  id VARCHAR(26) PRIMARY KEY,
  order_number VARCHAR(255) NOT NULL,
  total DECIMAL(10,2) NOT NULL,
  status VARCHAR(50) NOT NULL,
  customer VARCHAR(26) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE KEY uk_order_number (order_number),
  INDEX idx_order_number (order_number),
  INDEX idx_status (status),
  INDEX idx_customer (customer)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

#### 5.3.4 外键约束生成

**LOGICAL 模式**：不生成外键约束

```sql
-- 只有索引
INDEX idx_customer (customer)
```

**PHYSICAL 模式**：生成外键约束

```yaml
# 输入
- name: customer
  type: RELATION
  config:
    foreign_key: PHYSICAL
    on_delete: RESTRICT
    on_update: CASCADE
```

```sql
-- 输出
INDEX idx_customer (customer),
CONSTRAINT fk_order_customer
  FOREIGN KEY (customer) REFERENCES customers(id)
  ON DELETE RESTRICT
  ON UPDATE CASCADE
```

#### 5.3.5 索引生成规则

**自动创建索引的情况**：
1. `unique: true` → UNIQUE KEY
2. `indexed: true` → INDEX
3. `type: RELATION` → INDEX（外键列自动索引）
4. `format: UUID` + `unique: true` → PRIMARY KEY

```yaml
# 输入
fields:
  - name: id
    type: STRING
    format: UUID
    unique: true              # → PRIMARY KEY
    
  - name: email
    type: STRING
    unique: true              # → UNIQUE KEY
    indexed: true             # unique 已包含索引
    
  - name: status_code
    type: STRING
    indexed: true             # → INDEX
    
  - name: customer
    type: RELATION            # → INDEX（自动）
```

```sql
-- 输出
PRIMARY KEY (id),
UNIQUE KEY uk_email (email),
INDEX idx_status_code (status_code),
INDEX idx_customer (customer)
```

### 5.4 Runtime Code 生成

#### 5.4.1 Resolver 生成

**Query Resolver**：

```go
// 自动生成
func (r *QueryResolver) Order(ctx context.Context, id string) (*Order, error) {
    return r.orderService.FindByID(ctx, id)
}

func (r *QueryResolver) Orders(
    ctx context.Context,
    where *OrderWhereInput,
    orderBy []*OrderOrderByInput,
    skip *int,
    take *int,
) ([]*Order, error) {
    return r.orderService.FindMany(ctx, where, orderBy, skip, take)
}
```

**Mutation Resolver**：

```go
// 自动生成
func (r *MutationResolver) CreateOrder(
    ctx context.Context,
    input CreateOrderInput,
) (*Order, error) {
    // 1. UUID 字段自动生成
    input.ID = generateUUIDV7()
    
    // 2. 逻辑外键校验（如果是 LOGICAL 模式）
    if err := r.validateForeignKeys(ctx, input); err != nil {
        return nil, err
    }
    
    // 3. 创建记录
    return r.orderService.Create(ctx, input)
}
```

#### 5.4.2 枚举校验生成

```go
// 自动生成（ENUM 类型字段）
func (r *MutationResolver) validateEnumFields(
    ctx context.Context,
    input CreateOrderInput,
) error {
    // 校验 status 字段（枚举值）
    enumDef, err := r.enumService.GetEnumDefinition(ctx, "order_status")
    if err != nil {
        return err
    }
    
    if !enumDef.HasOptionCode(input.Status) {
        return bizerrors.Errorf(
            "PARAM_INVALID.ORDER",
            "invalid status value: %s",
            input.Status,
        )
    }
    
    return nil
}
```

#### 5.4.3 逻辑外键校验生成

```go
// 自动生成（LOGICAL 模式）
func (r *MutationResolver) validateForeignKeys(
    ctx context.Context,
    input CreateOrderInput,
) error {
    // 根据 RelationConfig 自动生成校验逻辑
    // 校验 customer 字段（关联到 Customer.id）
    if input.Customer != nil {
        // 查询目标模型的 target_field
        // SELECT COUNT(*) FROM customers WHERE id = input.Customer
        exists, err := r.customerService.ExistsByField(
            ctx, 
            "id",              // ← 从 RelationConfig.TargetField 获取
            *input.Customer,
        )
        if err != nil {
            return err
        }
        if !exists {
            return bizerrors.Errorf(
                "PARAM_INVALID.ORDER",
                "customer %s not found",
                *input.Customer,
            )
        }
    }
    
    // 如果关联到非主键字段（如 sku）
    // SELECT COUNT(*) FROM products WHERE sku = input.Product
    if input.Product != nil {
        exists, err := r.productService.ExistsByField(
            ctx,
            "sku",             // ← target_field = sku
            *input.Product,
        )
        if err != nil {
            return err
        }
        if !exists {
            return bizerrors.Errorf(
                "PARAM_INVALID.ORDER_ITEM",
                "product sku %s not found",
                *input.Product,
            )
        }
    }
    
    return nil
}
```

### 5.5 特殊字段处理

#### 5.5.1 UUID 字段

```yaml
- name: id
  type: STRING
  format: UUID
  required: true
  unique: true
```

**生成行为**：
1. **Database**: `VARCHAR(26) PRIMARY KEY`
2. **GraphQL Input**: 不包含在 `CreateXxxInput` 和 `UpdateXxxInput` 中
3. **Runtime**: 创建时自动生成 UUID v7
4. **不可更新**: Update 操作忽略此字段

#### 5.5.2 自动时间戳字段

```yaml
- name: created_at
  type: DATETIME
  format: DATETIME
  config:
    auto_now_add: true

- name: updated_at
  type: DATETIME
  format: DATETIME
  config:
    auto_now: true
```

**生成行为**：
1. **Database**: 
   ```sql
   created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
   ```
2. **GraphQL Input**: 不包含在 Input 中
3. **Runtime**: 自动设置时间

#### 5.5.3 枚举字段

```yaml
- name: status
  type: ENUM
  enum_name: order_status
  required: true
```

**生成行为**：
1. **Database**: 
   ```sql
   status VARCHAR(50) NOT NULL,
   INDEX idx_status (status)
   ```
2. **GraphQL Type**: 
   ```graphql
   type Order {
     status: OrderStatus!
   }
   
   enum OrderStatus {
     PENDING
     PAID
     SHIPPED
   }
   ```
3. **GraphQL Input**: 
   ```graphql
   input CreateOrderInput {
     status: OrderStatus!    # 枚举值
   }
   ```
4. **Runtime**: 
   - 存储 code 值（字符串）
   - 创建/更新时校验值是否在枚举定义中
   - 不需要 JOIN 查询

#### 5.5.4 关联字段

```yaml
- name: customer
  type: RELATION
  config:
    target_model: customer
    target_field: id        # ← 必须指定
    type: MANY_TO_ONE
    foreign_key_mode: LOGICAL
```

**生成行为**：
1. **Database**: 
   ```sql
   customer VARCHAR(26) NOT NULL,  -- 列类型从 Customer.id 推导
   INDEX idx_customer (customer)
   ```
2. **GraphQL Type**: 
   ```graphql
   type Order {
     customer: Customer!    # 关联对象类型
   }
   ```
3. **GraphQL Input**: 
   ```graphql
   input CreateOrderInput {
     customer: ID!    # 输入类型从 target_field (UUID) 推导
   }
   ```
4. **Runtime**: 
   - 存储外键值到 `customer` 列
   - 查询时通过 `target_field` 加载关联对象：`WHERE id = customer`
   - 支持嵌套过滤

**关键点**：
- 外键列名 = 字段名（`customer`）
- 外键列类型从 `target_field` 的类型推导（UUID → VARCHAR(26)）
- GraphQL 输入类型从 `target_field` 推导（UUID → ID!）
- 查询时使用 `target_field` 作为 WHERE 条件

### 5.6 产物生成流程

```
┌─────────────────────────────────────────────┐
│  1. 解析模型定义                             │
│     - 读取 YAML/JSON                        │
│     - 验证模型和字段                         │
│     - 构建内存模型                           │
├─────────────────────────────────────────────┤
│  2. 推导物理结构                             │
│     - 关联字段 → 创建 ModelRelation         │
│     - 关联关系 → 外键列                      │
│     - target_field → 列类型和输入类型       │
│     - 索引推导                               │
├─────────────────────────────────────────────┤
│  3. 生成 GraphQL Schema                     │
│     - Type 定义                             │
│     - Query/Mutation 入口                   │
│     - Input 类型                            │
│     - Filter 类型                           │
├─────────────────────────────────────────────┤
│  4. 生成 Database Schema                    │
│     - CREATE TABLE                          │
│     - INDEX                                 │
│     - FOREIGN KEY (if PHYSICAL)             │
├─────────────────────────────────────────────┤
│  5. 生成 Runtime Code                       │
│     - Resolver                              │
│     - Validator (枚举 + LOGICAL 外键)       │
└─────────────────────────────────────────────┘
```

---

## 附录 A：关联关系完整设计

### A.1 关联关系的三层结构

```
┌─────────────────────────────────────────────┐
│  用户定义层（语义）                          │
│  - FieldDefinition.Type = RELATION          │
│  - RelationConfig:                          │
│    - target_model: "customer"               │
│    - target_field: "id"        ← 核心配置   │
├─────────────────────────────────────────────┤
│  系统管理层（关联关系）                      │
│  - ModelRelation（自动创建）:               │
│    - source: Order.customer                 │
│    - target: Customer.id       ← 完整映射   │
│    - type: MANY_TO_ONE                      │
│    - foreign_key_mode: LOGICAL              │
├─────────────────────────────────────────────┤
│  产物生成层（物理存储）                      │
│  - Column: customer VARCHAR(26)             │
│  - Index: idx_customer                      │
│  - FK Constraint (可选)                     │
│  - GraphQL Input: customer: ID!             │
└─────────────────────────────────────────────┘
```

### A.2 target_field 的作用总结

| 作用 | 说明 | 示例 |
|------|------|------|
| **明确关联映射** | 指定关联到目标模型的哪个字段 | `Order.customer → Customer.id` |
| **支持非主键关联** | 可关联到任何唯一字段 | `OrderItem.product → Product.sku` |
| **决定列类型** | 外键列类型从 target_field 推导 | `sku (String)` → `VARCHAR(255)` |
| **决定输入类型** | GraphQL 输入从 target_field 推导 | `id (UUID)` → `ID!` |
| **生成外键约束** | 物理外键引用的目标列 | `REFERENCES customers(id)` |
| **校验逻辑** | 逻辑外键查询的字段 | `WHERE id = ?` |

### A.3 推导规则汇总

**数据库列类型推导**：
```
target_field 类型        → 外键列类型
STRING/UUID             → VARCHAR(26)
STRING                  → VARCHAR(255)
NUMBER/INTEGER          → INT
NUMBER/DECIMAL          → DECIMAL(p,s)
```

**GraphQL 输入类型推导**：
```
target_field 类型        → GraphQL 输入类型
STRING/UUID             → ID!
STRING                  → String!
NUMBER/INTEGER          → Int!
NUMBER/DECIMAL          → Float!
```

**物理外键约束**：
```sql
FOREIGN KEY (字段名) REFERENCES 目标表(target_field)
--           ↑                    ↑       ↑
--     关联字段名            目标表名   目标列名
```

**逻辑外键校验**：
```go
// 查询目标记录是否存在
SELECT COUNT(*) FROM 目标表 WHERE target_field = 输入值
```

---

## 附录 B：枚举系统设计

### 当前系统的枚举实现

ModelCraft 采用**配置化枚举系统**，枚举定义存储在系统数据库中，作为项目级配置管理。

**核心组件**：

1. **EnumDefinition**（枚举定义）
   - 存储位置：系统数据库（非用户数据库）
   - 作用域：项目级（OrgName + ProjectSlug）
   - 包含：code/label 映射 + 选项列表

2. **FieldEnumAssociation**（字段枚举关联）
   - 管理字段与枚举定义的关联关系
   - 多对一关系：多个字段可引用同一枚举定义

3. **EnumOption**（枚举选项）
   - Code: 存储在数据库中的值
   - Label: 显示给用户的文本
   - Order: 排序
   - Description: 选项描述

**关键特性**：
- ✅ 枚举定义通过管理界面配置
- ✅ 支持单选和多选模式
- ✅ 字段通过 `EnumName` 引用枚举定义
- ❌ 不允许运行时动态增删（需通过配置界面）
- ❌ 不生成独立的 CRUD API

**与字典表的区别**：

| 维度 | 枚举（ENUM） | 字典表（Regular Model） |
|------|-------------|----------------------|
| 本质 | 系统配置 | 业务数据 |
| 存储 | 系统数据库 | 用户数据库 |
| API | 无独立 CRUD | 完整 CRUD |
| 引用方式 | 直接存储 code | 存储外键 |
| 查询性能 | 高（无 JOIN） | 需要 JOIN |
| 扩展性 | 固定结构 | 任意字段 |

---

## 附录 C：ONE_TO_MANY 方案对比（已确定）

**推荐方案**：单向定义 + 自动推导反向字段（参见 4.6 节）

### 已采用的设计

- ✅ 只在外键方定义关联字段（MANY_TO_ONE）
- ✅ 配置 `reverse` 参数指定反向字段名
- ✅ 系统自动在目标模型生成虚拟字段
- ✅ 符合物理真相（数据库只有外键）
- ✅ 避免双向定义的一致性问题

### 不采用的方案

❌ **方案 A：双向定义**
- 需要在两个模型中都定义字段
- 容易出现配置不一致
- 增加用户心智负担

❌ **方案 B：只保留 MANY_TO_ONE**
- 需要用户手动编写反向查询
- API 体验不佳

---

**文档版本**: v3.0  
**最后更新**: 2026-03-09  
**变更说明**: 
- v3.0: 更新关联关系设计，强调 target_field 的作用
- v2.0: 移除 Enum Model，新增 ENUM 字段类型
- v1.0: 初始版本
