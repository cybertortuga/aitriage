---
name: Silent Luxury Cyber-Brutalist
colors:
  surface: '#131313'
  surface-dim: '#131313'
  surface-bright: '#3a3939'
  surface-container-lowest: '#0e0e0e'
  surface-container-low: '#1c1b1b'
  surface-container: '#201f1f'
  surface-container-high: '#2a2a2a'
  surface-container-highest: '#353534'
  on-surface: '#e5e2e1'
  on-surface-variant: '#c4c7c8'
  inverse-surface: '#e5e2e1'
  inverse-on-surface: '#313030'
  outline: '#8e9192'
  outline-variant: '#444748'
  surface-tint: '#c6c6c7'
  primary: '#ffffff'
  on-primary: '#2f3131'
  primary-container: '#e2e2e2'
  on-primary-container: '#636565'
  inverse-primary: '#5d5f5f'
  secondary: '#d0bcff'
  on-secondary: '#3c0091'
  secondary-container: '#571bc1'
  on-secondary-container: '#c4abff'
  tertiary: '#ffffff'
  on-tertiary: '#690006'
  tertiary-container: '#ffdad6'
  on-tertiary-container: '#c41f21'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e2e2e2'
  primary-fixed-dim: '#c6c6c7'
  on-primary-fixed: '#1a1c1c'
  on-primary-fixed-variant: '#454747'
  secondary-fixed: '#e9ddff'
  secondary-fixed-dim: '#d0bcff'
  on-secondary-fixed: '#23005c'
  on-secondary-fixed-variant: '#5516be'
  tertiary-fixed: '#ffdad6'
  tertiary-fixed-dim: '#ffb4ac'
  on-tertiary-fixed: '#410002'
  on-tertiary-fixed-variant: '#93000d'
  background: '#131313'
  on-background: '#e5e2e1'
  surface-variant: '#353534'
typography:
  display-lg:
    fontFamily: Inter
    fontSize: 48px
    fontWeight: '700'
    lineHeight: '1.1'
    letterSpacing: -0.04em
  headline-lg:
    fontFamily: Inter
    fontSize: 32px
    fontWeight: '600'
    lineHeight: '1.2'
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Inter
    fontSize: 24px
    fontWeight: '600'
    lineHeight: '1.3'
  headline-sm:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '600'
    lineHeight: '1.4'
  body-lg:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.6'
  body-sm:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.5'
  mono-metrics:
    fontFamily: Geist
    fontSize: 20px
    fontWeight: '600'
    lineHeight: '1'
    letterSpacing: 0.05em
  mono-data:
    fontFamily: Geist
    fontSize: 13px
    fontWeight: '400'
    lineHeight: '1.5'
  label-caps:
    fontFamily: Geist
    fontSize: 11px
    fontWeight: '700'
    lineHeight: '1'
    letterSpacing: 0.1em
spacing:
  unit: 4px
  gutter: 16px
  margin-mobile: 16px
  margin-desktop: 32px
  panel-padding: 24px
  data-density-compact: 8px
  data-density-comfortable: 16px
---

## Brand & Style

This design system embodies "Silent Luxury Cyber-Brutalist"—a sophisticated, high-stakes aesthetic designed for elite cybersecurity operations. It rejects the softness of consumer tech in favor of raw, architectural precision. The "Silent Luxury" aspect is expressed through extreme restraint, generous negative space within a high-density framework, and a monochromatic foundation. The "Cyber-Brutalist" influence manifests in sharp 90-degree angles, 1px structural grids, and a total absence of decorative flourishes like shadows or gradients.

The target audience consists of senior security analysts and CISO-level executives who require immediate, unvarnished truth from their data. The UI evokes a sense of "technological authority" and "calm under pressure," treating every pixel as a vital piece of intelligence.

## Colors

