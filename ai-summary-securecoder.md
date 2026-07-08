# ai-summary-securecoder

## Обзор

Прокачать ИИ-сводку безопасности в Simple Dashboard так, чтобы она работала по принципу SecureCoder: сначала использовать фактический evidence из AITriage DB и контекста проекта, затем формулировать выводы без догадок. Текущий текст для `ait-dodo-landing` ошибочно утверждает backend/API/auth/rate-limit риски, хотя read-only аудит целевого проекта показывает static Vite + React landing page без backend routes, auth flows, forms, database client or request handlers.

Цель: dedicated `/api/ai-summary` должен отдавать evidence-grounded summary, а Simple UI не должен строить слабый prompt через общий `/api/chat`.

## Аудит

- MCP-инструменты проверены через `tool_search`.
- Доступны MCP: GitHub, Chrome DevTools, Node REPL, Firebase, Dart, Cloud Run.
- Отдельный local filesystem MCP для чтения/записи локальных файлов не доступен в exposed tools; локальный read-only аудит выполнен через `rg`, `find`, `sed`, `nl`, `git status`.
- `/Users/afedotov/Documents/GitHub/ait-dodo-landing` не является git repo в этой папке.
- `ait-dodo-landing` содержит `package.json`, `vite.config.ts`, `src/main.tsx`, `src/App.tsx`, компоненты landing page; backend/server/API routes не найдены.
- `ait-dodo-landing/SECURITY_TRIAGE_REPORT.md` уже фиксирует: `Authentication Middleware Missing` и `Rate Limiting Missing` являются false positive для текущего static landing scope, `.env.example missing` не является vulnerability без env usage.
- В `aitriage` рабочее дерево уже содержит много незакоммиченных изменений; их не откатывать.

## Research

- Go `database/sql`: `QueryContext`/`QueryRowContext` предназначены для SELECT with placeholder args; это подходит для безопасного DB evidence сборщика.
  Source: https://pkg.go.dev/database/sql
- Go database querying guide: использовать `QueryContext` для multiple rows and `QueryRow` для single row.
  Source: https://go.dev/doc/database/querying
- Go `net/url`: `Values.Get` возвращает первое query value or empty string; подходит для `lang`.
  Source: https://pkg.go.dev/net/url
- MDN `URLSearchParams`: стандартный browser API для query string building.
  Source: https://developer.mozilla.org/en-US/docs/Web/API/URLSearchParams
- MDN Fetch API: browser API for fetching resources; current UI already uses direct fetch in this component.
  Source: https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API

## Затронутые файлы и области

- `internal/server/server.go`
  - `handleSummary`
  - helper logic for summary evidence collection and fallback rendering
  - DB reads from `products`, `findings`, `engagements`
- `web/src/services/securityService.ts`
  - optional `lang` query support for `getAISummary`
- `web/src/pages/SimpleDashboardPage.tsx`
  - AI Security Summary button handler
  - remove frontend prompt guessing path for this feature
- `internal/server/server_test.go`
  - test fake LLM and summary evidence contract
- Optional validation only:
  - `go test ./internal/server`
  - `npm run build` inside `web` if TypeScript surface changes

## Задача 1: Backend evidence contract

### Описание

Replace the weak `/api/ai-summary` implementation that only sends counts and titles to the LLM. Build a structured evidence prompt from product metadata, latest scan path, severity/status counts, top findings with rule ID, file, line, status, AI triage status/summary, description, and fix suggestion.

The system prompt must reuse SecureCoder framing and forbid speculation:

- Do not infer backend/API/auth/rate-limit exposure unless evidence shows request handlers, API routes, forms, auth/session, database, or network services.
- Distinguish confirmed true positive, false positive/suppressed, and needs manual review.
- If evidence is scanner-only and project-scope uncertain, say verification needed rather than declaring critical status.
- Cite concrete files/rules.
- For RU/EN, respond in requested language.

### Затронутые файлы

- `internal/server/server.go`
- `internal/server/server_test.go`

### Подзадачи

- [x] Add small internal structs/helpers for summary evidence.
- [x] Query product metadata and latest engagement path when `product_id` is provided.
- [x] Query active and suppressed findings with rule/file/status details.
- [x] Build SecureCoder summary system/user messages with anti-speculation rules.
- [x] Improve deterministic fallback so it also avoids unverified backend/auth claims.
- [x] Add focused server tests with fake LLM checking prompt evidence and anti-speculation instructions.

### Критерии приёмки

