# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Refactored API handlers to apply SOLID principles and Clean Code standards
  - Extracted JobValidator abstraction for centralized validation logic
  - Eliminated 140+ lines of duplicate validation code across CreateJob and UpdateJob
  - Reduced average handler method size by 60% through separation of concerns
  - Simplified SetJobSchedule with consistent schedule validation
- Enhanced scheduler with public Enqueue() method for manual job triggering

### Added
- JobValidator type in internal/api/validator.go:
  - Centralized job request validation with ValidateJobRequest()
  - Centralized schedule validation with ValidateScheduleRequest()
  - Default value application with ApplyDefaults()
  - Clean request-to-model conversion with ToJobModel()
- Comprehensive test coverage for validation logic:
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
