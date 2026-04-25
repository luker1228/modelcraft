# ModelCraft UI Design Direction: Stripe-Inspired Redesign

> **Status**: Proposal — 产品定位与设计方向文档，不涉及代码变更
> **Date**: 2025-04-25
> **Reference**: dashboard.stripe.com

---

## 1. 为什么要转向 Stripe 方向

### 1.1 产品定位对齐

ModelCraft 是一个**数据建模与管理平台**，核心用户是开发者和技术型产品经理。这与 Stripe 的用户画像高度重合：

| 维度 | Stripe | ModelCraft |
|------|--------|------------|
| 用户 | 开发者、技术 PM | 开发者、技术 PM |
| 核心操作 | 配置、监控、管理 | 建模、配置、管理 |
| 数据密度 | 高（交易、日志、报表） | 高（模型、字段、枚举、关系） |
| 信任需求 | 极高（金融数据） | 高（数据库 schema 决策） |
| 使用频率 | 日常工具 | 日常工具 |

Stripe Dashboard 的设计语言传递的核心信息是：**"你可以信任这个工具来做重要的决策"**。这正是 ModelCraft 需要的。

### 1.2 当前设计的不足

ModelCraft 当前 UI 是标准的"Tailwind + shadcn/ui 开箱"风格，功能完整但缺乏辨识度：

- **色温偏暖灰** — 与 Stripe 的冷蓝灰相比，少了"精密工具"的感觉
- **边框定义层级** — 大量使用 `border` 划分区域，视觉上偏"网格化"
- **字重范围过宽** — 300-700 的 weight 分布缺乏节制感
- **标题过大** — 32px h1 在 dashboard 场景下显得"内容稀薄"
- **缺少品牌色记忆点** — Blue-600 是 Tailwind 默认色，没有品牌独特性

---

## 2. Stripe 设计语言核心解构

### 2.1 设计哲学：Restraint（克制）

Stripe 的设计系统建立在一个核心原则上：**通过移除而非添加来创造高级感**。

具体表现为：

- **色彩克制** — 整个界面只有一个强调色 `#635BFF`，且仅用于按钮、链接、选中态
- **字重克制** — 只用 400/500/600 三个 weight，绝不使用 700+
- **阴影克制** — 最大阴影的 opacity 仅 0.06，多层叠加模拟真实光照
- **圆角克制** — 主要使用 6-8px，避免过度圆润的"消费级"感觉
- **动效克制** — 功能性过渡，不做装饰性动画

### 2.2 色彩体系

```
┌─────────────────────────────────────────────────┐
│  背景层     #F6F8FA   冷蓝灰，仅比白色暗 2%     │
│  内容层     #FFFFFF   纯白卡片，靠阴影浮起       │
│  文字主     #1A1F36   冷灰+蓝底色，沉稳权威     │
│  文字次     #697386   中性冷灰                   │
│  文字三     #8792A2   更浅冷灰                   │
│  边框       #E3E8EE   蓝灰边框（非暖灰）         │
│  强调色     #635BFF   Indigo，唯一亮色           │
│  强调悬停   #7A73FF   微亮                       │
│  强调底色   rgba(99,91,255,0.08)  8% 透明度     │
└─────────────────────────────────────────────────┘
```

**关键洞察**：所有中性色都带有蓝色底色（blue undertone），而非暖灰。这是整体"冷感/精密感"的来源。

### 2.3 阴影体系 — 核心差异点

Stripe **不用边框来定义卡片**，而是用**极其微妙的多层阴影**：

```css
/* Small — 列表项、输入框 */
box-shadow: 0 1px 1px rgba(0,0,0,0.03), 
            0 1px 2px rgba(0,0,0,0.04);

/* Medium — 卡片、面板 */
box-shadow: 0 2px 4px rgba(0,0,0,0.04), 
            0 4px 8px rgba(0,0,0,0.04), 
            0 1px 1px rgba(0,0,0,0.02);

/* Large — 模态框、浮层 */
box-shadow: 0 4px 8px rgba(0,0,0,0.04), 
            0 8px 24px rgba(0,0,0,0.06), 
            0 1px 2px rgba(0,0,0,0.02);
```

这种"shadow-first"而非"border-first"的方式，让界面有**深度感但不压抑**。

### 2.4 排版体系

```
字号范围：12px → 13px → 14px → 16px → 20px → 24px（到顶）
字重范围：400（正文）/ 500（标签、导航）/ 600（标题、关键数字）
行高：1.3（紧凑）/ 1.5（正文）
字间距：-0.01em（微收紧）
```

**关键洞察**：最大字号仅 24px。在 dashboard 产品中，大标题是一种"浪费空间"的行为。Stripe 选择把空间留给数据，而不是留给标题。

