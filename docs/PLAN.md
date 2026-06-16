# AITriage — Security Presence Checker for Entropy-Based Projects

## Проблема

Вайбкодеры деплоят проекты без базовых security-практик. Не потому что написали плохой код — а потому что **вообще не подумали** о безопасности.

Существующие инструменты (Semgrep, Trivy, Snyk) ищут **конкретные баги**: SQLi в строке 42, CVE в зависимости, секрет в коде. Но никто не отвечает на вопрос: **"какие security-практики в этом проекте отсутствуют?"**

## Решение

**AITriage** — детерминированный presence-checker. Не ищет баги. Проверяет, **есть ли** в проекте ключевые security-практики.

Semgrep находит плохой код. AITriage находит **отсутствие хорошего кода**.

### Формат

- CLI: `aitriage scan .`
- CI/CD: GitHub Action
- Детерминированный: запустил дважды — одинаковый результат
- Без LLM в принятии решений

---

## Presence Checks

13 security-практик. Для каждой один вопрос: **"присутствует ли это в проекте?"**

| # | Практика | Вопрос | Что это НЕ |
|---|----------|--------|-----------|
| P-01 | **Authentication** | Есть ли механизм аутентификации? | Не проверяем корректность реализации auth |
| P-02 | **Authorization** | Есть ли проверка ролей / прав доступа? | Не проверяем RBAC логику |
| P-03 | **Input Validation** | Валидируется ли пользовательский ввод? | Не ищем конкретный SQLi |
| P-04 | **Rate Limiting** | Есть ли ограничение частоты запросов? | Не проверяем пороги |
| P-05 | **Error Handling** | Скрываются ли stack traces от пользователя? | Не проверяем все error paths |
| P-06 | **Security Headers** | Настроены ли CSP, HSTS, X-Frame-Options? | Не валидируем значения |
| P-07 | **CORS Policy** | Явно ли сконфигурирован CORS? | Не проверяем allowed origins |
| P-08 | **Secrets Management** | Секреты в env/vault, а не в коде? | Не сканируем git history (для этого есть gitleaks) |
| P-09 | **Parameterized Queries** | Используется ORM / prepared statements? | Не ищем конкретные SQLi |
| P-10 | **HTTPS Enforcement** | Принудительный ли HTTPS? | Не проверяем TLS конфиг |
| P-11 | **Dependency Lockfile** | Зафиксированы ли версии зависимостей? | Не сканируем CVE (для этого есть Trivy) |
| P-12 | **Logging** | Есть ли security-логирование? | Не проверяем полноту логов |
| P-13 | **CSRF Protection** | Есть ли защита от CSRF? | Не проверяем реализацию |

### Отличие от существующих инструментов

```
Semgrep:  "В строке 42 SQL-запрос собран через конкатенацию строк"
Trivy:   "Зависимость express@4.17.1 имеет CVE-2024-XXXX"
gitleaks: "AWS ключ найден в файле config.js"

AITriage: "В этом проекте ОТСУТСТВУЮТ: аутентификация, rate limiting, input validation"
```

---

## Как определить присутствие практики

Присутствие проверяется через **маркеры** — специфичные для каждого фреймворка файлы, импорты, паттерны конфигурации.

### Пример: P-01 Authentication

**Next.js:**
- Файл `middleware.ts` с auth-логикой
- Импорт `next-auth`, `@clerk/nextjs`, `@auth/core`
- `getServerSession()`, `auth()` вызовы в API routes

**FastAPI:**
- `Depends(get_current_user)` в эндпоинтах
- Импорт `fastapi.security` (OAuth2PasswordBearer, HTTPBearer)
- JWT-related импорты (python-jose, PyJWT)

**Express.js:**
- `passport`, `express-jwt`, `jsonwebtoken` в dependencies
- middleware с `req.user`, `req.isAuthenticated()`

**Логика:** если найден хотя бы один маркер → **PRESENT**. Если ни одного → **ABSENT**.

### Пример: P-04 Rate Limiting

**Next.js:** `@upstash/ratelimit`, `rate-limiter-flexible` в deps, middleware с rate-logic
**FastAPI:** `slowapi` в deps, `Limiter` в коде
**Express.js:** `express-rate-limit` в deps, `rateLimit()` middleware

### Confidence

Не всё определяется однозначно. Rate limiting может быть на уровне nginx/Cloudflare, а не в коде. Поэтому каждый результат имеет confidence:

- **PRESENT** — маркер найден, практика есть
- **ABSENT** — маркеры не найдены (может быть реализовано на уровне инфра — AITriage не видит)
- **UNKNOWN** — стек не поддерживается

---

## Поддерживаемые стеки

Presence checks привязаны к фреймворку. Каждый стек — набор маркеров для каждой из 13 практик.

### Phase 1 (MVP)
- **Next.js** (App Router) — самый массовый у вайбкодеров
- **FastAPI** — второй по популярности для AI-проектов

