# CLI 资源发现与 Agent 自省

---

## 1. 概述

资源发现和自省是 Agent-First 设计的核心。Agent 需要能够：
- **发现**可访问的项目、数据库、模型（`mc catalog`）
- **了解**模型的字段、类型、关系、限制（`mc describe`）
- **理解** CLI 自身的命令结构和参数（`mc schema`）

---

## 2. `mc catalog` — 资源列表

### 2.1 `mc catalog projects`

列出当前 EndUser 可访问的项目。

数据来源：登录时返回的 `projects` 列表 / EndUser GraphQL `listAccessibleProjects`。

```bash
mc catalog projects
```

```json
{
  "ok": true,
  "projects": [
    { "slug": "sales", "title": "销售系统" },
    { "slug": "hr", "title": "HR 系统" }
  ]
}
```

### 2.2 `mc catalog databases`

列出指定项目内的数据库。

数据来源：EndUser GraphQL `modelDatabaseCatalog` 查询。

```bash
mc catalog databases --project sales
```

```json
{
  "ok": true,
  "databases": [
    { "name": "maindb" },
    { "name": "analytics" }
  ]
}
```

若省略 `--project`，使用当前活跃项目。

### 2.3 `mc catalog models`

列出指定数据库内的模型。

数据来源：EndUser GraphQL `modelCatalog` 查询。

```bash
mc catalog models --project sales --database maindb
```

```json
{
  "ok": true,
  "models": [
    { "name": "users", "title": "用户表", "databaseName": "maindb" },
    { "name": "orders", "title": "订单表", "databaseName": "maindb" },
    { "name": "departments", "title": "部门表", "databaseName": "maindb" }
  ]
}
```

### 2.4 `mc catalog` 无子命令

若不带子命令，输出当前上下文的完整资源树：

```bash
mc catalog
```

```json
{
  "ok": true,
  "context": {
    "server": "https://mc.example.com",
    "org": "acme",
    "currentProject": "sales"
  },
  "projects": [
    {
      "slug": "sales",
      "title": "销售系统",
      "databases": [
        {
          "name": "maindb",
          "models": [
            { "name": "users", "title": "用户表" },
            { "name": "orders", "title": "订单表" }
          ]
        }
      ]
    }
  ]
}
```

---

## 3. `mc describe` — 资源元数据

### 3.1 描述模型

输出模型的完整元数据：字段列表、类型、关系、查询限制。

Agent 在执行查询前调用 `mc describe` 了解模型结构。

```bash
mc describe sales.maindb.users
```

```json
{
  "ok": true,
  "model": {
    "name": "users",
    "title": "用户表",
    "database": "maindb",
    "project": "sales",
    "displayField": "username",
    "fields": [
      {
        "name": "id",
        "type": "String",
        "isPrimary": true,
        "isRequired": true,
        "description": "主键"
      },
      {
        "name": "username",
        "type": "String",
        "isRequired": true,
        "isUnique": true
      },
      {
        "name": "email",
        "type": "String",
        "isRequired": false
      },
      {
        "name": "age",
        "type": "Int",
        "isRequired": false
      },
      {
        "name": "status",
        "type": "Enum",
        "isRequired": true,
        "enumValues": ["active", "inactive", "deleted"]
      },
      {
        "name": "departmentId",
        "type": "String",
        "isRequired": false,
        "description": "外键 → departments"
      },
      {
        "name": "department",
        "type": "Relation",
        "relationType": "ManyToOne",
        "relatedModel": "departments",
        "foreignKey": "departmentId"
      },
      {
        "name": "createdAt",
        "type": "DateTime",
        "isRequired": true
      },
      {
        "name": "updatedAt",
        "type": "DateTime",
        "isRequired": true
      }
    ],
    "limits": {
      "maxTake": 100,
      "defaultTake": 20,
      "note": "Limits are enforced server-side. Use --take and --skip for pagination."
    }
  }
}
```

### 3.2 描述数据库

```bash
mc describe sales.maindb
```

```json
{
  "ok": true,
  "database": {
    "name": "maindb",
    "project": "sales",
    "modelCount": 5,
    "models": [
      { "name": "users", "title": "用户表" },
      { "name": "orders", "title": "订单表" },
      { "name": "departments", "title": "部门表" },
      { "name": "products", "title": "产品表" },
      { "name": "order_items", "title": "订单明细表" }
    ]
  }
}
```

### 3.3 路径段数判断

| 路径段数 | 解析为 | 示例 |
|----------|--------|------|
| 1 段 | 模型（使用当前 project + database） | `mc describe users` |
| 2 段 | 数据库.模型 或 项目.数据库 | `mc describe maindb.users` |
| 3 段 | 项目.数据库.模型 | `mc describe sales.maindb.users` |

2 段歧义处理：优先解析为 `database.model`（更常用）。若需指定为 `project.database`，使用 `--project` 标志覆盖。

---

## 4. `mc schema` — CLI 自省

Agent 专用。输出 CLI 自身的命令结构和参数 schema，供 Agent 自主发现可用能力。

### 4.1 `mc schema commands`

输出完整命令树：

