package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/taskflow/taskflow/internal/auth"
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
	store *store.Store
}

// NewJobHandlers creates job handlers
func NewJobHandlers(st *store.Store) *JobHandlers {
	return &JobHandlers{store: st}
}

// ListJobs handles GET /api/jobs
func (h *JobHandlers) ListJobs(w http.ResponseWriter, r *http.Request) {
	var createdBy *int
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if role := r.Header.Get("X-User-Role"); role != "admin" {
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

	if role != "admin" {
		WriteError(w, http.StatusForbidden, "Only admins can create jobs", "UNAUTHORIZED")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	var req struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		Script            string `json:"script"`
		WorkingDir        string `json:"working_dir"`
		TimeoutSeconds    int    `json:"timeout_seconds"`
		RetryCount        int    `json:"retry_count"`
		RetryDelaySeconds int    `json:"retry_delay_seconds"`
		NotifyEmails      string `json:"notify_emails"`
		NotifyOn          string `json:"notify_on"`
		Timezone          string `json:"timezone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body", "VALIDATION_ERROR")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Script == "" {
		WriteError(w, http.StatusBadRequest, "Name and script are required", "VALIDATION_ERROR")
		return
	}

	// Validate name length
	if len(req.Name) > 255 {
		WriteError(w, http.StatusBadRequest, "Job name too long (max 255 characters)", "VALIDATION_ERROR")
		return
	}

	// Validate script length (1MB max, set during execution)
	if len(req.Script) > 1000000 {
		WriteError(w, http.StatusBadRequest, "Script too long (max 1MB)", "VALIDATION_ERROR")
		return
	}

	// Validate timeout
	if req.TimeoutSeconds < 1 || req.TimeoutSeconds > 86400 {
		WriteError(w, http.StatusBadRequest, "Timeout must be between 1 and 86400 seconds", "VALIDATION_ERROR")
		return
	}

	// Validate retry values
	if req.RetryCount < 0 || req.RetryCount > 10 {
		WriteError(w, http.StatusBadRequest, "Retry count must be between 0 and 10", "VALIDATION_ERROR")
		return
	}

	// Validate retry delay
	if req.RetryDelaySeconds < 0 || req.RetryDelaySeconds > 86400 {
		WriteError(w, http.StatusBadRequest, "Retry delay must be between 0 and 86400 seconds", "VALIDATION_ERROR")
		return
	}

	// Validate notify_on
	if req.NotifyOn != "" && req.NotifyOn != "always" && req.NotifyOn != "failure" && req.NotifyOn != "success" {
		WriteError(w, http.StatusBadRequest, "Invalid notify_on value", "VALIDATION_ERROR")
		return
	}

	// Set defaults
	if req.WorkingDir == "" {
		req.WorkingDir = "/tmp"
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 3600
	}
	if req.NotifyOn == "" {
		req.NotifyOn = "failure"
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	job := &store.Job{
		Name:              req.Name,
		Description:       req.Description,
		Script:            req.Script,
		WorkingDir:        req.WorkingDir,
		TimeoutSeconds:    req.TimeoutSeconds,
		RetryCount:        req.RetryCount,
		RetryDelaySeconds: req.RetryDelaySeconds,
		Enabled:           true,
		NotifyEmails:      req.NotifyEmails,
		NotifyOn:          req.NotifyOn,
		Timezone:          req.Timezone,
		CreatedBy:         userID,
	}

	job, err := h.store.CreateJob(job)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create job", "INTERNAL_ERROR")
		return
	}

	WriteJSON(w, http.StatusCreated, job)
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

// Health handles GET /health
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok"}`)
}
