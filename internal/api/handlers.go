package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/scheduler"
	"github.com/taskflow/taskflow/internal/store"
)

// AuthHandlers handles authentication endpoints
type AuthHandlers struct {
	store      *store.Store
	jwtManager *auth.JWTManager
}

// NewAuthHandlers creates auth handlers
func NewAuthHandlers(st *store.Store, jwtManager *auth.JWTManager) *AuthHandlers {
	return &AuthHandlers{store: st, jwtManager: jwtManager}
}

// Login handles POST /api/auth/login
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	user, err := h.store.GetUserByUsername(req.Username)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Invalid credentials", "INVALID_CREDENTIALS")
		return
	}

	// Verify password
	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		WriteError(w, http.StatusUnauthorized, "Invalid credentials", "INVALID_CREDENTIALS")
		return
	}

	// Update last login
	if err := h.store.UpdateUserLastLogin(user.ID); err != nil {
		log.Printf("Failed to update last login: %v\n", err)
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Role, 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// SetupStatus handles GET /setup/status
func (h *AuthHandlers) SetupStatus(w http.ResponseWriter, r *http.Request) {
	count, err := h.store.UserCount()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to check setup status", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"needs_setup": count == 0,
	})
}

// CreateFirstAdmin handles POST /setup/admin
func (h *AuthHandlers) CreateFirstAdmin(w http.ResponseWriter, r *http.Request) {
	// Check if setup is needed
	count, err := h.store.UserCount()
	if err != nil || count > 0 {
		WriteError(w, http.StatusForbidden, "Setup already completed", "CONFLICT")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Username and password are required", "VALIDATION_ERROR")
		return
	}

	// Hash password
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to hash password", "INTERNAL_ERROR")
		return
	}

	// Create user
	user, err := h.store.CreateUser(req.Username, req.Email, hash, "admin")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create user", "INTERNAL_ERROR")
		return
	}

	// Generate token
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Role, 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// JobHandlers handles job endpoints
type JobHandlers struct {
	store     *store.Store
	scheduler *scheduler.Scheduler
	validator *JobValidator
}

// NewJobHandlers creates job handlers
func NewJobHandlers(st *store.Store, sched *scheduler.Scheduler) *JobHandlers {
	return &JobHandlers{
		store:     st,
		scheduler: sched,
		validator: NewJobValidator(),
	}
}

// ListJobs handles GET /api/jobs
func (h *JobHandlers) ListJobs(w http.ResponseWriter, r *http.Request) {
	var createdBy *int
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if role := r.Header.Get("X-User-Role"); role != internal.RoleAdmin {
			// Non-admin users only see their own jobs
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
				return
			}
			createdBy = &userID
		}
	}

	jobs, err := h.store.ListJobs(createdBy)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to list jobs", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  jobs,
		"total": len(jobs),
	})
}

// CreateJob handles POST /api/jobs
func (h *JobHandlers) CreateJob(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	role := r.Header.Get("X-User-Role")

	if role != internal.RoleAdmin {
		WriteError(w, http.StatusForbidden, "Only admins can create jobs", "UNAUTHORIZED")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	if validErr := h.validator.ValidateJobRequest(&req); validErr != nil {
		WriteError(w, http.StatusBadRequest, validErr.Message, validErr.Code)
		return
	}

	// Validate schedule if provided
	if req.Schedule != nil {
		if validErr := h.validator.ValidateScheduleRequest(req.Schedule); validErr != nil {
			WriteError(w, http.StatusBadRequest, validErr.Message, validErr.Code)
			return
		}
	}

	h.validator.ApplyDefaults(&req)

	newJob := h.validator.ToJobModel(&req, nil)
	newJob.Enabled = true
	newJob.CreatedBy = userID

	createdJob, err := h.store.CreateJob(newJob)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create job", "INTERNAL_ERROR")
		return
	}

	// Save schedule if provided
	if req.Schedule != nil {
		schedule := &store.Schedule{
			JobID:    createdJob.ID,
			Years:    req.Schedule.Years,
			Months:   req.Schedule.Months,
			Days:     req.Schedule.Days,
			Weekdays: req.Schedule.Weekdays,
			Hours:    req.Schedule.Hours,
			Minutes:  req.Schedule.Minutes,
		}
		if err := h.store.SetJobSchedule(createdJob.ID, schedule); err != nil {
			// Job was created but schedule failed - delete the job to maintain consistency
			if delErr := h.store.DeleteJob(createdJob.ID); delErr != nil {
				log.Printf("Failed to rollback job %s after schedule error: %v\n", createdJob.ID, delErr)
			}
			WriteError(w, http.StatusInternalServerError, "Failed to set schedule", "INTERNAL_ERROR")
			return
		}
	}

	WriteJSON(w, http.StatusCreated, createdJob)
}

