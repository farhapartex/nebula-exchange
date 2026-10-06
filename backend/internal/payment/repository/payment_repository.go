package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

const oneOpenCheckoutPerUserIndex = "payments_one_open_checkout_per_user"

var (
	ErrPaymentNotFound    = errors.New("payment not found")
	ErrOpenCheckoutExists = errors.New("the player already has an open checkout")
)

type SettledPaymentPosition struct {
	PaidAt time.Time
	ID     uuid.UUID
}

type PaymentRepository interface {
	CreateWithChapters(ctx context.Context, payment *models.Payment) error
	ListOpenForUser(ctx context.Context, userID uuid.UUID) ([]models.Payment, error)
	FindForUser(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) (models.Payment, error)
	LockByID(ctx context.Context, paymentID uuid.UUID) (models.Payment, error)
	LockByPaymentIntent(ctx context.Context, paymentIntentID string) (models.Payment, error)
	ListPaidForUser(ctx context.Context, userID uuid.UUID, excludedPaymentID uuid.UUID) ([]models.Payment, error)
	ListSettledForUser(ctx context.Context, userID uuid.UUID, after *SettledPaymentPosition, limit int) ([]models.Payment, error)
	Update(ctx context.Context, payment *models.Payment, changes map[string]any) error
}

type GormPaymentRepository struct {
	database *gorm.DB
}

func NewPaymentRepository(database *gorm.DB) *GormPaymentRepository {
	return &GormPaymentRepository{database: database}
}

func (repository *GormPaymentRepository) CreateWithChapters(ctx context.Context, payment *models.Payment) error {
	err := database.Session(ctx, repository.database).Omit("Plan").Create(payment).Error
	if constraintName, isUniqueViolation := database.ViolatedUniqueConstraint(err); isUniqueViolation && constraintName == oneOpenCheckoutPerUserIndex {
		return ErrOpenCheckoutExists
	}
	return err
}

func (repository *GormPaymentRepository) ListOpenForUser(ctx context.Context, userID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"user_id": userID, "status": models.PaymentStatusOpen}).
		Find(&payments).Error
	return payments, err
}

func (repository *GormPaymentRepository) FindForUser(ctx context.Context, paymentID uuid.UUID, userID uuid.UUID) (models.Payment, error) {
	var payment models.Payment
	err := database.Session(ctx, repository.database).
		Preload("Chapters").
		Where(map[string]any{"id": paymentID, "user_id": userID}).
		Take(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Payment{}, ErrPaymentNotFound
	}
	return payment, err
}

func (repository *GormPaymentRepository) LockByID(ctx context.Context, paymentID uuid.UUID) (models.Payment, error) {
	return repository.lockOne(ctx, map[string]any{"id": paymentID})
}

func (repository *GormPaymentRepository) LockByPaymentIntent(ctx context.Context, paymentIntentID string) (models.Payment, error) {
	return repository.lockOne(ctx, map[string]any{"stripe_payment_intent_id": paymentIntentID})
}

func (repository *GormPaymentRepository) lockOne(ctx context.Context, conditions map[string]any) (models.Payment, error) {
	var payment models.Payment
	err := database.Session(ctx, repository.database).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where(conditions).
		Take(&payment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Payment{}, ErrPaymentNotFound
	}
	if err != nil {
		return models.Payment{}, err
	}
	err = database.Session(ctx, repository.database).
		Where(map[string]any{"payment_id": payment.ID}).
		Order("chapter_number").
		Find(&payment.Chapters).Error
	return payment, err
}

func (repository *GormPaymentRepository) ListPaidForUser(ctx context.Context, userID uuid.UUID, excludedPaymentID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := database.Session(ctx, repository.database).
		Preload("Chapters").
		Where(map[string]any{"user_id": userID, "status": models.PaymentStatusPaid}).
		Not(map[string]any{"id": excludedPaymentID}).
		Find(&payments).Error
	return payments, err
}

func (repository *GormPaymentRepository) ListSettledForUser(ctx context.Context, userID uuid.UUID, after *SettledPaymentPosition, limit int) ([]models.Payment, error) {
	query := database.Session(ctx, repository.database).
		Preload("Plan").
		Preload("Chapters", func(chapters *gorm.DB) *gorm.DB { return chapters.Order("chapter_number") }).
		Where(map[string]any{"user_id": userID, "status": models.SettledPaymentStatuses})
	if after != nil {
		query = query.Where("(paid_at, id) < (?, ?)", after.PaidAt, after.ID)
	}
	var payments []models.Payment
	err := query.Order("paid_at DESC").Order("id DESC").Limit(limit).Find(&payments).Error
	return payments, err
}

func (repository *GormPaymentRepository) Update(ctx context.Context, payment *models.Payment, changes map[string]any) error {
	return database.Session(ctx, repository.database).
		Model(payment).
		Omit("Plan", "Chapters").
		Updates(changes).Error
}
