# Component Design Quick Start

> ModelCraft B2B design — solid colors, 1px borders, no gradients, no glass-morphism.

---

## 1. Primary Action Button

```tsx
<Button className="bg-[#2563eb] hover:bg-[#1d4ed8] text-white border-0 h-9 px-4 rounded-md font-medium text-sm gap-2 transition-all duration-200">
  <Plus className="w-4 h-4" strokeWidth={1.5} />
  创建项目
</Button>
```

## 2. Secondary / Cancel Button

```tsx
<Button variant="outline" className="bg-[#fafafa] hover:bg-white text-gray-900 border border-gray-200 hover:border-gray-300 h-9 px-4 rounded-md font-medium text-sm transition-all duration-200">
  取消
</Button>
```

## 3. Destructive Button

```tsx
<Button className="bg-[#ef4444] hover:bg-[#dc2626] text-white border-0 h-9 px-4 rounded-md font-medium text-sm gap-2 transition-all duration-200">
  <Trash2 className="w-4 h-4" strokeWidth={1.5} />
  删除
</Button>
```

## 4. Content Card

```tsx
<div className="bg-white border border-gray-200 rounded-lg p-4 transition-all duration-200 hover:border-blue-100 hover:shadow-sm">
  <h3 className="text-base font-semibold text-gray-900 mb-2">Card Title</h3>
  <p className="text-sm text-gray-500 mb-3">Card description or metadata.</p>
  <div className="flex items-center gap-2 mb-3">
    <span className="inline-flex items-center px-3 py-1 rounded bg-[#ecfdf5] text-[#059669] text-xs font-semibold">
      Active
    </span>
  </div>
  <div className="text-xs text-gray-400">Updated 2 hours ago</div>
</div>
```

## 5. Status Badges

```tsx
{/* Success */}
<span className="inline-flex items-center px-3 py-1 rounded bg-[#ecfdf5] text-[#059669] text-xs font-semibold">Active</span>

{/* Warning */}
<span className="inline-flex items-center px-3 py-1 rounded bg-[#fef3c7] text-[#d97706] text-xs font-semibold">Draft</span>

{/* Error */}
<span className="inline-flex items-center px-3 py-1 rounded bg-[#fee2e2] text-[#ef4444] text-xs font-semibold">Error</span>

{/* Info / Primary */}
<span className="inline-flex items-center px-3 py-1 rounded bg-[#dbeafe] text-[#2563eb] text-xs font-semibold">In Progress</span>
```

## 6. Form Fields

```tsx
<div className="space-y-4">
  <div>
    <label className="block text-sm font-medium text-gray-900 mb-1.5">
      Project Name
    </label>
    <input
      type="text"
      placeholder="Enter project name"
      className="w-full border border-gray-200 rounded-md px-3 py-2 text-sm text-gray-900 bg-white
                 placeholder:text-gray-400 focus:outline-none focus:border-[#2563eb]
                 focus:ring-2 focus:ring-[rgba(37,99,235,0.1)] transition-all duration-200"
    />
  </div>

  <div>
    <label className="block text-sm font-medium text-gray-900 mb-1.5">
      Description
    </label>
    <textarea
      rows={4}
      placeholder="Optional description"
      className="w-full border border-gray-200 rounded-md px-3 py-2 text-sm text-gray-900 bg-white
                 placeholder:text-gray-400 focus:outline-none focus:border-[#2563eb]
                 focus:ring-2 focus:ring-[rgba(37,99,235,0.1)] transition-all duration-200 resize-none"
    />
  </div>

  <div className="flex gap-3">
    <Button className="bg-[#2563eb] hover:bg-[#1d4ed8] text-white ...">Create</Button>
    <Button variant="outline" className="...">Cancel</Button>
  </div>
</div>
```

## 7. Data Table

```tsx
<div className="border border-gray-200 rounded-lg overflow-hidden">
  <table className="w-full text-sm border-collapse">
    <thead className="bg-gray-50 border-b border-gray-200">
      <tr>
        <th className="text-left px-3 py-3 font-semibold text-gray-900">Name</th>
        <th className="text-left px-3 py-3 font-semibold text-gray-900">Status</th>
        <th className="text-left px-3 py-3 font-semibold text-gray-900">Actions</th>
      </tr>
    </thead>
    <tbody>
      {/* Normal row */}
      <tr className="border-b border-gray-200 hover:bg-gray-50 transition-colors">
        <td className="px-3 py-3 text-gray-500">Project Alpha</td>
        <td className="px-3 py-3">
          <span className="inline-flex items-center px-3 py-1 rounded bg-[#ecfdf5] text-[#059669] text-xs font-semibold">
            Active
          </span>
        </td>
        <td className="px-3 py-3">
          <Button variant="ghost" size="icon" className="w-8 h-8 hover:bg-gray-100">
            <MoreHorizontal className="w-4 h-4" strokeWidth={1.5} />
          </Button>
        </td>
      </tr>
      {/* Selected row — use #dadee5, NOT rgba blue */}
      <tr className="border-b border-gray-200 bg-[#dadee5]">
        <td className="px-3 py-3 text-gray-500">Project Beta</td>
        ...
      </tr>
    </tbody>
  </table>
</div>
```

