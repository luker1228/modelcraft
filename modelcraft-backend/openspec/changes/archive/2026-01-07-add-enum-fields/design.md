# 枚举字段设计文档

## Context

系统需要支持枚举字段类型,包括单选和多选场景。参考关联关系的设计模式,枚举定义作为独立的元数据资源进行中心化管理,字段通过引用方式关联枚举定义。

### 现有相关设计
- **关联关系设计**: 字段通过 `RelationId` 引用 `ModelRelation`,关联数据存储在独立的 `model_relations` 表
- **字段类型系统**: 基于 `FormatType` 的字段类型体系,支持 STRING、NUMBER、DATE、DATETIME 等格式
- **验证配置**: 通过 `ValidationConfig` 结构存储字段级别的验证规则

## Goals / Non-Goals

### Goals
- 建立中心化的枚举元数据管理机制
- 支持枚举的复用和统一维护
- 支持单选和多选两种枚举模式
- 提供完整的枚举值验证
- 保持与关联关系设计的一致性

### Non-Goals
- 不支持枚举值的动态计算或条件枚举
- 不支持枚举的层级结构或树形枚举
- 不实现枚举值的国际化(可在后续版本考虑)

## Decisions

### 决策1: 单表存储枚举定义和选项

**选择**: 使用单表 + JSON字段存储枚举选项

**理由**:
- 枚举选项通常作为整体使用,不需要单独查询某个选项
- 简化数据库设计和查询逻辑
- 减少JOIN操作,提高查询性能
- 枚举选项的增删改通常是批量操作

**替代方案**: 两表存储(枚举定义表 + 枚举选项表)
- 优点: 更规范的数据库设计,选项可单独管理
- 缺点: 需要额外的表和JOIN查询,复杂度增加

### 决策2: 物理字段存储key,逻辑字段提供展示

**选择**: 枚举字段依赖物理字段存储key值

**实现**:
- 单选枚举: 物理字段类型为 `STRING`,存储枚举选项的key
- 多选枚举: 物理字段类型为 `STRING`,存储JSON数组格式的key列表

**理由**:
- 与关联关系设计保持一致
- 物理存储轻量,只存储标识符
- 逻辑字段可以动态加载完整的枚举信息(key + value)
- 方便数据库索引和查询

### 决策3: 字段通过 EnumName 引用枚举定义

**选择**: 在 `field_definitions` 表增加 `enum_name` 字段,使用枚举的业务标识而非ID

**理由**:
- 与关联关系中使用 `ModelName` 的设计模式一致
- 支持模型和枚举的导入导出场景
- 避免ID冲突问题,name作为全局唯一的业务标识
- 支持外键约束,保证引用完整性
- 支持枚举的复用,多个字段可引用同一个枚举定义

**字段定义扩展**:
```go
type FieldDefinition struct {
    // ... 现有字段
    EnumName *string         `json:"enumName,omitempty"`  // 关联的枚举业务标识
    Enum     *EnumDefinition `json:"enum,omitempty"`      // 查询时加载的枚举详情
}
```

**导入导出优势**:
- 导出: 使用 `enumName` 引用,与环境无关
- 导入: 根据 `name` 匹配枚举,如不存在则一起导入
- 可移植: 不依赖特定环境的ID生成规则

### 决策4: 严格的枚举选项验证

**选择**: key和value必须非空,在创建和更新时强制验证

**验证规则**:
1. 枚举选项的key不能为空字符串
2. 枚举选项的value不能为空字符串
3. 同一枚举定义中的key必须唯一
4. 枚举定义至少包含一个选项

**理由**:
- 确保数据完整性和一致性
- 避免空key导致的查询和匹配问题
- 防止用户误操作创建无效枚举

### 决策5: 兼容现有 EnumValues 字段

**选择**: 保留 `ValidationConfig.EnumValues` 字段,支持两种枚举方式

**使用场景**:
- **简单枚举**: 使用 `EnumValues` 字段,适合不需要复用的简单场景
- **中心化枚举**: 使用 `EnumId` 引用,适合需要复用和统一管理的场景

**互斥规则**: 一个字段不能同时设置 `EnumValues` 和 `EnumName`

## Architecture

### 数据库设计

#### model_enums 表
```sql
CREATE TABLE model_enums (
    id VARCHAR(36) PRIMARY KEY COMMENT '枚举ID,使用UUIDV7,内部使用',
    name VARCHAR(100) NOT NULL UNIQUE COMMENT '枚举英文标识,全局唯一,用于导入导出和字段关联',
    title VARCHAR(200) NOT NULL COMMENT '枚举显示名称',
    description TEXT COMMENT '枚举描述',
    options JSON NOT NULL COMMENT '枚举选项数组 [{"key": "active", "value": "激活", "order": 1, "description": ""}]',
    is_multi_select BOOLEAN DEFAULT FALSE COMMENT '是否支持多选',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    INDEX idx_name (name)
) COMMENT='枚举定义表-中心化管理枚举元数据';
```

