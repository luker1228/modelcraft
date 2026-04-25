# ModelCraft UI 规范

> **优先级：最高** — 所有 UI 开发决策的单一真相源。与 DESIGN.md 保持同步，冲突时以本文件为准。
> **版本**: v2.0 (2025-04-25) — Stripe Dashboard 方向重构
> **状态**: ✅ 已确认

---

## 目录

1. [设计语言](#1-设计语言)
2. [Design Tokens](#2-design-tokens)
3. [排版](#3-排版)
4. [阴影与层级](#4-阴影与层级)
5. [间距与圆角](#5-间距与圆角)
6. [组件规范](#6-组件规范)
   - 6.1 [Button](#61-button)
   - 6.2 [Badge](#62-badge)
   - 6.3 [Table / DataTable](#63-table--datatable)
   - 6.4 [Toolbar & 搜索框](#64-toolbar--搜索框)
   - 6.5 [Sidebar & Navigation](#65-sidebar--navigation)
   - 6.6 [PageHeader](#66-pageheader)
   - 6.7 [Form & Input](#67-form--input)
   - 6.8 [Modal / Dialog](#68-modal--dialog)
   - 6.9 [EmptyState](#69-emptystate)
   - 6.10 [Skeleton & Loading](#610-skeleton--loading)
7. [动效](#7-动效)
8. [禁止事项（Anti-patterns）](#8-禁止事项)

---

## 1. 设计语言

**北极星参照：Stripe Dashboard**

核心原则：**通过移除而非添加来创造高级感。** 每个元素都要赚到它存在的权利。

| 维度 | 决策 |
|------|------|
| 色温 | 冷蓝灰（blue undertone），不用暖灰 |
| 层级表达 | Shadow-first：卡片靠多层微阴影浮起，不依赖粗边框 |
| 强调色 | 单一 Indigo `#4F46E5`，仅用于可交互元素 |
| 字体 | Inter only（Fira Code 用于技术标识符）|
| 最大字号 | 20px（dashboard 场景不用 32px 大标题）|
| 字重上限 | 600（`font-semibold`）|
| 动效 | 功能性，150ms ease-out，无装饰动画 |

---

## 2. Design Tokens

### 2.1 色彩（Light Mode）

#### 背景层级

| Token | HEX | 用途 |
|-------|-----|------|
| `--canvas` / `--bg` | `#F6F8FA` | 页面背景，所有面板的地基 |
| `--surface` | `#FFFFFF` | 卡片、侧边栏、Topbar、弹窗 |
| `--structure-muted` | `#EBEEF2` | 搜索框背景（比 canvas 深一档，凹入质感）|

#### 文字层级

| Token | HEX | 用途 |
|-------|-----|------|
| `--ink-deep` / `--foreground` | `#1A1F36` | 主要文字、标题、激活 thead |
| `--ink-mid` / `--muted-foreground` | `#697386` | 次要文字、描述、辅助信息 |
| `--ink-muted` / `--text-tertiary` | `#8792A2` | 占位符、图标静息态、section header |

#### 结构层级

| Token | HEX | 用途 |
|-------|-----|------|
| `--structure-border` / `--border` | `#E3E8EE` | 卡片边框、分割线、输入框边框 |
| `--structure-muted-border` | `#D8DDE5` | 搜索框边框（比 border 深一档）|

#### 强调色

| Token | HEX | 用途 |
|-------|-----|------|
| `--primary` | `#4F46E5` | Primary button、active nav 左条、tab 下划线、focus ring |
| `--primary-hover` | `#6366F1` | Primary button hover |
| `--primary-muted` / `--accent` | `rgba(79,70,229,0.08)` | 选中态背景、info badge 底色 |

#### 语义色

| Token | HEX | 用途 |
|-------|-----|------|
| `--success` | `#059669` | 活跃/健康状态 badge、成功 toast |
| `--warning` | `#D97706` | 警告状态 badge |
| `--destructive` | `#EF4444` | 错误状态、删除操作 |

### 2.2 CSS Variables（globals.css）

```css
:root {
  /* Background */
  --background:        210 17% 98%;   /* #F6F8FA */
  --card:              0 0% 100%;     /* #FFFFFF */
  --muted:             210 17% 98%;   /* #F6F8FA — 同 background */

  /* Text */
  --foreground:        228 45% 16%;   /* #1A1F36 */
  --muted-foreground:  217 14% 49%;   /* #697386 */

  /* Border */
  --border:            213 19% 91%;   /* #E3E8EE */
  --input:             213 19% 91%;   /* #E3E8EE */

  /* Primary (Indigo-600) */
  --primary:           239 84% 59%;   /* #4F46E5 */
  --primary-foreground: 0 0% 100%;
  --ring:              239 84% 59%;   /* focus ring 同 primary */

  /* Accent (selected state) */
  --accent:            239 84% 59% / 0.08;   /* rgba(79,70,229,0.08) */
  --accent-foreground: 239 84% 47%;          /* #4338CA */

  /* Sidebar */
  --sidebar-background: 0 0% 100%;
  --sidebar-accent:     239 84% 59% / 0.08;
  --sidebar-border:     213 19% 91%;

  /* Signal */
  --destructive:       0 84.2% 60.2%;
  --success:           160 84% 30%;
  --warning:           32 95% 44%;

  /* Radius */
  --radius: 0.5rem;  /* base = 8px，组件按需覆盖 */
}
```

---

## 3. 排版

### 3.1 字体

```
正文/UI:  Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif
技术标识: 'Fira Code', monospace
```

**Space Grotesk 已移除。** 所有场景统一用 Inter。

### 3.2 字号体系

| 级别 | 字号 | 字重 | 行高 | 用途 |
|------|------|------|------|------|
| Page Title | 20px | 600 | 1.3 | 页面标题，每页最多一个 |
| Section Title | 16px | 600 | 1.4 | 卡片标题、分组标题 |
| Body | 14px | 400 | 1.5 | 正文、表格数据单元格 |
| Label / Small | 13px | 400/500 | 1.4 | 表单说明、次要信息 |
| Table Header | 11px | 500 | 1.4 | thead uppercase 列标签 |
| Caption | 12px | 400 | 1.4 | 时间戳、元数据 |
| Mono | 12px | 400 | 1.6 | 技术标识符（Fira Code）|

### 3.3 规则

- **字重上限 600**。不使用 `font-bold`（700）及以上。
- **最小字号 12px**。低于 12px 不传达信息，是噪声。
- 表头（thead th）使用 `text-transform: uppercase; letter-spacing: 0.06em`，视觉上像"标签"而非文本。
- 技术标识符（model name、field slug、ID、SQL）一律使用 Fira Code mono。

---

## 4. 阴影与层级

### 4.1 阴影体系

```css
/* Stripe-style multi-layer subtle shadows */
--shadow-sm:  0 1px 1px rgba(0,0,0,0.03),
              0 1px 2px rgba(0,0,0,0.04);

--shadow-md:  0 2px 4px rgba(0,0,0,0.04),
              0 4px 8px rgba(0,0,0,0.04),
              0 1px 1px rgba(0,0,0,0.02);

--shadow-lg:  0 4px 8px rgba(0,0,0,0.04),
              0 8px 24px rgba(0,0,0,0.06),
              0 1px 2px rgba(0,0,0,0.02);
```

### 4.2 用途规则

| 层级 | Shadow | 场景 |
|------|--------|------|
| 静息卡片 | `shadow-md` | 页面内容卡片、DataTable 容器 |
| 按钮 | `shadow-sm` | Primary button 静息态 |
| Dropdown / Popover | `shadow-lg` | 浮层，脱离文档流 |
| Modal / Sheet | `shadow-lg` | 最高层级 |
| 卡片 hover | `shadow-lg`（升级）| 可点击卡片悬停反馈 |

**Flat-By-Default Rule**：Sidebar、Table row、Nav item 静息态均无阴影。阴影只出现在浮起或悬停状态。

### 4.3 卡片层级模型

```
#F6F8FA (页面背景)
  └── #FFFFFF + shadow-md (内容卡片)
```

**只有两层。禁止 Card 嵌套 Card。**

---

## 5. 间距与圆角

### 5.1 圆角

| Token | 值 | 用途 |
|-------|-----|------|
| `rounded-sm` | 4px | Badge、小元素 |
| `rounded-md` | 6px | Button、Input、Nav item |
| `rounded-lg` | 8px | Card、Table 容器（主要使用）|
| `rounded-xl` | 12px | 仅 Modal |

### 5.2 间距节奏

| 场景 | 间距 |
|------|------|
| 表单字段间距 | 16-20px (`gap-4` / `gap-5`) |
| Section 间距 | 24-32px |
| 卡片内边距 | 24px (`p-6`) |
| 紧凑内边距（Settings）| 16px (`p-4`) |
| Topbar 高度 | 48px |
| Table row 高度 | 48px（comfortable）/ 40px（dense）|
| Table header 高度 | 38-40px |
| Sidebar nav item 高度 | 36px |

---

## 6. 组件规范

### 6.1 Button

```
层级（高→低）:
Primary    bg:#4F46E5  text:white   shadow-sm   → 每个页面/modal 最多一个
Outline    bg:transparent  border:#E3E8EE  text:#1A1F36
Ghost      bg:transparent  text:#697386  hover:rgba(0,0,0,0.04)
Destructive  bg:#EF4444  text:white  → 仅在确认 Modal 内使用

尺寸:
default:  h-9 (36px), px-4, text-13px, font-weight 500
sm:       h-8 (32px), px-3, text-12px, font-weight 500
icon:     36×36, padding 0
```

**规则：**
- 每页只有一个 Primary button
- 危险操作（删除）的 Destructive 按钮仅在 AlertDialog 确认步骤中出现，不直接暴露在列表行
- 图标按钮（行内操作）使用 Ghost variant
- 字重统一 500（`font-medium`），不用 400 也不用 600

---

### 6.2 Badge

```
形状: rounded-sm (4px)  — 禁止 rounded-full pill 形
尺寸: h-5 (20px), px-2 (8px), text-11px, font-weight 500

变体:
  default/info:    bg rgba(79,70,229,0.08)   text #4F46E5
  success/active:  bg rgba(5,150,105,0.08)   text #059669
  warning:         bg rgba(217,119,6,0.08)   text #D97706
  destructive:     bg rgba(239,68,68,0.08)   text #EF4444
  neutral:         bg #F6F8FA  border #E3E8EE  text #697386
```

**规则：**
- 禁止 `bg-gradient-*` 渐变 badge
- 禁止 `shadow-sm` 在 badge 上
- 语义色只在传达状态时使用（不用绿色 button 代表"非危险"）

---

### 6.3 Table / DataTable

#### 结构

```
┌─ TableCard (bg:#FFF, rounded-lg, shadow-md) ──────────────┐
│  ┌─ Workspace Header ─────────────────────────────────────┐ │
│  │  [Icon] 模型名  · 记录数         [Export] [Schema] [+] │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌─ Toolbar (bg:#F6F8FA) ─────────────────────────────────┐ │
│  │  [🔍 搜索框 bg:#EBEEF2 border:#D8DDE5]  [筛选] [列]   │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌─ thead (bg:#FFF, border-bottom: 2px #E3E8EE) ──────────┐ │
│  │  □  ID ↕   用户名 ↕   邮箱   状态   角色   积分   时间 │ │
│  └────────────────────────────────────────────────────────┘ │
│     data rows (bg:transparent, 48px, hover:rgba(0,0,0,0.015))│
│  ┌─ Footer (border-top: 1px) ─────────────────────────────┐ │
│  │  共 1,284 条                    ‹ 1 2 3 ... 184 ›      │ │
│  └────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

#### 表头（thead）规范 ✅ 已确认

```css
thead tr {
  background: #FFFFFF;              /* 纯白，无灰色填充 */
  border-bottom: 2px solid #E3E8EE; /* 加粗下边框替代灰色背景 */
}
th {
  font-size: 11px;
  font-weight: 500;
  color: #1A1F36;                   /* 深色文字，非 muted */
  text-transform: uppercase;
  letter-spacing: 0.06em;
  height: 38px;
}
th:hover { color: #4F46E5; }        /* hover 染 indigo */
```

#### 数据行规范

```css
td {
  height: 48px;                     /* comfortable 模式 */
  font-size: 13px;
  color: #1A1F36;
  border-bottom: 1px solid #E3E8EE;
}
tr:last-child td { border-bottom: none; }
tr:hover td { background: rgba(0,0,0,0.015); }

/* 次要列 */
td.muted { color: #697386; }
/* 技术标识符 */
td.mono  { font-family: 'Fira Code', monospace; font-size: 12px; }
/* 数字列 */
td.num   { text-align: right; color: #697386; }
```

---

### 6.4 Toolbar & 搜索框

**已确认决策：**

```
Toolbar 背景:  #F6F8FA（与页面背景同色）
搜索框背景:    #EBEEF2（比 toolbar 深一档，形成凹入质感）
搜索框边框:    #D8DDE5（比 border 深一档）
搜索框高度:    30px
搜索框圆角:    6px (rounded-md)
```

```css
.toolbar {
  background: #F6F8FA;
  border-bottom: 1px solid #E3E8EE;
  padding: 10px 16px;
}
.toolbar-search {
  background: #EBEEF2;
  border: 1px solid #D8DDE5;
  border-radius: 6px;
  height: 30px;
  font-size: 13px;
  color: #8792A2;           /* placeholder */
}
```

**原则：** 搜索框必须比 toolbar 背景深一个层级，才有"输入框凹进去"的空间感。

---

### 6.5 Sidebar & Navigation

```
宽度: 240px (expanded) / 64px (icon-only collapsed)
背景: #FFFFFF
右边框: 1px solid #E3E8EE

Section Header:
  font-size: 11px
  font-weight: 500
  color: #8792A2
  text-transform: uppercase
  letter-spacing: 0.06em
  margin-top: 16px（首个 section 无 margin-top）

Nav Item:
  height: 36px
  padding: 0 10px
  border-radius: 6px
  border-left: 3px solid transparent
  gap: 10px (icon + text)
  font-size: 14px

  默认态: color #697386, icon #8792A2
  Hover:  background rgba(0,0,0,0.04), color #1A1F36
  Active: background rgba(79,70,229,0.08)
          color #4F46E5
          font-weight 500
          border-left-color #4F46E5

Icon: 16px, stroke-width 1.5

Sidebar 底部:
  用户头像 + 用户名 + 设置入口
  不在 Topbar 放用户信息
```

---

### 6.6 PageHeader

```
标题:   20px, font-weight 600, color #1A1F36, letter-spacing -0.01em
描述:   13px, font-weight 400, color #697386, margin-top 4px
操作区: flex, gap 8px, 垂直对齐标题行（flex items-start justify-between）
底部间距: margin-bottom 24px

规则:
- 不使用下边框分割 pageheader 和内容区（用间距替代）
- 描述文字可选，若内容自明则省略
```

---

### 6.7 Form & Input

```
Input:
  height: 36px
  border: 1px solid #E3E8EE
  border-radius: 6px
  background: transparent (#FFFFFF)
  font-size: 14px
  color: #1A1F36
  placeholder: color #8792A2

  focus:
    border-color: #4F46E5
    box-shadow: 0 0 0 3px rgba(79,70,229,0.1)

  disabled:
    opacity: 0.5
    cursor: not-allowed

Label:
  font-size: 13px
  font-weight: 500
  color: #1A1F36
  margin-bottom: 6px

Required mark: * 紧跟 label，color #EF4444

Helper text: 12px, color #697386, margin-top 4px
Error text:  12px, color #EF4444, margin-top 4px（替换 helper）

字段间距: 16-20px
Form 最大宽度: max-w-lg (512px) 或 max-w-xl (576px)，不占满内容区
```

---

### 6.8 Modal / Dialog

```
遮罩:   rgba(0,0,0,0.4)   — 不用 0.8，过暗
容器:   bg #FFFFFF, rounded-lg (8px→12px for modal only), shadow-lg
宽度:   sm:400px / md:480px / lg:640px
动画:   fade-in + scale(0.97→1), 150ms ease-out

Header:  padding 20px 24px 16px
  标题: 15px, font-weight 600, color #1A1F36
  描述: 13px, color #697386, margin-top 3px
  分割线: border-bottom 1px #E3E8EE

Body:    padding 20px 24px

Footer:  padding 14px 24px
  分割线: border-top 1px #E3E8EE
  按钮:  右对齐, gap 8px

危险确认 Modal:
  标题前加 ⚠ 图标
  Confirm 使用 Destructive variant
  Cancel 在左侧（先出现）
```

---

### 6.9 EmptyState

```
┌─────────────────────────┐
│                          │
│    [灰色 icon 40px]      │   color: #8792A2，无彩色圆形背景
│                          │
│    No models yet         │   14px, font-weight 500, #1A1F36
│    Create your first...  │   13px, color #697386, margin-top 4px
│                          │
│    [+ Create Model]      │   primary button sm, margin-top 16px
│                          │
└─────────────────────────┘

容器:  不用 Card，不用虚线边框
       直接 flex column items-center，内容区垂直居中
最大宽度: 360px（文字区域）

禁止:
  ❌ border-2 border-dashed（虚线框）
  ❌ size-16 bg-blue-100 rounded-full（彩色圆圈装饰）
  ❌ text-lg font-semibold（过大标题）
  ❌ p-16（过度 padding 撑开）
```

---

### 6.10 Skeleton & Loading

```
Skeleton:
  background: #E3E8EE（静态底色）
  shimmer: linear-gradient left→right, 1.5s infinite
  形状匹配实际内容（表格行形状 / 卡片形状）

Component-level Error State:
  ┌─────────────────────────────────────┐
  │ ⚠ 加载失败               [重试]    │
  │   无法加载数据，请稍后重试。         │
  └─────────────────────────────────────┘
  background: rgba(239,68,68,0.04)
  border: 1px solid #FCA5A5
  border-radius: 6px
  font-size: 13px

Toast (sonner):
  保留现有配置，样式跟随 token 更新
```

---

## 7. 动效

| 场景 | 时长 | 缓动 | 属性 |
|------|------|------|------|
| 按钮 hover/focus | 150ms | ease-out | color, background, box-shadow |
| 卡片 hover | 150ms | ease-out | box-shadow |
| Dropdown 出现 | 150ms | ease-out | opacity, transform(scale) |
| Modal 打开 | 150ms | ease-out | opacity, transform(scale) |
| Sheet/Drawer 打开 | 300ms | ease-in-out | transform(translateX) |
| Sheet/Drawer 关闭 | 200ms | ease-in-out | transform(translateX) |
| Accordion | 200ms | ease-out | height |
| Skeleton shimmer | 1500ms | linear | background-position |

**规则：**
- 所有动画使用 `transform` 和 `opacity`，禁止动画 `width/height/left/top`（触发 layout）
- 禁止 `infinite` 循环动画（skeleton shimmer 除外）
- 禁止 `hover:scale-*` 缩放（消费级）
- 使用 `transition-colors`、`transition-shadow`、`transition-opacity` 等具体属性，禁止 `transition-all`

---

## 8. 禁止事项

### 视觉装饰

| 禁止 | 原因 |
|------|------|
| `bg-gradient-*` 渐变 | 零渐变原则 |
| `backdrop-filter blur` 玻璃效果 | 消费级装饰 |
| `bg-white/80` 半透明白 | 用实色 `#FFFFFF` |
| `hover:scale-*` 缩放 | 消费级，干扰操作判断 |
| `rounded-full` badge/pill | 用 `rounded-sm` (4px) |
| Card 嵌套 Card | 只有两个层级：canvas + surface |
| 大面积品牌色块 | dashboard 内不做色块装饰 |
| Space Grotesk | 已移除，Inter only |

### 表格

| 禁止 | 正确做法 |
|------|----------|
| thead 灰色填充背景 | 纯白 + 2px 下边框 |
| 搜索框与 toolbar 同色 | 搜索框必须深一档 (#EBEEF2) |
| `font-bold` (700) thead | `font-medium` (500) |

### Empty State

| 禁止 | 正确做法 |
|------|----------|
| `border-2 border-dashed` 虚线框 | 无边框，直接居中内容 |
| `bg-blue-100 rounded-full` 彩色圆圈 | 纯灰色 icon `#8792A2` |
| `text-lg font-semibold` 大标题 | `text-sm font-medium` |

### 代码

| 禁止 | 正确做法 |
|------|----------|
| `text-gray-600`、`bg-slate-200` 等硬编码颜色 | 语义化 token |
| `transition-all` | `transition-colors`、`transition-shadow` 等 |
| `font-bold`、`font-extrabold`、`font-black` | `font-semibold` 封顶 |
| 直接修改 `contract/` 目录 | 通过 `front-contract-pull` skill |

---

## 参考

- [DESIGN.md](../../../DESIGN.md) — 设计系统 token 与组件规范（与本文保持同步）
- [PRODUCT.md](../../../PRODUCT.md) — 产品定位与设计原则
- [design-direction-stripe.md](./design-direction-stripe.md) — Stripe 方向分析与 token 迁移方案
- [ui-redesign-plan.md](./ui-redesign-plan.md) — 分阶段改造路线图
- Demo pages: `demo-roles.html`、`demo-model-editor.html`
