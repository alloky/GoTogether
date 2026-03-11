# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoTogether is a meeting scheduling application with three independently deployable services:

- **`backend/`** — Go REST API (chi router, PostgreSQL via pgx, JWT auth, golang-migrate)
- **`frontend/`** — React + TypeScript SPA (Vite, Ant Design, React Query, React Router)
- **`tgbot/`** — Telegram bot (telebot.v3) that wraps the backend API
- **`e2e/`** — Playwright end-to-end tests (targets `http://localhost:3000`)
- **`tgbot-tests/`** — Jest integration tests for the Telegram bot (uses `telegram-test-api`)

All services are orchestrated via `docker-compose.yml`.

## Commands

### Backend (Go)
```bash
cd backend
go build ./...                    # Build
go test ./...                     # Run all tests
go test ./internal/handler/...    # Run a specific package's tests
go test -run TestName ./...       # Run a single test
```

### Frontend (Node)
```bash
cd frontend
npm run dev       # Dev server on http://localhost:3000 (via Vite)
npm run build     # TypeScript check + Vite production build
npm run preview   # Serve the production build locally
```

### E2E Tests (Playwright)
```bash
cd e2e
npm test              # Run all tests (headless)
npm run test:headed   # Run with browser visible
npm run test:ui       # Playwright UI mode
```

### Telegram Bot Tests (Jest)
```bash
cd tgbot-tests
npm test    # Run all Jest tests sequentially
```

### Full Stack
```bash
docker compose up --build    # Build and start all services
docker compose up db         # Start only PostgreSQL
```

## Architecture

### Backend layering (`backend/internal/`)
Follows a strict layered architecture:
- **`domain/`** — Core types (`Meeting`, `User`, `TimeSlot`, `Participant`, `Vote`) and repository interfaces. No dependencies on infrastructure.
- **`repository/postgres/`** — PostgreSQL implementations of the domain repository interfaces.
- **`service/`** — Business logic (`AuthService`, `MeetingService`). Depends on domain interfaces, not concrete repos.
- **`handler/`** — HTTP handlers, middleware, and router. Injects services. Auth middleware validates JWT and sets user context.
- **`config/`** — Loads config from environment variables with defaults.

Migrations live in `backend/migrations/` and are run automatically at startup via `golang-migrate`.

### Backend environment variables
| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | `postgres://gotogether:gotogether@localhost:5432/gotogether?sslmode=disable` | |
| `JWT_SECRET` | `dev-secret-change-me` | |
| `PORT` | `8080` | |
| `CORS_ORIGIN` | `http://localhost:3000` | |
| `MIGRATIONS_PATH` | `/migrations` | |
| `SMTP_HOST` | `` | Empty → falls back to log-only sender |
| `SMTP_PORT` | `465` | Use 465 (implicit TLS/SMTPS) or 587 (STARTTLS) |
| `SMTP_USER` | `` | Empty → no auth (local relay) |
| `SMTP_PASSWORD` | `` | |
| `SMTP_FROM` | `` | Sender address shown in emails |

### Email / SMTP
`backend/internal/service/email_service.go` has two send paths:
- **Port 465** — implicit TLS (SMTPS). Used for Gmail and most external providers.
- **Port 587/25** — plain connection upgraded via STARTTLS if the server advertises it.
- **No credentials** — auth is skipped entirely (suitable for local relays).

When `SMTP_HOST` is empty the backend uses `LogEmailSender`, which prints the code to stdout only.

All five SMTP vars are read from the root `.env` file (loaded via `env_file: .env` in docker-compose).

### API routes (`/api/`)
- `POST /auth/register`, `POST /auth/login` — public
- `GET /auth/me`, `GET /users/search` — JWT-protected
- `/meetings/` — full CRUD, plus `/confirm`, `/participants`, `/participants/rsvp`, `/votes`, `/tags` — all JWT-protected

### Frontend structure (`frontend/src/`)
- **`api/`** — Axios client + typed fetch functions for auth and meetings
- **`pages/`** — Top-level route components (Dashboard, CreateMeeting, MeetingDetail, Calendar, Login, Register)
- **`components/`** — Reusable UI components grouped by domain (`auth/`, `layout/`, `meetings/`)
- **`hooks/`** — Custom React hooks (likely wrapping React Query)
- **`context/`** — React context providers (likely auth context)

### Telegram Bot (`tgbot/internal/`)
- **`bot/`** — Bot struct, handler registration, `ConversationManager` (multi-step flows), `VoteStore`
- **`apiclient/`** — HTTP client that calls the backend API
- **`auth/`** — Auth manager for mapping Telegram users to backend JWT tokens
- **`config/`** — Loads `TELEGRAM_BOT_TOKEN`, `BACKEND_URL`, `JWT_SECRET` from environment

The bot requires `TELEGRAM_BOT_TOKEN` set (via `.env` file for docker-compose). For tests, `telegram-test-api` acts as a fake Telegram server.

### Key domain concepts
A `Meeting` has: organizer, participants (with RSVP status: `invited`/`accepted`/`declined`), time slots, votes per slot, tags, and a status (`pending`/`confirmed`/`cancelled`). The organizer confirms a meeting by selecting a winning time slot.