**字段说明**:
- `id`: 主键,内部使用,数据库级别的唯一标识
- `name`: 业务标识,全局唯一,用于字段关联、导入导出、API调用
- `title`: 显示名称,用于UI展示
- `options`: JSON数组存储枚举选项
- `is_multi_select`: 标记该枚举是否允许多选

**options JSON 结构**:
```json
[
    {
        "key": "active",
        "value": "激活",
        "order": 1,
        "description": "账户激活状态"
    },
    {
        "key": "inactive",
        "value": "未激活", 
        "order": 2,
        "description": "账户未激活状态"
    }
]
```

#### field_definitions 表修改
```sql
ALTER TABLE field_definitions 
ADD COLUMN enum_name VARCHAR(100) NULL COMMENT '关联的枚举名称(业务标识),用于导入导出',
ADD CONSTRAINT fk_field_enum_name 
    FOREIGN KEY (enum_name) REFERENCES model_enums(name) 
    ON DELETE SET NULL 
    ON UPDATE CASCADE;

CREATE INDEX idx_field_enum_name ON field_definitions(enum_name);
```

**字段说明**:
- `enum_name`: 关联枚举的业务标识(name),而非ID
- 外键约束到 `model_enums.name`
- `ON DELETE SET NULL`: 删除枚举时,字段的 enum_name 设置为 NULL
- `ON UPDATE CASCADE`: 枚举 name 更新时,字段引用自动更新

### 领域模型设计

#### EnumOption 值对象
```go
type EnumOption struct {
    Key         string  `json:"key"`                    // 枚举key,必须非空且唯一
    Value       string  `json:"value"`                  // 枚举显示值,必须非空
    Order       int     `json:"order"`                  // 显示排序
    Description string  `json:"description,omitempty"`  // 选项描述
}

func (eo *EnumOption) Validate() error {
    if strings.TrimSpace(eo.Key) == "" {
        return errors.New("enum option key cannot be empty")
    }
    if strings.TrimSpace(eo.Value) == "" {
        return errors.New("enum option value cannot be empty")
    }
    return nil
}
```

#### EnumDefinition 聚合根
```go
type EnumDefinition struct {
    ID            string        `json:"id"`
    Name          string        `json:"name"`          // 英文标识,全局唯一
    Title         string        `json:"title"`         // 显示名称
    Description   string        `json:"description"`   // 描述
    Options       []EnumOption  `json:"options"`       // 枚举选项列表
    IsMultiSelect bool          `json:"isMultiSelect"` // 是否多选
    CreatedAt     time.Time     `json:"createdAt"`
    UpdatedAt     time.Time     `json:"updatedAt"`
}

func (ed *EnumDefinition) Validate() error {
    if ed.Name == "" {
        return errors.New("enum name cannot be empty")
    }
    if ed.Title == "" {
        return errors.New("enum title cannot be empty")
    }
    if len(ed.Options) == 0 {
        return errors.New("enum must have at least one option")
    }
    
    // 验证选项
    keySet := make(map[string]bool)
    for _, opt := range ed.Options {
        if err := opt.Validate(); err != nil {
            return err
        }
        if keySet[opt.Key] {
            return fmt.Errorf("duplicate enum option key: %s", opt.Key)
        }
        keySet[opt.Key] = true
    }
    
    return nil
}
```

#### FieldDefinition 扩展
```go
type FieldDefinition struct {
    // ... 现有字段
    EnumName     *string           `json:"enumName,omitempty"`   // 关联的枚举业务标识
    Enum         *EnumDefinition   `json:"enum,omitempty"`       // 查询时加载的枚举详情
    
    // 现有的简单枚举支持(保留兼容)
    Validation   *ValidationConfig `json:"validation"`
}

func (fd *FieldDefinition) Validate() error {
    // ... 现有验证逻辑
    
    // 新增: 枚举字段验证
    if fd.Type.Format == FormatEnum || fd.Type.Format == FormatEnumArray {
        // 检查互斥性
        hasEnumName := fd.EnumName != nil && *fd.EnumName != ""
        hasEnumValues := fd.Validation != nil && len(fd.Validation.EnumValues) > 0
        
        if hasEnumName && hasEnumValues {
            return errors.New("cannot use both enumName and enumValues")
        }
        if !hasEnumName && !hasEnumValues {
            return errors.New("enum field must have either enumName or enumValues")
        }
    }
    
    return nil
}
```

### 字段类型扩展

```go
const (
    // ... 现有格式类型
    FormatEnum      FormatType = "ENUM"       // 单选枚举
    FormatEnumArray FormatType = "ENUM_ARRAY" // 多选枚举
)

func init() {
    fieldTypeMap[FormatEnum] = &FieldType{
        SchemaType: SchemaTypeString, 
        Format:     FormatEnum, 
        Title:      "枚举(单选)",
    }
    fieldTypeMap[FormatEnumArray] = &FieldType{
        SchemaType: SchemaTypeArray, 
        Format:     FormatEnumArray, 
        Title:      "枚举(多选)",
    }
}
```

