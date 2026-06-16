# База знаний: AST-анализ и Tree-sitter для Go

## 1. Сравнение библиотек Go для Tree-sitter
| Параметр | `tree-sitter/go-tree-sitter` (Официальная) | `smacker/go-tree-sitter` (Legacy) |
| :--- | :--- | :--- |
| **Поддержка** | Официальная организация Tree-sitter | Сообщество (с 2018 года) |
| **Подход** | Модульный (грамматики отдельно) | Монолитный (грамматики вшиты) |
| **Рекомендация** | **Использовать для новых проектов** | Только для поддержки старого кода |

**Почему официальная:** Она легче, так как мы сами выбираем, какие языки поддерживать. Это позволяет избежать раздувания бинарника и гарантирует совместимость с обновлениями движка.

## 2. Язык запросов (S-expressions)
Tree-sitter использует Lisp-подобный синтаксис для поиска паттернов в дереве.

### Базовый синтаксис:
- `(node_type)` — найти узел указанного типа.
- `field: (node_type)` — найти узел внутри поля родителя.
- `@capture_name` — пометить узел для извлечения результата.
- `(#match? @name "regex")` — предикат для фильтрации захваченных узлов.

### Примеры для безопасности:
1. **JS/TS: Опасное использование innerHTML**
   ```scheme
   (assignment_expression
     left: (member_expression
       property: (property_identifier) @prop)
     (#eq? @prop "innerHTML"))
   ```

2. **Go: Поиск использования http.ListenAndServe без ограничений**
   ```scheme
   (call_expression
     function: (selector_expression
       field: (field_identifier) @func)
     arguments: (argument_list
       (string) @addr)
     (#eq? @func "ListenAndServe")
     (#match? @addr ":[0-9]+"))
   ```

## 3. Архитектурные нюансы в Go
- **CGO_ENABLED=1:** Обязательное условие. Без этого Tree-sitter не скомпилируется.
- **Парсинг:** Оптимально создавать парсер один раз на поток (используя `sync.Pool`) и менять язык через `SetLanguage`.
- **Утечки памяти:** ВСЕ объекты (Tree, Parser, Query, Cursor) имеют метод `Close()`, который удаляет C-объекты. Забытый `defer Close()` приведет к утечке, которую Go GC не исправит.

## 4. Инструменты отладки
- [Tree-sitter Playground](https://tree-sitter.github.io/tree-sitter/playground) — лучшее место для тестирования запросов.
- [AST Explorer](https://astexplorer.net/) — для визуализации структуры дерева разных языков.
