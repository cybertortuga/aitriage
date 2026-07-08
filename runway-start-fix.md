# runway-start-fix

## Обзор
Фиксим запуск SecureCoder Agent Runway из simple-интерфейса. Сейчас POST на backend уходит, но UI остается на шаге 0 и не показывает понятный прогресс или ошибку, из-за чего кнопка выглядит ненажимаемой.

## Затронутые файлы и области
- `web/src/pages/SimpleDashboardPage.tsx` — состояние Runway, polling, кнопка запуска.
- `web/src/locales/en/pages.json` — текст кнопки/статуса Runway.
- `web/src/locales/ru/pages.json` — русский текст кнопки/статуса Runway.
- `internal/server/server.go` — старт backend Runway-сессии и начальный статус.

## Задача 1: Исправить состояние запуска Runway
### Описание
После успешного старта backend-пайплайна UI должен переходить в состояние выполнения, возобновлять polling и не плодить повторные запуски по клику.
### Затронутые файлы
- `web/src/pages/SimpleDashboardPage.tsx`
- `web/src/locales/en/pages.json`
- `web/src/locales/ru/pages.json`
- `internal/server/server.go`
### Подзадачи
- [x] Обновить frontend: после успешного `/api/runway/start/:id` переводить Runway на шаг выполнения и не блокировать polling.
- [x] Обновить frontend: показывать loading/disabled-состояние кнопки запуска и восстанавливать ошибки из session.
- [x] Обновить backend: при старте выставлять `running/current_step=1`, очищать старую ошибку и не запускать повторно уже running-сессию.
### Критерии приёмки
- [x] Клик по `START AUTOMATED AUDIT` меняет UI на состояние запуска/выполнения.
- [x] Повторный клик не создает параллельный запуск той же running-сессии.
- [x] Ошибка backend-пайплайна видна в интерфейсе, а не выглядит как мертвый клик.
### Риски
- Реальный LLM-пайплайн может упасть по внешней квоте 429; это не чинится UI-правкой, но ошибка должна стать видимой.

## Задача 2: Проверить в локальном веб-режиме
### Описание
Пересобрать/перезапустить web-контейнер и проверить запуск Runway на `aitriage-test-vulnerabilities`.
### Затронутые файлы
- Docker/web runtime без изменения дополнительных исходников.
### Подзадачи
- [x] Запустить форматирование/сборку или релевантные проверки.
- [x] Перезапустить локальный web-режим на `http://localhost:8080/`.
- [x] Проверить через API/browser, что кнопка работает и статус/ошибка отображаются.
### Критерии приёмки
- [x] Локальный `http://localhost:8080/` отвечает после перезапуска.
- [x] Runway session после клика имеет `status=running` или `failed`, но не зависает визуально на шаге 0 без сообщения.
### Риски
- Если в БД осталась старая broken-сессия со статусом `running`, может понадобиться удалить только сгенерированные тестовые Runway-сессии.

## Задача 3: Добавить живой прогресс Runway по этапам
### Описание
Backend должен сохранять текущий этап ИИ-пайплайна Runway в `runway_sessions`, а simple UI должен показывать конкретный шаг и сообщение, чтобы пользователь видел, что аудит не завис.
### Затронутые файлы
- `internal/models/enterprise.go`
- `internal/server/db.go`
- `internal/server/repositories/runway_repo.go`
- `internal/server/server.go`
- `internal/agent/graph/state.go`
- `internal/agent/graph/orchestrator.go`
- `web/src/pages/SimpleDashboardPage.tsx`
- `web/src/locales/en/pages.json`
- `web/src/locales/ru/pages.json`
- `runway-start-fix.md`
### Подзадачи
- [x] Добавить поле progress message в модель, БД и runway repository.
- [x] Прокинуть progress callback из server Runway goroutine в graph.Run.
- [x] Обновлять progress на этапах: контекст, threat model, PoC, health check, report, fix spec, summary/completed.
- [x] Показывать в UI конкретный текущий этап, иконку/анимацию и подсказку вместо общего “аудит выполняется”.
- [ ] Проверить сборку, Docker web и runtime-состояние через API/browser.
### Критерии приёмки
- [ ] `GET /api/runway?product_id=...` возвращает `current_step > 1` и `progress_message` во время долгого Runway.
- [ ] Simple UI меняет текст статуса при переходе между этапами.
- [ ] Пользователь видит конкретный текущий этап, даже если LLM-запрос длится долго.
### Риски
- Текущий running-процесс был запущен старым кодом; после пересборки для проверки понадобится удалить только тестовую session и запустить Runway заново.