### 数据存储示例

#### 场景: 用户状态枚举字段

**枚举定义**:
```json
{
    "id": "enum_user_status_001",
    "name": "UserStatus",
    "title": "用户状态",
    "description": "用户账户的状态枚举",
    "isMultiSelect": false,
    "options": [
        {"key": "active", "value": "激活", "order": 1},
        {"key": "inactive", "value": "未激活", "order": 2},
        {"key": "suspended", "value": "暂停", "order": 3}
    ]
}
```

**字段定义**:
```json
{
    "modelId": "model_user_001",
    "name": "status",
    "title": "状态",
    "type": {"format": "ENUM"},
    "enumName": "UserStatus"
}
```

**实际数据存储**:
- 物理字段 `status` 存储: `"active"`
- 查询时逻辑字段返回: `{"key": "active", "value": "激活"}`

#### 场景: 用户标签多选枚举

**枚举定义**:
```json
{
    "id": "enum_user_tags_001",
    "name": "UserTags",
    "title": "用户标签",
    "isMultiSelect": true,
    "options": [
        {"key": "vip", "value": "VIP会员", "order": 1},
        {"key": "verified", "value": "已认证", "order": 2},
        {"key": "premium", "value": "高级用户", "order": 3}
    ]
}
```

**字段定义**:
```json
{
    "modelId": "model_user_001",
    "name": "tags",
    "title": "标签",
    "type": {"format": "ENUM_ARRAY"},
    "enumName": "UserTags"
}
```

**实际数据存储**:
- 物理字段 `tags` 存储: `["vip", "verified"]` (JSON数组)
- 查询时逻辑字段返回: `[{"key": "vip", "value": "VIP会员"}, {"key": "verified", "value": "已认证"}]`

## Risks / Trade-offs

### 风险1: JSON字段查询性能

**风险**: JSON字段不支持高效的索引和查询

**缓解**:
- 枚举选项通常数量较少(< 100个选项)
- 查询主要通过枚举ID,不需要查询内部JSON
- 如有性能问题,可在后续版本拆分为两表

### 风险2: 枚举名称变更影响字段引用

**风险**: 修改枚举的 name 可能导致字段引用失效

**缓解**:
- 数据库外键使用 `ON UPDATE CASCADE`,自动更新字段引用
- 提供枚举重命名的影响分析功能
- 建议创建后不要随意修改枚举 name
- API 层面限制枚举 name 的修改权限

### 风险3: 枚举选项值变更影响历史数据

**风险**: 修改或删除枚举选项可能导致历史数据引用失效

**缓解**:
- 提供枚举选项变更的影响分析功能
- 删除枚举前检查是否有字段引用
- 建议使用"禁用"而非"删除"选项
- 考虑在后续版本增加枚举变更历史追踪

### 风险4: 兼容性复杂度

**风险**: 同时支持 EnumValues 和 EnumId 增加系统复杂度

**缓解**:
- 明确互斥规则,不允许同时使用
- 提供迁移工具,帮助用户从 EnumValues 迁移到 EnumName
- 在文档中明确说明两种方式的适用场景

## Migration Plan

### 数据库迁移步骤

1. **创建枚举表**
```sql
CREATE TABLE model_enums (...);
```

2. **修改字段表**
```sql
ALTER TABLE field_definitions ADD COLUMN enum_name VARCHAR(100) NULL;
ALTER TABLE field_definitions ADD CONSTRAINT fk_field_enum_name ...;
```

3. **数据迁移** (可选)
- 扫描现有使用 EnumValues 的字段
- 为常用的枚举值创建枚举定义
- 更新字段引用新的枚举定义

### 代码部署步骤

1. 部署数据库迁移脚本
2. 部署新版本代码(支持枚举功能)
3. 验证枚举创建和字段关联功能
4. 逐步迁移现有枚举字段(非强制)

### 回滚方案

- 如需回滚,保留 `enum_name` 字段数据
- 回滚代码版本,枚举字段回退到 EnumValues 模式
- 枚举表数据保留,不影响后续重新部署

## Open Questions

1. **枚举国际化**: 是否需要支持多语言的枚举值? 
   - 建议: 后续版本考虑,可在 options JSON 中增加 `i18n` 字段

2. **枚举权限控制**: 是否需要控制枚举的可见性和可编辑性?
   - 建议: 后续版本考虑,可增加 `editable` 和 `deletable` 字段

3. **枚举值颜色/图标**: 是否需要支持枚举选项的UI展示配置?
   - 建议: 可在 options 的 JSON 中增加扩展字段,如 `color`, `icon`

4. **枚举依赖关系**: 是否需要支持级联枚举(一个枚举的选项依赖另一个枚举的值)?
   - 建议: 需求明确后在后续版本实现
