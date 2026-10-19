# LevelUp

[![CI](https://github.com/cdlinkin/todoapp/actions/workflows/ci.yml/badge.svg)](https://github.com/cdlinkin/todoapp/actions/workflows/ci.yml)

Учебный проект: todo-приложение на Go. Начиналось как простой CRUD, в итоге обросло авторизацией, тестами, Swagger, Docker и CI. Фронтенд — одна HTML-страница, её отдаёт сам сервер.

Что умеет:

- регистрация и вход (пароль хранится как bcrypt-хеш, дальше работа по JWT);
- задачи: создать, поправить, отметить выполненной, удалить. Каждый видит только свои;
- статистика: сколько создано и сделано, процент выполнения, среднее время на задачу, можно выбрать период;
- Swagger UI, чтобы потыкать API без Postman.

Стек: Go 1.25 (`net/http` без фреймворков), PostgreSQL 18 через `pgx`, миграции на `golang-migrate`, `golang-jwt`, `zap` для логов, `swag` для документации. Окружение поднимается через Docker Compose, команды собраны в Makefile.

## Как устроен код

Всё лежит в `internal/` и разбито по фичам (`users`, `tasks`, `statistics`, `auth`, `web`). Внутри каждой фичи три слоя: `transport/http` разбирает запрос, `service` отвечает за логику, `repository/postgres` ходит в базу. Общие штуки (логгер, пул соединений, middleware, JWT) лежат в `internal/core`.

Пара решений, которые здесь важны:

- Сервис зависит от интерфейса репозитория, а не от конкретной реализации. Интерфейс объявлен там, где он используется. Благодаря этому сервисы и хендлеры тестируются с фейками, без базы.
- У пользователя три структуры: DTO для JSON, доменная модель и модель для таблицы. Звучит как лишнее, но, например, хеш пароля просто не может уйти в ответ API: в DTO нет такого поля.
- Ошибки помечаются маркерами вроде `ErrNotFound` или `ErrConflict` и пробрасываются через `%w`. В HTTP-статус они превращаются в одном месте.
- Задачи и пользователи обновляются с проверкой `version`, так что одновременное изменение одной записи вернёт 409, а не перезатрёт чужие правки.

## Запуск

Понадобятся Go 1.25+, Docker и make. Makefile я писал под Windows: `levelup-run` и цели для очистки используют cmd/PowerShell, остальное просто вызывает `docker compose` и `go`.

Сначала `.env`:

```bash
cp .env.example .env
```

Пример заполнения:

```env
HTTP_ADDR=:5050
HTTP_SHUTDOWN_TIMEOUT=30s
ALLOWED_ORIGINS=http://localhost:5050,null

POSTGRES_USER=admin
POSTGRES_PASSWORD=change-me
POSTGRES_DB=todoapp
POSTGRES_TIMEOUT=10s
POSTGRES_PORT=9998

LOGGER_LEVEL=DEBUG

# минимум 32 байта, иначе приложение не стартует. Например: openssl rand -base64 32
AUTH_JWT_SECRET=
AUTH_TOKEN_TTL=24h
```

`POSTGRES_HOST` и `LOGGER_FOLDER` в `.env` не нужны: `make levelup-run` подставляет `localhost` и `./out/logs`, а для Docker они прописаны в `docker-compose.yaml`.

Дальше база, миграции и само приложение:

```bash
make env-up
make migrate-up
make levelup-run
```

`env-up` поднимает Postgres и пробрасывает его порт на `127.0.0.1:9998`. Если сразу после первого запуска `migrate-up` ругается на `connection refused`, значит база ещё инициализируется, подожди пару секунд.

После этого:

- приложение: http://localhost:5050/
- Swagger: http://localhost:5050/swagger/index.html

Можно запустить и полностью в Docker: вместо `make levelup-run` выполнить `make levelup-deploy`.

## API

Всё лежит под `/api/v1`. Без токена открыты только `/auth/register` и `/auth/login`, для остальных нужен заголовок `Authorization: Bearer <token>`.

```
POST   /auth/register
POST   /auth/login            -> access_token
GET    /auth/me

GET    /tasks?limit=&offset=
POST   /tasks
GET    /tasks/{id}
PATCH  /tasks/{id}
DELETE /tasks/{id}

GET    /statistics?from=YYYY-MM-DD&to=YYYY-MM-DD

GET    /users/{id}
PATCH  /users/{id}
DELETE /users/{id}
```

В PATCH поле можно не передавать (останется как было), передать значение или передать `null`, чтобы очистить.

Про доступ: на чужую задачу сервер отвечает 404, как будто её нет. На чужой профиль 403. При входе неверный пароль и несуществующий email дают одинаковый 401, чтобы по ответу нельзя было понять, зарегистрирован ли человек. Аккаунт с задачами удалить нельзя, будет 409.

Подробнее про запросы и ответы смотри в Swagger.

Быстрый пример через curl:

```bash
curl -X POST http://localhost:5050/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Ivan Ivanov","email":"ivan@mail.ru","password":"secret123"}'

curl -X POST http://localhost:5050/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ivan@mail.ru","password":"secret123"}'

curl http://localhost:5050/api/v1/tasks -H "Authorization: Bearer <token>"
```

## Тесты

```bash
make test
```

База для тестов не нужна. Покрыто:

- доменная логика: валидация, PATCH по принципу «всё или ничего», хеширование паролей;
- JWT: в том числе отказ на чужой секрет, просроченный токен и `alg: none`;
- auth-middleware;
- сервисы с фейковыми репозиториями, включая проверку, что чужие данные нельзя прочитать или изменить;
- HTTP-хендлеры через `httptest`.

## CI

На каждый PR и пуш в `main` GitHub Actions прогоняет `gofmt`, `go vet`, `go test -race`, сборку и сборку Docker-образа.

## Что ещё есть в Makefile

`make help` покажет все команды. Чаще всего нужны эти:

```
env-up / env-down        поднять / остановить базу
env-cleanup              остановить и удалить данные базы
migrate-up / migrate-down
migrate-create seq=name  новая миграция
levelup-run              запуск локально
levelup-deploy           запуск в Docker
test, vet, build
swagger-gen              пересобрать Swagger после правок в аннотациях
```
