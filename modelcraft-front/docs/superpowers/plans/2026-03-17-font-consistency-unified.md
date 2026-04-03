# 字体一致性统一 Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 统一项目中所有字体权重、字体大小和文字颜色用法，确保与 STYLE.md 和 typography.ts 设计规范一致。

**Architecture:** 
- 将所有 `font-bold`(700) 和 `font-extrabold`(800) 替换为规范定义的 `font-semibold`(600)
- 将分散的 `text-gray-*` 颜色替换为 shadcn 语义化变量（`text-foreground`、`text-muted-foreground`）
- 删除非规范的视觉效果（如 `drop-shadow-md`、`bg-clip-text` 渐变）
- 保留 `text-5xl` 等大号字体（品牌展示页可用）

**Tech Stack:** Tailwind CSS, shadcn/ui CSS variables, TypeScript/TSX

---

## 映射规则

### 字体权重映射

| 当前用法 | 替换为 | 文件数 | 实例数 |
|---|---|---|---|
| `font-bold` (700) | `font-semibold` (600) | 7 | 11+ |
| `font-extrabold` (800) | `font-semibold` (600) | 1 | 3 |
| `font-black` (900) | — | 0 | 0 |

### 文字颜色映射

| 当前用法 | 对应含义 | 替换为 | 使用场景 |
|---|---|---|---|
| `text-gray-900` | 主要文本（标题、正文） | `text-foreground` | 所有标题、主内容 |
| `text-gray-700` | 次要主文本 | `text-foreground` 或 `text-muted-foreground` | 表单标签、主描述 |
| `text-gray-600` | 次要文本（描述） | `text-muted-foreground` | 副标题、次要描述 |
| `text-gray-500` | 辅助文本 | `text-muted-foreground` | 占位符、禁用文本、图标 |
| `text-gray-400` | 最淡文本 | `text-muted-foreground` | 装饰性图标、极弱信息 |

> **说明**：由于 `--muted-foreground` 在 globals.css 中对应 `220 8.9% 46.1%`（≈gray-500），所有 gray-* 变体统一映射到 `text-foreground` 或 `text-muted-foreground` 两档。`text-foreground` 用于主文本和标题，`text-muted-foreground` 用于次要和辅助文本。

---

## Chunk 1: 字体权重统一 - 高优先级文件

### 文件修改清单

- Modify: `src/app/auth/callback/page.tsx:172, 209, 250` — 替换 `font-extrabold` → `font-semibold`
- Modify: `src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx:121, 143, 161, 172, 202, 226, 235` — 替换 `font-bold` → `font-semibold`，删除 `drop-shadow-md` 和渐变效果
- Modify: `src/components/ui/editor-sidebar.tsx:259` — 替换 `font-bold` → `font-semibold`

#### Task 1.1: 修复 auth/callback/page.tsx 的 font-extrabold

**文件:** `src/app/auth/callback/page.tsx`

- [ ] **Step 1: 读取文件，确认 3 处 font-extrabold 的位置**

Run: 
```bash
grep -n "font-extrabold" /root/modelcraft_project/modelcraft-front/src/app/auth/callback/page.tsx
```

Expected output:
```
172:              <h2 className="mb-4 text-3xl font-extrabold text-foreground">
209:              <h2 className="mb-4 text-3xl font-extrabold text-foreground">
250:          <h2 className="mb-4 text-3xl font-extrabold text-foreground">
```

- [ ] **Step 2: 替换第一处（172 行）**

使用 Edit 工具，将：
```tsx
<h2 className="mb-4 text-3xl font-extrabold text-foreground">
```

替换为：
```tsx
<h2 className="mb-4 text-3xl font-semibold text-foreground">
```

- [ ] **Step 3: 替换第二处（209 行）**

同样替换 209 行的 `font-extrabold` → `font-semibold`

- [ ] **Step 4: 替换第三处（250 行）**

同样替换 250 行的 `font-extrabold` → `font-semibold`

- [ ] **Step 5: 验证替换**

Run:
```bash
grep -n "font-extrabold" /root/modelcraft_project/modelcraft-front/src/app/auth/callback/page.tsx
```

Expected: 无输出（已全部替换）

- [ ] **Step 6: 验证 font-semibold 已正确替换**

Run:
```bash
grep -n "text-3xl font-semibold" /root/modelcraft_project/modelcraft-front/src/app/auth/callback/page.tsx | grep "mb-4"
```

