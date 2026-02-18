package signup

import (
	"context"
	"fmt"
	"time"

	"dvith.com/go-service-api/pkg/database"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	Email         string     `db:"email" json:"email"`
	Password      string     `db:"password" json:"-"`
	FullName      string     `db:"full_name" json:"full_name"`
	Username      string     `db:"username" json:"username"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	EmailVerified bool       `db:"email_verified" json:"email_verified"`
	VerifiedAt    *time.Time `db:"verified_at" json:"verified_at"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at" json:"deleted_at"`
}

// SignupRepository handles user signup operations
type SignupRepository struct {
	db *database.DBPool
}

// NewSignupRepository creates a new signup repository
func NewSignupRepository(db *database.DBPool) *SignupRepository {
	return &SignupRepository{
		db: db,
	}
}

// SaveUser saves a new user to the database
func (repo *SignupRepository) SaveUser(ctx context.Context, user *User) (*User, error) {
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Generate new UUID if not provided
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Set default values
	if !user.IsActive {
		user.IsActive = true // default is active
	}

	query := `
		INSERT INTO users (id, email, password, full_name, username, is_active, email_verified, verified_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, email, password, full_name, username, is_active, email_verified, verified_at, created_at, updated_at, deleted_at
	`

	row := repo.db.QueryRow(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Password,
		user.FullName,
		user.Username,
		user.IsActive,
		user.EmailVerified,
		user.VerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	// Scan the returned row
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FullName,
		&user.Username,
		&user.IsActive,
		&user.EmailVerified,
		&user.VerifiedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}
