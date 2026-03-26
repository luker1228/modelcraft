# sqlc 自定义类型规范

> **优先级: 高** - 定义 ModelCraft Go 后端的 sqlc 自定义类型实现标准。

## 核心原则

- **JSON 类型字段必须使用 `db:"type:json"` 标签**，不可省略
- 自定义类型必须实现 `sql.Scanner` 和 `driver.Valuer` 接口
- `Scan` 方法必须先检查 `fmt.Stringer` 接口，再处理标准类型

## 自定义类型实现模板

### StringSlice 示例

```go
package dbgen

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
)

// StringSlice 自定义类型用于处理字符串数组
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
    if len(s) == 0 {
        return nil, nil
    }
    return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
    if value == nil {
        *s = nil
        return nil
    }

    var bytes []byte

    // 参考 sqlc 官方 datatypes 包的做法，先检查 fmt.Stringer 接口
    if str, ok := value.(fmt.Stringer); ok {
        bytes = []byte(str.String())
    } else {
        switch v := value.(type) {
        case []byte:
            if len(v) == 0 {
                *s = nil
                return nil
            }
            bytes = v
        case string:
            if v == "" {
                *s = nil
                return nil
            }
            bytes = []byte(v)
        default:
            return fmt.Errorf("cannot scan %T into StringSlice", value)
        }
    }

    if len(bytes) == 0 {
        *s = nil
        return nil
    }

    return json.Unmarshal(bytes, s)
}
```

## 结构体定义

### JSON 字段必须有 type 标签

```go
// ModelRelationPO 结构体定义
type ModelRelationPO struct {
    ID                 string      `db:"primaryKey" json:"id"`
    ModelId            string      `json:"modelId"`
    Name               string      `json:"name"`
    RelationType       string      `json:"relation_type"`
    ModelName          string      `json:"modelName"`
    SourceFields       StringSlice `db:"type:json" json:"source_fields"`       // ✅ 必须有 db:"type:json"
    TargetFields       StringSlice `db:"type:json" json:"target_fields"`       // ✅ 必须有 db:"type:json"
    ThroughTable       *string     `json:"through_table"`
    CreatedAt          time.Time   `db:"autoCreateTime" json:"created_at"`
    UpdatedAt          time.Time   `db:"autoUpdateTime" json:"updated_at"`
}
```

## 错误示例

### ❌ 错误 1：JSON 字段缺少 db:"type:json" 标签

```go
type ModelRelationPO struct {
    SourceFields StringSlice `json:"source_fields"` // 禁止！必须添加 db:"type:json"
}
```

### ❌ 错误 2：Scan 方法没有先检查 fmt.Stringer 接口

```go
func (s *StringSlice) Scan(value interface{}) error {
    switch v := value.(type) {
    case []byte:  // 禁止！应该先检查 fmt.Stringer
        return json.Unmarshal(v, s)
    }
}
```

### ❌ 错误 3：Scan 方法处理了非标准类型

```go
func (s *StringSlice) Scan(value interface{}) error {
    switch v := value.(type) {
    case []interface{}:  // 禁止！不是 sqlc 标准做法
        for _, item := range v {
            *s = append(*s, item.(string))
        }
    }
}
```

## 原理说明

### 1. `db:"type:json"` 标签

明确告诉 sqlc 字段类型，确保 MySQL 驱动返回正确的数据类型，避免 `&[]` 等异常类型。

### 2. `fmt.Stringer` 接口

sqlc 官方 `datatypes` 包推荐做法，处理未知类型时先尝试调用 `String()` 方法。

### 3. 只处理标准类型

`[]byte` 和 `string` 是数据库驱动返回的标准类型，其他类型应返回错误而非尝试转换。

## 常见自定义类型

### StringSlice - 字符串数组

```go
type StringSlice []string
```

适用于存储标签、权限列表等字符串数组。

### IntSlice - 整数数组