Expected: 3 行输出确认

- [ ] **Step 7: Commit**

```bash
git add src/app/auth/callback/page.tsx
git commit -m "fix: replace font-extrabold with font-semibold in auth callback page

Standardize font weight to match design spec (600 instead of 800) for consistency with STYLE.md and typography.ts guidelines."
```

---

#### Task 1.2: 修复 guide/page.tsx 的多个 font-bold 和视觉效果

**文件:** `src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx`

- [ ] **Step 1: 读取文件，确认所有 font-bold 位置**

Run:
```bash
grep -n "font-bold" /root/modelcraft_project/modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx
```

Expected: 7+ 处输出

- [ ] **Step 2: 检查第 121 行的 drop-shadow 和 font-bold**

Read lines 115-130 to confirm context

- [ ] **Step 3: 修复 121 行 — 移除 drop-shadow-md，替换 font-bold → font-semibold**

原代码（约 121 行）：
```tsx
<h1 className="mb-3 font-heading text-3xl font-bold leading-tight text-white drop-shadow-md md:text-4xl">
  欢迎使用 ModelCraft
</h1>
```

修改为：
```tsx
<h1 className="mb-3 font-heading text-3xl font-semibold leading-tight text-white md:text-4xl">
  欢迎使用 ModelCraft
</h1>
```

- [ ] **Step 4: 修复 143 行 — 替换 font-bold → font-semibold**

原：`<h2 className="mb-4 flex items-center gap-2 font-heading text-xl font-bold text-slate-800">`

修改为：`<h2 className="mb-4 flex items-center gap-2 font-heading text-xl font-semibold text-slate-800">`

- [ ] **Step 5: 修复 161 行 — 删除渐变效果和 font-bold**

原代码（约 161 行）使用 `bg-clip-text` 和 `text-transparent` 的渐变：
```tsx
<div className={`bg-gradient-to-r font-heading text-3xl font-bold ${getColorStyles(stat.color)} bg-clip-text text-transparent`}>
```

修改为（使用纯色，不用渐变）：
```tsx
<div className={`font-heading text-3xl font-semibold ${getColorStyles(stat.color)}`}>
```

- [ ] **Step 6: 修复 172 行 — 替换 font-bold → font-semibold**

原：`<h2 className="mb-6 font-heading text-2xl font-bold text-slate-800">开始指南</h2>`

修改为：`<h2 className="mb-6 font-heading text-2xl font-semibold text-slate-800">开始指南</h2>`

- [ ] **Step 7: 修复 202 行 — 替换 font-bold → font-semibold**

原：`<h3 className="font-heading text-xl font-bold text-slate-800">{step.title}</h3>`

修改为：`<h3 className="font-heading text-xl font-semibold text-slate-800">{step.title}</h3>`

- [ ] **Step 8: 修复 226 行 — 替换 font-bold → font-semibold**

原：`<h2 className="mb-6 font-heading text-2xl font-bold text-slate-800">快速链接</h2>`

修改为：`<h2 className="mb-6 font-heading text-2xl font-semibold text-slate-800">快速链接</h2>`

- [ ] **Step 9: 修复 235 行 — 替换 font-bold → font-semibold**

原：`<h3 className="mb-2 font-heading font-bold text-slate-800 transition-colors group-hover:text-blue-600">`

修改为：`<h3 className="mb-2 font-heading font-semibold text-slate-800 transition-colors group-hover:text-blue-600">`

- [ ] **Step 10: 验证替换完成**

Run:
```bash
grep -n "font-bold" /root/modelcraft_project/modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx
```

Expected: 无输出

- [ ] **Step 11: 验证 font-semibold 已到位**

Run:
```bash
grep -n "font-semibold" /root/modelcraft_project/modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx | head -5
```

Expected: 5+ 行输出确认

- [ ] **Step 12: 验证 drop-shadow 已移除**

Run:
```bash
grep -n "drop-shadow" /root/modelcraft_project/modelcraft-front/src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx
```

Expected: 无输出

- [ ] **Step 13: Commit**

```bash
git add src/app/org/[orgName]/projects/[projectSlug]/guide/page.tsx
git commit -m "fix: standardize font weights and remove non-spec visual effects in guide page

- Replace font-bold with font-semibold for all headings (align with design spec)
- Remove drop-shadow-md effect from hero heading (violates clean design principle)
- Remove gradient text effect (use solid colors per STYLE.md)"
```

