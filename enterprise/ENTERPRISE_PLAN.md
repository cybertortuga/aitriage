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
