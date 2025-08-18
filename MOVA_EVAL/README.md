
# Cursor Context DB (SQLite, мінімум)

Тільки те, що треба для **Cursor**: зберігати снапшоти контексту й швидко відновлювати їх у робочу сесію.

- Файл БД: `state/.cursor_ctx.db` (SQLite, WAL)
- Схема: `schema.sql`
- CLI: `.tools/ctx.py`

## Використання (3 команди)

```bash
# 1) Ініціалізувати (одноразово або при зміні schema.sql)
python .tools/ctx.py init

# 2) Зберегти снапшот
python .tools/ctx.py save --title "Refactor plan" --text "Короткий контекст для Cursor" --tags mova

# 3) Відновити
python .tools/ctx.py show --id 1 --format text  # вивід скопіювати в Cursor
# або
python .tools/ctx.py export --id 1 --out ctx.txt
```

Переглянути останні:
```bash
python .tools/ctx.py last --limit 10
```