---

#### Task 1.3: 修复 editor-sidebar.tsx 的 font-bold

**文件:** `src/components/ui/editor-sidebar.tsx`

- [ ] **Step 1: 读取 259 行附近的代码**

Run:
```bash
sed -n '255,265p' /root/modelcraft_project/modelcraft-front/src/components/ui/editor-sidebar.tsx
```

- [ ] **Step 2: 替换 font-bold → font-semibold**

原：`className="group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-bold uppercase tracking-widest text-slate-700..."`

修改为：`className="group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-semibold uppercase tracking-widest text-slate-700..."`

- [ ] **Step 3: 验证替换**

Run:
```bash
grep -n "font-bold" /root/modelcraft_project/modelcraft-front/src/components/ui/editor-sidebar.tsx
```

Expected: 无输出

- [ ] **Step 4: 验证 font-semibold 已到位**

Run:
```bash
grep -n "font-semibold" /root/modelcraft_project/modelcraft-front/src/components/ui/editor-sidebar.tsx | grep "text-xs"
```

- [ ] **Step 5: Commit**

```bash
git add src/components/ui/editor-sidebar.tsx
git commit -m "fix: replace font-bold with font-semibold in editor sidebar

Standardize section header text weight to design spec (600 instead of 700)."
```

---

## Chunk 2: 文字颜色统一 — 常见颜色替换

### 文件修改清单

- Modify: `src/app/login/page.tsx` — 替换 text-gray-* → text-foreground/text-muted-foreground
- Modify: `src/components/layout/AppLayout.tsx` — 替换 text-gray-* → text-foreground/text-muted-foreground
- Modify: `src/components/project/ProjectCard.tsx` — 替换 text-slate-* → text-foreground/text-muted-foreground
- Modify: `src/app/org-selector/page.tsx` — 替换 text-slate-* → text-foreground/text-muted-foreground
- Modify: `src/components/settings/layout.tsx` — 替换 text-foreground（已使用规范变量）
- 其他文件（DatabaseConfigFields、ErrorHistoryDialog 等）— 集中修复

#### Task 2.1: 修复 login/page.tsx 的颜色

**文件:** `src/app/login/page.tsx`

- [ ] **Step 1: 统计 text-gray-* 和 text-slate-* 分布**

Run:
```bash
grep -n "text-gray-\|text-slate-" /root/modelcraft_project/modelcraft-front/src/app/login/page.tsx | wc -l
```

Expected: 15+ 处

- [ ] **Step 2: 替换 text-gray-900 → text-foreground**

原文件中所有 `text-gray-900` 替换为 `text-foreground`（标题、主文本）

示例替换位置：
- 86 行：`<span className="text-2xl font-semibold text-foreground">`
- 103 行：`<h1 className="text-5xl font-semibold leading-tight text-foreground">`
- 121 行、163 行等的标题

- [ ] **Step 3: 替换 text-gray-600 → text-muted-foreground**

原文件中所有 `text-gray-600` 替换为 `text-muted-foreground`（次要描述、副文本）

示例替换位置：
- 117 行：描述文本
- 151 行：登录表单副标题
- 194 行、204 行：次要描述

- [ ] **Step 4: 替换 text-gray-500 → text-muted-foreground**

原文件中所有 `text-gray-500` 替换为 `text-muted-foreground`（辅助文本）

示例替换位置：
- 152 行、181 行：页脚文本

- [ ] **Step 5: 验证替换完成**

Run:
```bash
grep -n "text-gray-" /root/modelcraft_project/modelcraft-front/src/app/login/page.tsx
```

Expected: 无输出

- [ ] **Step 6: 检查 text-foreground 替换**

Run:
```bash
grep -n "text-foreground" /root/modelcraft_project/modelcraft-front/src/app/login/page.tsx | head -3
```

Expected: 至少 3 行输出确认

- [ ] **Step 7: Commit**

```bash
git add src/app/login/page.tsx
git commit -m "refactor: standardize text colors to shadcn semantic variables in login page

- Replace text-gray-900 with text-foreground (primary text/titles)
- Replace text-gray-600 with text-muted-foreground (secondary descriptions)
- Replace text-gray-500 with text-muted-foreground (auxiliary text)"
```

---

#### Task 2.2: 修复 components/layout/AppLayout.tsx 的颜色

