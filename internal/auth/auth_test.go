package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHashPassword tests password hashing
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "secure_password_123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_that_should_still_hash_correctly_1234567890",
			wantErr:  false,
		},
		{
			name:     "special characters",
			password: "p@$$w0rd!#%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			// Hash should be bcrypt format
			assert.True(t, len(hash) > 20)
		})
	}
}

// TestVerifyPassword tests password verification
func TestVerifyPassword(t *testing.T) {
	password := "correct_password"
	hash, err := HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name      string
		hash      string
		password  string
		wantMatch bool
	}{
		{
			name:      "correct password",
			hash:      hash,
			password:  password,
			wantMatch: true,
		},
		{
			name:      "incorrect password",
			hash:      hash,
			password:  "wrong_password",
			wantMatch: false,
		},
		{
			name:      "empty password",
			hash:      hash,
			password:  "",
			wantMatch: false,
		},
		{
			name:      "case sensitive",
			hash:      hash,
			password:  "CORRECT_PASSWORD",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(tt.hash, tt.password)
			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

// TestJWTGeneration tests JWT token generation
func TestJWTGeneration(t *testing.T) {
	jwtMgr := NewJWTManager("test-secret-key-at-least-32-bytes-long")

	tests := []struct {
		name      string
		userID    int
		username  string
		role      string
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "valid token",
			userID:    1,
			username:  "testuser",
			role:      "admin",
			expiresIn: 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "zero expires uses default",
			userID:    2,
			username:  "user2",
			role:      "user",
			expiresIn: 0,
			wantErr:   false,
		},
		{
			name:      "short expiry",
			userID:    3,
			username:  "user3",
			role:      "user",
			expiresIn: 1 * time.Minute,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtMgr.GenerateToken(tt.userID, tt.username, tt.role, tt.expiresIn)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, token)
			// Token should have 3 parts separated by 2 dots
			dotCount := 0
			for _, char := range token {
				if char == '.' {
					dotCount++
				}
			}
			assert.Equal(t, 2, dotCount, "JWT should have 3 parts separated by 2 dots")
		})
	}
}

// TestJWTValidation tests JWT token validation
func TestJWTValidation(t *testing.T) {
	secret := "test-secret-key-at-least-32-bytes-long"
	jwtMgr := NewJWTManager(secret)

	// Create a valid token
	token, err := jwtMgr.GenerateToken(42, "testuser", "admin", 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		wantError bool
		wantID    int
		wantUser  string
		wantRole  string
	}{
		{
			name:      "valid token",
			token:     token,
			wantError: false,
			wantID:    42,
			wantUser:  "testuser",
			wantRole:  "admin",
		},
		{
			name:      "invalid token format",
			token:     "invalid.token.format",
			wantError: true,
		},
		{
			name:      "empty token",
			token:     "",
			wantError: true,
		},
		{
			name:      "tampered token",
			token:     token + "tampered",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtMgr.ValidateToken(tt.token)
			if tt.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, claims.UserID)
			assert.Equal(t, tt.wantUser, claims.Username)
			assert.Equal(t, tt.wantRole, claims.Role)
		})
	}
}

// TestJWTWithDifferentSecret tests that tokens are secret-dependent
func TestJWTWithDifferentSecret(t *testing.T) {
	secret1 := "secret-one-at-least-32-bytes-long-1"
	secret2 := "secret-two-at-least-32-bytes-long-2"

	jwtMgr1 := NewJWTManager(secret1)
	jwtMgr2 := NewJWTManager(secret2)

	token, err := jwtMgr1.GenerateToken(1, "user", "user", 24*time.Hour)
	require.NoError(t, err)

	// Should validate with same secret
	_, err = jwtMgr1.ValidateToken(token)
	require.NoError(t, err)

	// Should NOT validate with different secret
	_, err = jwtMgr2.ValidateToken(token)
	require.Error(t, err)
}
