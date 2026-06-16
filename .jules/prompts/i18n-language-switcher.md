# Task: Implement EN/RU Language Switcher for AITriage Web Dashboard

## Overview

Implement a complete internationalization (i18n) system for the AITriage web dashboard that allows users to switch between English and Russian languages. The language preference must persist across page reloads (use `localStorage`). The switcher UI must match the existing "Silent Luxury / Cyber-Brutalist" design language exactly.

**Do NOT install any external i18n libraries** (no `react-i18next`, no `i18next`). Build a lightweight custom solution using React Context.

---

## Project Structure Context

The web frontend is located at `web/`. It uses:
- **React 19** with TypeScript 6
- **Vite 8** as bundler
- **Tailwind CSS v4** (via `@tailwindcss/vite` plugin)
- **Framer Motion** for animations
- **JetBrains Mono** as the only font (monospaced, cyber-brutalist aesthetic)

### Key directories:
```
web/src/
├── App.tsx              # Main app shell with Header, Sidebar, Footer
├── main.tsx             # Entry point
├── index.css            # Global styles and Tailwind theme tokens
├── types.ts             # Shared TypeScript types
├── ui/                  # Reusable UI primitives (DO NOT modify internals)
│   ├── Brand.tsx         # Logo + "Forensic_Engine" subtitle
│   ├── NavItem.tsx       # (exported from Brand.tsx) Sidebar navigation button
│   ├── PageLayout.tsx    # Page shell with title/subtitle/actions
│   ├── DataGrid.tsx      # Table component
│   ├── MechanicalButton.tsx # Styled button (primary/outline/error variants)
│   └── Telemetry.tsx     # TelemetryCard + TelemetrySidebar
├── components/          # Feature components
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   ├── LoginPage.tsx
│   ├── Dashboard.tsx
│   ├── Triage.tsx
│   ├── Browser.tsx
│   ├── Chat.tsx
│   ├── AdminPanel.tsx
│   ├── Dependencies.tsx
│   ├── DependencyGraph.tsx
│   ├── FoundryLoader.tsx
│   ├── Foundry.tsx
│   ├── SASTView.tsx
│   └── SourceViewer.tsx
└── hooks/
    └── useScan.ts       # Scan orchestration hook
```

### TypeScript Config:
The project uses `verbatimModuleSyntax: true` in `tsconfig.app.json`. This means:
- All type-only imports MUST use `import type { ... }` syntax
- All regular imports MUST use `import { ... }` syntax
- Mixing will cause build failure

### Build command:
`npm run build` runs `tsc -b && vite build`. The TypeScript compiler is strict:
- `noUnusedLocals: true`
- `noUnusedParameters: true`
- `erasableSyntaxOnly: true`

**Your code MUST pass `npm run build` without errors.**

---

## Step 1: Create the i18n Infrastructure

### 1.1 Create `web/src/i18n/types.ts`

