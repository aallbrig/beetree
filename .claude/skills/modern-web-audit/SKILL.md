---
name: modern-web-audit
description: Audit beetreecraft.com against modern web product conventions (hamburger icon, tap targets, dark mode, social cards, loading states) and fix what feels dated. Use when improving how modern the website looks and behaves, and during website evaluations.
---

# Modern web audit

People judge a product in seconds by whether it behaves like the sites they already use. Check each convention on the Hugo site (`website/`) and the SPA (`web/`), at 390 px wide (phone) and 1280 px wide, in light and dark.

## Checklist

- **Navigation**: the mobile menu is an icon button (three bars that become an X), not a word; it has `aria-label`, `aria-expanded` and a 44 by 44 px target. Dropdowns open on hover and on keyboard focus. The current page is marked. A skip-to-content link exists.
- **Header and footer**: sticky header that does not cover anchors; the same footer, with the version, on every page.
- **Touch**: every control is at least 44 px; no hover-only features; no horizontal scroll at 320 px.
- **Theme**: follows `prefers-color-scheme` until the visitor chooses; the choice persists; no flash of the wrong theme on load.
- **Typography and spacing**: a readable measure (60 to 75 characters), a consistent type scale, generous spacing, one accent colour used with purpose.
- **Feedback**: loading states (skeletons or spinners) instead of blank screens; toasts for success and errors; disabled states on buttons while a request runs; empty states that say what to do next.
- **Forms**: labels tied to inputs, the right input types, inline validation, `autocomplete` hints, visible focus rings.
- **Sharing and discovery**: Open Graph and Twitter cards with an image, canonical URL, favicon set, `theme-color`, a web app manifest.
- **Performance**: images sized and lazy-loaded, no layout shift, fonts that do not block, first load under 2 seconds on a throttled phone.
- **Accessibility**: AA contrast, landmarks, one `h1` per page, alt text, focus order, `prefers-reduced-motion` honoured.
- **Polish**: no emojis in the UI, consistent button styles, 404 page that helps, smooth but short transitions (150 ms).

## Method

1. Load each page type (home, learn, docs, resources, builder, explore, forum, profile) at both widths and walk the list.
2. Write each miss as one line: page, convention, what you saw.
3. Fix the highest-impact miss that ships this iteration, with a gate: a `check_site.sh` assertion on the HTML (for example, the menu toggle has `aria-label` and no visible "Menu" text), or a unit test.
4. File the rest as `loop-idea` issues.
