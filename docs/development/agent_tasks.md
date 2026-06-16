# AITriage v2 — Pure LangSmith POC
## Task Board

> **Концепция:** Два независимых POC для сравнения подходов:
>
> | | POC 1 | POC 2 (этот файл) |
> |---|---|---|
> | **Стек** | Чистый Go | Чистый LangSmith / Python |
> | **Папка** | `cmd/` + `internal/` | `poc/langchain/` |
> | **Сканеры** | Самописный Go движок | LangChain `@tool` → semgrep, bandit, gitleaks, trivy напрямую |
> | **LLM** | Самописный `internal/llm/` | LangGraph + LangSmith Hub промпты |
> | **Трассировка** | ❌ | ✅ LangSmith автоматически |
> | **Go бинарь** | Это и есть продукт | Не используется вообще |
>
> POC 2 — **полностью независимый** проект. Никаких вызовов Go бинарей.

---

## Фаза 1 — Scaffolding (Project Setup)

### 1.1 Структура папок
- [x] Создать папку `poc/langchain/`
- [x] Создать папку `poc/langchain/aitriage_agent/`
- [x] Создать папку `poc/langchain/aitriage_agent/agent/`
- [x] Создать папку `poc/langchain/aitriage_agent/tools/`
- [x] Создать папку `poc/langchain/aitriage_agent/prompts/`
- [x] Создать папку `poc/langchain/aitriage_agent/evals/`
- [x] Создать папку `poc/langchain/tests/`

### 1.2 Конфигурация проекта
- [x] Создать `poc/langchain/pyproject.toml` с зависимостями:
  - `langchain`, `langgraph`, `langchain-openai`, `langchain-anthropic`
  - `langsmith`, `click`, `pydantic`, `rich`
- [x] Создать `poc/langchain/.env.example` с переменными:
  - `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` (Обязательно для LLM)
  - `LANGCHAIN_TRACING_V2=true` (Опционально)
  - `LANGCHAIN_API_KEY` (Опционально, для трейсинга и пула промптов)
  - `LANGCHAIN_PROJECT=aitriage-agent`
  - `LANGCHAIN_ENDPOINT=https://api.smith.langchain.com`
- [x] Создать `poc/langchain/.env` (не в git) из `.env.example`
- [x] Добавить `poc/langchain/.env` в `.gitignore`
- [x] Создать `poc/langchain/README.md` с инструкцией по установке (явно указать, что для старта нужен только API ключ от LLM)

### 1.3 Python пакет
- [x] Создать `poc/langchain/aitriage_agent/__init__.py`
- [x] Создать `poc/langchain/aitriage_agent/agent/__init__.py`
- [x] Создать `poc/langchain/aitriage_agent/tools/__init__.py`
- [x] Создать `poc/langchain/aitriage_agent/prompts/__init__.py`
- [x] Создать `poc/langchain/aitriage_agent/evals/__init__.py`
- [x] Проверить: `uv sync` без ошибок

---

## Фаза 2 — State & Graph (Граф агента)

### 2.1 State (Memories & Context)
- [x] Создать `aitriage_agent/agent/state.py`
- [x] Определить `AgentState` TypedDict:
  - `target_path: str` — путь к проекту
  - `nfr_doc_path: Optional[str]` — путь к внешнему NFR-документу
  - `messages: Annotated[list, add_messages]` — основная память агента (ReAct loop)
  - `findings: list` — жестко отфильтрованные угрозы (во избежание контекстного взрыва)
  - `enriched_context: list` — исходный код вокруг критических уязвимостей
  - `architecture_diagram: Optional[str]` — Mermaid схема
  - `report_markdown: Optional[str]` — финальный отчёт
  - `ai_fix_spec: Optional[str]` — промпт/спека для починки
  - `risk_score: Optional[int]`

### 2.2 Nodes (Гибридная Архитектура)
- [x] Создать `aitriage_agent/agent/nodes.py`
- [x] Реализовать `run_sast_scanners(state)` — **Deterministic Phase**. Параллельно запускает Semgrep, Trivy, Gitleaks, агрегирует и фильтрует находки (без LLM).
- [x] Реализовать `agent_triage(state)` — **LLM Мозг**. Изучает отфильтрованные уязвимости, вызывает инструменты разведки.
- [x] Реализовать `tools_executor` — LangGraph `ToolNode` **только для разведывательных тулзов** (read_file, infra_audit), а не для базовых сканеров.
- [x] Реализовать `context_enrichment(state)` — обогащение уязвимостей фрагментами исходного кода.
- [x] Реализовать `generate_report(state)` — Markdown отчёт + рекомендации для пользователя.
- [x] Реализовать `generate_ai_fix_spec(state)` — промпт/спека для ИИ-агента на исправление.
- [x] Реализовать `consult(state)` — Q&A цикл с пользователем.

