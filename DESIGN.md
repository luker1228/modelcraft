---
name: ModelCraft
description: Data model and access-control configuration platform for enterprise ops teams
colors:
  action-indigo: "#4F46E5"
  action-indigo-hover: "#6366F1"
  action-indigo-surface: "rgba(79,70,229,0.08)"
  canvas: "#F6F8FA"
  surface: "#ffffff"
  ink-deep: "#1A1F36"
  ink-mid: "#697386"
  ink-muted: "#8792A2"
  structure-border: "#E3E8EE"
  structure-muted: "#EBEEF2"
  structure-muted-border: "#D8DDE5"
  selected-state: "rgba(79,70,229,0.08)"
  selected-foreground: "#4F46E5"
  signal-critical: "#f04343"
  signal-success: "#10b981"
  signal-warning: "#f59e0b"
  chart-1: "hsl(12, 76%, 61%)"
  chart-2: "hsl(173, 58%, 39%)"
  chart-3: "hsl(197, 37%, 24%)"
  chart-4: "hsl(43, 74%, 66%)"
  chart-5: "hsl(27, 87%, 67%)"
dark-mode:
  background: "hsl(224, 71.4%, 4.1%)"
  foreground: "hsl(210, 20%, 98%)"
  primary: "hsl(239, 84%, 67%)"
  primary-foreground: "hsl(0, 0%, 100%)"
  muted: "hsl(215, 27.9%, 16.9%)"
  muted-foreground: "hsl(217.9, 10.6%, 64.9%)"
  border: "hsl(215, 27.9%, 16.9%)"
  destructive: "hsl(0, 62.8%, 30.6%)"
sidebar:
  background: "#ffffff"
  foreground: "hsl(240, 5.3%, 26.1%)"
  primary: "hsl(240, 5.9%, 10%)"
  primary-foreground: "hsl(0, 0%, 98%)"
  accent: "rgba(79,70,229,0.08)"
  accent-foreground: "#1e2a3b"
  border: "hsl(220, 13%, 91%)"
  ring: "hsl(217.2, 91.2%, 59.8%)"
typography:
  display:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  headline:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  title:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 600
    lineHeight: 1.4
    letterSpacing: "-0.01em"
  body:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
  label:
    fontFamily: "Inter, system-ui, sans-serif"
    fontSize: "0.6875rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "0.06em"
    textTransform: "uppercase"
  mono:
    fontFamily: "Fira Code, monospace"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.6
rounded:
  sm: "4px"
  md: "6px"
  lg: "8px"
  xl: "12px"
  2xl: "16px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
breakpoints:
  container-max: "1400px"
  container-padding: "2rem"
motion:
  fade-in:
    duration: "0.25s"
    easing: "ease-out"
    transform: "translateY(4px) -> translateY(0)"
  slide-in-right:
    duration: "0.25s"
    easing: "ease-out"
    transform: "translateX(10px) -> translateX(0)"
  accordion:
    duration: "0.2s"
    easing: "ease-out"
  sheet-open:
    duration: "0.5s"
    easing: "ease-in-out"
  sheet-close:
    duration: "0.3s"
    easing: "ease-in-out"