---

## 3. ModelCraft 品牌色方向建议

### 3.1 方案 A：Indigo 路线（最接近 Stripe）

```
Primary:       #635BFF (Indigo)
Primary Hover: #7A73FF
Primary Muted: rgba(99, 91, 255, 0.08)
```

**优势**：开发者群体高度认可的"技术品牌"色（Stripe、Figma、Twitch 同系）
**风险**：直接使用同色会被认为是"Stripe 仿品"

### 3.2 方案 B：Deep Blue 路线（差异化）

```
Primary:       #4F46E5 (Indigo-600，偏蓝)
Primary Hover: #6366F1
Primary Muted: rgba(79, 70, 229, 0.08)
```

**优势**：在 Indigo 色系内做微调，保留冷调科技感同时有独立性
**风险**：较低

### 3.3 方案 C：Violet 路线（更独特）

```
Primary:       #7C3AED (Violet-600)
Primary Hover: #8B5CF6
Primary Muted: rgba(124, 58, 237, 0.08)
```

**优势**：更独特的品牌记忆点，保持冷调
**风险**：紫色在 B2B 场景中接受度略低于蓝/靛蓝

### 推荐：方案 B

在 Stripe 的色彩语言基础上做 15° 色相偏移，既能继承"精密工具"的感觉，又有清晰的品牌独立性。

---

## 4. 完整 Design Token 迁移方案

### 4.1 色彩迁移

| Token | 当前值 | 新值 | 变更说明 |
|-------|--------|------|----------|
| `--background` | `#fafafa` | `#F6F8FA` | 冷蓝灰背景 |
| `--card` / `--surface` | `#ffffff` | `#FFFFFF` | 不变，靠阴影浮起 |
| `--foreground` | `#111827` | `#1A1F36` | 冷灰+蓝底色 |
| `--muted-foreground` | `#6b7280` | `#697386` | 冷灰 |
| `--border` | `#e5e7eb` | `#E3E8EE` | 蓝灰边框 |
| `--primary` | `#2563eb` | `#4F46E5` | Indigo-600 |
| `--primary-foreground` | `#ffffff` | `#FFFFFF` | 不变 |
| `--accent` | (current) | `rgba(79,70,229,0.08)` | 极低透明度 |
| `--destructive` | `#ef4444` | `#EF4444` | 保持（红色通用） |
| `--success` | `#059669` | `#059669` | 保持（绿色通用） |
| `--warning` | `#d97706` | `#D97706` | 保持（琥珀色通用） |

### 4.2 排版迁移

| Token | 当前值 | 新值 |
|-------|--------|------|
| 字体 | Inter + Space Grotesk | Inter only（统一） |
| h1 | 32px/600 | 24px/600 |
| h2 | 24px/600 | 20px/600 |
| h3 | 16px/600 | 16px/600（不变） |
| body | 14px/400 | 14px/400（不变） |
| small | 13px/400 | 13px/400（不变） |
| caption | 12px/400 | 12px/400（不变） |
| max weight | 700 | **600**（硬上限） |
| letter-spacing | default | `-0.01em` |

### 4.3 阴影迁移

| Token | 当前值 | 新值 |
|-------|--------|------|
| `--shadow-sm` | (minimal/none) | `0 1px 1px rgba(0,0,0,0.03), 0 1px 2px rgba(0,0,0,0.04)` |
| `--shadow-md` | (minimal/none) | `0 2px 4px rgba(0,0,0,0.04), 0 4px 8px rgba(0,0,0,0.04), 0 1px 1px rgba(0,0,0,0.02)` |
| `--shadow-lg` | (minimal/none) | `0 4px 8px rgba(0,0,0,0.04), 0 8px 24px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.02)` |

### 4.4 圆角迁移

| Token | 当前值 | 新值 | 用途 |
|-------|--------|------|------|
| `--radius-sm` | 4px | 4px | Badge、小元素 |
| `--radius-md` | 6px | 6px | 按钮、输入框 |
| `--radius-lg` | 8px | 8px | 卡片（主要使用） |
| `--radius-xl` | 12px | 12px | 仅 Modal |

---

## 5. 界面模式变更

### 5.1 从 Border-First 到 Shadow-First

**Before（当前）:**
```
┌─────────────────────┐  ← 1px solid #e5e7eb
│  Card Content       │
└─────────────────────┘
```

**After（Stripe 风格）:**
```
╔═════════════════════╗  ← 无边框（或仅 1px #E3E8EE）
║  Card Content       ║  ← shadow-md 提供层级感
╚═════════════════════╝
```