## Задача 4: Сохранить все Runway-артефакты во вкладки отчета
### Описание
После завершения Runway backend генерирует AI Fix Specification и Actionable Summary, но не сохраняет их в поля `security_plan` и `remediation` таблицы `runway_sessions`. Из-за этого вкладки `Security Plan` и `Remediation` в истории отчета остаются неактивными, хотя пайплайн прошел успешно.
### Затронутые файлы
- `internal/server/server.go`
- `runway-start-fix.md`
### Подзадачи
- [x] Сохранить `state.SummaryMarkdown` в `session.SecurityPlan`.
- [x] Сохранить `state.AIFixSpec` в `session.Remediation`.
- [x] Проверить, что текущие доступные артефакты session экспортируются в проектный `aitriage/`.
### Критерии приёмки
- [x] Новые завершенные Runway-сессии имеют непустые поля `security_plan` и `remediation`, если соответствующие артефакты были сгенерированы.
- [x] Вкладки `Security Plan` и `Remediation` становятся активными для новых завершенных отчетов.
- [x] Экспорт создает markdown-файл в проекте.
### Риски
- Уже завершенная session `17` не содержит утерянные `AIFixSpec/SummaryMarkdown` в SQLite; без повторного запуска или отдельного восстановления можно экспортировать только сохраненные поля.

## Риски и зависимости
- Проверка зависит от локального Docker/web-режима и доступности тестового продукта `aitriage-test-vulnerabilities`.
- Внешний Gemini/OpenAI-compatible endpoint уже возвращал 429; это ожидаемый внешний отказ, он должен быть показан пользователю.

## Финальный отчёт
- Исправлен frontend Runway: успешный старт переводит UI на шаг выполнения, polling не блокируется локальным loading-state, кнопка запуска блокируется на время активной сессии.
- Исправлен backend Runway: старт выставляет `status=running`, `current_step=1`, очищает старую ошибку и не запускает повторно уже running-сессию.
- Добавлены RU/EN-тексты для состояния запуска Runway.
- Проверки: `jq empty web/src/locales/en/pages.json web/src/locales/ru/pages.json`, `go test ./internal/server/...`, `npm --prefix web run build`, `docker compose up -d --build web`.
- Runtime-проверка: `http://localhost:8080/` отвечает, продукт `aitriage-test-vulnerabilities`, session `id=16` перешла в `status=running`, `current_step=1`; повторный `POST /api/runway/start/16` вернул `Runway scan already running` без создания новой session.
- Дополнительная диагностика session `id=17`: завершилась успешно, но `security_plan` и `remediation` остались пустыми из-за отсутствующего backend-маппинга. Исправлено сохранение `state.SummaryMarkdown` → `security_plan` и `state.AIFixSpec` → `remediation` для новых запусков.
- Экспорт текущей session `id=17` выполнен: создан `/Users/afedotov/Documents/GitHub/aitriage-test-vulnerabilities/aitriage/runway-report-17-2026-07-08.md`. В этом файле есть сохраненные секции Threat Model, Verification (PoC), Audit Report; секции Security Plan и Remediation отсутствуют у старой session, потому что не были записаны в SQLite при завершении.