## 8. Alert Messages

```tsx
{/* Success */}
<div className="flex items-start gap-3 px-4 py-3 rounded-md border border-[rgba(5,150,105,0.2)] bg-[#ecfdf5] text-[#059669] text-sm mb-4">
  <CheckCircle className="w-4 h-4 flex-shrink-0 mt-0.5" strokeWidth={1.5} />
  <span>Operation completed successfully.</span>
</div>

{/* Warning */}
<div className="flex items-start gap-3 px-4 py-3 rounded-md border border-[rgba(217,119,6,0.2)] bg-[#fef3c7] text-[#d97706] text-sm mb-4">
  <AlertTriangle className="w-4 h-4 flex-shrink-0 mt-0.5" strokeWidth={1.5} />
  <span>Warning message here.</span>
</div>

{/* Error */}
<div className="flex items-start gap-3 px-4 py-3 rounded-md border border-[rgba(239,68,68,0.2)] bg-[#fee2e2] text-[#ef4444] text-sm mb-4">
  <XCircle className="w-4 h-4 flex-shrink-0 mt-0.5" strokeWidth={1.5} />
  <span>Error message here.</span>
</div>

{/* Info */}
<div className="flex items-start gap-3 px-4 py-3 rounded-md border border-[rgba(37,99,235,0.2)] bg-[#dbeafe] text-[#2563eb] text-sm mb-4">
  <Info className="w-4 h-4 flex-shrink-0 mt-0.5" strokeWidth={1.5} />
  <span>Informational message here.</span>
</div>
```

## 9. Dropdown Menu

```tsx
<DropdownMenuContent align="end" className="w-56 border border-gray-200 shadow-md rounded-lg bg-white" sideOffset={8}>
  <DropdownMenuLabel>
    <div className="flex flex-col space-y-0.5">
      <p className="text-sm font-semibold text-gray-900">User Name</p>
      <p className="text-xs text-gray-500">user@example.com</p>
    </div>
  </DropdownMenuLabel>
  <DropdownMenuSeparator className="border-gray-200" />
  <DropdownMenuItem className="cursor-pointer text-sm text-gray-700 focus:bg-gray-100 focus:text-gray-900">
    <Settings className="mr-2 h-4 w-4" strokeWidth={1.5} />
    Settings
  </DropdownMenuItem>
  <DropdownMenuItem className="cursor-pointer text-sm text-gray-700 focus:bg-gray-100 focus:text-gray-900">
    <LogOut className="mr-2 h-4 w-4" strokeWidth={1.5} />
    Sign out
  </DropdownMenuItem>
</DropdownMenuContent>
```

> **No colored icons in menus** — no red for logout, no gradients on icon containers. Plain gray icons only.

---

## Decision Trees

### Which button type?

```
Is it the main/primary action?
├─ YES → Primary (blue solid)
└─ NO
   └─ Is it dangerous (delete, remove)?
      ├─ YES → Destructive (red solid)
      └─ NO
         └─ Is it a dismissal (cancel, close)?
            ├─ YES → Secondary (outlined)
            └─ NO → Ghost (transparent)
```

### Which badge color?

```
Success / active / connected → Success (green bg)
Warning / draft / pending    → Warning (amber bg)
Error / failed / disabled    → Error (red bg)
Info / in-progress / primary → Primary (blue bg)
```

### What background for selected state?

```
Row selected / item active → bg-[#dadee5]   ✅
                           → bg-blue-50     ❌ (too light)
                           → rgba blue      ❌ (too light)
```

### Input focus ring?

```
focus:border-[#2563eb] focus:ring-2 focus:ring-[rgba(37,99,235,0.1)]
```

---

## Common Mistakes to Avoid

| Wrong | Correct |
|-------|---------|
| `bg-gradient-to-r from-blue-600 to-indigo-600` on button | `bg-[#2563eb] hover:bg-[#1d4ed8]` |
| `bg-white/70 backdrop-blur-sm` on card | `bg-white border border-gray-200` |
| `shadow-blue-500/30` on button | No colored shadow on buttons |
| Red icon on "logout" menu item | Plain gray icon `text-gray-700` |
| `bg-blue-50` for selected row | `bg-[#dadee5]` |
| `font-medium` on body text | `font-normal` |
| Decorative blob `blur-3xl animate-pulse` | Nothing — clean background |
