package inapp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/notify/inapp/inappstore"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/pagination"
)

const maximumIDsPerReadRequest = 100

type markReadBody struct {
	IDs []string `json:"ids" binding:"max=100"`
	All bool     `json:"all"`
}

type listCursor struct {
	BeforeID uuid.UUID `json:"before_id"`
}

type Handler struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewHandler(pool *pgxpool.Pool, now func() time.Time) *Handler {
	return &Handler{pool: pool, now: now}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	notificationRoutes := router.Group("/me/notifications", authentication.RequireUser())
	notificationRoutes.GET("", handler.listNotifications)
	notificationRoutes.GET("/unread-count", handler.countUnread)
	notificationRoutes.POST("/read", handler.markRead)
}

func (handler *Handler) listNotifications(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	parameters := inappstore.ListNotificationsParams{
		UserID:     userID,
		UnreadOnly: context.Query("unread_only") == "true",
		RowLimit:   int32(pageRequest.FetchLimit()),
	}
	if pageRequest.HasCursor() {
		cursor, err := request.DecodeCursorPosition[listCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		parameters.BeforeID = &cursor.BeforeID
	}
	notificationRows, err := inappstore.New(handler.pool).ListNotifications(context.Request.Context(), parameters)
	if err != nil {
		response.WriteError(context, fmt.Errorf("list notifications: %w", err))
		return
	}
	notifications := make([]Notification, 0, len(notificationRows))
	for _, notificationRow := range notificationRows {
		notifications = append(notifications, notificationFromRow(notificationRow))
	}
	notificationPage, err := pagination.BuildPage(notifications, pageRequest, func(notification Notification) listCursor {
		return listCursor{BeforeID: notification.ID}
	})
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList(context, http.StatusOK, notificationPage)
}

func (handler *Handler) countUnread(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	unreadCount, err := inappstore.New(handler.pool).CountUnreadNotifications(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, fmt.Errorf("count unread notifications: %w", err))
		return
	}
	response.WriteData(context, http.StatusOK, map[string]int32{"unread_count": unreadCount})
}

func (handler *Handler) markRead(context *gin.Context) {
	var body markReadBody
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	if !body.All && len(body.IDs) == 0 {
		response.WriteError(context, apierror.ValidationFailed(map[string]string{"ids": "list notification ids or set all to true"}))
		return
	}
	notificationIDs := make([]uuid.UUID, 0, min(len(body.IDs), maximumIDsPerReadRequest))
	for _, rawID := range body.IDs {
		notificationID, err := uuid.Parse(rawID)
		if err != nil {
			response.WriteError(context, apierror.ValidationFailed(map[string]string{"ids": "must contain notification ids"}))
			return
		}
		notificationIDs = append(notificationIDs, notificationID)
	}
	userID, _ := authentication.UserIDFrom(context)
	readAt := handler.now()
	markedCount, err := inappstore.New(handler.pool).MarkNotificationsRead(context.Request.Context(), inappstore.MarkNotificationsReadParams{
		ReadAt:          &readAt,
		UserID:          userID,
		MarkAll:         body.All,
		NotificationIds: notificationIDs,
	})
	if err != nil {
		response.WriteError(context, fmt.Errorf("mark notifications read: %w", err))
		return
	}
	response.WriteData(context, http.StatusOK, map[string]int32{"marked": markedCount})
}
