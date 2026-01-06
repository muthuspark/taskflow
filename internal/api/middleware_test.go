package api

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/store"
)

// TestAuthMiddleware tests authentication middleware
func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key-at-least-32-bytes-long"
	jwtMgr := auth.NewJWTManager(secret)
	testStore := store.NewTestStore(t)
	defer testStore.Close()

	// Create a test user in the database
	_, err := testStore.CreateUser("admin", "admin@example.com", "test-password", "admin")
	require.NoError(t, err)

	// Generate valid token
	token, err := jwtMgr.GenerateToken(1, "admin", "admin", 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectStatus   int
		expectUserID   string
		expectUserRole string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectStatus:   http.StatusOK,
			expectUserID:   "1",
			expectUserRole: "admin",
		},
		{
			name:         "missing auth header",
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid auth format",
			authHeader:   "InvalidFormat " + token,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "malformed token",
			authHeader:   "Bearer malformed.token.here",
			expectStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := AuthMiddleware(jwtMgr, testStore)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Header.Get("X-User-ID")
				userRole := r.Header.Get("X-User-Role")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(userID + ":" + userRole))
			})

			// Wrap with middleware
			wrappedHandler := middleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Record response
			recorder := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(recorder, req)

			// Verify status
			assert.Equal(t, tt.expectStatus, recorder.Code)

			// Verify headers if successful
			if tt.expectStatus == http.StatusOK {
				assert.Equal(t, tt.expectUserID, req.Header.Get("X-User-ID"))
				assert.Equal(t, tt.expectUserRole, req.Header.Get("X-User-Role"))
			}
		})
	}
}

// TestUserIDConversion tests that user ID is properly converted to string
func TestUserIDConversion(t *testing.T) {
	tests := []struct {
		name     string
		userID   int
		expected string
	}{
		{
			name:     "small number",
			userID:   1,
			expected: "1",
		},
		{
			name:     "medium number",
			userID:   42,
			expected: "42",
		},
		{
			name:     "large number",
			userID:   999999,
			expected: "999999",
		},
		{
			name:     "zero",
			userID:   0,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strconv.Itoa(tt.userID)
			assert.Equal(t, tt.expected, result)
			// Verify it's NOT a Unicode character
			assert.NotEqual(t, string(rune(tt.userID)), result)
		})
	}
}

// TestCORSMiddleware tests CORS header handling
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		allowedOrigins    string
		method            string
		expectAllowOrigin string
	}{
		{
			name:              "wildcard origin",
			allowedOrigins:    "*",
			method:            "GET",
			expectAllowOrigin: "*",
		},
		{
			name:              "specific origin",
			allowedOrigins:    "https://example.com",
			method:            "GET",
			expectAllowOrigin: "https://example.com",
		},
		{
			name:              "options request",
			allowedOrigins:    "*",
			method:            "OPTIONS",
			expectAllowOrigin: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := CORSMiddleware(tt.allowedOrigins)

			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := middleware(testHandler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			recorder := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectAllowOrigin, recorder.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", recorder.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "Content-Type, Authorization", recorder.Header().Get("Access-Control-Allow-Headers"))

			// OPTIONS requests should return 200 immediately
			if tt.method == "OPTIONS" {
				assert.Equal(t, http.StatusOK, recorder.Code)
			}
		})
	}
}
