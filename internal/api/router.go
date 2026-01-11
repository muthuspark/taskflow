package api

import (
	"net/http"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/scheduler"
	"github.com/taskflow/taskflow/internal/store"
)

// NewRouter creates and configures the HTTP router
func NewRouter(st *store.Store, jwtManager *auth.JWTManager, wsHub *WSHub, corsOrigins string, sched *scheduler.Scheduler, apiBasePath string) *http.ServeMux {
	mux := http.NewServeMux()

	// Handlers
	authHandlers := NewAuthHandlers(st, jwtManager)
	jobHandlers := NewJobHandlers(st, sched)
	runHandlers := NewRunHandlers(st)
	scheduleHandlers := NewScheduleHandlers(st)
	dashboardHandlers := NewDashboardHandlers(st)
	analyticsHandlers := NewAnalyticsHandlers(st)

	// Middleware
	authMw := AuthMiddleware(jwtManager, st)
	corsMw := CORSMiddleware(corsOrigins)
	bodyLimitMw := RequestBodyLimitMiddleware(internal.MaxRequestBodySize)

	// Health check (no auth required)
	mux.HandleFunc("GET /health", Health)

	// Config endpoint (no auth required) - provides runtime config to frontend
	// Uses /taskflow-app prefix to avoid conflicts with other services behind nginx
	mux.HandleFunc("GET /taskflow-app/config", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{
			"api_base_path": apiBasePath,
		})
	})

	// Setup endpoints (no auth required for initial setup)
	mux.HandleFunc("GET /setup/status", authHandlers.SetupStatus)
	mux.Handle("POST /setup/admin", bodyLimitMw(http.HandlerFunc(authHandlers.CreateFirstAdmin)))

	// Auth endpoints (no auth required for login)
	mux.Handle("POST "+apiBasePath+"/auth/login", bodyLimitMw(http.HandlerFunc(authHandlers.Login)))

	// Auth endpoints (requires auth)
	mux.Handle("PUT "+apiBasePath+"/auth/password", bodyLimitMw(authMw(http.HandlerFunc(authHandlers.ChangePassword))))

	// Protected endpoints - wrap with auth middleware
	// Jobs endpoints
	mux.Handle("GET "+apiBasePath+"/jobs", authMw(http.HandlerFunc(jobHandlers.ListJobs)))
	mux.Handle("POST "+apiBasePath+"/jobs", bodyLimitMw(authMw(http.HandlerFunc(jobHandlers.CreateJob))))
	mux.Handle("GET "+apiBasePath+"/jobs/{id}", authMw(http.HandlerFunc(jobHandlers.GetJob)))
	mux.Handle("PUT "+apiBasePath+"/jobs/{id}", bodyLimitMw(authMw(http.HandlerFunc(jobHandlers.UpdateJob))))
	mux.Handle("DELETE "+apiBasePath+"/jobs/{id}", authMw(http.HandlerFunc(jobHandlers.DeleteJob)))
	mux.Handle("POST "+apiBasePath+"/jobs/{id}/run", authMw(http.HandlerFunc(jobHandlers.TriggerJob)))

	// Schedule endpoints
	mux.Handle("GET "+apiBasePath+"/jobs/{id}/schedule", authMw(http.HandlerFunc(scheduleHandlers.GetJobSchedule)))
	mux.Handle("PUT "+apiBasePath+"/jobs/{id}/schedule", bodyLimitMw(authMw(http.HandlerFunc(scheduleHandlers.SetJobSchedule))))

	// Runs endpoints
	mux.Handle("GET "+apiBasePath+"/runs", authMw(http.HandlerFunc(runHandlers.ListRuns)))
	mux.Handle("GET "+apiBasePath+"/runs/{id}", authMw(http.HandlerFunc(runHandlers.GetRun)))
	mux.Handle("GET "+apiBasePath+"/runs/{id}/logs", authMw(http.HandlerFunc(runHandlers.GetRunLogs)))

	// Dashboard endpoints
	mux.Handle("GET "+apiBasePath+"/dashboard/stats", authMw(http.HandlerFunc(dashboardHandlers.GetStats)))

	// Analytics endpoints
	mux.Handle("GET "+apiBasePath+"/analytics/overview", authMw(http.HandlerFunc(analyticsHandlers.GetOverallStats)))
	mux.Handle("GET "+apiBasePath+"/analytics/execution-trends", authMw(http.HandlerFunc(analyticsHandlers.GetExecutionTrends)))
	mux.Handle("GET "+apiBasePath+"/analytics/job-stats", authMw(http.HandlerFunc(analyticsHandlers.GetJobStats)))
	mux.Handle("GET "+apiBasePath+"/analytics/jobs/{id}/duration-trends", authMw(http.HandlerFunc(analyticsHandlers.GetJobDurationTrends)))

	// WebSocket endpoints (no auth middleware applied here - handler manages auth internally)
	mux.HandleFunc("GET "+apiBasePath+"/ws/logs", wsHub.HandleLogsWebSocket)

	// Return wrapped mux with CORS and other global middleware
	wrappedMux := http.NewServeMux()
	wrappedMux.Handle("/", corsMw(mux))

	return wrappedMux
}
