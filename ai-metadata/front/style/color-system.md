# ModelCraft Color System Reference

> Based on STYLE.md — Restrained B2B style. Solid colors only. No gradients, no transparency.

---

## CSS Variables

```css
/* Primary */
--primary: #2563eb;
--primary-hover: #1d4ed8;
--primary-light: #dbeafe;

/* Semantic */
--success: #059669;
--success-light: #ecfdf5;
--warning: #d97706;
--warning-light: #fef3c7;
--destructive: #ef4444;
--destructive-light: #fee2e2;

/* Neutral */
--text-primary: #111827;
--text-secondary: #6b7280;
--text-tertiary: #9ca3af;
--border: #e5e7eb;
--bg-primary: #fafafa;
--bg-secondary: #ffffff;
--selected: #dadee5;   /* hsl(215 20% 88%) — selected row/item background */
```

---

## Color Usage Guide

### Text Colors

| Hex | Tailwind | Use For |
|-----|----------|---------|
| `#111827` | `text-gray-900` | Headings, card titles, primary labels |
| `#6b7280` | `text-gray-500` | Body text, descriptions, table cells |
| `#9ca3af` | `text-gray-400` | Timestamps, hints, disabled text, placeholders |
| `#2563eb` | `text-blue-600` | Links, active nav items |

### Background Colors

| Hex | Tailwind | Use For |
|-----|----------|---------|
| `#ffffff` | `bg-white` | Cards, modals, input backgrounds |
| `#fafafa` | `bg-gray-50` | Page background, table header |
| `#dbeafe` | `bg-blue-100` | Primary badge background |
| `#ecfdf5` | `bg-green-50` | Success badge/alert background |
| `#fef3c7` | `bg-amber-50` | Warning badge/alert background |
| `#fee2e2` | `bg-red-50` | Error badge/alert background |
| `#dadee5` | `bg-[#dadee5]` | **Selected row/item** (not `bg-blue-50` — too light) |

### Border Colors

| Hex | Tailwind | Use For |
|-----|----------|---------|
| `#e5e7eb` | `border-gray-200` | Default borders (cards, inputs, table rows) |
| `#d1d5db` | `border-gray-300` | Hover border on secondary buttons |
| `#2563eb` | `border-blue-600` | Input focus state |
| `#dbeafe` | `border-blue-100` | Card hover border |

---

## Button Color Specs

### Primary Button
```
background:       #2563eb
background-hover: #1d4ed8
text:             #ffffff
border:           none
height:           36px (h-9)
padding:          8px 16px (px-4 py-2)
border-radius:    6px (rounded-md)
font-weight:      500 (font-medium)
```

### Secondary Button
```
background:       #fafafa
background-hover: #ffffff
text:             #111827
border:           1px solid #e5e7eb
border-hover:     1px solid #d1d5db
height:           36px (h-9)
padding:          8px 16px (px-4 py-2)
border-radius:    6px (rounded-md)
font-weight:      500 (font-medium)
```

### Ghost Button
```
background:       transparent
background-hover: #fafafa
text:             #111827
border:           none
height:           36px (h-9)
padding:          8px 12px (px-3 py-2)
border-radius:    6px (rounded-md)
font-weight:      500 (font-medium)
```

### Destructive Button
```
background:       #ef4444
background-hover: #dc2626
text:             #ffffff
border:           none
height:           36px (h-9)
padding:          8px 16px (px-4 py-2)
border-radius:    6px (rounded-md)
font-weight:      500 (font-medium)
```

---

## Badge Color Specs

All badges: `padding: 4px 12px`, `border-radius: 4px`, `font-size: 12px`, `font-weight: 500`

| Variant | Background | Text |
|---------|-----------|------|
| Success | `#ecfdf5` | `#059669` |
| Warning | `#fef3c7` | `#d97706` |
| Error/Destructive | `#fee2e2` | `#ef4444` |
| Primary/Info | `#dbeafe` | `#2563eb` |

---

## Alert Color Specs

All alerts: `padding: 12px 16px`, `border-radius: 6px`, `gap: 12px` (icon + content), `font-size: 13px`