```typescript
export type Locale = 'en' | 'ru';

export interface Translations {
  // ─── Sidebar ───
  sidebar: {
    systemModules: string;
    privilegedAccess: string;
    dashboard: string;
    vulnerabilities: string;
    sourceViewer: string;
    supplyChain: string;
    securityOracle: string;
    identityControl: string;
    settings: string;
    operational: string;
  };

  // ─── Header ───
  header: {
    systemNominal: string;
    engineVersion: string;
    secureUplink: string;
    sessionContext: string;
    localTime: string;
  };

  // ─── Footer ───
  footer: {
    consoleActive: string;
    encryption: string;
    latency: string;
    buffers: string;
    nominal: string;
    copyright: string;
  };

  // ─── Login Page ───
  login: {
    secureAccessRequired: string;
    operatorIdentity: string;
    accessToken: string;
    initializeSession: string;
    verifyingCredentials: string;
    criticalFailure: string;
    invalidCredentials: string;
    authServerUnreachable: string;
    warningText: string;
    statusStandby: string;
  };

  // ─── Dashboard ───
  dashboard: {
    title: string;
    subtitle: string;
    orchestrateScan: string;
    securityScore: string;
    fileResources: string;
    criticalVectors: string;
    lastSync: string;
    aggregateStability: string;
    totalObjectCount: string;
    immediateActionRequired: string;
    temporalMarker: string;
    scanOrchestrationLogs: string;
    realtimeFeed: string;
    scanInitialized: string;
    foundVectors: string;
    networkTopology: string;
    apiUplink: string;
    established: string;
    dbSync: string;
    aiCore: string;
    ready: string;
    quickActions: string;
    refreshCaches: string;
    clearLogBuffers: string;
    emergencyHalt: string;
    systemTelemetry: string;
  };

  // ─── Triage ───
  triage: {
    title: string;
    subtitle: string;
    queryPlaceholder: string;
    runFullAudit: string;
    id: string;
    severity: string;
    threatDescription: string;
    vectorLocation: string;
    threatAnalysis: string;
    severityScore: string;
    incidentSummary: string;
    sourceIndex: string;
    executeSmartFix: string;
    markFalsePositive: string;
    awaitingSelection: string;
    selectVector: string;
  };

  // ─── Browser ───
  browser: {
    title: string;
    subtitle: string;
    scanTarget: string;
    parentDir: string;
  };

  // ─── Chat ───
  chat: {
    title: string;
    subtitle: string;
    agentInput: string;
    orchestratorOutput: string;
    orchestratorOnline: string;
    askPlaceholder: string;
    transmit: string;
    suggestedQueries: string;
    queries: string[];
    model: string;
    tokenUsage: string;
    sessionPayload: string;
    orchestrationDelay: string;
    oracleContext: string;
  };

  // ─── Admin ───
  admin: {
    title: string;
    subtitle: string;
    provisionAgent: string;
    uid: string;
    emailIdentity: string;
    clearance: string;
    status: string;
    lastAuthTimestamp: string;
  };

  // ─── FoundryLoader ───
  loader: {
    foundryActive: string;
    engagingSubsystems: string;
    activeProcess: string;
    initializingCore: string;
    loadRatio: string;
    nominal: string;
    cautionText: string;
  };

  // ─── Empty State (App.tsx) ───
  emptyState: {
    noSecurityContext: string;
    requiresProject: string;
    projectBrowser: string;
    openBrowserTerminal: string;
    initializingBiometrics: string;
    moduleSyncing: string;
    systemCriticalException: string;
  };

  // ─── Common ───
  common: {
    unknown: string;
    never: string;
    completed: string;
  };

  // ─── Language Switcher ───
  langSwitcher: {
    label: string;
  };
}
```

### 1.2 Create `web/src/i18n/en.ts`

Export a `const en: Translations` object. Extract ALL hardcoded English strings from every component listed above. Here is the COMPLETE mapping of every string currently in the codebase that needs translation:

**Sidebar.tsx strings:**
- `":: SYSTEM_MODULES"` → `sidebar.systemModules`
- `":: PRIVILEGED_ACCESS"` → `sidebar.privilegedAccess`
- `"Dashboard"` → `sidebar.dashboard`
- `"Vulnerabilities"` → `sidebar.vulnerabilities`
- `"Source Viewer"` → `sidebar.sourceViewer`
- `"Supply Chain"` → `sidebar.supplyChain`
- `"Security Oracle"` → `sidebar.securityOracle`
- `"Identity Control"` → `sidebar.identityControl`
- `"[ SETTINGS ]"` → `sidebar.settings`
- `"OPERATIONAL"` → `sidebar.operational`

**Header.tsx strings:**
- `"> SYSTEM_NOMINAL"` → `header.systemNominal`
- `"ENGINE_v1.4.2_LTS"` → `header.engineVersion`
- `"SECURE_UPLINK_ESTABLISHED"` → `header.secureUplink`
- `"SESSION_CONTEXT"` → `header.sessionContext`
- `"LOCAL_TIME"` → `header.localTime`