The palette is rooted in a "Strict Dark" philosophy. The background is pure black (#000000), providing an infinite void that maximizes contrast for critical alerts. Surfaces are layered using deep charcoals to create a subtle sense of hierarchy without breaking the monolithic feel.

Functional colors are used sparingly and with high intentionality. **Crimson** is reserved exclusively for immediate threats, while **Violet** denotes AI-augmented insights and autonomous actions. All interactive elements use high-contrast white-on-black or accented borders to ensure zero ambiguity in high-pressure triage scenarios.

## Typography

Typography is a dual-system approach. **Inter** provides high legibility for narrative content, headings, and system navigation. Its neutral, geometric construction supports the brutalist aesthetic while maintaining a professional "corporate" tone.

**Geist** is the engine of the design system, used for all data-dense areas including IP addresses, timestamps, terminal feeds, and metrics. It ensures that characters are distinct (preventing confusion between '0' and 'O') and aligns perfectly to the 1px grid. 

For display headings, use tight letter-spacing to emphasize the "Luxury" feel. For labels and technical metadata, use all-caps with increased letter-spacing to create a "tactical" instrumentation look.

## Layout & Spacing

The layout follows a "Modular Grid" philosophy. Content is housed in rigid panels defined by 1px borders. 

- **Grid:** A 12-column system is used for desktop layouts, but the visual priority is on the "Container" or "Panel." 
- **Density:** High density is the default. Information should be packed tightly to allow analysts to see the maximum amount of data without scrolling.
- **Micro-spacing:** Built on a 4px baseline. All padding and margins must be multiples of 4 (e.g., 4, 8, 12, 16, 24, 32, 48, 64).
- **Responsive:** On mobile, panels stack vertically. Sidebars on desktop collapse into icon-only rails to preserve horizontal space for data tables.

## Elevation & Depth

This design system explicitly rejects shadows and blurs. Depth is achieved through **Z-index layering** and **Border Color Contrast**.

- **Level 0 (Base):** #000000. The canvas.
- **Level 1 (Panels):** #111111 with a #333333 border.
- **Level 2 (In-panel widgets):** #1A1A1A with a #444444 border.
- **Level 3 (Modals/Popovers):** #111111 with a #FFFFFF (1px) border to signify immediate focus.

To create "visual punch" without shadows, use **Solid Fills**. For example, a hovering state for a list item should change the background from transparent to #1A1A1A instantly, with no transition or only a very fast (50ms) linear cut.

## Shapes

The shape language is strictly **Sharp**. Every corner is 0px. This reinforces the "Cyber-Brutalist" narrative and mimics the appearance of terminal windows and vintage radar displays. 

Buttons, input fields, cards, and even progress bars must utilize hard 90-degree angles. This lack of "softness" signals that the product is a precise instrument, not a consumer toy.

## Components

### Buttons
- **Primary:** Solid #FFFFFF background, #000000 text. Sharp corners. No shadow.
- **Secondary:** Transparent background, 1px #333333 border, #FFFFFF text.
- **Action (AI):** Solid #8B5CF6 background, #FFFFFF text.
- **Ghost:** Transparent background, #A1A1AA text. Underline on hover.

### Data Grids
- Headers must use `label-caps` typography with a #1A1A1A background.
- Row dividers are 1px solid #333333.
- Cell text uses `mono-data`.
- Status indicators are small solid squares (8x8px) in the respective alert color.

### Terminal Feeds
- Background: #000000. 
- Text: `mono-data`. 
- Scrollbars: Thin 2px solid lines in #333333, no rounded ends.
- Time-stamps should be muted (#A1A1AA).

### Input Fields
- 1px solid #333333 border. 
- Background: #000000. 
- Active state: 1px solid #FFFFFF.
- Error state: 1px solid #E53935.

### Charts
- Brutalist style: No curved lines. Use stepped lines for time-series data.
- Bar charts: Solid color blocks with 0px radius.
- Grid lines: 1px dotted or dashed #333333.