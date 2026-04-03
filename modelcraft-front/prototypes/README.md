# ModelCraft 原型开发指南

本目录存放 UI 原型 HTML 文件，作为设计的"唯一真相源"。

## 目录结构

```
prototypes/
├── README.md                    # 本文件
├── shared/                      # 共享资源
│   ├── tailwind.config.js      # Tailwind 配置（与项目同步）
│   ├── tailwind-base.css       # CSS 变量（与项目同步）
│   └── components/             # 可复用组件片段
│       ├── sidebar-workspace.html  # Sidebar - Workspace 模式
│       ├── sidebar-project.html    # Sidebar - Project 模式
│       └── topbar.html             # Topbar 顶部栏
├── layout/                      # 布局原型（Sidebar + Topbar + Content）
│   └── index.html
├── login/                       # 登录页原型
│   └── index.html
├── workspace/                   # 工作空间页原型
│   └── index.html
└── ...                          # 其他页面原型
```

## 核心原则

**永远先修改原型，再写 React 代码。**

原型文件是设计的"唯一真相源"。任何 UI 变更都必须：
1. 先在 `prototypes/` 目录下创建或修改 HTML 原型
2. 确认原型视觉效果符合预期
3. 再基于原型实现 React 组件

## 使用方法

### 1. 创建新原型

```bash
# 创建新页面目录
mkdir -p prototypes/new-page

# 创建 HTML 文件
touch prototypes/new-page/index.html
```

### 2. HTML 模板

每个原型 HTML 文件使用以下模板：

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

### 3. 预览原型

直接在浏览器中打开 HTML 文件：

```bash
# macOS
open prototypes/layout/index.html

# Linux
xdg-open prototypes/layout/index.html

# 或使用 VS Code Live Server 插件
```

### 4. 暗色模式预览

在 `<html>` 标签添加 `class="dark"`：

```html
<html lang="zh-CN" class="dark">
```

## 组件复用

### 使用共享组件片段

`shared/components/` 目录下的组件片段可用于：

1. **直接复制** - 复制 HTML 结构到新原型
2. **iframe 嵌入** - 在其他 HTML 中通过 iframe 引入
3. **参考实现** - 作为 React 组件的参考

### 示例：引用 Sidebar 组件

```html
<!-- 方式1：直接复制 sidebar-workspace.html 中的内容 -->

<!-- 方式2：iframe 嵌入 -->
<iframe src="../shared/components/sidebar-workspace.html" frameborder="0"></iframe>
```

## 设计系统

### 颜色规范

| 用途 | Hex | Tailwind | 使用场景 |
|------|-----|----------|---------|
| Primary | `#2563eb` | `blue-600` | 主按钮、链接、激活状态 |
| Primary Hover | `#1d4ed8` | `blue-700` | 悬停/聚焦状态 |
| Primary Light | `#dbeafe` | `blue-100` | 浅色背景 |
| Success | `#059669` | `emerald-600` | 活跃状态、成功消息 |
| Warning | `#d97706` | `amber-600` | 草稿状态、警告 |
| Error | `#ef4444` | `red-500` | 删除、错误 |
| Selected | `#dadee5` | - | 选中行/项目背景 |
| Border | `#e5e7eb` | `gray-200` | 边框 |
| Background | `#fafafa` | `gray-50` | 页面背景 |

### 可用颜色类名

使用项目定义的颜色变量，与正式代码完全一致：

| 类名 | 说明 |
|------|------|
| `bg-background` / `text-foreground` | 背景色 / 前景色 |
| `bg-primary` / `text-primary` | 主色调 |
| `bg-secondary` / `text-secondary` | 次要色 |
| `bg-muted` / `text-muted-foreground` | 静默色 |
| `bg-card` / `text-card-foreground` | 卡片色 |
| `bg-destructive` / `text-destructive` | 危险色 |
| `bg-selected` / `text-selected-foreground` | 选中色 |
| `bg-sidebar` / `text-sidebar-foreground` | 侧边栏色 |

### 可用圆角类名

| 类名 | 说明 |
|------|------|
| `rounded-sm` | 小圆角 |
| `rounded-md` | 中圆角 |
| `rounded-lg` | 大圆角 |
| `rounded-xl` | 超大圆角 |
| `rounded-2xl` | 特大圆角 |

## 工作流程

### 设计师

1. 在 `prototypes/<page>/` 目录下编写或修改 HTML
2. 使用项目统一的 Tailwind 类名
3. 在浏览器中预览效果
4. 提交代码到 Git

### 开发者

1. 查看原型 HTML 确认设计
2. 将 Tailwind 类名复制到 React 组件
3. 实现交互逻辑
4. 如有疑问，以原型为准

## 页面与原型对应关系

| 页面路由 | 原型文件 |
|----------|----------|
| `/login` | `prototypes/login/index.html` |
| `/org/[orgName]/workspace` | `prototypes/workspace/index.html` |
| `/org/[orgName]/projects/[slug]/*` | 参考 `prototypes/layout/index.html` |

## 注意事项

1. **保持同步** - `shared/` 目录下的配置需要与项目正式配置保持同步
2. **相对路径** - 引用共享资源时使用相对路径 `../shared/`
3. **Git 提交** - 原型文件纳入版本控制，方便追踪设计变更
4. **命名规范** - 目录名与 `src/app/` 下的路由对应，便于查找

## 常见问题

**Q: 原型不加载样式？**
A: 检查相对路径是否正确，确保 `../shared/` 能正确引用共享资源。

**Q: 图标不显示？**
A: 确保在页面底部调用了 `lucide.createIcons()`。

**Q: 暗色模式不生效？**
A: 检查 `<html>` 标签是否有 `class="dark"`。

**Q: 选中状态用什么背景色？**
A: 使用 `#dadee5`（`bg-[#dadee5]`），不要使用 `bg-blue-50` 或透明度效果。
