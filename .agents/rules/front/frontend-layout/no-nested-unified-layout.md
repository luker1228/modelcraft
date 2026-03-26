---
paths:
  - "src/app/org/[orgName]/project/[projectName]/**/*.tsx"
---

# Frontend Layout: No Nested UnifiedLayout in Project Pages

Project-level pages must NOT wrap content in `<UnifiedLayout>`. The layout is already provided by the shared `layout.tsx`.

## Requirements

- **Never** import or use `<UnifiedLayout>` in page components under `src/app/org/[orgName]/project/[projectName]/`
- The file `src/app/org/[orgName]/project/[projectName]/layout.tsx` already renders `<UnifiedLayout showProjectNav>` — adding it again in a page causes two `BreadcrumbHeader` topbars to render
- Page components under this path must return content directly (a plain `<div>` as root)

## Examples

### ✅ Good — page returns content directly

```tsx
// src/app/org/[orgName]/project/[projectName]/clusters/page.tsx
export default function ClustersPage() {
  return (
    <div className="h-full overflow-auto bg-sidebar">
      <div className="max-w-3xl mx-auto px-6 py-8">
        {/* page content */}
      </div>
    </div>
  )
}
```

### ❌ Bad — UnifiedLayout nested inside layout.tsx's UnifiedLayout

```tsx
// src/app/org/[orgName]/project/[projectName]/clusters/page.tsx
import { UnifiedLayout } from '@/components/layout/UnifiedLayout'

export default function ClustersPage() {
  return (
    // ❌ layout.tsx already wraps this in <UnifiedLayout showProjectNav>
    // This produces TWO BreadcrumbHeader topbars
    <UnifiedLayout showProjectNav pageTitle="数据库集群">
      <div>...</div>
    </UnifiedLayout>
  )
}
```

## Where `<UnifiedLayout>` IS allowed

| Path | Rule |
|------|------|
| `src/app/org/[orgName]/workspace/page.tsx` | ✅ Allowed — no parent layout.tsx wrapping it |
| `src/app/org/[orgName]/team/page.tsx` | ✅ Allowed — no parent layout.tsx wrapping it |
| `src/app/org/[orgName]/project/[projectName]/layout.tsx` | ✅ This is the single source of layout for all project pages |
| `src/app/org/[orgName]/project/[projectName]/**/page.tsx` | ❌ Must NOT use `<UnifiedLayout>` |

## Rationale

`layout.tsx` in Next.js App Router wraps all child pages automatically. Importing `<UnifiedLayout>` again in a child page creates a double-render of the topbar and sidebar, producing a broken UI with two navigation headers stacked on top of each other.
