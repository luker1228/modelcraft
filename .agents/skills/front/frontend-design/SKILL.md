---
name: frontend-design
description: Create distinctive, production-grade frontend interfaces with high design quality. Use this skill when the user asks to build web components, pages, artifacts, posters, or applications (examples include websites, landing pages, dashboards, React components, HTML/CSS layouts, or when styling/beautifying any web UI). Generates creative, polished code and UI design that avoids generic AI aesthetics.
license: Complete terms in LICENSE.txt
---

This skill guides creation of distinctive, production-grade frontend interfaces that avoid generic "AI slop" aesthetics. Implement real working code with exceptional attention to aesthetic details and creative choices.

## Project Style Reference

When working within this project, refer to the design system and styling guidelines before making any design decisions:

- Refer to @ai-metadata/front/style/quick-start.md for a quick overview of the design system.
- Refer to @ai-metadata/front/style/STYLE.md for the complete style guide.
- Refer to @ai-metadata/front/style/color-system.md for the color system and palette.
- Refer to @ai-metadata/front/style/tailwind-usage-policy.md for Tailwind CSS usage policy.
- Refer to @ai-metadata/front/style/design-system-demo-v2.html for a visual reference of the design system.

The user provides frontend requirements: a component, page, application, or interface to build. They may include context about the purpose, audience, or technical constraints.

## List Mutation Behavior

**After any mutation (create, update, delete), always prefer optimistic UI or cache update over a full page refetch.**

When a user deletes or creates an item in a list:
- **Remove/add it from the local list immediately** — do not wait for a round-trip `refetch()`
- Use Apollo `cache.modify` or `cache.evict` to update the list in-place
- Only fall back to `refetch()` when the mutation response doesn't provide enough data to reconstruct the list locally

This makes the list feel instant and responsive, which is critical to the user experience.

```tsx
// ✅ Good — remove from cache directly after delete
const [deleteEnum] = useMutation(DELETE_ENUM, {
  update(cache, _, { variables }) {
    cache.modify({
      fields: {
        enums(existing: Reference[], { readField }) {
          return existing.filter(ref => readField('name', ref) !== variables?.name)
        },
      },
    })
  },
})

// ❌ Avoid — full refetch causes visible flicker and extra network round-trip
const [deleteEnum] = useMutation(DELETE_ENUM, {
  onCompleted: () => refetch(),
})
```

---

## What NOT to Add Uninvited

**Do not add summary stats panels unless explicitly requested.**

A "summary stats panel" (统计盘) is a row of metric cards showing aggregated numbers (e.g., "Total: 3 enums", "Total usage: 80"). This pattern is often reflexively added to management pages but:

- It adds visual noise without user value in most CRUD pages
- It implies the user needs analytics they did not ask for
- It makes pages feel heavier and more complex than necessary

Only add stats/metrics when the user explicitly asks for them, or the page is clearly a **dashboard or analytics view** by nature.

**If in doubt — leave it out.**

---

## Design Thinking

Before coding, understand the context and commit to a BOLD aesthetic direction:
- **Purpose**: What problem does this interface solve? Who uses it?
- **Tone**: Pick an extreme: brutally minimal, maximalist chaos, retro-futuristic, organic/natural, luxury/refined, playful/toy-like, editorial/magazine, brutalist/raw, art deco/geometric, soft/pastel, industrial/utilitarian, etc. There are so many flavors to choose from. Use these for inspiration but design one that is true to the aesthetic direction.
- **Constraints**: Technical requirements (framework, performance, accessibility).
- **Differentiation**: What makes this UNFORGETTABLE? What's the one thing someone will remember?

**CRITICAL**: Choose a clear conceptual direction and execute it with precision. Bold maximalism and refined minimalism both work - the key is intentionality, not intensity.

Then implement working code (HTML/CSS/JS, React, Vue, etc.) that is:
- Production-grade and functional
- Visually striking and memorable
- Cohesive with a clear aesthetic point-of-view
- Meticulously refined in every detail

## Frontend Aesthetics Guidelines

Focus on:
- **Typography**: Choose fonts that are beautiful, unique, and interesting. Avoid generic fonts like Arial and Inter; opt instead for distinctive choices that elevate the frontend's aesthetics; unexpected, characterful font choices. Pair a distinctive display font with a refined body font.
- **Color & Theme**: Commit to a cohesive aesthetic. Use CSS variables for consistency. Dominant colors with sharp accents outperform timid, evenly-distributed palettes.
- **Motion**: Use animations for effects and micro-interactions. Prioritize CSS-only solutions for HTML. Use Motion library for React when available. Focus on high-impact moments: one well-orchestrated page load with staggered reveals (animation-delay) creates more delight than scattered micro-interactions. Use scroll-triggering and hover states that surprise.
- **Spatial Composition**: Unexpected layouts. Asymmetry. Overlap. Diagonal flow. Grid-breaking elements. Generous negative space OR controlled density.
- **Backgrounds & Visual Details**: Create atmosphere and depth rather than defaulting to solid colors. Add contextual effects and textures that match the overall aesthetic. Apply creative forms like gradient meshes, noise textures, geometric patterns, layered transparencies, dramatic shadows, decorative borders, custom cursors, and grain overlays.

NEVER use generic AI-generated aesthetics like overused font families (Inter, Roboto, Arial, system fonts), cliched color schemes (particularly purple gradients on white backgrounds), predictable layouts and component patterns, and cookie-cutter design that lacks context-specific character.

Interpret creatively and make unexpected choices that feel genuinely designed for the context. No design should be the same. Vary between light and dark themes, different fonts, different aesthetics. NEVER converge on common choices (Space Grotesk, for example) across generations.

**IMPORTANT**: Match implementation complexity to the aesthetic vision. Maximalist designs need elaborate code with extensive animations and effects. Minimalist or refined designs need restraint, precision, and careful attention to spacing, typography, and subtle details. Elegance comes from executing the vision well.

Remember: Claude is capable of extraordinary creative work. Don't hold back, show what can truly be created when thinking outside the box and committing fully to a distinctive vision.
