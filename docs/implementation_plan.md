# План реализации полного MVP (Все проверки + CI/CD)

Запрос "сделать все сразу" подразумевает закрытие оставшихся задач фаз MVP (Месяцы 1-3 в `PLAN.md`). 
Мы добавим все недостающие Presence Checks (13 маркеров) для Next.js и FastAPI, а также создадим нативную интеграцию с GitHub Actions.

## User Review Required

> [!CAUTION]  
> "Сделать всё сразу" — это огромный пул изменений. Я добавлю десятки новых проверок в `default_rules.yaml` и создам файлы для GitHub Action. Пожалуйста, подтверди этот план, прежде чем я начну генерировать код.
> Мы будем использовать подход Presence Checks ("присутствует ли библиотека/файл", а не классический поиск CVE/SQLi).

## Proposed Changes

### Правила Сканирования (Presence Checks)

Мы дополним `internal/engine/default_rules.yaml` недостающими правилами для покрытия всех 13 секьюрити-доменов. Я буду использовать поиск по `package.json` / `requirements.txt` / файлов конфигураций, как и задумано концепцией "Умного Grep'а".

#### [MODIFY] `internal/engine/default_rules.yaml`
- **P-01 Authentication:** (Уже есть NEXT-AUTH). Добавлю проверку импортов OAuth/JWT для FastAPI (`fastapi.security`, `jose`).
- **P-02 Authorization:** Добавлю проверку наличия логики ролей в middleware/guards (паттерны `Role`, `is_admin`, `Depends(get_current_active_user)`).
- **P-03 Input Validation:** Проверка библиотек (Next.js: `zod`, `yup`, `joi`. FastAPI: `pydantic` автоматически закрывает это, проверим наличие Pydantic схем).
- **P-04 Rate Limiting:** (Next.js: `@upstash/ratelimit`, `express-rate-limit`. FastAPI: `slowapi`).
- **P-05 Error Handling:** Наличие кастомных обработчиков ошибок (Next.js: файл `error.tsx`. FastAPI: `@app.exception_handler`).
- **P-06 Security Headers:** Проверка конфигурации заголовков (Next.js: `headers()` в `next.config.*`. FastAPI: `SecureHeaders` / middleware).
- **P-07 CORS Policy:** (Уже есть NEXT-CORS). Добавлю проверку использования `CORSMiddleware` для FastAPI.
- **P-08 Secrets Management:** Проверка загрузки envvars через `.env` (Next.js: не хранятся ли открыто, FastAPI: использование `pydantic-settings`).
- **P-09 Parameterized Queries:** Докажем использование ORM (Next.js: `prisma`, `drizzle`, `typeorm`. FastAPI: `sqlalchemy`, `sqlmodel`, `tortoise-orm`).
- **P-10 HTTPS Enforcement:** Будем считать "Условно выполнено" для Next.js (если Vercel/Cloudflare) или проверять наличие RedirectMiddleware. 
- **P-11 Dependency Lockfile:** Уже реализовано (`ENTR-02`).
- **P-12 Logging:** Наличие логгеров помимо `console.log` (Next.js: `winston`, `pino`. FastAPI: `loguru`, модуль `logging`).
- **P-13 CSRF Protection:** (Уже есть NEXT-CSRF). Добавлю проверку защиты для FastAPI (например, библ. `fastapi-csrf-protect`).

### Интеграция с GitHub 

Для выполнения задачи "CI/CD" мы упакуем Go-бинарник в Docker-based GitHub Action, чтобы юзеры могли писать: `uses: [твоя-репа]/@v1`.

#### [NEW] `action.yml`
Создание метаданных GitHub Action, определяющих входы, выходы и вызов Docker image.

#### [NEW] `Dockerfile`
Контейнеризация `aitriage`, компиляция Go и запуск проверок в окружении GitHub Actions `alpine`. 

#### [NEW] `.github/workflows/test-action.yml`
Workflow для самотестирования Action прямо в этом репозитории на каждый Push.

## Verification Plan

### Automated Tests
1. Прогнать билд бинарника `go build ./...`.
2. Запустить созданный AITriage-контейнер локально, чтобы проверить, что Docker корректно собирается.
3. Проверить `aitriage scan .` на собственной кодовой базе, чтобы убедиться, что YAML-правила загружаются без синтаксических ошибок.

### Manual Verification
После коммита Action заработает. Ты сможешь подключить его к любому своему пулл-реквисту, добавив 1 строчку в workflow и проверив SARIF репорты во вкладке **Security** на GitHub.
