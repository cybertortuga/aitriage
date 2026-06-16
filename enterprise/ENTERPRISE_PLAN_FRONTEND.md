# AITriage Enterprise — Frontend Plan

---

## PHASE 9: FRONTEND — AUTH & NAVIGATION

### 9.1 Auth Flow (уже есть, улучшить)
- [ ] `LoginPage.tsx` — уже есть, добавить "Forgot password" UI placeholder
- [ ] `AuthStore.ts` — хранить `user_id`, `global_role`, `email`, `full_name` из JWT
- [ ] `ProtectedRoute.tsx` — HOC для защиты роутов по GlobalRole
- [ ] `usePermissions()` hook — `can(action, resource)` — проверка прав в компонентах
- [ ] Redirect на `/login` при 401 с любого API запроса (axios interceptor)

### 9.2 Navigation / Sidebar
- [ ] Sidebar показывает пункты меню в зависимости от роли:
  - Все: Dashboard, Products, Findings, Kanban
  - analyst+: Scan
  - security_lead+: Engagements, Reports
  - admin+: Admin Panel, Audit Log
- [ ] Notification bell с badge (unread count)
- [ ] User avatar + role badge в нижней части sidebar
- [ ] Breadcrumb: `Dashboard > Product: MyApp > Engagement #12 > Finding #34`

---

## PHASE 10: FRONTEND — DASHBOARD

### 10.1 `DashboardPage.tsx` — Главная
- [ ] **Header**: "Security Posture Overview" + фильтр по product/date
- [ ] **Top KPI Bar** (горизонтальные карточки):
  - CRITICAL findings (красный badge)
  - HIGH findings (оранжевый)
  - SLA Breached (мигающий красный)
  - Overall Security Score (0-100, цветовая шкала)
- [ ] **Findings by Severity** — stacked bar chart (последние 30 дней)
- [ ] **Top Risky Products** — таблица с ProductName, Open Critical, Security Score, SLA Status
- [ ] **Recent Engagements** — последние 5 сканирований с линком
- [ ] **MTTR Gauge** — среднее время на исправление по severity
- [ ] **Findings Status Distribution** — pie chart: open/in_progress/fixed/wont_fix/fp
- [ ] Все блоки — "Silent Luxury" TUI style, monospace, dark

### 10.2 Notifications Panel
- [ ] Floating panel при клике на bell icon
- [ ] Группировка: Critical / Warning / Info
- [ ] "Mark all read" кнопка
- [ ] Каждая нотификация — кликабельная, ведёт к сущности

---

## PHASE 11: FRONTEND — PRODUCTS (DefectDojo-style)

### 11.1 `ProductsPage.tsx` — Список продуктов
- [ ] Таблица: Name, Type, Business Criticality, Open Findings (CRIT/HIGH), Security Score, SLA Status, Last Scan
- [ ] Фильтры: Product Type, Lifecycle, Criticality
- [ ] Поиск по имени
- [ ] Кнопка "New Product" (для security_lead+)
- [ ] Каждая строка — кликабельна, ведёт на `/products/:id`

### 11.2 `ProductDetailPage.tsx`
- [ ] **Header**: Product Name, Type, Repo URL (клик → Github), Lifecycle badge
- [ ] **Tabs**:
  - `Overview` — метрики, security score trend
  - `Engagements` — список сканов, кнопка "Start New Scan"
  - `Findings` — таблица всех findings с фильтрами
  - `Kanban` — доска (см. Phase 12)
  - `Members` — список участников + добавление (для owner)
  - `Settings` — SLA config, product settings (для owner+)

### 11.3 `ProductSettingsModal.tsx`
- [ ] Форма: Name, Description, Repo URL, Product Type, Lifecycle, Business Criticality
- [ ] SLA Settings: дни для CRITICAL/HIGH/MEDIUM/LOW
- [ ] Кнопка "Archive Product" (для admin+)

### 11.4 `ProductMembersTab.tsx`
- [ ] Таблица: Avatar, Username, Email, Role, Actions
- [ ] "Add Member" — search by username + select role
- [ ] Role dropdown: owner / security_lead / analyst / developer / viewer
- [ ] "Remove" — удалить из проекта

---

## PHASE 12: FRONTEND — KANBAN BOARD

### 12.1 `KanbanBoard.tsx` — Доска
- [ ] Установить `@dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`
- [ ] 5 колонок: **Backlog | Todo | In Progress | Review | Done**
- [ ] Drag-and-drop карточек между колонками
- [ ] При drop → `PATCH /api/findings/:id/move` + optimistic update
- [ ] Колонки скроллируются вертикально независимо
- [ ] Счётчик в заголовке каждой колонки
- [ ] Filter Bar над доской: Severity, Assignee, Search по title