**App.tsx footer strings:**
- `"CONSOLE_ACTIVE"` → `footer.consoleActive`
- `"ENCRYPTION: AES-256-GCM"` → `footer.encryption`
- `"LATENCY:"` → `footer.latency`
- `"BUFFERS:"` → `footer.buffers`
- `"NOMINAL"` → `footer.nominal`
- `"© 2026 CYBERTORTUGA_FOUNDRY"` → `footer.copyright`

**App.tsx empty state strings:**
- `"> NO_SECURITY_CONTEXT"` → `emptyState.noSecurityContext`
- `"AITriage requires an active project directory..."` → `emptyState.requiresProject`
- `"PROJECT_BROWSER"` → `emptyState.projectBrowser`
- `"[ OPEN_BROWSER_TERMINAL ]"` → `emptyState.openBrowserTerminal`
- `"INITIALIZING_BIOMETRICS..."` → `emptyState.initializingBiometrics`
- `"Module [ ... ] synchronization in progress..."` → `emptyState.moduleSyncing`
- `">> SYSTEM_CRITICAL_EXCEPTION:"` → `emptyState.systemCriticalException`

**LoginPage.tsx strings:**
- `"SECURE_ACCESS_REQUIRED"` → `login.secureAccessRequired`
- `"> OPERATOR_IDENTITY"` → `login.operatorIdentity`
- `"> ACCESS_TOKEN"` → `login.accessToken`
- `"[ INITIALIZE_SESSION ]"` → `login.initializeSession`
- `"[ VERIFYING_CREDENTIALS... ]"` → `login.verifyingCredentials`
- `"CRITICAL_FAILURE:"` → `login.criticalFailure`
- `"INVALID_CREDENTIALS"` → `login.invalidCredentials`
- `"AUTH_SERVER_UNREACHABLE"` → `login.authServerUnreachable`
- `"This system is protected by the AITriage Neural-Forensic Layer..."` → `login.warningText`
- `"STATUS: STANDBY"` → `login.statusStandby`

**Dashboard.tsx strings:**
- `"System_Dashboard"` → `dashboard.title`
- `"Operational oversight of project security posture."` → `dashboard.subtitle`
- `"[ ORCHESTRATE_SCAN ]"` → `dashboard.orchestrateScan`
- `"SECURITY_SCORE"` → `dashboard.securityScore`
- `"FILE_RESOURCES"` → `dashboard.fileResources`
- `"CRITICAL_VECTORS"` → `dashboard.criticalVectors`
- `"LAST_SYNC"` → `dashboard.lastSync`
- `":: SCAN_ORCHESTRATION_LOGS"` → `dashboard.scanOrchestrationLogs`
- `"REALTIME_FEED"` → `dashboard.realtimeFeed`
- `"NETWORK_TOPOLOGY"` → `dashboard.networkTopology`
- `"API_UPLINK"` → `dashboard.apiUplink`
- `"ESTABLISHED"` → `dashboard.established`
- `"DB_SYNC"` → `dashboard.dbSync`
- `"NOMINAL"` → `dashboard.nominal` (reuse from `footer.nominal` or keep separate)
- `"AI_CORE"` → `dashboard.aiCore`
- `"READY"` → `dashboard.ready`
- `"QUICK_ACTIONS"` → `dashboard.quickActions`
- `"[ REFRESH_CACHES ]"` → `dashboard.refreshCaches`
- `"[ CLEAR_LOG_BUFFERS ]"` → `dashboard.clearLogBuffers`
- `"[ EMERGENCY_HALT ]"` → `dashboard.emergencyHalt`
- `"SYSTEM_TELEMETRY"` → `dashboard.systemTelemetry`

