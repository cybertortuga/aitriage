# AITriage Enterprise Security Dashboard — Redesign Plan
## Theme: Silent Luxury Cyber-Brutalist

> **Status**: 🟡 PHASES 1–5 COMPLETE · Phase 6 (Bug Fixes DONE · Visual Audit IN PROGRESS)  
> **Design Spec**: `docs/design_mockups/stitch_aitriage_enterprise_security_dashboard/silent_luxury_cyber_brutalist/DESIGN.md`  
> **Reference HTMLs**: `docs/design_mockups/stitch_aitriage_enterprise_security_dashboard/*/code.html`  
> **Dev Server**: `http://localhost:5176`  
> **Last Full Audit**: 2026-05-15

---

## Full Application Audit (May 2026)

### Active Component Render Tree
```
App.tsx
├── /login → LoginPage (components/LoginPage.tsx)                ✅ DONE
└── /      → Layout (components/Layout.tsx)                      ✅ DONE
    ├── Header (components/Header.tsx)                            ✅ DONE
    ├── Sidebar (components/Sidebar.tsx)                          ✅ DONE
    └── [Outlet — routed pages]
        ├── /cc            CommandCenterPage & CCOverviewPanel    ✅ DONE
        ├── /              DashboardPage                          ✅ DONE
        ├── /products      ProductsPage                           ✅ DONE
        ├── /products/:id  ProductDetailPage                      ✅ DONE
        ├── /kanban        KanbanPage                             ✅ DONE
        │   └── KanbanBoard                                       ✅ DONE
        │       ├── Column.tsx                                    ✅ DONE
        │       ├── FindingCard.tsx                               ✅ DONE
        │       └── FindingDetailModal ← Modal.tsx                ❌ VIOLATIONS
        ├── /findings      FindingsPage                           ✅ DONE
        ├── /chat          AIChatPage                             ✅ DONE
        ├── /rules         RulesPage                              ✅ DONE
        ├── /topology      TopologyPage                           ✅ DONE
        ├── /scanners      ScannersPage                           ✅ DONE
        ├── /terminal      TerminalPage                           ✅ DONE
        ├── /reports       ReportsPage                            ✅ DONE
        └── /admin         AdminPanelPage                         ✅ DONE
            ├── UsersTab                                          ❌ VIOLATIONS
            ├── AuditLogTab                                       ❌ VIOLATIONS
            ├── SystemConfigTab                                   ✅ OK
            └── LoadingScreen                                     ❌ VIOLATIONS
RouteError (ui/ErrorBoundary.tsx)                                 ❌ VIOLATIONS
```

### Orphaned Legacy Components (NOT in active routes — dead code)
These exist in the codebase but are **not imported by any active route or component**.  
Do NOT delete — they may be referenced by future features or branch code.  
Do NOT spend time fixing them.

| File | Size | Why Orphaned |
|---|---|---|
| `components/Dashboard.tsx` | 8.6KB | Pre-redesign dashboard, replaced by DashboardPage |
| `components/Chat.tsx` | 7.1KB | Pre-redesign chat, replaced by AIChatPage |
| `components/AdminPanel.tsx` | 10.4KB | Pre-redesign admin, replaced by AdminPanelPage |
| `components/SASTView.tsx` | 5.8KB | Old SAST viewer, not routed |
| `components/EngagementsView.tsx` | 6.2KB | Old engagements view, not routed |
| `components/Foundry.tsx` | 4.0KB | Old foundry page, not routed |
| `components/FoundryLoader.tsx` | 6.5KB | Old scan loader overlay, not routed |
| `components/SourceViewer.tsx` | 3.8KB | Old source viewer, not routed |
| `components/Dependencies.tsx` | 4.2KB | Old deps view, not routed |
| `components/DependencyGraph.tsx` | 5.6KB | Old dep graph, not routed |
| `ui/DataGrid.tsx` | 1.2KB | Only used by orphaned components above |
| `ui/Telemetry.tsx` | 2.1KB | Only used by orphaned components above |
| `ui/Brand.tsx` | 2.1KB | Only used by orphaned FoundryLoader |
| `components/common/Breadcrumbs.tsx` | 2.0KB | Not imported anywhere active |
| `components/common/LoadingScreen.tsx` | 0.5KB | Used by AdminPanelPage — needs fix |
| `components/common/NotificationPanel.tsx` | 4.0KB | Not imported anywhere active |

---

## Remaining Issues — Detailed Bug List

### 🔴 BUG-1: Undefined CSS Tokens `text-success` / `border-success` / `bg-success`
**Severity**: HIGH — Components render with no color, broken UI  
**Root Cause**: `--color-success` and `--color-warning` were never added to `index.css @theme`  
**Affected Files**:
- `components/admin/UsersTab.tsx` — `border-success/40`, `text-success`, `bg-success/5`
- `components/admin/AuditLogTab.tsx` — `text-success`
- `components/common/LoadingScreen.tsx` — referenced via AdminPanelPage  