```json
{
  "name": "mc",
  "version": "1.0.0",
  "commands": [
    {
      "name": "auth",
      "description": "Authentication management",
      "subcommands": [
        {
          "name": "login",
          "description": "Authenticate with ModelCraft server",
          "usage": "mc auth login [flags]",
          "flags": [
            { "name": "--server", "type": "string", "required": true, "description": "Server URL" },
            { "name": "--org", "type": "string", "required": true, "description": "Organization name" },
            { "name": "--username", "type": "string", "required": true, "description": "EndUser username" },
            { "name": "--password", "type": "string", "required": true, "description": "EndUser password" },
            { "name": "--project", "type": "string", "required": false, "description": "Project to select after login" }
          ]
        },
        {
          "name": "logout",
          "description": "Log out and revoke tokens",
          "usage": "mc auth logout"
        },
        {
          "name": "refresh",
          "description": "Refresh access token",
          "usage": "mc auth refresh"
        },
        {
          "name": "status",
          "description": "Show current authentication status",
          "usage": "mc auth status"
        },
        {
          "name": "switch-project",
          "description": "Switch to a different project",
          "usage": "mc auth switch-project --project <slug>",
          "flags": [
            { "name": "--project", "type": "string", "required": true, "description": "Target project slug" }
          ]
        }
      ]
    },
    {
      "name": "query",
      "description": "Query records from a model (findMany)",
      "usage": "mc query <resource-path> [flags]",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model as project.database.model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": false, "description": "Filter condition in JSON" },
        { "name": "--select", "type": "json-array", "required": false, "description": "Fields to return" },
        { "name": "--orderBy", "type": "json", "required": false, "description": "Sort order {\"field\":\"asc|desc\"}" },
        { "name": "--take", "type": "int", "required": false, "default": 20, "description": "Records to return. Server enforces max; use 'mc describe' to check." },
        { "name": "--skip", "type": "int", "required": false, "default": 0, "description": "Records to skip" }
      ]
    },
    {
      "name": "get",
      "description": "Get a single record (findUnique)",
      "usage": "mc get <resource-path> --where '<json>' [flags]",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": true, "description": "Unique identifier condition" },
        { "name": "--select", "type": "json-array", "required": false, "description": "Fields to return" }
      ]
    },
    {
      "name": "create",
      "description": "Create a new record",
      "usage": "mc create <resource-path> --data '<json>' [flags]",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--data", "type": "json", "required": true, "description": "Record data" },
        { "name": "--select", "type": "json-array", "required": false, "description": "Fields to return" }
      ]
    },
    {
      "name": "update",
      "description": "Update an existing record",
      "usage": "mc update <resource-path> --where '<json>' --data '<json>' [flags]",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": true, "description": "Record identifier" },
        { "name": "--data", "type": "json", "required": true, "description": "Fields to update" },
        { "name": "--select", "type": "json-array", "required": false, "description": "Fields to return" }
      ]
    },
    {
      "name": "delete",
      "description": "Delete a record",
      "usage": "mc delete <resource-path> --where '<json>'",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": true, "description": "Record identifier" }
      ]
    },
    {
      "name": "count",
      "description": "Count records matching a filter",
      "usage": "mc count <resource-path> [--where '<json>']",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": false, "description": "Filter condition" }
      ]
    },
    {
      "name": "aggregate",
      "description": "Run aggregate functions on a model",
      "usage": "mc aggregate <resource-path> [flags]",
      "args": [
        { "name": "resource-path", "required": true, "description": "Target model" }
      ],
      "flags": [
        { "name": "--where", "type": "json", "required": false, "description": "Filter condition" },
        { "name": "--avg", "type": "json-array", "required": false, "description": "Fields to average" },
        { "name": "--sum", "type": "json-array", "required": false, "description": "Fields to sum" },
        { "name": "--min", "type": "json-array", "required": false, "description": "Fields to find minimum" },
        { "name": "--max", "type": "json-array", "required": false, "description": "Fields to find maximum" }
      ]
    },
    {
      "name": "describe",
      "description": "Show metadata for a resource (fields, types, relations, limits)",
      "usage": "mc describe <resource-path>"
    },
    {
      "name": "catalog",
      "description": "List accessible resources",
      "subcommands": [
        { "name": "projects", "description": "List accessible projects" },
        { "name": "databases", "description": "List databases in a project" },
        { "name": "models", "description": "List models in a database" }
      ]
    },
    {
      "name": "schema",
      "description": "CLI self-introspection for agents",
      "subcommands": [
        { "name": "commands", "description": "Full command tree as JSON" },
        { "name": "query", "description": "Query command schema detail" },
        { "name": "flags", "description": "Global flags schema" }
      ]
    },
    {
      "name": "version",
      "description": "Show CLI version"
    }
  ]
}
```

### 4.2 `mc schema flags`

输出全局标志定义：

```json
{
  "globalFlags": [
    { "name": "--output", "short": "-o", "type": "enum", "values": ["json", "yaml"], "default": "json", "description": "Output format" },
    { "name": "--server", "type": "string", "description": "Override server URL from credentials" },
    { "name": "--project", "type": "string", "description": "Override current project" },
    { "name": "--database", "type": "string", "description": "Override current database" },
    { "name": "--quiet", "short": "-q", "type": "bool", "description": "Suppress non-essential output" },
    { "name": "--verbose", "short": "-v", "type": "bool", "description": "Verbose logging to stderr" }
  ]
}
```

### 4.3 Agent 使用流程

典型 Agent 工作流：

```
1. mc schema commands          → 了解 CLI 能力
2. mc auth login ...           → 认证
3. mc catalog projects         → 发现可用项目
4. mc catalog databases        → 发现可用数据库
5. mc catalog models           → 发现可用模型
6. mc describe <model>         → 了解模型字段和限制
7. mc query <model> --where .. → 执行查询
8. mc create <model> --data .. → 创建数据
```

Agent 可以跳过 1-5，直接从已知的资源路径开始。自省命令的价值在于 Agent **初次接触未知 ModelCraft 实例**时的自主发现。
