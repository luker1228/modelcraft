# 公共 Layout 样式使用指南

## 📁 文件结构

```
prototypes/
├── shared/
│   └── layout-styles.css          ← 公共 layout 样式
├── team/
│   └── team-management-prototype.html  ← ✅ 已使用公共样式
└── workspace/
    ├── index.html                      ← ⚠️ 待更新（使用 Tailwind）
    └── index.html.backup               ← 原始备份
```

## ✅ 已完成

1. **创建了 `/prototypes/shared/layout-styles.css`**
   - 包含完整的 Layout 结构（topbar + sidebar + main-content）
   - CSS 变量（STYLE.md 规范）
   - 公共组件样式（按钮、徽章、表格、输入框等）

2. **更新了 `/prototypes/team/team-management-prototype.html`**
   - 移除了所有重复的 layout 样式
   - 引用公共样式：`<link href="../shared/layout-styles.css" rel="stylesheet">`

## 🚀 如何使用公共样式创建新页面

### 模板代码

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>页面标题 - ModelCraft</title>
  
  <!-- Fonts -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
  
  <!-- 🎯 引用公共 Layout 样式 -->
  <link href="../shared/layout-styles.css" rel="stylesheet">
  
  <!-- Lucide Icons -->
  <script src="https://unpkg.com/lucide@latest"></script>
  
  <!-- 页面特定样式（如果需要） -->
  <style>
    /* 在这里添加页面特定的样式 */
  </style>
</head>
<body>
  <div class="layout">
    <!-- ===== Topbar ===== -->
    <header class="topbar">
      <div class="topbar-left">
        <button class="topbar-org-selector">
          <div class="topbar-logo">
            <i data-lucide="sparkles"></i>
          </div>
        </button>
      </div>

      <div class="topbar-center">
        <nav class="topbar-breadcrumb">
          <span style="font-weight: 500; color: var(--text-primary);">组织名</span>
          <span class="topbar-breadcrumb-separator">/</span>
          <span style="color: var(--text-primary);">页面名</span>
        </nav>
      </div>

      <div class="topbar-right">
        <button class="topbar-btn" title="搜索">
          <i data-lucide="search"></i>
        </button>
        <button class="topbar-btn" title="通知">
          <i data-lucide="bell"></i>
          <span class="topbar-btn-badge"></span>
        </button>
        <button class="topbar-btn" title="刷新">
          <i data-lucide="refresh-cw"></i>
        </button>
        <button class="topbar-btn" title="帮助">
          <i data-lucide="help-circle"></i>
        </button>

        <div class="topbar-divider"></div>

        <button class="topbar-user">
          <div class="topbar-user-avatar">张</div>
          <i data-lucide="chevron-down"></i>
        </button>
      </div>
    </header>

    <!-- ===== Main Area ===== -->
    <div class="main-area">
      <!-- Sidebar -->
      <aside class="sidebar">
        <div class="sidebar-content">
          <div class="sidebar-section">
            <div class="sidebar-section-label">WORKSPACE</div>
            <nav>
              <a href="#" class="sidebar-nav-item">
                <i data-lucide="folder-open"></i>
                <span>项目</span>
              </a>
              <a href="#" class="sidebar-nav-item active">
                <i data-lucide="users"></i>
                <span>团队</span>
              </a>
              <a href="#" class="sidebar-nav-item">
                <i data-lucide="settings"></i>
                <span>组织设置</span>
              </a>
            </nav>
          </div>
        </div>

        <div class="sidebar-footer">
          <button class="sidebar-toggle">
            <i data-lucide="panel-left-close"></i>
          </button>
        </div>
      </aside>

      <!-- 🎯 只需要实现这部分内容 -->
      <main class="main-content">
        <div style="max-width: 1280px; margin: 0 auto; padding: 24px;">
          
          <!-- 页面标题 -->
          <div style="margin-bottom: 32px;">
            <h1 style="font-size: 24px; font-weight: 600; color: var(--text-primary);">
              页面标题
            </h1>
          </div>

          <!-- 你的页面内容 -->
          <div style="background: white; border: 1px solid var(--border); border-radius: 8px; padding: 24px;">
            <!-- 在这里添加页面内容 -->
          </div>

        </div>
      </main>
    </div>
  </div>

  <script>
    lucide.createIcons();
  </script>
</body>
</html>
```

## 📦 公共样式包含的组件

### 1. Layout 结构
- `.layout` - 主容器
- `.topbar` - 顶部栏
- `.sidebar` - 侧边栏
- `.main-content` - 主内容区

### 2. 按钮
- `.btn-primary` - 主要按钮
- `.btn-secondary` - 次要按钮
- `.btn-ghost` - 幽灵按钮

### 3. 徽章
- `.badge-success` - 成功状态
- `.badge-warning` - 警告状态
- `.badge-primary` - 主要状态
- `.badge-neutral` - 中性状态

### 4. 表格
- `table`, `thead`, `tbody`, `th`, `td` - 统一的表格样式

### 5. 输入框
- `input[type="text"]`, `textarea` - 统一的输入框样式

### 6. CSS 变量
```css
--primary: #2563eb
--primary-hover: #1d4ed8
--success: #059669
--warning: #d97706
--text-primary: #111827
--text-secondary: #6b7280
--border: #e5e7eb
--bg-primary: #fafafa
--bg-secondary: #ffffff
```

## ⚠️ Workspace 页面说明

workspace 页面目前使用 Tailwind CSS utility classes，内容比较复杂。

**两种选择：**

1. **保持现状**：workspace 继续使用 Tailwind + 内联样式
2. **迁移到公共样式**：需要手动替换 Tailwind classes 为公共样式类

建议：**保持现状**，因为 workspace 页面功能比较完整，迁移成本较高。

新页面直接使用公共样式即可。

## 📝 最佳实践

1. **新页面**：直接使用公共样式模板
2. **只实现内容区**：`<main class="main-content">` 内部
3. **复用 topbar 和 sidebar**：从模板复制即可
4. **页面特定样式**：放在 `<style>` 标签中，不要修改公共样式

## 示例

参考 `/prototypes/team/team-management-prototype.html`，这是一个使用公共样式的完整示例。
