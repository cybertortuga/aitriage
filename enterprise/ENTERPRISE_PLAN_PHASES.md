# AITriage Enterprise — Phases & Tasks

---

## PHASE 1: DATABASE FOUNDATION
**Цель**: SQLite schema + migration layer + репозитории

### 1.1 `internal/server/db.go` — Инициализация БД
- [ ] `InitDB(path string) (*sql.DB, error)` — открыть/создать SQLite файл
- [ ] Включить WAL mode: `PRAGMA journal_mode=WAL`
- [ ] Включить foreign keys: `PRAGMA foreign_keys=ON`
- [ ] Установить connection pool: `SetMaxOpenConns(1)` (SQLite single-writer)
- [ ] `RunMigrations(db *sql.DB)` — создать все таблицы если не существуют
- [ ] `SeedDefaultData(db *sql.DB)` — создать default admin, default product type
- [ ] Перенести создание default admin из `auth.go` сюда

### 1.2 `internal/server/models.go` — Go structs
- [ ] `User` struct с полями из DB schema
- [ ] `ProductType` struct
- [ ] `Product` struct
- [ ] `ProductMember` struct
- [ ] `Engagement` struct
- [ ] `Finding` struct (полный, включая SLA, dedup, risk acceptance)
- [ ] `FindingNote` struct
- [ ] `AuditLog` struct
- [ ] `Notification` struct
- [ ] JSON теги для всех полей
- [ ] Хелпер методы: `Finding.IsOverSLA()`, `Finding.DaysToSLA()`

### 1.3 `internal/server/repositories/user_repo.go`
- [ ] `GetByUsername(username string) (*User, error)`
- [ ] `GetByID(id int64) (*User, error)`
- [ ] `Create(u *User) (int64, error)`
- [ ] `Update(u *User) error`
- [ ] `Delete(id int64) error`
- [ ] `List(limit, offset int) ([]User, error)`
- [ ] `UpdateLastLogin(id int64) error`

### 1.4 `internal/server/repositories/product_repo.go`
- [ ] `Create(p *Product) (int64, error)`
- [ ] `GetByID(id int64) (*Product, error)`
- [ ] `List(userID int64, role string) ([]Product, error)` — фильтрация по доступу
- [ ] `Update(p *Product) error`
- [ ] `Delete(id int64) error`
- [ ] `AddMember(productID, userID int64, role string) error`
- [ ] `RemoveMember(productID, userID int64) error`
- [ ] `GetMembers(productID int64) ([]ProductMember, error)`
- [ ] `GetUserRole(productID, userID int64) (string, error)`

### 1.5 `internal/server/repositories/engagement_repo.go`
- [ ] `Create(e *Engagement) (int64, error)`
- [ ] `GetByID(id int64) (*Engagement, error)`
- [ ] `ListByProduct(productID int64) ([]Engagement, error)`
- [ ] `Complete(id int64) error`
- [ ] `GetLatestByProduct(productID int64) (*Engagement, error)`

### 1.6 `internal/server/repositories/finding_repo.go`
- [ ] `Create(f *Finding) (int64, error)`
- [ ] `BulkCreate(findings []Finding) error` — для batch после scan
- [ ] `GetByID(id int64) (*Finding, error)`
- [ ] `ListByEngagement(engagementID int64, filters FindingFilters) ([]Finding, error)`
- [ ] `ListByProduct(productID int64, filters FindingFilters) ([]Finding, error)`
- [ ] `UpdateStatus(id int64, status, column string, updatedBy int64) error`
- [ ] `UpdateAssignee(id int64, assigneeID int64) error`
- [ ] `MarkDuplicate(id, duplicateOfID int64) error`
- [ ] `FindByHash(hash string, productID int64) (*Finding, error)` — deduplication
- [ ] `SetRiskAccepted(id int64, reason string, expiry *time.Time, byUser int64) error`
- [ ] `SetFalsePositive(id int64, reason string, byUser int64) error`
- [ ] `GetKanbanBoard(productID int64) (map[string][]Finding, error)`
- [ ] `UpdateSLAStatus(db *sql.DB) error` — cron job для пометки breached
- [ ] `FindingFilters` struct: severity, status, assignee, sla_breached, date_range

