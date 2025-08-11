# Server Management Backend

A robust backend service that simulates virtual server lifecycles including provisioning, state management, billing, logging, and observability without the need for actual infrastructure.

## ✨ Features

- **Virtual Server Management**: Full lifecycle support (provision, start, stop, reboot, terminate)
- **Billing System**: Automated uptime-based billing with configurable rates
- **Resource Optimization**: Idle instance detection and reaping
- **Monitoring**: Built-in metrics and structured logging
- **REST API**: Fully documented with OpenAPI/Swagger
- **Persistence**: PostgreSQL with atomic IP allocation
- **DevOps Ready**: Docker Compose setup and CI/CD friendly
- **Developer Experience**: Hot-reload support with Air

## 🚀 Quick Start

### Prerequisites
- Go 1.22+ ([Installation Guide](https://golang.org/doc/install))
- Docker & Docker Compose ([Installation Guide](https://docs.docker.com/get-docker/))
- (Optional) [Air](https://github.com/cosmtrek/air) for hot-reload

### Local Development

#### Using Go
```bash
git clone <repo>
cd Sever-Management

# Install dependencies
go mod download

# Generate Swagger docs
make swag

# Start the application
make run

# Or run with hot-reload
make air-local
```

#### Using Docker
```bash
# Start all services (app + dependencies)
make docker-up

# Or run with hot-reload in Docker
make air-docker
```

## 🧪 Testing

```bash
# Run unit tests
make test

# Run linters
make lint

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## 📚 API Documentation

### Interactive Documentation
- Generate: `make swag`
- View: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### API Endpoints

#### Server Management
- `POST /server` - Provision a new server
- `GET /servers` - List all servers
- `GET /servers/{id}` - Get server details
- `POST /servers/{id}/action` - Perform an action on a server (start/stop/reboot/terminate)
- `GET /servers/{id}/logs` - Retrieve server logs

#### System Health
- `GET /healthz` - Health check endpoint
- `GET /readyz` - Readiness check endpoint
- `GET /metrics` - Prometheus metrics endpoint

## 🏗️ Architecture

### Core Components

1. **API Layer**
   - RESTful endpoints using Chi router
   - Request validation and error handling
   - Structured logging with request IDs
   - Metrics collection

2. **Service Layer**
   - Business logic implementation
   - Server state management
   - Billing calculations
   - Background workers (idle reaper, billing daemon)

3. **Persistence**
   - PostgreSQL for data storage
   - Atomic operations for IP allocation
   - Connection pooling and retries

4. **Observability**
   - Structured JSON logging
   - Prometheus metrics
   - Request tracing

### Data Flow
1. Request → API Handler → Service → Repository → Database
2. Background workers monitor and update server states
3. Metrics and logs are emitted throughout the process

## 🔧 Configuration

Configuration is managed through environment variables with sensible defaults:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=server_management
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# Billing
BILLING_RATE=0.01  # per minute
IDLE_TIMEOUT=30    # minutes
```

## 📦 Deployment

### Docker
```bash
docker build -t server-management .
docker run -p 8080:8080 server-management
```


- Built with Go 
- Uses Chi for routing
- PostgreSQL for persistence
- Prometheus for metrics



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