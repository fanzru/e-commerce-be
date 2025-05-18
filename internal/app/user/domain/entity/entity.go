package entity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRole defines the role of a user
type UserRole string

const (
	// RoleAdmin is for administrative users
	RoleAdmin UserRole = "admin"
	// RoleCustomer is for regular customers
	RoleCustomer UserRole = "customer"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // Never expose password in JSON responses
	Name      string     `json:"name"`
	Role      UserRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// RefreshToken represents a refresh token for authentication
type RefreshToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUser creates a new user with the given details
func NewUser(email, password, name string, role UserRole) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ComparePassword checks if the provided password matches the user's password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// NewRefreshToken creates a new refresh token for the given user
func NewRefreshToken(userID uuid.UUID, tokenStr string, expiresInDays int) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: now.Add(time.Duration(expiresInDays) * 24 * time.Hour),
		CreatedAt: now,
	}
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return rt.ExpiresAt.Before(time.Now())
}