| Variant | Background | Text | Border |
|---------|-----------|------|--------|
| Success | `#ecfdf5` | `#059669` | `rgba(5,150,105,0.2)` |
| Warning | `#fef3c7` | `#d97706` | `rgba(217,119,6,0.2)` |
| Error | `#fee2e2` | `#ef4444` | `rgba(239,68,68,0.2)` |
| Info | `#dbeafe` | `#2563eb` | `rgba(37,99,235,0.2)` |

---

## Card Specs

```
background:       #ffffff
border:           1px solid #e5e7eb
border-hover:     1px solid #dbeafe    (blue-100)
border-radius:    8px (rounded-lg)
padding:          16px (p-4)
shadow-hover:     0 1px 3px rgba(0,0,0,0.05)
transition:       all 0.2s ease
```

---

## Input / Form Specs

```
background:       #ffffff
border:           1px solid #e5e7eb
border-focus:     1px solid #2563eb
border-radius:    6px (rounded-md)
padding:          8px 12px (px-3 py-2)
font-size:        14px (text-sm)
focus-ring:       0 0 0 3px rgba(37,99,235,0.1)
placeholder:      #9ca3af (text-gray-400)
```

---

## Table Specs

```
font-size:        14px (text-sm)
border-collapse:  collapse

thead:
  background:     #fafafa (bg-gray-50)
  border-bottom:  1px solid #e5e7eb

th:
  padding:        12px (p-3)
  font-weight:    600 (font-semibold)
  text-color:     #111827

td:
  padding:        12px (p-3)
  border-bottom:  1px solid #e5e7eb
  text-color:     #6b7280

row-hover:        background #fafafa
row-selected:     background #dadee5  ← USE THIS, not rgba blue
```

---

## Shadow Levels

| Level | Value | Use For |
|-------|-------|---------|
| None | `none` | Most elements |
| Subtle | `0 1px 3px rgba(0,0,0,0.05)` | Card hover |
| Default | `0 4px 6px rgba(0,0,0,0.1)` | Modals, dropdowns |

**Principle**: Prefer borders over shadows. Use shadows only for floating elements (modals, dropdowns).

---

## Border Radius

| Value | Tailwind | Use For |
|-------|----------|---------|
| 4px | `rounded` or `rounded-sm` | Badges |
| 6px | `rounded-md` | Inputs, buttons |
| 8px | `rounded-lg` | Cards, modals |
| 12px | `rounded-xl` | Large containers |

---

## Spacing Reference

Base unit: **4px**

| Tailwind | Pixels | Common Use |
|---------|--------|------------|
| `gap-2` | 8px | Icon-to-text gap in buttons |
| `p-3` | 12px | Alert padding, badge content |
| `p-4` | 16px | Card padding, form field spacing |
| `gap-3` | 12px | Button group gap |
| `mb-4` / `space-y-4` | 16px | Form field vertical spacing |
| `p-6` / `gap-6` | 24px | Section padding |
| `p-8` / `gap-8` | 32px | Large section gaps |

---

## ❌ What NOT to Do

```tsx
// ❌ No gradients on buttons
<Button className="bg-gradient-to-r from-blue-600 to-indigo-600">Create</Button>

// ❌ No glass-morphism
<Card className="bg-white/70 backdrop-blur-sm">...</Card>

// ❌ No decorative blobs
<div className="absolute bg-gradient-to-br from-blue-400/20 to-cyan-400/20 rounded-full blur-3xl" />

// ❌ No glow effects
<div className="absolute inset-0 bg-blue-600 rounded-2xl blur-lg opacity-50" />

// ❌ Wrong selected row color (too light)
<tr className="bg-blue-50"> or <tr className="bg-[rgba(37,99,235,0.05)]">

// ✅ Correct
<tr className="bg-[#dadee5]">
```

---

## Typography Quick Reference

Font stack: `-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`

```tsx
// Page title
<h1 className="text-3xl font-semibold text-gray-900">Title</h1>

// Section heading  
<h2 className="text-2xl font-semibold text-gray-900">Section</h2>

// Card title
<h3 className="text-base font-semibold text-gray-900">Card</h3>

// Body text
<p className="text-sm text-gray-500">Description</p>

// Caption / hint
<span className="text-xs text-gray-400">Timestamp</span>

// Form label
<label className="block text-sm font-medium text-gray-900 mb-1.5">Label</label>
```