**Triage.tsx strings:**
- `"Security_Triage"` → `triage.title`
- `"Interactive vulnerability resolution..."` → `triage.subtitle`
- `"QUERY:"` / `"FILTER_VECTORS..."` → `triage.queryPlaceholder`
- `"[ RUN_FULL_AUDIT ]"` → `triage.runFullAudit`
- Column headers: `"ID"`, `"SEV"`, `"THREAT_DESCRIPTION"`, `"VECTOR_LOCATION"`
- `"THREAT_ANALYSIS"` → `triage.threatAnalysis`
- `"SEVERITY_SCORE"` → `triage.severityScore`
- `"INCIDENT_SUMMARY"` → `triage.incidentSummary`
- `"SOURCE_INDEX"` → `triage.sourceIndex`
- `"[ EXECUTE_SMART_FIX ]"` → `triage.executeSmartFix`
- `"[ MARK_FALSE_POSITIVE ]"` → `triage.markFalsePositive`

**Chat.tsx strings:**
- `"Security_Oracle"` → `chat.title`
- `"LLM-Powered natural language audit & discovery."` → `chat.subtitle`
- `"AGENT_INPUT"` / `"ORCHESTRATOR_OUTPUT"` → `chat.agentInput` / `chat.orchestratorOutput`
- `"ORCHESTRATOR_ONLINE. Awaiting security query..."` → `chat.orchestratorOnline`
- `"ASK_SECURITY_CONTEXT..."` → `chat.askPlaceholder`
- `"[ TRANSMIT ]"` → `chat.transmit`
- `":: SUGGESTED_QUERIES"` → `chat.suggestedQueries`
- The 4 suggested query strings → `chat.queries` array

**FoundryLoader.tsx strings:**
- `"Foundry_Active"` → `loader.foundryActive`
- `"Engaging_Security_Subsystems"` → `loader.engagingSubsystems`
- `"ACTIVE_PROCESS"` → `loader.activeProcess`
- `"INITIALIZING_CORE..."` → `loader.initializingCore`
- `"LOAD_RATIO"` → `loader.loadRatio`
- `"[ NOMINAL ]"` → `loader.nominal`
- `"CAUTION: DO NOT INTERRUPT SYSTEM ORCHESTRATION..."` → `loader.cautionText`

**AdminPanel.tsx strings:**
- `"System_Administration"` → `admin.title`
- `"User Management & Access Control Layer."` → `admin.subtitle`
- `"[ PROVISION_NEW_AGENT ]"` → `admin.provisionAgent`
- Column headers: `"UID"`, `"EMAIL_IDENTITY"`, `"CLEARANCE"`, `"STATUS"`, `"LAST_AUTH_TIMESTAMP"`

### 1.3 Create `web/src/i18n/ru.ts`

Export a `const ru: Translations` object. **CRITICAL STYLING RULE:** Russian translations must maintain the same cyber-brutalist, "audit-grade" tone. Do NOT use casual Russian. Use dry, technical, military-style terminology with underscores between words, uppercase everywhere. Examples:

