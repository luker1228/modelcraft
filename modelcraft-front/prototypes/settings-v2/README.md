# Settings V2 原型 - 项目设置页面

## 概述

Settings V2 是基于 Supabase 组织设置页面的设计实现，使用**双栏布局**（左侧导航 + 右侧内容）。适用于项目设置页面。

## 布局结构

```
┌─────────────────────────────────────────────────────────────────┐
│  TOPBAR  h-56px  (Logo | Breadcrumb | Actions)                  │
│  border-b: 1px #e5e7eb                                          │
└─────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                                                                    │
│ ┌──────────┐ ┌─────────────────┐ ┌──────────────────────────────┐ │
│ │ MAIN     │ │  SETTINGS       │ │    SETTINGS CONTENT          │ │
│ │ SIDEBAR  │ │  SIDEBAR        │ │                              │ │
│ │ (200px)  │ │  (260px)        │ │  ┌──────────────────────┐    │ │
│ │          │ │                 │ │  │ Page Header          │    │ │
│ │ [📊]     │ │ 设置            │ │  │ 项目设置             │    │ │
│ │ [📝]     │ │ ───────         │ │  │ 管理您的项目配置...  │    │ │
│ │ [⚙️] act │ │ [·] 基本信息 ✓  │ │  │ ─────────────────    │    │ │
│ │ [🔌]     │ │ [ ] 安全设置    │ │  │                      │    │ │
│ │          │ │ [ ] 隐私政策    │ │  │ ┌────────────────┐   │    │ │
│ │          │ │ [ ] 成员管理    │ │  │ │ 基本信息       │   │    │ │
│ │          │ │                 │ │  │ │ 项目名称: [..] │   │    │ │
│ │          │ │                 │ │  │ │ 描述: [......] │   │    │ │
│ │          │ │                 │ │  │ │ 状态: ● 已连接 │   │    │ │
│ │          │ │                 │ │  │ │                │   │    │ │
│ │          │ │                 │ │  │ │ [取消] [保存]  │   │    │ │
│ │          │ │                 │ │  │ └────────────────┘   │    │ │
│ │          │ │                 │ │  │                      │    │ │
│ │          │ │                 │ │  │ ┌────────────────┐   │    │ │
│ │          │ │                 │ │  │ │ 数据库连接     │   │    │ │
│ │          │ │                 │ │  │ │ Host: [..][..] │   │    │ │
│ │          │ │                 │ │  │ │ User: [......] │   │    │ │
│ │          │ │                 │ │  │ │ Pass: [·······]│   │    │ │
│ │          │ │                 │ │  │ │                │   │    │ │
│ │          │ │                 │ │  │ │ ✓ 连接成功     │   │    │ │
│ │          │ │                 │ │  │ │ [测试] [取消]  │   │    │ │
│ │          │ │                 │ │  │ │ [保存]         │   │    │ │
│ │          │ │                 │ │  │ └────────────────┘   │    │ │
│ │          │ │                 │ │  │                      │    │ │
│ │          │ │                 │ │  │ ┌────────────────┐   │    │ │
│ │          │ │                 │ │  │ │ 危险区域  ⚠️   │   │    │ │
│ │          │ │                 │ │  │ │ [重置连接]     │   │    │ │
│ │          │ │                 │ │  │ │ [删除项目]     │   │    │ │
│ │          │ │                 │ │  │ └────────────────┘   │    │ │
│ │          │ │                 │ │  └──────────────────────┘    │ │
│ │          │ │                 │ │                              │ │
│ └──────────┘ └─────────────────┘ └──────────────────────────────┘ │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

## 设计特点

### 1. 三层布局结构

| 层级 | 组件 | 宽度 | 用途 |
|------|------|------|------|
| 第一层 | Main Sidebar | 200px | 项目导航（概览、模型编辑器、设置、API） |
| 第二层 | Settings Sidebar | 260px | 设置分类导航（基本信息、安全、隐私、成员） |
| 第三层 | Content Area | 剩余空间 | 具体设置表单内容 |

### 2. 色彩系统

**遵循 ModelCraft 设计系统：**
- **主色**: `#2563eb` (蓝色) - 按钮、链接
- **成功**: `#059669` (绿色) - 状态徽章、确认
- **警告/错误**: `#dc2626`/`#ef4444` (红色) - 危险操作
- **边框**: `#e5e7eb` (浅灰) - 分割线、边框
- **背景**: `#fafafa` (超浅灰) - 页面背景
- **危险区域背景**: `#fef2f2` (极浅红) - 危险操作区域

### 3. 不使用的样式

