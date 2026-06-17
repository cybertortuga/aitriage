# readme-ci-cd

## Обзор
Актуализировать README.md, полностью пересобрав его по актуальным данным и возможностям AITriage. Особое внимание уделить разделу CI/CD: как работает пайплайн, роль pre-built Docker-образа в GHCR, двухслойная модель проверок (deterministic gate + AI advisor), интеграция с GitHub Actions и CodeQL/SARIF, политики ИБ и механизм Health Check.

## Затронутые файлы и области
- [README.md](../../../README.md) — полная пересборка и актуализация
- [.aitriage.yaml.example](../../../.aitriage.yaml.example) — обновление до актуальной структуры (синхронизация с docs/aitriage.yaml.example)

## Задача 1: Аудит и сбор актуальных данных для README.md
### Описание
Собрать все актуальные факты о работе CI/CD, механизмах Docker-escalation, поддержке AST, Entropy checks, baseline/suppression, watch mode, и SBOM.
### Затронутые файлы
- [README.md](../../../README.md)
### Подзадачи
- [x] Проверить все поддерживаемые CLI команды и флаги (`scan`, `agent`, `fix`, `baseline`, `watch`, `sbom`, `rules`, `init`, `install-mcp`).
- [x] Сформулировать подробное описание логики CI/CD (GHCR pre-built Docker-образ, two-layer model, GitHub annotations, step summaries, CodeQL/SARIF).
- [x] Проверить структуру профилей политики ИБ (baseline, standard, strict) и их влияние на вердикт CI.
### Критерии приёмки
- [x] Собран полный список актуального функционала, параметров и команд.
### Риски
Нет.

## Задача 2: Обновление файла .aitriage.yaml.example в корне
### Описание
Обновить .aitriage.yaml.example, чтобы он содержал актуальные настройки, включая блок health_check, аналогично docs/aitriage.yaml.example.
### Затронутые файлы
- [.aitriage.yaml.example](../../../.aitriage.yaml.example)
### Подзадачи
- [x] Заменить устаревшие комментарии и параметры на актуальную структуру с `health_check:` политиками.
### Критерии приёмки
- [x] Файл содержит пример полной настройки `health_check: profile, fail_on, minimum_score, max_critical, max_high, max_medium, block_sources, block_classes`.
### Риски
Нет.

## Задача 3: Полная пересборка и актуализация README.md
### Описание
Написать новый, структурированный и стилистически выверенный README.md, содержащий актуальные примеры интеграции и подробное описание архитектуры CI/CD.
### Затронутые файлы
- [README.md](../../../README.md)
### Подзадачи
- [x] Описать двухслойную модель CI/CD: Детерминированный гейт (Layer 1) и ИИ-Ассистент (Layer 2).
- [x] Добавить в README.md точные примеры workflows для GitHub Actions с использованием официального Docker Action (`cybertortuga/aitriage@v1`).
- [x] Объяснить, как работает оптимизация сборки через GHCR (pre-built образ сокращает время с 15 минут до нескольких секунд).
- [x] Описать логику оценки ИБ-политики (verdict, passed, blocking reasons).
- [x] Актуализировать таблицы правил и поддерживаемых команд.
### Критерии приёмки
- [x] README.md содержит актуальные примеры файлов конфигурации и GitHub Actions.
- [x] Раздел CI/CD детально описывает двухслойный подход и использование Docker-образа из GHCR.
### Риски
- Риск упустить обратную совместимость параметров или опечататься в YAML конфигах. Будет проверено вручную.

## Риски и зависимости
Изменения носят исключительно документарный характер, поэтому риски нарушения работоспособности кода приложения отсутствуют. Важно гарантировать корректность YAML-примеров, чтобы пользователи не сталкивались с синтаксическими ошибками в CI/CD.

## Финальный отчёт
1. **Аудит**: Детально исследованы исходный код CLI-интерфейсов (`scan.go`, `agent.go`), конфигурационные механизмы в `internal/healthpolicy` и `internal/report/healthcheck`, а также существующие workflow-пайплайны в `.github/workflows/`.
2. **Конфигурация**: Файл `.aitriage.yaml.example` обновлён до актуального стандарта, включая полный набор полей блока `health_check` (`profile`, `fail_on`, `minimum_score`, `max_critical`, `max_high`, `max_medium`, `block_sources`, `block_classes`).
3. **README**: Полностью пересобран `README.md`. В раздел CI/CD добавлено описание двухслойной архитектуры (Layer 1: Deterministic Gate + SARIF, Layer 2: AI Advisor / PR Agent), преимуществ использования pre-built контейнера из GHCR (`cybertortuga/aitriage@v1`), а также подробностей политики ИБ (health-profile: baseline, standard, strict).
4. **Тестирование**: Запущен полный цикл тестов `go test -p 1 ./...`, все тесты успешно прошли (cached/зелёные).
