# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## TaskFlow Overview

TaskFlow is a lightweight, self-hosted task scheduler with:
- **Backend**: Go with SQLite, single binary deployment
- **Frontend**: Vue.js SPA with Vite
- **Core Features**: Job scheduling (cron-like), script execution, real-time logs via WebSocket, CPU/memory metrics, JWT auth with admin/user roles

## Common Commands

### Building and Running

```bash
# Build both frontend and backend
make build

# Run the built binary (generates random JWT_SECRET)
make run

# Development mode with hot-reload
make dev           # Terminal 1: Backend server on :8080
make dev-frontend  # Terminal 2: Frontend dev server (optional)

# Frontend only
make build-frontend
make dev-frontend

# Backend only
make build-backend
make dev

# Clean all artifacts
make clean
```

### Testing

```bash
# Run all tests (uses in-memory SQLite for isolation)
make test

# Run specific test package
go test -v ./internal/auth

# Run single test function
go test -v ./internal/auth -run TestHashPassword

# Run tests with coverage
go test -cover ./...
```

## Architecture Overview

### Core Data Flow

1. **User Request** → API handlers (internal/api) → Store operations (internal/store)
2. **Job Scheduling** → Scheduler checks every minute (internal/scheduler/scheduler.go:71-82) → Matcher validates cron rules → JobQueue enqueues
3. **Job Execution** → Executor spawns bash subprocess → Streams logs to database → Updates run status with exit code
4. **Real-time Logs** → WebSocket hub (internal/api/websocket.go) → Client subscriptions

### Request Processing Pipeline

```
HTTP Request
    ↓
CORSMiddleware (internal/api/middleware.go)
    ↓
AuthMiddleware (validates JWT, loads user context)
    ↓
Route Handler (internal/api/handlers.go)
    ↓
Store Operations (internal/store/*.go)
    ↓
Database (SQLite)
```

### Job Lifecycle

```
Job Created → Scheduled by Scheduler → Enqueued → Executed (one at a time)
                                                    ↓
                                            Logs streamed to DB
                                                    ↓
                                            Metrics captured
                                                    ↓
                                            Run record updated
```

### Key Components

#### 1. **Store** (internal/store/sqlite.go)
- Central database abstraction
- Methods for CRUD operations on jobs, runs, users, logs, metrics
- Automatic migrations on startup
- Uses `sql.DB` for connection pooling

#### 2. **Scheduler** (internal/scheduler/scheduler.go)
- Runs every 60 seconds (time.Minute ticker)
- Calls Matcher to check if job should run based on schedule
- Prevents duplicate runs in same minute
- Uses JobQueue for sequential execution
- Single-threaded: only one job runs at a time

#### 3. **Executor** (internal/executor/executor.go)
- Spawns bash subprocess for each job
- Respects timeout context (ctx.WithTimeout)
- Streams stdout/stderr to database
- Captures exit code and determines success/failure/timeout
- Called directly by job handler in main.go:63-76

#### 4. **API Router** (internal/api/router.go)
- Standard Go http.ServeMux
- Handlers grouped: auth (login, setup), jobs (CRUD), runs (history, logs)
- WebSocket endpoint for live log streaming
- CORS and Auth middleware

#### 5. **Authentication** (internal/auth/jwt.go, password.go)
- JWT signing with HS256
- Bootstrap mode: first user creation without auth (setup endpoint disabled after)
- Role-based access control: admin can create jobs, users can only run their own

### Database Schema (Automatic via Migrations)

```
users           - User accounts with roles (admin/user)
jobs            - Job definitions (script, timeout, retry settings)
schedules       - Cron-like scheduling rules (flexible year/month/day/hour/minute)
runs            - Execution history (status, exit code, timestamps)
logs            - Streaming job output (stdout, stderr, system messages)
metrics         - CPU/memory samples during execution (2s interval)
metrics_aggregate - Hourly/daily aggregated statistics
settings        - Key-value configuration store
```

## Important Architectural Decisions

### Sequential Job Execution
Jobs run one at a time in FIFO order (see CanExecute in executor.go:154-164). This is by design for Phase 1 to avoid resource contention on single-server deployments.

### Bootstrap Mode
On first run with no users, the `/setup/admin` endpoint allows creating the first admin without JWT auth. After first user is created, this endpoint becomes disabled. See handlers.go for bootstrap logic.

### Scheduler Prevents Duplicates
The scheduler checks if a job already ran in the same minute (scheduler.go:108-120) to prevent accidental duplicate executions when the clock rolls over.

### Log Retention
Runs older than 30 days (configurable via LOG_RETENTION_DAYS) are automatically deleted daily at cleanup time. See main.go:83-93 for cleanup goroutine.

### Timezone Support
Jobs have a timezone field. The Matcher respects timezones when evaluating cron schedules. See internal/scheduler/matcher.go for logic.

## Configuration

All config via environment variables (loaded in internal/config/config.go):

```
PORT=8080                      # HTTP listen port
DB_PATH=taskflow.db            # SQLite file path
JWT_SECRET=<required>          # HMAC secret for JWT signing
LOG_LEVEL=info                 # Logging verbosity
LOG_RETENTION_DAYS=30          # Delete runs older than this
ALLOWED_ORIGINS=*              # CORS allowed origins
SMTP_SERVER/PORT/USERNAME/PASSWORD  # Optional email notifications
```

## Testing Strategy

- Tests use in-memory SQLite (":memory:" connection string) for speed and isolation
- Helper: testutil.go provides setup/teardown utilities
- Test packages: auth_test.go, middleware_test.go, handlers_test.go, executor_test.go, matcher_test.go
- Run with `make test` or `go test -v ./...`

## Common Tasks

### Adding a New API Endpoint
1. Define handler in internal/api/handlers.go
2. Register route in internal/api/router.go with appropriate middleware
3. Add store method in internal/store/*.go if needed
4. Write tests in internal/api/handlers_test.go

### Adding a New Store Operation
1. Add SQL query in relevant file (internal/store/{jobs,runs,logs,users,metrics}.go)
2. Add test in that same file
3. Call from handlers

### Debugging Job Execution
1. Check logs: query logs table for run_id
2. Check metrics: query metrics table for run_id
3. Check run status: query runs table for status/error_message
4. Enable LOG_LEVEL=debug for verbose logging

### Modifying Scheduler Logic
- Scheduler checks jobs every minute (time.Minute ticker)
- Matcher determines if cron schedule matches current time (see matcher_test.go for examples)
- Avoid blocking operations in scheduler main loop

## Frontend Location

Vue.js SPA: web/frontend/
- Vite build tool
- npm install && npm run build (integrated in make build)
- Development: npm run dev on separate port
- API client: web/frontend/src/services/api.js
- State management: Pinia stores in web/frontend/src/stores/

## Performance Notes

- Single-threaded job execution prevents resource contention
- In-memory SQLite caching improves query performance
- WebSocket streaming reduces log polling overhead
- Metrics sampled at 2-second intervals during job execution (can be adjusted in executor.go)

## Debugging Tips

1. Enable debug logging: `LOG_LEVEL=debug`
2. Inspect database: `sqlite3 taskflow.db`
3. Check scheduler state: See IsRunning() in scheduler.go
4. Monitor job queue: See JobQueue in scheduler/queue.go
5. Watch WebSocket connections: api/websocket.go maintains active client list
