---
name: AITriage Design System
colors:
  surface: '#0d1515'
  surface-dim: '#0d1515'
  surface-bright: '#323b3b'
  surface-container-lowest: '#081010'
  surface-container-low: '#151d1d'
  surface-container: '#192121'
  surface-container-high: '#232b2c'
  surface-container-highest: '#2e3637'
  on-surface: '#dce4e4'
  on-surface-variant: '#b9caca'
  inverse-surface: '#dce4e4'
  inverse-on-surface: '#2a3232'
  outline: '#849495'
  outline-variant: '#3a494a'
  surface-tint: '#00dce5'
  primary: '#e9feff'
  on-primary: '#003739'
  primary-container: '#00f5ff'
  on-primary-container: '#006c71'
  inverse-primary: '#00696e'
  secondary: '#b8c3ff'
  on-secondary: '#002388'
  secondary-container: '#0043eb'
  on-secondary-container: '#c6ceff'
  tertiary: '#fff9f0'
  on-tertiary: '#3a3000'
  tertiary-container: '#ffdb3f'
  on-tertiary-container: '#736000'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#63f7ff'
  primary-fixed-dim: '#00dce5'
  on-primary-fixed: '#002021'
  on-primary-fixed-variant: '#004f53'
  secondary-fixed: '#dde1ff'
  secondary-fixed-dim: '#b8c3ff'
  on-secondary-fixed: '#001356'
  on-secondary-fixed-variant: '#0035be'
  tertiary-fixed: '#ffe16c'
  tertiary-fixed-dim: '#e7c427'
  on-tertiary-fixed: '#221b00'
  on-tertiary-fixed-variant: '#544600'
  background: '#0d1515'
  on-background: '#dce4e4'
  surface-variant: '#2e3637'
typography:
  display-lg:
    fontFamily: JetBrains Mono
    fontSize: 24px
    fontWeight: '700'
    lineHeight: 32px
    letterSpacing: -0.02em
  header-md:
    fontFamily: JetBrains Mono
    fontSize: 18px
    fontWeight: '600'
    lineHeight: 24px
    letterSpacing: 0.05em
  body-base:
    fontFamily: JetBrains Mono
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
  code-sm:
    fontFamily: JetBrains Mono
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
  label-xs:
    fontFamily: JetBrains Mono
    fontSize: 11px
    fontWeight: '500'
    lineHeight: 14px
spacing:
  cell-x: 1ch
  cell-y: 1em
  container-padding: 2rem
  grid-gutter: 1px
  section-gap: 1.5rem
---

## Brand & Style

This design system embodies a "Silent Luxury" aesthetic tailored for high-stakes technical environments. It merges the raw, functional efficiency of a Terminal User Interface (TUI) with the polished restraint of a premium data suite. 

The style is rooted in **Technical Minimalism** and **Cyber-Brutalism**. It prioritizes density and precision over decorative elements. The emotional response is one of calm authority, intentional focus, and surgical accuracy. Every character, line, and gap serves a structural purpose, evoking the feel of a sophisticated mainframe or a high-end aviation instrument panel.

## Colors

The palette is anchored in deep neutrals to minimize eye strain during extended analytical sessions. 

- **Foundation:** The background utilizes Deep Slate (#121212) for the base and Charcoal (#1a1a1a) for surface elevation.
- **Accents:** Cyan and Cobalt Blue are used sparingly for interactivity, focus states, and data visualization highlights.
- **Semantics:** Color is strictly reserved for meaning. Blood Red indicates immediate action, Amber signals caution, and Slate Gray denotes routine or low-priority background processes.
- **Typography:** Off-white is used for primary data and headers to ensure high legibility, while muted gray is reserved for secondary metadata and inactive UI chrome.

## Typography

This design system uses **JetBrains Mono** across all levels to maintain a rigid, monospaced rhythm. The typography is designed for data-dense environments where character alignment is critical for scanning logs and tables.

- **Headlines:** Use uppercase and tracking (letter-spacing) to create a distinct hierarchy without increasing font size excessively.
- **Body:** Standardized at 14px for optimal readability in terminal windows.
- **Labels:** Small, uppercase, and often muted to provide context without competing with the data.
- **Alignment:** Strictly left-aligned or justified to the grid; centered text should be avoided to maintain the technical feel.

## Layout & Spacing

The layout philosophy follows a **Modular Grid System** based on character units. It utilizes a fluid layout that snaps to the underlying monospaced grid.

- **Structure:** Content is organized into "Panels" or "Cells" defined by unicode box-drawing characters.
- **Lipgloss-style Padding:** Internal padding within panels should be generous (typically 1-2 character widths) to ensure that dense data remains legible and "luxurious."
- **Visual Gaps:** Deliberate whitespace is used to separate high-level functional areas, preventing the UI from feeling cluttered despite its technical nature.
- **Responsiveness:** On smaller viewports, panels stack vertically. The priority is always given to the "Primary Data Stream," with sidebars collapsing into hidden overlays or bottom-docked tabs.

## Elevation & Depth

Depth in this design system is achieved through **Tonal Layering** and **Structural Framing** rather than shadows or blurs.

- **Primary Surface:** The lowest level is the `#121212` background.
- **Raised Panels:** Active containers use the `#1a1a1a` surface to subtly lift them from the background.
- **Framing:** High-contrast unicode borders (┌ ─ ┐) provide the primary means of separation.
- **Focus States:** Interactivity is indicated by high-contrast border color shifts (e.g., changing a gray border to Cyan) or by inverting the text and background colors of a specific line or cell.

## Shapes

The design system employs a **Strict Angular** shape language. 

- **Corners:** No rounded corners are permitted. All buttons, containers, and indicators must feature 90-degree angles to maintain the CLI-inspired aesthetic.
- **Dividers:** Horizontal and vertical lines are created using single or double-line unicode box characters (─ or ═).
- **Selection Brackets:** Active states can be further emphasized using square brackets `[` `]` or arrow indicators `>` to point to selected data points.

## Components

- **Buttons:** Styled as high-contrast blocks. Default state is a thin Cyan border with Cyan text. Hover/Active state inverts the component (Cyan background, Black text).
- **Gauges:** Horizontal bar charts utilizing block characters (█ █ █ ░ ░) to represent percentage or capacity. 
- **Tables:** Dense grids with column headers separated by a single line (─). Active rows use a full-width background highlight in Cobalt Blue or Cyan with inverted text.
- **Status Indicators:** Small solid squares (■) or hollow squares (□) colored according to the semantic palette (Red/Amber/Gray).
- **Input Fields:** Represented by a prompt character `>` followed by an underscore `_` cursor. The container is a simple rectangle defined by unicode borders.
- **Metrics Chips:** Key-value pairs displayed as `[ KEY : VALUE ]` with the key in muted gray and the value in primary off-white.