核心变更：卡片默认使用 `shadow-sm` 或 `shadow-md`，减少显式边框。背景色 `#F6F8FA` 与卡片 `#FFFFFF` 的 2% 明度差配合阴影，自然形成层级。

### 5.2 从 Large Headers 到 Compressed Headers

**Before:**
```
┌─────────────────────────────────┐
│  Projects              [+ New]  │  ← 32px h1
│                                 │
│  ┌───┐ ┌───┐ ┌───┐             │
```

**After:**
```
┌─────────────────────────────────┐
│  Projects              [+ New]  │  ← 20px h2，与操作按钮同行
│  ┌───┐ ┌───┐ ┌───┐             │  ← 内容区紧凑上移
```

### 5.3 导航风格

保持 Sidebar 导航模式（ModelCraft 的模型树需要），但视觉调整为：

- Sidebar 背景：从当前主题色 → `#FFFFFF` 或 `#F6F8FA`
- 选中态：从高亮背景 → `rgba(79,70,229,0.08)` + 左侧 2px indigo 条
- 文字：全部使用 500 weight，选中态 600

---

## 6. 不做什么（Anti-Patterns）

在这次方向调整中，以下是明确要**避免**的：

| 不做 | 原因 |
|------|------|
| 渐变背景 | Stripe 零渐变，纯色更可信 |
| 动效装饰 | 仅功能性过渡（0.15s ease），无弹性/华丽动画 |
| 深色 sidebar | Stripe 全白/浅灰，保持统一 |
| 圆角 > 12px | 消费级感觉，与 B2B 定位冲突 |
| 字重 700+ | 破坏 Stripe 风格的核心克制感 |
| 多色强调 | 只有一个品牌强调色，语义色（红绿黄）仅用于状态 |
| 卡片悬停变色 | 当前 `hover:border-blue-100` → 改为微增阴影 |
| 大面积品牌色块 | 登录页 brand panel 可保留，但 dashboard 内不做大色块 |

---

## 7. 实施优先级建议

如果决定推进，建议分三个阶段：

### Phase 1 — Token 层（影响范围：全局，改动集中）

- 更新 `globals.css` 中的 CSS 变量
- 更新 `tailwind.config.ts` 中的主题色
- 统一字体为 Inter only
- 添加阴影 token

**效果**：仅改 token 文件，全站色温和基础感觉立即转变。

### Phase 2 — 组件层（影响范围：ui/ 目录）

- Card 组件：border → shadow
- Button 组件：更新 hover/focus 态
- Badge 组件：调整色值
- Sidebar 组件：视觉风格调整
- 字号压缩：h1 32→24, h2 24→20

**效果**：组件级别的视觉品质升级。

### Phase 3 — 页面层（影响范围：各页面）

- 逐页审查 padding/spacing
- 表格/列表的信息密度优化
- 空状态、加载态的视觉统一
- 登录/注册页品牌化调整

**效果**：端到端的 Stripe 级体验。

---

## 8. 对比参考

为了直观感受方向差异，以下是关键视觉元素的 Before/After 对比：

### 按钮

```
Before:  bg-[#2563eb] rounded-md font-medium
After:   bg-[#4F46E5] rounded-[6px] font-medium shadow-sm
         hover:bg-[#6366F1] transition-all duration-150
```

### 卡片

```
Before:  border border-gray-200 bg-white rounded-lg
After:   bg-white rounded-lg shadow-[0_2px_4px_rgba(0,0,0,0.04),0_4px_8px_rgba(0,0,0,0.04)]
         hover:shadow-[0_4px_8px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.06)]
```

### 页面背景

```
Before:  bg-[#fafafa]  (暖灰)
After:   bg-[#F6F8FA]  (冷蓝灰)
```

### 文字层级

```
Before:  text-[#111827] / text-[#6b7280] / text-[#9ca3af]
After:   text-[#1A1F36] / text-[#697386] / text-[#8792A2]
```

---

## 9. 总结

这不是一次"换皮"，而是一次**产品气质的升级**：

| 维度 | 从 | 到 |
|------|------|------|
| 气质 | "标准 SaaS 工具" | "精密开发者工具" |
| 色温 | 暖灰 | 冷蓝灰 |
| 层级表达 | 边框 | 阴影 |
| 品牌色 | Tailwind 默认蓝 | 独特 Indigo |
| 排版 | 标准 | 压缩、克制 |
| 核心哲学 | 功能完整 | **克制即高级** |

Stripe 的设计语言之所以被开发者群体认为是"最好的 dashboard"，本质不是因为它好看，而是因为它传递了**"这个工具值得信任"**的信号。ModelCraft 作为数据建模工具，同样需要这种信任感。
