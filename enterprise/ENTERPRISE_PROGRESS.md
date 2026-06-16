# AITriage Enterprise — Progress Tracker

> **⚠️ ОБЯЗАТЕЛЬНО К ПРОЧТЕНИЮ ПЕРЕД НАЧАЛОМ РАБОТЫ ⚠️**
>
> Этот файл — единственный источник правды о прогрессе.
> **Строго следуй плану. Не пропускай шаги. Не переходи к следующей фазе без завершения текущей.**
> После завершения каждой подзадачи — обновляй статус в этом файле.
> Полный технический план: `ENTERPRISE_PLAN.md`, `ENTERPRISE_PLAN_PHASES.md`, `ENTERPRISE_PLAN_FRONTEND.md`

---

## ПРАВИЛА ДЛЯ AI-ИСПОЛНИТЕЛЯ

1. **Читай план перед каждой задачей** — открой соответствующий раздел `ENTERPRISE_PLAN_PHASES.md` или `ENTERPRISE_PLAN_FRONTEND.md`
2. **Один шаг за раз** — не делай несколько задач одновременно массово
3. **Обновляй этот файл** — после каждого завершённого пункта меняй `[ ]` на `[x]`
4. **При блокере** — запиши `[!] BLOCKED: <причина>` и остановись, спроси пользователя
5. **Не изменяй архитектуру** без согласования с пользователем
6. **Строго придерживайся стека**: SQLite + pure Go + React 19 + Tailwind CSS 4
7. **Стиль UI**: "Silent Luxury" / TUI / монospace / dark — смотри `ENTERPRISE_PLAN.md`
8. **Commit после каждой фазы**: `git commit -m "enterprise: phase N - <описание>"`

---

## ОБЩИЙ ПРОГРЕСС

| Фаза | Название | Статус | % |
|---|---|---|---|
| Phase 1 | Database Foundation | ✅ Completed | 100% |
| Phase 2 | Backend RBAC & Auth Refactor | ✅ Completed | 100% |
| Phase 3 | Product Management API | 🔲 Not Started | 0% |
| Phase 4 | Engagements & Findings API | 🔲 Not Started | 0% |
| Phase 5 | Kanban API | 🔲 Not Started | 0% |
| Phase 6 | Notifications API | 🔲 Not Started | 0% |
| Phase 7 | Audit Log API | 🔲 Not Started | 0% |
| Phase 8 | Dashboard Metrics API | 🔲 Not Started | 0% |
| Phase 9 | Frontend Auth & Navigation | 🔲 Not Started | 0% |
| Phase 10 | Frontend Dashboard | 🔲 Not Started | 0% |
| Phase 11 | Frontend Products | 🔲 Not Started | 0% |
| Phase 12 | Frontend Kanban Board | 🔲 Not Started | 0% |
| Phase 13 | Frontend Findings List | 🔲 Not Started | 0% |
| Phase 14 | Frontend Admin Panel | 🔲 Not Started | 0% |
| Phase 15 | Frontend Reports | 🔲 Not Started | 0% |
| Phase 16 | Docker & Deployment | 🔲 Not Started | 0% |
| Phase 17 | Final Audit | 🔲 Not Started | 0% |

---

## PHASE 1: DATABASE FOUNDATION
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 1

### 1.1 `internal/server/db.go`
- [x] `InitDB(path string)` — открыть/создать SQLite файл
- [x] WAL mode: `PRAGMA journal_mode=WAL`
- [x] Foreign keys: `PRAGMA foreign_keys=ON`
- [x] Connection pool: `SetMaxOpenConns(1)`
- [x] `RunMigrations(db)` — создать все таблицы
- [x] `SeedDefaultData(db)` — создать default admin + default product type

### 1.2 `internal/server/models.go`
- [x] `User` struct
- [x] `ProductType` struct
- [x] `Product` struct
- [x] `ProductMember` struct
- [x] `Engagement` struct
- [x] `Finding` struct (полный)
- [x] `FindingNote` struct
- [x] `AuditLog` struct
- [x] `Notification` struct
- [x] Хелпер методы: `Finding.IsOverSLA()`, `Finding.DaysToSLA()`

