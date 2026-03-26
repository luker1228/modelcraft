---
name: design-token-extractor
description: >
  在实现任何 React 组件或页面之前，必须使用此 skill 从 prototype HTML/CSS 中提取
  design token，并将其转换为 Tailwind class 映射，作为实现约束。
  当用户要求实现、还原、翻译 prototype 为 React 代码时，立刻调用此 skill。
  当用户提到"原型"、"还原"、"实现页面"、"写 React"、"HTML 转 React"时，必须先执行此 skill。
---

# Design Token Extractor

将 prototype HTML/CSS 的设计规范提取为 Tailwind token 映射，作为 React 实现的约束文档。

## 为什么需要这个流程

Prototype CSS 用具体像素值（`34px`、`13px`）定义组件规格，但 Tailwind 使用语义 class（`h-9`、`text-sm`）。
两套系统之间存在隐性误差：`h-9 = 36px ≠ 34px`，`text-sm = 14px ≠ 13px`。
不经过显式映射，AI 会按 Tailwind "默认感觉"选 class，导致字体、边距、高度系统性偏差。

**这个 skill 的作用**：在写任何 className 之前，先建立"原型规格 → Tailwind class"的精确对照表，让实现有据可依。

---

## 第一步：读取 Token 源文件

必须读取以下文件：

```
prototypes/shared/layout-styles.css      ← 组件级 token（btn、badge、form 等）
prototypes/shared/tailwind-base.css      ← 全局 token（颜色、字体变量）
```

如果当前页面有专属样式，还需读取对应的 prototype `<style>` 块：
- `prototypes/cluster/index.html` → cluster 页面组件
- `prototypes/org-settings/index.html` → org-settings 页面组件
- `prototypes/model-editor/index.html` → model-editor 页面组件
- 其他页面同理

---

## 第二步：建立 Tailwind Token 映射表

读完 CSS 后，按以下格式在脑中或输出中构建映射表。**这是实现的约束，不是建议。**

### 颜色 Token（来自 tailwind-base.css + layout-styles.css）

| CSS 变量 / 值 | 含义 | React 中使用 |
|---|---|---|
| `#2563eb` / `var(--primary)` | 主色蓝 | `bg-[#2563eb]` 或 `text-[#2563eb]` |
| `#1d4ed8` / `var(--primary-hover)` | hover 蓝 | `hover:bg-[#1d4ed8]` |
| `#dbeafe` / `var(--primary-light)` | 浅蓝背景 | `bg-[#dbeafe]` |
| `#111827` / `var(--text-primary)` | 主文字 | `text-gray-900` |
| `#6b7280` / `var(--text-secondary)` | 次要文字 | `text-gray-500` |
| `#9ca3af` / `var(--text-tertiary)` | 第三级文字 | `text-gray-400` |
| `#e5e7eb` / `var(--border)` | 边框 | `border-gray-200` |
| `#fafafa` / `var(--bg-primary)` | 页面背景 | `bg-gray-50` |
| `#ffffff` / `var(--bg-secondary)` | 卡片背景 | `bg-white` |
| `#dadee5` / `var(--selected)` | 选中态背景 | `bg-[#dadee5]` |
| `#059669` / `var(--success)` | 成功绿 | `text-[#059669]` |
| `#ecfdf5` / `var(--success-light)` | 成功浅背景 | `bg-[#ecfdf5]` |
| `#ef4444` / `var(--destructive)` | 危险红 | `text-red-500` |
| `#d97706` / `var(--warning)` | 警告橙 | `text-[#d97706]` |

### 组件 Token（来自 layout-styles.css，**精确值，不能随意替换 Tailwind 标准 class**）

#### 按钮（.btn-primary / .btn-secondary / .btn-ghost）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| height | 36px | `h-9` ✓ (36px，刚好对应) |
| padding-x | 16px | `px-4` |
| font-size | 14px | `text-sm` |
| font-weight | 500 | `font-medium` |
| border-radius | 6px | `rounded-[6px]` 或 `rounded-md` (8px，接近) |
| icon size | 16px | `w-4 h-4` |
| icon stroke-width | 1.5 | `strokeWidth={1.5}` |
| gap (icon + text) | 8px | `gap-2` |

> 注意：`layout-styles.css` 的 `.btn-primary` height 是 36px，但各 prototype `<style>` 中 `.btn` 可能是 32px。**以实际读取的 CSS 为准，不要凭记忆。**

#### 输入框（.form-input / input[type="text"]）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| height（form-input） | 34px | `h-[34px]`（不能用 h-8=32px 或 h-9=36px） |
| padding-x | 10px | `px-[10px]` 或 `px-2.5` |
| font-size | 13px | `text-[13px]`（不能用 text-sm=14px） |
| border-radius | 6px | `rounded-[6px]` |
| border-color | `#e5e7eb` | `border-gray-200` |
| height（通用 input） | auto（padding: 8px 12px） | `py-2 px-3` |
| font-size（通用） | 14px | `text-sm` |