**文件:** `src/components/layout/AppLayout.tsx`

- [ ] **Step 1: 统计 text-gray-* 分布**

Run:
```bash
grep -n "text-gray-" /root/modelcraft_project/modelcraft-front/src/components/layout/AppLayout.tsx
```

Expected: 8+ 处

- [ ] **Step 2: 替换 text-gray-900 → text-foreground**

位置：
- 243 行：`<span className="font-medium text-foreground">`（组织名称）
- 247 行：`<span className="text-foreground">`
- 337 行的 `hover:text-gray-900` → `hover:text-foreground`

- [ ] **Step 3: 替换 text-gray-500 → text-muted-foreground**

位置：
- 208 行：Search 图标 `text-muted-foreground`
- 233 行：Empty state `text-muted-foreground`
- 258 行、267 行、278 行、287 行的按钮图标 `text-muted-foreground`
- 323 行、337 行的导航项图标 `text-muted-foreground`

- [ ] **Step 4: 替换 text-gray-600 → text-muted-foreground**

位置：
- 323 行的 `text-gray-600 hover:...` → `text-muted-foreground hover:...`

- [ ] **Step 5: 验证替换**

Run:
```bash
grep -n "text-gray-" /root/modelcraft_project/modelcraft-front/src/components/layout/AppLayout.tsx
```

Expected: 无输出

- [ ] **Step 6: Commit**

```bash
git add src/components/layout/AppLayout.tsx
git commit -m "refactor: standardize text colors to shadcn semantic variables in AppLayout

- Replace text-gray-900 with text-foreground (primary navigation, breadcrumb)
- Replace text-gray-500/600 with text-muted-foreground (icons, secondary text, empty state)"
```

---

## Chunk 3: 组件库颜色统一 — 中等优先级文件

#### Task 3.1: 修复 project/ProjectCard.tsx

**文件:** `src/components/project/ProjectCard.tsx`

- [ ] **Step 1: 替换 text-slate-900 → text-foreground**

位置：
- 66 行：`<CardTitle className="text-lg font-semibold text-foreground..."`（卡片标题）

- [ ] **Step 2: 替换 text-slate-600 → text-muted-foreground**

位置：
- 69 行：`<CardDescription className="mt-1 line-clamp-2 text-sm text-muted-foreground">`
- 106 行：`<span className="text-muted-foreground">更新于...</span>`

- [ ] **Step 3: 替换 text-slate-500 → text-muted-foreground**

位置：
- 已在 106 行替换完成

- [ ] **Step 4: 验证替换**

Run:
```bash
grep -n "text-slate-" /root/modelcraft_project/modelcraft-front/src/components/project/ProjectCard.tsx
```

Expected: 无输出（应全部替换）

- [ ] **Step 5: Commit**

```bash
git add src/components/project/ProjectCard.tsx
git commit -m "refactor: standardize text colors to shadcn semantic variables in ProjectCard

- Replace text-slate-900 with text-foreground (card titles)
- Replace text-slate-600/500 with text-muted-foreground (descriptions, metadata)"
```

---

#### Task 3.2: 修复 org-selector/page.tsx 的颜色

**文件:** `src/app/org-selector/page.tsx`

- [ ] **Step 1: 检查文件中的颜色使用**

Run:
```bash
grep -n "text-slate-\|text-gray-\|text-foreground" /root/modelcraft_project/modelcraft-front/src/app/org-selector/page.tsx | head -5
```

- [ ] **Step 2: 验证 text-foreground 已在 256 行使用**

已在规范变量中，无需修改（本文件已部分符合规范）

- [ ] **Step 3: 确认无其他颜色问题**

Run:
```bash
grep -n "text-gray-\|text-slate-" /root/modelcraft_project/modelcraft-front/src/app/org-selector/page.tsx
```

Expected: 无输出（若有则修复）

---

## Chunk 4: 数据库配置和错误组件 — 低优先级文件

#### Task 4.1: 修复 database/DatabaseConfigFields.tsx

**文件:** `src/components/database/DatabaseConfigFields.tsx`

- [ ] **Step 1: 替换 text-gray-500 → text-muted-foreground**

位置：
- 61 行：Icon 颜色 `text-muted-foreground`
- 67 行：Description `text-muted-foreground`
- 113 行、160 行、197 行：Icon 颜色 `text-muted-foreground`