**Fix**: Add `--color-success: #4caf50` and `--color-warning: #f59e0b` tokens to `index.css @theme` block.

---

### 🔴 BUG-2: Design Violations in `Modal.tsx` (affects FindingDetailModal)
**Severity**: HIGH — Violates core spec rules (no blur, no shadow)  
**File**: `components/common/Modal.tsx`  
**Violations**:
- `backdrop-blur-sm` on overlay — **SPEC: No blur effects allowed**
- `shadow-2xl` on modal container — **SPEC: No shadows allowed**
- `scrollbar-thin` — not defined in our CSS (should be `cyber-scrollbar` or default)
- Scale animation `scale: 0.95 → 1` — borderline; spec prefers instant cuts  

**Fix**: Remove `backdrop-blur-sm`, `shadow-2xl`. Replace `scrollbar-thin` with `cyber-scrollbar`. Keep backdrop opacity, remove blur. Simplify animation.

---

### 🔴 BUG-3: `rounded-full` Violations (sharp corners rule)
**Severity**: MEDIUM — Every `rounded-*` class violates the 0px corner spec  
**Affected Files & Lines**:
- `components/common/LoadingScreen.tsx:6` — `rounded-full animate-spin` loading spinner
- `components/findings/FindingDetailModal.tsx:312` — `rounded-full animate-spin` loading spinner
- `ui/ErrorBoundary.tsx:37` — `rounded-full animate-ping` error dot
- `ui/ErrorBoundary.tsx:75` — `rounded-full animate-ping` error dot  

**Fix**: Replace spinning `rounded-full` circles with square pulsing elements. Replace `animate-spin` spinner with a brutalist square blink/pulse.

---

### 🔴 BUG-4: `backdrop-blur-sm` in `ErrorBoundary.tsx`
**Severity**: MEDIUM — Violates no-blur spec  
**File**: `ui/ErrorBoundary.tsx:35, :73`  
**Fix**: Remove `backdrop-blur-sm`, set explicit background color instead.

---

### 🟡 BUG-5: `App.css` — Leftover Vite Boilerplate
**Severity**: LOW — Uses undefined CSS variables (`var(--accent)`, `var(--border)`, etc.), adds dead CSS weight  
**File**: `web/src/App.css`  
**Fix**: Clear the file contents (keep empty, do not delete — Vite references it).

---