| English | Russian |
|---------|---------|
| `SYSTEM_MODULES` | `СИСТЕМНЫЕ_МОДУЛИ` |
| `Dashboard` | `Мониторинг` |
| `Vulnerabilities` | `Уязвимости` |
| `Source Viewer` | `Исходный_Код` |
| `Supply Chain` | `Цепочка_Поставок` |
| `Security Oracle` | `Оракул_Безопасности` |
| `Identity Control` | `Контроль_Доступа` |
| `OPERATIONAL` | `АКТИВЕН` |
| `SYSTEM_NOMINAL` | `СИСТЕМА_ШТАТНО` |
| `SECURE_UPLINK_ESTABLISHED` | `ЗАЩИЩЁННЫЙ_КАНАЛ_УСТАНОВЛЕН` |
| `SESSION_CONTEXT` | `КОНТЕКСТ_СЕССИИ` |
| `LOCAL_TIME` | `ЛОКАЛЬНОЕ_ВРЕМЯ` |
| `CONSOLE_ACTIVE` | `КОНСОЛЬ_АКТИВНА` |
| `ENCRYPTION: AES-256-GCM` | `ШИФРОВАНИЕ: AES-256-GCM` |
| `NO_SECURITY_CONTEXT` | `КОНТЕКСТ_БЕЗОПАСНОСТИ_ОТСУТСТВУЕТ` |
| `INITIALIZING_BIOMETRICS...` | `ИНИЦИАЛИЗАЦИЯ_БИОМЕТРИИ...` |
| `SECURE_ACCESS_REQUIRED` | `ТРЕБУЕТСЯ_АВТОРИЗАЦИЯ` |
| `OPERATOR_IDENTITY` | `ИДЕНТИФИКАТОР_ОПЕРАТОРА` |
| `ACCESS_TOKEN` | `ТОКЕН_ДОСТУПА` |
| `INITIALIZE_SESSION` | `ИНИЦИАЛИЗИРОВАТЬ_СЕССИЮ` |
| `System_Dashboard` | `Система_Мониторинга` |
| `Operational oversight of project security posture.` | `Оперативный контроль состояния безопасности проекта.` |
| `ORCHESTRATE_SCAN` | `ЗАПУСТИТЬ_СКАНИРОВАНИЕ` |
| `SECURITY_SCORE` | `ОЦЕНКА_БЕЗОПАСНОСТИ` |
| `FILE_RESOURCES` | `ФАЙЛОВЫЕ_РЕСУРСЫ` |
| `CRITICAL_VECTORS` | `КРИТИЧЕСКИЕ_ВЕКТОРЫ` |
| `Security_Triage` | `Сортировка_Угроз` |
| `THREAT_DESCRIPTION` | `ОПИСАНИЕ_УГРОЗЫ` |
| `VECTOR_LOCATION` | `РАСПОЛОЖЕНИЕ_ВЕКТОРА` |
| `EXECUTE_SMART_FIX` | `ПРИМЕНИТЬ_ИСПРАВЛЕНИЕ` |
| `MARK_FALSE_POSITIVE` | `ЛОЖНОЕ_СРАБАТЫВАНИЕ` |
| `Foundry_Active` | `Литейная_Активна` |
| `CAUTION: DO NOT INTERRUPT...` | `ВНИМАНИЕ: НЕ ПРЕРЫВАЙТЕ ОРКЕСТРАЦИЮ СИСТЕМЫ. КОНТЕКСТ БЕЗОПАСНОСТИ НЕСТАБИЛЕН.` |
| `Security_Oracle` | `Оракул_Безопасности` |
| `TRANSMIT` | `ПЕРЕДАТЬ` |
| `System_Administration` | `Администрирование_Системы` |
| `PROVISION_NEW_AGENT` | `СОЗДАТЬ_НОВОГО_АГЕНТА` |
| Chat suggested queries | Translate to equivalent Russian audit-style queries |

For the login page warning text in Russian:
`"Данная система защищена нейро-форензик слоем AITriage. Попытки несанкционированного доступа регистрируются и передаются в ядро оркестрации."`

### 1.4 Create `web/src/i18n/index.ts`

```typescript
export { en } from './en';
export { ru } from './ru';
export type { Locale, Translations } from './types';
```

---

## Step 2: Create the i18n Context and Hook

### Create `web/src/i18n/I18nContext.tsx`

