## Context

当前系统使用 `field_definitions.enum_name` 字段直接存储枚举的业务标识(business name)来关联枚举定义。这种设计存在以下问题:

1. **缺乏外键完整性**: `enum_name` 是业务标识而非主键,无法使用数据库外键约束
2. **引用追踪困难**: 无法高效查询某个枚举被哪些字段使用
3. **级联删除缺失**: 删除字段时需要手动清理枚举关联
4. **GraphQL API 设计混乱**: 枚举配置混在 `validationConfig` 中,职责不清晰

类比系统中 `model_relations` 表的设计,关系数据应该独立管理。本次重构引入专门的关联表来管理模型字段与枚举之间的多对一关系。

**约束条件**:
- 必须保持向后兼容 `ValidationConfig.EnumValues` 简单枚举功能
- 数据库迁移由现有脚本处理,本次变更不包含迁移实现
- 需要支持两种枚举使用模式:关联现有枚举 vs 创建新枚举

**相关利益方**:
- 前端开发团队: 需要适配新的 GraphQL API
- 数据库管理员: 需要执行数据迁移脚本

## Goals / Non-Goals

**Goals**:
- 使用关联表 `model_field_enum_associations` 管理字段-枚举关系
- 通过外键约束保证引用完整性
- 支持级联删除字段时自动清理关联
- 提供清晰的 GraphQL API (`enumConfig`) 区分枚举配置和字段验证配置
- 支持"关联现有枚举"和"创建新枚举"两种模式

**Non-Goals**:
- 不修改 `ValidationConfig.EnumValues` 简单枚举功能
- 不包含数据库迁移脚本实现(由现有工具处理)
- 不修改枚举定义本身的数据结构
- 不改变枚举的存储格式(物理字段存key,逻辑字段返回完整信息)

## Decisions

### Decision 1: 使用独立关联表而非嵌入字段

**决策**: 创建 `model_field_enum_associations` 表记录字段-枚举关联,而非在 `field_definitions` 表中直接存储 `enum_name`

**理由**:
- 类比 `model_relations` 表的设计模式,关系数据应独立管理
- 支持数据库级别的外键约束和引用完整性
- 方便查询某个枚举被哪些字段使用(反向查询)
- 支持级联删除: 删除字段自动删除关联,删除枚举前检查关联

**替代方案**:
- 方案A: 保留 `field_definitions.enum_name`,增加唯一索引
  - 缺点: 无法使用外键约束(name是业务标识非主键)
  - 缺点: 反向查询效率低(需要扫描所有字段)
- 方案B: 改为存储 `enum_id` (外键)
  - 缺点: 破坏现有导入导出逻辑(依赖name而非ID)
  - 缺点: 跨环境迁移困难(ID不同)

### Decision 2: GraphQL API 使用 enumConfig 替代 validationConfig.enum

**决策**: 新增 `AddFieldInput.enumConfig` 输入类型,移除 `ValidationConfigInput.enum` 字段

**理由**:
- 职责分离: 枚举配置与字段验证配置是不同的关注点
- API 清晰性: `enumConfig` 明确表达是枚举相关配置
- 扩展性: 枚举配置可以独立演进,不影响 `ValidationConfig` 结构
- 一致性: 与 `relationConfig` 的设计模式保持一致

**替代方案**:
- 方案A: 保留在 `validationConfig` 中,增加 `enumName` 字段
  - 缺点: 职责混乱,验证配置不应包含关联配置
  - 缺点: 难以扩展(如未来支持枚举权限等)
- 方案B: 使用两个独立的 mutation (connectEnum vs createEnum)
  - 缺点: API 复杂度增加
  - 缺点: 客户端需要两次调用

### Decision 3: 使用 connectEnum 布尔值区分创建vs关联

**决策**: `enumConfig` 包含 `connectEnum: Boolean!` 字段,true表示关联现有枚举,false表示创建新枚举

**理由**:
- 单一 API 入口,降低客户端复杂度
- 语义清晰: `connectEnum` 明确表达操作意图
- 与字段关系配置保持一致(未来可能有类似需求)

