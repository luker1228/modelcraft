---
paths:
  - "src/app/org/**/*.tsx"
---

# Frontend Layout: Unified Page Container Width

All content pages must use `max-w-7xl mx-auto p-6` as the standard content container — never fixed narrow widths like `max-w-xl`, `max-w-2xl`, or `max-w-3xl`.

## Requirements

- Use `max-w-7xl mx-auto p-6` for the content container on all pages
- Outer wrapper is `h-full overflow-auto bg-white` — simple, no decorative elements
- **No decorative background blobs** — do not add blur-3xl gradient balls to page backgrounds
- Cards and tables inside the container use `rounded-lg border bg-white` — not `rounded-xl bg-sidebar shadow-sm`

## Standard Page Shell

Every page must follow this exact outer structure:

```tsx
return (
  <div className="h-full overflow-auto bg-white">
    <div className="max-w-7xl mx-auto p-6">

      {/* Page title — always first */}
      <div className="mb-6">
        <h1 className="font-heading text-xl font-semibold tracking-tight text-slate-900">
          页面标题
        </h1>
      </div>

      {/* Page content */}

    </div>
  </div>
)
```

## Examples

### ✅ Good — clean, matches workspace and teams pattern

```tsx
return (
  <div className="h-full overflow-auto bg-white">
    <div className="max-w-7xl mx-auto p-6">
      <div className="mb-6">
        <h1 className="font-heading text-xl font-semibold tracking-tight text-slate-900">
          数据库集群
        </h1>
      </div>
      {/* content */}
    </div>
  </div>
)
```

### ❌ Bad — decorative blobs, wrong background, narrow width

```tsx
// ❌ Decorative blobs — do not add
return (
  <div className="h-full relative overflow-hidden bg-white">
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      <div className="absolute top-0 -left-4 w-72 h-72 bg-gradient-to-br from-blue-400/10 to-cyan-400/10 rounded-full blur-3xl animate-pulse" />
      <div className="absolute top-1/4 -right-12 w-96 h-96 bg-gradient-to-br from-violet-400/8 to-purple-400/8 rounded-full blur-3xl" />
    </div>
    <div className="h-full overflow-auto relative z-10">
      ...
    </div>
  </div>
)
```

```tsx
// ❌ Narrow width and wrong background
return (
  <div className="h-full overflow-auto bg-sidebar">
    <div className="max-w-xl mx-auto px-6 py-8">
      ...
    </div>
  </div>
)
```

## Card Style Inside Container

```tsx
// ✅ Correct
<div className="rounded-lg border bg-white overflow-hidden">
  ...
</div>

// ❌ Wrong
<div className="bg-sidebar rounded-xl border border-border shadow-sm overflow-hidden">
  ...
</div>
```

## Reference Pages

| Page | Path | Container |
|------|------|-----------|
| 工作台 | `workspace/page.tsx` | `max-w-7xl mx-auto p-6` ✅ |
| 团队 | `team/page.tsx` | `max-w-7xl mx-auto p-6` ✅ |
| 集群 | `cluster/page.tsx` | `max-w-7xl mx-auto p-6` ✅ |

## Rationale

Consistent container width and clean white backgrounds ensure all pages feel like part of the same product. Decorative blob backgrounds are reserved for the login/marketing page only — app content pages should be clean and distraction-free.