- [ ] **Step 2: 替换 text-gray-700 → text-foreground**

位置：
- 63 行：Field label `text-foreground`
- 115 行、162 行、199 行：Field label `text-foreground`

- [ ] **Step 3: 验证替换**

Run:
```bash
grep -n "text-gray-" /root/modelcraft_project/modelcraft-front/src/components/database/DatabaseConfigFields.tsx
```

Expected: 无输出

- [ ] **Step 4: Commit**

```bash
git add src/components/database/DatabaseConfigFields.tsx
git commit -m "refactor: standardize text colors in DatabaseConfigFields

- Replace text-gray-700 with text-foreground (labels)
- Replace text-gray-500 with text-muted-foreground (icons, descriptions)"
```

---

#### Task 4.2: 修复 error/ErrorHistoryDialog.tsx

**文件:** `src/components/error/ErrorHistoryDialog.tsx`

- [ ] **Step 1: 替换 text-gray-500 → text-muted-foreground**

位置：
- 90 行：Empty state `text-muted-foreground`
- 158 行：Timestamp `text-muted-foreground`

- [ ] **Step 2: 替换 text-gray-700 → text-foreground**

位置：
- 156 行：Error message `text-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/components/error/ErrorHistoryDialog.tsx
git commit -m "refactor: standardize text colors in ErrorHistoryDialog

- Replace text-gray-700 with text-foreground (error messages)
- Replace text-gray-500 with text-muted-foreground (empty state, timestamps)"
```

---

#### Task 4.3: 修复 ui/identity-form-section.tsx

**文件:** `src/components/ui/identity-form-section.tsx`

- [ ] **Step 1: 替换 text-gray-700 → text-foreground**

位置：
- 120 行：Form label `text-foreground`
- 462 行：Section title `text-foreground`

- [ ] **Step 2: 替换 text-gray-500 → text-muted-foreground**

位置：
- 127 行：Description `text-muted-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/components/ui/identity-form-section.tsx
git commit -m "refactor: standardize text colors in identity-form-section

- Replace text-gray-700 with text-foreground (labels, titles)
- Replace text-gray-500 with text-muted-foreground (descriptions)"
```

---

#### Task 4.4: 修复 error/GraphQLErrorDialog.tsx

**文件:** `src/components/error/GraphQLErrorDialog.tsx`

- [ ] **Step 1: 替换 text-gray-600 → text-muted-foreground**

位置：
- 135 行、142 行：Error details `text-muted-foreground`

- [ ] **Step 2: Commit**

```bash
git add src/components/error/GraphQLErrorDialog.tsx
git commit -m "refactor: standardize text color in GraphQLErrorDialog

- Replace text-gray-600 with text-muted-foreground (error details)"
```

---

## Chunk 5: 编辑器和工作区组件 — 低优先级文件

#### Task 5.1: 修复 editor-sidebar.tsx 和 editor-layout.tsx 中的 text-slate-*

**文件:** `src/components/ui/editor-sidebar.tsx` 和 `src/components/ui/editor-layout.tsx`

- [ ] **Step 1: 修复 editor-sidebar.tsx**

替换位置：
- 311 行：`text-slate-900` → `text-foreground`
- 350 行：`text-slate-500` → `text-muted-foreground`
- 368 行、370 行：`text-slate-400/500` → `text-muted-foreground`

- [ ] **Step 2: 修复 editor-layout.tsx**

替换位置：
- 102 行：`text-slate-800` → `text-foreground`
- 104 行：`text-slate-500` → `text-muted-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/components/ui/editor-sidebar.tsx src/components/ui/editor-layout.tsx
git commit -m "refactor: standardize text colors in editor components

- Replace text-slate-900/800 with text-foreground (titles)
- Replace text-slate-500/400 with text-muted-foreground (secondary text)"
```

---

#### Task 5.2: 修复 workspace 和其他工作区文件中的 text-slate-*

**文件:** `src/app/org/[orgName]/workspace/page.tsx`

- [ ] **Step 1: 替换 text-slate-900 → text-foreground**

位置：
- 392 行、442 行：标题和提示 `text-foreground`

- [ ] **Step 2: 替换 text-slate-600/500 → text-muted-foreground**

位置：
- 393 行（project slug）、401 行、405 行、443 行 `text-muted-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/app/org/[orgName]/workspace/page.tsx
git commit -m "refactor: standardize text colors in workspace page

- Replace text-slate-900 with text-foreground (headings)
- Replace text-slate-600/500 with text-muted-foreground (metadata, descriptions)"
```