// GetJob handles GET /api/jobs/{id}
func (h *JobHandlers) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	// Validate job ID is not empty
	if jobID == "" {
		WriteError(w, http.StatusBadRequest, "Job ID is required", "INVALID_ID")
		return
	}

	job, err := h.store.GetJob(jobID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Job not found", "NOT_FOUND")
		return
	}

	WriteJSON(w, http.StatusOK, job)
}

// UpdateJob handles PUT /api/jobs/{id}
func (h *JobHandlers) UpdateJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	role := r.Header.Get("X-User-Role")

	if role != internal.RoleAdmin {
		WriteError(w, http.StatusForbidden, "Only admins can update jobs", "UNAUTHORIZED")
		return
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	if validErr := h.validator.ValidateJobRequest(&req); validErr != nil {
		WriteError(w, http.StatusBadRequest, validErr.Message, validErr.Code)
		return
	}

	// Validate schedule if provided
	if req.Schedule != nil {
		if validErr := h.validator.ValidateScheduleRequest(req.Schedule); validErr != nil {
			WriteError(w, http.StatusBadRequest, validErr.Message, validErr.Code)
			return
		}
	}

	job := h.validator.ToJobModel(&req, &jobID)
	job.Enabled = req.Enabled

	if err := h.store.UpdateJob(job); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update job", "INTERNAL_ERROR")
		return
	}

	// Update schedule if provided
	if req.Schedule != nil {
		schedule := &store.Schedule{
			JobID:    jobID,
			Years:    req.Schedule.Years,
			Months:   req.Schedule.Months,
			Days:     req.Schedule.Days,
			Weekdays: req.Schedule.Weekdays,
			Hours:    req.Schedule.Hours,
			Minutes:  req.Schedule.Minutes,
		}
		if err := h.store.SetJobSchedule(jobID, schedule); err != nil {
			WriteError(w, http.StatusInternalServerError, "Job updated but failed to set schedule", "INTERNAL_ERROR")
			return
		}
	}

	updatedJob, _ := h.store.GetJob(jobID)
	WriteJSON(w, http.StatusOK, updatedJob)
}

// DeleteJob handles DELETE /api/jobs/{id}
func (h *JobHandlers) DeleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	role := r.Header.Get("X-User-Role")

	if role != internal.RoleAdmin {
		WriteError(w, http.StatusForbidden, "Only admins can delete jobs", "UNAUTHORIZED")
		return
	}

	if err := h.store.DeleteJob(jobID); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete job", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Job deleted successfully",
	})
}

// TriggerJob handles POST /api/jobs/{id}/run
func (h *JobHandlers) TriggerJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	// Verify job exists
	job, err := h.store.GetJob(jobID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Job not found", "NOT_FOUND")
		return
	}

	if !job.Enabled {
		WriteError(w, http.StatusBadRequest, "Job is not enabled", "INVALID_STATE")
		return
	}

	// Create a run with manual trigger type
	run, err := h.store.CreateRun(jobID, "manual")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create run", "INTERNAL_ERROR")
		return
	}

	// Enqueue the job with the run to maintain sequential execution
	h.scheduler.EnqueueWithRun(job, run)

	WriteJSON(w, http.StatusCreated, run)
}

// RunHandlers handles run endpoints
type RunHandlers struct {
	store *store.Store
}

// NewRunHandlers creates run handlers
func NewRunHandlers(st *store.Store) *RunHandlers {
	return &RunHandlers{store: st}
}

// ListRuns handles GET /api/runs
func (h *RunHandlers) ListRuns(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	var jobIDPtr *string
	if jobID != "" {
		jobIDPtr = &jobID
	}

	runs, err := h.store.ListRuns(jobIDPtr, limit, offset)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to list runs", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"runs":  runs,
		"total": len(runs),
	})
}

// GetRun handles GET /api/runs/{id}
func (h *RunHandlers) GetRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")

	// Validate run ID is not empty
	if runID == "" {
		WriteError(w, http.StatusBadRequest, "Run ID is required", "INVALID_ID")
		return
	}

	run, err := h.store.GetRun(runID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "Run not found", "NOT_FOUND")
		return
	}

	WriteJSON(w, http.StatusOK, run)
}

// GetRunLogs handles GET /api/runs/{id}/logs
func (h *RunHandlers) GetRunLogs(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")

	logs, err := h.store.GetLogs(runID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get logs", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
	})
}

// ScheduleHandlers handles schedule endpoints
type ScheduleHandlers struct {
	store *store.Store
}