```typescript
import React, { createContext, useContext, useState, useCallback } from 'react';
import type { Locale, Translations } from './types';
import { en } from './en';
import { ru } from './ru';

const translations: Record<Locale, Translations> = { en, ru };

interface I18nContextValue {
  locale: Locale;
  t: Translations;
  setLocale: (locale: Locale) => void;
  toggleLocale: () => void;
}

const I18nContext = createContext<I18nContextValue | null>(null);

const getInitialLocale = (): Locale => {
  const saved = localStorage.getItem('aitriage-locale');
  if (saved === 'en' || saved === 'ru') return saved;
  // Auto-detect from browser
  const browserLang = navigator.language.toLowerCase();
  if (browserLang.startsWith('ru')) return 'ru';
  return 'en';
};

export const I18nProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [locale, setLocaleState] = useState<Locale>(getInitialLocale);

  const setLocale = useCallback((newLocale: Locale) => {
    setLocaleState(newLocale);
    localStorage.setItem('aitriage-locale', newLocale);
    document.documentElement.lang = newLocale;
  }, []);

  const toggleLocale = useCallback(() => {
    setLocale(locale === 'en' ? 'ru' : 'en');
  }, [locale, setLocale]);

  const value: I18nContextValue = {
    locale,
    t: translations[locale],
    setLocale,
    toggleLocale,
  };

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

export const useI18n = (): I18nContextValue => {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n must be used within I18nProvider');
  return ctx;
};
```

---

## Step 3: Wrap the App

In `web/src/main.tsx`, wrap `<App />` with `<I18nProvider>`:

```tsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { I18nProvider } from './i18n/I18nContext.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <I18nProvider>
      <App />
    </I18nProvider>
  </StrictMode>,
)
```

---

## Step 4: Add the Language Switcher UI

### 4.1 Create `web/src/ui/LanguageSwitcher.tsx`

This component renders a segmented control that looks like: `[ EN ] | RU` or `EN | [ RU ]`

Design requirements:
- Use existing design tokens from `index.css` (colors like `primary-fixed-dim`, `outline-variant`, `surface-container`, `on-surface`, `on-surface-variant`)
- Zero border-radius (this is cyber-brutalist — no rounded corners anywhere)
- Font: JetBrains Mono (inherited from `font-code` class)
- Text size: `text-[10px]`, uppercase, bold, tracking-wide
- Active segment: `bg-primary-fixed-dim text-background font-black`
- Inactive segment: `bg-transparent text-on-surface-variant hover:text-on-surface`
- Container: `border border-outline-variant`
- Total width: compact, fits in the Header

```tsx
import React from 'react';
import { useI18n } from '../i18n/I18nContext';
import type { Locale } from '../i18n/types';

export const LanguageSwitcher: React.FC = () => {
  const { locale, setLocale } = useI18n();

  const segments: { id: Locale; label: string }[] = [
    { id: 'en', label: 'EN' },
    { id: 'ru', label: 'RU' },
  ];

  return (
    <div className="flex border border-outline-variant">
      {segments.map((seg) => (
        <button
          key={seg.id}
          onClick={() => setLocale(seg.id)}
          className={`px-3 py-1.5 text-[10px] font-black uppercase tracking-[0.2em] transition-all ${
            locale === seg.id
              ? 'bg-primary-fixed-dim text-background'
              : 'bg-transparent text-on-surface-variant hover:text-on-surface hover:bg-surface-container'
          }`}
        >
          {seg.label}
        </button>
      ))}
    </div>
  );
};
```

### 4.2 Place it in `Header.tsx`

Insert `<LanguageSwitcher />` in the Header, right before the `SESSION_CONTEXT` / `LOCAL_TIME` block (the right side of the header). Add a vertical separator `<div className="h-8 w-[1px] bg-outline-variant" />` between it and adjacent elements.

---

## Step 5: Replace All Hardcoded Strings

In EVERY component, add `const { t } = useI18n();` and replace all hardcoded strings with their `t.section.key` equivalent.

### Components to modify (exhaustive list):