#### 卡片（.card）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| background | white | `bg-white` |
| border | 1px solid #e5e7eb | `border border-gray-200` |
| border-radius | 8px | `rounded-lg` (8px ✓) |
| overflow | hidden | `overflow-hidden` |

#### 卡片 Header（.card-header）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| padding | 16px 20px | `px-5 py-4` |
| border-bottom | 1px solid #e5e7eb | `border-b border-gray-200` |
| display | flex, align-items: center | `flex items-center` |
| gap | 8px | `gap-2` |
| title font-size | **14px**（不是 20px！） | `text-sm` |
| title font-weight | 600 | `font-semibold` |
| title color | #111827 | `text-gray-900` |

> ⚠️ 常见错误：AI 常把 card-header-title 渲染为 `text-xl`（20px），原型实际是 `14px`。

#### 卡片 Body（.card-body）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| padding | 20px | `p-5` |

#### 卡片 Footer（.card-footer）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| padding | 12px 20px | `py-3 px-5` |
| border-top | 1px solid #e5e7eb | `border-t border-gray-200` |
| background | #fafafa | `bg-gray-50` |
| justify | flex-end | `flex justify-end` |
| gap | 8px | `gap-2` |

#### 表单行（.form-row）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| display | flex, align-items: start | `flex items-start` |
| gap | **24px**（不是 16px！） | `gap-6` |
| margin-bottom | **24px**（不是 16px！） | `mb-6` |
| label 列宽 | 38% | `w-[38%] flex-none` |

> ⚠️ 常见错误：AI 常用 `gap-4`（16px），原型实际是 `gap-6`（24px）。

#### 表单 Label（.form-label）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| font-size | 14px | `text-sm` |
| font-weight | 500 | `font-medium` |
| color | #111827 / #374151 | `text-gray-700` |

#### 表单 Hint（.form-hint）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| font-size | 12px | `text-xs` |
| color | #9ca3af | `text-gray-400` |
| margin-top | 4px | `mt-1` |

#### Badge（.badge）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| padding | 4px 12px | `py-1 px-3` |
| border-radius | 4px | `rounded` |
| font-size | 12px | `text-xs` |
| font-weight | 500 | `font-medium` |

#### 侧边栏导航项（.sidebar-nav-item）

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| padding | 8px 12px | `py-2 px-3` |
| border-radius | 6px | `rounded-[6px]` |
| font-size | 14px | `text-sm` |
| font-weight | 500 / active:600 | `font-medium` / `font-semibold` |
| gap | 12px | `gap-3` |
| icon size | 16px | `w-4 h-4` |
| icon stroke-width | 1.5 | `strokeWidth={1.5}` |

#### Layout 尺寸

| 属性 | 原型值 | Tailwind class |
|---|---|---|
| topbar height | 56px | `h-14` |
| sidebar width (expanded) | 200px | `w-[200px]` |
| sidebar width (collapsed) | 64px | `w-16` |

---

## 第三步：实现规则

拿到 token 映射表后，实现 React 时遵守以下规则：

1. **精确高度优先**：如果原型值不是 Tailwind 标准尺寸（如 34px），使用 `h-[34px]` 而非就近取整。
2. **字体大小精确**：13px 用 `text-[13px]`，不能用 `text-sm`（14px）。14px 才能用 `text-sm`。
3. **间距不猜**：form-row 的 gap 是 24px（`gap-6`），不是 16px（`gap-4`）。不确定就回查 CSS。
4. **card-header-title 是 text-sm**：不是 text-xl，不是 text-lg，是 14px 的小标题。
5. **icon 统一 strokeWidth={1.5}**：所有 Lucide icon 加 `strokeWidth={1.5}`。
6. **颜色用具体值**：品牌色用 `bg-[#2563eb]`，不要用 `bg-blue-600`（值不同）。

---

## 第四步：实现前输出 Checklist

在开始写 className 前，先输出以下 checklist 确认关键值：

```
✅ 已读取 layout-styles.css
✅ 已读取 prototype <style> 块（如有）
📐 form-input height: 34px → h-[34px]
📐 btn height: 36px → h-9
📐 form-row gap: 24px → gap-6
📐 card-header-title: 14px → text-sm font-semibold
📐 icon stroke: 1.5 → strokeWidth={1.5}
```

---

## 快速参考：常见错误对照

| ❌ 错误写法 | ✅ 正确写法 | 原因 |
|---|---|---|
| `text-xl font-semibold` (card title) | `text-sm font-semibold` | 原型 14px，不是 20px |
| `h-9` (form-input) | `h-[34px]` | h-9=36px，原型是 34px |
| `gap-4` (form-row) | `gap-6` | 原型 24px，不是 16px |
| `text-sm` (form-input) | `text-[13px]` | text-sm=14px，原型是 13px |
| `bg-blue-600` | `bg-[#2563eb]` | 颜色值不同 |
| `strokeWidth` 未设置 | `strokeWidth={1.5}` | 原型 stroke-width: 1.5 |
| `px-5 py-4` (card-header) | `px-5 py-4` | ✓ 正确（16px 20px） |
| `py-3 px-5` (card-footer) | `py-3 px-5` | ✓ 正确（12px 20px） |
