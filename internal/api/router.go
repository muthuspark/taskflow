package api

import (
	"net/http"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/store"
)

// NewRouter creates and configures the HTTP router
func NewRouter(st *store.Store, jwtManager *auth.JWTManager, wsHub *WSHub, corsOrigins string) *http.ServeMux {
	mux := http.NewServeMux()

	// Handlers
	authHandlers := NewAuthHandlers(st, jwtManager)
	jobHandlers := NewJobHandlers(st)
	runHandlers := NewRunHandlers(st)
	scheduleHandlers := NewScheduleHandlers(st)
	dashboardHandlers := NewDashboardHandlers(st)

	// Middleware
	authMw := AuthMiddleware(jwtManager, st)
	corsMw := CORSMiddleware(corsOrigins)
	bodyLimitMw := RequestBodyLimitMiddleware(internal.MaxRequestBodySize)

	// Health check (no auth required)
	mux.HandleFunc("GET /health", Health)

	// Setup endpoints (no auth required for initial setup)
	mux.HandleFunc("GET /setup/status", authHandlers.SetupStatus)
	mux.Handle("POST /setup/admin", bodyLimitMw(http.HandlerFunc(authHandlers.CreateFirstAdmin)))

	// Auth endpoints (no auth required for login)
	mux.Handle("POST /api/auth/login", bodyLimitMw(http.HandlerFunc(authHandlers.Login)))

	// Protected endpoints - wrap with auth middleware
	// Jobs endpoints
	mux.Handle("GET /api/jobs", authMw(http.HandlerFunc(jobHandlers.ListJobs)))
	mux.Handle("POST /api/jobs", bodyLimitMw(authMw(http.HandlerFunc(jobHandlers.CreateJob))))
	mux.Handle("GET /api/jobs/{id}", authMw(http.HandlerFunc(jobHandlers.GetJob)))
	mux.Handle("PUT /api/jobs/{id}", bodyLimitMw(authMw(http.HandlerFunc(jobHandlers.UpdateJob))))
	mux.Handle("DELETE /api/jobs/{id}", authMw(http.HandlerFunc(jobHandlers.DeleteJob)))
	mux.Handle("POST /api/jobs/{id}/run", authMw(http.HandlerFunc(jobHandlers.TriggerJob)))

	// Schedule endpoints
	mux.Handle("GET /api/jobs/{id}/schedule", authMw(http.HandlerFunc(scheduleHandlers.GetJobSchedule)))
	mux.Handle("PUT /api/jobs/{id}/schedule", bodyLimitMw(authMw(http.HandlerFunc(scheduleHandlers.SetJobSchedule))))

	// Runs endpoints
	mux.Handle("GET /api/runs", authMw(http.HandlerFunc(runHandlers.ListRuns)))
	mux.Handle("GET /api/runs/{id}", authMw(http.HandlerFunc(runHandlers.GetRun)))
	mux.Handle("GET /api/runs/{id}/logs", authMw(http.HandlerFunc(runHandlers.GetRunLogs)))

	// Dashboard endpoints
	mux.Handle("GET /api/dashboard/stats", authMw(http.HandlerFunc(dashboardHandlers.GetStats)))

	// WebSocket endpoints (no auth middleware applied here - handler manages auth internally)
	mux.HandleFunc("GET /api/ws/logs", wsHub.HandleLogsWebSocket)

	// Return wrapped mux with CORS and other global middleware
	wrappedMux := http.NewServeMux()
	wrappedMux.Handle("/", corsMw(mux))

	return wrappedMux
}
