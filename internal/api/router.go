package api

import (
	"net/http"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/store"
)

// NewRouter creates and configures the HTTP router
func NewRouter(st *store.Store, jwtManager *auth.JWTManager, corsOrigins string) *http.ServeMux {
	mux := http.NewServeMux()

	// Handlers
	authHandlers := NewAuthHandlers(st, jwtManager)
	jobHandlers := NewJobHandlers(st)

	// Health check (no auth required)
	mux.HandleFunc("GET /health", Health)

	// Setup endpoints (no auth required for initial setup)
	mux.HandleFunc("GET /setup/status", authHandlers.SetupStatus)
	mux.HandleFunc("POST /setup/admin", authHandlers.CreateFirstAdmin)

	// Auth endpoints (no auth required for login)
	mux.HandleFunc("POST /api/auth/login", authHandlers.Login)

	// Protected endpoints - wrap with auth middleware
	authMw := AuthMiddleware(jwtManager, st)

	// Jobs endpoints
	mux.Handle("GET /api/jobs", authMw(http.HandlerFunc(jobHandlers.ListJobs)))
	mux.Handle("POST /api/jobs", authMw(http.HandlerFunc(jobHandlers.CreateJob)))
	mux.Handle("GET /api/jobs/{id}", authMw(http.HandlerFunc(jobHandlers.GetJob)))

	// Apply CORS middleware
	corsMw := CORSMiddleware(corsOrigins)

	// Return wrapped mux with CORS and other global middleware
	wrappedMux := http.NewServeMux()
	wrappedMux.Handle("/", corsMw(mux))

	return wrappedMux
}
