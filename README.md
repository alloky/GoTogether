# GoTogether

A meeting scheduling app that lets groups propose time slots, vote on availability, and confirm meetings. Available as a web app and a Telegram bot.

## Features

- Create meetings with multiple candidate time slots
- Invite participants and track RSVP status (accepted / declined)
- Vote on time slots; organizer confirms the winning slot
- Tag meetings and browse a public meeting calendar
- Full Telegram bot interface for creating and managing meetings

## Stack

| Layer | Technology |
|---|---|
| Backend | Go, chi, PostgreSQL, JWT, golang-migrate |
| Frontend | React 18, TypeScript, Vite, Ant Design, React Query |
| Telegram bot | Go, telebot.v3 |
| E2E tests | Playwright |
| Bot tests | Jest + telegram-test-api |

## Getting Started

**Prerequisites:** Docker and Docker Compose.

1. Copy the example env file and add your Telegram bot token:
   ```bash
   echo "TELEGRAM_BOT_TOKEN=your_token_here" > .env
   ```

2. Start all services:
   ```bash
   docker compose up --build
   ```

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |

The backend runs database migrations automatically on startup.

## Development

### Backend
```bash
cd backend
go test ./...        # run tests
go build ./...       # build
```

### Frontend
```bash
cd frontend
npm run dev          # dev server on http://localhost:3000
npm run build        # production build
```

### E2E Tests
Requires the full stack running (`docker compose up`).
```bash
cd e2e
npm test
```

### Telegram Bot Tests
```bash
cd tgbot-tests
npm test
```

## Architecture

```
backend/
  cmd/server/        # entrypoint
  internal/
    domain/          # core types and repository interfaces
    repository/      # PostgreSQL implementations
    service/         # business logic
    handler/         # HTTP handlers and router
  migrations/        # SQL migrations (auto-applied at startup)

frontend/src/
  api/               # typed API client (axios + React Query)
  pages/             # route-level components
  components/        # reusable UI components

tgbot/internal/
  bot/               # handlers, conversation flows, vote store
  apiclient/         # backend HTTP client
  auth/              # Telegram ↔ backend JWT mapping
```

## Environment Variables

### Backend
| Variable | Default |
|---|---|
| `DATABASE_URL` | `postgres://gotogether:gotogether@localhost:5432/gotogether?sslmode=disable` |
| `JWT_SECRET` | `dev-secret-change-me` |
| `PORT` | `8080` |
| `CORS_ORIGIN` | `http://localhost:3000` |

### Telegram Bot
| Variable | Required |
|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes |
| `BACKEND_URL` | Defaults to `http://127.0.0.1:8080` |
| `JWT_SECRET` | Must match backend |