components:
  button-primary:
    backgroundColor: "{colors.action-indigo}"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
    fontWeight: 500
  button-primary-hover:
    backgroundColor: "{colors.action-indigo-hover}"
    textColor: "{colors.surface}"
    rounded: "{rounded.md}"
  button-outline:
    backgroundColor: "transparent"
    textColor: "{colors.ink-deep}"
    border: "1px solid {colors.structure-border}"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    height: "36px"
    fontWeight: 500
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.ink-mid}"
    rounded: "{rounded.md}"
    padding: "8px 12px"
    height: "36px"
    fontWeight: 500
  button-ghost-hover:
    backgroundColor: "rgba(0,0,0,0.04)"
    textColor: "{colors.ink-deep}"
    rounded: "{rounded.md}"
  badge-default:
    backgroundColor: "{colors.action-indigo-surface}"
    textColor: "{colors.action-indigo}"
    rounded: "{rounded.sm}"
    padding: "0 7px"
    height: "20px"
    fontSize: "11px"
    fontWeight: 500
  badge-success:
    backgroundColor: "rgba(5,150,105,0.08)"
    textColor: "#059669"
    rounded: "{rounded.sm}"
  badge-warning:
    backgroundColor: "rgba(217,119,6,0.08)"
    textColor: "#D97706"
    rounded: "{rounded.sm}"
  badge-destructive:
    backgroundColor: "rgba(239,68,68,0.08)"
    textColor: "#EF4444"
    rounded: "{rounded.sm}"
  badge-neutral:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink-mid}"
    border: "1px solid {colors.structure-border}"
    rounded: "{rounded.sm}"
  input-default:
    backgroundColor: "transparent"
    textColor: "{colors.ink-deep}"
    border: "1px solid {colors.structure-border}"
    rounded: "{rounded.md}"
    padding: "4px 12px"
    height: "36px"
  input-focus:
    borderColor: "{colors.action-indigo}"
    boxShadow: "0 0 0 3px rgba(79,70,229,0.1)"
  toolbar-search:
    backgroundColor: "{colors.structure-muted}"
    border: "1px solid {colors.structure-muted-border}"
    rounded: "{rounded.md}"
    height: "30px"
    note: "搜索框比 toolbar 背景深一档，形成凹入质感"
  card-default:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-deep}"
    rounded: "{rounded.lg}"
    padding: "24px"
    shadow: "stripe-md"
  table-thead:
    backgroundColor: "{colors.surface}"
    borderBottom: "2px solid {colors.structure-border}"
    thColor: "{colors.ink-deep}"
    thFontSize: "11px"
    thFontWeight: 500
    thTextTransform: "uppercase"
    thLetterSpacing: "0.06em"
    note: "纯白背景 + 2px 前景色下边框，无灰色填充"
  table-toolbar:
    backgroundColor: "{colors.canvas}"
    note: "Toolbar 与页面背景同色 #F6F8FA"
---

# Design System: ModelCraft

## 1. Overview

**Creative North Star: "Stripe Dashboard — Precision Tool"**

ModelCraft's visual system is built for operators, not audiences. Every screen is a structured working surface: hierarchy is visible, states are unambiguous, and nothing decorates itself at the expense of the task at hand. The design philosophy is restraint — premium feel through removal, not addition.

The type system runs on a single voice: Inter at all scales. Headers, labels, body copy, and navigation all use Inter. Fira Code appears wherever a string is also a technical value (model names, field identifiers, table slugs, SQL). One font family creates cohesion without personality competition.

