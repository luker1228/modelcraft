# ModelCraft Design System

> ModelCraft Design System - Restrained B2B Style. Professional, clear, and functional.

---

## Table of Contents

1. [Color System](#1-color-system)
2. [Typography](#2-typography)
3. [Spacing System](#3-spacing-system)
4. [Border Radius](#4-border-radius)
5. [Shadows](#5-shadows)
6. [Components](#6-components)
7. [Forms](#7-forms)
8. [Tables](#8-tables)
9. [Alerts](#9-alerts)
10. [Icons](#10-icons)
11. [Design Principles](#11-design-principles)

---

## 1. Color System

### 1.1 Primary Brand Color

| Purpose | Hex | Usage |
|---------|-----|-------|
| Primary | `#2563eb` | Main buttons, links, active states |
| Primary Hover | `#1d4ed8` | Button hover, focus states |
| Primary Light | `#dbeafe` | Light backgrounds, badge backgrounds |

**Usage**: Solid color only. No gradients, no transparency effects.

```css
--primary: #2563eb;
--primary-hover: #1d4ed8;
--primary-light: #dbeafe;
```

### 1.2 Semantic Colors

| Purpose | Hex | Usage |
|---------|-----|-------|
| Success | `#059669` | Active status, success messages |
| Success Light | `#ecfdf5` | Success badge/alert backgrounds |
| Warning | `#d97706` | Draft status, warning alerts |
| Warning Light | `#fef3c7` | Warning badge/alert backgrounds |
| Destructive | `#ef4444` | Delete actions, errors |
| Destructive Light | `#fee2e2` | Error badge/alert backgrounds |

**Usage**: Semantic colors convey meaning. Use light variants for backgrounds, solid for text/icons.

```css
--success: #059669;
--success-light: #ecfdf5;
--warning: #d97706;
--warning-light: #fef3c7;
--destructive: #ef4444;
--destructive-light: #fee2e2;
```

### 1.3 Neutral Colors

| Purpose | Hex | Usage |
|---------|-----|-------|
| Text Primary | `#111827` | Main text, headings |
| Text Secondary | `#6b7280` | Secondary text, descriptions |
| Text Tertiary | `#9ca3af` | Timestamps, hints, disabled |
| Border | `#e5e7eb` | Card borders, dividers |
| Background Primary | `#fafafa` | Page background, section backgrounds |
| Background Secondary | `#ffffff` | Card backgrounds, containers |
| Selected | `#dadee5` | Row/item selected state background |

```css
--text-primary: #111827;
--text-secondary: #6b7280;
--text-tertiary: #9ca3af;
--border: #e5e7eb;
--bg-primary: #fafafa;
--bg-secondary: #ffffff;
--selected: #dadee5; /* hsl(215 20% 88%) */
```

### 1.4 Color Principles

- **No gradients** - Use solid colors only
- **No transparency effects** - Use solid backgrounds
- **No decorative glows** - Keep design clean and professional
- **Semantic meaning** - Colors communicate status and action
- **High contrast** - Ensure text is readable on all backgrounds

---

## 2. Typography

### 2.1 Font Family

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
```

Use system fonts for optimal performance and native appearance.

### 2.2 Font Hierarchy

| Element | Size | Weight | Usage |
|---------|------|--------|-------|
| Page Title | 32px | 600 | Main page headings (h1) |
| Section Heading | 24px | 600 | Section titles (h2) |
| Card Title | 16px | 600 | Card/modal headings (h3) |
| Subsection | 14px | 600 | Form labels, badge text |
| Body Text | 14px | 400 | Main content, descriptions |
| Small Text | 13px | 400 | Secondary descriptions |
| Tertiary Text | 12px | 400 | Timestamps, hints |

### 2.3 Line Height

- Default: `1.5` (line-height: 150%)
- Use consistent line height across all text elements

---

## 3. Spacing System

### 3.1 Base Unit

All spacing uses a **4px base unit**:

| Name | Value | Pixels | Usage |
|------|-------|--------|-------|
| xs | 0.25rem | 4px | Minimal gaps |
| sm | 0.5rem | 8px | Icon-to-text, tight spacing |
| md | 1rem | 16px | Default padding, gap between items |
| lg | 1.5rem | 24px | Section spacing, card padding |
| xl | 2rem | 32px | Large gaps, bottom margins |
| 2xl | 3rem | 48px | Page sections, major spacing |

### 3.2 Component Spacing

| Component | Padding | Gap |
|-----------|---------|-----|
| Button | 8px 16px | N/A |
| Button Small | 6px 12px | N/A |
| Button Large | 10px 24px | N/A |
| Card | 16px | N/A |
| Input | 8px 12px | N/A |
| Badge | 4px 12px | N/A |
| Alert | 12px 16px | 12px gap (content) |

---

## 4. Border Radius

| Value | Pixels | Usage |
|-------|--------|-------|
| 4px | 4px | Small elements, badges |
| 6px | 6px | Input fields |
| 8px | 8px | Cards, buttons, modals |
| 12px | 12px | Large components |

Keep border radius **subtle and consistent** for a restrained appearance.

---

## 5. Shadows

### 5.1 Shadow Levels

| Level | CSS | Usage |
|-------|-----|-------|
| None | `none` | Most elements |
| Subtle | `0 1px 3px rgba(0, 0, 0, 0.05)` | Card hover state |
| Default | `0 4px 6px rgba(0, 0, 0, 0.1)` | Modals, dropdowns |

**Principle**: Use shadows sparingly. Borders are preferred for defining element boundaries.

---

## 6. Components

### 6.1 Buttons

#### Primary Button

```html
<button class="btn-primary">
  <svg><!-- Icon --></svg>
  Create Project
</button>
```

**CSS**:
```css
.btn-primary {
  background: #2563eb;
  color: white;
  padding: 8px 16px;
  height: 36px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-primary:hover {
  background: #1d4ed8;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary svg {
  width: 16px;
  height: 16px;
  stroke: currentColor;
  stroke-width: 1.5;
}
```

**Sizes**:
- **Small**: `padding: 6px 12px; height: 32px; font-size: 12px;`
- **Default**: `padding: 8px 16px; height: 36px; font-size: 14px;`
- **Large**: `padding: 10px 24px; height: 40px; font-size: 16px;`

**Usage**:
- Call-to-action buttons
- Primary form submissions
- Main page actions

#### Secondary Button

```html
<button class="btn-secondary">Cancel</button>
```

**CSS**:
```css
.btn-secondary {
  background: #fafafa;
  color: #111827;
  border: 1px solid #e5e7eb;
  padding: 8px 16px;
  height: 36px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-secondary:hover {
  background: white;
  border-color: #d1d5db;
}
```

**Usage**:
- Alternative actions
- Dismissal actions (Cancel, Close)
- Secondary form buttons

#### Ghost Button

```html
<button class="btn-ghost">
  <svg><!-- Icon --></svg>
  View
</button>
```

**CSS**:
```css
.btn-ghost {
  background: transparent;
  color: #111827;
  border: none;
  padding: 8px 12px;
  height: 36px;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-ghost:hover {
  background: #fafafa;
}
```

**Usage**:
- Icon buttons
- Tertiary actions
- Menu items
- Overflow menus

#### Icon Button

```html
<button class="btn-icon">
  <svg><!-- Icon --></svg>
</button>
```

**CSS**:
```css
.btn-icon {
  width: 36px;
  padding: 0;
  height: 36px;
  border-radius: 6px;
}
```

**Usage**:
- Standalone icon actions
- Menu triggers
- Quick actions

#### Destructive Button

```html
<button class="btn-destructive">
  <svg><!-- Trash icon --></svg>
  Delete
</button>
```

**CSS**:
```css
.btn-destructive {
  background: #ef4444;
  color: white;
  padding: 8px 16px;
  height: 36px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-destructive:hover {
  background: #dc2626;
}
```

**Usage**:
- Delete actions
- Dangerous operations
- Irreversible actions

### 6.2 Cards

```html
<div class="card">
  <h3 class="card-title">Project Name</h3>
  <p class="card-description">Description text</p>
  <div style="display: flex; gap: 8px; margin-bottom: 12px;">
    <span class="badge badge-success">Active</span>
  </div>
  <div class="card-footer">Additional info</div>
</div>
```

**CSS**:
```css
.card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.2s ease;
}

.card:hover {
  border-color: #dbeafe;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
  margin-bottom: 8px;
}

.card-description {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 12px;
}

.card-footer {
  font-size: 12px;
  color: #9ca3af;
}
```

**Usage**:
- Project listings
- Data displays
- Summary cards
- Grid layouts

### 6.3 Badges

```html
<span class="badge badge-success">Active</span>
<span class="badge badge-warning">Draft</span>
<span class="badge badge-destructive">Error</span>
<span class="badge badge-primary">In Progress</span>
```

**CSS**:
```css
.badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  width: fit-content;
}

.badge-success {
  background: #ecfdf5;
  color: #059669;
}

.badge-warning {
  background: #fef3c7;
  color: #d97706;
}

.badge-destructive {
  background: #fee2e2;
  color: #ef4444;
}

.badge-primary {
  background: #dbeafe;
  color: #2563eb;
}
```

**Usage**:
- Status indicators (Active, Draft, Archived)
- Project states
- Role/permission labels
- Progress indicators

---

## 7. Forms

### 7.1 Input Fields

```html
<div class="input-group">
  <label class="label">Project Name</label>
  <input type="text" placeholder="Enter project name">
</div>
```

**CSS**:
```css
.input-group {
  margin-bottom: 16px;
}

.label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #111827;
  margin-bottom: 6px;
}

input, textarea {
  font-family: inherit;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 8px 12px;
  font-size: 14px;
  background: white;
  color: #111827;
  transition: all 0.2s ease;
  width: 100%;
}

input:focus, textarea:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

input::placeholder, textarea::placeholder {
  color: #9ca3af;
}
```

**Usage**:
- Text inputs
- Search fields
- Text areas
- Form fields

### 7.2 Form Principles

- **Clear labels** - Always label form fields
- **Blue focus ring** - Primary color focus indicator
- **Single column** - Stack inputs vertically by default
- **Consistent spacing** - 16px between fields
- **Clear feedback** - Show validation errors clearly

### 7.3 Form Layout

```html
<form>
  <div class="input-group">
    <label class="label">Field Label</label>
    <input type="text" placeholder="Placeholder text">
  </div>
  
  <div class="input-group">
    <label class="label">Description</label>
    <textarea placeholder="Placeholder text" rows="4"></textarea>
  </div>

  <div style="display: flex; gap: 12px;">
    <button type="submit" class="btn-primary">Create</button>
    <button type="button" class="btn-secondary">Cancel</button>
  </div>
</form>
```

---

## 8. Tables

```html
<table>
  <thead>
    <tr>
      <th>Column</th>
      <th>Column</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Data</td>
      <td>Data</td>
    </tr>
  </tbody>
</table>
```

**CSS**:
```css
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

thead {
  background: #fafafa;
  border-bottom: 1px solid #e5e7eb;
}

th {
  text-align: left;
  padding: 12px;
  font-weight: 600;
  color: #111827;
}

td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
  color: #6b7280;
}

tbody tr:hover {
  background: #fafafa;
}

tbody tr.selected {
  background: #dadee5; /* --selected */
}
```

**Selected State**:
- Use `#dadee5` (HSL 215 20% 88%) for selected rows — darker than hover, clearly distinguishable
- Do **not** use `rgba(37, 99, 235, 0.05)` — too light, insufficient contrast

**Usage**:
- Data listings
- Project tables
- Member tables
- Results tables

---

## 9. Alerts

### 9.1 Success Alert

```html
<div class="alert alert-success">
  <svg><!-- Check icon --></svg>
  <div>Project created successfully</div>
</div>
```

**CSS**:
```css
.alert {
  border-radius: 6px;
  padding: 12px 16px;
  margin-bottom: 16px;
  font-size: 13px;
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.alert svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
  margin-top: 2px;
  stroke: currentColor;
  stroke-width: 1.5;
}

.alert-success {
  background: #ecfdf5;
  color: #059669;
  border: 1px solid rgba(5, 150, 105, 0.2);
}

.alert-warning {
  background: #fef3c7;
  color: #d97706;
  border: 1px solid rgba(217, 119, 6, 0.2);
}

.alert-error {
  background: #fee2e2;
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.alert-info {
  background: #dbeafe;
  color: #2563eb;
  border: 1px solid rgba(37, 99, 235, 0.2);
}
```

**Alert Icons**:
- **Success**: Check mark (✓)
- **Warning**: Alert triangle (⚠)
- **Error**: X or error circle (✕)
- **Info**: Info circle (ⓘ)

**Usage**:
- Status messages
- Error notifications
- User feedback
- Confirmations

---

## 10. Icons

### 10.1 Icon Library

Use **Lucide React** for all icons. Keep icons **simple and minimal**.

```tsx
import { 
  Plus, 
  Edit, 
  Trash2, 
  Eye,
  ChevronDown,
  Settings,
  Users,
  MoreHorizontal
} from "lucide-react"
```

### 10.2 Icon Specifications

**Stroke Settings**:
- `stroke-width="1.5"` for all icons
- `stroke-linecap="round"`
- `stroke-linejoin="round"`

**Example**:
```jsx
<Plus 
  className="w-4 h-4" 
  strokeWidth={1.5}
/>
```

### 10.3 Icon Sizes

| Size | Pixels | Tailwind | Usage |
|------|--------|----------|-------|
| xs | 12px | `w-3 h-3` | Decorative |
| sm | 14px | `w-3.5 h-3.5` | Secondary |
| md | 16px | `w-4 h-4` | Button/nav |
| lg | 20px | `w-5 h-5` | Emphasis |
| xl | 24px | `w-6 h-6` | Large icons |

### 10.4 Icon Colors

```tsx
// Text Secondary (default)
<Plus className="text-gray-600" />

// Primary Brand Color
<Plus className="text-blue-600" />

// Success
<Check className="text-green-600" />

// Warning
<AlertCircle className="text-amber-600" />

// Destructive
<Trash2 className="text-red-500" />

// White (on colored backgrounds)
<Plus className="text-white" />
```

### 10.5 Icon + Text Pattern

```jsx
<Button>
  <Plus className="w-4 h-4" />
  <span>Create Project</span>
</Button>
```

**Gap**: `gap-8px` between icon and text

### 10.6 Icon Principles

- **Stroke only** - No filled icons
- **Consistent stroke width** - 1.5 for all icons
- **Explicit sizing** - Use CSS classes, never guess
- **Vertically centered** - Align with adjacent text
- **No emoji** - Always use SVG icons
- **Simple and minimal** - Avoid complex shapes

---

## 11. Design Principles

### 11.1 Restraint

- **Solid colors only** - No gradients, no decorative effects
- **Minimal shadows** - Use borders for definition
- **No animations** - Keep interactions snappy (0.2s transitions max)
- **Clear hierarchy** - Visual weight matches importance
- **No decorative elements** - Every element serves a function

### 11.2 Clarity

- **Clear labels** - Every control should have a label
- **Obvious states** - Active/hover/disabled states must be obvious
- **High contrast** - Text must be easily readable
- **Consistent spacing** - Predictable layout

### 11.3 Functionality

- **Purpose-driven** - Every element serves a function
- **Predictable behavior** - Standard interactions expected
- **No distractions** - Focus user attention on content
- **B2B appropriate** - Professional, business-focused aesthetic

### 11.4 Consistency

- **Unified button styles** - Primary/Secondary/Ghost/Destructive
- **Consistent spacing** - 4px base unit throughout
- **Unified colors** - Limited semantic color palette
- **Consistent borders** - 1px borders on all cards/inputs
- **Consistent icons** - All icons from Lucide React

### 11.5 Accessibility

- **Color contrast** - All text meets WCAG AA standards
- **Focus indicators** - Clear blue focus ring on all interactive elements
- **Semantic HTML** - Proper heading hierarchy, labels on inputs
- **Disabled states** - Clearly indicate disabled controls
- **Alt text** - Icons should be labeled with `aria-label`

---

## Appendix: Quick Reference

### Color Hex Values
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
--selected: #dadee5; /* hsl(215 20% 88%) - selected row/item background */
```

### Common Class Names
```css
.btn-primary      /* Primary action button */
.btn-secondary    /* Alternative action */
.btn-ghost        /* Tertiary/icon button */
.btn-destructive  /* Delete/dangerous action */
.btn-icon         /* Icon-only button */
.card             /* Content container */
.card-title       /* Card heading */
.card-description /* Card description */
.card-footer      /* Card metadata */
.badge            /* Status indicator */
.badge-success    /* Success status */
.badge-warning    /* Warning status */
.badge-destructive /* Error status */
.badge-primary    /* Primary status */
.alert            /* Alert message */
.alert-success    /* Success alert */
.alert-warning    /* Warning alert */
.alert-error      /* Error alert */
.alert-info       /* Info alert */
.label            /* Form label */
.input-group      /* Form field wrapper */
```

### Tailwind Classes (Reference)
```
/* Text */
text-primary (#111827) → text-gray-900
text-secondary (#6b7280) → text-gray-500
text-tertiary (#9ca3af) → text-gray-400

/* Background */
bg-primary (#fafafa) → bg-gray-50
bg-secondary (#ffffff) → bg-white

/* Border */
border (#e5e7eb) → border-gray-200

/* Rounded */
rounded-sm (4px) → rounded-sm
rounded-md (6px) → rounded-md
rounded-lg (8px) → rounded-lg

/* Spacing */
p-4 (16px) → p-4
gap-2 (8px) → gap-2
mb-6 (24px) → mb-6

/* Font */
font-semibold (600) → font-semibold
font-medium (500) → font-medium
font-normal (400) → font-normal

/* Icon */
w-4 h-4 (16px) → w-4 h-4
w-5 h-5 (20px) → w-5 h-5
stroke-2 → stroke-width="1.5"
```

### Implementation Checklist

When implementing new components:
- [ ] Use solid colors only (no gradients)
- [ ] Add 1px border for cards/containers
- [ ] Set `border-radius: 6px` for inputs, `8px` for cards
- [ ] Use `gap: 8px` for icon+text
- [ ] Add `transition: all 0.2s ease` for interactive elements
- [ ] Implement blue focus ring: `box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1)`
- [ ] Use Lucide React icons with `stroke-width={1.5}`
- [ ] Maintain 16px spacing between form fields
- [ ] Use `#dadee5` (`--selected`) for row/item selected state background
- [ ] Use semantic colors (green for success, orange for warning, red for destructive)
- [ ] Test color contrast with WCAG AA standards
