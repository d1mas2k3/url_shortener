# URL Shortener

Сервис для сокращения ссылок. Принимает длинный URL, возвращает короткий код из 10 символов.

## API

### Сократить ссылку

```
POST /api/v1/links
Content-Type: application/json

{ "url": "https://example.com/some/very/long/path" }
```

Ответ:
```json
{ "short_url": "http://localhost:8080/links/aB3_kZ9mQp" }
```

Если тот же URL отправить повторно - вернётся тот же короткий код (дедупликация).

### Получить оригинальный URL

```
GET /api/v1/links/aB3_kZ9mQp
```

Ответ:
```json
{ "original_url": "https://example.com/some/very/long/path" }
```

## Хранилище

Сервис поддерживает два хранилища, выбор задаётся переменной окружения `APP_STORAGE`:

- `memory` — хранение в памяти приложения. Быстро, данные теряются при перезапуске.
- `postgres` — хранение в PostgreSQL. Данные сохраняются между перезапусками.

## Запуск

### In-memory режим

Создай `.env` файл в корне (пример ниже) и запусти:

```bash
make run-memory
```

### Postgres режим

```bash
make up           # поднять контейнер PostgreSQL
make init-db      # создать таблицу
make run-postgres
```

### Docker

```bash
make docker-build        # собрать образ
make docker-run-memory   # запустить в memory-режиме
```

## Пример .env

```env
# Приложение
APP_BASE_URL=http://localhost:8080

# HTTP сервер
HTTP_ADDR=:8080
HTTP_SHUTDOWN_TIMEOUT=30s

# Логгер
LOGGER_LEVEL=DEBUG
LOGGER_FOLDER=./logs

# PostgreSQL (нужно только для режима postgres)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=url_shortener
POSTGRES_TIMEOUT=5s

# Docker Compose
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=url_shortener
PROJECT_ROOT=.
```

## Тесты

```bash
make test        # запустить тесты
make test-race   # запустить тесты с проверкой гонок данных
```

Покрыты юнит-тестами:
- In-memory репозиторий: сохранение, поиск по коду и URL, дубли
- Сервис: дедупликация, валидация пустого URL, resolve несуществующего кода

## Технические детали

**Генерация кода** — `crypto/rand` генерирует 10 случайных байт, каждый маппируется
на символ алфавита из 63 знаков (a-z, A-Z, 0-9, `_`). Итого 63^10 ≈ 3.5 * 10^17
возможных комбинаций. При коллизии кода сервис автоматически генерирует новый (retry).

**Конкурентность** — in-memory хранилище использует `sync.RWMutex` с двумя индексами
(code→url и url→code), что позволяет множеству горутин читать одновременно.
Postgres-режим использует пул соединений `pgxpool`.

**Сверх ТЗ:**
- Структурированное логирование (zap) — запись в консоль и в файл одновременно
- Middleware цепочка: Request ID, логирование каждого запроса, трассировка времени, перехват паник
- Graceful shutdown — при Ctrl+C сервер дожидается завершения текущих запросов
- Таймауты на запросы к БД — защита от зависания при проблемах с PostgreSQL
- Трёхслойная архитектура с разделением через интерфейсы — легко добавить новое хранилище

## Стек

- Go 1.25
- [pgx](https://github.com/jackc/pgx) — драйвер PostgreSQL
- [zap](https://github.com/uber-go/zap) — логирование
- [envconfig](https://github.com/kelseyhightower/envconfig) — конфигурация из env
- [validator](https://github.com/go-playground/validator) — валидация запросов
- [uuid](https://github.com/google/uuid) — генерация Request ID
