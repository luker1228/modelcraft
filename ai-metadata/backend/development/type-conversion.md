# 类型转换规范

> **优先级: 高** - 定义 ModelCraft Go 后端的类型转换标准。

## 核心原则

- **禁止使用 Go 原生类型断言 `x.(T)`**，必须使用 `github.com/spf13/cast` 包进行类型转换
- 需要区分"值缺失"与"类型错误"时，使用带 `E` 后缀的函数（返回 error）
- 只关心转换结果时，使用不带 `E` 的函数（零值降级）

## 规则

| 场景 | 使用方式 |
|------|----------|
| 必须有值，类型错误需报错 | `cast.ToStringE(v)` + `bizerrors.Wrap(err, ...)` |
| 可选值，类型错误降级为零值 | `cast.ToString(v)`，再判断空值 |

## 示例

### ✅ 正确示例

```go
import "github.com/spf13/cast"

// 必填字段：类型错误需报错
schemaType, err := cast.ToStringE(schemaMap["type"])
if err != nil {
    return nil, bizerrors.Wrap(err, "invalid type")
}

// 必填字段：值为空也需报错
title := cast.ToString(schemaMap["title"])
if title == "" {
    return nil, bizerrors.New("required metadata 'title' is missing")
}

// 可选字段：类型错误静默降级为零值
description := cast.ToString(schemaMap["description"])

// 数值类型
port, err := cast.ToIntE(config["port"])
if err != nil {
    return nil, bizerrors.Wrap(err, "invalid port")
}

// 布尔类型
enabled := cast.ToBool(config["enabled"])

// 切片类型
tags := cast.ToStringSlice(config["tags"])
```

### ❌ 错误示例

```go
// 禁止：原生类型断言，panic 风险
schemaType := schemaMap["type"].(string)

// 禁止：ok 模式虽然安全，但代码冗长，且项目统一用 cast
schemaType, ok := schemaMap["type"].(string)
if !ok {
    return nil, bizerrors.New("invalid type")
}
```

## 常用函数

### 字符串转换

```go
cast.ToString(v)      // 转换为 string，失败返回 ""
cast.ToStringE(v)     // 转换为 string，失败返回 error
```

### 数值转换

```go
cast.ToInt(v)         // 转换为 int，失败返回 0
cast.ToIntE(v)        // 转换为 int，失败返回 error
cast.ToInt64(v)       // 转换为 int64，失败返回 0
cast.ToInt64E(v)      // 转换为 int64，失败返回 error
cast.ToFloat64(v)     // 转换为 float64，失败返回 0.0
cast.ToFloat64E(v)    // 转换为 float64，失败返回 error
```

### 布尔转换

```go
cast.ToBool(v)        // 转换为 bool，失败返回 false
cast.ToBoolE(v)       // 转换为 bool，失败返回 error
```

### 切片转换

```go
cast.ToStringSlice(v)    // 转换为 []string，失败返回 []string{}
cast.ToStringSliceE(v)   // 转换为 []string，失败返回 error
cast.ToIntSlice(v)       // 转换为 []int，失败返回 []int{}
cast.ToIntSliceE(v)      // 转换为 []int，失败返回 error
```

### Map 转换

```go
cast.ToStringMap(v)           // 转换为 map[string]interface{}
cast.ToStringMapE(v)          // 转换为 map[string]interface{}，失败返回 error
cast.ToStringMapString(v)     // 转换为 map[string]string
cast.ToStringMapStringE(v)    // 转换为 map[string]string，失败返回 error
```

## 选择策略

### 使用 `ToXxxE` 的场景

- 必填字段，值缺失或类型错误需要明确报错
- 需要区分"值不存在"和"类型错误"
- 业务逻辑依赖转换成功

### 使用 `ToXxx` 的场景

- 可选字段，类型错误可以降级为零值
- 配置项，有默认值兜底
- 不需要区分错误类型的场景
