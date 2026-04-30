# 🔐 ModelCraft Permission Bundle 权限包完整文档集

本目录包含关于 ModelCraft 权限包（Permission Bundle）下"数据权限配置"的完整研究和参考文档。

## 📚 文档导航

### 1. 🔍 **PERMISSION_BUNDLE_RESEARCH.md** (研究报告)
**适合**: 想要深入理解权限系统的开发者

内容包括：
- 核心发现：截图字段的字段对应关系
- 完整的 Item 数据结构定义
- 两种权限模式（PRESET vs CUSTOM）的详细对比
- 数据库 Schema 完整分析
- 数据流示例
- 设计原理深度解读

**快速定位**：
- 想知道 `d97f7a43...` 和 `读写全部` 对应什么字段？ → 第"核心发现"段
- 需要理解 EndUserBundleDataPermissionItem 结构？ → 第"完整的 Item 字段列表"段
- 想学习两种权限模式？ → 第"两种权限授权模式详解"段

---

### 2. 📋 **PERMISSION_BUNDLE_QUICK_REF.md** (快速参考卡片)
**适合**: 需要快速查阅关键信息的开发者

内容包括：
- 🎯 截图字段解读速查表
- 🏗️ 核心数据结构概览
- 📊 数据库表关系图
- 📁 关键文件位置索引
- 🚀 常见操作示例
- 💡 关键设计原则

**快速定位**：
- 需要查预设权限值的含义？ → 第"两种权限模式"段
- 不知道某个字段在哪个文件定义的？ → 第"关键文件位置"段
- 想看添加权限的代码示例？ → 第"常见操作"段

---

### 3. 📊 **PERMISSION_BUNDLE_SUMMARY.txt** (速查表)
**适合**: 需要一个全能型对照表的开发者

内容包括：
- 关键概念表
- 预设值完整对比表（Box Drawing 美化）
- 数据库表关系ASCII图
- SQLC 查询操作速查
- 文件位置速查表
- JSON 策略字段详解
- GraphQL 操作示例

**快速定位**：
- 需要看表格式对比？ → 第"两种权限授权模式的完整对比"段
- 想要 ASCII 关系图？ → 第"数据库表关系图"段

---

### 4. 📖 **PERMISSION_BUNDLE_ANALYSIS.md** (完整分析文档)
**适合**: 需要全面系统理解的开发者

内容包括：
- GraphQL Schema 完整定义 (500+ 行)
- 数据库 Schema 逐表解析
- SQLC 查询文件详细说明
- 关键字段映射表
- 完整的 Item 结构示例（JSON格式）
- 数据流关系图

**快速定位**：
- 需要查 GraphQL 类型的完整定义？ → 第"GraphQL Schema 定义"段
- 想看数据库每个字段的注释？ → 第"数据库 Schema"段
- 需要 SQLC 查询的完整说明？ → 第"SQLC 查询"段

---

## 🎯 使用场景指南

### 场景 1：我要快速弄清"数据权限配置"列表项的字段含义
1. 打开 **PERMISSION_BUNDLE_RESEARCH.md**
2. 跳到"核心发现"段
3. 查看字段对应表 ✅

**耗时**: 2 分钟

---

### 场景 2：我要实现一个新功能，需要理解权限系统架构
1. 打开 **PERMISSION_BUNDLE_RESEARCH.md**
2. 阅读"完整的 Item 字段列表"段
3. 学习"两种权限授权模式详解"段
4. 查看"数据库 Schema 关系图"段
5. 研究"常见操作代码示例"段

**耗时**: 30 分钟

---

### 场景 3：我在写 GraphQL 查询，需要知道 dataPermissionItems 的字段
1. 打开 **PERMISSION_BUNDLE_QUICK_REF.md**
2. 查看"核心数据结构"段的 EndUserBundleDataPermissionItem 定义
3. 或打开 **PERMISSION_BUNDLE_ANALYSIS.md**
4. 搜索"type EndUserBundleDataPermissionItem"

**耗时**: 5 分钟

---

### 场景 4：我要查某个 SQLC 查询的功能
1. 打开 **PERMISSION_BUNDLE_SUMMARY.txt**
2. 查看"SQLC 查询操作速查表"
3. 或打开 **PERMISSION_BUNDLE_ANALYSIS.md**
4. 搜索查询名称

**耗时**: 3 分钟

---

### 场景 5：我要查某个 GraphQL/数据库字段在代码中的具体位置
1. 打开 **PERMISSION_BUNDLE_QUICK_REF.md**
2. 查看"关键文件位置"表
3. 跳到对应行号

**耗时**: 2 分钟

---

### 场景 6：我要学习权限包的两种授权模式（PRESET vs CUSTOM）
1. 打开 **PERMISSION_BUNDLE_QUICK_REF.md**
2. 查看"两种权限模式"小节
3. 查看数据库存储示例

或

1. 打开 **PERMISSION_BUNDLE_SUMMARY.txt**
2. 查看"两种权限授权模式的完整对比"表格

**耗时**: 10 分钟