### 🟡 BUG-6: `FindingDetailModal.tsx` — `border-warning` Undefined Token
**Severity**: LOW — Token `warning` doesn't exist in theme; falls back gracefully but incorrect  
**File**: `components/findings/FindingDetailModal.tsx` (EngagementsView uses it but it's orphaned)  
**Fix**: After BUG-1 fix adds `--color-warning`, this resolves automatically.

---

## Task Checklist

### Phase 1–5 — COMPLETE ✅
All pages, layout shell, and secondary pages have been rewritten per the Silent Luxury design system.

### Phase 6 — Bug Fixes & Visual Audit (COMPLETE)

#### 6.1 — CSS Token Additions ✅
- [x] Add `--color-success: #4caf50` to `index.css @theme`
- [x] Add `--color-warning: #f59e0b` to `index.css @theme`
- [x] Add `--color-on-success` and `--color-on-warning` companion tokens

#### 6.2 — Fix Modal.tsx (Level 3 elevation spec compliance) ✅
- [x] Remove `backdrop-blur-sm` from overlay div
- [x] Remove `shadow-2xl` from modal container
- [x] Replace `scrollbar-thin` with `cyber-scrollbar`
- [x] Set `bg-[#111111] border border-white` on modal container (Level 3 spec)
- [x] Simplify animation — opacity only, no scale

#### 6.3 — Fix rounded corners violations ✅
- [x] `LoadingScreen.tsx` — replaced with 3-bar brutalist pulsing indicator
- [x] `FindingDetailModal.tsx` — replaced with violet pulsing bars
- [x] `ErrorBoundary.tsx` — replaced `rounded-full animate-ping` with square `animate-pulse`

#### 6.4 — Fix ErrorBoundary.tsx ✅
- [x] Remove `backdrop-blur-sm` from both `ErrorBoundary` and `RouteError`
- [x] Set `bg-surface-container-low` on error card

#### 6.5 — Clean App.css ✅
- [x] Cleared all Vite boilerplate (file preserved, now empty)

#### 6.6 — Visual Audit Pass (all 12 routes)
Walk through every route in browser and note regressions:
- [x] `/login` — LoginPage
- [x] `/cc` (and `/`) — CommandCenterPage & CCOverviewPanel (Headers compacted, AI Summary and Workbench added)
- [x] `/findings` — Triage (master/detail) (Header compacted)
- [x] `/products` — Assets grid (Header compacted)
- [x] `/products/:id` — Asset detail (tabs) (Header compacted)
- [x] `/kanban` — Remediation Hub + FindingDetailModal (Header compacted)
- [x] `/rules` — Policies & Rules (Header compacted)
- [x] `/reports` — Reports & SBOM (Header compacted)
- [x] `/admin` — RBAC (all 4 tabs: Users, Audit Logs, System Config, API Keys) (Header compacted)
- [x] `/chat` — AI Copilot (Header compacted)
- [x] `/topology` — Architecture graph (Header compacted)
- [x] `/terminal` — Terminal feed
- [x] `/scanners` — Scanner cards (Header compacted)

#### 6.7 — Final Cleanup
- [ ] Verify no TypeScript errors in build output
- [ ] Verify no undefined class warnings in browser console
- [ ] Update this plan status to COMPLETE

---

## Design System Reference

### Core Rules (Non-Negotiable)
1. **0px border radius everywhere** — no `rounded-*` classes
2. **No shadows** — no `shadow-*` classes
3. **No blur** — no `backdrop-blur-*` or `blur-*` classes
4. **No gradients** — no `bg-gradient-*`
5. **Instant transitions only** — no `transition-all`, use `transition-none` or `duration-[50ms]`
6. **Status = solid squares** — 8×8px `status-dot` class, no circles

### Color Tokens (index.css @theme)
| Token | Value | Use |
|---|---|---|
| `--color-background` | `#131313` | Page canvas |
| `--color-surface-container-lowest` | `#0e0e0e` | Darkest panels, sidebar |
| `--color-surface-container-low` | `#1c1b1b` | Secondary panels |
| `--color-surface-container` | `#201f1f` | Default panels |
| `--color-surface-container-high` | `#2a2a2a` | Grid headers |
| `--color-surface-container-highest` | `#353534` | Active nav items |
| `--color-on-surface` | `#e5e2e1` | Primary text |
| `--color-on-surface-variant` | `#c4c7c8` | Secondary/muted text |
| `--color-outline-variant` | `#444748` | Subtle borders |
| `--color-outline` | `#8e9192` | Mid borders |
| `--color-primary` | `#ffffff` | White accent, active elements |
| `--color-error` | `#ffb4ab` | Critical/error |
| `--color-error-container` | `#93000a` | Error backgrounds |
| `--color-success` | `#4caf50` | ✅ NEEDS ADDING — healthy/ok states |
| `--color-warning` | `#f59e0b` | ✅ NEEDS ADDING — high severity |

### Elevation System
| Level | Use | Background | Border |
|---|---|---|---|
| 0 — Canvas | Page background | `#000000` | none |
| 1 — Panels | Cards, sidebar, main panels | `#111111` | `1px #333333` |
| 2 — In-panel | Widgets, table headers | `#1a1a1a` | `1px #444444` |
| 3 — Modals | Dialogs, popovers | `#111111` | `1px #ffffff` |

### Typography Scale
| Token | Font | Size | Weight | Letter-spacing |
|---|---|---|---|---|
| `display-lg` | Inter | 48px | 700 | -0.04em |
| `title-lg` | Inter | 22px | 700 | -0.02em |
| `headline-lg` | Inter | 32px | 600 | -0.02em |
| `headline-md` | Inter | 24px | 600 | — |
| `headline-sm` | Inter | 18px | 600 | — |
| `body-lg` | Inter | 16px | 400 | — |
| `body-sm` | Inter | 14px | 400 | — |
| `mono-metrics` | Geist | 20px | 600 | +0.05em |
| `mono-data` | Geist | 13px | 400 | — |
| `label-caps` | Geist | 11px | 700 | +0.1em · ALL CAPS |

### Utility Classes (index.css @layer components)
| Class | Purpose |
|---|---|
| `.cyber-panel` | Level 1 panel: `bg-#111 border border-#333` |
| `.cyber-widget` | Level 2 widget: `bg-#1a1a1a border border-#444` |
| `.cyber-modal` | Level 3 modal: `bg-#111 border border-#fff` |
| `.cyber-grid-header` | Data grid header row |
| `.cyber-grid-row` | Data grid body row |
| `.cyber-input` | Input field: `bg-#000 border-#333, focus:border-#fff` |
| `.cyber-scrollbar` | 2px thin scrollbar |
| `.status-dot` | 8×8px solid square indicator |
| `.btn-primary` | White fill button |
| `.btn-secondary` | Transparent + border button |
| `.btn-action` | Violet (#8b5cf6) AI action button |
| `.btn-ghost` | Text-only ghost button |
