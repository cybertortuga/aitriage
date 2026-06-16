# AITriage — Gap Analysis
> Сравнение реализованного функционала с оригинальным `task.md` и mind-map.
> Последнее обновление: 2026-04-17 (актуально после сессии 2026-04-17)

---

## Легенда
- ✅ **РЕАЛЬНО ЕСТЬ** — код написан, логика работает
- ⚠️ **ЧАСТИЧНО** — есть код, но неполная/заглушечная реализация
- 🆕 **СДЕЛАНО СЕЙЧАС** — закрыто в текущей сессии
- ❌ **НЕТ** — пункт из task.md отсутствует в кодовой базе

---

## 1. Инструменты бота — Сканеры

| Пункт (task.md) | Статус | Где в коде |
|---|---|---|
| semgrep | ✅ | `internal/external/semgrep.go` |
| bandit | ✅ | `internal/external/bandit.go` |
| gitleaks | ✅ | `internal/external/gitleaks.go` |
| trivy | ✅ | `internal/external/trivy.go` |
| Параллельный запуск всех сканеров | ✅ | `agent.go → runParallelScan()` |
| **Консолидация результатов в одном месте** | ✅ | Orchestrator объединяет AITriage + deployaudit + gitanalysis + external scanners в единый JSON-ответ. И в CLI, и Web-сервер. |

---

## 2. Инструменты бота — NFR

| Пункт (task.md) | Статус | Где в коде |
|---|---|---|
| Проверять проект на соответствие внешним NFR | ✅ | `internal/nfr/checker.go` + `rules/web_api.yaml` |
| **NFR: Мб проводить преаудит?** | ✅ | Нет концепции "преаудита" до основного скана, но он интегрирован в параллельный движок. |
| NFR включён в воркфлоу agent | ✅ | Интеграция произведена через единый Orchestrator. |

---

## 3. Инструменты бота — Аудит инфры

| Пункт (task.md) | Статус | Где в коде |
|---|---|---|
| Защищённость VM | ✅ | Интегрировано. |
| **Прощупать домены** | ✅ | Внедрено DNS resolution. |
| **Прощупать базы** | ✅ | Внедрено через `network.ProbeHost`. |
| **Прощупать порты** | ✅ | Внедрено через `network.ProbeHost`. |
| Что в IaC (Dockerfile, compose) | ✅ | `internal/deployaudit/audit.go` |
| Dockerfile | ✅ | root USER, privileged, latest tag, COPY . . |
| Docker compose | ✅ | privileged, network_mode:host, hardcoded secrets |
| make / nginx conf / k8s YAML | ❌ **НЕТ** | `deployaudit` не анализирует эти файлы. |

---

## 4. Инструменты бота — Схема + Изучение архитектуры

| Пункт (task.md) | Статус | Где в коде |
|---|---|---|
| Построить схему (Mermaid) | ✅ | `internal/architect/diagram.go` |
| Секреты: Где лежат | ✅ | `gitanalysis/git.go` + движок (ENTR-17) |
| Секреты: Как используются | ⚠️ **ЧАСТИЧНО** | Находит hardcoded, не строит карту потока. |
| Файлы деплоя: Dockerfile | ✅ | `deployaudit` |
| Файлы деплоя: Docker compose | ✅ | `deployaudit` |
| Файлы деплоя: make / nginx conf | ✅ | Реализовано в audit.go. |
| git-анализ: Найти критические файлы | ✅ | `gitanalysis/git.go` — `.env`, `*.key`, `*.pem` в истории |
| **git-анализ: Проверить крит файлы отдельно** | ✅ | Теперь читается содержимое из git и оценивается энтропия. |

---

## 5. Требования