### 1.3 `internal/server/repositories/user_repo.go`
- [x] `GetByUsername()`
- [x] `GetByID()`
- [x] `Create()`
- [x] `Update()`
- [x] `Delete()`
- [x] `List()`
- [x] `UpdateLastLogin()`

### 1.4 `internal/server/repositories/product_repo.go`
- [x] `Create()`
- [x] `GetByID()`
- [x] `List()` с фильтрацией по доступу
- [x] `Update()`
- [x] `Delete()`
- [x] `AddMember()`
- [x] `RemoveMember()`
- [x] `GetMembers()`
- [x] `GetUserRole()`

### 1.5 `internal/server/repositories/engagement_repo.go`
- [x] `Create()`
- [x] `GetByID()`
- [x] `ListByProduct()`
- [x] `Complete()`
- [x] `GetLatestByProduct()`

### 1.6 `internal/server/repositories/finding_repo.go`
- [x] `Create()`
- [x] `BulkCreate()`
- [x] `GetByID()`
- [x] `ListByEngagement()` с фильтрами
- [x] `ListByProduct()` с фильтрами
- [x] `UpdateStatus()`
- [x] `UpdateAssignee()`
- [x] `MarkDuplicate()`
- [x] `FindByHash()` — deduplication
- [x] `SetRiskAccepted()`
- [x] `SetFalsePositive()`
- [x] `GetKanbanBoard()`
- [x] `UpdateSLAStatus()` — для cron job

### 1.7 `internal/server/repositories/audit_repo.go`
- [x] `Log(entry)`
- [x] `List(filters)`

### 1.8 Migration `users.json` → SQLite
- [x] Проверить наличие `users.json` при старте
- [x] Импортировать пользователей в DB
- [x] Переименовать файл в `users.json.migrated`
- [x] Логировать миграцию

**✅ Phase 1 Complete когда:** `go build ./...` проходит, SQLite файл создаётся, пользователи мигрируют

---

## PHASE 2: BACKEND RBAC & AUTH REFACTOR
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 2

### 2.1 JWT Claims расширение
- [x] Добавить `UserID`, `GlobalRole`, `Email` в Claims struct
- [x] Обновить `handleLogin` — из DB
- [x] Обновить `handleMe` — полный профиль
- [x] Настроить `exp`: 24h standard

### 2.2 `permissionMiddleware`
- [x] `RequireGlobalRole(roles ...string)`
- [x] `RequireProductRole(productID, roles ...string)`
- [x] `ExtractClaims()` хелпер
- [x] JSON 403 response с `required` полем

### 2.3 User CRUD API (мигрировать на DB)
- [x] `GET /api/admin/users` → DB
- [x] `POST /api/admin/users` → DB
- [x] `PUT /api/admin/users/:username`
- [x] `DELETE /api/admin/users/:username` (soft delete)
- [x] `POST /api/admin/users/:username/reset-password`
- [x] Все операции → `audit_log`

### 2.4 User Profile API
- [x] `GET /api/me` — полный профиль
- [x] `PUT /api/me` — email, full_name
- [x] `PUT /api/me/password`

**✅ Phase 2 Complete когда:** login/logout работает через DB, RBAC middleware блокирует неавторизованных

---

## PHASE 3: PRODUCT MANAGEMENT API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 3

### 3.1 Product Types API
- [ ] `GET /api/product-types`
- [ ] `POST /api/product-types`
- [ ] `PUT /api/product-types/:id`
- [ ] `DELETE /api/product-types/:id`

### 3.2 Products API
- [ ] `GET /api/products`
- [ ] `GET /api/products/:id`
- [ ] `POST /api/products`
- [ ] `PUT /api/products/:id`
- [ ] `DELETE /api/products/:id`

### 3.3 Product Members API
- [ ] `GET /api/products/:id/members`
- [ ] `POST /api/products/:id/members`
- [ ] `PUT /api/products/:id/members/:userID`
- [ ] `DELETE /api/products/:id/members/:userID`

### 3.4 Product Metrics API
- [ ] `GET /api/products/:id/metrics`
- [ ] Findings by severity агрегация
- [ ] SLA compliance %
- [ ] MTTR расчёт
- [ ] Security Score расчёт
- [ ] Trend 30/60/90 дней

