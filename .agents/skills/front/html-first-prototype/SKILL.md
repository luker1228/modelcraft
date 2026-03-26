---
name: html-first-prototype
description: |
  HTML-First Prototype 工作流。用于所有 UI 相关任务。

  必须使用此 skill 的情况：
  - 创建新页面或新组件
  - 修改现有 UI 的布局、样式或结构
  - 调整颜色、间距、字体等视觉元素
  - 重构 UI 组件的 HTML 结构

  核心原则：**永远先修改 prototypes/ 目录下的 HTML 原型，确认设计后再实现 React 代码**。
  原型是设计的"唯一真相源"，React 实现必须与原型保持一致。
---

# HTML-First Prototype 工作流

## 核心原则

**永远先修改原型，再写 React 代码。**

原型文件是设计的"唯一真相源"。任何 UI 变更都必须：
1. 先在 `prototypes/` 目录下创建或修改 HTML 原型
2. 确认原型视觉效果符合预期
3. 再基于原型实现 React 组件

## 工作流程

### Step 1: 检查或创建原型目录

```
prototypes/
├── shared/                      # 共享资源（设计系统源）
│   ├── tailwind.config.js      # Tailwind 配置（可修改）
│   └── tailwind-base.css       # CSS 变量（可修改）
├── <page-name>/                 # 页面原型目录
│   ├── index.html              # 主 HTML 文件
│   └── assets/                 # 页面专属资源
└── README.md                    # 使用指南
```

**重要说明：**
- `shared/` 目录中的文件是设计系统的源文件
- 可以直接修改 `tailwind.config.js` 和 `tailwind-base.css`
- 修改后需要同步到 React 项目（见 Step 3.5）

### Step 2: 创建或修改 HTML 原型

**新建原型时，使用以下模板：**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Page Name - ModelCraft Prototype</title>

  <!-- Fonts -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Space+Grotesk:wght@500;600;700&family=Fira+Code:wght@400;500&display=swap" rel="stylesheet">

  <!-- Tailwind CSS -->
  <script src="https://cdn.tailwindcss.com"></script>
  <script src="../shared/tailwind.config.js"></script>
  <link href="../shared/tailwind-base.css" rel="stylesheet">

  <!-- Lucide Icons -->
  <script src="https://unpkg.com/lucide@latest"></script>
</head>
<body>
  <!-- 在此编写 UI -->

  <!-- Initialize Lucide Icons -->
  <script>
    lucide.createIcons();
  </script>
</body>
</html>
```

### Step 3: 在浏览器中预览原型

直接在浏览器打开 HTML 文件确认效果。

**暗色模式预览：** 在 `<html>` 标签添加 `class="dark"`

### Step 3.5: 同步设计系统到 React（如有修改）

如果你在原型中修改了设计系统配置（颜色、间距、字体等），需要同步到 React 项目：

```bash
npm run sync-tailwind
```

这个命令会：
- 将 `prototypes/shared/tailwind.config.js` → `tailwind.config.ts`
- 将 `prototypes/shared/tailwind-base.css` → `src/app/globals.css`
- 自动创建备份文件到 `.tailwind-backups/`

**何时需要运行：**
- ✅ 修改了原型的 `tailwind.config.js`
- ✅ 修改了原型的 `tailwind-base.css`
- ✅ 添加了新的颜色、字体、间距定义
- ❌ 只修改了 HTML 结构（不需要同步）

### Step 4: 基于原型实现 React 组件

将原型中的 Tailwind 类名直接复制到 React 组件中。

**对应关系：**
- `prototypes/login/index.html` → `src/app/login/page.tsx`
- `prototypes/workspace/index.html` → `src/app/org/[orgName]/workspace/page.tsx`
- `prototypes/dashboard/index.html` → `src/app/org/[orgName]/project/[projectSlug]/dashboard/page.tsx`

## 设计规范

> 完整规范见 `ai-metadata/front/style/STYLE.md`

### 核心原则

- **实色为主** - 禁止渐变，禁止透明度效果，禁止装饰性光晕
- **边框优先** - 使用边框定义元素边界，而非阴影
- **克制简洁** - B2B 风格，专业清晰

### 颜色系统

| 用途 | Hex | Tailwind | 使用场景 |
|------|-----|----------|---------|
| Primary | `#2563eb` | `blue-600` | 主按钮、链接、激活状态 |
| Primary Hover | `#1d4ed8` | `blue-700` | 悬停/聚焦状态 |
| Primary Light | `#dbeafe` | `blue-100` | 浅色背景 |
| Success | `#059669` | `emerald-600` | 活跃状态、成功消息 |
| Success Light | `#ecfdf5` | `emerald-50` | 成功背景 |
| Warning | `#d97706` | `amber-600` | 草稿状态、警告 |
| Warning Light | `#fef3c7` | `amber-50` | 警告背景 |
| Error | `#ef4444` | `red-500` | 删除、错误 |
| Error Light | `#fee2e2` | `red-50` | 错误背景 |
| Selected | `#dadee5` | - | 选中行/项目背景 |
| Border | `#e5e7eb` | `gray-200` | 边框 |
| Background | `#fafafa` | `gray-50` | 页面背景 |
| Card | `#ffffff` | `white` | 卡片背景 |

