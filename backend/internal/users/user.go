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
