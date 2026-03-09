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
# Edit monitoring/.env — fill in ALERT_BOT_TOKEN and ALERT_CHAT_ID
```

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
🔗 View in GlitchTip
```

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

## Next step: application instrumentation

Add to `backend` and `tgbot` (separate implementation task):
- `github.com/getsentry/sentry-go` + `sentrychi` middleware — point DSN at GlitchTip
- `go.opentelemetry.io/contrib/instrumentation/github.com/go-chi/chi/v5/otelchi`
- `github.com/exaring/otelpgx` for SQL trace spans
- OTel metric instruments for HTTP request counters and latency histograms

The `sentry-go` SDK works identically with GlitchTip — just use the GlitchTip DSN:
```go
sentry.Init(sentry.ClientOptions{
    Dsn: "http://<key>@localhost:8100/1",
})
```
