# Task: Full Web Frontend Functionality Audit, Bug Fix & TUI Feature Parity

## Mission

The AITriage web dashboard currently has a beautiful "Silent Luxury / Cyber-Brutalist" UI, but most features are either broken, disconnected from the backend API, or use mock data instead of real API calls. Your job is to:

1. **Audit every component** — find all bugs, broken API calls, type mismatches, and dead code
2. **Connect every feature to the real backend API** — replace all mock data with actual API calls
3. **Achieve TUI feature parity** — every feature available in the Go TUI must work in the web UI
4. **Fix all TypeScript compilation errors** — `npm run build` must succeed with ZERO errors

**Do NOT change any CSS classes, design tokens, or visual styling.** The UI design is final.

---

## Architecture Overview

### Stack
- **Backend**: Go HTTP server at `internal/server/server.go` — listens on port 8080
- **Frontend**: React 19 + TypeScript 6 + Vite 8 + Tailwind CSS v4
- **Deployment**: Docker Compose — frontend (nginx on :80 → :3000 host), backend (:8080)
- **Proxy**: In dev, Vite proxies `/api/*` to `http://localhost:8080`. In Docker, nginx proxies to `http://backend:8080`

### TypeScript Config Constraints
- `verbatimModuleSyntax: true` — type-only imports MUST use `import type { ... }`
- `noUnusedLocals: true` — no unused variables
- `noUnusedParameters: true` — no unused function parameters
- `erasableSyntaxOnly: true`
- **Build command**: `npm run build` runs `tsc -b && vite build`

---

## Part 1: Complete Backend API Reference

These are ALL the endpoints in `internal/server/server.go`. Every single one must have a working frontend integration.

### `POST /api/login`
**Request:** `{ "username": string, "password": string }`
**Response:** `{ "ok": true, "username": string, "is_admin": boolean }`
**Side effect:** Sets `HttpOnly` cookie named `token` (JWT, 24h expiry)
**Auth:** None required (public endpoint)
**Default credentials:** username=`admin`, password=`admin` (created automatically on first run)

### `GET /api/me`
**Response (authenticated):** `{ "ok": true, "username": string, "is_admin": boolean }`
**Response (unauthenticated):** 401 status
**Auth:** Cookie-based JWT

### `POST /api/scan`
**Request:** `{ "path": string, "stack"?: string, "external"?: boolean }`
**Response:**
```json
{
  "ok": true,
  "scan_id": "SCAN-1715000000",
  "findings": [
    {
      "id": "ENTR-001",
      "name": "Hardcoded Secret Detected",
      "severity": "critical",
      "file": "config.py",
      "line": 42,
      "suggestion": "Move to environment variable",
      "owasp": "A02:2021",
      "audit_status": "open"
    }
  ],
  "dependencies": [
    {
      "name": "express",
      "version": "4.18.2",
      "type": "direct",
      "ecosystem": "npm"
    }
  ],
  "stacks": ["node", "python"],
  "security_score": 72,
  "security_grade": "C",
  "duration": "1.234s"
}
```
**Auth:** JWT required
**IMPORTANT:** This is a synchronous, potentially long-running call (5-30 seconds). The frontend must show loading progress during this time.

### `GET /api/browser?path=<path>`
**Response:**
```json
{
  "ok": true,
  "path": "/some/path",
  "entries": [
    { "name": "src", "is_dir": true, "path": "/some/path/src" },
    { "name": "main.go", "is_dir": false, "path": "/some/path/main.go" }
  ]
}
```
**Error responses:** 403 (permission denied), 404 (not found), 500 (internal error)
**Auth:** JWT required

### `POST /api/triage`
**Request:** `{ "project": string, "id": string, "file": string, "action": "IGNORE" | "FIX" | "OPEN" }`
**Response:** `{ "ok": true }`
**Auth:** JWT required
**What it does:** Writes audit status to `.aitriage/audit.json` in the project directory

### `POST /api/chat`
**Request:** `{ "messages": [{ "role": "user"|"system"|"assistant", "content": string }] }`
**Response:** `{ "ok": true, "content": string }` or `{ "ok": false, "error": "AI Consultant is offline..." }`
**Auth:** JWT required
**IMPORTANT:** This uses the LLM client (Gemini). It may return `ok: false` if no API key is configured. The frontend must handle this gracefully.

### `POST /api/analyze`
**Request:** `{ "id": string, "type": string }`
**Response:** `{ "ok": true, "analysis": string }` or `{ "ok": false, "error": string }`
**Auth:** JWT required
**What it does:** Sends the finding ID to the LLM for deep analysis