### 12.2 `FindingCard.tsx` — Карточка finding
- [ ] Severity badge (цвет по severity: CRITICAL=red, HIGH=orange, MEDIUM=yellow, LOW=blue, INFO=gray)
- [ ] Rule ID монospace (напр. `SEC-0042`)
- [ ] Title (truncated, full tooltip)
- [ ] File path + line
- [ ] Assignee avatar (или "Unassigned")
- [ ] SLA indicator: зеленый/желтый/красный + "X days left / BREACHED"
- [ ] Tags: `VERIFIED` / `DUPLICATE` / `FP` / `RISK_ACCEPTED`
- [ ] Quick actions при hover:
  - Assign to me
  - Mark Fixed
  - Open Detail

### 12.3 `FindingDetailModal.tsx` — Детальный просмотр
- [ ] **Tabs**:
  - `Overview` — полное описание, impact, CWE/CVE
  - `Code` — code snippet с highlight + file path
  - `Fix Suggestion` — AI рекомендация (из scan)
  - `Notes` — комментарии (textarea + список)
  - `History` — audit trail этого finding
- [ ] **Actions Panel** (правая колонка):
  - Status selector (dropdown)
  - Kanban column selector
  - Assign To (search user)
  - "Accept Risk" button → modal с reason + expiry
  - "Mark False Positive" button → modal с reason
  - "Mark Verified" button (analyst+)
  - "Mark Duplicate" → search другой finding
- [ ] **SLA Block**: deadline, days remaining, breached indicator

---

## PHASE 13: FRONTEND — FINDINGS LIST

### 13.1 `FindingsPage.tsx`
- [ ] Глобальный список всех findings (cross-product)
- [ ] **Фильтры**:
  - Product (multi-select)
  - Severity (checkboxes: CRITICAL/HIGH/MEDIUM/LOW/INFO)
  - Status (multi-select)
  - Assignee (dropdown)
  - SLA Breached (toggle)
  - Date range (created_at)
  - Rule ID (text search)
- [ ] **Сортировка**: severity, date, sla_deadline, status
- [ ] **Таблица**: #ID, Severity, Title, Product, File, Status, Assignee, SLA, Date
- [ ] Bulk Actions: assign, change status, export
- [ ] Export → CSV/JSON (текущий отфильтрованный список)
- [ ] Клик на строку → `FindingDetailModal`

---

## PHASE 14: FRONTEND — ADMIN PANEL

### 14.1 `AdminPanel.tsx` — Роутинг
- [ ] Доступен только для admin/superadmin
- [ ] **Tabs**: Users | Product Types | Audit Log | System Settings | Notifications Config

### 14.2 `UsersTab.tsx`
- [ ] Таблица: ID, Username, Email, Full Name, Global Role, Active, Last Login, Actions
- [ ] "Create User" button → modal:
  - Username (required)
  - Email
  - Full Name
  - Password
  - Global Role (select)
  - Is Active (toggle)
- [ ] "Edit User" → тот же modal заполненный
- [ ] "Reset Password" → modal с новым паролем
- [ ] "Deactivate" / "Activate" toggle
- [ ] "Delete" — только для superadmin, только неактивных
- [ ] Поиск по username/email

### 14.3 `AuditLogTab.tsx`
- [ ] Таблица: Timestamp, User, Action, Entity Type, Entity ID, Old Value, New Value, IP
- [ ] Фильтры: action type, user, date range, entity type
- [ ] Раскрываемые строки для просмотра full old/new value
- [ ] "Export" → JSON

### 14.4 `SystemSettingsTab.tsx`
- [ ] JWT Secret rotation UI
- [ ] Default SLA values (глобальные)
- [ ] Session timeout config
- [ ] Scanner binary path config
- [ ] LLM API key management (с маскировкой)

---

## PHASE 15: FRONTEND — REPORTS

### 15.1 `ReportsPage.tsx`
- [ ] **Executive Summary Report** — по product:
  - Security Score trend
  - Open findings by severity (pie chart)
  - SLA compliance %
  - MTTR
  - Fixed vs Open ratio
- [ ] **Engagement Report** — итоги конкретного скана:
  - Все findings
  - Comparison с предыдущим engagement (new/fixed/regressed)