**替代方案**:
- 方案A: 根据是否提供 `options` 自动判断
  - 缺点: 隐式行为,容易产生歧义
  - 缺点: 无法明确区分用户意图(可能是误操作)
- 方案B: 使用 `union` 类型(ConnectEnumConfig | CreateEnumConfig)
  - 缺点: GraphQL union 输入类型支持有限
  - 缺点: 客户端处理复杂

### Decision 4: 外键关联使用 enum_id 而非 enum_name

**决策**: `model_field_enum_associations` 表存储 `enum_id` 外键关联 `model_enums.id`

**理由**:
- 数据库外键约束要求关联主键
- 防止删除被引用的枚举(RESTRICT)
- 性能优化: ID 索引比 name 索引效率高

**导入导出处理**:
- 导出时: 通过 JOIN 查询获取 enum_name
- 导入时: 先按 name 查找或创建枚举,再创建关联

## Schema Design

### 数据库表结构

```sql
CREATE TABLE IF NOT EXISTS `model_field_enum_associations` (
  `model_id` VARCHAR(36) NOT NULL COMMENT '模型ID',
  `field_name` VARCHAR(64) NOT NULL COMMENT '字段名称',
  `enum_id` VARCHAR(36) NOT NULL COMMENT '枚举ID',
  `cluster_name` VARCHAR(64) NOT NULL COMMENT '集群名称',
  `database_name` VARCHAR(64) NOT NULL COMMENT '数据库名称',
  `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

  PRIMARY KEY (`model_id`, `field_name`),

  CONSTRAINT `fk_field_enum_assoc_field`
    FOREIGN KEY (`model_id`, `field_name`)
    REFERENCES `field_definitions` (`model_id`, `name`)
    ON DELETE CASCADE,

  CONSTRAINT `fk_field_enum_assoc_enum`
    FOREIGN KEY (`enum_id`)
    REFERENCES `model_enums` (`id`)
    ON DELETE RESTRICT,

  KEY `idx_enum_id` (`enum_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**外键说明**:
- `ON DELETE CASCADE`: 删除字段时自动删除关联
- `ON DELETE RESTRICT`: 删除枚举时检查是否被引用

### GraphQL Schema

```graphql
input EnumOptionInput {
  key: String!
  value: String!
  order: Int!
  description: String
}

input EnumConfigInput {
  enumName: String!
  options: [EnumOptionInput!]
  description: String
  connectEnum: Boolean!
}

input AddFieldInput {
  name: String!
  title: String!
  format: FormatType!
  storageHint: String
  nonNull: Boolean = false
  required: Boolean = false
  description: String
  validationConfig: ValidationConfigInput
  relationConfig: RelationConfigInput
  enumConfig: EnumConfigInput  # 新增
}
```

## Implementation Flow

### 创建枚举字段流程

```
1. GraphQL Resolver 接收 AddFieldInput
   └─> 检查 format 是否为 ENUM/ENUM_ARRAY
       └─> 检查 enumConfig 是否存在

2. 如果 enumConfig.connectEnum = true
   ├─> EnumService.FindByName(enumConfig.enumName)
   ├─> 如果不存在 → 返回错误 "enum not found"
   └─> 如果存在 → 获取 enum.ID

3. 如果 enumConfig.connectEnum = false
   ├─> 验证 enumConfig.options 必须非空
   ├─> EnumService.Create(enumName, options, description)
   ├─> 如果名称冲突 → 返回错误 "enum name already exists"
   └─> 如果成功 → 获取新创建的 enum.ID

4. FieldService.CreateField(fieldDef)
   └─> 创建字段定义

5. FieldEnumAssociationRepository.Create(association)
   └─> 创建关联记录 (model_id, field_name, enum_id)

6. Transaction.Commit()
```

### 删除字段流程

```
1. FieldService.DeleteField(modelID, fieldName)
   └─> 删除字段定义

2. 数据库自动触发
   └─> CASCADE DELETE 删除 model_field_enum_associations 记录
```

### 删除枚举流程

```
1. EnumService.DeleteEnum(enumID)
   ├─> FieldEnumAssociationRepository.FindByEnumID(enumID)
   ├─> 如果有关联 → 返回错误 "enum is referenced by fields: [list]"
   └─> 如果无关联 → 删除枚举定义
```

### 查询字段流程

```
1. GraphQL Query 查询 Field
   └─> FieldRepository.FindByModelID(modelID)
       └─> 返回 FieldDefinition 列表

2. Resolver 加载枚举
   ├─> FieldEnumAssociationRepository.FindByField(modelID, fieldName)
   ├─> 如果有关联 → EnumRepository.FindByID(enumID)
   └─> 填充 FieldDefinition.Enum 字段
```

## Risks / Trade-offs

### Risk 1: 数据迁移失败风险

**风险**: 现有 `field_definitions.enum_name` 数据迁移到新表时可能失败

**影响**: 导致字段-枚举关联丢失,影响业务功能

**缓解措施**:
- 迁移脚本需要验证所有 enum_name 都能找到对应的枚举ID
- 对于找不到的枚举,记录日志但不中断迁移
- 提供回滚脚本
- 在测试环境充分验证

### Risk 2: GraphQL API 破坏性变更

**风险**: 移除 `validationConfig.enum` 导致现有客户端无法工作

**影响**: 前端应用无法创建枚举字段

**缓解措施**:
- 在 CHANGELOG 中明确标注 BREAKING CHANGE
- 提供迁移指南和示例代码
- 如果需要,可以考虑保留 `validationConfig.enum` 一个版本作为过渡(标记为 deprecated)
- 与前端团队提前沟通变更

### Risk 3: 性能影响

**风险**: 查询字段时需要 JOIN 关联表和枚举表,可能影响性能

**影响**: 字段列表查询变慢

**缓解措施**:
- 在 `model_field_enum_associations.enum_id` 上创建索引
- 使用 DataLoader 批量加载枚举(避免 N+1 查询)
- 如果性能问题明显,考虑在字段表增加冗余的 `enum_name` 字段(只读)

### Trade-off: 存储冗余 vs 引用完整性

**选择**: 优先保证引用完整性,接受关联表的存储开销

**理由**:
- 关联表记录很小,存储成本可接受
- 外键约束带来的数据一致性收益远大于存储成本
- 便于未来扩展(如添加关联元数据)

## Migration Plan

### 步骤1: 数据库变更(由现有脚本处理)

- 创建 `model_field_enum_associations` 表
- 迁移 `field_definitions.enum_name` 数据到新表
- 删除 `field_definitions.enum_name` 列

### 步骤2: 代码变更(本次实施)

- Domain Layer: 移除 EnumName,增加 FieldEnumAssociation
- Infrastructure: 实现关联表仓储
- Application: 调整字段和枚举服务逻辑
- GraphQL: 更新 schema 和 resolvers

### 步骤3: 部署和验证

1. 部署后端服务(包含新旧 API)
2. 前端切换到新 API (`enumConfig`)
3. 监控日志和错误率
4. 验证核心功能(创建/删除/查询枚举字段)

### 回滚策略

如果部署后发现严重问题:

1. 回滚代码到上一版本
2. 如果数据库已变更,执行回滚脚本:
   - 恢复 `field_definitions.enum_name` 列
   - 从 `model_field_enum_associations` 回填数据
   - 删除 `model_field_enum_associations` 表

**注意**: 回滚窗口应在数据库变更后24小时内

## Open Questions

1. **是否需要在过渡期保留 `validationConfig.enum` 支持?**
   - 如果前端团队需要更长的适配时间,可以考虑保留一个版本
   - 标记为 `@deprecated` 并在文档中说明替代方案

2. **导入导出时枚举名称冲突如何处理?**
   - 当前设计: 导入时如果枚举名称已存在,报错
   - 是否需要支持自动重命名?(如 UserStatus -> UserStatus_2)
   - 建议: 先实现严格检查,根据用户反馈决定是否增加灵活性

3. **是否需要支持字段批量修改枚举关联?**
   - 当前设计: 只在创建字段时设置枚举关联
   - 未来可能需要 `updateFieldEnumAssociation` mutation
   - 建议: 本次不实现,等待实际需求
