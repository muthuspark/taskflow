package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/taskflow/taskflow/internal/auth"
	"github.com/taskflow/taskflow/internal/store"
)

// ContextKey is used for context values
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// AuthMiddleware checks JWT tokens and adds user info to context
func AuthMiddleware(jwtManager *auth.JWTManager, store *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				WriteError(w, http.StatusUnauthorized, "Missing authorization header", "INVALID_TOKEN")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				WriteError(w, http.StatusUnauthorized, "Invalid authorization header", "INVALID_TOKEN")
				return
			}

			token := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "Invalid token", "INVALID_TOKEN")
				return
			}

			// Get user from database
			user, err := store.GetUser(claims.UserID)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "User not found", "INVALID_TOKEN")
				return
			}

			// Add user to context
			r.Header.Set("X-User-ID", strconv.Itoa(user.ID))
			r.Header.Set("X-User-Role", user.Role)

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
