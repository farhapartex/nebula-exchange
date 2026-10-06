package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

const (
	checkoutLifetime          = 31 * time.Minute
	checkoutCreationGrace     = time.Minute
	checkoutCurrency          = "usd"
	checkoutReturnParameter   = "checkout"
	checkoutCancelledValue    = "cancelled"
	checkoutSuccessPath       = "/subscription"
	checkoutCancelPath        = "/fight"
	checkoutProductNamePrefix = "Street Born"
)

var (
	ErrPaymentsUnavailable       = apierror.ServiceUnavailable("Payments are not available right now. Try again later.")
	ErrCheckoutAlreadyStarting   = apierror.Conflict("A checkout is already being started. Try again in a moment.")
	ErrPreviousCheckoutCompleted = apierror.Conflict("Your previous payment just went through. Reload to see your chapters.")
	ErrCheckoutNotFound          = apierror.NotFound("Checkout not found")
)

type CheckoutStatus string

const (
	CheckoutStatusOpen    CheckoutStatus = "OPEN"
	CheckoutStatusPaid    CheckoutStatus = "PAID"
	CheckoutStatusExpired CheckoutStatus = "EXPIRED"
)

type StartedCheckout struct {
	ID          uuid.UUID
	CheckoutURL string
	ExpiresAt   time.Time
}

type CheckoutState struct {
	ID     uuid.UUID
	Status CheckoutStatus
}

type CheckoutService interface {
	Start(ctx context.Context, userID uuid.UUID, planID string, chapterCount int) (StartedCheckout, error)
	State(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID) (CheckoutState, error)
}

type CheckoutDependencies struct {
	Plans           repository.PlanRepository
	Payments        repository.PaymentRepository
	Chapters        ChapterCatalog
	Ownership       ChapterOwnership
	Unlocker        ChapterUnlocker
	Gateway         CheckoutGateway
	Transactions    TransactionRunner
	FrontendBaseURL string
	Logger          *slog.Logger
	Now             Clock
}

type checkoutService struct {
	dependencies CheckoutDependencies
	quoter       chapterQuoter
	settlement   paymentSettlement
}

func NewCheckoutService(dependencies CheckoutDependencies) CheckoutService {
	return &checkoutService{
		dependencies: dependencies,
		quoter:       chapterQuoter{chapters: dependencies.Chapters, ownership: dependencies.Ownership},
		settlement: paymentSettlement{
			payments:     dependencies.Payments,
			unlocker:     dependencies.Unlocker,
			transactions: dependencies.Transactions,
			logger:       dependencies.Logger,
			now:          dependencies.Now,
		},
	}
}

func (checkout *checkoutService) Start(ctx context.Context, userID uuid.UUID, planID string, chapterCount int) (StartedCheckout, error) {
	if checkout.dependencies.Gateway == nil {
		return StartedCheckout{}, ErrPaymentsUnavailable
	}
	plan, option, err := checkout.quote(ctx, userID, planID, chapterCount)
	if err != nil {
		return StartedCheckout{}, err
	}
	if err := checkout.expireOpenCheckouts(ctx, userID); err != nil {
		return StartedCheckout{}, err
	}

	payment, err := checkout.newPayment(userID, plan, option)
	if err != nil {
		return StartedCheckout{}, err
	}
	if err := checkout.dependencies.Payments.CreateWithChapters(ctx, &payment); err != nil {
		if errors.Is(err, repository.ErrOpenCheckoutExists) {
			return StartedCheckout{}, ErrCheckoutAlreadyStarting
		}
		return StartedCheckout{}, err
	}

	createdCheckout, err := checkout.dependencies.Gateway.CreateCheckout(ctx, CheckoutRequest{
		PaymentID:          payment.ID,
		ProductName:        checkoutProductName(option),
		ProductDescription: checkoutProductDescription(plan, option),
		AmountCents:        payment.TotalCents,
		Currency:           payment.Currency,
		SuccessURL:         checkout.returnURL(checkoutSuccessPath, payment.ID.String()),
		CancelURL:          checkout.returnURL(checkoutCancelPath, checkoutCancelledValue),
		ExpiresAt:          payment.CheckoutExpiresAt,
	})
	if err != nil {
		checkout.dependencies.Logger.ErrorContext(ctx, "stripe checkout could not be created", slog.String("payment_id", payment.ID.String()), slog.Any("error", err))
		if updateErr := checkout.dependencies.Payments.Update(ctx, &payment, map[string]any{"status": models.PaymentStatusFailed}); updateErr != nil {
			return StartedCheckout{}, updateErr
		}
		return StartedCheckout{}, ErrPaymentsUnavailable
	}
	if err := checkout.dependencies.Payments.Update(ctx, &payment, map[string]any{
		"stripe_checkout_session_id": createdCheckout.SessionID,
		"checkout_url":               createdCheckout.URL,
	}); err != nil {
		return StartedCheckout{}, err
	}
	return StartedCheckout{ID: payment.ID, CheckoutURL: createdCheckout.URL, ExpiresAt: payment.CheckoutExpiresAt}, nil
}