---

## 🗂️ 文件对应表

| 需要查什么 | 推荐文档 | 快速定位 |
|----------|--------|--------|
| 截图字段含义 | RESEARCH | "核心发现" |
| Item完整结构 | RESEARCH | "完整的Item字段列表" |
| GraphQL类型定义 | ANALYSIS | "GraphQL Schema定义" |
| 数据库表字段 | ANALYSIS | "数据库Schema" |
| SQLC查询列表 | ANALYSIS 或 SUMMARY | "SQLC查询" |
| 文件位置索引 | QUICK_REF 或 ANALYSIS | "关键文件位置" |
| 预设值含义 | QUICK_REF 或 SUMMARY | "两种权限模式" |
| 常见操作代码 | QUICK_REF 或 SUMMARY | "常见操作" |
| 表格对比 | SUMMARY | "预设值对比表" |
| ASCII关系图 | SUMMARY | "数据库表关系图" |

---

## 🎓 学习路径建议

### 初级（了解基本概念）
1. 阅读 **PERMISSION_BUNDLE_QUICK_REF.md** 的"🎯 截图中的字段解读"
2. 阅读"🏗️ 核心数据结构"
3. 查看"📊 两种权限模式"

**目标**: 能解释"d97f7a43..."和"读写全部"对应什么  
**耗时**: 15 分钟

---

### 中级（理解工作原理）
1. 阅读 **PERMISSION_BUNDLE_RESEARCH.md** 完整报告
2. 专注"🎯 两种权限授权模式详解"段
3. 学习"🏗️ 数据库 Schema 关系图"段

**目标**: 能解释权限配置如何存储和查询  
**耗时**: 1 小时

---

### 高级（能独立开发新功能）
1. 深入研究 **PERMISSION_BUNDLE_ANALYSIS.md**
2. 查看实际的 GraphQL schema 文件
3. 查看实际的数据库 migration 文件
4. 查看实际的 SQLC 查询代码
5. 查看实际的 Go 数据模型代码

**目标**: 能从 GraphQL 到数据库完整地实现新功能  
**耗时**: 2-3 小时

---

## 📍 核心文件位置速查

### GraphQL Schema
```
modelcraft-backend/api/graph/project/schema/rbac.graphql
  - 第 189-204 行: EndUserBundleDataPermissionItem 类型
  - 第 62-67 行: EndUserPermissionPreset enum
  - 第 72-75 行: DataPermissionGrantType enum
```

### 数据库 Schema
```
modelcraft-backend/db/schema/mysql/13_rbac_permissions.sql
  - 第 98-138 行: end_user_bundle_data_permission_items 表 ⭐

modelcraft-backend/db/schema/mysql/14_rbac_bundle_snapshots.sql
  - 第 12-32 行: end_user_permission_bundle_snapshots 表
```

### SQLC 查询
```
modelcraft-backend/db/queries/rbac/bundle.sql
  - 第 48-89 行: Bundle Item 相关查询
```

### Go 代码
```
modelcraft-backend/internal/infrastructure/dbgen/models.go
  - EndUserBundleDataPermissionItem struct 定义
```

---

## 🔗 相关链接

- 完整 GraphQL Schema: `api/graph/project/schema/rbac.graphql`
- 权限系统应用层: `internal/app/rbac/`
- 数据库访问层: `internal/infrastructure/repository/`
- 前端合约: `modelcraft-front/contract/graph/org/schema/`

---

## 💬 常见问题

**Q: 为什么有这么多文档？**  
A: 不同的用户有不同的需求。选择适合你的文档即可：
- 快速查阅 → QUICK_REF 或 SUMMARY
- 深入学习 → RESEARCH
- 完整参考 → ANALYSIS

**Q: 应该从哪个文档开始？**  
A: 如果是第一次接触权限包系统，建议从 PERMISSION_BUNDLE_QUICK_REF.md 开始，然后根据需要深入查看其他文档。

**Q: 这些文档会不会过时？**  
A: 这些文档是基于 2026-05-01 的代码生成的。如果代码有更新，可能需要更新文档。建议查看实际的源代码文件作为权威参考。

**Q: 我想贡献补充内容？**  
A: 欢迎！请直接编辑相应的 .md 或 .txt 文件，或者创建新的参考文档。

---

## 📊 文档统计

| 文档 | 大小 | 类型 | 内容 |
|-----|-----|-----|------|
| PERMISSION_BUNDLE_RESEARCH.md | 14 KB | Markdown | 研究报告 |
| PERMISSION_BUNDLE_QUICK_REF.md | 8 KB | Markdown | 快速参考 |
| PERMISSION_BUNDLE_ANALYSIS.md | 17 KB | Markdown | 完整分析 |
| PERMISSION_BUNDLE_SUMMARY.txt | 16 KB | 纯文本 | 速查表 |
| **总计** | **55 KB** | | |

---

**最后更新**: 2026-05-01  
**文档版本**: 1.0  
**适用项目**: ModelCraft  
**作者**: AI Research Assistant