### `GET /api/file?path=<path>`
**Response:** `{ "ok": true, "content": "file contents as string" }`
**Auth:** JWT required

### `GET /api/health`
**Response:** `{ "ok": true, "tools": { "semgrep": bool, "bandit": bool, "gitleaks": bool, "trivy": bool } }`
**Auth:** JWT required

### `GET /api/admin/users` (admin only)
**Response:** `{ "users": [{ "username": string, "is_admin": boolean }] }`

### `POST /api/admin/users` (admin only)
**Request:** `{ "username": string, "password": string, "is_admin": boolean }`
**Response:** `{ "ok": true }`

### `DELETE /api/admin/users?username=<name>` (admin only)
**Response:** `{ "ok": true }`
**Constraint:** Cannot delete `admin` user

---

## Part 2: Current Bugs to Fix

### BUG 1: LoginPage.tsx — Not calling real API
**File:** `web/src/components/LoginPage.tsx`
**Problem:** The `handleLogin` function (line ~22) does NOT call `/api/login`. It just checks `if (username === 'admin')` client-side. This means:
- No JWT token is set
- All subsequent API calls will fail with 401
- Password field is completely ignored

**Fix:** Replace the client-side check with a real `POST /api/login` call. On success, the backend sets the JWT cookie automatically via `Set-Cookie`. Then call `onLogin()` with the response data.

```typescript
const handleLogin = async (e: React.FormEvent) => {
  e.preventDefault();
  setIsScanning(true);
  setError(null);
  
  try {
    const resp = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    const data = await resp.json();
    
    if (data.ok) {
      onLogin({ username: data.username, is_admin: data.is_admin });
    } else {
      setError(data.error || 'INVALID_CREDENTIALS');
    }
  } catch (err) {
    setError('AUTH_SERVER_UNREACHABLE');
  } finally {
    setIsScanning(false);
  }
};
```

### BUG 2: App.tsx — scan report data mapping mismatch
**File:** `web/src/App.tsx`
**Problem:** The Dashboard receives `report?.files_count` and `report?.timestamp` but the actual scan API response does NOT have those fields. The real response has:
- `security_score` ✓
- `findings` ✓
- `dependencies` ✓
- `stacks` ✓
- `scan_id` ✓
- `duration` ✓
- NO `files_count` — must be computed or removed
- NO `timestamp` — must use current time or `duration`
- NO `path` — must be stored from the scan request

**Fix:** Store the scanned path in `useScan` state. Compute `files_count` from browser or just show total findings. Use `new Date().toISOString()` for timestamp.

### BUG 3: useScan.ts — Loading animation doesn't progress
**File:** `web/src/hooks/useScan.ts`
**Problem:** The scan creates 5 loading steps but NEVER updates their status during the scan. The `progress` stays at 0 until it jumps to 100 on completion. This makes the FoundryLoader animation completely static.

**Fix:** Use a timer to simulate step progression while waiting for the API response:
```typescript
// Start a progress simulation interval
let stepIndex = 0;
const interval = setInterval(() => {
  if (stepIndex < 5) {
    setState(prev => ({
      ...prev,
      progress: ((stepIndex + 1) / 5) * 80, // Max 80% until real completion
      currentStep: prev.steps[stepIndex]?.name || '',
      steps: prev.steps.map((s, i) => ({
        ...s,
        status: i < stepIndex ? 'done' : i === stepIndex ? 'active' : 'pending'
      }))
    }));
    stepIndex++;
  }
}, 1500);

try {
  const resp = await fetch(/* ... */);
  // ... handle response
} finally {
  clearInterval(interval);
}
```

### BUG 4: Chat.tsx — Uses mock responses instead of real API
**File:** `web/src/components/Chat.tsx`
**Problem:** The `handleSend` function (line ~18) uses `setTimeout` with a hardcoded mock response. It does NOT call `/api/chat`.

**Fix:** Use the `chatWithAI` function from `useScan` hook, or call `/api/chat` directly with proper message format:
```typescript
const handleSend = async () => {
  if (!input.trim()) return;
  const userMsg: Message = { role: 'user', content: input };
  setMessages(prev => [...prev, userMsg]);
  setInput('');
  setLoading(true);
  
  try {
    const resp = await fetch('/api/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        messages: [...messages, userMsg].map(m => ({ role: m.role, content: m.content }))
      }),
    });
    const data = await resp.json();
    
    if (data.ok) {
      setMessages(prev => [...prev, { role: 'assistant', content: data.content }]);
    } else {
      setMessages(prev => [...prev, { 
        role: 'assistant', 
        content: `SYSTEM_ERROR: ${data.error || 'Neural core unreachable'}`
      }]);
    }
  } catch (err) {
    setMessages(prev => [...prev, { 
      role: 'assistant', 
      content: 'CRITICAL_FAILURE: Network connection to AI core severed.'
    }]);
  } finally {
    setLoading(false);
  }
};
```

