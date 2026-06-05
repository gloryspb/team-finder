# Team Finder

Team Finder - учебный full-stack проект для курсовой работы на тему «Платформа для поиска команд в онлайн-игры».

В проекте реализована платформа, где игрок может зарегистрироваться, заполнить профиль, создать объявление о поиске команды, найти подходящие объявления через фильтры и отправить заявку на вступление. Владелец объявления может просматривать входящие заявки и принимать или отклонять их.

## Цель проекта

Цель проекта - разработать веб-приложение с клиент-серверной архитектурой, REST API, авторизацией пользователей и хранением данных в реляционной базе данных. Проект сделан как учебный пример full-stack приложения с разделением backend, frontend, базы данных, миграций и тестов.

## Используемый стек

- Backend: Go 1.22+, Echo.
- API: REST + JSON.
- Auth: JWT + bcrypt.
- Database: PostgreSQL.
- Migrations: Goose.
- Cache/service state: Redis.
- Frontend: React, TypeScript, Vite, React Router.
- Development/deployment: Docker, Docker Compose.
- Tests: Go unit tests и fuzz tests для `AuthService` и `ListingService`.

## Основная функциональность

- регистрация и вход пользователя;
- хранение JWT в `localStorage`;
- получение текущего пользователя и профиля;
- редактирование профиля игрока;
- список игр с режимами и ролями;
- создание объявлений о поиске команды;
- фильтрация объявлений по игре, роли, региону, режиму, статусу и поисковой строке;
- просмотр подробностей объявления;
- отправка заявки на объявление;
- запрет заявки на собственное объявление;
- запрет повторной заявки;
- просмотр входящих и исходящих заявок;
- принятие и отклонение заявок владельцем объявления;
- закрытие объявления.

## Запуск через Docker

```bash
cp .env.example .env
docker compose up --build
```

После запуска доступны:

- frontend: `http://localhost:5173`;
- backend API: `http://localhost:8080/api`;
- healthcheck: `http://localhost:8080/health`;
- PostgreSQL: `localhost:5432`;
- Redis: `localhost:6379`.

Backend автоматически применяет Goose-миграции при старте. Через миграцию также добавляются сидовые игры: Dota 2, Counter-Strike 2, Valorant, League of Legends, Overwatch 2 и Apex Legends.

## Запуск без Docker

Для запуска без Docker нужен локальный PostgreSQL. Redis необязателен: если Redis недоступен, backend продолжает работу и выводит предупреждение в лог.

Backend:

```powershell
cd backend
$env:DATABASE_URL="postgres://teamfinder:teamfinder@localhost:5432/teamfinder?sslmode=disable"
$env:REDIS_ADDR="localhost:6379"
$env:JWT_SECRET="dev-secret-change-me"
$env:ALLOW_ORIGINS="http://localhost:5173"
go run ./cmd/api
```

Frontend:

```powershell
cd frontend
npm install
$env:VITE_API_URL="http://localhost:8080/api"
npm run dev
```

## Миграции

В Docker миграции применяются автоматически при запуске backend.

Ручной запуск миграций из папки `backend`:

```bash
goose -dir migrations postgres "postgres://teamfinder:teamfinder@localhost:5432/teamfinder?sslmode=disable" up
goose -dir migrations postgres "postgres://teamfinder:teamfinder@localhost:5432/teamfinder?sslmode=disable" down
```

## Тестирование

Обычные unit-тесты:

```bash
cd backend
go test ./...
```

Fuzz-тестирование AuthService:

```bash
go test ./internal/services -run=^$ -fuzz=FuzzRegister -fuzztime=10s
go test ./internal/services -run=^$ -fuzz=FuzzLogin -fuzztime=10s
```

Fuzz-тестирование ListingService:

```bash
go test ./internal/services -run=^$ -fuzz=FuzzCreateListing -fuzztime=10s
go test ./internal/services -run=^$ -fuzz=FuzzApplicationStatus -fuzztime=10s
```

## API endpoints

Auth:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/me`
- `GET /api/me/profile`
- `PUT /api/me/profile`

Games:

- `GET /api/games`
- `POST /api/games` - только admin
- `PUT /api/games/:id` - только admin
- `DELETE /api/games/:id` - только admin

Listings:

- `GET /api/listings?game_id=&role=&region=&mode=&status=&search=`
- `GET /api/listings/:id`
- `POST /api/listings`
- `PUT /api/listings/:id`
- `PATCH /api/listings/:id/close`
- `DELETE /api/listings/:id`

Applications:

- `POST /api/listings/:id/applications`
- `GET /api/applications/outgoing`
- `GET /api/applications/incoming`
- `PATCH /api/applications/:id/status`

Ошибки API возвращаются в формате:

```json
{"error":"message"}
```

## Сценарии проверки

1. Зарегистрировать первого пользователя.
2. Заполнить или изменить профиль игрока.
3. Создать объявление о поиске команды.
4. Зарегистрировать второго пользователя.
5. Найти объявление первого пользователя через фильтры.
6. Отправить заявку на объявление от второго пользователя.
7. Войти под первым пользователем и открыть входящие заявки.
8. Принять или отклонить заявку.
9. Закрыть объявление и проверить изменение статуса.

## Развёртывание frontend на GitHub Pages

GitHub Pages подходит только для статического frontend. Backend на Go, PostgreSQL и Redis должны быть размещены отдельно, например на Render, Railway, Fly.io или VPS.

Для сборки frontend в GitHub Actions используется workflow:

```text
.github/workflows/deploy-frontend.yml
```

В переменной репозитория `VITE_API_URL` указывается публичный адрес backend API, например:

```text
https://example-backend.com/api
```

После публикации адрес frontend будет иметь вид:

```text
https://<username>.github.io/<repository>/
```

## Структура проекта

```text
/backend
  /cmd/api/main.go
  /internal/config
  /internal/domain
  /internal/ports
  /internal/services
  /internal/repositories/postgres
  /internal/handlers/http
  /internal/middleware
  /internal/validation
  /migrations
  Dockerfile
/frontend
  /src/api
  /src/components
  /src/pages
  /src/routes
  /src/types
  Dockerfile
/docker-compose.yml
/.env.example
/README.md
```
