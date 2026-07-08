# url-checker

Асинхронный сервис проверки доступности HTTP-эндпоинтов. Клиент отправляет пачку URL, сразу получает ID задачи, а результаты проверок собираются в фоне конкурентно.

![CI](https://github.com/HotHeadd/url-health-checker/actions/workflows/ci.yml/badge.svg)

## Что делает

`url-checker` принимает по HTTP список URL, конкурентно проверяет каждый (статус-код и время ответа) и сохраняет агрегированный результат под ID задачи. Отправка неблокирующая: `POST` сразу возвращает ID (`202 Accepted`), проверки идут в фоновой горутине, а результат клиент забирает отдельным `GET`. Так задержка запроса не зависит от того, сколько времени занимает опрос медленных или недоступных хостов.

## Стек

- **Go 1.26** — стандартная библиотека `net/http` с роутингом по методам (Go 1.22+, `GET /checks/{id}`), без веб-фреймворка
- **PostgreSQL 16** через **pgx/pgxpool** (напрямую, не через `database/sql`)
- **goose** — миграции схемы
- **slog** — структурированное JSON-логирование
- **golangci-lint** (v2) и **GitHub Actions** — CI
- **testify** + `net/http/httptest` — тесты

## Модель данных в БД

Две таблицы (см. `db/migrations`):

- `tasks` — строка на каждую отправку: `id` (UUID), `proc_status` (`pending` / `done` / `failed`), `created_at`.
- `results` — строка на каждый проверенный URL, внешний ключ на `tasks`, индекс по `task_id`.

## Быстрый старт

### Требования

- Go 1.26+
- Docker (для контейнера Postgres)
- CLI [`goose`](https://github.com/pressly/goose) (для миграций)

### 1. Конфигурация

Скопируй пример env-файла и задай строку подключения:

```sh
cp example.env .env
```

В `.env` нужен `DATABASE_URL`, например:

```
DATABASE_URL=postgres://urlchecker:computer@localhost:5432/urlchecker?sslmode=disable
```

### 2. Запуск БД и миграции

```sh
make db-up        # поднять контейнер Postgres
make migrate-up   # применить миграции
```

### 3. Запуск сервера

```sh
make run          # слушает :8080
```

## API

### `GET /health`

Проверить сервер.

```sh
curl -i http://localhost:8080/health
```

```json
{ "status": "ok" }
```

### `POST /checks`

Отправить пачку URL. Сразу возвращает ID задачи; проверки идут в фоне.

```sh
curl -i -X POST http://localhost:8080/checks \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://example.com", "https://httpbin.org/status/500"]}'
```

Ответ (`202 Accepted`):

```json
{ "id": "1f7c...", "status": "pending" }
```

### `GET /checks/{id}`

Забрать результат отправленной задачи.

```sh
curl -i http://localhost:8080/checks/1f7c...
```

Ответ (`200 OK`) по готовности:

```json
{
  "id": "1f7c...",
  "status": "done",
  "results": [
    { "url": "https://example.com", "status_code": 200, "request_duration_ms": 142, "error": "" },
    { "url": "https://httpbin.org/status/500", "status_code": 500, "request_duration_ms": 210, "error": "" }
  ]
}
```

Возвращает `400` при некорректном UUID и `404` для неизвестной задачи.

## Запуск

```sh
make test         # прогнать тесты
make test-race    # тесты с race-детектором
make lint         # golangci-lint
make help         # список всех таргетов
```

## Лицензия

MIT
