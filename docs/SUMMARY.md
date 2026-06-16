# AITriage — Security Presence Checker

## Что это

Open-source CLI-инструмент + GitHub Action, который проверяет entropy-based проекты на **наличие security-практик**.

Не ищет конкретные уязвимости (для этого есть Semgrep, Trivy, Snyk). Отвечает на вопрос: **"Какие security-практики в проекте отсутствуют?"**

## Зачем

LLM генерирует рабочий код, но без security awareness. Люди без ИБ-опыта деплоят проекты, в которых отсутствуют базовые вещи: аутентификация, rate limiting, валидация ввода. Существующие SAST/SCA-инструменты ищут баги в написанном коде, но не проверяют отсутствие целых практик.

## Что проверяет

13 presence checks — для каждого ответ PRESENT / ABSENT:

| Практика | Вопрос |
|----------|--------|
| Authentication | Есть ли механизм аутентификации? |
| Authorization | Есть ли проверка ролей/прав? |
| Input Validation | Валидируется ли пользовательский ввод? |
| Rate Limiting | Есть ли ограничение частоты запросов? |
| Error Handling | Скрываются ли stack traces? |
| Security Headers | Настроены ли CSP, HSTS? |
| CORS Policy | Явно ли сконфигурирован CORS? |
| Secrets Management | Секреты в env, а не в коде? |
| Parameterized Queries | ORM / prepared statements? |
| HTTPS Enforcement | Принудительный HTTPS? |
| Dependency Lockfile | Зафиксированы ли версии зависимостей? |
| Logging | Есть ли security-логирование? |
| CSRF Protection | Есть ли защита от CSRF? |

## Как работает

Определяет фреймворк проекта → ищет **маркеры** (файлы, импорты, зависимости, паттерны конфигов), специфичные для каждого стека. Маркер найден → PRESENT. Не найден → ABSENT.

Детерминированный, без LLM. Запустил дважды — одинаковый результат.

## Ключевая фраза

> Semgrep находит плохой код. AITriage находит отсутствие хорошего кода.

## Формат поставки

- **CLI:** `aitriage scan .`
- **GitHub Action** для CI/CD с fail-on-grade
- **Выходные форматы:** terminal, JSON, SARIF (GitHub Security tab)
- Итоговая оценка A-F по проценту присутствующих практик

## Стеки

- **MVP:** Next.js, FastAPI
- **Phase 2:** Express.js, Django
- **Phase 3:** Go, Rails

## Roadmap (6 мес.)

| Период | Результат |
|--------|-----------|
| Месяц 1-2 | MVP: CLI + 13 чеков для Next.js и FastAPI |
| Месяц 3 | GitHub Action + SARIF + PR-комментарии |
| Месяц 4 | Express.js, Django |
| Месяц 5-6 | npm publish, GitHub Marketplace, документация |

## Техстек

TypeScript, Node.js, regex + file checks (без AST для MVP), vitest.