func (checkout *checkoutService) quote(ctx context.Context, userID uuid.UUID, planID string, chapterCount int) (models.Plan, PlanOption, error) {
	plan, err := checkout.dependencies.Plans.FindActive(ctx, planID)
	if errors.Is(err, repository.ErrPlanNotFound) {
		return models.Plan{}, PlanOption{}, apierror.ValidationFailed(map[string]string{"plan_id": "is not available"})
	}
	if err != nil {
		return models.Plan{}, PlanOption{}, err
	}
	discountTiers, err := ParseDiscountTiers(plan.DiscountTiers)
	if err != nil {
		return models.Plan{}, PlanOption{}, err
	}
	chaptersToBuy, err := checkout.quoter.chaptersToBuy(ctx, userID)
	if err != nil {
		return models.Plan{}, PlanOption{}, err
	}
	for _, option := range optionsForPlan(plan.Kind, chaptersToBuy, discountTiers) {
		if option.ChapterCount == chapterCount {
			return plan, option, nil
		}
	}
	return models.Plan{}, PlanOption{}, apierror.ValidationFailed(map[string]string{"chapter_count": "is not offered by this plan"})
}

func (checkout *checkoutService) expireOpenCheckouts(ctx context.Context, userID uuid.UUID) error {
	openPayments, err := checkout.dependencies.Payments.ListOpenForUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, openPayment := range openPayments {
		if err := checkout.expireOpenCheckout(ctx, openPayment); err != nil {
			return err
		}
	}
	return nil
}

func (checkout *checkoutService) expireOpenCheckout(ctx context.Context, openPayment models.Payment) error {
	if openPayment.StripeCheckoutSessionID == nil {
		if checkout.dependencies.Now().Sub(openPayment.CreatedAt) < checkoutCreationGrace {
			return ErrCheckoutAlreadyStarting
		}
		return checkout.dependencies.Payments.Update(ctx, &openPayment, map[string]any{"status": models.PaymentStatusFailed})
	}
	sessionID := *openPayment.StripeCheckoutSessionID
	expireErr := checkout.dependencies.Gateway.ExpireCheckout(ctx, sessionID)
	if expireErr == nil {
		return checkout.settlement.markExpired(ctx, openPayment.ID)
	}
	snapshot, err := checkout.dependencies.Gateway.FetchCheckout(ctx, sessionID)
	if err != nil {
		checkout.dependencies.Logger.ErrorContext(ctx, "previous stripe checkout could not be expired", slog.String("payment_id", openPayment.ID.String()), slog.Any("error", expireErr))
		return ErrPaymentsUnavailable
	}
	if snapshot.IsPaid {
		if err := checkout.settlement.markPaid(ctx, openPayment.ID, snapshot); err != nil {
			return err
		}
		return ErrPreviousCheckoutCompleted
	}
	if snapshot.IsExpired {
		return checkout.settlement.markExpired(ctx, openPayment.ID)
	}
	checkout.dependencies.Logger.ErrorContext(ctx, "previous stripe checkout is still open", slog.String("payment_id", openPayment.ID.String()), slog.Any("error", expireErr))
	return ErrPaymentsUnavailable
}