### BUG 5: Triage.tsx — EXECUTE_SMART_FIX and MARK_FALSE_POSITIVE buttons do nothing
**File:** `web/src/components/Triage.tsx`
**Problem:** Both `MechanicalButton` elements on lines ~111 and ~115 have no `onClick` handlers. They are completely decorative.

**Fix:** Wire them to the `/api/triage` endpoint:
- "EXECUTE_SMART_FIX" → call triage with `action: "FIX"`
- "MARK_FALSE_POSITIVE" → call triage with `action: "IGNORE"`

The Triage component needs to receive `onTriage` and `onAnalyze` callback props (or directly call the API). After triage, update the finding's `audit_status` in the parent state.

### BUG 6: Triage.tsx — Missing AI Analysis feature
**Problem:** The TUI has a feature where you can press a key to get AI-powered analysis of a finding via `/api/analyze`. The web Triage view has no equivalent — there's no "Analyze" button and no place to display AI analysis results.

**Fix:** Add an "[ AI_DEEP_ANALYSIS ]" button in the detail sidebar. When clicked, call `/api/analyze` with the finding ID. Display the result in a new expandable section below the INCIDENT_SUMMARY.

### BUG 7: AdminPanel.tsx — Not connected to any API
**File:** `web/src/components/AdminPanel.tsx`
**Problem:** The component receives `users`, `onUpdateRole`, `onUpdateStatus` as props, but in `App.tsx` these are hardcoded as `users={[]}`, `onUpdateRole={() => {}}`, `onUpdateStatus={() => {}}`. No real API calls.

**Fix:**
1. On mount, fetch users from `GET /api/admin/users`
2. Add "Create User" form that calls `POST /api/admin/users`
3. Add "Delete" button per user that calls `DELETE /api/admin/users?username=X`
4. Refresh user list after each mutation

### BUG 8: Dashboard.tsx — Telemetry sidebar shows hardcoded status
**Problem:** The right sidebar shows "API_UPLINK: ESTABLISHED", "DB_SYNC: NOMINAL", "AI_CORE: READY" — these are all hardcoded strings. They should reflect actual system health.

**Fix:** On Dashboard mount, call `GET /api/health` and display real tool availability status. Show which external tools (semgrep, bandit, gitleaks, trivy) are installed.

### BUG 9: Browser.tsx — Entry click doesn't use `path` field
**Problem:** The backend returns `{ name, is_dir, path }` but the Browser component reconstructs the path manually (`currentPath + '/' + entry.name`). This is fragile and may break with host prefix paths.

**Fix:** Use `entry.path` directly from the API response. Update the `Entry` interface to include `path: string`.

### BUG 10: services/api.ts — Exists but is NOT USED
**File:** `web/src/services/api.ts`
**Problem:** There's a complete API service layer with typed functions (`api.getBrowser()`, `api.startScan()`, `api.getFile()`, etc.) but NO component uses it. All components make raw `fetch()` calls with no error handling.

**Fix:** Either:
- (Preferred) Refactor ALL components to use `api.*` functions from `services/api.ts`
- OR delete `services/api.ts` and keep raw fetch calls with proper typing

The `api.ts` file also has outdated types — `sendChat` sends `{ message, context }` but the backend expects `{ messages: Message[] }`. Fix the mismatch.

### BUG 11: types.ts — ScanStatusValue type is not used correctly
**Problem:** `types.ts` defines `ScanStatusValue` but `useScan.ts` defines its own `ScanStatus` type. There are two competing type systems.

**Fix:** Consolidate. Use ONE source of truth for all types. Move scan-related types into `types.ts`. Have `useScan.ts` import them.

### BUG 12: SourceViewer.tsx — File viewer component exists but is never rendered
**File:** `web/src/components/SourceViewer.tsx`
**Problem:** This component exists but is not mounted anywhere in `App.tsx`. The `/api/file` endpoint exists but has no frontend consumer.

**Fix:** When a user clicks on a file path in the Triage view or Browser view, show the SourceViewer with the file contents loaded from `/api/file?path=X`. This could be a modal or a slide-over panel.