### 1.7 `internal/server/repositories/audit_repo.go`
- [ ] `Log(entry *AuditLog) error`
- [ ] `List(filters AuditFilters) ([]AuditLog, error)`

### 1.8 Migration от `users.json` → SQLite
- [ ] При старте: проверить есть ли `users.json`
- [ ] Если есть → импортировать пользователей в DB
- [ ] Переименовать `users.json` → `users.json.migrated`
- [ ] Логировать миграцию

---

## PHASE 2: BACKEND RBAC & AUTH REFACTOR

### 2.1 Расширение JWT Claims
- [ ] Добавить в Claims: `UserID int64`, `GlobalRole string`, `Email string`
- [ ] Обновить `handleLogin` — достать user из DB вместо map
- [ ] Обновить `handleMe` — вернуть полный профиль с ролью
- [ ] Добавить `exp` claim — 24h для обычных, 1h для admin сессий

### 2.2 `permissionMiddleware`
- [ ] `RequireGlobalRole(roles ...string) http.HandlerFunc` — проверить global_role из JWT
- [ ] `RequireProductRole(productID string, roles ...string)` — проверить через product_members
- [ ] `ExtractClaims(r *http.Request) (*Claims, error)` — хелпер
- [ ] Middleware поддерживает AND/OR логику ролей
- [ ] При 403 — возвращать JSON `{"error": "insufficient_permissions", "required": "..."}`

### 2.3 Обновление User CRUD API
- [ ] `GET /api/admin/users` — список всех юзеров из DB (только superadmin/admin)
- [ ] `POST /api/admin/users` — создать с хешированным паролем, записать в DB
- [ ] `PUT /api/admin/users/:username` — обновить роль, email, активность
- [ ] `DELETE /api/admin/users/:username` — мягкое удаление (is_active=0)
- [ ] `POST /api/admin/users/:username/reset-password` — сменить пароль
- [ ] Все операции пишут в `audit_log`

### 2.4 User Profile API
- [ ] `GET /api/me` — полный профиль (username, email, role, продукты)
- [ ] `PUT /api/me` — обновить email, full_name
- [ ] `PUT /api/me/password` — смена пароля (нужен старый пароль)

---

## PHASE 3: PRODUCT MANAGEMENT API (DefectDojo-style)

### 3.1 Product Types
- [ ] `GET /api/product-types` — список
- [ ] `POST /api/product-types` — создать (admin+)
- [ ] `PUT /api/product-types/:id` — редактировать
- [ ] `DELETE /api/product-types/:id` — удалить

### 3.2 Products (Projects)
- [ ] `GET /api/products` — список доступных текущему юзеру продуктов
- [ ] `GET /api/products/:id` — детальная инфа + статистика
- [ ] `POST /api/products` — создать (security_lead+)
- [ ] `PUT /api/products/:id` — редактировать (owner+)
- [ ] `DELETE /api/products/:id` — архивировать (admin+)

### 3.3 Product Members
- [ ] `GET /api/products/:id/members` — список участников
- [ ] `POST /api/products/:id/members` — добавить участника `{user_id, role}`
- [ ] `PUT /api/products/:id/members/:userID` — сменить роль
- [ ] `DELETE /api/products/:id/members/:userID` — убрать из проекта

### 3.4 Product Metrics
- [ ] `GET /api/products/:id/metrics` — агрегированные данные:
  - Findings по severity (CRITICAL/HIGH/MEDIUM/LOW/INFO)
  - Findings по status (open/in_progress/fixed/etc)
  - SLA compliance % (сколько в SLA vs сколько нарушили)
  - MTTR (Mean Time To Remediate)
  - Security Score (производная от open critical/high)
  - Trend за 30/60/90 дней

### 3.5 SLA Configuration
- [ ] `PUT /api/products/:id/sla` — настроить SLA дни по severity
- [ ] Background task каждые 1h: проверять просроченные findings, ставить `sla_breached=1`
- [ ] При нарушении SLA → создавать Notification для security_lead

---

## PHASE 4: ENGAGEMENTS & FINDINGS API

### 4.1 Engagements
- [ ] `GET /api/products/:id/engagements` — список сканирований
- [ ] `GET /api/engagements/:id` — детали + все findings
- [ ] `POST /api/products/:id/engagements` — создать вручную
- [ ] `PUT /api/engagements/:id` — обновить статус
- [ ] `POST /api/engagements/:id/complete` — завершить engagement

