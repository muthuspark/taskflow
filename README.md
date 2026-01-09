# TaskFlow - Task Scheduler & Runner

A lightweight, self-hosted task scheduler with a Go backend, SQLite storage, and Vue.js web UI.

## Features

- **Go Backend**: Single binary, easy deployment
- **SQLite Storage**: Zero external dependencies
- **Web UI**: Real-time updates, live logs, resource monitoring
- **Job Scheduling**: Cron-like scheduling with visual multi-selector
- **Task Execution**: Sequential job execution with shell scripts
- **Resource Tracking**: CPU and memory monitoring during execution
- **User Authentication**: JWT-based auth with admin/user roles
- **Bootstrap Mode**: First-time setup creates admin account

## Quick Start

### Requirements

- Go 1.22+
- Node.js 18+ (for building frontend)
- Make

### Build & Run

```bash
# Build frontend and backend
make build

# Run in foreground (useful for debugging)
./bin/taskflow

# Or run as a background daemon
./bin/taskflow start

# Access the UI at http://localhost:8080
```

### Service Commands

TaskFlow supports service-style commands for daemon management:

```bash
./taskflow          # Start in foreground
./taskflow start    # Start as background daemon
./taskflow stop     # Stop the running daemon
./taskflow status   # Check if TaskFlow is running
./taskflow help     # Show help message
```

When started with `start`, TaskFlow:
- Runs in the background
- Creates a PID file (`taskflow.pid`) in the executable directory
- Can be stopped gracefully with `./taskflow stop`

### Development Mode

Terminal 1 - Backend:
```bash
make dev
```

Terminal 2 - Frontend (optional for hot-reload):
```bash
make dev-frontend
# Then in your browser, set API endpoint to http://localhost:8080
```

## Configuration

Set environment variables:

```bash
export PORT=8080                    # HTTP port (default: 8080)
export DB_PATH=/path/to/taskflow.db # Database path (default: taskflow.db)
export JWT_SECRET=your-secret-key   # JWT signing secret (auto-generated if not set)
export API_BASE_PATH=/api           # API base path (default: /api)
export LOG_LEVEL=info               # Log level: debug, info, warn, error
export ALLOWED_ORIGINS=*            # CORS origins (default: *)
export LOG_RETENTION_DAYS=30        # Days to keep run logs (default: 30)

# Optional: SMTP for email notifications
export SMTP_SERVER=smtp.example.com
export SMTP_PORT=587
export SMTP_USERNAME=user@example.com
export SMTP_PASSWORD=password
```

**Notes:**
- If `JWT_SECRET` is not set, a random secret is generated at startup. This means user sessions won't persist across restarts. For production, set a fixed secret.
- `API_BASE_PATH` is a runtime configuration. The frontend fetches it from `/taskflow-app/config` at startup, so you only need to set it on the backend. This is useful when deploying behind a reverse proxy (e.g., nginx) at a subpath like `/taskflow/api`.

## Database

The application automatically creates and initializes SQLite on first run. The database includes:

- **users**: User accounts with roles
- **jobs**: Job definitions
- **schedules**: Cron-like scheduling rules
- **runs**: Job execution history
- **logs**: Streaming job logs
- **metrics**: CPU/memory usage during execution
- **metrics_aggregate**: Hourly/daily aggregated metrics
- **settings**: Application settings

## API Endpoints

### System (No Auth Required)

- `GET /health` - Health check
- `GET /taskflow-app/config` - Get runtime configuration (API base path)
- `GET /setup/status` - Check if setup is needed
- `POST /setup/admin` - Create first admin user

### Authentication (No Auth Required)

- `POST /api/auth/login` - Login and get JWT token

### Jobs (Auth Required)

- `GET /api/jobs` - List jobs
- `POST /api/jobs` - Create job (admin only)
- `GET /api/jobs/:id` - Get job details
- `PUT /api/jobs/:id` - Update job
- `DELETE /api/jobs/:id` - Delete job
- `POST /api/jobs/:id/run` - Trigger manual execution

