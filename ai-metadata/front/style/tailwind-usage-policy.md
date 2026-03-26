# Tailwind CSS Usage Policy

Enforce consistent use of Tailwind utility classes as the primary styling approach, with clear rules for the rare cases that require other techniques.

## Requirements

### 基础与布局（100% Tailwind）

- **所有**基础样式和布局必须使用 Tailwind 工具类
- 禁止在组件文件中用行内 `style` 定义静态布局（如 `style={{ display: 'flex' }}`，改用 `className="flex"`）
- 禁止在 `*.tsx` 文件中使用 `<style>` 标签定义布局样式

### 组件样式（95% Tailwind）

- 组件样式默认使用 Tailwind 工具类
- 对于**高度复用的复杂样式**，使用 `@apply` 提取为自定义工具类，定义在 `@layer components` 中
- `@apply` 样式统一放入 `src/app/globals.css` 或专用的 `src/styles/components.css`，禁止分散在各组件文件中

### 动态样式

- 动态值（如用户自定义颜色、运行时计算的尺寸）：结合 **CSS 变量** + 行内 `style` 属性
- CSS 变量优先在 `tailwind.config.js` 的 `theme.extend` 中定义，使其可通过工具类引用
- 禁止在行内 `style` 中硬编码静态值（应改用工具类）

### 复杂动画 / 特定选择器

- 复杂关键帧动画（`@keyframes`）和特定伪类/伪元素选择器：写在 `globals.css` 或通过 Tailwind 插件扩展
- 简单动画优先使用 Tailwind 内置工具类（`animate-spin`、`animate-pulse` 等）

### 覆盖第三方库

- 覆盖第三方库样式必须创建**独立文件**（如 `src/styles/overrides.css`），不得混入组件文件
- 通过提高 CSS 特异性覆盖；仅在特异性无效时谨慎使用 `!important`
- 在文件顶部注释说明覆盖的目标库和原因

## Examples

### ✅ Good — 100% Tailwind 布局

```tsx
export function Card({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-4 rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
      {children}
    </div>
  )
}
```

### ✅ Good — @apply 提取复用样式（globals.css）

```css
/* src/app/globals.css */
@layer components {
  .card-base {
    @apply rounded-lg border border-gray-200 bg-white p-6 shadow-sm;
  }

  .input-base {
    @apply w-full rounded-md border border-gray-300 px-3 py-2 text-sm
           focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }
}
```

```tsx
export function Input(props: InputProps) {
  return <input className="input-base" {...props} />
}
```

### ✅ Good — 动态样式使用 CSS 变量 + 行内 style

```tsx
// 仅运行时动态值才用行内 style
export function BrandButton({ color }: { color: string }) {
  return (
    <button
      className="rounded px-4 py-2 text-white"
      style={{ backgroundColor: color }}
    >
      Click
    </button>
  )
}
```

### ✅ Good — 复杂动画在 globals.css

```css
/* src/app/globals.css */
@keyframes slide-in-from-right {
  from { transform: translateX(100%); opacity: 0; }
  to   { transform: translateX(0);    opacity: 1; }
}

.animate-slide-in {
  animation: slide-in-from-right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

### ✅ Good — 第三方库覆盖独立文件

```css
/* src/styles/overrides.css */
/* 覆盖目标：react-datepicker v4.x — 修正弹出层 z-index 以适配 Modal */
.react-datepicker-popper {
  z-index: 9999 !important;
}
```

---

### ❌ Bad — 在组件中写静态行内样式

```tsx
export function Card({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
      {children}
    </div>
  )
}
```

### ❌ Bad — @apply 分散在组件文件中

```tsx
export function Button() {
  return (
    <>
      <style>{`.my-btn { @apply px-4 py-2 bg-blue-600 text-white rounded; }`}</style>
      <button className="my-btn">Click</button>
    </>
  )
}
```

### ❌ Bad — 第三方覆盖混入全局样式

```css
/* src/app/globals.css */
body { font-family: 'Inter', sans-serif; }
.react-datepicker { z-index: 9999 !important; } /* ← 应放入 overrides.css */
```

## Rationale

- **一致性**：统一工具类使样式决策集中化，减少命名冲突和特异性问题
- **可维护性**：`@apply` 集中在 `@layer components`，修改一处影响所有复用点
- **可读性**：工具类直接在 JSX 中可见，无需跳转到单独 CSS 文件
- **第三方隔离**：独立的 `overrides.css` 使覆盖规则易于追踪和清理