---

#### Task 5.3: 修复其他工作区相关文件

**文件:** 
- `src/components/layout/UserMenu.tsx`
- `src/app/org/[orgName]/team/page.tsx`
- `src/app/org/[orgName]/settings/layout.tsx`

针对每个文件：

- [ ] **Step 1: 替换 text-slate-900 → text-foreground**
- [ ] **Step 2: 替换 text-slate-600/500 → text-muted-foreground**
- [ ] **Step 3: Commit 每个文件**

```bash
git add src/components/layout/UserMenu.tsx
git commit -m "refactor: standardize text colors in UserMenu"

git add src/app/org/[orgName]/team/page.tsx
git commit -m "refactor: standardize text colors in team page"

git add src/app/org/[orgName]/settings/layout.tsx
git commit -m "refactor: standardize text colors in settings layout"
```

---

## Chunk 6: 创建和指南页面 — 余下文件

#### Task 6.1: 修复 org/create/page.tsx

**文件:** `src/app/org/create/page.tsx`

- [ ] **Step 1: 替换 text-slate-* → text-foreground/text-muted-foreground**

位置：
- 189 行：`text-slate-700` → `text-foreground`（代码块文本）
- 242 行：`text-slate-600` → `text-muted-foreground`（链接）

- [ ] **Step 2: Commit**

```bash
git add src/app/org/create/page.tsx
git commit -m "refactor: standardize text colors in org create page"
```

---

#### Task 6.2: 修复 cluster/page.tsx 中的 text-slate-* 和 text-gray-*

**文件:** `src/app/org/[orgName]/projects/[projectSlug]/cluster/page.tsx`

- [ ] **Step 1: 替换 text-gray-900 → text-foreground**

位置：
- 246 行、300 行：标题 `text-foreground`

- [ ] **Step 2: 替换 text-gray-500 → text-muted-foreground**

位置：
- 299 行：Icon `text-muted-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/app/org/[orgName]/projects/[projectSlug]/cluster/page.tsx
git commit -m "refactor: standardize text colors in cluster page"
```

---

#### Task 6.3: 修复 team/page.tsx

**文件:** `src/app/org/[orgName]/team/page.tsx`

- [ ] **Step 1: 替换 text-slate-900 → text-foreground**

位置：
- 68 行、126 行、251 行：标题 `text-foreground`

- [ ] **Step 2: 替换 text-slate-600/500 → text-muted-foreground**

位置：
- 77 行、111 行、129 行、136 行、147 行 `text-muted-foreground`

- [ ] **Step 3: Commit**

```bash
git add src/app/org/[orgName]/team/page.tsx
git commit -m "refactor: standardize text colors in team page"
```

---

#### Task 6.4: 修复 projects/page.tsx

**文件:** `src/app/projects/page.tsx`

- [ ] **Step 1: 替换 text-gray-600 → text-muted-foreground**

位置：
- 31 行、42 行：重定向消息 `text-muted-foreground`

- [ ] **Step 2: Commit**

```bash
git add src/app/projects/page.tsx
git commit -m "refactor: standardize text colors in projects page"
```

---

## Chunk 7: 验证和最终测试

#### Task 7.1: 全局验证 — 确保无遗漏

- [ ] **Step 1: 验证所有 font-bold 已替换**

Run:
```bash
grep -r "font-bold" /root/modelcraft_project/modelcraft-front/src --include="*.tsx" --include="*.ts" | grep -v "node_modules" | grep -v "\.d\.ts"
```

Expected: 仅在 `src/lib/typography.ts` 注释中出现（定义而非使用），无其他实际用法

- [ ] **Step 2: 验证所有 font-extrabold 已替换**

Run:
```bash
grep -r "font-extrabold" /root/modelcraft_project/modelcraft-front/src --include="*.tsx"
```

Expected: 无输出

- [ ] **Step 3: 验证所有 text-gray-* 已替换**

Run:
```bash
grep -r "text-gray-" /root/modelcraft_project/modelcraft-front/src --include="*.tsx" | grep -v "hover:text-gray" | head -5
```

Expected: 最多 5 行（容许部分 hover 状态的 text-gray-*），非关键

- [ ] **Step 4: 验证已使用 text-foreground 和 text-muted-foreground**

