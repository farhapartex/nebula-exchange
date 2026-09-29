package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (repository *Repository) FindCredentialsByEmail(ctx context.Context, database usersstore.DBTX, email string) (Credentials, bool, error) {
	credentialRow, err := usersstore.New(database).FindUserCredentialsByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return Credentials{}, false, nil
	}
	if err != nil {
		return Credentials{}, false, err
	}
	return Credentials{
		User: User{
			ID:            credentialRow.ID,
			Email:         credentialRow.Email,
			Username:      credentialRow.Username,
			Status:        Status(credentialRow.Status),
			IsActive:      credentialRow.IsActive,
			IsAdmin:       credentialRow.IsAdmin,
			CreatedAt:     credentialRow.CreatedAt,
			LastLoginAt:   credentialRow.LastLoginAt,
			TotpEnabledAt: credentialRow.TotpEnabledAt,
		},
		PasswordHash: credentialRow.PasswordHash,
	}, true, nil
}

func (repository *Repository) FindByID(ctx context.Context, database usersstore.DBTX, userID uuid.UUID) (User, bool, error) {
	userRow, err := usersstore.New(database).FindUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return User{
		ID:            userRow.ID,
		Email:         userRow.Email,
		Username:      userRow.Username,
		Status:        Status(userRow.Status),
		IsActive:      userRow.IsActive,
		IsAdmin:       userRow.IsAdmin,
		CreatedAt:     userRow.CreatedAt,
		ActivatedAt:   userRow.ActivatedAt,
		LastLoginAt:   userRow.LastLoginAt,
		TotpEnabledAt: userRow.TotpEnabledAt,
	}, true, nil
}

func (repository *Repository) RecordLogin(ctx context.Context, database usersstore.DBTX, userID uuid.UUID, loggedInAt time.Time) error {
	return usersstore.New(database).RecordLogin(ctx, usersstore.RecordLoginParams{ID: userID, LastLoginAt: &loggedInAt})
}

func (repository *Repository) Activate(ctx context.Context, database usersstore.DBTX, userID uuid.UUID, activatedAt time.Time) (User, bool, error) {
	activatedRow, err := usersstore.New(database).ActivateUser(ctx, usersstore.ActivateUserParams{ID: userID, ActivatedAt: activatedAt})
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, err
	}
	return User{
		ID:          activatedRow.ID,
		Email:       activatedRow.Email,
		Username:    activatedRow.Username,
		Status:      Status(activatedRow.Status),
		IsActive:    activatedRow.IsActive,
		IsAdmin:     activatedRow.IsAdmin,
		CreatedAt:   activatedRow.CreatedAt,
		ActivatedAt: activatedRow.ActivatedAt,
		LastLoginAt: activatedRow.LastLoginAt,
	}, true, nil
}

func (repository *Repository) UpdatePasswordHash(ctx context.Context, database usersstore.DBTX, userID uuid.UUID, passwordHash string, updatedAt time.Time) (string, error) {
	return usersstore.New(database).UpdatePasswordHash(ctx, usersstore.UpdatePasswordHashParams{
		ID:           userID,
		PasswordHash: passwordHash,
		UpdatedAt:    updatedAt,
	})
}

func (repository *Repository) UpdateUsername(ctx context.Context, database usersstore.DBTX, userID uuid.UUID, username string, updatedAt time.Time) error {
	_, err := usersstore.New(database).UpdateUsername(ctx, usersstore.UpdateUsernameParams{ID: userID, Username: username, UpdatedAt: updatedAt})
	return translateUniqueViolation(err)
}

func (repository *Repository) FindPasswordHash(ctx context.Context, database usersstore.DBTX, userID uuid.UUID) (string, error) {
	return usersstore.New(database).FindPasswordHashByID(ctx, userID)
}
