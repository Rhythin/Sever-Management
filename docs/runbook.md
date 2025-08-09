# Runbook: Server Management Backend

## 1. Setup

### Local Development
- Clone repo, install Go, Docker, Air (optional)
- `make swag` to generate Swagger docs
- `make run` or `make air-local` for hot-reload

### Docker Compose
- `make docker-up` to start app + Postgres
- `make air-docker` for Docker hot-reload
- DB credentials: see `docker-compose.yml`

### Database
- Auto-migration on startup (see `internal/persistence/db.go`)
- Schema: [../schema.sql](../schema.sql)

## 2. Monitoring
- **Health:** `/healthz`, `/readyz`
- **Logs:** Structured, with request ID (see stdout or Docker logs)
- **Swagger:** `/swagger/index.html` for API docs

## 3. Alerting
- Set up Prometheus alert rules for:
  - High error rates (5xx responses)
  - High latency
  - DB connection failures
  - App not ready/healthy
- Use log aggregation (e.g., Loki, ELK) to alert on error logs

## 4. Failure Modes & Recovery
- **DB Down:** App will log and retry; check Postgres container/logs
- **App Crash:** Use `docker-compose restart app` or `make docker-up`
- **Migration Fail:** See logs for errors; check schema and DB connectivity
- **Idle Reaper/Billing Fail:** See logs for errors; restart app if needed
- **IP Allocation Fail:** DB constraint violation; check for IP exhaustion

## 5. Operational Notes
- Use `make` targets for all common operations
- For local dev, use Air for fast feedback
- For production, use Docker Compose or Kubernetes

---
