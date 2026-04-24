---
name: ModelCraft
description: Data model and access-control configuration platform for enterprise ops teams
colors:
  action-blue: "#2563eb"
  action-blue-surface: "#dbeafe"
  canvas: "#fafafa"
  surface: "#ffffff"
  ink-deep: "#050d1f"
  ink-mid: "#374151"
  ink-muted: "#6b7280"
  structure-border: "#e2e6ec"
  structure-muted: "#f1f3f7"
  selected-state: "#dadee5"
  signal-critical: "#f04343"
  signal-success: "#10b981"
  signal-warning: "#f59e0b"
typography:
  display:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "-0.01em"
  headline:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.005em"
  title:
    fontFamily: "Space Grotesk, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 600
    lineHeight: 1.4
  body:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "0.01em"
  mono:
    fontFamily: "Fira Code, monospace"
    fontSize: "0.8125rem"
    fontWeight: 400
    lineHeight: 1.6
rounded:
  sm: "4px"
  md: "6px"
  lg: "8px"
  xl: "12px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.action-blue}"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
  button-primary-hover:
    backgroundColor: "#1d4ed8"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
  button-outline:
    backgroundColor: "transparent"
    textColor: "{colors.ink-mid}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.ink-muted}"
    rounded: "{rounded.md}"
    padding: "8px 12px"
    height: "36px"
  button-ghost-hover:
    backgroundColor: "{colors.structure-muted}"
    textColor: "{colors.ink-deep}"
    rounded: "{rounded.md}"
  badge-default:
    backgroundColor: "{colors.action-blue}"
    textColor: "{colors.surface}"
    rounded: "9999px"
    padding: "2px 10px"
  badge-secondary:
    backgroundColor: "{colors.action-blue-surface}"
    textColor: "{colors.action-blue}"
    rounded: "9999px"
    padding: "2px 10px"
  badge-outline:
    backgroundColor: "transparent"
    textColor: "{colors.ink-mid}"
    rounded: "9999px"
    padding: "2px 10px"
  input-default:
    backgroundColor: "transparent"
    textColor: "{colors.ink-deep}"
    rounded: "{rounded.md}"
    padding: "4px 12px"
    height: "36px"
  card-default:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-deep}"
    rounded: "{rounded.lg}"
    padding: "24px"
---

# Design System: ModelCraft

## 1. Overview

**Creative North Star: "The Engineering Blueprint"**

ModelCraft's visual system is built for operators, not audiences. Every screen is a structured working surface: hierarchy is visible, states are unambiguous, and nothing decorates itself at the expense of the task at hand. The metaphor is a well-drafted technical drawing — each element occupies exactly the space it needs, labels are terse and accurate, and white space exists to separate, not impress.

The type system runs on two distinct voices. Space Grotesk carries headings and structural labels with a quiet geometric confidence — the kind of precision a brief has, not the kind a poster has. Inter handles everything at reading scale: body copy, table rows, input labels. Fira Code appears wherever a string is also a technical value (model names, field identifiers, table slugs). Together they create density without noise.