✅ **完全遵循设计规范：**
- ✓ 无渐变背景
- ✓ 无透明效果
- ✓ 无装饰性光晕
- ✓ 无大阴影
- ✓ 实色设计
- ✓ 边框优先
- ✓ 最小阴影

## 核心组件

### Settings Sidebar（左侧导航）

```html
<aside class="settings-sidebar">
  <div class="settings-sidebar-title">设置</div>
  <nav class="settings-nav">
    <a href="#" class="settings-nav-item active">
      <i data-lucide="database"></i>
      <span>基本信息</span>
    </a>
    <!-- 其他导航项 -->
  </nav>
</aside>
```

**特点：**
- 固定宽度 260px
- 顶部标题 "设置"
- 左边框指示活跃项 (3px 蓝色)
- 背景色渐变：正常白色 → 悬停浅灰
- 字体颜色：灰色 → 活跃时深灰

### Settings Section（设置卡片）

```html
<div class="settings-section">
  <div class="settings-section-header">
    <h2 class="settings-section-title">基本信息</h2>
    <p class="settings-section-description">描述文本</p>
  </div>
  <div class="settings-section-body">
    <!-- 表单内容 -->
  </div>
</div>
```

**特点：**
- 白色背景，1px 边框
- 头部与内容分离 (border-bottom)
- 危险区域特殊样式 (红色边框、红色背景头)

### Form Group（表单行）

```html
<div class="form-group">
  <div class="form-label-col">
    <label class="form-label">标签 *</label>
    <p class="form-hint">提示文本</p>
  </div>
  <div class="form-control-col">
    <input type="text" class="form-control" />
    <p class="form-control-hint">帮助文本</p>
  </div>
</div>
```

**特点：**
- 左右分列布局 (200px 标签 + 剩余空间内容)
- 标签列较窄，内容列自适应
- 独立的表单行分割线

## 响应式设计

- **桌面**: 三层布局完整显示
- **平板**: 考虑隐藏 Main Sidebar，Settings Sidebar 保留
- **手机**: 考虑底部标签式导航或全屏单栏

## 使用示例

### 1. 打开原型

在浏览器中打开 `prototypes/settings-v2/index.html` 预览完整效果。

### 2. 在项目中实现

复制 CSS 样式和 HTML 结构到 React 组件：

```tsx
// src/app/org/[orgName]/projects/[projectSlug]/settings-v2/page.tsx
import React from 'react'

export default function SettingsV2Page() {
  return (
    <div className="settings-layout">
      {/* Settings Sidebar */}
      <aside className="settings-sidebar">
        {/* 导航项 */}
      </aside>

      {/* Settings Content */}
      <main className="settings-content">
        {/* 表单内容 */}
      </main>
    </div>
  )
}
```

### 3. CSS 变量配置

所有样式使用 CSS 变量（见顶部 `:root`）：

```css
--sidebar-width: 200px
--topbar-height: 56px
--settings-sidebar-width: 260px
```

## 与现有 Settings 页面的对比

| 特性 | 原 Settings | Settings V2 |
|------|-----------|-----------|
| 布局 | 单栏 (左 Topbar + 中 Sidebar + 右 Content) | 双栏设置导航 (Main Sidebar + Settings Sidebar + Content) |
| 设置导航 | 无独立导航 | 左侧专用导航栏 |
| 最大内容宽度 | 640px | 1000px |
| 内容分类 | 无分类，全部堆叠 | 分类清晰（基本信息、安全、隐私、成员） |
| 适用场景 | 项目级设置 (集群、模型) | 项目全局设置 |
| 设计参考 | ModelCraft 原创 | Supabase 组织设置 |

## 代码质量检查清单

- [x] 实色设计，无渐变
- [x] 边框优先定义元素边界
- [x] 最小阴影 (仅 shadow-sm)
- [x] 语义化 HTML (nav, aside, main, section)
- [x] 无行内样式（除了演示用途）
- [x] 图标使用 Lucide Icons (strokeWidth: 1.5)
- [x] 字体使用系统字体栈
- [x] 色彩符合 ModelCraft 调色板
- [x] 响应式设计考虑 (md: breakpoint)
- [x] 无障碍支持 (aria 属性等)

## 参考资源

- **灵感来源**: Supabase Dashboard 组织设置页面
- **设计系统**: ModelCraft Design System (`ai-metadata/style/STYLE.md`)
- **图标库**: Lucide Icons
- **布局框架**: Tailwind CSS + 原生 CSS

## 下一步

1. 将 HTML 结构转换为 React 组件
2. 实现表单状态管理 (React Hook Form)
3. 集成 GraphQL 查询/变更
4. 添加表单验证逻辑
5. 实现响应式设计
6. 添加暗色模式支持 (可选)