### 3.5 SLA Configuration
- [ ] `PUT /api/products/:id/sla`
- [ ] SLA deadline расчёт при создании finding
- [ ] Background cron: `UpdateSLAStatus()` каждый час
- [ ] Notification при SLA breach

**✅ Phase 3 Complete когда:** можно создать Product, добавить участников, получить метрики

---

## PHASE 4: ENGAGEMENTS & FINDINGS API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 4

### 4.1 Engagements API
- [ ] `GET /api/products/:id/engagements`
- [ ] `GET /api/engagements/:id`
- [ ] `POST /api/products/:id/engagements`
- [ ] `PUT /api/engagements/:id`
- [ ] `POST /api/engagements/:id/complete`

### 4.2 Findings CRUD API
- [ ] `GET /api/findings` (с фильтрами)
- [ ] `GET /api/findings/:id`
- [ ] `PUT /api/findings/:id`
- [ ] `POST /api/findings/:id/notes`
- [ ] `GET /api/findings/:id/notes`
- [ ] `POST /api/findings/:id/risk-accept`
- [ ] `POST /api/findings/:id/false-positive`
- [ ] `POST /api/findings/:id/verify`
- [ ] `POST /api/findings/:id/duplicate`

### 4.3 Deduplication Engine
- [ ] SHA256 hash при создании finding
- [ ] `FindByHash()` lookup при BulkCreate
- [ ] Regression detection (fixed → re-opened)
- [ ] Audit log dedup decision

### 4.4 Scan → DB Integration
- [ ] После `/api/scan` → создать Engagement автоматически
- [ ] Парсить `RichScanResult` → `BulkCreate` findings
- [ ] Применить deduplication
- [ ] Вычислить SLA deadlines
- [ ] Вернуть `engagement_id` в ответе
- [ ] Обновить `/api/triage` → работает с DB

**✅ Phase 4 Complete когда:** после scan findings появляются в DB, triage обновляет статус в DB

---

## PHASE 5: KANBAN API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 5

- [ ] `GET /api/kanban?product_id=X`
- [ ] `PATCH /api/findings/:id/move`
- [ ] `PATCH /api/findings/:id/assign`
- [ ] Audit log при движении
- [ ] Notification назначенному

**✅ Phase 5 Complete когда:** Kanban API возвращает сгруппированные findings

---

## PHASE 6: NOTIFICATIONS API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 6

- [ ] `GET /api/notifications`
- [ ] `PUT /api/notifications/:id/read`
- [ ] `PUT /api/notifications/read-all`
- [ ] `DELETE /api/notifications/:id`
- [ ] Events: assigned, SLA breach, critical finding, risk accepted, fixed

**✅ Phase 6 Complete когда:** пользователь получает нотификации в UI

---

## PHASE 7: AUDIT LOG API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 7

- [ ] `GET /api/audit-log` с фильтрами
- [ ] `GET /api/audit-log?entity_type=finding&entity_id=X`
- [ ] Pagination (limit/offset)

**✅ Phase 7 Complete когда:** audit log показывает историю действий

---

## PHASE 8: DASHBOARD METRICS API
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 8

- [ ] `GET /api/dashboard`
- [ ] Total products
- [ ] Open findings by severity
- [ ] SLA breached count
- [ ] Recent engagements
- [ ] Top risky products
- [ ] Findings trend (30 days)
- [ ] MTTR by severity
- [ ] Overall security score

**✅ Phase 8 Complete когда:** dashboard API возвращает все метрики

---

## PHASE 9: FRONTEND — AUTH & NAVIGATION
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 9

- [ ] `AuthStore.ts` — добавить `user_id`, `global_role`, `email`
- [ ] `ProtectedRoute.tsx` — HOC по GlobalRole
- [ ] `usePermissions()` hook
- [ ] 401 interceptor → redirect to login
- [ ] Sidebar — роль-зависимое меню
- [ ] Notification bell с badge
- [ ] Breadcrumb навигация

**✅ Phase 9 Complete когда:** роли блокируют доступ к страницам в UI

---

## PHASE 10: FRONTEND — DASHBOARD
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 10