- [ ] Export → PDF (через `window.print()` + print CSS) / JSON / CSV
- [ ] "Share Report" → генерирует публичный read-only токен (опционально)

---

## PHASE 16: DOCKER & DEPLOYMENT

### 16.1 Docker Compose обновление
- [ ] `web` container — отдельный multi-stage build (Node → nginx или копирование в Go container)
- [ ] `backend` container — Go binary + SQLite файл в volume
- [ ] `db-volume` — persist SQLite file между перезапусками
- [ ] Health check для backend: `GET /api/health`
- [ ] Environment variables:
  - `JWT_SECRET` (обязателен в production)
  - `DB_PATH` (путь к SQLite файлу)
  - `LOG_LEVEL`
  - `LLM_API_KEY`

### 16.2 Build & CI
- [ ] `Makefile` target: `make enterprise-up` — полная сборка и запуск
- [ ] `Makefile` target: `make db-migrate` — ручной запуск миграций
- [ ] `Makefile` target: `make create-admin` — интерактивное создание superadmin
- [ ] `.env.example` — шаблон переменных для deployment

---

## PHASE 17: FINAL AUDIT

### 17.1 Security Hardening
- [ ] Все admin endpoints защищены `permissionMiddleware`
- [ ] JWT expiry проверяется на каждом запросе
- [ ] Rate limiting на `/api/login` (max 5 попыток за 15 мин)
- [ ] SQL injection: только prepared statements (параметры через `?`)
- [ ] XSS: Content-Security-Policy header
- [ ] Secrets: JWT_SECRET из ENV, не захардкожен
- [ ] `users.json.migrated` удаляется или помещается в .gitignore

### 17.2 Testing Checklist
- [ ] Login / Logout flow
- [ ] Role enforcement: viewer не может triagить
- [ ] Scan → DB → Kanban pipeline E2E
- [ ] Deduplication: второй одинаковый скан не создаёт дубли
- [ ] SLA breach notification появляется
- [ ] Kanban drag-and-drop обновляет status в DB
- [ ] Risk acceptance скрывает finding из active count
- [ ] Audit log записывает все действия

### 17.3 Documentation
- [ ] `README.md` — Enterprise Setup Guide
- [ ] API docs (markdown table всех endpoints)
- [ ] RBAC guide: кто что может
- [ ] Docker deployment guide

---

## SUMMARY: Список всех новых файлов

### Backend (Go)
```
internal/server/db.go
internal/server/models.go
internal/server/repositories/user_repo.go
internal/server/repositories/product_repo.go
internal/server/repositories/engagement_repo.go
internal/server/repositories/finding_repo.go
internal/server/repositories/audit_repo.go
internal/server/repositories/notification_repo.go
internal/server/handlers/product_handlers.go
internal/server/handlers/engagement_handlers.go
internal/server/handlers/finding_handlers.go
internal/server/handlers/kanban_handlers.go
internal/server/handlers/notification_handlers.go
internal/server/handlers/audit_handlers.go
internal/server/handlers/report_handlers.go
internal/server/middleware/permission.go
internal/server/middleware/audit.go
internal/server/sla/scheduler.go
```

### Frontend (React/TypeScript)
```
web/src/pages/DashboardPage.tsx
web/src/pages/ProductsPage.tsx
web/src/pages/ProductDetailPage.tsx
web/src/pages/FindingsPage.tsx
web/src/pages/KanbanPage.tsx
web/src/pages/AdminPanel.tsx
web/src/pages/ReportsPage.tsx
web/src/components/kanban/KanbanBoard.tsx
web/src/components/kanban/KanbanColumn.tsx
web/src/components/kanban/FindingCard.tsx
web/src/components/findings/FindingDetailModal.tsx
web/src/components/findings/FindingFilters.tsx
web/src/components/products/ProductSettingsModal.tsx
web/src/components/products/ProductMembersTab.tsx
web/src/components/admin/UsersTab.tsx
web/src/components/admin/AuditLogTab.tsx
web/src/components/admin/SystemSettingsTab.tsx
web/src/components/common/NotificationPanel.tsx
web/src/components/common/SLABadge.tsx
web/src/components/common/SeverityBadge.tsx
web/src/components/common/RoleBadge.tsx
web/src/hooks/usePermissions.ts
web/src/hooks/useNotifications.ts
web/src/hooks/useProducts.ts
web/src/hooks/useFindings.ts
web/src/api/products.ts
web/src/api/findings.ts
web/src/api/kanban.ts
web/src/api/admin.ts
web/src/api/notifications.ts
web/src/types/enterprise.ts
```
