package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type SubscriptionChapterResponse struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
}

type SubscriptionResponse struct {
	ID              uuid.UUID                     `json:"id"`
	PlanID          string                        `json:"plan_id"`
	PlanName        string                        `json:"plan_name"`
	PlanKind        models.PlanKind               `json:"plan_kind"`
	Status          models.PaymentStatus          `json:"status"`
	Chapters        []SubscriptionChapterResponse `json:"chapters"`
	SubtotalCents   string                        `json:"subtotal_cents"`
	DiscountPercent int                           `json:"discount_percent"`
	DiscountCents   string                        `json:"discount_cents"`
	TotalCents      string                        `json:"total_cents"`
	PaidAt          time.Time                     `json:"paid_at"`
	RefundedAt      *time.Time                    `json:"refunded_at"`
}

type SubscriptionHandler struct {
	subscriptionService service.SubscriptionService
}

func NewSubscriptionHandler(subscriptionService service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subscriptionService: subscriptionService}
}

func (handler *SubscriptionHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/subscriptions", authentication.RequireUser(), handler.listSubscriptions)
}

func (handler *SubscriptionHandler) listSubscriptions(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after *service.SubscriptionCursor
	if pageRequest.HasCursor() {
		cursor, err := request.DecodeCursorPosition[service.SubscriptionCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		after = &cursor
	}
	userID, _ := authentication.UserIDFrom(context)
	paymentPage, err := handler.subscriptionService.List(context.Request.Context(), userID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	subscriptionResponses := make([]SubscriptionResponse, 0, len(paymentPage.Items))
	for _, payment := range paymentPage.Items {
		subscriptionResponses = append(subscriptionResponses, toSubscriptionResponse(payment))
	}
	response.WriteList(context, http.StatusOK, pagination.Page[SubscriptionResponse]{Items: subscriptionResponses, Info: paymentPage.Info})
}

func toSubscriptionResponse(payment models.Payment) SubscriptionResponse {
	chapterResponses := make([]SubscriptionChapterResponse, 0, len(payment.Chapters))
	for _, chapter := range payment.Chapters {
		chapterResponses = append(chapterResponses, SubscriptionChapterResponse{ID: chapter.ChapterID, Number: chapter.ChapterNumber, Title: chapter.ChapterTitle})
	}
	return SubscriptionResponse{
		ID:              payment.ID,
		PlanID:          payment.PlanID,
		PlanName:        payment.Plan.Name,
		PlanKind:        payment.Plan.Kind,
		Status:          payment.Status,
		Chapters:        chapterResponses,
		SubtotalCents:   formatCents(payment.SubtotalCents),
		DiscountPercent: payment.DiscountPercent,
		DiscountCents:   formatCents(payment.DiscountCents),
		TotalCents:      formatCents(payment.TotalCents),
		PaidAt:          *payment.PaidAt,
		RefundedAt:      payment.RefundedAt,
	}
}