Run:
```bash
grep -r "text-foreground\|text-muted-foreground" /root/modelcraft_project/modelcraft-front/src --include="*.tsx" | wc -l
```

Expected: 50+ 行（已有多处替换）

- [ ] **Step 5: 检查项目编译无错**

Run:
```bash
cd /root/modelcraft_project/modelcraft-front && npm run build 2>&1 | tail -20
```

Expected: Build 成功，或仅有非字体相关的警告

- [ ] **Step 6: 最终 commit - 统一提交验证**

```bash
git log --oneline | head -20
```

验证前面的所有 commit 都已成功应用

---

---

## Chunk 8: 增加 ESLint 规则强制语义变量

> 在代码统一替换完成后（Chunk 1-6）再配置此规则，避免尚未迁移的代码在 CI 中大量报错。

**策略**：使用 ESLint 内置的 `no-restricted-syntax` 规则，通过 AST 选择器匹配 `className` 字符串中含有禁止 Tailwind 类的模式，覆盖 JSX 的 `className`、`cn()` 调用、`clsx()` 调用三类场景。

**文件:**
- Modify: `.eslintrc.cjs`

#### Task 8.1: 在 .eslintrc.cjs 中添加 no-restricted-syntax 规则

- [ ] **Step 1: 了解当前规则配置**

Read `.eslintrc.cjs` 确认现有结构（已有 `tailwindcss/` 系列规则）

- [ ] **Step 2: 添加 no-restricted-syntax 规则到 .eslintrc.cjs**

在 `rules` 对象中新增以下规则：

```js
// 禁止使用非语义化字体权重（应使用 font-semibold/medium/normal）
'no-restricted-syntax': [
  'error',
  // 禁止在 className 字符串中使用 font-bold / font-extrabold / font-black
  {
    selector: 'JSXAttribute[name.name="className"] Literal[value=/\\bfont-(bold|extrabold|black)\\b/]',
    message: '禁止使用 font-bold/font-extrabold/font-black。请根据设计规范使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
  },
  {
    selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\bfont-(bold|extrabold|black)\\b/]',
    message: '禁止使用 font-bold/font-extrabold/font-black。请根据设计规范使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
  },
  // 禁止在 className 字符串中使用 text-gray-* 具体数值（应使用 text-foreground 或 text-muted-foreground）
  {
    selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
    message: '禁止使用 text-gray-* 具体值。请使用语义化变量：主文本用 text-foreground，次要/辅助文本用 text-muted-foreground。参见 STYLE.md 第 1.3 节。',
  },
  {
    selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
    message: '禁止使用 text-gray-* 具体值。请使用语义化变量：主文本用 text-foreground，次要/辅助文本用 text-muted-foreground。参见 STYLE.md 第 1.3 节。',
  },
  // 禁止在 className 字符串中使用 text-slate-* 具体数值
  {
    selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
    message: '禁止使用 text-slate-* 具体值。请使用语义化变量：主文本用 text-foreground，次要/辅助文本用 text-muted-foreground。参见 STYLE.md 第 1.3 节。',
  },
  {
    selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
    message: '禁止使用 text-slate-* 具体值。请使用语义化变量：主文本用 text-foreground，次要/辅助文本用 text-muted-foreground。参见 STYLE.md 第 1.3 节。',
  },
],
```

完整修改后的 `.eslintrc.cjs`：

