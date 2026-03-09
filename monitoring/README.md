# GoTogether Monitoring Stack

All monitoring services are integrated into the root `docker-compose.yml`.
Everything starts with a single `docker compose up -d`.

## Components

| Service | Port | Purpose |
|---|---|---|
| VictoriaMetrics | 8428 | Metrics storage (PromQL) + VMUI |
| VMAgent | 8429 | Scrapes exporters → VictoriaMetrics |
| VMAlert | 8880 | Evaluates alert rules |
| Alertmanager | 9093 | Routes alerts → Telegram group |
| Loki | 3100 | Log aggregation |
| Promtail | 9080 | Tails Docker container logs → Loki |
| Jaeger | 16686 | Distributed trace UI |
| OTel Collector | 4317/4318 | OTLP receiver → Jaeger + VictoriaMetrics + Loki |
| Grafana | 3001 | Unified dashboards |
| cAdvisor | 8081 | Container metrics |
| node-exporter | 9100 | Host OS metrics |
| postgres-exporter | 9187 | PostgreSQL metrics |
| GlitchTip | 8100 | Error tracking (Sentry-SDK compatible) |
| GlitchTip Redis | 6380 | GlitchTip task queue |
| sentry-relay | 9456 | GlitchTip webhook → Telegram |

## First-time setup

### 1. Configure secrets

```bash
cp monitoring/.env.example monitoring/.env
# Edit monitoring/.env — fill in ALERT_BOT_TOKEN, ALERT_CHAT_ID, and GLITCHTIP_SECRET_KEY
```

The `monitoring/.env` file is loaded by `alertmanager`, `glitchtip-web`, `glitchtip-worker`, and `sentry-relay`.

**Getting ALERT_CHAT_ID**:
1. Add your alert bot to the Telegram group
2. Send any message in the group
3. Visit: `https://api.telegram.org/bot<TOKEN>/getUpdates`
4. Find `"chat":{"id":...}` — that negative number is your chat ID

### 2. Start everything

```bash
docker compose up -d
```

That's it. The GlitchTip database is created automatically on first PostgreSQL init
via `monitoring/init-glitchtip-db.sql` (mounted into `/docker-entrypoint-initdb.d/`).

> **Note**: If your `db` volume already exists from a previous run, the init script
> will NOT re-execute (PostgreSQL only runs entrypoint scripts on fresh volumes).
> In that case, run manually:
> ```bash
> docker compose exec db psql -U gotogether -f /docker-entrypoint-initdb.d/10-glitchtip.sql
> ```

### 3. Verify everything is up

```bash
docker compose ps
```

### 4. Open Grafana

- URL: http://localhost:3001
- Login: `admin` / `gotogether`
- Datasources are auto-provisioned (VictoriaMetrics, Loki, Jaeger, Alertmanager)

### 5. Import community dashboards

In Grafana → Dashboards → Import:
- **Node Exporter Full**: ID `1860`
- **cAdvisor**: ID `14282`
- **PostgreSQL**: ID `9628`

Select `VictoriaMetrics` as the datasource when prompted.

### 6. Configure GlitchTip

1. Open GlitchTip: http://localhost:8100
2. Register an admin account (first user becomes admin)
3. Create an organization (e.g. `GoTogether`)
4. Create two projects:
   - `gotogether-backend` (platform: Go)
   - `gotogether-tgbot` (platform: Go)
5. Copy the **DSN** for each project — you'll use these when instrumenting the Go code:
   ```
   http://<key>@localhost:8100/1   (backend)
   http://<key>@localhost:8100/2   (tgbot)
   ```
6. Configure webhook alerts:
   - Go to **Settings → Projects → [project] → Alerts**
   - Add a **Webhook** alert with URL: `http://localhost:9456/webhook`
   - Set notification frequency (e.g. every event or every 10 minutes)

## Telegram alert message formats

### GlitchTip errors (via sentry-relay)

```
🆕 🔴 [CREATED] panic: runtime error: index out of range
📦 Project: gotogether-backend
📍 Culprit: handler.CreateMeeting
💬 Error: index out of range [3] with length 2
👁 Seen: 1 time(s)
🔗 View in Sentry
```

The link text reads "View in Sentry" (sentry-relay is compatible with both Sentry and GlitchTip webhook formats).

Level emojis: 💀 fatal · 🔴 error · 🟡 warning · 🔵 info · ⚪ other
Action emojis: 🆕 created · ✅ resolved · 👤 assigned · 🔕 ignored

### Alertmanager metric alerts

```
🔴 HighErrorRate
Severity: warning
📌 Backend HTTP 5xx rate above 1%
ℹ️ Error rate: 3.2%
⏰ Started: 2030-03-08 14:23:00 UTC
```

Resolved alerts show `✅`.

## Stopping

```bash
docker compose down
# To also remove volumes (clears all data):
docker compose down -v
```

## sentry-relay environment variables

`sentry-relay` reads these from `monitoring/.env` (via `env_file`) plus the compose `environment` block:

- **`ALERT_BOT_TOKEN`** — Telegram bot token (the alert bot, not the GoTogether bot)
- **`ALERT_CHAT_ID`** — Telegram group chat ID
- **`SENTRY_SECRET`** — optional; if set, validates the `Sentry-Hook-Signature` header on incoming webhooks
- **`GLITCHTIP_URL`** — base URL of the GlitchTip instance, used to build issue permalinks (default: `http://localhost:8100`)
- **`PORT`** — HTTP listen port (default: `9456`, set to `9456` in compose)

## Alert rules summary

### Metric alerts (VMAlert → Alertmanager)

Rules live in `monitoring/rules/alerts.yml` and are evaluated by VMAlert.

**containers** (interval 30s)
- `ContainerDown` (critical) — cAdvisor `container_last_seen` gap > 30s
- `ContainerRestartLoop` (warning) — restarted > 3 times in 5 min
- `ContainerHighCPU` (warning) — CPU > 80% for 5 min
- `ContainerHighMemory` (warning) — memory > 85% of limit for 5 min

**tgbot** (interval 30s)
- `TgbotDown` (critical) — tgbot container not seen for > 30s

**host** (interval 60s)
- `HostHighMemory` (warning) — host RAM > 85% for 5 min
- `HostDiskSpaceHigh` (warning) — disk `/` > 80% for 5 min
- `HostDiskSpaceCritical` (critical) — disk `/` > 95% for 2 min
- `HostHighLoad` (warning) — 5-min load average > 2 for 10 min

**backend** (interval 30s)
- `BackendHighErrorRate` (warning) — 5xx rate > 1% for 2 min
- `BackendCriticalErrorRate` (critical) — 5xx rate > 5% for 1 min
- `BackendHighLatency` (warning) — P95 latency > 2s for 5 min
- `BackendNoTraffic` (warning) — zero requests for 10 min
- `BackendDBPoolHigh` (warning) — pgx pool acquired connections > 90% of max for 2 min

**postgres** (interval 30s)
- `PostgresDown` (critical) — exporter cannot connect for 1 min
- `PostgresConnectionsHigh` (warning) — connections > 80% of max for 5 min
- `PostgresLongRunningQuery` (warning) — query running > 30s for 1 min
- `PostgresDeadlocks` (warning) — any deadlock in last 5 min
- `PostgresCheckpointsTooFrequent` (info) — requested checkpoint rate > 0.5/s for 5 min

**monitoring-health** (interval 60s)
- `VictoriaMetricsDown` (critical) — absent for 2 min
- `LokiDown` (warning) — absent for 2 min

### Loki log alerts (Loki ruler → Alertmanager)

Rules live in `monitoring/loki/rules/fake/alerts.yml` and are evaluated by the Loki ruler.

**log-alerts** (interval 30s)
- `LogFatalOrPanic` (critical) — log contains FATAL or panic in backend/tgbot/frontend/db
- `LogMigrationFailed` (critical) — backend log contains migration failure
- `LogBackendErrorRate` (warning) — more than 10 ERROR log lines per minute in backend
- `TgbotPollingErrors` (warning) — more than 3 Telegram polling errors in 5 min

### GlitchTip / Sentry alerts

GlitchTip handles error tracking alerts natively via webhook → `sentry-relay` → Telegram:
- **New issue detected** — fires immediately when `sentry-relay` receives a `created` webhook event
- Configure in GlitchTip: Settings → Project → Alerts → Webhook URL: `http://localhost:9456/webhook`

> **Known limitations:**
> - GlitchTip (open source) does not support issue frequency threshold alerts (e.g. >10/min).
>   Use the `LogBackendErrorRate` Loki alert as a substitute.
> - GlitchTip does not support anomaly detection or error-rate-spike alerts.
>   These require a paid Sentry plan or custom tooling.

## Backend metrics endpoint

The backend exposes Prometheus metrics at `GET /metrics` (unauthenticated, outside `/api`).

Metrics exported:
- `http_requests_total{method, path, status}` — HTTP request counter
- `http_request_duration_seconds{method, path}` — HTTP latency histogram (P50/P95/P99)
- `http_requests_in_flight` — current concurrent requests
- `pgx_pool_acquired_connections` — currently acquired DB connections
- `pgx_pool_idle_connections` — idle DB connections
- `pgx_pool_total_connections` — total DB connections in pool
- `pgx_pool_max_connections` — maximum pool size
- `pgx_pool_acquire_count_total` — cumulative successful acquires
- `pgx_pool_acquire_duration_seconds_total` — cumulative acquire wait time
- `pgx_pool_canceled_acquire_count_total` — acquires canceled by context
- `pgx_pool_empty_acquire_count_total` — acquires when pool was empty

VMAgent scrapes this endpoint via the `backend` job in `monitoring/prometheus.yml`.

## Next step: full OTel SDK + tracing + sentry-go

The backend currently exposes Prometheus metrics but does not send traces to Jaeger
or errors to GlitchTip. The next task is to add full OTel instrumentation and
sentry-go integration. See the copy-paste prompt below for this task.

The `sentry-go` SDK works identically with GlitchTip — just use the GlitchTip DSN:
```go
sentry.Init(sentry.ClientOptions{
    Dsn: "http://<key>@localhost:8100/1",
})
```
