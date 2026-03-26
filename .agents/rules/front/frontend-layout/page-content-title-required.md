---
paths:
  - "src/app/org/**/*.tsx"
---

# Frontend Layout: Content Pages Require a Page Title

Every content page must render a visible `<h1>` title inside the page content area, below the topbar.

## Requirements

- Every page under `src/app/org/` must include an `<h1>` title in the content body
- Use `font-heading text-xl font-semibold tracking-tight text-slate-900` for consistency
- The title goes at the top of the scrollable content area, **inside** the page container — not inside `<UnifiedLayout>` props or the topbar
- The `pageTitle` prop on `<UnifiedLayout>` (which sets the breadcrumb) is **not** a substitute for an in-page `<h1>` title — both are required for org-level pages

## Examples

### ✅ Good — visible h1 title in page content

```tsx
// team/page.tsx
return (
  <UnifiedLayout pageTitle="团队">
    <div className="h-full overflow-auto">
      <div className="max-w-7xl mx-auto p-6">

        {/* ✅ Page title visible in content area */}
        <div className="mb-6">
          <h1 className="font-heading text-xl font-semibold tracking-tight text-slate-900">
            团队成员
          </h1>
        </div>

        {/* rest of content */}
      </div>
    </div>
  </UnifiedLayout>
)
```

```tsx
// workspace/page.tsx
return (
  <UnifiedLayout pageTitle="所有项目">
    <div className="h-full overflow-auto">
      <div className="max-w-7xl mx-auto p-6">

        {/* ✅ Page title visible in content area */}
        <div className="mb-6">
          <h1 className="font-heading text-xl font-semibold tracking-tight text-slate-900">
            所有项目
          </h1>
        </div>

        {/* search bar, actions, grid */}
      </div>
    </div>
  </UnifiedLayout>
)
```

### ❌ Bad — content starts directly with functional elements, no title

```tsx
return (
  <UnifiedLayout pageTitle="所有项目">
    <div className="h-full overflow-auto">
      <div className="max-w-7xl mx-auto p-6">

        {/* ❌ No h1 title — page jumps straight to search bar */}
        <div className="mb-6 flex items-center justify-between gap-4">
          <SearchInput placeholder="搜索项目..." />
          <Button>创建项目</Button>
        </div>
      </div>
    </div>
  </UnifiedLayout>
)
```

## Title Placement Pattern

```
┌─────────────────────────────────────────────────┐
│  BreadcrumbHeader (topbar — h-14)               │  ← provided by UnifiedLayout / layout.tsx
├─────────────────────────────────────────────────┤
│  <div className="max-w-7xl mx-auto p-6">        │
│    <div className="mb-6">                       │
│      <h1 className="font-heading text-xl        │  ← required page title
│           font-semibold tracking-tight          │
│           text-slate-900">                      │
│        页面标题                                  │
│      </h1>                                      │
│    </div>                                       │
│    {/* rest of page content */}                 │
│  </div>                                         │
└─────────────────────────────────────────────────┘
```

## Typography Class

Always use this exact class combination for page titles to maintain visual consistency:

```
font-heading text-xl font-semibold tracking-tight text-slate-900
```

| Class | Purpose |
|-------|---------|
| `font-heading` | Space Grotesk — the project's display/heading font |
| `text-xl` | 20px — standard page title size |
| `font-semibold` | 600 weight — prominent but not heavy |
| `tracking-tight` | Tighter letter spacing for headings |
| `text-slate-900` | Near-black — highest contrast |

## Rationale

Without an in-page `<h1>`, content pages feel like they're missing context — the user sees the breadcrumb in the topbar but the content area gives no visual anchor. The title provides hierarchy and orients the user when they land on or scroll back to the top of a page.
