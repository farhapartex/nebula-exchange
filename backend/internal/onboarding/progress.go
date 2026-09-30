package onboarding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/ledger/ledgerstore"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Step struct {
	Key         string `json:"key"`
	IsCompleted bool   `json:"is_completed"`
}

type Progress struct {
	Steps          []Step `json:"steps"`
	CompletedCount int    `json:"completed_count"`
}

var checklist = []struct {
	key         string
	journalType ledger.JournalType
}{
	{"paid_entry_fee", ledger.JournalEntryFee},
	{"first_mission", ledger.JournalMissionLoot},
	{"first_craft", ledger.JournalCraftOutput},
	{"first_trade", ledger.JournalTradeFill},
}

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me/onboarding", authentication.RequireUser(), handler.getProgress)
}

func (handler *Handler) getProgress(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	progress, err := handler.load(context.Request.Context(), userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, progress)
}

func (handler *Handler) load(ctx context.Context, userID uuid.UUID) (Progress, error) {
	journalTypes := make([]string, 0, len(checklist))
	for _, checklistStep := range checklist {
		journalTypes = append(journalTypes, string(checklistStep.journalType))
	}
	seenTypes, err := ledgerstore.New(handler.pool).ListPlayerJournalTypesAmong(ctx, ledgerstore.ListPlayerJournalTypesAmongParams{
		UserID:       &userID,
		JournalTypes: journalTypes,
	})
	if err != nil {
		return Progress{}, fmt.Errorf("load onboarding progress: %w", err)
	}
	seen := map[string]bool{}
	for _, seenType := range seenTypes {
		seen[seenType] = true
	}
	progress := Progress{Steps: make([]Step, 0, len(checklist))}
	for _, checklistStep := range checklist {
		isCompleted := seen[string(checklistStep.journalType)]
		progress.Steps = append(progress.Steps, Step{Key: checklistStep.key, IsCompleted: isCompleted})
		if isCompleted {
			progress.CompletedCount++
		}
	}
	return progress, nil
}