1. **`web/src/App.tsx`** — empty state, footer, loading screen, error banner
2. **`web/src/components/Header.tsx`** — all status text
3. **`web/src/components/Sidebar.tsx`** — nav labels, section headers, status text
4. **`web/src/components/LoginPage.tsx`** — form labels, buttons, warning text
5. **`web/src/components/Dashboard.tsx`** — page title/subtitle, telemetry labels, log text, sidebar labels, button labels
6. **`web/src/components/Triage.tsx`** — page title/subtitle, column headers, detail sidebar labels, action buttons
7. **`web/src/components/Chat.tsx`** — page title/subtitle, message role labels, placeholder, button, suggested queries
8. **`web/src/components/AdminPanel.tsx`** — page title/subtitle, column headers, button labels
9. **`web/src/components/FoundryLoader.tsx`** — title, status text, progress labels, caution text
10. **`web/src/components/Browser.tsx`** — page title/subtitle, action buttons
11. **`web/src/components/Dependencies.tsx`** — page title/subtitle, column headers
12. **`web/src/components/DependencyGraph.tsx`** — page title/subtitle
13. **`web/src/components/Foundry.tsx`** — page title/subtitle
14. **`web/src/components/SASTView.tsx`** — page title/subtitle, column headers
15. **`web/src/components/SourceViewer.tsx`** — page title/subtitle

### Example transformation for Sidebar.tsx:

BEFORE:
```tsx
<div className="text-[10px] ...">:: SYSTEM_MODULES</div>
<NavItem label="Dashboard" ... />
```

AFTER:
```tsx
const { t } = useI18n();
// ...
<div className="text-[10px] ...">:: {t.sidebar.systemModules}</div>
<NavItem label={t.sidebar.dashboard} ... />
```

---

## Step 6: Brand Component Special Handling

The `Brand` component in `web/src/ui/Brand.tsx` renders `"AITriage"` and `"Forensic_Engine"`. The brand name `"AITriage"` must NOT be translated — it stays as-is in both languages. The subtitle `"Forensic_Engine"` should be translated to `"Форензик_Движок"` in Russian. You will need to:

1. Add `subtitle` to the `Translations` interface under a `brand` section
2. Pass the `useI18n()` hook result into Brand, OR make Brand accept an optional `subtitle` prop and pass it from the parent components (Header, LoginPage, FoundryLoader)

The simpler approach: make `Brand` accept an optional `subtitle?: string` prop, defaulting to `"Forensic_Engine"`. Then in each parent that renders `<Brand />`, pass `t.brand.subtitle`.

---

## Step 7: Validation Checklist

Before marking this task as complete, verify:

- [ ] `npm run build` succeeds with zero errors
- [ ] All imports use `import type` for type-only imports
- [ ] No unused variables or imports remain
- [ ] Language switcher appears in the Header and toggles instantly
- [ ] localStorage persists the choice across page reloads
- [ ] ALL visible text in every component switches between EN and RU
- [ ] Russian text maintains the uppercase, underscore-separated, cyber-brutalist aesthetic
- [ ] The `<html lang="">` attribute updates when language changes
- [ ] No external i18n libraries were installed

---

## Files Created (new):
- `web/src/i18n/types.ts`
- `web/src/i18n/en.ts`
- `web/src/i18n/ru.ts`
- `web/src/i18n/index.ts`
- `web/src/i18n/I18nContext.tsx`
- `web/src/ui/LanguageSwitcher.tsx`

## Files Modified:
- `web/src/main.tsx` (wrap with I18nProvider)
- `web/src/App.tsx`
- `web/src/components/Header.tsx`
- `web/src/components/Sidebar.tsx`
- `web/src/components/LoginPage.tsx`
- `web/src/components/Dashboard.tsx`
- `web/src/components/Triage.tsx`
- `web/src/components/Chat.tsx`
- `web/src/components/AdminPanel.tsx`
- `web/src/components/FoundryLoader.tsx`
- `web/src/components/Browser.tsx`
- `web/src/components/Dependencies.tsx`
- `web/src/components/DependencyGraph.tsx`
- `web/src/components/Foundry.tsx`
- `web/src/components/SASTView.tsx`
- `web/src/components/SourceViewer.tsx`
- `web/src/ui/Brand.tsx`
