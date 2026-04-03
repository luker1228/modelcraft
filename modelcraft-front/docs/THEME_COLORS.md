# Unified Theme Colors

## Overview

This document explains the unified color system used across the ModelCraft application. All background colors for sidebar, cards, and main content areas are centralized in CSS variables for easy global updates.

## CSS Variables

Located in `src/app/globals.css` under `@layer base`:

### Light Mode (`:root`)
```css
--container-bg: 0 0% 100%;           /* Pure white */
--sidebar-background: 0 0% 100%;      /* Pure white */
```

### Dark Mode (`.dark`)
```css
--container-bg: 240 5.9% 10%;        /* Dark background */
--sidebar-background: 240 5.9% 10%;   /* Dark background */
```

## CSS Classes

Located in `src/app/globals.css` under `@layer components`:

- `.container-bg` - Basic container background
- `.container-bg-blur` - Container background with blur effect (glass morphism)

## Tailwind Classes

The theme uses these Tailwind classes for backgrounds:
- `bg-sidebar` - Maps to `--sidebar-background`
- `bg-selected` - Maps to `--selected` for hover/active states
- `hover:bg-sidebar/95` - Hover state with reduced opacity

## TypeScript Constants

Located in `src/lib/theme-colors.ts`:

```typescript
import { 
  CONTAINER_BG_CLASS,
  CONTAINER_BG_BLUR_CLASS,
  CONTAINER_BG_HOVER_CLASS,
  CARD_CONTAINER_CLASS
} from '@/lib/theme-colors'
```

## Usage Examples

### 1. Using Tailwind Classes (Recommended)

```tsx
// Simple background
<div className="bg-sidebar">Content</div>

// With blur effect
<div className="bg-sidebar backdrop-blur-sm">Content</div>

// Card styling
<Card className="group bg-sidebar backdrop-blur-sm border-0 shadow-md hover:bg-sidebar/95 transition-all cursor-pointer">
  Content
</Card>
```

### 2. Using TypeScript Constants

```tsx
import { CARD_CONTAINER_CLASS } from '@/lib/theme-colors'

<Card className={CARD_CONTAINER_CLASS}>
  Content
</Card>
```

### 3. Using CSS Variables

```css
.my-component {
  background-color: hsl(var(--container-bg));
}

.my-component-with-blur {
  background-color: hsl(var(--container-bg));
  backdrop-filter: blur(4px);
}
```

## Components Using Unified Colors

### Light Mode (Default)
- **Sidebar** - `bg-sidebar` (white)
- **Cards** - `bg-sidebar` (white)
- **Input Fields** - `bg-sidebar` (white)
- **Buttons** - `bg-sidebar` (white)
- **Dropdowns** - `bg-sidebar` (white)

### Dark Mode
- **Sidebar** - `bg-sidebar` (dark)
- **Cards** - `bg-sidebar` (dark)
- **All other components** - Automatically inherit dark theme

## Files Using Unified Colors

1. **src/components/project/ProjectCard.tsx**
   - Project card component using `bg-sidebar`

2. **src/app/org/[orgName]/workspace/page.tsx**
   - Project cards, search input, buttons using `bg-sidebar`

3. **Other pages and components**
   - Dashboard cards
   - Guide page cards
   - CMS page components

## How to Update All Colors Globally

### Option 1: Update CSS Variables (Recommended)

Edit `src/app/globals.css`:

```css
:root {
  /* Change light mode background */
  --container-bg: 0 0% 96%;  /* Changed from pure white */
}

.dark {
  /* Change dark mode background */
  --container-bg: 240 5.9% 12%;  /* Changed from previous dark */
}
```

All components automatically update because they use `bg-sidebar` which maps to `--sidebar-background`, which is set to the value of `--container-bg`.

### Option 2: Update Sidebar CSS Variables

Edit `src/app/globals.css`:

```css
:root {
  --sidebar-background: 0 0% 96%;  /* All sidebar components update */
}
```

### Option 3: Update in Code

If you need to change a specific component, use the Tailwind classes directly:

```tsx
// Before (using unified class)
<Card className="bg-sidebar">Content</Card>

// After (specific color)
<Card className="bg-blue-50">Content</Card>
```

## Related CSS Variables

For reference, here are other related theme variables:

- `--selected` - Hover/active state color (215 20% 88%)
- `--tech-bg` - Tech background pattern (0 0% 98%)
- `--sidebar-accent` - Maps to `--selected`
- `--sidebar-border` - Border color (220 13% 91%)

## Notes

1. **Synchronization**: `--container-bg` and `--sidebar-background` should always have the same value for consistency.

2. **Transparency**: When using `bg-sidebar/70` or `bg-sidebar/80`, ensure you're intentionally creating a transparent effect (e.g., for glass morphism).

3. **Blur Effects**: Pair transparent backgrounds with `backdrop-blur-sm` or `backdrop-blur` for glass morphism effects.

4. **Theme Switching**: The app automatically switches between light and dark mode based on the system setting or user preference. Ensure both `:root` and `.dark` CSS variables are defined.

## Future Enhancements

- [ ] Add CSS variable for secondary background color
- [ ] Add gradient combinations
- [ ] Add animation/transition timing variables
- [ ] Document color contrast ratios for accessibility
