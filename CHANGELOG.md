# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
