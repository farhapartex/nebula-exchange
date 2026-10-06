package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type SubscriptionCursor struct {
	PaidAt time.Time `json:"paid_at"`
	ID     uuid.UUID `json:"id"`
}

type SubscriptionService interface {
	List(ctx context.Context, userID uuid.UUID, after *SubscriptionCursor, pageRequest pagination.Request) (pagination.Page[models.Payment], error)
}

type subscriptionService struct {
	payments repository.PaymentRepository
}

func NewSubscriptionService(payments repository.PaymentRepository) SubscriptionService {
	return &subscriptionService{payments: payments}
}

func (subscriptions *subscriptionService) List(ctx context.Context, userID uuid.UUID, after *SubscriptionCursor, pageRequest pagination.Request) (pagination.Page[models.Payment], error) {
	var afterPosition *repository.SettledPaymentPosition
	if after != nil {
		afterPosition = &repository.SettledPaymentPosition{PaidAt: after.PaidAt, ID: after.ID}
	}
	payments, err := subscriptions.payments.ListSettledForUser(ctx, userID, afterPosition, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[models.Payment]{}, err
	}
	return pagination.BuildPage(payments, pageRequest, func(payment models.Payment) SubscriptionCursor {
		return SubscriptionCursor{PaidAt: *payment.PaidAt, ID: payment.ID}
	})
}
