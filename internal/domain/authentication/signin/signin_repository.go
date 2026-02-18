package signin

import (
	"context"
	"fmt"
	"time"

	"dvith.com/go-service-api/pkg/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type SigninRepository struct {
	db *database.DBPool
}

// NewSignupRepository creates a new signup repository
func NewSigninRepository(db *database.DBPool) *SigninRepository {
	return &SigninRepository{
		db: db,
	}
}

func (repo *SigninRepository) FindUser(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be nil")
	}

	query := `
		SELECT id, email, password, full_name, username, is_active, email_verified, verified_at, created_at, updated_at, deleted_at
		FROM users
		WHERE is_active = true AND email = $1
	`

	row := repo.db.QueryRow(ctx, query, email)

	// Scan the returned row
	var user User
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
