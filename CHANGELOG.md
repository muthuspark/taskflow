# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive test suite covering all critical functionality (1,223 lines of test code)
  - Authentication module tests: 11 tests with 87% coverage (password hashing, JWT generation/validation)
  - API handlers tests: 14 tests validating input validation and all job creation bounds
  - API middleware tests: 9 tests for auth middleware and CORS protection
  - Executor tests: 6 tests for script validation and execution limits
  - Scheduler tests: 13 tests for cron-like pattern matching and edge cases
- Test infrastructure helper: in-memory SQLite database factory for test isolation
- TEST_REPORT.md: Comprehensive documentation of all 57 test cases and coverage metrics

### Changed
- Improved test database setup with proper in-memory SQLite instances instead of custom mocks
- Updated test helpers to use NewTestStore() for realistic database testing
- Enhanced test isolation and cleanup patterns across all test modules

### Fixed
- Fixed Mock Store type incompatibility in handlers, executor, and middleware tests
- Fixed JWT token validation test logic (corrected dot counting for 3-part tokens)
- Fixed error message validation in script size limit tests
- Removed unused imports and ensured clean test builds

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