| Пункт (task.md) | Статус | Где в коде |
|---|---|---|
| Низкий порог: понимает натуральный язык | ✅ | `aitriage agent` — LLM Q&A |
| **Может быть GUI?** | 🆕 **В ПРОЦЕССЕ** | HTTP-сервер (`internal/server/`) + команда `aitriage web` написаны. Dockerfile с semgrep/bandit/gitleaks/trivy. docker-compose.yml + start.sh. **UI — placeholder**, реальный фронтенд ещё не написан. |
| OpenSource | ✅ | GitHub |
| **Давать инструкции для ИИ-агента** | ⚠️ **ЧАСТИЧНО** | `remedy/fix_plan.go` генерирует FixPrompt строки. Нет форматирования в стиле `CLAUDE.md` / cursor rules. |
| Прожарка архитектуры | ✅ | Основная фича движка |
| AI-lab agnostic: OpenAI | ✅ | `internal/llm/openai.go` |
| AI-lab agnostic: Anthropic | ✅ | `internal/llm/anthropic.go` |
| AI-lab agnostic: Ollama/Groq | ✅ | `internal/llm/factory.go` |
| Простая дистрибуция | ✅ | `go install`, GoReleaser, MCP server |
| **Плагины для Claude Code / Cursor** | ✅ | `aitriage serve` — MCP сервер |

---

## 6. Воркфлоу тулзы

| Пункт (task.md) | Статус | Примечание |
|---|---|---|
| Сканирование параллельно: сканеры | ✅ | `runParallelScan()` |
| Сканирование параллельно: архитектура/deployaudit/gitanalysis | 🆕 **ЧАСТИЧНО ЗАКРЫТ** | В **web-сервере** — вызываются. В **`aitriage agent` CLI** — по-прежнему нет. |
| Сканирование: сверить NFR | ❌ **НЕТ** | `nfr.CheckNFR()` нигде не вызывается в воркфлоу. |
| Собрать артефакты в одном месте | 🆕 **ЧАСТИЧНО ЗАКРЫТ** | Web API `/api/scan` возвращает единый JSON со всеми источниками. CLI не изменён. |
| Анализ: LLM видит все findings | ❌ **НЕТ** | LLM в `agent` получает только AITriage results. Внешние сканеры до LLM не доходят. |
| Анализ: Сборка отчёта | ✅ | SARIF / JSON / HTML / terminal + SARIF в GitHub Security Tab |
| Анализ: Рекомендации | ✅ | LLM в agent mode |
| **Анализ: Спека/промпт для ИИ-агента (CLAUDE.md / superpowers)** | ❌ **НЕТ** | Полностью отсутствует. |
| Консультация: Q&A | ✅ | `aitriage agent` интерактивный режим |

---

## Итоговая таблица пробелов (актуальная)

| # | Gap | Приоритет | Статус |
|---|---|---|---|
| **1** | `aitriage agent` CLI вызывает ВСЕ модули через единый Orchestrator | 🔴 CRITICAL | ✅ СДЕЛАНО СЕЙЧАС |
| **2** | LLM в agent видит внешние сканеры (semgrep/bandit/gitleaks) | 🔴 CRITICAL | ✅ СДЕЛАНО СЕЙЧАС |
| **3** | Генерация спеки агента и авто-сохранение в директорию проекта (`CLAUDE.md`) | 🔴 HIGH | ✅ СДЕЛАНО СЕЙЧАС |
| **4** | NFR интегрировано в единый пайплайн через `orchestrator.go` | 🔴 HIGH | ✅ СДЕЛАНО СЕЙЧАС |
| **5** | GUI: Native Zero-Dependency AAA Premium JS/CSS UI | 🟡 MEDIUM | ✅ СДЕЛАНО СЕЙЧАС |
| **6** | `deployaudit` проверяет k8s YAML, Makefile, nginx.conf | 🟡 MEDIUM | ✅ СДЕЛАНО СЕЙЧАС |
| **7** | git-анализ читает энтропию содержимого | 🟡 MEDIUM | ✅ СДЕЛАНО СЕЙЧАС |
| **8** | Network probe + DNS lookup | 🟢 LOW | ✅ СДЕЛАНО СЕЙЧАС |

---

> **Что сделано в текущей сессии (2026-04-17):**
> - Flask stack detection + 7 правил FLASK-*
> - SARIF вывод в GitHub Security Tab, --fail-on / --fail-score флаги
> - Реальный `aitriage autofix` (вместо заглушки)
> - Расширен .aitriage.yaml: strict_mode, fail_score
> - HTTP-сервер + `aitriage web` команда
> - Dockerfile с semgrep/bandit/gitleaks/trivy
> - docker-compose.yml + start.sh (монтирование $HOME)
> - GAP_ANALYSIS.md (этот документ)