### Runs

- `GET /api/runs` - List execution history
- `GET /api/runs/:id` - Get run details
- `GET /api/runs/:id/logs` - Get logs (HTTP)
- `WS /api/runs/:id/logs/live` - Stream logs (WebSocket)

### Metrics

- `GET /api/jobs/:id/metrics` - Historical metrics for job
- `GET /api/runs/:id/metrics` - Metrics for specific run
- `GET /api/dashboard/stats` - System statistics

## Project Structure

```
taskflow/
├── cmd/taskflow/              # Entry point
├── internal/
│   ├── api/                   # HTTP handlers and middleware
│   ├── auth/                  # JWT and password hashing
│   ├── config/                # Configuration
│   ├── executor/              # Job execution
│   ├── scheduler/             # Scheduling logic
│   ├── store/                 # Database operations
│   └── timezone/              # Timezone utilities
├── web/
│   ├── embed.go               # Static file embedding
│   └── frontend/              # Vue.js SPA
│       ├── src/
│       │   ├── components/    # Vue components
│       │   ├── stores/        # Pinia state management
│       │   ├── services/      # API client
│       │   ├── App.vue        # Root component
│       │   └── main.js        # Entry point
│       ├── vite.config.js
│       └── package.json
├── go.mod                     # Go dependencies
├── Makefile                   # Build commands
└── README.md
```

## Core Concepts

### Job Execution Model

Jobs execute sequentially (one at a time) in a FIFO queue. This prevents resource contention and simplifies single-server deployments.

### Bootstrap Mode

On first run with no users, the app enters bootstrap mode. A setup endpoint creates the first admin account without authentication. After the first user is created, the setup endpoint is disabled.

### User Roles

| Feature | Admin | User |
|---------|-------|------|
| Create jobs | ✅ | ❌ |
| Edit own jobs | ✅ | ✅ |
| Edit others' jobs | ✅ | ❌ |
| Run jobs | ✅ | ✅ |
| View own logs | ✅ | ✅ |
| Manage users | ✅ | ❌ |

### Data Retention

- Runs and logs older than 30 days are automatically deleted
- Cleanup runs daily at midnight UTC
- Configurable via `LOG_RETENTION_DAYS` environment variable

## Testing

Run the test suite:

```bash
make test
```

Tests use in-memory SQLite for speed and isolation.

## Deployment

### Native Binary (Recommended)

```bash
# Build the binary
make build

# Copy binary and set up data directory
mkdir -p /opt/taskflow/data
cp bin/taskflow /opt/taskflow/
cd /opt/taskflow

# Start as daemon
./taskflow start

# Check status
./taskflow status

# Stop when needed
./taskflow stop
```

### Using systemd (Optional)

For automatic startup on boot, create `/etc/systemd/system/taskflow.service`:

```ini
[Unit]
Description=TaskFlow Task Scheduler
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/taskflow
ExecStart=/opt/taskflow/taskflow
Environment=DB_PATH=/opt/taskflow/data/taskflow.db
Environment=JWT_SECRET=your-production-secret
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then enable and start:
```bash
sudo systemctl enable taskflow
sudo systemctl start taskflow
```

### Docker

A Dockerfile can be created for containerized deployment.

## Performance Notes

- Single-threaded job execution prevents resource exhaustion
- In-memory SQLite caching improves query performance
- WebSocket streaming reduces log polling overhead
- Metrics sampled at 2-second intervals during job execution

## Limitations & Future Work

- **Phase 1**: No built-in rate limiting (deploy behind Nginx for production)
- **Phase 2**: E2E tests with Playwright, advanced scheduling features
- **Phase 2**: Redis caching layer, multi-server deployments
- **Phase 2**: Backup/restore utilities, audit logging

## Contributing

See the PRD for detailed implementation guidelines.

## License

MIT