```js
module.exports = {
  root: true,
  extends: ['next/core-web-vitals', 'plugin:tailwindcss/recommended'],
  plugins: ['tailwindcss'],
  rules: {
    'react/no-unescaped-entities': 'off',
    'tailwindcss/classnames-order': 'warn',
    'tailwindcss/enforces-negative-arbitrary-values': 'warn',
    'tailwindcss/enforces-shorthand': 'warn',
    'tailwindcss/migration-from-tailwind-2': 'off',
    'tailwindcss/no-arbitrary-value': 'off',
    'tailwindcss/no-contradicting-classname': 'error',
    'tailwindcss/no-custom-classname': 'off',
    'tailwindcss/no-unnecessary-arbitrary-value': 'warn',

    // --- 字体规范强制 ---
    // 禁止使用超出设计规范的字体权重和非语义化颜色
    // 设计规范参见: src/lib/typography.ts, ai-metadata/style/STYLE.md
    'no-restricted-syntax': [
      'error',
      // 字体权重：禁止 font-bold / font-extrabold / font-black
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\bfont-(bold|extrabold|black)\\b/]',
        message: '禁止使用 font-bold/font-extrabold/font-black。请使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\bfont-(bold|extrabold|black)\\b/]',
        message: '禁止使用 font-bold/font-extrabold/font-black。请使用 font-semibold (600) 或 font-medium (500)。参见 src/lib/typography.ts。',
      },
      // 文字颜色：禁止 text-gray-{400-900}，应改用语义化变量
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-gray-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-gray-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-gray-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      // 文字颜色：禁止 text-slate-{400-900}，应改用语义化变量
      {
        selector: 'JSXAttribute[name.name="className"] Literal[value=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-slate-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
      {
        selector: 'JSXAttribute[name.name="className"] TemplateLiteral > TemplateElement[value.raw=/\\btext-slate-(400|500|600|700|800|900)\\b/]',
        message: '禁止使用 text-slate-* 具体值。主文本用 text-foreground，次要文本用 text-muted-foreground。参见 STYLE.md §1.3。',
      },
    ],
  },
  settings: {
    tailwindcss: {
      config: 'tailwind.config.ts',
      cssFiles: ['**/*.css', '!**/node_modules', '!**/.*', '!**/dist', '!**/build'],
    },
  },
}
```

- [ ] **Step 3: 运行 lint 检查，确保规则生效**

Run:
```bash
cd /root/modelcraft_project/modelcraft-front && npm run lint 2>&1 | grep "no-restricted-syntax" | head -5
```

**注意：此步骤应在 Chunk 1-6 全部完成后执行。** 如果代码已全部迁移，Expected 输出为空（无违规）。若还有遗漏，错误信息会精确指出文件和行号。

- [ ] **Step 4: 若有剩余违规，逐条修复**

根据 lint 报错修复剩余的 `font-bold`、`text-gray-*`、`text-slate-*` 用法，再次运行：

```bash
cd /root/modelcraft_project/modelcraft-front && npm run lint 2>&1 | grep "no-restricted-syntax"
```

Expected: 无输出

- [ ] **Step 5: 确认 typography.ts 本身不受误报影响**

`src/lib/typography.ts` 中的 `font-bold` 是常量定义（非 JSXAttribute），不在规则覆盖范围内，无需特殊处理。

Run:
```bash
cd /root/modelcraft_project/modelcraft-front && npx eslint src/lib/typography.ts 2>&1 | grep "no-restricted-syntax"
```

Expected: 无输出（typography.ts 不包含 JSX，不触发规则）

- [ ] **Step 6: Commit**

```bash
git add .eslintrc.cjs
git commit -m "feat: add ESLint rules to enforce semantic font variables

- Forbid font-bold/font-extrabold/font-black in JSX className (use font-semibold/medium per typography.ts)
- Forbid text-gray-*/text-slate-* in JSX className (use text-foreground/text-muted-foreground per STYLE.md §1.3)
- Rules cover both string literals and template literals in className attributes"
```

---

## 总结

此计划包含 **20+ 个细化任务**，分 **8 个 Chunk** 完成：

1. **Chunk 1** - 字体权重统一（`font-bold` → `font-semibold`）：3 个高优先级文件
2. **Chunk 2** - 主要页面颜色统一（登录、布局）：2 个文件，影响范围广
3. **Chunk 3** - 组件库颜色统一（卡片、选择器）：2 个文件
4. **Chunk 4** - 数据库和错误组件颜色统一：4 个文件
5. **Chunk 5** - 编辑器和工作区颜色统一：5 个文件
6. **Chunk 6** - 创建和指南页面颜色统一：4 个文件
7. **Chunk 7** - 全局验证和最终测试：1 个验证任务
8. **Chunk 8** - ESLint 规则强制语义变量：修改 `.eslintrc.cjs`，防止未来代码引入相同问题

**修改统计**：
- 总文件数：20+ 个
- 总行数修改：100+ 行
- 字体权重替换：14+ 处
- 颜色替换：80+ 处

**验收标准**：
- ✅ 无 `font-bold` 或 `font-extrabold` 在实际代码中使用
- ✅ 所有标题和主文本使用 `text-foreground`
- ✅ 所有次要/辅助文本使用 `text-muted-foreground`
- ✅ 项目编译成功，无样式相关错误
- ✅ 所有 commit 信息清晰记录修改原因
