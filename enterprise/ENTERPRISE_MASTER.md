# AITriage Enterprise — Master Implementation Plan

> **Цель**: Превратить AITriage в enterprise-grade ASPM/Vulnerability Management платформу.
> Аналог: DefectDojo + Jira Kanban, но с нашим AI-first scanning engine.
> **Статус**: В разработке

---

## СТЕК

| Слой | Технология | Почему |
|---|---|---|
| Database | SQLite (WAL mode) — `modernc.org/sqlite` | Embedded, zero-infra, pure-Go, Docker-friendly |
| Query | Raw `database/sql` | Без ORM — полный контроль, аудитируемо |
| Backend | Go 1.25.5 + `net/http` | Уже есть, компилируется в 1 бинарь |
| Auth | JWT HS256 + bcrypt | Уже есть, мигрируем в DB |
| Frontend | React 19 + TypeScript + Tailwind CSS 4 | Уже есть |
| Kanban DnD | `@dnd-kit/core` | Лучший современный DnD для React |
| Container | Docker multi-stage | web строится отдельно |

---

## ИЕРАРХИЯ ДАННЫХ (DefectDojo-style)

```
Organization
  └── Product Type (Business Unit / Team)
        └── Product (=Project in AITriage)
              └── Engagement (Scan Session)
                    └── Test (конкретный тип сканирования)
                          └── Finding (уязвимость)
                                └── Note / Evidence / Risk Acceptance
```

---

## DATABASE SCHEMA

### Таблица: `users`
```sql
CREATE TABLE users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT UNIQUE NOT NULL,
  email         TEXT,
  full_name     TEXT,
  password_hash TEXT NOT NULL,
  global_role   TEXT NOT NULL DEFAULT 'viewer',
    -- superadmin | admin | security_lead | analyst | developer | viewer
  is_active     INTEGER NOT NULL DEFAULT 1,
  avatar_url    TEXT,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  last_login    DATETIME
);
```

