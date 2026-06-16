# AITriage — Анализ конкурентов и стратегический выбор
> Дата: 14 апреля 2026  
> Контекст: Команда сформулировала требования к "ИИ-агенту security" (task.md + mindmap).  
> Цель: Определить — строить своё или использовать готовое.

---

## 1. Ландшафт рынка (апрель 2026)

Рынок AI Security Tools взорвался за последние 6 месяцев. Тренды:
- **Agentic Workflows** — инструменты перешли от "scan-and-report" к автономным агентам
- **Remediation-First** — фокус на автоматическом исправлении, а не только обнаружении
- **PR-Native Enforcement** — безопасность на уровне Pull Request
- **OWASP Agentic AI Top 10** (декабрь 2025) — новая таксономия рисков для автономных AI

### Ключевые open-source проекты

| Проект | ⭐ | Язык | Фокус | Ссылка |
|---|---|---|---|---|
| **Ship-Safe** | 407 | Node.js | 22 агента, 80+ классов атак, agentic loop | [github](https://github.com/asamassekou10/ship-safe) |
| **PentestAgent** | 2000 | Python | Мультиагентный AI-пентестер, MCP server | [github](https://github.com/GH05TCREW/pentestagent) |
| **GoThreatScope** | ~100 | Go | MCP-сервер: SBOM + vuln + secrets | [github](https://github.com/anotherik/GoThreatScope) |
| **SeCoRA** | ~200 | Python | AI SAST через OpenAI, OWASP Top 10 | [github](https://github.com/shivamsaraswat/secora) |
| **Strix** | ~300 | Python | Автономный AI-пентестер с верификацией PoC | [github](https://github.com/usestrix/strix) |
| **Semgrep** | 10K+ | OCaml | Rule-based SAST, industry standard | [github](https://github.com/semgrep/semgrep) |
| **AITriage** (наш) | 0 | Go | O(N) SAST, Tree-Sitter AST, Shannon Entropy | — |

---

## 2. Детальное сравнение: AITriage vs Ship-Safe

Ship-Safe — наш прямой конкурент. Покрывает ~95% требований из task.md.

### 2.1 Полная матрица требований task.md

| Требование (из task.md) | Ship-Safe | AITriage | Комментарий |
|---|---|---|---|
| **Сканеры** | | | |
| ├ semgrep | ✅ InjectionTester | ❌ | Ship-Safe имеет свой аналог |
| ├ bandit | ✅ EntropyPatternAgent | ❌ | Частично покрыт |
| ├ gitleaks | ✅ Secrets + GitHistoryScanner | ✅ Shannon Entropy | Разные подходы |
| └ trivy | ✅ deps audit + ConfigAuditor | ❌ | — |
| **NFR** | | | |
| ├ Проверка на соответствие | ✅ SOC2, ISO 27001, NIST | ❌ | — |
| └ Преаудит | ✅ `npx ship-safe checklist` | ❌ | — |
| **Аудит инфры** | | | |
| ├ Защищённость VM | ✅ CICDScanner | ❌ | — |
| ├ Порты/домены/базы | ✅ APIFuzzer | ❌ | — |
| └ IaC | ✅ Terraform, K8s, Docker | ⚠️ только Dockerfile | — |
| **Архитектура** | | | |
| ├ Секреты (где/как) | ✅ 50+ паттернов + verification | ✅ Entropy + AST | AST точнее |
| ├ Файлы деплоя | ✅ ConfigAuditor | ⚠️ частично | — |
| └ Git-анализ | ✅ GitHistoryScanner | ❌ | — |
| **Требования** | | | |
| ├ Натуральный язык | ✅ Claude Code plugin | ❌ | — |
| ├ GUI | ✅ webapp + VS Code ext | ❌ | CLI only |
| ├ OpenSource | ✅ MIT | ✅ MIT | — |
| ├ Инструкции AI на фикс | ✅ agentic loop | ❌ stub | — |
| ├ Прожарка архитектуры | ✅ 22 агента | ✅ AST + rules | — |
| ├ AI-lab agnostic | ✅ 12+ провайдеров | ❌ нет LLM | — |
| └ Дистрибуция | | | |
| 　├ Легко установить | ✅ `npx` | ✅ `go install` | Оба хороши |
| 　└ Плагины IDE | ✅ Claude hooks + VS Code | ❌ | — |
| **Воркфлоу** | | | |
| ├ Параллельные сканеры | ✅ 22 агента параллельно | ✅ goroutines | — |
| ├ Собрать артефакты | ✅ .ship-safe/context.json | ⚠️ SARIF только | — |
| ├ Консолидация | ✅ ScoringEngine | ✅ scorer | — |
| ├ Сборка отчёта | ✅ HTML/PDF/SARIF/JSON/CSV/MD | ✅ HTML/SARIF/JSON | — |
| ├ Рекомендации | ✅ remediation plan | ⚠️ suggestion поле | — |
| ├ Промпт для AI-фикса | ✅ `--agentic` loop | ❌ | — |
| └ Консультация (Q&A) | ✅ через Claude Code | ❌ | — |

**Итого: Ship-Safe 26/28 ✅ vs AITriage 9/28 ✅**

### 2.2 Что есть у AITriage и НЕТ у Ship-Safe

| Преимущество AITriage | Почему это важно | Ship-Safe альтернатива |
|---|---|---|
| **Tree-Sitter AST** | Структурный анализ кода, а не regex. Не даёт false positives на комментарии/строки | LLM-powered deep analysis (regex + AI) |
| **Shannon Entropy в AST-контексте** | Ищет секреты в контексте переменных/присваиваний, а не в произвольных строках | 50+ regex паттернов + API verification |
| **O(N) single-pass** | Один проход по файлам, все проверки за раз | 22 агента, каждый ходит по файлам отдельно |
| **Go binary** | 10-100x быстрее Node.js на больших монорепах | Node.js |
| **Zero dependencies** | `go install` → один бинарник, ноль runtime | Требует Node.js 18+ |
| **Полностью offline** | Детерминированный результат без API-ключей | Scanning offline, но deep analysis требует API |
| **Entropy stripping** | AST-aware разделение кода и комментариев перед анализом | Regex-based filtering |

### 2.3 Что есть у Ship-Safe и НЕТ у AITriage

| Преимущество Ship-Safe | Проодробнее |
|---|---|
| 22 специализированных агента | InjectionTester, AuthBypassAgent, SSRFProber, SupplyChainAudit, ConfigAuditor, SupabaseRLSAgent, LLMRedTeam, MCPSecurityAgent, AgenticSecurityAgent, RAGSecurityAgent, MemoryPoisoningAgent, PIIComplianceAgent, EntropyPatternAgent, ExceptionHandlerAgent, AgentConfigScanner, MobileScanner, GitHistoryScanner, CICDScanner, APIFuzzer, ManagedAgentScanner, HermesSecurityAgent, AgentAttestationAgent |
| LLM deep analysis | Верификация exploitability через любой LLM |
| Agentic loop | scan → auto-annotate fixes → re-scan until score ≥ target |
| Claude Code hooks | PreToolUse/PostToolUse — блокирует секреты ДО записи на диск |
| Claude Code plugin | `/ship-safe` команды прямо в IDE |
| VS Code extension | GUI в редакторе |
| Webapp (SaaS) | Веб-дашборд для команд |
| Secrets verification | Проверяет через API провайдеров жив ли утёкший ключ |
| 12+ LLM провайдеров | Anthropic, OpenAI, Google, Ollama, Groq, Together, Mistral, DeepSeek, xAI, Perplexity, LM Studio |
| 5 стандартов OWASP | Web 2025, Mobile 2024, LLM 2025, CI/CD, Agentic AI |
| Compliance mapping | SOC 2 Type II, ISO 27001:2022, NIST AI RMF |
| Policy-as-code | `.ship-safe.policy.json` — командные стандарты |
| Baseline management | Принять текущие находки, показывать только регрессии |
| Diff scanning | Сканировать только изменённые файлы |
| CI/CD pipeline | `npx ship-safe ci .` с exit codes и PR comments |
| GitHub Action | Готовый action для CI |
| SBOM | CycloneDX 1.5 |
| MCP server | `npx ship-safe mcp` |
| Incremental scanning | Кеширование результатов, пересканирует только изменённые |
| Supply chain hardening | Pinned SHA, OIDC, CODEOWNERS, provenance |
| Hermes Agent integration | Нативная интеграция с NousResearch Hermes |
| Legal audit | DMCA, leaked-source derivatives |

---

## 3. Сравнение: AITriage vs PentestAgent

PentestAgent — **другая ниша** (offensive pentesting), но его **фреймворк** теоретически можно переиспользовать.

| Критерий | PentestAgent | AITriage |
|---|---|---|
| **Назначение** | Active pentesting (атака) | Code audit (защита) |
| **Язык** | Python 3.10+ | Go 1.25+ |
| **LLM** | LiteLLM (любой провайдер) | Нет |
| **MCP** | ✅ server + client | ❌ |
| **TUI** | ✅ Textual | ❌ CLI only |
| **Multi-agent** | ✅ crew mode | ❌ |
| **Инструменты** | terminal, browser, web_search, notes | AST engine, entropy, entropy |
| **Playbooks** | thp3_recon, thp3_web, thp3_network | ❌ |
| **Docker** | ✅ base + Kali | ❌ |
| **RAG** | ✅ FAISS + sentence-transformers | ❌ |
| **Звёзды** | 2000 | 0 |
| **Установка** | `pip install -e .` (тяжёлая) | `go install` (лёгкая) |

**Вывод:** Форк PentestAgent — 50K+ строк чужого Python. Мы меняем ВСЁ (промпты, инструменты, playbooks) кроме каркаса (~500 строк). Не оправдано.

---

## 4. Сравнение: AITriage vs GoThreatScope

GoThreatScope — ближайший архитектурный аналог (Go + MCP + Security).

| Критерий | GoThreatScope | AITriage |
|---|---|---|
| Язык | Go | Go |
| MCP server | ✅ | ❌ (планируется) |
| SBOM | ✅ | ❌ |
| Vuln detection (osv.dev) | ✅ | ❌ |
| Secrets scanning | ✅ (gitleaks fallback) | ✅ (Shannon Entropy) |
| AST-анализ | ❌ | ✅ (Tree-Sitter) |
| Entropy code detection | ❌ | ✅ |
| Метрики | ✅ | ❌ |
| Звёзды | ~100 | 0 |

**Вывод:** GoThreatScope — proof that Go + MCP security server works. Но у него нет нашего главного козыря (AST).

---

## 5. Технический анализ Ship-Safe

### Архитектура
- **Язык:** JavaScript (49%), TypeScript (38%), CSS (11%)
- **Runtime:** Node.js 18+
- **Установка:** `npx ship-safe audit .` (npm)
- **Зависимости:** 5 прямых (минимально)
- **Кеш:** `.ship-safe/context.json` + `.ship-safe/llm-cache.json`
- **Contributors:** 2 (автор + Claude AI)

### Метод сканирования
- **Pattern matching** — regex паттерны для каждого агента
- **Entropy scoring** — для секретов
- **LLM deep analysis** — для верификации exploitability (опционально)
- **НЕ использует AST** — это главное архитектурное отличие от AITriage

### Scoring System
- Стартует с 100
- 8 категорий с весами (Secrets 15%, Code Vulns 15%, Dependencies 13%, Auth 15%, Config 8%, Supply Chain 12%, API 10%, AI/LLM 12%)
- Confidence levels: high 100%, medium 60%, low 30%
- Грейды: A (90-100), B (75-89), C (60-74), D (40-59), F (0-39)

### Сильные стороны
1. Невероятно широкий охват (22 агента, 80+ классов атак)
2. Agentic loop — автоматический fix + re-scan
3. Native Claude Code integration (hooks + plugin)
4. Supply chain hardening (практикует что проповедует)
5. Активная разработка (v8.0.0, коммиты каждый день)

### Слабые стороны
1. **Regex-based** — false positives на комментариях и строках
2. **Node.js** — медленнее Go на порядки
3. **Deep analysis зависит от LLM** — стоит денег, нужен API ключ
4. **Один автор + AI** — bus factor = 1
5. **SaaS ambitions** — shipsafecli.com/pricing (может стать freemium)

---

## 6. Стратегические опции

### Опция A: Использовать Ship-Safe "как есть"
- **Усилие:** 0
- **Выгода:** 95% task.md закрыто немедленно
- **Риск:** Зависимость от чужого проекта, нет ownership
- **Когда выбрать:** Если цель — просто иметь рабочий инструмент сейчас

### Опция B: AITriage → Go MCP Server (специализация)
- **Усилие:** 8-10 дней (~2800 строк)
- **Выгода:** Уникальный AST-powered MCP tool, один бинарник, нулевые зависимости
- **Риск:** Конкурируем с Ship-Safe на узком сегменте
- **Когда выбрать:** Если хотим свой продукт с уникальной ценностью
- **Позиционирование:** "Быстрый, точный, оффлайновый SAST для CI/CD и MCP"
- **Подробный план:** см. [INTEGRATION_PLAN.md](./INTEGRATION_PLAN.md)

### Опция C: Форк PentestAgent
- **Усилие:** 4-6 недель
- **Выгода:** Standalone agent с TUI и чатом
- **Риск:** 50K+ строк чужого Python, два языка навсегда
- **Когда выбрать:** Никогда (Ship-Safe делает это лучше)
- **Вердикт: ❌ Не рекомендуется**

### Опция D: Контрибьютить в Ship-Safe
- **Усилие:** 1-2 недели (PR с AST engine)
- **Выгода:** Community exposure, наш код в проекте с 407 звёздами
- **Риск:** PR может не принять, теряем IP
- **Когда выбрать:** Если хотим influence без ownership burden

### Опция E: Ship-Safe для команды + AITriage как R&D
- **Усилие:** Минимальное
- **Выгода:** Команда сразу получает рабочий инструмент, а мы спокойно развиваем AST-ядро
- **Риск:** Размытый фокус
- **Когда выбрать:** Если хотим и результат сейчас, и собственный продукт потом

---

## 7. Рекомендация

**Для немедленного результата** → Опция A (Ship-Safe)  
`npx ship-safe audit .` на проектах команды. Завтра.

**Для долгосрочного продукта** → Опция B (AITriage MCP)  
Позиционирование: *"Fastest deterministic SAST engine. Zero dependencies. One binary. MCP-native."*

**Оптимальный гибрид** → Опция E  
- Команда использует Ship-Safe для ежедневной работы
- Мы параллельно превращаем AITriage в MCP tool
- Через 2-3 недели AITriage дополняет Ship-Safe как "быстрое AST-ядро"

---

## Приложения

### A. Полный список MCP tools (Go SDK)
Official Go SDK: `github.com/modelcontextprotocol/go-sdk`  
API: `mcp.NewServer()` → `mcp.AddTool()` → `server.Run()` (stdio/SSE)

### B. Ссылки
- [Ship-Safe GitHub](https://github.com/asamassekou10/ship-safe)
- [Ship-Safe Website](https://shipsafecli.com)
- [PentestAgent GitHub](https://github.com/GH05TCREW/pentestagent)
- [GoThreatScope GitHub](https://github.com/anotherik/GoThreatScope)
- [SeCoRA GitHub](https://github.com/shivamsaraswat/secora)
- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [OWASP Agentic AI Top 10](https://owasp.org/www-project-agentic-ai-top-10/)