- [x] `/api/ai-summary?product_id=...&lang=ru` prompts the LLM with product/file/rule/status evidence, not only titles.
- [x] Prompt explicitly forbids guessing project type or backend/API/auth exposure without evidence.
- [x] False-positive / risk-accepted / resolved findings are not counted as active critical blockers.
- [x] Fallback text cannot claim critical auth/rate-limit exposure solely from scanner titles.
- [x] `go test ./internal/server` passes or any failure is documented.

### Риски

- Existing DB may contain historical stale findings from older scans; summary must include statuses and avoid treating suppressed data as active.
- If no LLM client is configured, fallback must still be useful and safe.
- Do not refactor unrelated `handleChat`, Runway, or cache code.

## Задача 2: Simple Dashboard route

### Описание

Route the AI Security Summary button through dedicated `/api/ai-summary` instead of building a generic chat prompt in React. This removes the frontend instruction that asks the model to describe what the project "likely is".

### Затронутые файлы

- `web/src/services/securityService.ts`
- `web/src/pages/SimpleDashboardPage.tsx`

### Подзадачи

- [x] Add optional `lang` parameter to `securityService.getAISummary`.
- [x] Import/use `securityService` in `SimpleDashboardPage.tsx`.
- [x] Replace direct `/api/chat` prompt generation with `securityService.getAISummary(summaryProjectId, aiSummaryLang)`.
- [x] Remove now-unused local summary finding/count variables if TypeScript flags them.

### Критерии приёмки

- [x] Clicking Generate calls `/api/ai-summary` with product id and language.
- [x] Frontend no longer contains prompt text asking the model to infer what the project likely is.
- [x] UI still shows loading/error/result states.
- [x] Frontend build/typecheck passes or any failure is documented.

### Риски

- `securityService.ts` uses axios while this component currently uses direct `fetch`; import must not create circular dependencies.
- Existing UI layout should not change.

## Задача 3: Verification and final report

### Описание

Run focused verification and write the result into this plan file.

### Затронутые файлы

- `ai-summary-securecoder.md`

### Подзадачи

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./internal/server`.
- [x] Run `npm run build` in `web` if frontend TypeScript changed.
- [x] Inspect final diff for unrelated edits.
- [x] Append final report to this file.

### Критерии приёмки

- [x] All planned checkboxes are updated.
- [x] No unrelated code paths are changed by this task; pre-existing dirty Runway/worktree edits remain outside this summary change.
- [x] Final report records commands, results, and residual risks.

### Риски

- Current branch has many pre-existing uncommitted changes; final diff must identify only the files touched for this task.
- Existing unrelated changes may make broad test/build commands noisy.

## Риски и зависимости

- LLM output is not deterministic; quality must be enforced through a strict evidence contract plus safe deterministic fallback.
- Summary correctness depends on DB statuses being current. The feature should be honest when evidence is stale or incomplete.
- The target screenshot/example is in Russian; RU is the default language, but EN should remain supported.

## Финальный отчёт

### Что сделано

- `/api/ai-summary` переведён с weak counts/title prompt на SecureCoder-style evidence contract: product metadata, latest scan path, project shape markers, optional local `SECURITY_TRIAGE_REPORT.md`, active/suppressed severity counts, top findings with rule/file/line/status/AI triage fields.
- System prompt теперь запрещает выдумывать backend/API/auth/rate-limit exposure без evidence: request handlers, API routes, forms, auth/session, DB, network services.
- Deterministic fallback также не объявляет critical auth/rate-limit риск только по scanner title.
- Simple Dashboard больше не строит generic `/api/chat` prompt с фразой "what this project likely is"; Generate вызывает dedicated `securityService.getAISummary(productId, lang)`.
- Добавлен серверный тест, который проверяет, что LLM получает anti-speculation instructions и evidence по false-positive/static Vite проекту.

### Проверки

- `gofmt -w internal/server/server.go internal/server/server_test.go` — OK.
- `go test ./internal/server` — OK.
- `npm run build` в `web` — OK. Осталось стандартное предупреждение Vite/Rolldown: основной chunk больше 500 kB.
- `git diff --check -- internal/server/server.go internal/server/server_test.go web/src/services/securityService.ts web/src/pages/SimpleDashboardPage.tsx ai-summary-securecoder.md` — OK.
- `rg` проверка старого frontend prompt — OK: prompt с "what this project likely is" удалён из `SimpleDashboardPage.tsx`.

### Остаточные риски

- Качество финальной формулировки всё ещё зависит от LLM, но теперь ограничено evidence contract и безопасным fallback.
- Если DB содержит stale findings от старых scan runs, summary будет честно показывать статусы, но корректность зависит от актуальности этих статусов.
- Worktree до задачи уже был dirty; в `web/src/pages/SimpleDashboardPage.tsx` остаются pre-existing Runway changes, которые не относятся к AI summary.