**常见错误：**
- ❌ 使用渐变 `bg-gradient-to-r from-blue-600 to-indigo-600`
- ❌ 使用毛玻璃 `bg-white/70 backdrop-blur-sm`
- ❌ 使用 `bg-blue-50` 作为选中背景（应使用 `#dadee5`）

### Styled System（样式系统）

#### 1. 布局系统

**Sidebar 宽度：**
- 展开状态：`200px` (`12.5rem` / `w-[200px]`)
- 折叠状态：`64px` (`4rem` / `w-16`)

**Topbar 高度：**
- 固定高度：`56px` (`3.5rem` / `h-14`)

**容器宽度：**
```html
<!-- 标准内容容器 -->
<div class="max-w-7xl mx-auto">
  <!-- 1280px 最大宽度，居中对齐 -->
</div>

<!-- 搜索框容器 -->
<div class="max-w-md">
  <!-- 448px 最大宽度 -->
</div>
```

#### 2. 间距系统

**统一间距规范：**

| 间距值 | Tailwind | 用途 |
|--------|----------|------|
| `4px` | `p-1`, `gap-1` | 最小间距 |
| `8px` | `p-2`, `gap-2` | 紧凑元素（图标+文字） |
| `12px` | `p-3`, `gap-3` | 中等间距（按钮内边距） |
| `16px` | `p-4`, `gap-4` | 标准间距（卡片内边距） |
| `24px` | `p-6`, `gap-6` | 大间距（页面内边距） |
| `32px` | `p-8`, `gap-8` | 特大间距（区块间距） |

**常用间距组合：**
```html
<!-- 页面容器 -->
<div class="p-6">  <!-- 24px 内边距 -->

<!-- 卡片 -->
<div class="p-4 gap-3">  <!-- 16px 内边距，12px 间距 -->

<!-- 工具栏 -->
<div class="flex items-center gap-3">  <!-- 12px 间距 -->

<!-- 按钮 -->
<button class="h-9 px-4 gap-2">  <!-- 36px 高度，16px 水平内边距，8px 图标间距 -->
```

#### 3. 字体系统

**字体家族：**
- 正文：`Inter` (font-family: Inter, system-ui, sans-serif)
- 标题：`Space Grotesk` (font-family: Space Grotesk, system-ui, sans-serif)
- 代码：`Fira Code` (font-family: Fira Code, monospace)

**字体大小规范：**

| 用途 | 大小 | Tailwind | Font Weight |
|------|------|----------|-------------|
| 大标题 | `32px` | `text-[32px]` | `font-semibold` (600) |
| 页面标题 | `24px` | `text-2xl` | `font-bold` (700) |
| 卡片标题 | `16px` | `text-base` | `font-semibold` (600) |
| 正文 | `14px` | `text-sm` | `font-normal` (400) |
| 小字 | `12px` | `text-xs` | `font-normal` (400) |
| Section Label | `10px` | `text-[10px]` | `font-mono uppercase` |

**示例：**
```html
<!-- 页面大标题 -->
<h1 class="text-[32px] font-semibold tracking-tight text-slate-900">
  所有项目
</h1>

<!-- 卡片标题 -->
<h3 class="text-base font-semibold text-gray-900">
  电商平台项目
</h3>

<!-- Section Label -->
<div class="text-[10px] font-mono uppercase tracking-wider text-muted-foreground">
  WORKSPACE
</div>
```

#### 4. 边框圆角系统

| 用途 | 圆角值 | Tailwind |
|------|--------|----------|
| 按钮 | `6px` | `rounded-md` |
| 卡片 | `8px` | `rounded-lg` |
| 徽章/头像 | `9999px` | `rounded-full` |
| Logo | `8px` | `rounded-lg` |
| 输入框 | `6px` | `rounded-md` |

#### 5. 阴影系统

**原则：边框优先，谨慎使用阴影**

仅在悬停状态使用轻微阴影：
```html
<!-- 卡片悬停 -->
<div class="hover:shadow-sm">
  <!-- shadow-sm = 0 1px 2px 0 rgba(0, 0, 0, 0.05) -->
</div>
```

**禁止使用的阴影：**
- ❌ `shadow-lg` (过重)
- ❌ `shadow-xl` (过重)
- ❌ `shadow-2xl` (过重)
- ❌ 带颜色的阴影 `shadow-blue-500/30` (装饰性)

#### 6. 图标系统

**使用 Lucide Icons，统一规范：**

| 用途 | 尺寸 | Stroke Width | Tailwind |
|------|------|--------------|----------|
| 主导航 | `16px` | `1.5` | `w-4 h-4` |
| 按钮图标 | `14px` | `1.5` | `w-3.5 h-3.5` |
| Topbar 图标 | `16px` | `1.5` | `w-4 h-4` |
| 大图标 | `32px` | `1.5` | `w-8 h-8` |