- [ ] KPI Bar (CRITICAL/HIGH/SLA Breached/Score)
- [ ] Findings by Severity chart (30 days)
- [ ] Top Risky Products таблица
- [ ] Recent Engagements (5)
- [ ] MTTR gauge
- [ ] Status Distribution pie
- [ ] Notification Panel

**✅ Phase 10 Complete когда:** Dashboard загружается с реальными данными

---

## PHASE 11: FRONTEND — PRODUCTS
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 11

- [ ] `ProductsPage.tsx` — таблица продуктов
- [ ] `ProductDetailPage.tsx` — 6 вкладок
- [ ] `ProductSettingsModal.tsx`
- [ ] `ProductMembersTab.tsx`

**✅ Phase 11 Complete когда:** можно создать продукт, добавить участников через UI

---

## PHASE 12: FRONTEND — KANBAN BOARD
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 12

- [ ] Установить `@dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities`
- [ ] `KanbanBoard.tsx` — 5 колонок
- [ ] Drag-and-drop с API sync
- [ ] `FindingCard.tsx` — severity, SLA, assignee
- [ ] `FindingDetailModal.tsx` — 5 вкладок
- [ ] Filter bar

**✅ Phase 12 Complete когда:** карточки можно перетаскивать, статус обновляется в DB

---

## PHASE 13: FRONTEND — FINDINGS LIST
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 13

- [ ] `FindingsPage.tsx` с фильтрами
- [ ] Таблица findings
- [ ] Bulk actions
- [ ] Export CSV/JSON
- [ ] `FindingDetailModal` интеграция

**✅ Phase 13 Complete когда:** можно фильтровать, сортировать, экспортировать findings

---

## PHASE 14: FRONTEND — ADMIN PANEL
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 14

- [ ] `AdminPanel.tsx` — routing + tabs
- [ ] `UsersTab.tsx` — CRUD пользователей
- [ ] `AuditLogTab.tsx` — история действий
- [ ] `SystemSettingsTab.tsx` — конфиг системы

**✅ Phase 14 Complete когда:** superadmin может управлять всеми пользователями через UI

---

## PHASE 15: FRONTEND — REPORTS
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 15

- [ ] Executive Summary Report
- [ ] Engagement Report с comparison
- [ ] Export PDF/JSON/CSV

**✅ Phase 15 Complete когда:** можно экспортировать security report

---

## PHASE 16: DOCKER & DEPLOYMENT
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 16

- [ ] `docker-compose.yml` обновлён (web отдельный контейнер)
- [ ] SQLite volume для persistence
- [ ] Health check
- [ ] ENV vars: `JWT_SECRET`, `DB_PATH`, `LLM_API_KEY`
- [ ] `Makefile` targets: `enterprise-up`, `db-migrate`, `create-admin`
- [ ] `.env.example`

**✅ Phase 16 Complete когда:** `docker-compose up` поднимает полный стек

---

## PHASE 17: FINAL AUDIT
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 17

### Security
- [ ] Все admin endpoints защищены middleware
- [ ] Rate limiting на `/api/login`
- [ ] Только prepared statements
- [ ] JWT_SECRET из ENV
- [ ] Content-Security-Policy header

### Testing
- [ ] Login/Logout flow
- [ ] RBAC: viewer не может triagить (проверить вручную)
- [ ] Scan → DB → Kanban E2E тест
- [ ] Deduplication: повторный скан не создаёт дубли
- [ ] SLA breach notification
- [ ] Audit log записывает все действия

### Documentation
- [ ] `README.md` enterprise setup guide
- [ ] API endpoints table
- [ ] RBAC guide
- [ ] Docker deployment guide

**✅ Phase 17 Complete когда:** все security checks пройдены, документация обновлена**

---

## BLOCKED / ISSUES LOG

| Date | Phase | Issue | Resolution |
|---|---|---|---|
| — | — | — | — |

---

## CHANGELOG

| Date | Phase | What was done |
|---|---|---|
| 2026-05-14 | Setup | SQLite dep installed (`modernc.org/sqlite`) |
| 2026-05-14 | Setup | Plan files created: ENTERPRISE_PLAN.md, ENTERPRISE_PLAN_PHASES.md, ENTERPRISE_PLAN_FRONTEND.md |