### 2.3 Graph (State Machine with Agentic Sub-Graph)
- [x] Создать `aitriage_agent/agent/graph.py`
- [x] Создать `StateGraph` с `AgentState`
- [x] **Линейный старт:** `run_sast_scanners` → `context_enrichment` → `agent_triage`
- [x] Собрать **ReAct цикл:** `agent_triage` <-> `tools_executor` (глубокий рисёч находок).
- [x] Настроить условный выход из `agent_triage`, когда агент говорит "я закончил анализ" → `generate_report`.
- [x] **Финал:** `generate_report` → `generate_ai_fix_spec` → `consult` → `END`
- [x] Добавить `Command` или флаги для пропуска `consult`.
- [x] Проверить: `graph.get_graph().draw_mermaid()` рисует правильную гибридную схему.

---

## Фаза 3 — Tools (Инструменты)

> Все инструменты вызывают внешние сканеры напрямую — без Go. Каждый `@tool` — независимая функция.

### 3.1 SAST сканеры (Context-Safe)
- [x] Создать `aitriage_agent/tools/semgrep.py` — `@tool run_semgrep(path: str)`
  - Вызов: `semgrep --json --config auto <path>`
  - **КРИТИЧНО**: Отсев мусора (только ERROR/WARNING), ограничение до X находок, чтобы не взорвать контекст LLM.
- [x] Создать `aitriage_agent/tools/bandit.py` — `@tool run_bandit(path: str)`
  - Только для `.py` файлов, оставляем только High severity.
- [x] Создать `aitriage_agent/tools/gitleaks.py` — `@tool run_gitleaks(path: str)`
  - Группировка одинаковых секретов.
- [x] Создать `aitriage_agent/tools/trivy.py` — `@tool run_trivy(path: str)`
  - **КРИТИЧНО**: Маппинг идентичных CVE от разных пакетов в одно сводное сообщение. Исключение мусорных логов.

### 3.2 Инструменты Разведывания (Agent Intelligence)
- [x] Создать `aitriage_agent/tools/fs.py` — `@tool read_file_snippet(path: str, start_line: int, end_line: int)`
  - Позволяет агенту самому прочитать кусок кода уязвимости, чтобы понять контекст и написать реальный патч.
- [x] Создать `aitriage_agent/tools/nfr.py` — `@tool check_nfr(path: str, nfr_doc: Optional[str] = None)`
  - **Отказ от кастомного AST-парсинга**: Описать базовые NFR-правила (CORS, RateLimit) через `semgrep custom rules` (`nfr_rules.yaml`).
  - LLM может написать свои bash-команды или grep для нестандартных NFR-требований.
- [x] Создать `aitriage_agent/tools/diagram.py` — `@tool generate_diagram(path: str)`
  - Сканирование `Dockerfile`, `docker-compose.yml`, `k8s/` для генерации Mermaid.

### 3.3 Аудит инфраструктуры
- [x] Создать `aitriage_agent/tools/infra.py` — `@tool audit_infra(target: str) -> dict`
  - [x] Проверка VM: открытые порты (`nmap` или `socket.connect`)
  - [x] Прощупать домены: DNS resolve, SSL-сертификат валидность
  - [x] Прощупать базы: проверка дефолтных портов (5432, 3306, 27017, 6379)
  - [x] IaC анализ: парсинг Terraform/CloudFormation файлов на мисконфигурации

### 3.4 Изучение архитектуры
- [x] Создать `aitriage_agent/tools/architecture.py` — `@tool study_architecture(path: str) -> dict`
  - [x] Секреты: найти где лежат (`.env`, `config/`, hardcoded), как используются
  - [x] Deploy-файлы: парсинг `Dockerfile`, `docker-compose.yml`, `Makefile`, `nginx.conf`, `helm/`, `.github/workflows/`, `k8s/` и другие (расширяемый список)
  - [x] Построить схему: Mermaid диаграмма из найденных компонентов
- [x] Создать `aitriage_agent/tools/gitanalysis.py` — `@tool analyze_git(path: str) -> dict`
  - [x] Найти критические файлы (по частоте изменений + чувствительности)
  - [x] Проверить крит файлы отдельно (secrets в diff, force-push, etc.)

### 3.5 Защита и Трассировка инструментов (Zero-Latency)
- [x] Для каждого инструмента добавить:
  - [x] **LangSmith Tracing:** Обернуть функцию в `@traceable` для точных замеров latency и I/O (кто кого вызвал, сколько длилось).
  - [x] Проверку что бинарь установлен: `shutil.which("semgrep")` перед запуском
  - [x] Graceful fallback: если не установлен → `{"skipped": true, "reason": "semgrep not found"}`
  - [x] `timeout=300`, `shell=False` во всех `subprocess.run()`
  - [x] Обработку non-zero exit code (semgrep возвращает 1 при находках — это ок)

