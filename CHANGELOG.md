# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.0] - 2026-01-12

### Added
- Email notifications on job completion
  - Sends emails when jobs succeed, fail, or timeout based on job settings
  - Configure recipients and trigger conditions per job (always/success/failure)
  - Supports SMTP with TLS (port 465) and STARTTLS (ports 25, 587)
  - Notification fields added to job create and edit forms
  - Debug logging for notification delivery troubleshooting
- Test SMTP settings button in Account Settings
  - Sends test email to admin's email address to verify SMTP configuration
  - `POST /api/settings/smtp/test` endpoint
- User email update functionality
  - Edit email address in Account Settings page
  - `PUT /api/auth/email` endpoint

### Improved
- Added logging documentation to README
  - Application logs (stdout/stderr, debug mode, systemd)
  - Job execution logs (database storage, web UI viewing)

## [0.6.1] - 2026-01-11

### Fixed
- Error messages not displaying in Account Settings page (password change and SMTP settings)
  - Frontend was incorrectly accessing `error.message` when API returns `error` as a string

## [0.6.0] - 2026-01-11

### Added
- SMTP configuration in admin settings
  - New "SMTP Configuration" section in Account Settings page (admin only)
  - Configure SMTP server, port, username, password, from name, and from email
  - Settings stored in database, overriding environment variables
  - Password masked in UI for security
  - `GET /api/settings/smtp` endpoint to retrieve settings
  - `PUT /api/settings/smtp` endpoint to update settings
  - New `internal/store/settings.go` with generic key-value store methods

## [0.5.0] - 2026-01-11

### Added
- Full-page log view for printing
  - New `/runs/:id/logs` route opens logs in a dedicated print-friendly page
  - "Open Logs in New Tab" button in run detail header
  - "Full Page Logs" button in run detail sidebar
  - Print button triggers browser print dialog
  - Print-optimized CSS hides UI chrome (header, navigation, footer) when printing
- Account settings page with password change
  - Click username in navigation header to access `/account` page
  - Displays account information (username, email, role)
  - Password change form with current password verification
  - Minimum 6 character validation for new passwords
- Backend password change API
  - `PUT /api/auth/password` endpoint (authenticated)
  - Verifies current password before allowing change
  - Uses bcrypt for secure password hashing

### Added
- Analytics dashboard with execution trends and job duration analysis
  - New `/analytics` page accessible from main navigation
  - Overview stats: total runs, success rate, runs in last 24h, average duration
  - Execution trends chart with toggle between success rate and run counts view
  - Configurable time range (7, 14, 30, 90 days)
  - Job statistics table showing per-job performance metrics (success rate, avg/min/max duration)
  - Job duration trends chart for analyzing individual job performance over time
  - Color-coded success rates (green >90%, yellow >70%, red <70%)
- Backend analytics API endpoints:
  - `GET /api/analytics/overview` - overall system statistics
  - `GET /api/analytics/execution-trends` - daily success/failure/timeout counts
  - `GET /api/analytics/job-stats` - per-job performance statistics
  - `GET /api/analytics/jobs/{id}/duration-trends` - duration trends for specific job
- Chart.js integration for data visualization (Line charts with vue-chartjs)

### Removed
- CPU/Memory metrics collection feature (removed due to unreliable child process tracking)
  - Removed MetricsGauge and MetricsPanel components
  - Removed metrics.go from executor package
  - Removed gopsutil dependency
  - Note: metrics database schema retained for potential future use

- GitHub Actions workflow for automated releases (`.github/workflows/release.yml`)
  - Triggers on semantic version tags (v*.*.*)
  - Cross-compiles binaries for Linux, macOS, and Windows (AMD64 + ARM64)
  - Runs tests before release, generates checksums
  - Auto-generates release notes from commits
- GitHub Pages landing page (`index.html`) for project showcase
  - Clean, minimal design matching the application's W3Techs-inspired style
  - Hero section with project description and download/GitHub links
  - Screenshot preview, feature highlights, and quick-start guide
  - Configuration reference table and technology stack overview
  - Fully responsive design for mobile devices
  - Single-file HTML with embedded CSS for easy deployment