```go
type IntSlice []int

func (s IntSlice) Value() (driver.Value, error) {
    if len(s) == 0 {
        return nil, nil
    }
    return json.Marshal(s)
}

func (s *IntSlice) Scan(value interface{}) error {
    // 实现同 StringSlice
}
```

### JSONMap - JSON 对象

```go
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
    if len(m) == 0 {
        return nil, nil
    }
    return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
    // 实现同 StringSlice
}
```

## 测试建议

```go
func TestStringSlice_ScanAndValue(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected StringSlice
        wantErr  bool
    }{
        {"nil", nil, nil, false},
        {"empty string", "", nil, false},
        {"empty bytes", []byte{}, nil, false},
        {"valid json bytes", []byte(`["a","b"]`), StringSlice{"a", "b"}, false},
        {"valid json string", `["a","b"]`, StringSlice{"a", "b"}, false},
        {"invalid type", 123, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var s StringSlice
            err := s.Scan(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !reflect.DeepEqual(s, tt.expected) {
                t.Errorf("Scan() got = %v, want %v", s, tt.expected)
            }
        })
    }
}
```

## sqlc 常见坑

### 1. 禁止使用 `db.Model`

`db.Model` 内置 `DeletedAt` 字段，会自动启用软删除，可能导致查询结果不符合预期。

```go
// ✅ 正确
type UserPO struct {
    ID        string    `db:"primaryKey" json:"id"`
    CreatedAt time.Time `db:"autoCreateTime" json:"created_at"`
    UpdatedAt time.Time `db:"autoUpdateTime" json:"updated_at"`
}

// ❌ 禁止
type UserPO struct {
    db.Model  // 禁止！会自动启用软删除
    Name string
}
```

### 2. struct 更新零值问题

使用 struct 进行 `Updates()` 时，零值字段（`""`, `0`, `false`, `nil`）**不会被更新**。

```go
// ✅ 方式1：使用 map 更新零值
db.Model(&user).Updates(map[string]interface{}{
    "name":   "",
    "status": 0,
    "active": false,
})

// ✅ 方式2：使用 Select 明确指定字段
db.Model(&user).Select("Name", "Status", "Active").Updates(user)

// ❌ 禁止：零值字段不会被更新
user.Name = ""
user.Status = 0
user.Active = false
db.Model(&user).Updates(user)  // Name/Status/Active 不会被更新
```

### 3. 预加载避免 N+1 查询

查询关联数据时必须使用 `Preload`，否则会产生 N+1 查询问题。

```go
// ✅ 正确：预加载关联数据
var models []ModelPO
db.Preload("Fields").Preload("Fields.Relation").Find(&models)

// ❌ 禁止：N+1 查询问题
var models []ModelPO
db.Find(&models)
for _, m := range models {
    db.Where("model_id = ?", m.ID).Find(&m.Fields)  // 每次循环都查询
}
```

### 4. 外键定义格式

外键定义时，`foreignKey` 是当前模型的字段，`references` 是关联模型的字段。

```go
type FieldDefinitionPO struct {
    ModelID string           `db:"primaryKey"`
    Name    string           `db:"primaryKey"`
    // foreignKey: 当前模型用于关联的字段
    // references: 关联模型中对应的字段
    Relation *ModelRelationPO `db:"foreignKey:model_id,name;references:model_id,name"`
}

type ModelRelationPO struct {
    ModelID   string `db:"column:model_id"`   // 对应 references 中的 model_id
    ModelName string `db:"column:model_name"` // 对应 references 中的 name
}
```

### 5. 查询条件不要用 map

`Where(map)` 的行为可能不符合预期，建议使用明确条件。

```go
// ✅ 正确
db.Where("org_name = ? AND project_slug = ?", orgName, projectSlug).Find(&models)

// ❌ 禁止
db.Where(map[string]interface{}{
    "org_name": orgName,
    "project_slug": projectSlug,
}).Find(&models)  // 行为可能不符合预期
```