### Phase 2
- Express.js
- Flask / Django

### Phase 3
- Go (Gin, net/http)
- Rails

### Определение стека
Автоматическое, по маркерам:
- `next.config.*` → Next.js
- `requirements.txt` + `fastapi` → FastAPI
- `package.json` + `express` → Express.js

---

## Scoring

Простая модель: процент присутствующих практик.

```
Score = (PRESENT checks / Total applicable checks) × 100%

A: 85-100%  — security-практики на месте
B: 70-84%   — есть пробелы
C: 50-69%   — серьёзные пробелы
D: 30-49%   — критический минимум
F: 0-29%    — security отсутствует как концепция
```

Без весов в MVP. Если нужно — добавим потом (auth важнее чем lockfile).

---

## CLI

```bash
# Установка
npm install -g aitriage

# Скан проекта
aitriage scan .

# Конкретный стек (если автоопределение не сработало)
aitriage scan . --stack nextjs

# JSON для CI
aitriage scan . --format json

# SARIF для GitHub Security tab
aitriage scan . --format sarif
```

### Пример вывода

```
🔍 AITriage v0.1.0 — Security Presence Check

📁 Project: ./my-entropy-app
🔧 Detected stack: Next.js (App Router)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

 🔴 ABSENT   Authentication
              No auth library or middleware detected.
              → Consider: NextAuth.js, Clerk, or custom middleware.ts

 🔴 ABSENT   Rate Limiting
              No rate limiting middleware or library detected.
              → Consider: @upstash/ratelimit, rate-limiter-flexible

 🔴 ABSENT   Input Validation
              No validation library detected in API routes.
              → Consider: Zod, Yup, or joi for request validation

 ✅ PRESENT  Parameterized Queries
              Prisma ORM detected (package.json)

 ✅ PRESENT  HTTPS Enforcement
              Next.js on Vercel enforces HTTPS by default

 ✅ PRESENT  Dependency Lockfile
              package-lock.json found

 🟡 ABSENT   CORS Policy
              No explicit CORS configuration in next.config.
              ⚠ May be handled by hosting platform (Vercel, etc.)

 🟡 ABSENT   Security Headers
              No security headers in next.config.js
              → Consider: headers() in next.config.js

 ✅ PRESENT  Error Handling
              Custom error.tsx found in app/

 ✅ PRESENT  Secrets Management
              .env in .gitignore, env vars used via process.env

 🔴 ABSENT   CSRF Protection
              No CSRF middleware detected.

 ✅ PRESENT  Logging
              winston detected in package.json

 🔴 ABSENT   Authorization
              No role/permission checks detected in routes.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Score: 46% — Grade D
 6/13 practices present · 4 critical absent
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## CI/CD — GitHub Action

```yaml
name: AITriage
on: [push, pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aitriage/action@v1
        with:
          fail-on-grade: D   # Fail PR если grade D или F
```

---

## Техстек

| Компонент | Технология | Почему |
|-----------|-----------|--------|
| CLI | Node.js / TypeScript | npm ecosystem, целевая аудитория там же |
| Парсинг | regex + file checks | Presence checks не требуют AST — ищем импорты, файлы, конфиги |
| GitHub Action | TypeScript action | Нативная интеграция |
| Тесты | vitest | Каждый маркер = unit test |

**Примечание:** для MVP не нужен Tree-sitter / AST. Presence checks работают на уровне: "есть ли import X?", "есть ли файл Y?", "есть ли пакет Z в dependencies?". Это grep, не парсинг.

---

## Roadmap

### Месяц 1-2: MVP

- [ ] CLI скелет (TypeScript, Commander.js)
- [ ] Авто-определение стека
- [ ] 13 presence checks для Next.js
- [ ] 13 presence checks для FastAPI
- [ ] Scoring + terminal output
- [ ] JSON output
- [ ] Тесты на сгенерированных LLM проектах

### Месяц 3: CI/CD

- [ ] GitHub Action
- [ ] SARIF output → GitHub Security tab
- [ ] PR комментарии с результатами
- [ ] README badge

### Месяц 4: Расширение стеков

- [ ] Express.js markers
- [ ] Django markers
- [ ] Улучшение маркеров по результатам тестов на реальных проектах

### Месяц 5-6: Polish

- [ ] npm publish
- [ ] GitHub Marketplace
- [ ] Документация по добавлению своих маркеров
- [ ] (Опционально) интеграция Trivy/gitleaks как дополнительных проверок через `--with-trivy`, `--with-gitleaks`

---

## Чем это НЕ является

- NOT SAST — не ищет баги в конкретных строках (для этого Semgrep)
- NOT SCA — не сканирует CVE в зависимостях (для этого Trivy)
- NOT Secret scanner — не ищет утечки ключей (для этого gitleaks)
- NOT LLM-powered — детерминированный
- NOT Pentest — не атакует приложение

**AITriage отвечает на один вопрос: "Какие security-практики в этом проекте отсутствуют?"**
