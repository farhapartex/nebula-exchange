package users

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusUnverified     Status = "UNVERIFIED"
	StatusPendingPayment Status = "PENDING_PAYMENT"
	StatusActive         Status = "ACTIVE"
	StatusFrozen         Status = "FROZEN"
	StatusBanned         Status = "BANNED"
)

type User struct {
	ID              uuid.UUID
	Email           string
	Username        string
	Status          Status
	IsActive        bool
	IsAdmin         bool
	TermsAcceptedAt time.Time
	ActivatedAt     *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NewUser struct {
	Email           string
	Username        string
	PasswordHash    string
	TermsAcceptedAt time.Time
}

type TakenIdentifiers struct {
	IsEmailTaken    bool
	IsUsernameTaken bool
}

func (takenIdentifiers TakenIdentifiers) Any() bool {
	return takenIdentifiers.IsEmailTaken || takenIdentifiers.IsUsernameTaken
}

type Credentials struct {
	User         User
	PasswordHash string
}

type Profile struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	Status      Status     `json:"status"`
	IsActive    bool       `json:"is_active"`
	IsAdmin     bool       `json:"is_admin"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (user User) Profile() Profile {
	return Profile{
		ID:          user.ID,
		Email:       user.Email,
		Username:    user.Username,
		Status:      user.Status,
		IsActive:    user.IsActive,
		IsAdmin:     user.IsAdmin,
		CreatedAt:   user.CreatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