---

## Part 3: Missing Features (TUI → Web Parity)

These features exist in the TUI (`internal/ui/tui/`) but are missing from the web:

### FEATURE 1: Scan Progress with Step Animation
The TUI shows real-time progress as each scan phase completes. The web must simulate this during the synchronous `/api/scan` call using timed step transitions in `useScan.ts`.

### FEATURE 2: Finding Detail with Source Code Preview
In the TUI, selecting a finding shows the source code snippet with the vulnerable line highlighted. The web must:
1. When a finding is selected in Triage, call `GET /api/file?path=<finding.file>`
2. Display the source code with syntax highlighting around `finding.line`
3. Show it in the TelemetrySidebar or a dedicated panel

### FEATURE 3: AI-Powered Finding Analysis
The TUI allows analyzing any finding with AI. Add an "[ AI_ANALYZE ]" button in the Triage detail panel that calls `POST /api/analyze`.

### FEATURE 4: Scan History / Re-scan
After scanning once, the user should be able to re-scan the same project from Dashboard without going back to Browser. Add a "[ RE-SCAN ]" button that uses the stored project path.

### FEATURE 5: Health Dashboard
Call `GET /api/health` on Dashboard load and show which external security tools are available (semgrep, bandit, gitleaks, trivy) with green/red indicators.

### FEATURE 6: Dependency Tree in Detail
The scan response includes `dependencies[]` with `name`, `version`, `type`, `ecosystem`. The Dependencies view already renders this but it's never populated because `report?.dependencies` comes from the scan response — ensure the data flows correctly.

---

## Part 4: Data Flow Architecture

Here's how data SHOULD flow (fix any deviations):

```
LoginPage → POST /api/login → cookie set → App renders main layout
                                          ↓
App.tsx (useScan hook holds global scan state)
  │
  ├─ Browser → GET /api/browser → user navigates → clicks "INITIALIZE_AUDIT"
  │                                                  ↓
  │                                          useScan.startScan(path)
  │                                                  ↓
  │                                          POST /api/scan → response stored in useScan.report
  │                                                  ↓
  ├─ Dashboard ← reads useScan.report (security_score, findings count, stacks)
  │            ← GET /api/health (tool availability)
  │
  ├─ Triage ← reads useScan.report.findings
  │         → POST /api/triage (triage action) → updates local finding status
  │         → POST /api/analyze (AI analysis) → shows analysis text
  │         → GET /api/file (source preview)
  │
  ├─ Dependencies ← reads useScan.report.dependencies
  │
  ├─ Chat → POST /api/chat (real LLM conversation)
  │
  └─ AdminPanel → GET/POST/DELETE /api/admin/users
```

---

## Part 5: Detailed Fixes for Each Component

### `web/src/hooks/useScan.ts`
1. Store the scanned `path` in state (for triage and re-scan)
2. Store `stacks`, `security_grade`, `scan_id`, `duration` from the API response
3. Animate loading steps with a timer during scan
4. Map API response fields correctly:
   - `data.findings` → findings array (note: API returns `name` not `description`, `suggestion` not `remediation`)
   - `data.dependencies` → dependencies array
   - `data.security_score` → score
   - `data.stacks` → detected tech stacks
5. Add a `rescan()` function that calls `startScan` with the stored path

### `web/src/types.ts`
Consolidate ALL types here. The `Finding` interface must match the actual API response:
```typescript
export interface Finding {
  id: string;
  name: string;       // API field (was incorrectly "title" or "description")
  severity: string;
  file: string;
  line: number;
  suggestion: string; // API field (was incorrectly "remediation")
  owasp?: string;
  audit_status: string;
  ai_analysis?: string; // Added by frontend after /api/analyze call
}
```

### `web/src/components/LoginPage.tsx`
- Replace mock auth with real `POST /api/login`
- Handle 401 response properly
- Show specific error messages from API

### `web/src/components/Dashboard.tsx`
- Accept scan metadata (stacks, grade, duration, scan_id) as props
- Call `GET /api/health` on mount, display tool status in sidebar
- Add "[ RE-SCAN ]" button that triggers rescan
- Fix data mapping (no `files_count`, no `timestamp` from API)

