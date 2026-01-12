package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateUser creates a new user
func (s *Store) CreateUser(username, email, passwordHash, role string) (*User, error) {
	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}

	result, err := s.db.Exec(
		`INSERT INTO users (username, email, password_hash, role, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		username, email, passwordHash, role, user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user id: %w", err)
	}

	user.ID = int(id)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *Store) GetUser(id int) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, last_login
		 FROM users WHERE id = ?`,
		id,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *Store) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		`SELECT id, username, email, password_hash, role, created_at, last_login
		 FROM users WHERE username = ?`,
		username,
	).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}
	return user, nil
}

// ListUsers retrieves all users
func (s *Store) ListUsers() ([]*User, error) {
	rows, err := s.db.Query(
		`SELECT id, username, email, password_hash, role, created_at, last_login
		 FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*User, 0)
	for rows.Next() {
		user := &User{}
		var lastLogin sql.NullTime

		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Role, &user.CreatedAt, &lastLogin,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// UpdateUser updates user information
func (s *Store) UpdateUser(id int, email, role string) error {
	result, err := s.db.Exec(
		`UPDATE users SET email = ?, role = ? WHERE id = ?`,
		email, role, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserLastLogin updates the last_login timestamp
func (s *Store) UpdateUserLastLogin(id int) error {
	_, err := s.db.Exec(
		`UPDATE users SET last_login = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

// DeleteUser deletes a user
func (s *Store) DeleteUser(id int) error {
	result, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UserCount returns the total number of users
func (s *Store) UserCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// UpdateUserEmail updates the email for a user
func (s *Store) UpdateUserEmail(id int, email string) error {
	result, err := s.db.Exec(
		`UPDATE users SET email = ? WHERE id = ?`,
		email, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserPassword updates the password hash for a user
func (s *Store) UpdateUserPassword(id int, passwordHash string) error {
	result, err := s.db.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`,
		passwordHash, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}