The color system is deliberately narrow. Action Indigo (#4F46E5) is the only saturated color with semantic weight. It appears on primary buttons, active nav items, tab underlines, and selected states — never decoratively. Status colors (critical, success, warning) exist only in explicit feedback contexts. The rest of the palette is cool blue-gray neutrals: canvas #F6F8FA, surface #FFFFFF, borders #E3E8EE.

**Key Characteristics:**
- Page background #F6F8FA (cool blue-gray); white surface cards float above with subtle multi-layer shadow
- Table headers: pure white background + 2px foreground-color bottom border; no gray fill
- Toolbar background matches page canvas (#F6F8FA); search input one step darker (#EBEEF2) for recessed feel
- Shadow-first hierarchy: cards use shadow-md, floating layers use shadow-lg; borders are structural not decorative
- Single font: Inter only; Fira Code for technical identifiers
- Indigo as action color only; no purely decorative use of hue
- Spacing rhythm: compact in navigation, comfortable in tables (48px rows) and forms
- States are always explicit: selected, hover, loading, empty, error each have a distinct visual signature
- Motion is purposeful and brief: 150ms ease-out for interactions, no decorative animation
- Container max-width 1400px, left-aligned content within

## 2. Colors: The Blueprint Palette

A narrow, functional palette where the neutrals do the heavy lifting and action blue earns every appearance.

### Primary
- **Action Indigo** (`#4F46E5`): Primary button fills, active sidebar nav left-stripe, tab underlines, focus rings, selected-state accents. Never used as background outside of button fills and nav states.
- **Action Indigo Hover** (`#6366F1`): Lightened indigo for hover states on primary buttons.
- **Action Indigo Surface** (`rgba(79,70,229,0.08)`): Tinted background for selected nav items, info badges, and selected-state chips. Pairs with action-indigo text.

### Neutral
- **Canvas** (`#F6F8FA`): Page-level background. Cool blue-gray, 2% darker than pure white. All panels sit on this surface.
- **Surface** (`#FFFFFF`): Cards, table backgrounds, sidebar, topbar, popover. Lifted above canvas via shadow.
- **Structure Muted** (`#EBEEF2`): Search input background inside toolbars. One step darker than canvas for recessed/inset feel. Border: `#D8DDE5`.
- **Ink Deep** (`#1A1F36`): Primary text: headings, table cell primary content, active thead labels.
- **Ink Mid** (`#697386`): Secondary text: descriptions, metadata, supporting copy.
- **Ink Muted** (`#8792A2`): De-emphasized text: placeholder copy, icon fills at rest, section headers.
- **Structure Border** (`#E3E8EE`): Table borders, card borders, dividers, input outlines. Cool blue-gray tint.
- **Selected State** (`rgba(79,70,229,0.08)`): Active background for sidebar nav items and selected list entries. Pairs with action-indigo text and left border accent.
- **Selected Foreground** (`#4F46E5`): Text/icon color paired with selected-state background.

### Signal
- **Critical** (`#f04343`): Destructive actions, error messages, error-state input rings. Appears in badge-destructive and inline form errors. Never decorative.
- **Success** (`#10b981`): Success toasts, status badges, confirmation states. Never paired with action blue.
- **Warning** (`#f59e0b`): Advisory states, cautionary badges. Reserved for genuine warnings only.

### Chart Colors
Five distinct hues for data visualization contexts only. Never used as UI accent colors.
- **Chart 1** (`hsl(12, 76%, 61%)`): Warm coral
- **Chart 2** (`hsl(173, 58%, 39%)`): Teal
- **Chart 3** (`hsl(197, 37%, 24%)`): Deep slate-blue
- **Chart 4** (`hsl(43, 74%, 66%)`): Amber
- **Chart 5** (`hsl(27, 87%, 67%)`): Orange

### Dark Mode

Dark mode inverts the surface hierarchy while preserving the same semantic roles. Key shifts:
- **Background** flips from `#fafafa` (canvas) to `hsl(224, 71.4%, 4.1%)` (near-black).
- **Primary** shifts from saturated blue to a lighter `hsl(213, 97%, 87%)` for adequate contrast on dark surfaces.
- **Muted / border** tokens compress into the same dark band (`hsl(215, 27.9%, 16.9%)`), distinguishable by opacity in use.
- **Signal colors** (success, warning) are not re-declared and fall through from light mode. Critical darkens to `hsl(0, 62.8%, 30.6%)`.

CSS custom properties handle the swap via `.dark` class on `<html>`. All Tailwind utilities consume tokens through `hsl(var(--*))`, so dark mode requires zero class changes in components.

> **Current state:** Dark mode tokens are defined in CSS but the product does not ship a dark mode toggle. The tokens exist for future readiness and for users who force `prefers-color-scheme: dark` at the OS level.

### Named Rules
**The One Voice Rule.** Action blue is the single saturated color with behavioral meaning. Any other use of hue is a signal color (critical / success / warning). If you are reaching for a second accent color for variety, you are wrong. The palette's restraint is not a limitation; it is the point.

**The Signal Purity Rule.** Status colors (critical, success, warning) exist only when communicating a state. A green badge on a healthy resource is correct. A green button for a non-destructive action is not.

## 3. Typography

**Display / Heading Font:** Space Grotesk (system-ui fallback, sans-serif)
**Body / UI Font:** Inter (system-ui fallback, sans-serif)
**Mono Font:** Fira Code (monospace)

**Character:** Space Grotesk brings a structured, slightly unconventional geometry that reads as precise without being cold. Inter at small sizes is pure instrument: invisible utility, no ego. Fira Code signals technical identity for any value that is simultaneously a display string and a machine-readable identifier.

### Hierarchy
- **Display** (Space Grotesk, 700, 1.5rem/24px, line-height 1.25, tracking -0.01em): Page-level headings. Used once per view for the primary landmark. The only weight-700 usage in the system; all other headings use 600.
- **Headline** (Space Grotesk, 600, 1.25rem/20px, line-height 1.3, tracking -0.005em): Section titles within a page, card titles, sheet headers. The `font-heading` utility class.
- **Title** (Space Grotesk, 600, 1rem/16px, line-height 1.4): Sub-section labels, table panel headers, form group labels.
- **Body** (Inter, 400, 0.875rem/14px, line-height 1.5): Table cell content, form field values, description text. Maximum line length 65-75ch on prose surfaces.
- **Label** (Inter, 500, 0.75rem/12px, line-height 1.4, tracking 0.01em): Column headers, metadata tags, badge text, compact UI labels. Never in all-caps.
- **Mono** (Fira Code, 400, 0.8125rem/13px, line-height 1.6): Technical identifiers: model names, field slugs, SQL expressions, IDs. Applied via `font-mono` utility.

### Weight Policy
The weight scale is intentionally narrow:
- **400 (normal):** Body text, form inputs, descriptions. The default.
- **500 (medium):** Reserved exclusively for technical identifiers rendered in `font-mono` (code, enum names, API identifiers). Never on Inter or Space Grotesk.
- **600 (semibold):** Headings (headline, title), section labels, table headers, badges. The primary emphasis weight.
- **700 (bold):** Page-level display headings only (`pageTitle` in `typography.ts`). One per view. Do not use for inline emphasis or sub-headings.

`font-extrabold` (800) and `font-black` (900) are banned entirely.

### Named Rules
**The Serif Exception.** The `.font-label` utility (Times New Roman) exists as a section label ornament -- a nod to the engineering-document register. Use it exclusively for static, decorative section labels, not for any interactive or data-bearing text.

**The Scale Minimum.** No text appears at less than 12px (0.75rem). Below this threshold, the label is not legible in standard office lighting; it is noise.

## 4. Elevation

This system is flat by default. Surfaces are differentiated primarily through background color contrast (canvas -> surface) and structural borders, not shadow depth. Shadows are reserved for floating or detached layers, not for resting state surface hierarchy.

**Scene context:** An operator at a desk in an office, 27-inch monitor, standard ambient light, mid-day. Shadows that work in a darkened showroom become noise in this environment. The response is a flat system that reads clearly in any ambient condition.

### Shadow Vocabulary
- **Ambient Lift** (`shadow-sm`: `0 1px 2px rgba(0,0,0,0.05)`): Subtle differentiation on interactive surfaces at rest. Used on default buttons. Functional minimum; not a visual statement.
- **Overlay** (`shadow-md`: `0 4px 6px -1px rgba(0,0,0,0.1)`): Popovers, select dropdowns, context menus. Signals detachment from the document flow.
- **Modal** (`shadow-lg`: `0 10px 15px -3px rgba(0,0,0,0.1)`): Dialogs, sheets, alert dialogs. The highest elevation layer.

### Named Rules
**The Flat-By-Default Rule.** Cards, table rows, sidebar items, nav entries -- all flat at rest. Shadows appear only when an element is elevated above the document flow (floating) or has lifted into a hover/active state that needs spatial separation. If you are adding a shadow to a resting card, you are decorating, not communicating.

## 5. Motion

Motion in ModelCraft is functional, not expressive. Every animation exists to orient the user spatially (where did this element come from?) or to smooth a state transition (accordion expand). No animation loops, no decorative motion, no attention-seeking entrance sequences.

### Animation Vocabulary
- **Fade In** (`animate-fade-in`, 250ms ease-out): Content appearing in place. Translates 4px upward while fading in. Used for page section reveals and lazy-loaded content.
- **Slide In Right** (`animate-slide-in-right`, 250ms ease-out): Side-panel content. Translates 10px from right while fading in. Used for sheet/drawer body content after the container has opened.
- **Accordion** (`animate-accordion-down` / `animate-accordion-up`, 200ms ease-out): Radix collapsible content. Height animates to/from `var(--radix-accordion-content-height)`.
- **Sheet Open/Close** (500ms / 300ms ease-in-out): Drawer slides from edge. Asymmetric timing: opening is slower (deliberate reveal), closing is faster (dismissal should feel instant).

### Named Rules
**The 300ms Ceiling.** No entrance animation exceeds 300ms. Sheet open (500ms) is the sole exception because it animates a large surface area and uses ease-in-out to prevent jarring snap. Everything else completes in 250ms or less.

**The No-Layout-Animation Rule.** Never animate `width`, `height`, `left`, `right`, or `top` on elements in the document flow. These trigger layout recalculation. Use `transform` and `opacity` exclusively. Accordion is the one exception (height animation on a collapsible, managed by Radix).

**The No-Loop Rule.** No CSS animation uses `infinite` iteration. Skeleton pulses are the only acceptable repeating pattern, and those use Tailwind's built-in `animate-pulse`.

## 6. Components

### Buttons
The button vocabulary is small and intentional. Each variant has exactly one behavioral role.

- **Shape:** Gently rounded (6px radius). Not pill-shaped; not sharp-cornered.
- **Primary** (`bg-primary text-white shadow-sm`, height 36px, padding 8px 16px): The single most consequential action on any given surface. Only one primary button per page section.
- **Hover / Focus:** Primary darkens to `#1d4ed8`; focus ring is 1px `ring-ring`. No scale transforms, no glow effects.
- **Outline** (`border border-input bg-background`, height 36px, padding 8px 16px): Secondary actions that need presence without dominance. Used for "edit", "add", cancel controls.
- **Ghost** (`hover:bg-muted`, height 36px, padding 8px 12px): Low-emphasis actions in dense contexts (table row actions, icon buttons). Background appears only on hover.
- **Destructive** (`bg-destructive text-white`): Reserved for confirmed irreversible actions inside AlertDialog confirmations. Never as the first action presented.
- **Size scale:** default (h-9), sm (h-8, px-3, text-xs), icon (36x36, padding 0).

### Badges / Status Chips
Square-rounded (4px) for labeling and status communication. Text at 11px (0.6875rem), font-weight 500. Height 20px. No pills (rounded-full banned).

- **Default** (indigo-surface bg, indigo text): Custom/user-defined labels.
- **Success** (rgba green 8%, green text): Active/healthy states.
- **Warning** (rgba amber 8%, amber text): Degraded/warning states.
- **Destructive** (rgba red 8%, red text): Error/banned states.
- **Neutral** (canvas bg, border, muted text): System/implicit labels, neutral classification.

### Cards / Containers
- **Corner Style:** Softly rounded (8px, `rounded-lg`).
- **Background:** Surface white (`#ffffff`), sitting on canvas (`#fafafa`).
- **Shadow:** `shadow-sm` at rest. Not elevated further.
- **Border:** `border border-border` (`#e2e6ec`). Present and structural; not decorative.
- **Internal Padding:** 24px all sides (`p-6`) for card panels. `p-4` for compact content areas inside settings.

### Tables
The primary data surface. Tables are flat with hairline row separators.

- **Header row:** Pure white background (`#FFFFFF`). 2px solid `#E3E8EE` bottom border. Text `#1A1F36`, 11px uppercase, font-weight 500, letter-spacing 0.06em. Height 38-40px.
- **Toolbar:** Background matches page canvas (`#F6F8FA`). Search input uses `#EBEEF2` background + `#D8DDE5` border — one step darker than toolbar for inset feel.
- **Body rows:** White background, 48px height (comfortable) or 40px (dense). Border-bottom 1px `#E3E8EE` on each row except last. Hover: `rgba(0,0,0,0.015)`.
- **Text alignment:** Primary text left, numbers right, status badges centered.
- **Action column:** Ghost icon button (⋮), visible on row hover.
- **Empty state:** Simple — gray icon + short text + primary CTA. No dashed cards, no colored circles.

### Inputs / Fields
- **Style:** Transparent background, single border (`border border-input`), 6px radius, height 36px.
- **Focus:** 1px `ring-ring` (action blue) inset glow. No animated border-width change.
- **Error:** `ring-destructive` focus ring; red helper text below.
- **Disabled:** `opacity-50 cursor-not-allowed`. Background unchanged.

### Navigation (App Sidebar)
- **Structure:** Left sidebar, surface-white (`#FFFFFF`) background, 1px `#E3E8EE` right border. Fixed width 240px expanded / 64px icon-only collapsed.
- **Nav items:** 36px height, 10px H padding, `rounded-md`. Icon (16px, stroke-width 1.5) + text label. 3px left border (transparent by default).
- **Default state:** Text `#697386`, icon `#8792A2`.
- **Hover:** `rgba(0,0,0,0.04)` background, text `#1A1F36`.
- **Active:** `rgba(79,70,229,0.08)` background, text `#4F46E5`, font-weight 500, left border `3px solid #4F46E5`.
- **Section headers:** 11px uppercase, font-weight 500, color `#8792A2`, letter-spacing 0.06em. Margin-top 16px between sections. Not clickable.
- **User area:** Bottom of sidebar, not in topbar.

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

## 7. Do's and Don'ts

### Do:
- **Do** use action indigo (`#4F46E5`) only on primary interactive elements: buttons, active nav left-stripe, tab underlines, focus rings. Its scarcity is what makes it legible as "this is the action."
- **Do** use shadow-first hierarchy: cards use `shadow-md`, floating layers use `shadow-lg`. Borders are structural (always 1px `#E3E8EE`), not the primary depth signal.
- **Do** use Inter for all text. Fira Code only for technical identifiers (model names, field slugs, IDs, SQL).
- **Do** show explicit states for every interactive surface: empty, loading, error, success. A component with no empty state is an unfinished component.
- **Do** use `ghost` and `outline` button variants for table row actions. Primary buttons belong to primary actions on a screen, not to every row in a table.
- **Do** keep text at or above 0.75rem (12px). Below this, the content is not communication; it is visual noise.
- **Do** vary spacing deliberately: 16px between form fields, 24px between sections, 8px inside compact nav items. Uniform padding everywhere produces monotony, not order.
- **Do** use semantic Tailwind tokens (`bg-primary`, `text-foreground`, `border-border`) instead of raw color values (`bg-blue-600`, `text-gray-700`). Semantic tokens are the only path to dark mode and theme consistency.
- **Do** use `transform` and `opacity` for animations. These are GPU-composited and do not trigger layout recalculation.

### Don't:
- **Don't** use Oracle/SAP-style thick panels, deeply layered card shadows, or heavy chrome.
- **Don't** use consumer-product softness — gradients, candy-colored badges, `rounded-full` pills, `hover:scale` zoom, or gamified copy.
- **Don't** use glassmorphism, `backdrop-filter`, gradient-filled cards, or animated gradient text.
- **Don't** use `bg-white/80` or any semi-transparent white backgrounds — always use solid `#FFFFFF`.
- **Don't** use a gray fill for table headers (thead). Use pure white background + 2px bottom border instead.
- **Don't** use the same background color for both the toolbar and the search input inside it — the search box must be one step darker to feel recessed.
- **Don't** place action indigo on backgrounds, headers, or decorative elements. Its appearance should predict interactivity or selection.
- **Don't** use `font-bold` (700) or heavier. Weight scale stops at `font-semibold` (600).
- **Don't** use Space Grotesk. Inter is the only permitted font family (Fira Code for mono).
- **Don't** use badge shape `rounded-full`. Badges use `rounded-sm` (4px).
- **Don't** use information density so low that users must scroll through three screens to find one configuration option. Appropriate density is a feature.
- **Don't** use hard-coded Tailwind color classes (`text-gray-600`, `bg-slate-200`). Always use semantic token equivalents.
- **Don't** use `transition-all`. Specify exact properties: `transition-colors`, `transition-shadow`, `transition-opacity`.
- **Don't** animate `width`, `height`, `left`, `right`, or `top` on document-flow elements. Use `transform: translate()` instead.