### `web/src/components/Triage.tsx`
- Accept `onTriage(id, file, action)` and `onAnalyze(id)` callbacks
- Wire "EXECUTE_SMART_FIX" to `onTriage(id, file, 'FIX')`
- Wire "MARK_FALSE_POSITIVE" to `onTriage(id, file, 'IGNORE')`
- Add "[ AI_ANALYZE ]" button → `onAnalyze(id)` → display result
- Display `finding.name` (not `finding.description`) as the primary text
- Display `finding.suggestion` in the detail sidebar
- Display `finding.owasp` mapping if present
- Display `finding.audit_status` with visual indicator (OPEN=default, IGNORED=dimmed, FIXED=green)
- Add source code preview: when finding selected, call `GET /api/file?path=<file>` and show relevant lines

### `web/src/components/Chat.tsx`
- Replace mock response with real `POST /api/chat`
- Send full conversation history as `messages` array
- Handle `ok: false` gracefully (show error in chat as system message)
- Add loading indicator while waiting for AI response

### `web/src/components/AdminPanel.tsx`
- Fetch users from `GET /api/admin/users` on mount
- Add inline "Create User" form (username, password, is_admin checkbox)
- Add "Delete" button per user row → `DELETE /api/admin/users?username=X`
- Disable delete on "admin" user
- Refresh list after mutations
- Remove props dependency — make it self-contained with its own state

### `web/src/components/Browser.tsx`
- Use `entry.path` from API response instead of manually building paths
- Update `Entry` interface to include `path: string`

### `web/src/components/Dependencies.tsx`
- Already mostly correct — just ensure data flows from `useScan.report.dependencies`

### `web/src/App.tsx`
- Pass all necessary callbacks to child components
- Pass `triageFinding` and `analyzeFinding` to Triage
- Make AdminPanel self-contained (remove empty prop stubs)
- Store scan path for re-scan capability
- Fix the `report?.files_count` reference (doesn't exist in API response)
- Fix the `report?.timestamp` reference (doesn't exist in API response)

---

## Part 6: Validation Checklist

Before marking complete, verify ALL of the following:

- [ ] `npm run build` succeeds with ZERO TypeScript errors
- [ ] All `import type` syntax is correct for `verbatimModuleSyntax`
- [ ] No unused variables or imports
- [ ] Login works with `admin`/`admin` credentials via real API
- [ ] JWT cookie is set after login (check via browser DevTools)
- [ ] `/api/me` returns user data after login
- [ ] Browser page loads directory listing from `/api/browser`
- [ ] Clicking directories navigates into them
- [ ] "INITIALIZE_AUDIT" button triggers a real scan via `/api/scan`
- [ ] FoundryLoader shows animated progress during scan
- [ ] Dashboard displays real scan results (security score, findings count)
- [ ] Dashboard health check shows real tool availability from `/api/health`
- [ ] Triage view lists all findings from scan
- [ ] Clicking a finding shows its details in the sidebar
- [ ] "EXECUTE_SMART_FIX" button calls `/api/triage` with action=FIX
- [ ] "MARK_FALSE_POSITIVE" button calls `/api/triage` with action=IGNORE
- [ ] Finding audit_status updates visually after triage action
- [ ] Chat sends real messages to `/api/chat` (graceful error if no API key)
- [ ] Dependencies view shows real dependency data from scan
- [ ] Admin panel loads users from `/api/admin/users`
- [ ] Admin can create new users
- [ ] Admin can delete non-admin users
- [ ] Logout clears the token cookie and returns to login page
- [ ] No console errors in browser DevTools during normal usage

---

## Files to Modify

| File | Changes |
|------|---------|
| `web/src/types.ts` | Consolidate all types, match API response shapes |
| `web/src/hooks/useScan.ts` | Fix data mapping, add progress animation, store path |
| `web/src/services/api.ts` | Fix `sendChat` signature, add `analyze`, `health`, `admin` calls, OR delete if unused |
| `web/src/App.tsx` | Fix report field references, wire callbacks, fix AdminPanel props |
| `web/src/components/LoginPage.tsx` | Real API login |
| `web/src/components/Dashboard.tsx` | Health check, fix data references, add re-scan |
| `web/src/components/Triage.tsx` | Wire buttons, add AI analysis, add source preview |
| `web/src/components/Chat.tsx` | Real API chat |
| `web/src/components/AdminPanel.tsx` | Full CRUD via API |
| `web/src/components/Browser.tsx` | Use entry.path, fix interface |

## Files to NOT Modify (design is final)
- `web/src/index.css`
- `web/src/ui/Brand.tsx`
- `web/src/ui/PageLayout.tsx`
- `web/src/ui/DataGrid.tsx`
- `web/src/ui/MechanicalButton.tsx`
- `web/src/ui/Telemetry.tsx`
- `web/nginx.conf`
- `web/vite.config.ts`
- Any Go backend files
