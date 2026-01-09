package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/taskflow/taskflow/internal/api"
	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/config"
	"github.com/taskflow/taskflow/internal/executor"
	"github.com/taskflow/taskflow/internal/scheduler"
	"github.com/taskflow/taskflow/internal/store"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Validate required config
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Initialize database
	log.Printf("Initializing database at %s\n", cfg.DBPath)
	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v\n", err)
	}
	defer db.Close()

	// Create default admin user if no users exist
	count, err := db.UserCount()
	if err == nil && count == 0 {
		hash, err := auth.HashPassword("password")
		if err == nil {
			_, err = db.CreateUser("admin", "admin@localhost", hash, "admin")
			if err != nil {
				log.Printf("Warning: Failed to create default admin user: %v\n", err)
			} else {
				log.Println("Created default admin user with credentials admin/password")
			}
		}
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize scheduler and executor
	sched := scheduler.New(db)
	exec := executor.New(db)

	// Create WebSocket hub with CORS validation
	wsHub := api.NewWSHub(cfg.AllowedOrigins)
	go wsHub.Run()

	// Wire up executor to broadcast logs and status via WebSocket
	exec.SetLogBroadcaster(func(runID string, stream string, content string, timestamp time.Time) {
		wsHub.Broadcast(api.WSMessage{
			Type:      "log",
			RunID:     runID,
			Timestamp: timestamp.Format(time.RFC3339),
			Data: map[string]string{
				"stream":  stream,
				"content": content,
			},
		})
	})
	exec.SetStatusBroadcaster(func(runID string, status string) {
		wsHub.Broadcast(api.WSMessage{
			Type:      "status",
			RunID:     runID,
			Timestamp: time.Now().Format(time.RFC3339),
			Data: map[string]string{
				"status": status,
			},
		})
	})

	// Create HTTP router (pass wsHub and scheduler for job processing)
	router := api.NewRouter(db, jwtManager, wsHub, cfg.AllowedOrigins, sched)

	// Create main handler that combines router and file server
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route API, health, and setup requests to the router
		if isAPIPath(r.URL.Path) {
			router.ServeHTTP(w, r)
			return
		}

		// Serve frontend files or SPA index.html
		path := r.URL.Path
		if path == "" {
			path = "/"
		}

		if path == "/" {
			http.ServeFile(w, r, "web/frontend/dist/index.html")
			return
		}

		filePath := "web/frontend/dist" + path
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// For SPA, serve index.html for non-asset routes
		if !isAssetPath(path) {
			http.ServeFile(w, r, "web/frontend/dist/index.html")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mainHandler,
	}

	// Start scheduler
	go func() {
		jobHandler := func(job *store.Job, existingRun *store.Run) error {
			var run *store.Run
			var err error

			// Use existing run if provided (manual triggers), otherwise create new one (scheduled)
			if existingRun != nil {
				run = existingRun
			} else {
				run, err = db.CreateRun(job.ID, "scheduled")
				if err != nil {
					log.Printf("Failed to create run: %v\n", err)
					return err
				}
			}

			// Execute the job
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(job.TimeoutSeconds)*time.Second)
			defer cancel()

			return exec.Execute(ctx, run, job)
		}

		if err := sched.Start(context.Background(), jobHandler); err != nil {
			log.Printf("Failed to start scheduler: %v\n", err)
		}
	}()

	// Start cleanup service (delete old runs daily)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := db.DeleteOldRuns(cfg.LogRetentionDays); err != nil {
				log.Printf("Failed to cleanup old runs: %v\n", err)
			}
		}
	}()

	// Start server in background
	go func() {
		log.Printf("Starting TaskFlow on %s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sched.Stop()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	}

	log.Println("Shutdown complete")
}

// isAPIPath checks if a path should be handled by the API router
func isAPIPath(path string) bool {
	return path == "/health" ||
		path == "/setup/status" ||
		strings.HasPrefix(path, "/setup/") ||
		strings.HasPrefix(path, "/api/") ||
		strings.HasPrefix(path, "/ws/")
}

// isAssetPath checks if a path is likely an asset file
func isAssetPath(path string) bool {
	if path == "/" || path == "/index.html" {
		return true
	}
	// Use filepath.Ext() for idiomatic extension checking with map lookup
	validExts := map[string]bool{
		".js": true, ".css": true, ".png": true,
		".svg": true, ".jpg": true, ".jpeg": true,
	}
	return validExts[filepath.Ext(path)]
}
