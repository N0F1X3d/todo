```md
# Todo (microservices)

Небольшое TODO-приложение на Go в формате микросервисов.

---

## 🧩 Архитектура

Основной поток:
```

HTTP client → api-service → gRPC → db-service → PostgreSQL

```

Поток событий (логирование):
```

api-service → Kafka → event-logger-service → ./logs

```

---

## 🧱 Сервисы

- **db-service** — gRPC-сервис задач
  - хранение задач в PostgreSQL
  - Redis-кеш задач (по ID, TTL) для оптимизации запросов к БД
  - миграции применяются автоматически при старте (golang-migrate)

- **api-service** — HTTP API (ходит в db-service по gRPC)

- **event-logger-service** — Kafka consumer, пишет события в `./logs`

- **postgres** — база данных
- **redis** — кеш (опционально, включается через env)
- **kafka + zookeeper** — брокер и координатор

---

## 📦 Стек технологий

- **Go**
- **gRPC**
- **PostgreSQL**
- **Redis**
- **Kafka / Zookeeper**
- **Docker / Docker Compose**
- **Taskfile**
- **Protobuf**
- **SQL migrations (golang-migrate)**

---

## 📁 Структура проекта

```

todo
├── db-service
│   ├── cmd/db-service          # Точка входа
│   ├── internal
│   │   ├── config              # cleanenv config (DB_/GRPC_/REDIS_)
│   │   ├── repository          # Работа с БД (+ Redis cache)
│   │   ├── service             # Бизнес-логика
│   │   └── server              # gRPC server
│   ├── proto                   # gRPC proto-файлы
│   ├── migrations              # SQL-миграции
│   ├── Dockerfile
│   └── go.mod
├── api-service
│   ├── cmd/api                 # Точка входа
│   ├── internal
│   │   └── ...
│   ├── Dockerfile
│   └── go.mod
├── event-logger-service
│   ├── cmd/...
│   ├── Dockerfile
│   └── ...
├── docker-compose.yml
├── Taskfile.yml
└── README.md

````

---

## ⚙️ Конфигурация (env)

### db-service (cleanenv)

**PostgreSQL**
- `DB_HOST` (в Docker: `postgres`)
- `DB_PORT` (обычно `5432`)
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_SSLMODE` (обычно `disable`)
- `DB_TIMEOUT` (например `5s`)

**gRPC**
- `GRPC_HOST` (обычно `0.0.0.0`)
- `GRPC_PORT` (например `50051`)

**Redis (кеш задач)**
- `REDIS_ENABLED` (`true/false`)
- `REDIS_HOST` (в Docker: `redis`)
- `REDIS_PORT` (обычно `6379`)
- `REDIS_PASSWORD` (если нужен)
- `REDIS_DB` (обычно `0`)
- `REDIS_TTL` (например `5m`) — TTL кеша задач

### api-service

- `HTTP_HOST` (обычно `0.0.0.0`)
- `HTTP_PORT` (например `8080`)
- `GRPC_HOST` (в Docker: `db-service`)
- `GRPC_PORT` (например `50051`)
- (если используется Kafka) параметры брокера/топика из env

---

## 🚀 Быстрый старт (Docker)

### 🔹 Требования
- Docker
- Docker Compose

### 🔹 Запуск всех сервисов
```bash
docker compose up -d --build
````

Проверить логи:

```bash
docker compose logs -f db-service
docker compose logs -f api-service
```

После старта:

* **API**: `http://localhost:8080`
* **gRPC (db-service)**: `localhost:50051`

---

## ⚠️ ВАЖНО: volume Postgres НЕ УДАЛЯТЬ

У тебя есть volume с данными PostgreSQL (например `postgres_data`). Он хранит состояние БД.

✅ Безопасно остановить (volume останется):

```bash
docker compose down
```

❌ НЕЛЬЗЯ (удалит volume и данные):

```bash
docker compose down -v
docker volume prune
```

---

## 🧬 Генерация gRPC-кода

```bash
task proto
```

---

## 🧪 Тестирование

Тесты находятся в `db-service`.

```bash
task test
```

или вручную:

```bash
cd db-service
go test ./... -v
```

---

## 🔍 Ручное тестирование gRPC

Можно использовать:

* `grpcurl`
* BloomRPC
* Postman (gRPC mode)

Пример:

```bash
grpcurl -plaintext localhost:50051 list
```

---

## 📜 Реализованные методы (db-service)

* `CreateTask`
* `GetTaskByID`
* `GetAllTasks`
* `CompleteTask`
* `DeleteTask`

---

## 📌 Статус проекта

Проект в активной разработке.

Реализовано:

* db-service (PostgreSQL + Redis cache + миграции)
* api-service (HTTP → gRPC)
* Kafka/Zookeeper + event-logger-service

Идеи на будущее:

* метрики/трейсинг (Prometheus/OpenTelemetry)
* Kubernetes deployment
* CI/CD (GitHub Actions)

---

## 🖤 Лицензия

MIT

```
::contentReference[oaicite:0]{index=0}
```
