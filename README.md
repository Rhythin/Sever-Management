# Server Management Backend

A backend service simulating virtual server lifecycles—provisioning, state transitions, billing, logging, observability—without managing real infrastructure.

## Features
- Virtual server FSM: provision, start, stop, reboot, terminate
- Uptime-based billing, idle reaper, metrics, logs
- REST API with OpenAPI/Swagger docs
- Structured logging with request ID
- Postgres persistence, atomic IP allocation
- Docker Compose for local dev
- Hot-reload (Air) for Go and Docker workflows

## Setup

### Prerequisites
- Go 1.22+
- Docker & Docker Compose
- (Optional) [Air](https://github.com/cosmtrek/air) for hot-reload

### Local Go Dev
```sh
git clone <repo>
cd Sever-Management
make swag
make run
# or hot-reload:
make air-local
```

### Docker Compose Dev
```sh
make docker-up
# or hot-reload:
make air-docker
```

### Tests & Lint
```sh
make test
make lint
```

### Swagger Docs
- Generate: `make swag`
- View: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## API Spec

| Method | Path                  | Description                        |
|--------|-----------------------|------------------------------------|
| POST   | /server               | Provision a new server             |
| GET    | /servers/{id}         | Get server metadata                |
| POST   | /servers/{id}/action  | Start, stop, reboot, terminate     |
| GET    | /servers              | List servers (filter, paginate)    |
| GET    | /servers/{id}/logs    | Last 100 lifecycle events          |
| GET    | /metrics              | Prometheus metrics                 |
| GET    | /healthz, /readyz     | Health/readiness endpoints         |

See [Swagger UI](http://localhost:8080/swagger/index.html) for full details.

## Design Notes
- **Clean architecture:** cmd/, internal/api, internal/domain, internal/persistence, internal/service
- **Dependency injection:** Uber fx
- **Structured logging:** zap, request ID middleware
- **Atomic IP allocation:** DB transaction, unique constraint
- **Observability:** Prometheus, structured logs, request tracing
- **Schema:** See [schema.sql](./schema.sql)
- **Runbook:** See [docs/runbook.md](./docs/runbook.md)

---
MIT License