- TaskFlow UI screenshot in README.md for visual project preview
  - High-resolution PNG image (2662x1600) showcasing the W3Techs-inspired interface
  - Helps users understand the application appearance before installation

### Changed
- Complete frontend UI redesign from lo-fi monochrome to W3Techs-inspired style
  - Replaced ultra-minimal black/white Tailwind design with cleaner traditional interface
  - New color palette using soft blues (#bcd4ec, #99c2e5) for improved visual hierarchy
  - Switched from Source Code Pro monospace to Verdana font family (13px base)
  - Simplified navigation with centered links, gradient top border, and featured banner
  - Tables now use alternating row colors and blue header backgrounds
  - Status badges display color-coded backgrounds (green/red/yellow) for clear status
  - Cards and boxes feature light blue headers with gray borders
  - Removed Tailwind CSS in favor of vanilla CSS for smaller bundle size
- Updated StatusBadge component with simplified class mappings for new design system
- All view components (Dashboard, Jobs, Runs, JobCreate, JobDetail, RunDetail, Login) adapted for new design

### Added
- Service-style daemon commands for production deployments:
  - `./taskflow` - Start in foreground (useful for Docker/debugging)
  - `./taskflow start` - Start as background daemon with PID file
  - `./taskflow stop` - Gracefully stop running daemon
  - `./taskflow status` - Check if TaskFlow is running
  - `./taskflow help` - Show usage information
- Auto-generated JWT secret when `JWT_SECRET` env var is not set
  - Generates secure 32-byte random secret at startup
  - Warning logged that sessions won't persist across restarts
- Runtime API base path configuration via `/taskflow-app/config` endpoint
  - Frontend fetches config at startup, no rebuild needed
  - Useful for reverse proxy deployments at custom subpaths
  - Set `API_BASE_PATH` env var on backend only
- Schedule selection during job creation with preset options:
  - Daily at 9 AM (default), Hourly, Daily at midnight
  - Weekdays at 9 AM, Weekly on Monday, Monthly on 1st
  - Custom schedule with minute/hour/weekday selectors
  - Cron expression preview
- Working directory field in job create and edit forms
- Schedule can now be included in job create/update API requests

### Changed
- Default schedule for new jobs is "Daily at 9 AM" instead of every minute (prevents immediate execution)
- Job creation now rolls back if schedule save fails (data consistency)
- Job update returns error if schedule update fails

### Improved (Code Quality)
- Replaced direct error comparisons with `errors.Is()` for sql.ErrNoRows
- Replaced type assertions with `errors.As()` for exec.ExitError extraction
- Replaced manual time field comparison with `time.Truncate().Equal()`
- Replaced manual extension loop with `filepath.Ext()` and map lookup
- Replaced `strings.Split(" ")` with `strings.Fields()` for Bearer token parsing
- Replaced manual slice loop with `slices.ContainsFunc()` for origin validation
- Replaced chained `||` checks with map-based lookup for enum validation
- Added explicit error handling for JSON marshal operations in SetJobSchedule

## [0.3.0] - 2026-01-07

### Changed
- Redesigned jobs listing interface from card grid to data table
  - Replaced JobCard component with table-based layout for better information density
  - Added visible columns for job metadata (name, description, status, timeout, retries, created date)
  - Improved discoverability of job details without requiring modal/detail view navigation
  - Refined login page styling with improved visual hierarchy (reduced border weight)
- Fixed API response data parsing in frontend services
  - jobs.js: Updated to extract jobs from response.data.data.jobs
  - runs.js: Updated to extract runs from response.data.data.runs and logs from response.data.data.logs
- Updated font stack for improved readability
  - Added Source Code Pro Google Font integration
  - Refined typography across all pages

### Added
- New StatusBadge component for consistent job status display
- Date formatting helper in JobsView for consistent date representation
- Preconnect directives in index.html for Google Fonts optimization
- Table styling with proper column widths and hover states

### Fixed
- Schedule field name mismatches between API and frontend (days_of_month → days, days_of_week → weekdays)
- Modal overlay styling (removed opacity for cleaner appearance)

### Refactored API handlers to apply SOLID principles and Clean Code standards
- Extracted JobValidator abstraction for centralized validation logic
- Eliminated 140+ lines of duplicate validation code across CreateJob and UpdateJob
- Reduced average handler method size by 60% through separation of concerns
- Simplified SetJobSchedule with consistent schedule validation
- Enhanced scheduler with public Enqueue() method for manual job triggering

### Test Coverage
- Added comprehensive test coverage for validation logic:
  - 4 new test functions with 31+ new assertions
  - 100% code coverage for JobValidator
  - Tests for happy path, error cases, and edge cases

## [0.2.1] - 2026-01-07

### Fixed
- Resolve critical API errors in run retrieval endpoints (GET /api/runs/{id}, GET /api/runs)
  - Fixed NULL handling in Run.ErrorMsg field (changed from string to *string)
  - Fixed SQL query construction bug in ListRuns() with fragmented WHERE/pagination clauses
  - Fixed dashboard stats endpoint (GET /api/dashboard/stats) cascading failure
- Improved API handler robustness with run ID validation in GetRun()

### Changed
- Refactored job execution status determination into dedicated finalizeRun() helper method
  - Reduces Execute() method complexity by 33% (120 → 80 lines)
  - Isolates status/exit code/error message logic for better maintainability
- Extracted nullable field conversion logic into populateRunPointers() helper
  - Eliminates 30 lines of duplicate code in GetRun() and ListRuns()
  - Single source of truth for NULL-to-pointer conversion pattern

### Added
- Comprehensive test suite for refactored code:
  - 12 edge case tests for populateRunPointers() (NULL handling, various field types, time values)
  - 10 focused tests for finalizeRun() (success, timeout, failure scenarios, duration calculation)
  - All tests pass with 100% success rate (104+ total tests)

## [0.2.0] - 2026-01-06

### Added
- Complete backend implementation (Go) with all core services:
  - API handlers with JWT authentication and role-based access control (admin/user)
  - SQLite store layer with automatic schema migrations and connection pooling
  - Job scheduler that evaluates cron-like expressions every 60 seconds
  - Executor for script validation and subprocess management with timeout support
  - WebSocket hub for real-time log streaming to connected clients
  - Middleware for CORS protection and request authentication
- Complete frontend implementation (Vue.js with Vite):
  - Responsive SPA for job management and execution monitoring
  - Real-time log viewer with WebSocket integration
  - API client abstraction layer and Pinia state management
  - Metrics dashboard with CPU/memory visualization
- Build and deployment infrastructure:
  - Makefile with targets for development and production builds
  - Docker support with Dockerfile for containerized deployments
  - Single-binary backend with environment variable configuration
- Developer tools and documentation:
  - commi_all.sh script for simplified commit workflow
  - Example scripts for common operational tasks (backup, health checks)
  - Product requirements document (prd.txt)

### Architecture
- Single-threaded sequential job execution (FIFO queue) to prevent resource contention
- Bootstrap mode allowing first admin creation without authentication
- Scheduler duplicate prevention via state checking within same minute
- Automatic log retention with configurable cleanup (default 30 days)
- Timezone-aware cron schedule matching

## [0.1.0] - 2026-01-06

### Added
- Initial TaskFlow implementation with complete feature set
  - Job scheduling with cron-like syntax
  - Script execution with timeout support
  - Real-time log streaming via WebSocket
  - CPU/memory metrics tracking
  - JWT authentication with role-based access control
  - SQLite database with automatic migrations

### Security
- Authentication bypass prevention (proper user ID conversion)
- Command injection mitigation (script validation and size limits)
- WebSocket CSRF protection (origin validation)
- Comprehensive input validation for all API endpoints
- Null pointer safety in error handling

### Documentation
- CLAUDE.md: Development guidance for Claude Code
- Architecture overview and data flow documentation
- API endpoint documentation
- Database schema documentation
- Deployment instructions