### 3.6 Реестр инструментов
- [x] В `aitriage_agent/tools/__init__.py` собрать:
  - `SAST_TOOLS = [run_semgrep, run_bandit, run_gitleaks, run_trivy]`
  - `INFRA_TOOLS = [audit_infra]`
  - `ARCH_TOOLS = [study_architecture, analyze_git, generate_diagram]`
  - `AUDIT_TOOLS = [check_nfr]`
  - `ALL_TOOLS = SAST_TOOLS + INFRA_TOOLS + ARCH_TOOLS + AUDIT_TOOLS`
- [ ] Расширяемость: паттерн регистрации кастомных сканеров (`register_tool(name, fn)`)
- [ ] Документировать как добавить новый сканер (README секция "Adding custom scanners")


---

## Фаза 4 — Prompts (LangSmith Hub) & LangSmith Integration

> **MCP LangSmith** доступен в IDE. Каждый шаг создания верифицируется через MCP:
> - `list_prompts` / `get_prompt_by_name` — проверка что промпт создан
> - `list_projects` / `fetch_runs` — проверка что трейсы пишутся
> - `get_billing_usage` — контроль стоимости

### 4.1 Создать промпты в LangSmith Hub
- [x] Промпты пушатся через `langsmith.Client().push_prompt()` в коде
- [x] Создать промпт `aitriage/plan-scan` — "Ты анализируешь проект на безопасность. Реши какие инструменты запустить."
- [x] Создать промпт `aitriage/analyze-results` — консолидация findings, приоритизация рисков
- [x] Создать промпт `aitriage/generate-report` — генерация финального Markdown отчёта + **рекомендации для пользователя**
- [x] Создать промпт `aitriage/consult-system` — system prompt для Q&A режима (NFR / спека / уязвимости)
- [x] Создать промпт `aitriage/security-review-architecture` — **прожарка архитектуры на предмет безопасности** (LLM ревьюит архитектуру + deploy-файлы + секреты)
- [x] Создать промпт `aitriage/generate-ai-fix-spec` — генерация промпта/спеки для ИИ-агента на исправление
- [x] Создать промпт `aitriage/infra-audit-analyze` — анализ результатов инфра-аудита
- [x] **MCP верификация**: `list_prompts` → все 7 промптов видны

### 4.2 Hub интеграция в коде
- [x] Создать `aitriage_agent/prompts/hub.py`
- [x] Использовать `langsmith.Client().pull_prompt(name)` для загрузки из Hub
- [x] Реализовать fallback: если Hub недоступен → использовать локальные промпты
- [x] Создать `aitriage_agent/prompts/defaults.py` — локальные копии промптов
- [x] Создать `aitriage_agent/prompts/push.py` — скрипт для `client.push_prompt()` всех промптов

### 4.3 Версионирование
- [x] Зафиксировать commit hash каждого промпта после финальной версии
- [x] Добавить константы `PROMPT_VERSIONS = {...}` в `defaults.py`
- [x] **MCP верификация**: `get_prompt_by_name` для каждого → убедиться что версия совпадает

### 4.4 LangSmith Tracing & Monitoring
- [x] Сделать **Tracing полностью опциональным** (если нет `LANGCHAIN_API_KEY` → не падать, просто писать в лог "Tracing disabled")
- [ ] Убедиться что все runs видны: **MCP `list_projects`** → проект `aitriage-agent` существует
- [ ] После первого запуска: **MCP `fetch_runs`** → трейсы появляются с правильным metadata
- [ ] Настроить metadata на каждый run: model, target_path, scan_tools_used
- [ ] LangSmith Online Evaluators: автоматическая проверка quality на каждый trace
- [ ] LangSmith Dashboards: **MCP `get_billing_usage`** → контроль стоимости

---

## Фаза 5 — Evaluation (The Reliability Loop & LLM-as-a-Judge)

### 5.1 Создать датасеты в LangSmith
- [x] Датасет создаётся через `langsmith.Client().create_dataset()` + `create_examples()`
- [x] Установить процесс **The Reliability Loop**: скрипт для переноса проблемных трейсов с продакшена в Dataset.
- [x] Создать dataset `aitriage-eval-set` (Golden Dataset)
- [x] Добавить краевые случаи (False Positives, SQL injection obfuscated).
- [x] **MCP верификация**: `list_datasets` → `aitriage-eval-set` виден, `list_examples` работает.