**示例：**
```html
<!-- 按钮中的图标 -->
<button class="flex items-center gap-2">
  <i data-lucide="plus" class="w-3.5 h-3.5"></i>
  新建项目
</button>

<!-- 导航图标 -->
<a href="#" class="flex items-center gap-3">
  <i data-lucide="folder-open" class="w-4 h-4"></i>
  <span>项目</span>
</a>
```

#### 7. 动画与过渡

**统一过渡时长：`200ms`**

```html
<!-- 标准过渡 -->
<button class="transition-all duration-200 hover:bg-[#1d4ed8]">

<!-- 颜色过渡 -->
<div class="transition-colors duration-200">

<!-- 阴影过渡 -->
<div class="transition-shadow duration-200 hover:shadow-sm">
```

**禁止过长的动画：**
- ❌ `duration-300` (太慢)
- ❌ `duration-500` (太慢)

### 组件规范

**主按钮：**
```html
<button class="h-9 px-4 bg-[#2563eb] hover:bg-[#1d4ed8] text-white border-0 rounded-md font-medium text-sm gap-2 transition-all duration-200 inline-flex items-center">
  <i data-lucide="plus" class="w-4 h-4"></i>
  创建项目
</button>
```

**内容卡片：**
```html
<div class="bg-white border border-gray-200 rounded-lg p-4 transition-all duration-200 hover:border-blue-100 hover:shadow-sm">
  <h3 class="text-base font-semibold text-gray-900 mb-2">Card Title</h3>
  <p class="text-sm text-gray-500 mb-3">Description text.</p>
  <span class="inline-flex items-center px-3 py-1 rounded bg-[#ecfdf5] text-[#059669] text-xs font-semibold">Active</span>
</div>
```

**状态徽章：**
```html
<!-- Active/Success -->
<span class="inline-flex items-center px-3 py-1 rounded bg-[#ecfdf5] text-[#059669] text-xs font-semibold">活跃</span>

<!-- Draft/Warning -->
<span class="inline-flex items-center px-3 py-1 rounded bg-[#fef3c7] text-[#d97706] text-xs font-semibold">草稿</span>

<!-- Archived -->
<span class="inline-flex items-center px-3 py-1 rounded bg-gray-100 text-gray-600 text-xs font-semibold">已归档</span>

<!-- Error -->
<span class="inline-flex items-center px-3 py-1 rounded bg-[#fee2e2] text-[#ef4444] text-xs font-semibold">错误</span>
```

**表单输入：**
```html
<input
  type="text"
  placeholder="搜索项目..."
  class="w-full h-9 px-4 border border-gray-200 rounded-md text-sm text-gray-900 bg-white
         placeholder:text-gray-400 focus:outline-none focus:border-[#2563eb]
         focus:ring-2 focus:ring-[rgba(37,99,235,0.1)] transition-all duration-200"
/>
```

## 检查清单

在实现 React 代码之前，确认：

- [ ] 原型 HTML 文件已创建/修改
- [ ] 原型在浏览器中预览效果符合预期
- [ ] 暗色模式也已预览（如适用）
- [ ] 遵循 ModelCraft 设计系统（无渐变、无毛玻璃）
- [ ] Tailwind 类名已准备好复制到 React 组件

## 常见问题

**Q: 只是小改动（如改个颜色）也要先改原型吗？**

是的。即使是小改动，也建议先改原型确认效果，避免反复调试。

**Q: 原型和 React 代码不一致怎么办？**

以原型为准。如果原型已确认的设计，React 实现必须匹配。

**Q: 如何同步项目的 Tailwind 配置更新？**

如果项目的 `tailwind.config.ts` 或 `globals.css` 有更新，需要同步到：
- `prototypes/shared/tailwind.config.js`
- `prototypes/shared/tailwind-base.css`

**Q: 选中状态应该用什么背景色？**

使用 `#dadee5`（`bg-[#dadee5]`），不要使用 `bg-blue-50` 或 `rgba(37,99,235,0.05)`，后者太浅不明显。

## 相关文档

| 文档 | 内容 |
|------|------|
| `ai-metadata/front/style/STYLE.md` | 完整设计系统规范 |
| `ai-metadata/front/style/quick-start.md` | 组件快速参考 |
| `prototypes/README.md` | 原型开发指南 |
| `scripts/README.md` | 自动化脚本说明 |

## 同步 Tailwind 配置

### 何时需要同步

当您修改设计系统时，运行同步脚本：

```bash
npm run sync-tailwind
```

**需要同步的情况：**
- ✅ 修改 `tailwind.config.ts`（颜色、字体、间距等）
- ✅ 修改 `src/app/globals.css`（CSS 变量、自定义类）
- ✅ 添加新的设计 token

**脚本作用：**
- 自动将 React 项目配置同步到原型文件
- `tailwind.config.ts` → `prototypes/shared/tailwind.config.js`
- `globals.css` → `prototypes/shared/tailwind-base.css`

### 配置同步检查清单

修改设计系统后：

- [ ] 运行 `npm run sync-tailwind`
- [ ] 在浏览器中打开原型文件验证效果
- [ ] 确认新的 Tailwind 类在原型中正常工作
- [ ] 将验证通过的类名复制到 React 组件

**注意：** 不要手动修改 `prototypes/shared/` 下的文件，它们由脚本自动生成。