// NewScheduleHandlers creates schedule handlers
func NewScheduleHandlers(st *store.Store) *ScheduleHandlers {
	return &ScheduleHandlers{store: st}
}

// GetJobSchedule handles GET /api/jobs/{id}/schedule
func (h *ScheduleHandlers) GetJobSchedule(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	schedule, err := h.store.GetJobSchedule(jobID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get schedule", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, schedule)
}

// SetJobSchedule handles PUT /api/jobs/{id}/schedule
func (h *ScheduleHandlers) SetJobSchedule(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	role := r.Header.Get("X-User-Role")

	if role != internal.RoleAdmin {
		WriteError(w, http.StatusForbidden, "Only admins can set schedules", "UNAUTHORIZED")
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	validator := NewJobValidator()
	if validErr := validator.ValidateScheduleRequest(&req); validErr != nil {
		WriteError(w, http.StatusBadRequest, validErr.Message, validErr.Code)
		return
	}

	schedule := &store.Schedule{
		JobID:    jobID,
		Years:    req.Years,
		Months:   req.Months,
		Days:     req.Days,
		Weekdays: req.Weekdays,
		Hours:    req.Hours,
		Minutes:  req.Minutes,
	}

	if err := h.store.SetJobSchedule(jobID, schedule); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to set schedule", "INTERNAL_ERROR")
		return
	}

	updatedSchedule, _ := h.store.GetJobSchedule(jobID)
	WriteJSON(w, http.StatusOK, updatedSchedule)
}

// DashboardHandlers handles dashboard endpoints
type DashboardHandlers struct {
	store *store.Store
}

// NewDashboardHandlers creates dashboard handlers
func NewDashboardHandlers(st *store.Store) *DashboardHandlers {
	return &DashboardHandlers{store: st}
}

// GetStats handles GET /api/dashboard/stats
func (h *DashboardHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.store.ListJobs(nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get stats", "INTERNAL_ERROR")
		return
	}

	runs, err := h.store.ListRuns(nil, 100, 0)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get stats", "INTERNAL_ERROR")
		return
	}

	activeJobs := 0
	for _, job := range jobs {
		if job.Enabled {
			activeJobs++
		}
	}

	successCount := 0
	failureCount := 0
	runningCount := 0

	for _, run := range runs {
		switch run.Status {
		case "success":
			successCount++
		case "failure", "timeout":
			failureCount++
		case "running":
			runningCount++
		}
	}

	totalCompleted := successCount + failureCount
	successRate := 0.0
	if totalCompleted > 0 {
		successRate = float64(successCount) / float64(totalCompleted)
	}

	recentRuns := runs
	if len(recentRuns) > 10 {
		recentRuns = recentRuns[:10]
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"total_jobs":   len(jobs),
		"active_jobs":  activeJobs,
		"success_rate": successRate,
		"running_now":  runningCount,
		"recent_runs":  recentRuns,
	})
}

// Health handles GET /health
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}

// AnalyticsHandlers handles analytics endpoints
type AnalyticsHandlers struct {
	store *store.Store
}

// NewAnalyticsHandlers creates analytics handlers
func NewAnalyticsHandlers(st *store.Store) *AnalyticsHandlers {
	return &AnalyticsHandlers{store: st}
}

// GetExecutionTrends handles GET /api/analytics/execution-trends
func (h *AnalyticsHandlers) GetExecutionTrends(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30 // default
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
		days = d
	}

	trends, err := h.store.GetExecutionTrends(days)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get execution trends", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"trends": trends,
		"days":   days,
	})
}

// GetJobStats handles GET /api/analytics/job-stats
func (h *AnalyticsHandlers) GetJobStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetJobStats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get job stats", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":  stats,
		"total": len(stats),
	})
}

// GetJobDurationTrends handles GET /api/analytics/jobs/{id}/duration-trends
func (h *AnalyticsHandlers) GetJobDurationTrends(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		WriteError(w, http.StatusBadRequest, "Job ID is required", "INVALID_ID")
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 30 // default
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
		days = d
	}

	trends, err := h.store.GetJobDurationTrends(jobID, days)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get duration trends", "INTERNAL_ERROR")
		return
	}

	// Get job name
	job, _ := h.store.GetJob(jobID)
	jobName := ""
	if job != nil {
		jobName = job.Name
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"job_id":   jobID,
		"job_name": jobName,
		"trends":   trends,
		"days":     days,
	})
}

// GetOverallStats handles GET /api/analytics/overview
func (h *AnalyticsHandlers) GetOverallStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetOverallStats()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to get overall stats", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusOK, stats)
}