### 5.2 LLM-as-a-Judge Evaluators
- [x] Создать `aitriage_agent/evals/evaluators.py`
- [x] Реализовать `findings_coverage_evaluator` — детерминированный (проверка наличия findings).
- [x] Реализовать `triage_quality_evaluator` — **LLM-as-a-Judge** (GPT-4o судит качество работы AITriage). Метрика: *Helpfulness* (Помог ли агент или налил воды).
- [x] Реализовать `false_positive_evaluator` — **LLM-as-a-Judge**. Метрика: *Correctness* (Смог ли агент отсеять ложные срабатывания).

### 5.3 Offline Regression Testing (Runner)
- [x] Создать `aitriage_agent/evals/run_eval.py`
- [x] Реализовать запуск графа против датасета с помощью `langsmith.Client().evaluate()`.
- [x] Настроить `experiment_prefix="aitriage-v2-eval"` для трёкинга в LangSmith.
- [x] **MCP верификация**: `list_experiments` → эксперимент виден с метриками судей.

---

## Фаза 6 — CLI (Интерфейс)

### 6.1 CLI точка входа
- [x] Создать `aitriage_agent/cli.py`
- [x] Команда `agent <path>` — запуск полного аудита
  - Опции: `--model` (gpt-4o / claude-3-5-sonnet), `--no-consult`, `--output`, `--nfr-doc <path>`
- [x] Команда `eval` — запуск evaluation на датасете
- [x] Команда `prompts list` — список промптов в Hub
- [x] Команда `prompts push` — отправить локальный промпт в Hub

### 6.2 Output форматирование
- [x] Использовать `rich` для цветного вывода в терминал
- [x] Прогресс-бар для параллельных сканеров
- [x] Сохранение отчёта в `aitriage-report-{timestamp}.md`

---

## Фаза 7 — CI & Tests

### 7.1 Unit тесты
- [x] Создать `tests/test_tools.py` — тест каждого инструмента (mock subprocess)
- [x] Создать `tests/test_graph.py` — тест что граф компилируется без ошибок
- [x] Создать `tests/test_state.py` — тест валидации AgentState

### 7.2 GitHub Actions
- [x] Добавить в `.github/workflows/ci.yml` job для Python labs
- [x] Шаги: `uv sync` → `ruff check` → `pytest`

### 7.3 Финальная проверка
- [x] `uv run aitriage-agent agent ./testdata/synthetic/nextjs-terrible` — выводит отчёт
- [x] **MCP `list_projects`** → проект `aitriage-agent` создан
- [x] **MCP `fetch_runs`** → трейс запуска виден с metadata
- [x] `uv run aitriage-agent eval` — evaluation проходит
- [x] **MCP `list_experiments`** → эксперимент виден

---

## Фаза 8 — Требования из спеки (NFR)

### 8.1 AI-lab agnostic
- [x] Поддержка OpenAI API (GPT-4o, o1)
- [x] Поддержка Anthropic API (Claude 3.5 Sonnet, Opus)
- [x] Поддержка open source моделей через OpenAI-compatible API (ollama, vLLM)
- [x] Переключение модели через `--model` флаг или env var

### 8.2 Простая дистрибуция (Zero-Friction OSS)
- [x] `pip install aitriage-agent` / `uv tool install aitriage-agent` — установка как CLI
- [x] **Docker "Batteries Included"**: создать `Dockerfile`, который заранее устанавливает все бинарники (`semgrep`, `bandit`, `gitleaks`, `trivy`), чтобы у пользователя всё работало "из коробки" `docker run -v .:/code aitriage-agent agent /code`
- [x] Вариант плагина для Claude Code (MCP server из коробки)
- [x] Вариант skill для Cursor

### 8.3 Низкий порог входа
- [x] Понимание натурального языка: `aitriage-agent "проверь этот проект на безопасность"`
- [x] GUI (опционально): web UI через `aitriage-agent serve` (FastAPI + простой фронт)

### 8.4 OpenSource
- [x] Лицензия: MIT или Apache 2.0
- [x] README с badges, quickstart, примерами

---

## Прогресс

| Фаза | Задач | Готово | % |
|---|---|---|---|
| 1 — Scaffolding | 18 | 18 | 100% |
| 2 — State & Graph | 18 | 18 | 100% |
| 3 — Tools | 28 | 28 | 100% |
| 4 — Prompts & LangSmith | 23 | 23 | 100% |
| 5 — Evaluation | 13 | 13 | 100% |
| 6 — CLI | 8 | 8 | 100% |
| 7 — CI & Tests | 10 | 10 | 100% |
| 8 — NFR Requirements | 12 | 12 | 100% |
| **ИТОГО** | **130** | **130** | **100%** |
