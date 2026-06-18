# Как устроен URL Shortener — подробный разбор

Этот документ объясняет как работает весь проект изнутри: что от чего зависит,
как данные проходят через слои, почему код написан именно так.

---

## Оглавление

1. [Что делает сервис](#что-делает-сервис)
2. [Структура папок](#структура-папок)
3. [Три слоя архитектуры](#три-слоя-архитектуры)
4. [Как данные проходят через слои — пример POST-запроса](#как-данные-проходят-через-слои--пример-post-запроса)
5. [Как данные проходят через слои — пример GET-запроса](#как-данные-проходят-через-слои--пример-get-запроса)
6. [Ключевой принцип: интерфейс живёт у потребителя](#ключевой-принцип-интерфейс-живёт-у-потребителя)
7. [Dependency Injection в main.go](#dependency-injection-в-maingo)
8. [Как работает генератор кодов](#как-работает-генератор-кодов)
9. [Как обрабатываются ошибки](#как-обрабатываются-ошибки)
10. [Как работает in-memory хранилище](#как-работает-in-memory-хранилище)
11. [Как работает postgres хранилище](#как-работает-postgres-хранилище)
12. [Middleware — что происходит до хендлера](#middleware--что-происходит-до-хендлера)
13. [Конфигурация через переменные окружения](#конфигурация-через-переменные-окружения)
14. [Запуск сервиса](#запуск-сервиса)
15. [Тесты](#тесты)

---

## Что делает сервис

Сервис принимает длинный URL и возвращает короткий код из 10 символов.

```
POST /api/v1/links
{ "url": "https://very-long-url.com/path?query=value" }

→ { "short_url": "http://localhost:8080/links/aB3_kZ9mQp" }
```

```
GET /api/v1/links/aB3_kZ9mQp

→ { "original_url": "https://very-long-url.com/path?query=value" }
```

Если один и тот же длинный URL отправить дважды — оба раза вернётся
один и тот же короткий код (дедупликация).

---

## Структура папок

```
cmd/
  server/
    main.go                         ← точка входа, сборка всего приложения

internal/
  core/                             ← переиспользуемая инфраструктура, не знает о фичах
    config/
      config.go                     ← общий конфиг: STORAGE, BASE_URL
    domain/
      link.go                       ← доменная сущность Link {Code, OriginalURL}
    errors/
      common.go                     ← sentinel-ошибки: ErrNotFound, ErrInvalidArgument и др.
    logger/
      logger.go                     ← обёртка над zap, запись в консоль + файл
      config.go                     ← конфиг логгера: LEVEL, FOLDER
    repository/
      postgres/
        pool/
          pool.go                   ← интерфейс Pool + реализация ConnectionPool (pgxpool)
          config.go                 ← конфиг пула: HOST, PORT, USER, PASSWORD, DB, TIMEOUT
    transport/
      http/
        middleware/
          middleware.go             ← тип Middleware, функция ChainMiddleware
          common.go                 ← RequestID, Logger, Trace, Panic middleware
        request/
          decode.go                 ← декодирование и валидация JSON из тела запроса
        response/
          handler.go                ← JSONResponse, ErrorResponse, PanicResponse
          writer.go                 ← обёртка над http.ResponseWriter для запоминания статус-кода
        server/
          server.go                 ← HTTPServer: запуск и graceful shutdown
          router.go                 ← APIVersionRouter: группировка маршрутов по версии API
          route.go                  ← структура Route {Method, Path, Handler, Middleware}
          config.go                 ← конфиг сервера: ADDR, SHUTDOWN_TIMEOUT

  features/
    links/                          ← фича "ссылки" — три подслоя
      repository/
        memory/
          repository.go             ← структура LinksMemoryRepository (две map + RWMutex)
          save.go                   ← метод Save
          get_by_code.go            ← метод GetByCode
          get_by_url.go             ← метод GetByURL
          repository_test.go        ← юнит-тесты in-memory репозитория
        postgres/
          repository.go             ← структура LinksPostgresRepository (pool)
          save.go                   ← метод Save (INSERT + маппинг unique violation)
          get_by_code.go            ← метод GetByCode (SELECT WHERE code = $1)
          get_by_url.go             ← метод GetByURL (SELECT WHERE original_url = $1)
      service/
        service.go                  ← структура LinkService, интерфейс LinksRepository
        shorten.go                  ← метод Shorten (дедупликация + генерация + retry)
        resolve.go                  ← метод Resolve (достать ссылку по коду)
        generator.go                ← функция generateCode (crypto/rand + алфавит 63 символа)
        service_test.go             ← юнит-тесты сервиса
      transport/
        http/
          transport.go              ← структура LinksHTTPHandler, интерфейс LinksService, Routes()
          shorten.go                ← HTTP-хендлер POST /links
          resolve.go                ← HTTP-хендлер GET /links/{code}
          dto_common.go             ← DTO структуры + функция linkDTOFromDomain

init.sql                            ← SQL-схема таблицы links
docker-compose.yaml                 ← поднимает контейнер PostgreSQL
Dockerfile                          ← multi-stage сборка Go-бинарника
Makefile                            ← команды запуска, тестов, docker
```

---

## Три слоя архитектуры

Всё приложение разбито на три слоя. Данные всегда идут сверху вниз:

```
HTTP-запрос
    ↓
Transport (transport/http/)   — знает об HTTP, не знает о БД
    ↓
Service (service/)            — знает о бизнес-правилах, не знает об HTTP и БД
    ↓
Repository (repository/)      — знает о БД, не знает об HTTP
    ↓
База данных / память
```

**Почему это важно?**

Каждый слой отвечает только за своё. Если завтра нужно сменить PostgreSQL на MongoDB —
меняешь только `repository/postgres/`, остальное не трогаешь. Если нужно добавить gRPC —
добавляешь `transport/grpc/`, сервис и репозиторий остаются нетронутыми.

---

## Как данные проходят через слои — пример POST-запроса

Запрос: `POST /api/v1/links` с телом `{"url": "https://example.com"}`

**Шаг 1 — Middleware цепочка** (до хендлера):
```
RequestID → Logger → Trace → Panic
```
- `RequestID` — присваивает запросу уникальный UUID (или берёт из заголовка X-Request-ID)
- `Logger` — создаёт логгер с прикреплёнными полями request_id и url, кладёт его в context
- `Trace` — запоминает время начала запроса, логирует "incoming request"
- `Panic` — оборачивает всё в recover(), если что-то паникует — поймает и вернёт 500

**Шаг 2 — Transport** (`shorten.go`):
```go
ctx := r.Context()                              // контекст с логгером внутри
log := core_logger.FromContext(ctx)             // достаём логгер
responseHandler := NewHTTPResponseHandler(...)  // инструмент для отправки ответов

var request ShortenRequest
DecodeAnyValidateRequest(r, &request)           // JSON → структура + валидация тегов

link, err := h.linksService.Shorten(ctx, request.URL)  // вызов сервиса

response := linkDTOFromDomain(link, h.baseURL)  // domain.Link → DTO
responseHandler.JSONResponse(response, 201)     // отправка ответа
```

**Шаг 3 — Service** (`shorten.go`):
```go
// 1. Валидация
if len(originalURL) == 0 → ErrInvalidArgument

// 2. Дедупликация: может этот URL уже сокращён?
link, err := repo.GetByURL(ctx, originalURL)
if err == nil → вернуть существующую ссылку (не создавать новую)

// 3. Генерация нового кода в цикле
for {
    code = generateCode()           // случайные 10 символов
    err = repo.Save(ctx, link)

    if ErrCodeExists → continue     // коллизия кода — retry
    if ErrURLExists  → GetByURL()   // гонка данных — другая горутина успела первой
    if err == nil    → return link  // успех
}
```

**Шаг 4 — Repository** (`save.go` — memory):
```go
r.mx.Lock()                         // полная блокировка на запись
defer r.mx.Unlock()

if _, ok := r.byURL[link.OriginalURL]; ok → ErrURLExists
if _, ok := r.byCode[link.Code]; ok       → ErrCodeExists

r.byURL[link.OriginalURL] = link.Code     // записываем в обе мапы
r.byCode[link.Code] = link.OriginalURL
```

**Ответ клиенту:**
```json
{ "short_url": "http://localhost:8080/links/aB3_kZ9mQp" }
```

---

## Как данные проходят через слои — пример GET-запроса

Запрос: `GET /api/v1/links/aB3_kZ9mQp`

**Transport** (`resolve.go`):
```go
code := r.PathValue("code")                    // "aB3_kZ9mQp" из URL-пути
link, err := h.linksService.Resolve(ctx, code)
responseHandler.JSONResponse(ResolveResponse{OriginalURL: link.OriginalURL}, 200)
```

**Service** (`resolve.go`):
```go
link, err := repo.GetByCode(ctx, code)
if err != nil → fmt.Errorf("getting by code: %w", err)  // пробрасываем ошибку
return link, nil
```

**Repository** (`get_by_code.go` — memory):
```go
r.mx.RLock()                        // блокировка только на чтение (можно несколько одновременно)
defer r.mx.RUnlock()

url, ok := r.byCode[code]
if !ok → ErrNotFound
return domain.Link{Code: code, OriginalURL: url}, nil
```

**Ответ клиенту:**
```json
{ "original_url": "https://example.com" }
```

---

## Ключевой принцип: интерфейс живёт у потребителя

Это самый важный архитектурный паттерн в проекте.

Обычно думают так: "репозиторий предоставляет интерфейс, сервис его использует".
Здесь наоборот: **сервис объявляет интерфейс того, что ему нужно**.

```go
// service/service.go — СЕРВИС объявляет что ему нужно от репозитория
type LinksRepository interface {
    Save(ctx context.Context, link domain.Link) error
    GetByCode(ctx context.Context, code string) (domain.Link, error)
    GetByURL(ctx context.Context, originalURL string) (domain.Link, error)
}
```

```go
// transport/http/transport.go — ТРАНСПОРТ объявляет что ему нужно от сервиса
type LinksService interface {
    Shorten(ctx context.Context, originalURL string) (domain.Link, error)
    Resolve(ctx context.Context, code string) (domain.Link, error)
}
```

**Что это даёт?**

Сервис не импортирует пакет репозитория. Транспорт не импортирует пакет сервиса.
Каждый верхний слой зависит только от своего интерфейса — абстракции, которую он сам
и определил. Конкретные реализации (`LinksMemoryRepository`, `LinksPostgresRepository`,
`LinkService`) удовлетворяют этим интерфейсам автоматически — в Go не нужно
писать `implements`, достаточно иметь нужные методы с правильными сигнатурами.

---

## Dependency Injection в main.go

Весь "монтаж" приложения происходит в `main.go`. Никаких DI-фреймворков — всё вручную:

```go
// 1. Логгер
logger := core_logger.NewLogger(core_logger.NewConfigMust())

// 2. Выбор хранилища (читаем APP_STORAGE из env)
switch appConfig.Storage {
case "memory":
    repo = links_memory_repository.NewLinksMemoryRepository()
case "postgres":
    pool = core_postgres_pool.NewConnectionPool(ctx, core_postgres_pool.NewConfigMust())
    repo = links_postgres_repository.NewLinksPostgresRepository(pool)
}

// 3. Сервис получает репозиторий
linksService := links_service.NewLinksService(repo)

// 4. Хендлер получает сервис
linksHandler := links_transport_http.NewLinksHTTPHandler(linksService, appConfig.BaseURL)

// 5. Сервер получает хендлер через роутер
httpServer := core_http_server.NewHTTPServer(config, logger, middleware...)
router := core_http_server.NewAPIVersionRouter(ApiVersion1)
router.RegisterRoutes(linksHandler.Routes()...)
httpServer.RegisterAPIRoutes(router)

// 6. Запуск с graceful shutdown
httpServer.Run(ctx)  // ctx отменяется при Ctrl+C
```

Зависимость идёт сверху вниз: `main` создаёт всё и передаёт вниз.
Никто ничего не создаёт сам внутри себя — всё получают через конструктор.

---

## Как работает генератор кодов

Файл: `service/generator.go`

```go
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
// 26 + 26 + 10 + 1 = 63 символа

func generateCode() (string, error) {
    b := make([]byte, 10)       // слайс из 10 байт
    rand.Read(b)                // заполняем случайными числами 0–255

    for i := range b {
        b[i] = alphabet[int(b[i]) % 63]  // каждое число → символ алфавита
    }

    return string(b), nil
}
```

**Почему `crypto/rand`, а не `math/rand`?**
`math/rand` — псевдослучайный генератор, предсказуемый при известном seed.
`crypto/rand` — криптографически стойкий генератор, непредсказуемый.
Для URL-shortener важно что коды нельзя угадать/предсказать.

**Сколько возможных комбинаций?**
63^10 ≈ 3,5 * 10^17 — это 350 квадриллионов вариантов.
Практически исчерпать невозможно.

**Почему цикл retry в `Shorten`?**
Теоретически два разных вызова могут сгенерировать одинаковый код одновременно.
Если `Save` вернул `ErrCodeExists` — просто генерируем новый код и пробуем снова.
При 63^10 комбинациях вероятность коллизии ничтожно мала.

---

## Как обрабатываются ошибки

В проекте используется паттерн "sentinel errors" + оборачивание через `%w`.

**Sentinel errors** — это заранее объявленные переменные-ошибки в `core/errors/common.go`:
```go
var (
    ErrNotFound        = errors.New("not found")
    ErrInvalidArgument = errors.New("invalid argument")
    ErrCodeExists      = errors.New("code already exists")
    ErrURLExists       = errors.New("url already exists")
)
```

**Оборачивание через слои:**
```
Repository:  fmt.Errorf("get by code=%q: %w", code, core_errors.ErrNotFound)
Service:     fmt.Errorf("getting by code: %w", err)
Transport:   responseHandler.ErrorResponse(err, "failed to resolve code")
```

На каждом уровне к ошибке добавляется контекст — где именно произошла ошибка.
При этом оригинальный sentinel сохраняется внутри цепочки благодаря `%w`.

**Как транспорт определяет HTTP-статус:**
```go
// response/handler.go
switch {
case errors.Is(err, ErrInvalidArgument): → 400 Bad Request
case errors.Is(err, ErrNotFound):        → 404 Not Found
default:                                 → 500 Internal Server Error
}
```

`errors.Is` "разворачивает" всю цепочку оборачиваний и находит sentinel внутри,
даже если он обёрнут в несколько слоёв `fmt.Errorf`.

**Ответ клиенту при ошибке:**
```json
{
    "message": "failed to resolve code",
    "error":   "getting by code: get by code=\"abc\": not found"
}
```

---

## Как работает in-memory хранилище

Файлы: `repository/memory/`

```go
type LinksMemoryRepository struct {
    mx     sync.RWMutex      // мьютекс для защиты от гонок данных
    byCode map[string]string  // code → originalURL
    byURL  map[string]string  // originalURL → code
}
```

**Два индекса** нужны для быстрого поиска в обе стороны:
- `GetByCode("aB3_kZ9mQp")` → ищет в `byCode`, O(1)
- `GetByURL("https://example.com")` → ищет в `byURL`, O(1)

Без второго индекса `GetByURL` пришлось бы перебирать все записи — O(n).

**RWMutex** — умный мьютекс:
- `RLock()` — блокировка на чтение. Несколько горутин могут читать одновременно.
- `Lock()` — блокировка на запись. Только одна горутина пишет, все остальные ждут.

Это важно при высокой нагрузке: сотни GET-запросов могут выполняться параллельно,
не мешая друг другу. Блокируются только когда кто-то пишет.

**Атомарность Save:**
Оба индекса обновляются под одним `Lock()` — нельзя ситуация когда
`byCode` уже записан, а `byURL` ещё нет. Либо оба, либо ни одного.

---

## Как работает postgres хранилище

Файлы: `repository/postgres/`

**Схема таблицы** (`init.sql`):
```sql
CREATE TABLE IF NOT EXISTS links (
    id           SERIAL PRIMARY KEY,
    code         VARCHAR(10)  NOT NULL UNIQUE,
    original_url TEXT         NOT NULL UNIQUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

`UNIQUE` на `code` и `original_url` — база данных сама гарантирует уникальность.
Попытка вставить дубль → PostgreSQL вернёт ошибку с кодом `23505` (unique_violation).

**Пул соединений** (`core/repository/postgres/pool/`):

Открывать новое соединение к БД на каждый запрос дорого (сотни миллисекунд).
`pgxpool` держит несколько соединений открытыми и переиспользует их.

```go
type Pool interface {
    QueryRow(ctx, sql, args...) pgx.Row   // один результат
    Query(ctx, sql, args...)    pgx.Rows  // много результатов
    Exec(ctx, sql, args...)               // без результата (INSERT/DELETE)
    OpTimeout() time.Duration             // таймаут операции
}
```

Репозиторий работает с интерфейсом `Pool`, а не с конкретной структурой —
это позволяет подменить реализацию в тестах без реальной БД.

**Таймаут на каждый запрос:**
```go
ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
defer cancel()
```

Если запрос к БД завис (сеть упала, БД перегружена) — через таймаут
контекст отменяется и запрос прерывается. Без этого горутина могла бы
зависнуть навсегда.

**Маппинг ошибки unique violation:**
```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    if pgErr.ConstraintName == "links_code_key" {
        return ErrCodeExists  // коллизия кода
    }
    return ErrURLExists       // дубль URL
}
```

PostgreSQL возвращает конкретный код ошибки `23505` и имя нарушенного constraint.
Мы переводим это в наши sentinel-ошибки — сервис не знает о PostgreSQL-специфике.

---

## Middleware — что происходит до хендлера

Middleware — это функция, которая оборачивает хендлер. Цепочка строится так:

```
запрос → RequestID → Logger → Trace → Panic → хендлер → ответ
```

Физически это выглядит как матрёшка:
```
Panic(
    Trace(
        Logger(
            RequestID(
                handler
            )
        )
    )
)
```

`ChainMiddleware` собирает эту матрёшку, проходя массив в обратном порядке:
```go
for i := len(m) - 1; i >= 0; i-- {
    h = m[i](h)  // каждый middleware оборачивает предыдущий
}
```

**RequestID** — добавляет уникальный ID запроса в заголовки. Нужен для трассировки:
если в логах несколько запросов, по request_id можно найти все строки одного.

**Logger** — кладёт логгер с request_id и URL в `context.Context`.
Хендлер достаёт его через `core_logger.FromContext(ctx)`.

**Trace** — логирует входящий запрос и итоговый статус-код после завершения хендлера.
Использует кастомный `ResponseWriter` который запоминает статус-код.

**Panic** — ловит паники через `recover()`. Без него одна паника в хендлере
уронила бы весь сервер. С ним — клиент получает 500, сервер продолжает работать.

---

## Конфигурация через переменные окружения

Каждый пакет читает свои env-переменные самостоятельно через `envconfig`.
Паттерн одинаковый везде: `NewConfig()` + `NewConfigMust()` (паникует при отсутствии обязательных).

| Переменная | Пакет | Описание |
|---|---|---|
| `APP_STORAGE` | core/config | `memory` или `postgres` |
| `APP_BASE_URL` | core/config | базовый URL сервиса, default: `http://localhost:8080` |
| `HTTP_ADDR` | transport/http/server | адрес и порт, например `:8080` |
| `HTTP_SHUTDOWN_TIMEOUT` | transport/http/server | таймаут graceful shutdown, default: `30s` |
| `LOGGER_LEVEL` | core/logger | уровень логов: DEBUG, INFO, WARN, ERROR |
| `LOGGER_FOLDER` | core/logger | папка для лог-файлов |
| `POSTGRES_HOST` | core/repository/postgres/pool | хост БД |
| `POSTGRES_PORT` | core/repository/postgres/pool | порт БД, default: `5432` |
| `POSTGRES_USER` | core/repository/postgres/pool | пользователь БД |
| `POSTGRES_PASSWORD` | core/repository/postgres/pool | пароль БД |
| `POSTGRES_DB` | core/repository/postgres/pool | имя базы данных |
| `POSTGRES_TIMEOUT` | core/repository/postgres/pool | таймаут запросов к БД |

**Почему паника при отсутствии обязательных переменных?**

Лучше упасть сразу при старте с понятным сообщением, чем молча запустить сервис
который всё равно не будет работать. Паника на старте — это предсказуемо.

---

## Запуск сервиса

**In-memory режим:**
```bash
make run-memory
# Читает HTTP_ADDR, APP_BASE_URL и другие из .env
# APP_STORAGE=memory — данные хранятся в памяти, теряются при перезапуске
```

**Postgres режим:**
```bash
make up          # поднять контейнер PostgreSQL
make init-db     # создать таблицу links
make run-postgres
```

**Docker:**
```bash
make docker-build         # собрать образ url-shortener
make docker-run-memory    # запустить контейнер в memory-режиме
```

**Тесты:**
```bash
make test        # go test ./...
make test-race   # go test -race ./... — обнаружение гонок данных
```

---

## Тесты

**`repository/memory/repository_test.go`** — тестирует in-memory репозиторий напрямую.

Тесты белого ящика (white-box): пакет тестов тот же что и код (`links_memory_repository`),
поэтому есть доступ к внутренним полям.

Каждый тест создаёт чистый репозиторий через `NewLinksMemoryRepository()` — изоляция.

**`service/service_test.go`** — тестирует сервис с реальным in-memory репозиторием.

Сервис получает `*LinksMemoryRepository` — тот самый, что реализует интерфейс
`LinksRepository`. Тест проверяет бизнес-логику: дедупликацию, валидацию, resolve.

**Почему именно такой подход к тестам сервиса?**

Изначально сервис тестировался через mock (ручную заглушку с функциями-полями).
Это позволяет изолировать сервис от репозитория и тестировать конкретные сценарии
(коллизия кода, гонка данных). Позже тесты были упрощены — используется реальный
memory-репозиторий, что проще для понимания, но покрывает меньше edge cases.

**Что проверяют тесты:**
- `TestSave_Success` — успешное сохранение
- `TestSave_DuplicateURL` — дубль URL → ErrURLExists
- `TestSave_DuplicateCode` — дубль кода → ErrCodeExists
- `TestGetByCode_Success` / `TestGetByCode_NotFound` — поиск по коду
- `TestGetByURL_Success` / `TestGetByURL_NotFound` — поиск по URL
- `TestShorten_Success` — новый URL → код длиной 10
- `TestShorten_EmptyURL` — пустой URL → ErrInvalidArgument
- `TestShorten_ExistingURL` — тот же URL дважды → один и тот же код
- `TestResolve_Success` — сохранить и найти обратно
- `TestResolve_NotFound` — несуществующий код → ErrNotFound