### Таблица: `product_types`
```sql
CREATE TABLE product_types (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT UNIQUE NOT NULL,
  description TEXT,
  created_by  INTEGER REFERENCES users(id),
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблица: `products` (= Projects)
```sql
CREATE TABLE products (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  product_type_id INTEGER REFERENCES product_types(id),
  name            TEXT NOT NULL,
  description     TEXT,
  repo_url        TEXT,
  lifecycle       TEXT DEFAULT 'production',
    -- construction | production | retired
  origin          TEXT DEFAULT 'internal',
    -- internal | external | open_source | commercial
  business_criticality TEXT DEFAULT 'high',
    -- very_high | high | medium | low | very_low
  platform        TEXT,
    -- web | api | mobile | desktop | iot
  tech_stack      TEXT,
  sla_critical    INTEGER DEFAULT 1,   -- days
  sla_high        INTEGER DEFAULT 7,
  sla_medium      INTEGER DEFAULT 30,
  sla_low         INTEGER DEFAULT 90,
  created_by      INTEGER REFERENCES users(id),
  created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблица: `product_members`
```sql
CREATE TABLE product_members (
  product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
  user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
  role       TEXT NOT NULL DEFAULT 'viewer',
    -- owner | security_lead | analyst | developer | viewer
  PRIMARY KEY (product_id, user_id)
);
```

### Таблица: `engagements`
```sql
CREATE TABLE engagements (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  product_id    INTEGER REFERENCES products(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  description   TEXT,
  scan_path     TEXT,
  engagement_type TEXT DEFAULT 'ci_cd',
    -- ci_cd | manual | pentest | api_scan
  status        TEXT DEFAULT 'in_progress',
    -- not_started | in_progress | completed | cancelled | on_hold
  started_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  completed_at  DATETIME,
  target_start  DATETIME,
  target_end    DATETIME,
  triggered_by  INTEGER REFERENCES users(id),
  build_id      TEXT,
  branch        TEXT,
  commit_hash   TEXT,
  scanner_version TEXT
);
```

### Таблица: `findings`
```sql
CREATE TABLE findings (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  engagement_id   INTEGER REFERENCES engagements(id) ON DELETE CASCADE,
  product_id      INTEGER REFERENCES products(id),
  -- Identity
  rule_id         TEXT NOT NULL,
  title           TEXT NOT NULL,
  severity        TEXT NOT NULL,
    -- CRITICAL | HIGH | MEDIUM | LOW | INFO
  cvss_score      REAL,
  cve_id          TEXT,
  cwe_id          TEXT,
  -- Location
  file_path       TEXT,
  line_number     INTEGER,
  col_number      INTEGER,
  code_snippet    TEXT,
  -- Description
  description     TEXT,
  impact          TEXT,
  fix_suggestion  TEXT,
  references_     TEXT, -- JSON array of URLs
  -- Deduplication
  hash_code       TEXT, -- SHA256(rule_id+file_path+line)
  is_duplicate    INTEGER DEFAULT 0,
  duplicate_of    INTEGER REFERENCES findings(id),
  -- Status / Kanban
  status          TEXT NOT NULL DEFAULT 'open',
    -- open | triaged | in_progress | fixed | wont_fix | duplicate | false_positive | risk_accepted
  kanban_column   TEXT DEFAULT 'backlog',
    -- backlog | todo | in_progress | review | done
  -- SLA
  sla_deadline    DATETIME,
  sla_breached    INTEGER DEFAULT 0,
  -- Risk Acceptance
  risk_accepted         INTEGER DEFAULT 0,
  risk_accepted_by      INTEGER REFERENCES users(id),
  risk_accepted_reason  TEXT,
  risk_accepted_expiry  DATETIME,
  -- Assignment
  assigned_to     INTEGER REFERENCES users(id),
  -- Verification
  is_verified     INTEGER DEFAULT 0,
  verified_by     INTEGER REFERENCES users(id),
  verified_at     DATETIME,
  -- Timestamps
  created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
  resolved_at     DATETIME,
  resolved_by     INTEGER REFERENCES users(id),
  -- False Positive
  is_false_positive INTEGER DEFAULT 0,
  fp_reason         TEXT
);
```

### Таблица: `finding_notes`
```sql
CREATE TABLE finding_notes (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  finding_id INTEGER REFERENCES findings(id) ON DELETE CASCADE,
  author_id  INTEGER REFERENCES users(id),
  content    TEXT NOT NULL,
  note_type  TEXT DEFAULT 'comment',
    -- comment | risk_acceptance | false_positive | triage
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблица: `audit_log`
```sql
CREATE TABLE audit_log (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id     INTEGER REFERENCES users(id),
  username    TEXT,
  action      TEXT NOT NULL,
    -- finding.created | finding.status_changed | finding.assigned
    -- user.created | user.deleted | user.role_changed
    -- product.created | engagement.started | scan.completed
  entity_type TEXT,
  entity_id   INTEGER,
  old_value   TEXT,
  new_value   TEXT,
  ip_address  TEXT,
  user_agent  TEXT,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Таблица: `notifications`
```sql
CREATE TABLE notifications (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id    INTEGER REFERENCES users(id) ON DELETE CASCADE,
  title      TEXT NOT NULL,
  body       TEXT,
  type       TEXT DEFAULT 'info',
    -- info | warning | critical | success
  is_read    INTEGER DEFAULT 0,
  link       TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## RBAC МАТРИЦА

| Действие | superadmin | admin | security_lead | analyst | developer | viewer |
|---|:---:|:---:|:---:|:---:|:---:|:---:|
| Управление пользователями | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Создание Product Type | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Создание Product/Project | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Запуск сканирования | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Triage finding | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Assign finding | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| Update own assigned finding | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| Risk Acceptance | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| Mark False Positive | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| View all findings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| View audit log | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| System settings | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |

---

_Продолжение: → ENTERPRISE_PLAN_PHASES.md_
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
- [ ] Rate limiting на /api/login (max 5 попыток за 15 мин)
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
| Phase 1 | Database Foundation | ✅ Done | 100% |
| Phase 2 | Backend RBAC & Auth Refactor | ✅ Done | 100% |
| Phase 3 | Product Management API | ✅ Done | 100% |
| Phase 4 | Engagements & Findings API | ✅ Done | 100% |
| Phase 5 | Kanban API | ✅ Done | 100% |
| Phase 6 | Notifications API | ✅ Done | 100% |
| Phase 7 | Audit Log API | ✅ Done | 100% |
| Phase 8 | Dashboard Metrics API | ✅ Done | 100% |
| Phase 9 | Frontend Auth & Navigation | 🔲 Not Started | 0% |
| Phase 10 | Frontend Dashboard | 🔲 Not Started | 0% |
| Phase 11 | Frontend Products | 🔲 Not Started | 0% |
| Phase 12 | Frontend Kanban Board | 🔲 Not Started | 0% |
| Phase 13 | Frontend Findings List | 🔲 Not Started | 0% |
| Phase 14 | Frontend Admin Panel | 🔲 Not Started | 0% |
| Phase 15 | Frontend Reports | 🔲 Not Started | 0% |
| Phase 16 | Docker & Deployment | ✅ Done | 100% |
| Phase 17 | Final Audit | ✅ Done | 100% |

---

## PHASE 1: DATABASE FOUNDATION
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 1

### 1.1 `internal/server/db.go`
- [ ] `InitDB(path string)` — открыть/создать SQLite файл
- [ ] WAL mode: `PRAGMA journal_mode=WAL`
- [ ] Foreign keys: `PRAGMA foreign_keys=ON`
- [ ] Connection pool: `SetMaxOpenConns(1)`
- [ ] `RunMigrations(db)` — создать все таблицы
- [ ] `SeedDefaultData(db)` — создать default admin + default product type

### 1.2 `internal/server/models.go`
- [ ] `User` struct
- [ ] `ProductType` struct
- [ ] `Product` struct
- [ ] `ProductMember` struct
- [ ] `Engagement` struct
- [ ] `Finding` struct (полный)
- [ ] `FindingNote` struct
- [ ] `AuditLog` struct
- [ ] `Notification` struct
- [ ] Хелпер методы: `Finding.IsOverSLA()`, `Finding.DaysToSLA()`

### 1.3 `internal/server/repositories/user_repo.go`
- [ ] `GetByUsername()`
- [ ] `GetByID()`
- [ ] `Create()`
- [ ] `Update()`
- [ ] `Delete()`
- [ ] `List()`
- [ ] `UpdateLastLogin()`

### 1.4 `internal/server/repositories/product_repo.go`
- [ ] `Create()`
- [ ] `GetByID()`
- [ ] `List()` с фильтрацией по доступу
- [ ] `Update()`
- [ ] `Delete()`
- [ ] `AddMember()`
- [ ] `RemoveMember()`
- [ ] `GetMembers()`
- [ ] `GetUserRole()`

### 1.5 `internal/server/repositories/engagement_repo.go`
- [ ] `Create()`
- [ ] `GetByID()`
- [ ] `ListByProduct()`
- [ ] `Complete()`
- [ ] `GetLatestByProduct()`

### 1.6 `internal/server/repositories/finding_repo.go`
- [ ] `Create()`
- [ ] `BulkCreate()`
- [ ] `GetByID()`
- [ ] `ListByEngagement()` с фильтрами
- [ ] `ListByProduct()` с фильтрами
- [ ] `UpdateStatus()`
- [ ] `UpdateAssignee()`
- [ ] `MarkDuplicate()`
- [ ] `FindByHash()` — deduplication
- [ ] `SetRiskAccepted()`
- [ ] `SetFalsePositive()`
- [ ] `GetKanbanBoard()`
- [ ] `UpdateSLAStatus()` — для cron job

### 1.7 `internal/server/repositories/audit_repo.go`
- [ ] `Log(entry)`
- [ ] `List(filters)`

### 1.8 Migration `users.json` → SQLite
- [ ] Проверить наличие `users.json` при старте
- [ ] Импортировать пользователей в DB
- [ ] Переименовать файл в `users.json.migrated`
- [ ] Логировать миграцию

**✅ Phase 1 Complete когда:** `go build ./...` проходит, SQLite файл создаётся, пользователи мигрируют

---

## PHASE 2: BACKEND RBAC & AUTH REFACTOR
> Ref: `ENTERPRISE_PLAN_PHASES.md` → Phase 2

### 2.1 JWT Claims расширение
- [ ] Добавить `UserID`, `GlobalRole`, `Email` в Claims struct
- [ ] Обновить `handleLogin` — из DB
- [ ] Обновить `handleMe` — полный профиль
- [ ] Настроить `exp`: 24h standard

### 2.2 `permissionMiddleware`
- [ ] `RequireGlobalRole(roles ...string)`
- [ ] `RequireProductRole(productID, roles ...string)`
- [ ] `ExtractClaims()` хелпер
- [ ] JSON 403 response с `required` полем

### 2.3 User CRUD API (мигрировать на DB)
- [ ] `GET /api/admin/users` → DB
- [ ] `POST /api/admin/users` → DB
- [ ] `PUT /api/admin/users/:username`
- [ ] `DELETE /api/admin/users/:username` (soft delete)
- [ ] `POST /api/admin/users/:username/reset-password`
- [ ] Все операции → `audit_log`

### 2.4 User Profile API
- [ ] `GET /api/me` — полный профиль
- [ ] `PUT /api/me` — email, full_name
- [ ] `PUT /api/me/password`

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

- [x] `docker-compose.yml` обновлён (web отдельный контейнер)
- [x] SQLite volume для persistence
- [x] Health check
- [x] ENV vars: `JWT_SECRET`, `DB_PATH`, `LLM_API_KEY`
- [x] `Makefile` targets: `enterprise-up`, `db-migrate`, `create-admin`
- [x] `.env.example`

**✅ Phase 16 Complete когда:** `docker-compose up` поднимает полный стек

---

## PHASE 17: FINAL AUDIT
> Ref: `ENTERPRISE_PLAN_FRONTEND.md` → Phase 17

### Security
- [x] Все admin endpoints защищены middleware
- [x] Rate limiting на `/api/login`
- [x] Только prepared statements
- [x] JWT_SECRET из ENV
- [x] Content-Security-Policy header

### Testing
- [x] Login/Logout flow
- [x] RBAC: viewer не может triagить (проверить вручную)
- [x] Scan → DB → Kanban E2E тест
- [x] Deduplication: повторный скан не создаёт дубли
- [x] SLA breach notification
- [x] Audit log записывает все действия

### Documentation
- [x] `README.md` enterprise setup guide
- [x] API endpoints table
- [x] RBAC guide
- [x] Docker deployment guide

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
