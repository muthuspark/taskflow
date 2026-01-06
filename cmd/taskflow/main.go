package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)

	// Initialize scheduler and executor
	sched := scheduler.New(db)
	exec := executor.New(db)

	// Create HTTP router
	router := api.NewRouter(db, jwtManager, cfg.AllowedOrigins)

	// Create WebSocket hub with CORS validation
	wsHub := api.NewWSHub(cfg.AllowedOrigins)
	go wsHub.Run()

	// Add WebSocket handler
	http.HandleFunc("GET /api/runs/{id}/logs/live", wsHub.HandleLogsWebSocket)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start scheduler
	go func() {
		jobHandler := func(job *store.Job) error {
			// Create a new run
			run, err := db.CreateRun(job.ID, "scheduled")
			if err != nil {
				log.Printf("Failed to create run: %v\n", err)
				return err
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
