package handler

import (
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
)

type UserProfileResponse struct {
	ID          uuid.UUID         `json:"id"`
	Email       string            `json:"email"`
	Username    string            `json:"username"`
	Status      models.UserStatus `json:"status"`
	IsActive    bool              `json:"is_active"`
	IsAdmin     bool              `json:"is_admin"`
	CreatedAt   time.Time         `json:"created_at"`
	LastLoginAt *time.Time        `json:"last_login_at"`
}

func toUserProfileResponse(user models.User) UserProfileResponse {
	return UserProfileResponse{
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