The color system is deliberately narrow. Action blue (#2563eb) is the only saturated color with semantic weight. It appears on primary buttons, active nav items, and selected states — never decoratively. Status colors (critical, success, warning) exist only in explicit feedback contexts and never coexist with action blue on the same element. The rest of the palette is cool-leaning neutrals that hold content without competing with it.

**Key Characteristics:**
- Flat surfaces with structural borders; shadows used only for floating layers (dropdowns, dialogs, sheets)
- Two-font system: display weight for structure, text weight for content
- Blue as action color only; no purely decorative use of hue
- Spacing rhythm varies deliberately: compact in navigation (nav items scan fast), generous in data tables (py-3 rows) and forms, with visible breathing room between section groups
- States are always explicit: selected, hover, loading, empty, error each have a distinct visual signature

## 2. Colors: The Blueprint Palette

A narrow, functional palette where the neutrals do the heavy lifting and action blue earns every appearance.

### Primary
- **Action Blue** (`#2563eb`): Primary button fills, active sidebar nav, focus rings, selected-state accents, and link text. Never used as background outside of button fills and nav states.
- **Action Blue Surface** (`#dbeafe`): Tinted background for selected nav items, secondary badges, and tag chips. Pairs with action-blue text. No standalone use as a decorative fill.

### Neutral
- **Canvas** (`#fafafa`): Page-level background. The ground surface on which all panels sit.
- **Surface** (`#ffffff`): Cards, table backgrounds, sidebar, popover. Lifted one step above canvas.
- **Ink Deep** (`#050d1f`): Primary text: headings, labels, table cell primary content.
- **Ink Mid** (`#374151`): Secondary text: descriptions, supporting copy, outlined button labels.
- **Ink Muted** (`#6b7280`): De-emphasized text: placeholder copy, metadata, icon fills at rest.
- **Structure Border** (`#e2e6ec`): Table borders, card borders, dividers, input outlines.
- **Structure Muted** (`#f1f3f7`): Alternating table row backgrounds, hover states for ghost buttons and nav items, disabled fills.
- **Selected State** (`#dadee5`): Explicit selected / active state background for list items and nav entries. More visible than muted, less prominent than primary.

### Signal
- **Critical** (`#f04343`): Destructive actions, error messages, error-state input rings. Appears in badge-destructive and inline form errors. Never decorative.
- **Success** (`#10b981`): Success toasts, status badges, confirmation states. Never paired with action blue.
- **Warning** (`#f59e0b`): Advisory states, cautionary badges. Reserved for genuine warnings only.

### Named Rules
**The One Voice Rule.** Action blue is the single saturated color with behavioral meaning. Any other use of hue is a signal color (critical / success / warning). If you are reaching for a second accent color for variety, you are wrong. The palette's restraint is not a limitation; it is the point.

**The Signal Purity Rule.** Status colors (critical, success, warning) exist only when communicating a state. A green badge on a healthy resource is correct. A green button for a non-destructive action is not.

## 3. Typography

**Display / Heading Font:** Space Grotesk (system-ui fallback, sans-serif)
**Body / UI Font:** Inter (system-ui fallback, sans-serif)
**Mono Font:** Fira Code (monospace)

**Character:** Space Grotesk brings a structured, slightly unconventional geometry that reads as precise without being cold. Inter at small sizes is pure instrument: invisible utility, no ego. Fira Code signals technical identity for any value that is simultaneously a display string and a machine-readable identifier.

### Hierarchy
- **Display** (Space Grotesk, 600, 1.5rem/24px, line-height 1.25, tracking -0.01em): Page-level headings. Used once per view for the primary landmark ("权限管理", "数据模型"). Never duplicated.
- **Headline** (Space Grotesk, 600, 1.25rem/20px, line-height 1.3, tracking -0.005em): Section titles within a page, card titles, sheet headers. The `font-heading` utility class.
- **Title** (Space Grotesk, 600, 1rem/16px, line-height 1.4): Sub-section labels, table panel headers, form group labels.
- **Body** (Inter, 400, 0.875rem/14px, line-height 1.5): Table cell content, form field values, description text. Maximum line length 65–75ch on prose surfaces.
- **Label** (Inter, 500, 0.75rem/12px, line-height 1.4, tracking 0.01em): Column headers, metadata tags, badge text, compact UI labels. Never in all-caps.
- **Mono** (Fira Code, 400, 0.8125rem/13px, line-height 1.6): Technical identifiers: model names, field slugs, SQL expressions, IDs. Applied via `font-mono` utility.

### Named Rules
**The Serif Exception.** The `.font-label` utility (Times New Roman) exists as a section label ornament — a nod to the engineering-document register. Use it exclusively for static, decorative section labels, not for any interactive or data-bearing text.

**The Scale Minimum.** No text appears at less than 12px (0.75rem). Below this threshold, the label is not legible in standard office lighting; it is noise.

## 4. Elevation

This system is flat by default. Surfaces are differentiated primarily through background color contrast (canvas → surface) and structural borders, not shadow depth. Shadows are reserved for floating or detached layers, not for resting state surface hierarchy.

**Scene context:** An operator at a desk in an office, 27-inch monitor, standard ambient light, mid-day. Shadows that work in a darkened showroom become noise in this environment. The response is a flat system that reads clearly in any ambient condition.

### Shadow Vocabulary
- **Ambient Lift** (`shadow-sm`: `0 1px 2px rgba(0,0,0,0.05)`): Subtle differentiation on interactive surfaces at rest. Used on default buttons. Functional minimum; not a visual statement.
- **Overlay** (`shadow-md`: `0 4px 6px -1px rgba(0,0,0,0.1)`): Popovers, select dropdowns, context menus. Signals detachment from the document flow.
- **Modal** (`shadow-lg`: `0 10px 15px -3px rgba(0,0,0,0.1)`): Dialogs, sheets, alert dialogs. The highest elevation layer.

### Named Rules
**The Flat-By-Default Rule.** Cards, table rows, sidebar items, nav entries — all flat at rest. Shadows appear only when an element is elevated above the document flow (floating) or has lifted into a hover/active state that needs spatial separation. If you are adding a shadow to a resting card, you are decorating, not communicating.

## 5. Components

### Buttons
The button vocabulary is small and intentional. Each variant has exactly one behavioral role.

- **Shape:** Gently rounded (6px radius). Not pill-shaped; not sharp-cornered.
- **Primary** (`bg-primary text-white shadow-sm`, height 36px, padding 8px 16px): The single most consequential action on any given surface. Only one primary button per page section.
- **Hover / Focus:** Primary darkens to `#1d4ed8`; focus ring is 1px `ring-ring`. No scale transforms, no glow effects.
- **Outline** (`border border-input bg-background`, height 36px, padding 8px 16px): Secondary actions that need presence without dominance. Used for "编辑", "添加", cancel controls.
- **Ghost** (`hover:bg-muted`, height 36px, padding 8px 12px): Low-emphasis actions in dense contexts (table row actions, icon buttons). Background appears only on hover.
- **Destructive** (`bg-destructive text-white`): Reserved for confirmed irreversible actions inside AlertDialog confirmations. Never as the first action presented.
- **Size scale:** default (h-9), sm (h-8, px-3, text-xs), icon (36×36, padding 0).

### Badges / Status Chips
Fully-rounded pills (9999px) for labeling and status communication. Text at 0.75rem, font-weight 600.

- **Default** (action-blue fill): Rare. Reserved for primary state labels like active/enabled.
- **Secondary** (blue-100 surface, blue-600 text): Most common. Selected/active state labels, system role tags.
- **Outline** (transparent fill, border, foreground text): Neutral classification tags. "系统", "普通".
- **Success / Warning / Destructive**: Signal colors only. Status badges for health, warnings, errors.

### Cards / Containers
- **Corner Style:** Softly rounded (8px, `rounded-lg`).
- **Background:** Surface white (`#ffffff`), sitting on canvas (`#fafafa`).
- **Shadow:** `shadow-sm` at rest. Not elevated further.
- **Border:** `border border-border` (`#e2e6ec`). Present and structural; not decorative.
- **Internal Padding:** 24px all sides (`p-6`) for card panels. `p-4` for compact content areas inside settings.

### Tables
The primary data surface. Tables are flat, borderless between rows, with hairline separators.

- **Header row:** `bg-muted/50`, `font-semibold text-muted-foreground`, 12px label size.
- **Body rows:** alternating between `bg-background` and `bg-muted/20`. Border-bottom on each row except last.
- **Hover state:** `hover:bg-muted/50` or `hover:bg-slate-50`. No animated fills.
- **Action column:** right-aligned ghost buttons. Visible on row hover in compact tables, always visible in explicit management tables.
- **Empty state:** centered, icon + headline + brief instruction. No tables without an empty state.

### Inputs / Fields
- **Style:** Transparent background, single border (`border border-input`), 6px radius, height 36px.
- **Focus:** 1px `ring-ring` (action blue) inset glow. No animated border-width change.
- **Error:** `ring-destructive` focus ring; red helper text below.
- **Disabled:** `opacity-50 cursor-not-allowed`. Background unchanged.

### Navigation (App Sidebar)
- **Structure:** Left sidebar, 100% white surface, hairline right border. Fixed width at rest.
- **Nav items:** 36px tap target, 12px horizontal padding, `rounded-md`. Icon left + text label.
- **Default state:** `text-muted-foreground` icon and text.
- **Hover:** `bg-muted text-foreground` or `hover:bg-gray-100`.
- **Active:** `bg-blue-100 text-blue-600`. The selected-state color pulls from action-blue-surface. No underlines, no side-stripe accents.
- **Section headers:** `text-xs font-semibold text-foreground` in all-lowercase (not uppercase). 8px top padding above first item in section.

### Sheets / Drawers
- **Trigger:** from the right, `sm:max-w-lg` (512px). Slides in with `ease-in-out`, 500ms open / 300ms close.
- **Background:** `bg-background` (surface white). `shadow-lg` on the leading edge.
- **Header:** `SheetTitle` in headline weight. `SheetDescription` in body/muted. 24px bottom margin before content.
- **Scrollable content area:** `ScrollArea` for long lists. Fixed header and action bar, scrolling body.

### Dialogs / Alerts
- **Usage boundary:** Dialogs for forms requiring focused attention (create, edit with multiple fields). AlertDialog exclusively for destructive confirmations. Inline alternatives exhausted first.
- **Max width:** 32rem (`sm:max-w-md`) for single-field dialogs; `sm:max-w-lg` for multi-step forms.
- **Shadow:** `shadow-lg`.
- **Destructive confirm button:** `bg-destructive text-destructive-foreground`. Cancel comes first visually.

## 6. Do's and Don'ts

### Do:
- **Do** use action blue (`#2563eb`) only on primary interactive elements: buttons, active nav, focus rings, and selected-state accents. Its scarcity is what makes it legible as "this is the action."
- **Do** flatten surfaces at rest. Cards, panels, and table rows have border definition, not shadow lift. Reserve `shadow-md` and `shadow-lg` for floating layers.
- **Do** use `font-heading` (Space Grotesk) for structural titles and `font-sans` (Inter) for all content text. The distinction is consistent throughout the product.
- **Do** show explicit states for every interactive surface: empty, loading, error, success. A component with no empty state is an unfinished component.
- **Do** use `ghost` and `outline` button variants for table row actions. Primary buttons belong to primary actions on a screen, not to every row in a table.
- **Do** keep text at or above 0.75rem (12px). Below this, the content is not communication; it is visual noise.
- **Do** vary spacing deliberately: 16px between form fields, 24px between sections, 8px inside compact nav items. Uniform padding everywhere produces monotony, not order.

### Don't:
- **Don't** use Oracle/SAP-style thick panels, deeply layered card shadows, or heavy chrome. If the UI looks like enterprise software from 2012, the shadows are too dark and the containers too nested.
- **Don't** use consumer-product softness — Canva/Shopify-style gradients, candy-colored badges, rounded-2xl pills everywhere, or encouraging/gamified copy. This is a professional tool.
- **Don't** use glassmorphism, backdrop-filter blurs, gradient-filled cards, or animated gradient text. These are decorations that fight legibility on a working surface.
- **Don't** use `border-left` or `border-right` greater than 1px as a colored accent stripe on cards, list items, or alerts. Rewrite with a tinted background or full border.
- **Don't** use background-clip gradient text. Use a single solid color; add emphasis via weight or size.
- **Don't** place action blue on backgrounds, headers, or decorative elements. Its appearance should predict interactivity or selection.
- **Don't** invent a third accent color. The palette is action blue + neutrals + signal colors. There is no room for a second brand color at this stage.
- **Don't** use information density so low that users must scroll through three screens to find one configuration option. Appropriate density is a feature.
