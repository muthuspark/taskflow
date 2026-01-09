package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
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

const pidFileName = "taskflow.pid"

func main() {
	// Handle service commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "start":
			startDaemon()
			return
		case "stop":
			stopDaemon()
			return
		case "status":
			checkStatus()
			return
		case "help", "-h", "--help":
			printUsage()
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}

	// Check if already running (unless we're the daemon process)
	isDaemon := os.Getenv("TASKFLOW_DAEMON") == "1"
	if !isDaemon {
		if pid, err := readPID(); err == nil && isProcessRunning(pid) {
			log.Fatalf("TaskFlow is already running (PID: %d). Use './taskflow stop' first.", pid)
		}
	}

	// Always write PID file (both foreground and daemon modes)
	if err := writePIDFile(); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
	defer removePIDFile()

	// Load configuration
	cfg := config.Load()

	// Auto-generate JWT secret if not provided
	if cfg.JWTSecret == "" {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			log.Fatalf("Failed to generate JWT secret: %v", err)
		}
		cfg.JWTSecret = hex.EncodeToString(secret)
		log.Println("Warning: JWT_SECRET not set, generated random secret. Sessions will not persist across restarts.")
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
	router := api.NewRouter(db, jwtManager, wsHub, cfg.AllowedOrigins, sched, cfg.APIBasePath)
	apiBasePath := cfg.APIBasePath

	// Create main handler that combines router and file server
	mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Route API, health, and setup requests to the router
		if isAPIPath(r.URL.Path, apiBasePath) {
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
func isAPIPath(path string, apiBasePath string) bool {
	return path == "/health" ||
		strings.HasPrefix(path, "/taskflow-app/") ||
		path == "/setup/status" ||
		strings.HasPrefix(path, "/setup/") ||
		strings.HasPrefix(path, apiBasePath+"/")
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

// getPIDFilePath returns the path to the PID file
func getPIDFilePath() string {
	// Use executable directory for PID file
	execPath, err := os.Executable()
	if err != nil {
		return pidFileName
	}
	return filepath.Join(filepath.Dir(execPath), pidFileName)
}

// writePIDFile writes the current process ID to the PID file
func writePIDFile() error {
	pidPath := getPIDFilePath()
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0644)
}

// removePIDFile removes the PID file
func removePIDFile() {
	os.Remove(getPIDFilePath())
}

// readPID reads the PID from the PID file
func readPID() (int, error) {
	pidPath := getPIDFilePath()
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

// isProcessRunning checks if a process with the given PID is running
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0 to check
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// startDaemon starts TaskFlow as a background daemon
func startDaemon() {
	// Check if already running
	if pid, err := readPID(); err == nil {
		if isProcessRunning(pid) {
			fmt.Printf("TaskFlow is already running (PID: %d)\n", pid)
			os.Exit(1)
		}
		// Stale PID file, remove it
		removePIDFile()
	}

	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Start the process in background
	cmd := exec.Command(execPath)
	cmd.Env = append(os.Environ(), "TASKFLOW_DAEMON=1")

	// Detach from terminal
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start the process
	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start TaskFlow: %v\n", err)
		os.Exit(1)
	}

	// Don't wait for the child process
	fmt.Printf("TaskFlow started (PID: %d)\n", cmd.Process.Pid)
}

// stopDaemon stops the running TaskFlow daemon
func stopDaemon() {
	pid, err := readPID()
	if err != nil {
		fmt.Println("TaskFlow is not running (no PID file found)")
		os.Exit(1)
	}

	if !isProcessRunning(pid) {
		fmt.Printf("TaskFlow is not running (stale PID: %d)\n", pid)
		removePIDFile()
		os.Exit(1)
	}

	// Send SIGTERM to gracefully stop
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process: %v\n", err)
		os.Exit(1)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to stop TaskFlow: %v\n", err)
		os.Exit(1)
	}

	// Wait for process to stop (with timeout)
	fmt.Printf("Stopping TaskFlow (PID: %d)...\n", pid)
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if !isProcessRunning(pid) {
			fmt.Println("TaskFlow stopped")
			return
		}
	}

	fmt.Println("TaskFlow did not stop in time, sending SIGKILL...")
	process.Signal(syscall.SIGKILL)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("TaskFlow killed")
}

// checkStatus checks if TaskFlow is running
func checkStatus() {
	pid, err := readPID()
	if err != nil {
		fmt.Println("TaskFlow is not running")
		os.Exit(1)
	}

	if isProcessRunning(pid) {
		fmt.Printf("TaskFlow is running (PID: %d)\n", pid)
	} else {
		fmt.Printf("TaskFlow is not running (stale PID file: %d)\n", pid)
		removePIDFile()
		os.Exit(1)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("TaskFlow - Lightweight Task Scheduler")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  taskflow          Start TaskFlow in foreground")
	fmt.Println("  taskflow start    Start TaskFlow as a background daemon")
	fmt.Println("  taskflow stop     Stop the running TaskFlow daemon")
	fmt.Println("  taskflow status   Check if TaskFlow is running")
	fmt.Println("  taskflow help     Show this help message")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  PORT              HTTP listen port (default: 8080)")
	fmt.Println("  DB_PATH           SQLite database path (default: taskflow.db)")
	fmt.Println("  JWT_SECRET        JWT signing secret (auto-generated if not set)")
	fmt.Println("  API_BASE_PATH     API base path (default: /api)")
	fmt.Println("  LOG_RETENTION_DAYS  Days to keep run logs (default: 30)")
	fmt.Println("  ALLOWED_ORIGINS   CORS allowed origins (default: *)")
}
