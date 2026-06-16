## 2024-05-18 - Missing ARIA labels in UI Components
**Learning:** Icon-only buttons lack ARIA labels, making them inaccessible to screen readers.
**Action:** Always add `aria-label` to icon-only buttons like the logout button in `Sidebar.tsx`. Focus-visible styles are missing on buttons and inputs across the app. Add `focus-visible:ring-2 focus-visible:ring-primary-fixed-dim focus-visible:outline-none` to `MechanicalButton.tsx` and text inputs like the one in `Chat.tsx`.
