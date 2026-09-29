package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"nebula-exchange/backend/internal/users/usersstore"
)

const (
	uniqueViolationCode      = "23505"
	emailUniqueConstraint    = "users_email_unique"
	usernameUniqueConstraint = "users_username_unique"
)

var (
	ErrEmailTaken    = errors.New("email is already registered")
	ErrUsernameTaken = errors.New("username is already taken")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (repository *Repository) FindTakenIdentifiers(ctx context.Context, database usersstore.DBTX, email, username string) (TakenIdentifiers, error) {
	checkResult, err := usersstore.New(database).CheckIdentifiersTaken(ctx, usersstore.CheckIdentifiersTakenParams{
		Email:    email,
		Username: username,
	})
	if err != nil {
		return TakenIdentifiers{}, err
	}
	return TakenIdentifiers{IsEmailTaken: checkResult.IsEmailTaken, IsUsernameTaken: checkResult.IsUsernameTaken}, nil
}

func (repository *Repository) Create(ctx context.Context, database usersstore.DBTX, newUser NewUser) (User, error) {
	userID, err := uuid.NewV7()
	if err != nil {
		return User{}, err
	}

	createdRow, err := usersstore.New(database).CreateUser(ctx, usersstore.CreateUserParams{
		ID:              userID,
		Email:           newUser.Email,
		Username:        newUser.Username,
		PasswordHash:    newUser.PasswordHash,
		TermsAcceptedAt: newUser.TermsAcceptedAt,
	})
	if err != nil {
		return User{}, translateUniqueViolation(err)
	}

	return User{
		ID:              createdRow.ID,
		Email:           createdRow.Email,
		Username:        createdRow.Username,
		Status:          Status(createdRow.Status),
		IsActive:        createdRow.IsActive,
		IsAdmin:         createdRow.IsAdmin,
		TermsAcceptedAt: createdRow.TermsAcceptedAt,
		ActivatedAt:     createdRow.ActivatedAt,
		LastLoginAt:     createdRow.LastLoginAt,
		CreatedAt:       createdRow.CreatedAt,
		UpdatedAt:       createdRow.UpdatedAt,
	}, nil
}

func translateUniqueViolation(err error) error {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != uniqueViolationCode {
		return err
	}
	switch postgresError.ConstraintName {
	case emailUniqueConstraint:
		return ErrEmailTaken
	case usernameUniqueConstraint:
		return ErrUsernameTaken
	default:
		return err
	}
}
