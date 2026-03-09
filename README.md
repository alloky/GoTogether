# GoTogether

A meeting scheduling app that lets groups propose time slots, vote on availability, and confirm meetings. Available as a web app and a Telegram bot.

## Features

- Create meetings with multiple candidate time slots
- Invite participants and track RSVP status (accepted / declined)
- Vote on time slots; organizer confirms the winning slot
- Tag meetings and browse a public meeting calendar
- Full Telegram bot interface for creating and managing meetings

## Stack

- **Backend** — Go, chi, PostgreSQL (pgx), JWT, golang-migrate
- **Frontend** — React 18, TypeScript, Vite, Ant Design, React Query
- **Telegram bot** — Go, telebot.v3
- **E2E tests** — Playwright
- **Bot tests** — Jest + telegram-test-api
- **Monitoring** — VictoriaMetrics, VMAgent, VMAlert, Alertmanager, Loki, Promtail, Jaeger, OTel Collector, Grafana, GlitchTip, sentry-relay

## Getting Started

**Prerequisites:** Docker and Docker Compose.

1. Create the root `.env` with your Telegram bot token:
   ```bash
   echo "TELEGRAM_BOT_TOKEN=your_token_here" > .env
   ```

2. Create `monitoring/.env` for alert notifications (copy the example):
   ```bash
   cp monitoring/.env.example monitoring/.env
   # Edit monitoring/.env — fill in ALERT_BOT_TOKEN, ALERT_CHAT_ID, GLITCHTIP_SECRET_KEY
   ```

3. Start all services:
   ```bash
   docker compose up --build
   ```

Application endpoints:

- **Frontend** — http://localhost:3000
- **Backend API** — http://localhost:8080

Monitoring endpoints:

- **Grafana** — http://localhost:3001 (admin / gotogether)
- **GlitchTip** — http://localhost:8100
- **VictoriaMetrics UI** — http://localhost:8428
- **Jaeger UI** — http://localhost:16686

The backend runs database migrations automatically on startup. The GlitchTip PostgreSQL database is created automatically on first `db` volume init via `monitoring/init-glitchtip-db.sql`.

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

monitoring/
  prometheus.yml     # VMAgent scrape config
  alertmanager.yml   # Alert routing → Telegram
  otelcol-config.yml # OTel Collector pipelines (traces, metrics, logs)
  rules/             # VMAlert alert rules (containers, host, backend, postgres)
  loki/              # Loki config
  promtail/          # Promtail config (Docker container log collection)
  grafana/           # Auto-provisioned datasources and dashboard pointers
  sentry-relay/      # GlitchTip webhook → Telegram forwarder (Go service)
  init-glitchtip-db.sql  # Creates glitchtip DB on first postgres init
  .env.example       # Alert secrets template
```

## Environment Variables

### Backend (set in `docker-compose.yml` or shell)
- **`DATABASE_URL`** — `postgres://gotogether:gotogether@localhost:5432/gotogether?sslmode=disable`
- **`JWT_SECRET`** — `dev-secret-change-me`
- **`PORT`** — `8080`
- **`CORS_ORIGIN`** — `http://localhost:3000`

### Telegram Bot (`.env` in project root)
- **`TELEGRAM_BOT_TOKEN`** — required
- **`BACKEND_URL`** — defaults to `http://127.0.0.1:8080`
- **`JWT_SECRET`** — must match backend

### Monitoring (`monitoring/.env`)
- **`ALERT_BOT_TOKEN`** — Telegram bot token for alert notifications (different from the GoTogether bot)
- **`ALERT_CHAT_ID`** — Telegram group chat ID to receive alerts (negative number)
- **`GLITCHTIP_SECRET_KEY`** — random secret for GlitchTip Django session signing
- **`SENTRY_SECRET`** — optional shared secret for webhook signature validation

## Monitoring

See [monitoring/README.md](monitoring/README.md) for full setup and configuration details.
