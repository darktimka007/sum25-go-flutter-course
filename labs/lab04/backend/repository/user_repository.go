package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"lab04-backend/models"
)

// UserRepository handles database operations for users
// This repository demonstrates MANUAL SQL approach with database/sql package
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}

	// Convert request to user model
	user := req.ToUser()

	// Insert into users table with RETURNING clause to get generated ID and timestamps
	query := `
		INSERT INTO users (name, email, created_at, updated_at) 
		VALUES (?, ?, ?, ?) 
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRow(query, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)

	// Scan the returned values
	err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return user, nil
}

// GetByID gets user by ID from database
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		WHERE id = ? AND deleted_at IS NULL`

	row := r.db.QueryRow(query, id)

	var user models.User
	err := user.ScanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// GetByEmail gets user by email from database
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		WHERE email = ? AND deleted_at IS NULL`

	row := r.db.QueryRow(query, email)

	var user models.User
	err := user.ScanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

// GetAll gets all users from database
func (r *UserRepository) GetAll() ([]models.User, error) {
	query := `
		SELECT id, name, email, created_at, updated_at 
		FROM users 
		WHERE deleted_at IS NULL 
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %v", err)
	}

	users, err := models.ScanUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan users: %v", err)
	}

	return users, nil
}

// Update updates user in database
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// First check if user exists
	existingUser, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Build dynamic UPDATE query based on non-nil fields
	var setParts []string
	var args []interface{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}

	if req.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *req.Email)
	}

	if len(setParts) == 0 {
		return existingUser, nil // Nothing to update
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())

	// Add WHERE clause parameter
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = ? AND deleted_at IS NULL",
		strings.Join(setParts, ", "),
	)

	_, err = r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	// Return updated user
	return r.GetByID(id)
}

// Delete deletes user from database
func (r *UserRepository) Delete(id int) error {
	// First check if user exists
	_, err := r.GetByID(id)
	if err != nil {
		return err
	}

	// Perform soft delete by setting deleted_at timestamp
	query := `
		UPDATE users 
		SET deleted_at = ? 
		WHERE id = ? AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found or already deleted", id)
	}

	return nil
}

// Count counts total number of users
func (r *UserRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %v", err)
	}

	return count, nil
}
