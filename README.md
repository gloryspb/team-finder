# Team Finder

Учебный full-stack проект по теме «Платформа для поиска команд в онлайн-игры». Приложение позволяет регистрироваться, вести профиль игрока, создавать объявления о поиске команды, фильтровать объявления и отправлять заявки.

## Стек

- Backend: Go 1.22+, Echo, REST API, JWT, bcrypt.
- Database: PostgreSQL, миграции Goose.
- Service state/cache: Redis подключается при старте.
- Frontend: React, TypeScript, Vite, React Router.
- Dev/deploy: Docker и Docker Compose.
- Тесты: Go unit tests и fuzz tests для `AuthService` и `ListingService`.

## Запуск

```bash
cp .env.example .env
docker compose up --build
```

После запуска:

- Frontend: http://localhost:5173
- Backend API: http://localhost:8080/api
- Healthcheck: http://localhost:8080/health
- PostgreSQL: localhost:5432
- Redis: localhost:6379

Backend применяет Goose-миграции автоматически при старте и добавляет сидовые игры: Dota 2, Counter-Strike 2, Valorant, League of Legends, Overwatch 2, Apex Legends.

## Развёртывание frontend на GitHub Pages

GitHub Pages разворачивает только статический frontend. Backend Go, PostgreSQL и Redis нужно разместить отдельно, например на Render, Railway, Fly.io или VPS. После публикации backend укажите его URL в переменной репозитория `VITE_API_URL`, например `https://your-api.example.com/api`.

Шаги:

1. Создайте репозиторий на GitHub и отправьте код в ветку `main`.
2. В настройках репозитория откройте `Settings -> Pages`.
3. В `Build and deployment` выберите источник `GitHub Actions`.
4. В `Settings -> Secrets and variables -> Actions -> Variables` добавьте `VITE_API_URL`.
5. Сделайте push в `main`. Workflow `.github/workflows/deploy-frontend.yml` соберёт `frontend` и опубликует сайт.

Адрес сайта будет вида `https://<username>.github.io/<repository>/`.

## Миграции

Автоматически в Docker:

```bash
docker compose up --build
```

Вручную из папки `backend`, если установлен `goose`:

```bash
goose -dir migrations postgres "postgres://teamfinder:teamfinder@localhost:5432/teamfinder?sslmode=disable" up
goose -dir migrations postgres "postgres://teamfinder:teamfinder@localhost:5432/teamfinder?sslmode=disable" down
```

## Тесты

```bash
cd backend
go test ./...
```

Fuzz-тесты запускаются стандартным Go:

```bash
go test ./internal/services -fuzz=FuzzRegister
go test ./internal/services -fuzz=FuzzCreateListing
go test ./internal/services -fuzz=FuzzApplicationStatus
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
- `POST /api/games` admin
- `PUT /api/games/:id` admin
- `DELETE /api/games/:id` admin

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

Ошибки возвращаются в формате:

```json
{"error":"message"}
```

## Тестовые сценарии

1. Зарегистрировать пользователя A и создать объявление.
2. Выйти, зарегистрировать пользователя B.
3. Найти объявление пользователя A через фильтры игры, роли, региона или режима.
4. Открыть подробности объявления и отправить заявку от пользователя B.
5. Войти пользователем A, открыть «Профиль и заявки».
6. Принять или отклонить входящую заявку.
7. Закрыть объявление и убедиться, что оно сменило статус.

## Структура

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
