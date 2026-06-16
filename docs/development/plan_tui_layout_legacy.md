# PLAN: Fix TUI Layout, Scrolling, and Overflow Issues

This plan addresses the critical bugs causing the dashboard to overflow, break layout, flicker, and scroll continuously instead of rendering as a proper full-screen application.

## User Review Required
> [!IMPORTANT]
> The issues you reported are caused by two main problems:
> 1. **Missing Alt Screen:** The application was rendering "inline" (printing to the standard terminal buffer) instead of using the alternate screen buffer, which full-screen apps like `k9s` use to prevent scrolling and screen corruption.
> 2. **Hardcoded Layout Heights:** The dashboard panels (e.g., `healthPanel` taking 15 lines, `sevPanel` taking 12 lines) were hardcoded. If your terminal window is shorter than ~32 lines, the layout overflows, causing Bubbletea to wrap lines and destroy the UI.

I will fix these by enabling the Alt Screen and making all layouts mathematically dynamic to fit exactly within your terminal's dimensions.

## Proposed Changes

### 1. `cmd/aitriage/scan.go`
Fix the application initialization so it locks into a full-screen view.
- **[MODIFY]** Change `tea.NewProgram(m)` to `tea.NewProgram(m, tea.WithAltScreen())`. This prevents scrolling and isolates the TUI from your terminal history.

### 2. `internal/ui/tui/view.go`
Refactor all sizing logic to be strictly dynamic relative to `m.Width` and `m.Height`, preventing overflow on any screen size.
- **[MODIFY]** `renderDashboard()`:
  - Calculate `availableHeight := m.Height - 6` (header, footer, padding).
  - Distribute `availableHeight` between the top chips and the bottom panels.
  - Make `activityPanel`, `healthPanel`, and `sevPanel` heights responsive instead of fixed 15/12/28.
- **[MODIFY]** `renderTriage()`:
  - Fix table height to exactly fit `m.Height - 8`.
  - Fix the right split-pane heights so `m.Viewport.Height` and `m.RemediationViewport.Height` add up precisely to `m.Height - 8`.
- **[MODIFY]** `renderChat()`:
  - Make `chatBox` take exactly `m.Height - 9` and `inputBox` take `3`, ensuring they don't exceed terminal height.
- **[MODIFY]** `renderAudit()` & `renderReports()`:
  - Set panel heights dynamically based on `m.Height`.

### 3. `internal/ui/tui/model.go` & `update.go`
- **[MODIFY]** Ensure `tea.WindowSizeMsg` updates the viewport dimensions correctly on resize.

## Verification Plan
1. **Automated Validation:** Build the binary (`go build ./cmd/aitriage`).
2. **Manual Verification:** Run `./aitriage scan . -i`. Verify that:
   - The TUI takes over the screen without any prior terminal text showing.
   - The UI does not scroll or flicker.
   - Resizing the terminal dynamically adjusts the tables and panels without breaking borders.
