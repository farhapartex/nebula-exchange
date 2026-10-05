package models

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusUnverified UserStatus = "UNVERIFIED"
	UserStatusActive     UserStatus = "ACTIVE"
	UserStatusFrozen     UserStatus = "FROZEN"
	UserStatusBanned     UserStatus = "BANNED"
)

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email           string     `gorm:"type:citext;not null"`
	Username        string     `gorm:"type:citext;not null"`
	PasswordHash    string     `gorm:"not null"`
	Status          UserStatus `gorm:"type:user_status;not null"`
	IsActive        bool       `gorm:"not null"`
	IsAdmin         bool       `gorm:"not null"`
	TermsAcceptedAt time.Time  `gorm:"not null"`
	ActivatedAt     *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (User) TableName() string {
	return "users"
}