func (checkout *checkoutService) newPayment(userID uuid.UUID, plan models.Plan, option PlanOption) (models.Payment, error) {
	paymentID, err := uuid.NewV7()
	if err != nil {
		return models.Payment{}, err
	}
	createdAt := checkout.dependencies.Now().UTC()
	paymentChapters := make([]models.PaymentChapter, 0, len(option.Chapters))
	for _, chapter := range option.Chapters {
		paymentChapters = append(paymentChapters, models.PaymentChapter{
			PaymentID:     paymentID,
			ChapterID:     chapter.ID,
			ChapterNumber: chapter.Number,
			ChapterTitle:  chapter.Title,
			PriceCents:    chapter.PriceCents,
		})
	}
	return models.Payment{
		ID:                paymentID,
		UserID:            userID,
		PlanID:            plan.ID,
		Status:            models.PaymentStatusOpen,
		Currency:          checkoutCurrency,
		SubtotalCents:     option.SubtotalCents,
		DiscountPercent:   option.DiscountPercent,
		DiscountCents:     option.DiscountCents,
		TotalCents:        option.TotalCents,
		CheckoutExpiresAt: createdAt.Add(checkoutLifetime),
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		Chapters:          paymentChapters,
	}, nil
}

func (checkout *checkoutService) returnURL(path string, checkoutValue string) string {
	query := url.Values{checkoutReturnParameter: []string{checkoutValue}}
	return strings.TrimRight(checkout.dependencies.FrontendBaseURL, "/") + path + "?" + query.Encode()
}

func (checkout *checkoutService) State(ctx context.Context, userID uuid.UUID, paymentID uuid.UUID) (CheckoutState, error) {
	payment, err := checkout.dependencies.Payments.FindForUser(ctx, paymentID, userID)
	if errors.Is(err, repository.ErrPaymentNotFound) {
		return CheckoutState{}, ErrCheckoutNotFound
	}
	if err != nil {
		return CheckoutState{}, err
	}
	if payment.Status == models.PaymentStatusOpen && payment.StripeCheckoutSessionID != nil && checkout.dependencies.Gateway != nil {
		if payment, err = checkout.refreshFromStripe(ctx, userID, payment); err != nil {
			return CheckoutState{}, err
		}
	}
	return CheckoutState{ID: payment.ID, Status: checkoutStatusOf(payment.Status)}, nil
}

func (checkout *checkoutService) refreshFromStripe(ctx context.Context, userID uuid.UUID, payment models.Payment) (models.Payment, error) {
	snapshot, err := checkout.dependencies.Gateway.FetchCheckout(ctx, *payment.StripeCheckoutSessionID)
	if err != nil {
		checkout.dependencies.Logger.WarnContext(ctx, "stripe checkout could not be fetched", slog.String("payment_id", payment.ID.String()), slog.Any("error", err))
		return payment, nil
	}
	switch {
	case snapshot.IsPaid:
		err = checkout.settlement.markPaid(ctx, payment.ID, snapshot)
	case snapshot.IsExpired:
		err = checkout.settlement.markExpired(ctx, payment.ID)
	default:
		return payment, nil
	}
	if err != nil {
		return models.Payment{}, err
	}
	return checkout.dependencies.Payments.FindForUser(ctx, payment.ID, userID)
}

func checkoutStatusOf(paymentStatus models.PaymentStatus) CheckoutStatus {
	if paymentStatus.WasPaid() {
		return CheckoutStatusPaid
	}
	if paymentStatus == models.PaymentStatusOpen {
		return CheckoutStatusOpen
	}
	return CheckoutStatusExpired
}

func checkoutProductName(option PlanOption) string {
	if option.ChapterCount == 1 {
		return fmt.Sprintf("%s: Chapter %d", checkoutProductNamePrefix, option.Chapters[0].Number)
	}
	return fmt.Sprintf("%s: Chapters %d to %d", checkoutProductNamePrefix, option.Chapters[0].Number, option.Chapters[len(option.Chapters)-1].Number)
}

func checkoutProductDescription(plan models.Plan, option PlanOption) string {
	chapterTitles := make([]string, 0, len(option.Chapters))
	for _, chapter := range option.Chapters {
		chapterTitles = append(chapterTitles, fmt.Sprintf("Chapter %d: %s", chapter.Number, chapter.Title))
	}
	description := plan.Name + ". " + strings.Join(chapterTitles, ", ")
	if option.DiscountPercent > 0 {
		description += fmt.Sprintf(". Includes %d%% off.", option.DiscountPercent)
	}
	return description
}
