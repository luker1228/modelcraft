# 3. 项目分离

> 代码位置：`internal/domain/project/`

## 概述

Project 是 Org 内的资源隔离单元。模型、数据库集群都归属于某个 Project。一个 Org 可以有多个 Project，Project 之间完全隔离。

## 核心实体

```
internal/domain/project/project.go

Project
├── OrgName      string          // 所属 Org（复合主键之一）
├── Slug         string          // 项目标识符（复合主键之一）
│                                // 格式：3-64 字符，小写字母开头，允许数字/下划线
├── Title        string          // 显示名称
├── Description  string
├── LoginURL     string          // 运行态登录跳转 URL（可选）
├── ClusterID    *string         // 关联的 DatabaseCluster ID（可选，1:1）
└── Status       ProjectStatus   // active | archived
```

## 复合主键

Project 使用 `(OrgName, Slug)` 作为复合主键，不使用自增 ID。这意味着：
- Slug 在同一 Org 内唯一
- 所有引用 Project 的地方都用 `(orgName, projectSlug)` 定位

## 与 DatabaseCluster 的关系

Project 与 DatabaseCluster 是 **1:1 可选关联**：

```
Project ──── (可选) ────▶ DatabaseCluster
```

- 一个 Project 最多关联一个 Cluster
- 未关联 Cluster 的 Project 无法使用运行态 GraphQL
- 通过 `SetCluster()` / `UnsetCluster()` 管理关联

## 生命周期

```
NewProject()
      │
      ▼
   active  ──Archive()──▶  archived
      │
   Activate()
      │
      ▼
   active
```

## ProjectScope 值对象

ProjectScope 是一个可复用的值对象，用于需要完整项目上下文的实体。它封装了 OrgName 和 ProjectSlug，确保这两个关键字段始终同时存在且有效。

代码位置：`internal/domain/project/project_scope.go`

### 结构定义

```go
type ProjectScope struct {
    OrgName     string  // 组织名称
    ProjectSlug string  // 项目标识符
}

// Validate 验证两个字段都不为空（包括空白字符串检查）
func (s *ProjectScope) Validate() error

// NewProjectScope 工厂方法：创建并验证
func NewProjectScope(orgName, projectSlug string) (ProjectScope, error)

// GetFullPath 返回 "orgName.projectSlug" 格式的完整路径
func (s *ProjectScope) GetFullPath() string
```

### 特点

- **复用性**: 被多个实体嵌入（ModelLocator、EnumDefinition、ModelGroup、FieldEnumAssociation）
- **验证**: `Validate()` 方法确保两个字段都不为空，拒绝仅空白字符的值
- **构造**: `NewProjectScope()` 工厂方法强制验证
- **路径**: `GetFullPath()` 返回 `"orgName.projectSlug"` 格式

### 使用示例

```go
scope, err := project.NewProjectScope("my-org", "my-project")
if err != nil {
    return err  // OrgName 或 ProjectSlug 为空时返回错误
}
path := scope.GetFullPath()  // "my-org.my-project"
```

### 在其他实体中的应用

ProjectScope 通过 **Go 结构体嵌入** 的方式被以下实体复用：

- **ModelLocator** — 扩展为 `org.project.database.model` 四层路径
- **EnumDefinition** — 表示属于某个项目的枚举类型
- **ModelGroup** — 表示属于某个项目的模型分组
- **FieldEnumAssociation** — 关联字段到枚举，确保同项目范围

嵌入后，这些实体的 `Validate()` 方法会首先调用 `ProjectScope.Validate()` 确保项目上下文有效。

## 相关文件

- `internal/domain/project/project.go` — Project 实体定义
- `internal/domain/project/project_scope.go` — ProjectScope 值对象定义
- `internal/domain/project/repository.go` — 仓储接口
