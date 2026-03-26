# ESLint 规则文档

本文档说明项目的 ESLint 配置和规范。

## 配置文件

- **位置**: `.eslintrc.cjs`
- **运行命令**: `npm run lint`

## 基础配置

### 继承的配置

```javascript
extends: [
  'next/core-web-vitals',          // Next.js 官方推荐规则
  'plugin:tailwindcss/recommended' // Tailwind CSS 最佳实践
]
```

### 插件

- `tailwindcss`: Tailwind CSS 类名检查和优化

## 核心规则

### 1. React 规则

#### `react/no-unescaped-entities`: OFF

```tsx
// ✅ 允许:无需转义单引号、双引号等字符
<div>Don't worry, it's fine</div>

// 这条规则已关闭,因为在现代 React 中通常不会引起问题
```

### 2. Tailwind CSS 规则

#### `tailwindcss/classnames-order`: WARN

类名应按 Tailwind 推荐顺序排列(通过 Prettier 插件自动修复)。

```tsx
// ✅ 推荐:按逻辑顺序排列
<div className="flex items-center justify-between p-4 bg-white rounded-lg shadow-md">

// ⚠️ 警告:顺序混乱
<div className="shadow-md bg-white flex rounded-lg p-4 items-center justify-between">
```

#### `tailwindcss/enforces-negative-arbitrary-values`: WARN

负值应使用 `-` 前缀而非任意值。

```tsx
// ✅ 推荐
<div className="-mt-4">

// ⚠️ 警告
<div className="mt-[-16px]">
```

#### `tailwindcss/enforces-shorthand`: WARN

应使用简写类名。

```tsx
// ✅ 推荐
<div className="m-4">

// ⚠️ 警告
<div className="mx-4 my-4">
```

#### `tailwindcss/no-contradicting-classname`: ERROR

禁止冲突的类名。

```tsx
// ❌ 错误:同时设置了 flex 和 block
<div className="flex block">

// ✅ 正确:只使用一个
<div className="flex">
```

#### `tailwindcss/no-unnecessary-arbitrary-value`: WARN

有标准类名时禁止使用任意值。

```tsx
// ✅ 推荐
<div className="p-4">

// ⚠️ 警告
<div className="p-[16px]">
```

#### `tailwindcss/no-arbitrary-value`: OFF

允许使用任意值(适用于设计系统外的特殊需求)。

```tsx
// ✅ 允许:特殊需求可以使用任意值
<div className="w-[347px]">
```

#### `tailwindcss/no-custom-classname`: OFF

允许使用自定义类名(如 CSS Modules 或全局样式)。

### 3. 设计系统强制规则

这些规则通过 `no-restricted-syntax` 实现,确保符合项目设计规范。

#### 字体权重限制 (ERROR)

**禁止**: `font-bold`, `font-extrabold`, `font-black`

**允许**: `font-medium` (500), `font-semibold` (600)

```tsx
// ❌ 错误
<h1 className="font-bold">标题</h1>

// ✅ 正确
<h1 className="font-semibold">标题</h1>
<p className="font-medium">正文</p>
```

**理由**: 设计系统中只使用 Medium (500) 和 Semibold (600) 两种字重,保持视觉一致性。

**参考**: `src/lib/typography.ts`, `ai-metadata/style/STYLE.md`

#### 文字颜色限制 (ERROR)

**禁止**: `text-gray-{400-900}`, `text-slate-{400-900}`

**允许**: 语义化颜色变量

```tsx
// ❌ 错误:直接使用灰度值
<p className="text-gray-600">文字</p>
<p className="text-slate-700">文字</p>

// ✅ 正确:使用语义化变量
<p className="text-foreground">主要文字</p>
<p className="text-muted-foreground">次要文字</p>
```

**理由**: 语义化变量支持主题切换,避免硬编码颜色值。

**参考**: `ai-metadata/style/STYLE.md` §1.3

## 常见错误和修复

### 错误 1: 使用了禁止的字体权重

```tsx
// ❌ ESLint Error
<div className="font-bold text-lg">标题</div>

// ✅ 修复
<div className="font-semibold text-lg">标题</div>
```

### 错误 2: 使用了非语义化颜色

```tsx
// ❌ ESLint Error
<p className="text-gray-500">次要信息</p>

// ✅ 修复
<p className="text-muted-foreground">次要信息</p>
```

### 错误 3: Tailwind 类名冲突

```tsx
// ❌ ESLint Error
<div className="flex block p-4">

// ✅ 修复
<div className="flex p-4">
```

## 自动修复

### 1. 自动修复 Tailwind 类名顺序

```bash
# 使用 Prettier + Tailwind 插件
npm run format
```

### 2. ESLint 自动修复

```bash
# 自动修复可修复的问题
npm run lint -- --fix
```

**注意**: 设计系统规则(字体权重、颜色)无法自动修复,需要手动调整。

## 编辑器集成

### VS Code

安装推荐扩展:

1. **ESLint** (`dbaeumer.vscode-eslint`)
2. **Tailwind CSS IntelliSense** (`bradlc.vscode-tailwindcss`)

配置 `.vscode/settings.json`:

```json
{
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ],
  "tailwindCSS.experimental.classRegex": [
    ["cn\\(([^)]*)\\)", "[\"'`]([^\"'`]*).*?[\"'`]"]
  ]
}
```

## 规则配置查看

查看完整 ESLint 配置:

```bash
npx eslint --print-config src/app/page.tsx
```

## 相关文档

- [设计系统规范](../style/STYLE.md) - 设计系统整体规范
- [颜色系统](../style/color-system.md) - 颜色使用指南
- [Tailwind 使用规范](../style/tailwind-usage-policy.md) - Tailwind CSS 使用策略
- [字体排版系统](../../src/lib/typography.ts) - 字体规范实现
