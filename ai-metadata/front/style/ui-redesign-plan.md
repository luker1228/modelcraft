# ModelCraft UI Redesign Plan

> **Status**: Plan — 不涉及代码变更
> **Date**: 2025-04-25
> **Reference**: dashboard.stripe.com (后台产品体验，非营销官网)
> **Prerequisite**: [design-direction-stripe.md](./design-direction-stripe.md)

---

## 目录

1. [当前 UI 与 Stripe Dashboard 的差距分析](#1-差距分析)
2. [信息架构建议](#2-信息架构建议)
3. [Layout 设计方案](#3-layout-设计方案)
4. [区域规范: Sidebar / Header / PageHeader / Content](#4-区域规范)
5. [组件规范: Table / Form / Modal / DetailPage / EmptyState](#5-组件规范)
6. [样板页选择](#6-样板页选择)
7. [分阶段改造路线](#7-分阶段改造路线)

---

## 1. 差距分析

### 1.1 逐项对照（10 个维度）

| # | Stripe 特点 | ModelCraft 现状 | 差距级别 | 具体差距 |
|---|---|---|---|---|
| 1 | 左侧导航清晰，模块分组明确 | 有 Sidebar 但结构扁平，workspace/project 导航混在 AppLayout 中 | **中** | 缺少分组标签（section header），auth 子菜单硬编码在 AppLayout，无视觉分组 |
| 2 | 顶部区域克制，不抢主内容 | Topbar 56px，含 org selector + breadcrumb + 5 个 icon button | **中** | 功能齐全但信息密度偏高，5 个 icon button（搜索/通知/刷新/帮助/用户）平铺在右侧，缺乏层级 |
| 3 | 页面标题/说明/主操作按钮位置稳定 | 有 PageHeader 组件（title + description + actions） | **小** | 组件存在但使用不一致，部分页面内联写 header 而非复用组件 |
| 4 | 表格信息密度高但不拥挤 | 各页面表格实现方式不统一 | **大** | 无统一 DataTable 组件，行高/列对齐/排序指示器不一致 |
| 5 | 详情页结构清晰 | 无标准的 Detail Page 布局 | **大** | 缺少「左侧主信息 + 右侧元数据」的 Detail Layout 模式 |
| 6 | 表单简洁，说明文案精准 | shadcn/ui Form 组件齐全 | **小** | 组件能力足够，但 FormDescription 和 validation 样式不够统一 |
| 7 | 空状态/错误状态/加载状态清楚 | 有 inline empty state，无统一抽象 | **大** | 空状态在各页面 inline 实现，风格不一致；缺少组件级 error state；skeleton 使用不系统 |
| 8 | 操作按钮层级明确，危险操作谨慎 | Button 有 variant（default/destructive/outline/ghost） | **小** | variant 定义合理，但页面使用中主次不够清晰 |
| 9 | 中性色/细边框/轻量阴影/清晰层级 | 暖灰色系 + border-first + 几乎无阴影 | **大** | 色温偏暖，层级完全依赖边框，无阴影体系 |
| 10 | 避免渐变/花哨动画/卡片套卡片/模板感 | globals.css 有 tech-bg-pattern 渐变变量、部分卡片嵌套 | **中** | 存在残留的渐变/玻璃效果变量，部分页面 Card 嵌套 Card |

### 1.2 核心差距总结

**必须解决的 3 个结构性问题：**

1. **色彩与层级体系** — 从暖灰 border-first 转向冷蓝灰 shadow-first
2. **信息密度组件** — 缺少统一的 DataTable、DetailPage、EmptyState 抽象
3. **导航分组** — Sidebar 需要清晰的 section 分组和稳定的视觉层级

**可以保留的优势：**

- shadcn/ui 组件体系完整，Button/Form/Dialog 等核心组件就位
- PageHeader/PageLayout 已有合理抽象，只需微调
- CSS 变量架构成熟，token 替换即可全局变色
- Tailwind 配置完善，改 token 不需要改组件代码

---

## 2. 信息架构建议

### 2.1 当前信息架构

```
Org Level (workspace)
├── 所有项目 (grid/list)
├── Settings
│   ├── 角色管理
│   ├── API Keys
│   ├── 登录设置
│   └── 通用设置
└── Team / Profile

Project Level
├── 模型列表 (sidebar tree)
│   └── 模型详情/字段编辑
├── 枚举管理
├── 数据库集群
└── SQL 编辑器
```

### 2.2 建议的信息架构（保持路由不变）

路由结构不变，但在 Sidebar 中通过**分组标签 (section header)** 明确模块归属：

```
Sidebar (Org Level)
┌─────────────────────┐
│ [Org Logo] OrgName  │  ← org selector (dropdown)
├─────────────────────┤
│ 项目                │  ← section header (12px, uppercase, muted)
│   所有项目          │
├─────────────────────┤
│ 设置                │  ← section header
│   角色与权限        │
│   API Keys          │
│   登录设置          │
│   通用设置          │
└─────────────────────┘

Sidebar (Project Level)
┌─────────────────────┐
│ [← Back] ProjectName│  ← back to org, project name
├─────────────────────┤
│ 数据建模            │  ← section header
│   模型              │  ← expandable tree
│   枚举              │
├─────────────────────┤
│ 数据管理            │  ← section header
│   数据库集群        │
│   SQL 编辑器        │
└─────────────────────┘
```

### 2.3 关键原则

- **Section Header** 是纯视觉分组，不可点击，`12px uppercase text-muted-foreground`
- 每个 section 之间有 `16px` 间距 + `1px` 细分割线
- Stripe 的分组方式：功能域 > 使用频率 > 字母顺序
- Sidebar 底部保留用户头像 + 设置入口（不放在 topbar）

---

## 3. Layout 设计方案

### 3.1 整体布局结构

```
┌──────────────────────────────────────────────────────┐
│ Topbar (48px)                                         │
│ [OrgName ▾]              [Search]  [?]  [Avatar ▾]   │
├────────────┬─────────────────────────────────────────┤
│            │                                          │
│  Sidebar   │  Content Area                            │
│  (240px)   │  ┌────────────────────────────────────┐ │
│            │  │ PageHeader                          │ │
│  section   │  │ Title          [Secondary] [Primary]│ │
│  headers   │  ├────────────────────────────────────┤ │
│  nav items │  │                                     │ │
│            │  │ Page Content                        │ │
│            │  │ (table / form / cards / detail)     │ │
│            │  │                                     │ │
│            │  └────────────────────────────────────┘ │
│            │                                          │
├────────────┴─────────────────────────────────────────┤
```

### 3.2 与当前 Layout 的差异

| 维度 | 当前 | 新方案 | 改动量 |
|------|------|--------|--------|
| Topbar 高度 | 56px (h-14) | **48px (h-12)** | 小 — 改 AppLayout 一处 |
| Topbar 内容 | org selector + breadcrumb + 5 icons | org selector + search + help + avatar | 中 — 精简 topbar，移除 breadcrumb（已有 sidebar 上下文）、通知/刷新 |
| Sidebar 宽度 | 240px / 64px collapsed | **240px / 64px collapsed** (不变) | 无 |
| Content 背景 | `#fafafa` (bg-muted) | `#F6F8FA` (冷蓝灰) | 小 — CSS 变量 |
| Content padding | px-6 py-8 | **px-8 py-6** (更宽水平留白，略收紧垂直) | 小 — PageLayout |
| Content max-width | 7xl (default) | **保持 7xl**，内容区居左而非居中 | 小 |

### 3.3 关键 Layout 决策

**Q: Breadcrumb 去哪了？**
Stripe Dashboard 不使用 breadcrumb。导航上下文完全由 sidebar 的选中态 + 页面标题传达。当前 ModelCraft 的 breadcrumb 放在 topbar 中间，占用了宝贵的水平空间。建议移除，但保留 Project Level 的 "← Back to projects" 按钮在 sidebar 顶部。

**Q: 通知和刷新按钮去哪了？**
Stripe 的 topbar 只有三个元素：左侧品牌/导航、中间搜索、右侧用户菜单。通知功能如果业务需要，放入 sidebar 底部或用户菜单中，不占用 topbar 空间。刷新按钮不应存在于 modern SPA 中 — 数据应该自动保持最新或由操作触发刷新。

---

## 4. 区域规范

### 4.1 Sidebar

```
┌─────────────────────────┐
│                          │
│  [Logo] Organization ▾   │  ← 48px height, 16px padding
│                          │
├──────────────────────────┤  ← 1px border-bottom (#E3E8EE)
│                          │
│  SECTION HEADER          │  ← 12px, font-weight 500, text-[#8792A2]
│                          │     uppercase, letter-spacing 0.05em
│  ○ Nav Item              │     padding-left 12px, margin-top 24px
│  ● Nav Item (active)     │
│  ○ Nav Item              │  ← 36px height, 12px H padding
│                          │     font-size 14px, font-weight 400
│  SECTION HEADER          │     active: font-weight 500
│                          │     active: bg rgba(79,70,229,0.08)
│  ○ Nav Item              │     active: left 3px border #4F46E5
│  ○ Nav Item              │     hover: bg rgba(0,0,0,0.04)
│                          │     icon: 16px, stroke-width 1.5
│                          │     icon-text gap: 12px
│                          │     border-radius: 6px
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ │
│                          │
│  [Avatar] UserName       │  ← sidebar 底部，48px height
│           Settings ⚙     │     用户菜单入口
│                          │
└──────────────────────────┘

宽度: 240px (expanded) / 64px (collapsed, icon only)
背景: #FFFFFF
右边框: 1px solid #E3E8EE
过渡: width 200ms ease-in-out
```

**与当前差异：**
- 新增 Section Header（当前无分组）
- 选中态从 `#DADEE5` 背景 → `rgba(79,70,229,0.08)` + 左侧 indigo 条
- 导航项高度从无固定值 → 标准化 36px
- 用户信息从 topbar 移至 sidebar 底部

### 4.2 Topbar (Header)

```
┌──────────────────────────────────────────────────────────┐
│  [≡]  ModelCraft         🔍 Search...        [?] [👤▾]  │
│  12px  16px gap           flex-1 max-w-480     gap-8     │
└──────────────────────────────────────────────────────────┘

高度: 48px
背景: #FFFFFF
下边框: 1px solid #E3E8EE
左侧: sidebar toggle (mobile only) + 品牌名 or 当前 Org
中间: Search input (placeholder 样式, ⌘K 快捷键提示)
右侧: help icon + user avatar dropdown
内边距: 0 16px (左右)
```

**设计决策：**
- 从 56px → 48px，节省 8px 给内容区
- 移除 breadcrumb — sidebar 选中态已提供位置上下文
- 移除通知 icon、刷新 icon — 精简至核心三元素
- Search 居中，`max-width: 480px`，`flex: 1`
- Search 样式: `bg-[#F6F8FA] border-none rounded-md h-8 text-sm placeholder:text-[#8792A2]`

### 4.3 PageHeader

```
┌──────────────────────────────────────────────────────────┐
│  Page Title                    [Outline Btn] [Primary Btn]│
│  Optional description text                                │
└──────────────────────────────────────────────────────────┘

上间距: 0 (紧贴 content area 顶部)
下间距: 24px (margin-bottom)
标题: 20px, font-weight 600, color #1A1F36, letter-spacing -0.01em
描述: 14px, font-weight 400, color #697386, margin-top 4px
操作区: flex gap-8px, 垂直居中对齐标题行

布局: flex justify-between items-start
```

**与当前 PageHeader 差异：**
- 标题从可选大小（h1 默认 32px）→ 固定 20px
- 底部 margin 从 32px (mb-8) → 24px
- 操作按钮与标题**同行**（当前已如此，保持）
- 不使用下边框分割（当前 `bordered` prop）— 用间距替代线条

### 4.4 Content Area

```
背景: #F6F8FA
内边距: 32px horizontal (px-8), 24px vertical (py-6)
内容最大宽: 保持 max-w-7xl，但不居中 (text-left)
overflow: auto (y-axis)
```

**内容区层级模型：**
```
#F6F8FA  ← 页面背景（冷蓝灰）
  └ #FFFFFF + shadow-md  ← 内容卡片（白色，阴影浮起）
      └ #F6F8FA  ← 嵌套背景区域（如表格 header，但极少使用）
```

只有两层。绝不出现 Card 嵌套 Card。

---

## 5. 组件规范

### 5.1 DataTable

```
┌──────────────────────────────────────────────────────────┐
│ ┌──────────────────────────────────────────────────────┐ │
│ │  Toolbar                                              │ │
│ │  [Search input]              [Filter ▾] [Export ▾]    │ │
│ ├──────────────────────────────────────────────────────┤ │
│ │  Name ↕      Type      Status      Updated     ···   │ │  ← header row
│ ├──────────────────────────────────────────────────────┤ │  ← 1px #E3E8EE
│ │  User Model  GraphQL   ● Active    2h ago       ⋮   │ │  ← data row
│ │  Order       REST      ● Active    1d ago       ⋮   │ │
│ │  Product     GraphQL   ○ Draft     3d ago       ⋮   │ │
│ ├──────────────────────────────────────────────────────┤ │
│ │  ← 1 2 3 ... 10 →              Showing 1-20 of 156  │ │  ← footer
│ └──────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘

容器: bg-white, rounded-lg (8px), shadow-md, 无外边框
Header row:
  - 高度: 40px
  - 背景: #F6F8FA (微灰)
  - 文字: 12px, font-weight 500, color #697386, uppercase
  - 上圆角: 8px (与容器一致)
  - 分割线: bottom 1px #E3E8EE

Data row:
  - 高度: 48px (comfortable) 或 40px (dense mode)
  - 背景: transparent (白色继承自容器)
  - 文字: 14px, font-weight 400, color #1A1F36
  - 次要列: 13px, color #697386
  - 分割线: bottom 1px #E3E8EE (最后一行无)
  - Hover: bg rgba(0,0,0,0.02)
  - 文字对齐: text left, numbers right, status center

Action column:
  - ⋮ more icon (ghost button, 出现在 hover)
  - dropdown menu for row actions

Footer:
  - 高度: 44px
  - 分割线: top 1px #E3E8EE
  - Pagination 左侧, count 右侧
  - 文字: 13px, color #697386

Toolbar (可选, 在表格上方):
  - 不嵌套在卡片内 — 直接在 PageHeader 和 Table 之间
  - Search: inline input, 无边框, bg #F6F8FA
  - Filters: outline buttons / dropdown
```

### 5.2 Form

```
标准 Form 布局:
┌──────────────────────────────────────┐
│  Form Section Title (16px/600)       │
│  Section description (13px/400)      │
│                                      │
│  Label *                             │  ← 13px, font-weight 500, #1A1F36
│  ┌────────────────────────────────┐  │
│  │ Input value                    │  │  ← h-9 (36px), rounded-md (6px)
│  └────────────────────────────────┘  │     border 1px #E3E8EE
│  Helper text or validation error     │     focus: border #4F46E5, ring 3px rgba(79,70,229,0.1)
│                                      │
│  Label                               │  ← field 间距: 20px (gap-5)
│  ┌────────────────────────────────┐  │
│  │ Input value                    │  │
│  └────────────────────────────────┘  │
│                                      │
│                 [Cancel]  [Save]      │  ← 右对齐, gap-8px
│                 outline    primary    │     Cancel=ghost/outline, Save=primary
└──────────────────────────────────────┘

表单宽度: max-w-lg (512px) 或 max-w-xl (576px)
              不占满整个内容区 — 窄表单更易扫读

Label: 13px, font-weight 500, color #1A1F36, margin-bottom 6px
Input: 14px, font-weight 400, color #1A1F36
       placeholder: color #8792A2
       disabled: bg #F6F8FA, color #8792A2
Helper: 12px, color #697386, margin-top 4px
Error:  12px, color #EF4444, margin-top 4px, 替换 helper

Required indicator: * 号紧跟 label，color #EF4444

Section 间距: 32px (如果 form 有多个 section)
```

### 5.3 Modal (Dialog)

```
┌─────────── Backdrop: rgba(0,0,0,0.4) ───────────────┐
│                                                       │
│    ┌───────────────────────────────────┐              │
│    │  Modal Title              [×]     │  ← header    │
│    │  Optional description              │    48px      │
│    ├───────────────────────────────────┤  ← 1px line  │
│    │                                    │              │
│    │  Content                           │  ← body     │
│    │  (form fields or confirmation)     │    p-6      │
│    │                                    │              │
│    ├───────────────────────────────────┤  ← 1px line  │
│    │            [Cancel]  [Confirm]     │  ← footer   │
│    └───────────────────────────────────┘    py-4 px-6 │
│                                                       │
└───────────────────────────────────────────────────────┘

容器: bg-white, rounded-lg (8px), shadow-lg
宽度: sm (400px) / md (480px) / lg (640px)
动画: fade-in + scale(0.95 → 1), 150ms ease-out

Header:
  - padding: 24px 24px 16px
  - 标题: 16px, font-weight 600, color #1A1F36
  - 描述: 13px, color #697386, margin-top 4px
  - 关闭按钮: 右上角, ghost, 16px icon

Footer:
  - padding: 16px 24px
  - border-top: 1px #E3E8EE
  - 按钮右对齐, gap 8px

危险操作 Modal:
  - 标题前加 ⚠ icon (amber) 或 🔴 icon (red)
  - Confirm 按钮使用 destructive variant
  - 增加确认输入 (输入名称才能删除)
```

### 5.4 Detail Page

```
┌──────────────────────────────────────────────────────────┐
│  PageHeader: Model "User"           [Edit] [⋮ More]     │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────────────────┐  ┌────────────────────┐    │
│  │  Main Content (span 8)   │  │  Metadata (span 4) │    │
│  │                          │  │                     │    │
│  │  Section: Fields         │  │  Status    ● Active │    │
│  │  ┌ DataTable ──────────┐ │  │  Created   Apr 20   │    │
│  │  │ field list...       │ │  │  Updated   2h ago   │    │
│  │  └─────────────────────┘ │  │  Slug      user     │    │
│  │                          │  │  API ID    mdl_xxx  │    │
│  │  Section: Relations      │  │                     │    │
│  │  ┌ DataTable ──────────┐ │  │  Section: Tags     │    │
│  │  │ relation list...    │ │  │  [tag1] [tag2]     │    │
│  │  └─────────────────────┘ │  │                     │    │
│  │                          │  │                     │    │
│  └──────────────────────────┘  └────────────────────┘    │
│                                                           │
└──────────────────────────────────────────────────────────┘

Layout: CSS Grid, grid-cols-12, gap-6 (24px)
Main: col-span-8
Sidebar metadata: col-span-4

Main Content:
  - Section 标题: 16px, font-weight 600, margin-bottom 16px
  - Section 间距: 32px

Metadata Panel:
  - bg-white, rounded-lg, shadow-sm
  - padding: 20px
  - Key-value pairs:
    - Key: 13px, font-weight 500, color #697386
    - Value: 14px, font-weight 400, color #1A1F36
    - 行间距: 16px
  - Section 分割: 1px #E3E8EE, margin 16px 0

当模型详情页已有左侧 sidebar 模型树时:
  - 不使用右侧 metadata panel（空间不够）
  - 改为 metadata 放在 PageHeader 下方的 inline bar
  - 或使用 Tabs 切换 "Fields" / "Settings" / "Info"
```

### 5.5 EmptyState

```
┌──────────────────────────────────────────────────────────┐
│                                                           │
│              ┌─────┐                                      │
│              │  📄 │  ← 40px icon, color #8792A2          │
│              └─────┘     (不用彩色背景圆圈)               │
│                                                           │
│           No models yet                                   │
│           14px, font-weight 500, color #1A1F36            │
│                                                           │
│     Create your first model to get started.               │
│     13px, font-weight 400, color #697386                  │
│                                                           │
│             [+ Create Model]                              │
│             primary button, size sm                       │
│                                                           │
└──────────────────────────────────────────────────────────┘

容器: 不使用虚线 Card。直接在内容区中央，flex column, items-center
最大宽度: 360px (文字区域)
图标: 简单线框 icon，#8792A2 色，40px — 不用蓝色圆形背景（当前写法）
标题: 14px, font-weight 500 — 不用 lg/semibold（当前 text-lg font-semibold）
描述: 13px, color #697386
按钮: 标准 primary button, margin-top 16px
垂直间距: icon→title 12px, title→desc 4px, desc→button 16px
```

**与当前差异：**
当前 empty state 使用 `Card border-2 border-dashed` + `p-16` + `size-16 bg-blue-100 rounded-full` 包裹 icon。这是典型的"模板感" — 过度装饰。Stripe 的 empty state 非常朴素：一个灰色 icon + 一行文字 + 一个按钮，没有多余的装饰容器。

### 5.6 状态反馈组件

**Loading State (Skeleton):**
```
骨架屏形状应匹配实际内容:
  - 表格: 表头行 + 5-8 行矩形条
  - 卡片: 圆角矩形 + 内部 2-3 条短矩形
  - 详情页: 左侧长矩形 + 右侧短矩形组

颜色: bg #E3E8EE (静态), shimmer overlay linear-gradient
动画: shimmer from left to right, 1.5s infinite
```

**Error State (组件级):**
```
┌──────────────────────────────────────┐
│ ⚠  Something went wrong       [Retry]│
│    Unable to load models.             │
└──────────────────────────────────────┘

背景: rgba(239,68,68,0.04) — 极浅红
边框: 1px #FCA5A5 (red-300)
圆角: 6px
文字: 13px, icon + title 同行, description 下方
Retry 按钮: outline, size sm
```

**Toast (已有 sonner):**
保持现有 sonner 配置，调整样式匹配新色系即可。

### 5.7 Badge

```
Status badges:
  Active/Success:  bg rgba(5,150,105,0.08), text #059669, border none
  Warning/Draft:   bg rgba(217,119,6,0.08), text #D97706
  Error/Failed:    bg rgba(239,68,68,0.08), text #EF4444
  Info/Default:    bg rgba(79,70,229,0.08), text #4F46E5
  Neutral:         bg #F6F8FA, text #697386

尺寸: h-5 (20px), px-2, text 12px, font-weight 500
圆角: 4px (rounded-sm) — 不用 full rounded pill
```

### 5.8 Button 层级规范

```
层级从高到低:

1. Primary   — bg #4F46E5, text white, shadow-sm
               用于: 页面主操作 (Create, Save, Confirm)
               每个页面/modal 最多一个

2. Secondary — bg #F6F8FA, text #1A1F36, border 1px #E3E8EE
               用于: 次要操作 (Cancel, Export, Filter)

3. Ghost     — bg transparent, text #697386
               用于: 内联操作 (Edit, View, icon-only buttons)
               hover: bg rgba(0,0,0,0.04)

4. Destructive — bg #EF4444, text white
                  用于: 危险操作 (Delete, Remove)
                  仅在确认 modal 内使用，不在页面直接暴露

5. Link      — text #4F46E5, no bg, underline on hover
               用于: 导航性链接，not actions
```

---

## 6. 样板页选择

### 推荐: Workspace 页（所有项目）

**路径**: `/src/app/org/[orgName]/workspace/page.tsx`

**选择理由：**

1. **覆盖组件最全** — 包含 PageHeader + Toolbar (search + view toggle) + Card Grid/List + Empty State + Create Dialog
2. **是用户首屏** — 登录后的第一个页面，设计方向的"第一印象"
3. **改动自包含** — 不依赖其他页面的改动（不像 model 详情页依赖 sidebar 树改造）
4. **A/B 可比** — 改前改后可以直接截图对比
5. **Token 验证** — 能验证色彩、阴影、排版 token 的实际效果

**样板页改造范围：**
- PageHeader: 标题 20px, 移除多余 description
- Card 组件: 移除 border, 使用 shadow-md
- Toolbar: search input 样式调整
- Empty State: 简化为朴素风格
- Create Dialog: 表单风格调整
- Grid/List 切换的视觉一致性

### 备选: 角色管理页

如果需要表格组件的样板，`/settings/roles` 是更好的选择 — 它有标准的 CRUD 表格，能验证 DataTable 规范。

---

## 7. 分阶段改造路线

### Phase 0 — Design Token 替换（影响全局，改动最小）

**改动文件：** 仅 2 个文件

| 文件 | 改动 |
|------|------|
| `globals.css` | 替换所有 CSS 变量值（色彩、阴影 token） |
| `tailwind.config.ts` | 移除 Space Grotesk heading font，添加 shadow token，调整 letter-spacing |

**具体变更：**
```
CSS Variables:
  --background:        0 0% 100%       → 210 17% 98%        (#F6F8FA)
  --foreground:        224 71.4% 4.1%  → 228 45% 16%        (#1A1F36)
  --muted-foreground:  220 8.9% 46.1%  → 217 14% 49%        (#697386)
  --border:            220 13% 91%     → 213 19% 91%        (#E3E8EE)
  --primary:           221 83% 53%     → 239 84% 59%        (#4F46E5)
  --ring:              221 83% 53%     → 239 84% 59%        (match primary)
  --sidebar-accent:    215 20% 88%     → 239 84% 59% / 0.08 (indigo 8%)

Tailwind Config:
  fontFamily.heading:  删除 Space Grotesk（统一用 Inter）
  letterSpacing.tight: -0.01em (新增)
  boxShadow:           添加 stripe-sm / stripe-md / stripe-lg

globals.css:
  删除 --tech-bg-pattern, --tech-surface, --tech-surface-hover（残留渐变变量）
```

**效果：** 不改任何组件代码，全站色温从暖灰切换为冷蓝灰，品牌色从 Blue-600 变为 Indigo-600。这一步的产出应立即截图对比。

### Phase 1 — 核心组件调整

**改动文件：** 约 5-8 个 UI 组件文件

| 组件 | 文件路径 | 改动 |
|------|---------|------|
| Card | `ui/card.tsx` | `border` → 无 border, 添加 shadow class |
| Button | `ui/button.tsx` | 移除 default shadow, 字重 normal→500 |
| Badge | `ui/badge.tsx` | 调整颜色为 rgba 透明底色风格 |
| PageHeader | `features/layout/PageHeader.tsx` | 标题 text-2xl → text-xl, mb-8 → mb-6 |
| PageLayout | `features/layout/PageLayout.tsx` | padding 调整 |
| CardTitle | `ui/card.tsx` | text-2xl → text-base, 移除 font-heading |

**效果：** 组件层面的视觉品质统一。所有使用这些组件的页面自动受益。

### Phase 2 — 样板页改造 (Workspace)

**改动文件：** 1 个页面文件 + 相关子组件

| 改动 | 详情 |
|------|------|
| Workspace page | 移除 Card border-dashed empty state, 重构为朴素风格 |
| ProjectCard | 卡片悬停效果从 border-blue-100 → shadow 提升 |
| CreateDialog | 表单样式微调 |
| Toolbar | search input 样式 |

**效果：** 第一个完整落地新设计方向的页面，可作为所有后续页面的参考。

### Phase 3 — Sidebar 与 Header 改造

**改动文件：** 2-3 个 layout 文件

| 改动 | 详情 |
|------|------|
| AppLayout topbar | 高度 56→48, 移除 breadcrumb/通知/刷新 |
| AppLayout sidebar | 添加 section header, 调整选中态样式 |
| Sidebar nav items | 统一 36px 高度，添加 left border active indicator |

**效果：** 导航体验全面升级，Stripe 风格的分组和层级。

### Phase 4 — 页面逐步迁移

按使用频率排序：

1. **Settings > Roles** — 标准表格页，验证 DataTable 规范
2. **Project > 模型列表/详情** — 核心业务页，验证 Detail Page 规范
3. **Settings > API Keys** — 表格 + 操作页
4. **Settings > 登录设置** — 表单页，验证 Form 规范
5. **其他页面** — 按需迁移

### Phase 5 — 状态层完善

| 改动 | 详情 |
|------|------|
| 抽象 EmptyState 组件 | 替代各页面 inline 实现 |
| 抽象 ErrorState 组件 | 组件级错误展示 |
| Skeleton 体系 | 为主要页面添加 content-shaped skeleton |

### 改造量预估

| Phase | 改动文件数 | 风险 |
|-------|-----------|------|
| Phase 0 | 2 | 极低 — 纯 token 替换 |
| Phase 1 | 5-8 | 低 — 基础组件，影响广但改动小 |
| Phase 2 | 3-5 | 低 — 单页面，可独立验证 |
| Phase 3 | 2-3 | 中 — layout 改动影响所有页面 |
| Phase 4 | 每页 1-2 | 低 — 逐页迁移，互不干扰 |
| Phase 5 | 3-5 新组件 | 低 — 新增组件，不改已有代码 |

---

## 附录 A: 完整 Design Token 对照表

### 色彩 (Light Theme)

| Token | 当前 HSL | 当前 HEX | 新 HSL | 新 HEX | 说明 |
|-------|---------|---------|--------|--------|------|
| `--background` | `0 0% 100%` | `#FFFFFF` | `210 17% 98%` | `#F6F8FA` | 冷蓝灰背景 |
| `--foreground` | `224 71.4% 4.1%` | `#030712` | `228 45% 16%` | `#1A1F36` | 冷灰主文字 |
| `--card` | `0 0% 100%` | `#FFFFFF` | `0 0% 100%` | `#FFFFFF` | 不变 |
| `--card-foreground` | `224 71.4% 4.1%` | `#030712` | `228 45% 16%` | `#1A1F36` | 同 foreground |
| `--primary` | `221 83% 53%` | `#2563EB` | `239 84% 59%` | `#4F46E5` | Indigo-600 |
| `--primary-foreground` | `0 0% 100%` | `#FFFFFF` | `0 0% 100%` | `#FFFFFF` | 不变 |
| `--muted` | `220 14.3% 95.9%` | `#F1F5F9` | `210 17% 98%` | `#F6F8FA` | 同 background |
| `--muted-foreground` | `220 8.9% 46.1%` | `#64748B` | `217 14% 49%` | `#697386` | 冷灰次文字 |
| `--border` | `220 13% 91%` | `#E2E8F0` | `213 19% 91%` | `#E3E8EE` | 蓝灰边框 |
| `--input` | `220 13% 91%` | `#E2E8F0` | `213 19% 91%` | `#E3E8EE` | 同 border |
| `--ring` | `221 83% 53%` | `#2563EB` | `239 84% 59%` | `#4F46E5` | 同 primary |
| `--accent` | `214 95% 93%` | `#DBEAFE` | `239 84% 59% / 0.08` | `rgba(79,70,229,0.08)` | Indigo 8% |
| `--accent-foreground` | `221 83% 40%` | `#1E40AF` | `239 84% 47%` | `#4338CA` | Indigo-700 |
| `--destructive` | `0 84.2% 60.2%` | `#EF4444` | `0 84.2% 60.2%` | `#EF4444` | 不变 |
| `--selected` | `215 20% 88%` | `#DADEE5` | `239 84% 59% / 0.08` | `rgba(79,70,229,0.08)` | 同 accent |
| `--sidebar-accent` | `215 20% 88%` | `#DADEE5` | `239 84% 59% / 0.08` | `rgba(79,70,229,0.08)` | 同 accent |

### 阴影 (新增)

| Token | 值 | 用途 |
|-------|-----|------|
| `shadow-stripe-sm` | `0 1px 1px rgba(0,0,0,0.03), 0 1px 2px rgba(0,0,0,0.04)` | 输入框、列表项 |
| `shadow-stripe-md` | `0 2px 4px rgba(0,0,0,0.04), 0 4px 8px rgba(0,0,0,0.04), 0 1px 1px rgba(0,0,0,0.02)` | 卡片、面板 |
| `shadow-stripe-lg` | `0 4px 8px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.02)` | Modal、浮层 |

### 排版

| Token | 当前 | 新值 |
|-------|------|------|
| Font family (body) | Inter | Inter (不变) |
| Font family (heading) | Space Grotesk | **Inter** (统一) |
| Letter spacing (tight) | — | `-0.01em` (新增) |
| Font weight max | 700 | **600** (硬上限) |

---

## 附录 B: 不改什么

| 维度 | 说明 |
|------|------|
| 路由结构 | `/org/[orgName]/project/[projectSlug]/...` 不变 |
| 业务逻辑 | 所有 hooks、store、GraphQL 操作不变 |
| 接口协议 | GraphQL schema、BFF 不变 |
| 组件库 | 继续使用 shadcn/ui，不引入新 UI 框架 |
| 图标库 | 继续使用 Lucide React |
| 状态管理 | Zustand + React Query 不变 |
| Dark mode | 保留 .dark 类，但优先做好 light theme |
| 依赖 | 不引入 AG Grid、TanStack Table 等大型依赖（除非后续表格需求明确） |

---

## 附录 C: 当前代码中需要清理的 Anti-Patterns

以下是审查现有代码后发现的、与 Stripe 方向冲突的具体写法：

### C.1 ProjectCard — 过度装饰

```tsx
// ❌ 当前写法 (workspace/page.tsx)
className="group cursor-pointer
  border border-slate-200/60 bg-white/80 shadow-sm backdrop-blur-sm
  transition-all duration-300
  hover:scale-[1.01] hover:border-slate-300 hover:bg-white hover:shadow-lg"

// ✅ Stripe 方向
className="group cursor-pointer
  bg-white rounded-lg shadow-stripe-md
  transition-shadow duration-150
  hover:shadow-stripe-lg"
```

问题：`backdrop-blur-sm`（玻璃效果）、`bg-white/80`（半透明）、`hover:scale-[1.01]`（缩放动画）都是消费级装饰，与克制方向冲突。

### C.2 Badge — 渐变色和圆形

```tsx
// ❌ 当前写法
<Badge className="border-0 bg-gradient-to-r from-emerald-100 to-teal-100
  text-xs font-semibold text-emerald-700 shadow-sm" />

// 基础 badge 样式
"inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold"

// ✅ Stripe 方向
<Badge variant="success" />
// 基础样式: rounded-sm (4px), font-medium (500), 无 border, 无 shadow
// success: bg-emerald-500/8 text-emerald-600
```

问题：`bg-gradient-to-r`（渐变）、`rounded-full`（药丸形）、`shadow-sm`（badge 加阴影）过度装饰。

### C.3 CardTitle — 字号过大

```tsx
// ❌ 当前写法 (card.tsx)
className="font-heading text-2xl font-semibold leading-none tracking-tight"

// ✅ Stripe 方向
className="text-base font-semibold leading-tight tracking-tight"
```

问题：`text-2xl`（24px）在卡片上下文中过大，且使用了 `font-heading`（Space Grotesk），应统一为 Inter。

### C.4 Empty State — 蓝色圆圈装饰

```tsx
// ❌ 当前写法
<Card className="border-2 border-dashed border-slate-300 bg-white">
  <CardContent className="p-16 text-center">
    <div className="mx-auto mb-4 flex size-16 items-center justify-center
      rounded-full bg-blue-100">
      <Search className="size-8 text-blue-600" />
    </div>
    <p className="mb-2 text-lg font-semibold">未找到匹配的项目</p>

// ✅ Stripe 方向
<div className="flex flex-col items-center py-16">
  <Search className="size-10 text-[#8792A2] mb-3" />
  <p className="text-sm font-medium text-foreground">未找到匹配的项目</p>
  <p className="text-[13px] text-muted-foreground mt-1">尝试调整搜索条件或创建新项目</p>
  <Button size="sm" className="mt-4">创建项目</Button>
</div>
```

问题：`border-2 border-dashed`（虚线框）、`p-16`（过度留白）、`size-16 bg-blue-100 rounded-full`（彩色圆圈装饰）都是模板感来源。

### C.5 Dialog Overlay — 过暗

```tsx
// ❌ 当前写法 (dialog.tsx)
className="fixed inset-0 z-50 bg-black/80"

// ✅ Stripe 方向
className="fixed inset-0 z-50 bg-black/40"
```

问题：80% 黑色遮罩过于沉重，Stripe 使用约 40% 透明度。

### C.6 globals.css — 残留的渐变/玻璃变量

```css
/* ❌ 需要删除的变量 */
--tech-bg-pattern: linear-gradient(135deg, ...);
--tech-surface: 0 0% 100% / 0.7;      /* 半透明白 */
--tech-surface-hover: 0 0% 100% / 0.9; /* 半透明白 */
```

这些变量是早期"科技感"设计方向的残留，与 Stripe 的纯色/实色方向冲突。

### C.7 Button — 字重偏轻

```tsx
// ❌ 当前写法
"text-sm font-normal"  // font-weight 400

// ✅ Stripe 方向
"text-sm font-medium"  // font-weight 500
```

Stripe 的按钮文字使用 500 weight，比 400 略重，提供更好的可点击性暗示。

### C.8 Table Header — 需要统一

```tsx
// 当前写法 (table.tsx)
TableHead: "h-10 px-2 font-semibold text-sm text-muted-foreground"

// ✅ Stripe 方向
TableHead: "h-10 px-3 font-medium text-xs uppercase tracking-wider text-muted-foreground bg-[#F6F8FA]"
```

Stripe 表头使用 12px 大写字母 + 更宽字间距，视觉上更像"标签"而非"文本"。
