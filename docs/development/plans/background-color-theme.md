# Background Color Theme Plan

## Goal

Add a user-facing background color selection to the web app, matching the existing accent palette flow, while keeping the current dark cyber-obsidian visual language intact.

## Research Notes

- MDN recommends CSS custom properties for shared values that need to be reused and changed in one place. This fits the current app because accent color already flows through `data-accent` and CSS variables.
- MDN `<input type="color">` supports live `input` and final `change` events, but the current product uses named palette buttons for accent selection. To stay consistent and avoid arbitrary low-contrast backgrounds, this change should ship as curated dark background palettes rather than an unrestricted color picker.
- WCAG contrast guidance requires at least 4.5:1 for normal text and 3:1 for large text. Because the app has many small mono labels, background palettes must stay dark and preserve existing `#f4f4f5` foreground contrast.

## Current Audit

- Early accent initialization: `web/index.html` reads `aitriage_accent` and sets `html[data-accent]`.
- Accent state/UI: `web/src/components/Header.tsx` owns the settings modal and palette buttons.
- Login accent state: `web/src/components/LoginPage.tsx` reads/writes the same accent key.
- Theme tokens: `web/src/index.css` defines Tailwind v4 `@theme` colors and runtime CSS variables.
- Main app shell: `web/src/components/Layout.tsx` uses `bg-background`, animated glow orbs, `grid-bg`, and `simple-bg-gradient`.
- Affected hardcoded background values:
  - Global CSS utilities: `luxury-glass`, `luxury-card`, `btn-secondary`, `cyber-input`, scrollbars, shimmer, markdown pre, simple background gradient.
  - Header settings modal and simple tab bar.
  - Login page shell, panel, input, footer picker.
  - Several deep page-specific hardcodes exist in large simple-mode pages. This pass will convert the shared shell/tokens first and touch local hardcodes only where they materially block the selected background from being visible.

## Scope And Estimate

- Size: medium frontend change.
- Risk: medium because the app uses Tailwind utilities plus direct CSS hex values.
- Expected files:
  - `web/index.html`
  - `web/src/index.css`
  - `web/src/components/Header.tsx`
  - `web/src/components/LoginPage.tsx`
  - `web/src/components/Layout.tsx`
  - `web/src/locales/en/components.json`
  - `web/src/locales/ru/components.json`
- Verification:
  - `npm run build` in `web`.
  - Browser check on settings modal, advanced shell, simple shell, and login shell.

## Implementation Plan

- [x] Audit current accent/theme architecture and background hardcodes.
- [x] Research CSS custom properties, color input behavior, and contrast constraints.
- [ ] Add `aitriage_background` early initialization in `web/index.html`.
- [ ] Add background CSS variables and `html[data-bg="..."]` palettes in `web/src/index.css`.
- [ ] Map Tailwind `@theme` background/surface colors to the new runtime variables.
- [ ] Update global shell/utilities to use background/surface variables instead of fixed obsidian hexes.
- [ ] Add `BACKGROUND_PALETTES`, state, persistence, and event sync in `Header.tsx`.
- [ ] Add the background palette block below the existing accent palette in Theme settings.
- [ ] Add EN/RU labels for background palette and palette names.
- [ ] Make login and main shell consume the background tokens.
- [ ] Run build and fix compile/type issues.
- [ ] Start/verify the local app visually.

## Design Decisions

- Keep accent and background independent: accent controls action/brand highlight; background controls base and surface depth.
- Use curated dark palettes to preserve contrast and avoid UI becoming unreadable.
- Set `data-bg` before React mounts to avoid initial flash.
- Reuse CSS variables so existing Tailwind classes such as `bg-background`, `bg-surface`, and `bg-surface-container-*` automatically pick up the selected background.

## Open Questions

- Whether to expose a fully custom color input later. Current answer: not in this pass, because arbitrary colors can break contrast and require a larger text/surface derivation system.