### 4.2 Findings CRUD
- [ ] `GET /api/findings` — список с фильтрами (product, engagement, severity, status, assignee, sla_breached)
- [ ] `GET /api/findings/:id` — полный finding с notes, history
- [ ] `PUT /api/findings/:id` — обновить (status, kanban_column, assigned_to)
- [ ] `POST /api/findings/:id/notes` — добавить комментарий
- [ ] `GET /api/findings/:id/notes` — история заметок
- [ ] `POST /api/findings/:id/risk-accept` — принять риск `{reason, expiry_date}`
- [ ] `POST /api/findings/:id/false-positive` — пометить FP `{reason}`
- [ ] `POST /api/findings/:id/verify` — верифицировать finding
- [ ] `POST /api/findings/:id/duplicate` — пометить дублем `{duplicate_of_id}`

### 4.3 Deduplication Engine
- [ ] При создании finding — вычислить SHA256 hash: `rule_id + file_path + line_number`
- [ ] `FindByHash()` — поискать существующий в том же product
- [ ] Если найден и статус != fixed → пометить новый как duplicate
- [ ] Если найден и статус == fixed → "regression", открыть заново
- [ ] Логировать dedup decision в audit_log

### 4.4 Scan → DB Integration (критично!)
- [ ] После `/api/scan` → автоматически создать Engagement
- [ ] Парсить findings из `RichScanResult` → записать в DB через `BulkCreate`
- [ ] Применить deduplication при bulk insert
- [ ] Вычислить SLA deadline для каждого finding (created_at + sla_days)
- [ ] Вернуть в ответе `engagement_id` чтобы frontend мог открыть результаты
- [ ] Обновить `/api/triage` — работать с DB findings, не с in-memory lastResult

---

## PHASE 5: KANBAN API

### 5.1 Kanban Board API
- [ ] `GET /api/kanban?product_id=X` — вернуть findings сгруппированные по `kanban_column`:
  ```json
  {
    "backlog": [...],
    "todo": [...],
    "in_progress": [...],
    "review": [...],
    "done": [...]
  }
  ```
- [ ] `PATCH /api/findings/:id/move` — переместить в другую колонку `{column, status}`
- [ ] `PATCH /api/findings/:id/assign` — назначить `{user_id}`
- [ ] Audit log при каждом движении
- [ ] Notification назначенному пользователю

### 5.2 Kanban Column Rules (автоматические переходы)
- [ ] `open` → Backlog (при создании)
- [ ] `triaged` → Todo (при triage)
- [ ] `in_progress` → In Progress (при assign)
- [ ] `fixed` → Done
- [ ] `wont_fix` → Done
- [ ] `false_positive` → Done (но с маркером)
- [ ] Manual override всегда возможен

---

## PHASE 6: NOTIFICATIONS API

- [ ] `GET /api/notifications` — список для текущего юзера
- [ ] `PUT /api/notifications/:id/read` — пометить прочитанным
- [ ] `PUT /api/notifications/read-all` — все прочитать
- [ ] `DELETE /api/notifications/:id` — удалить
- [ ] Events для создания нотификаций:
  - Finding assigned → notify assignee
  - SLA breached → notify security_lead + owner
  - New critical finding → notify security_lead
  - Risk acceptance approved → notify reporter
  - Finding fixed → notify reporter

---

## PHASE 7: AUDIT LOG API

- [ ] `GET /api/audit-log` — список с фильтрами (action, user, entity_type, date_range)
- [ ] `GET /api/audit-log?entity_type=finding&entity_id=123` — history конкретного finding
- [ ] Pagination: limit/offset
- [ ] Иммутабельный лог (только INSERT, никогда UPDATE/DELETE)

---

## PHASE 8: DASHBOARD METRICS API

- [ ] `GET /api/dashboard` — глобальные метрики:
  - Total products
  - Open findings by severity
  - SLA breached count
  - Recent engagements (5)
  - Top risky products (by open critical)
  - Findings trend (last 30 days, daily buckets)
  - MTTR по severity
  - Security score overall

---

_Продолжение: → ENTERPRISE_PLAN_FRONTEND.md_
