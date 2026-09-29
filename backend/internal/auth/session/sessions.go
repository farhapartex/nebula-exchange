package session

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/session/sessionstore"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/pagination"
)

var errSessionNotFound = apierror.NotFound("This session no longer exists")

type ActiveSession struct {
	ID           uuid.UUID `json:"id"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	StartedAt    time.Time `json:"started_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsCurrent    bool      `json:"is_current"`
}

type sessionCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type RevocationResult struct {
	WasCurrentSession bool
}

type Sessions struct {
	pool *pgxpool.Pool
	now  func() time.Time
}

func NewSessions(pool *pgxpool.Pool, now func() time.Time) *Sessions {
	return &Sessions{pool: pool, now: now}
}

func (sessions *Sessions) List(ctx context.Context, userID uuid.UUID, currentTokenHash []byte, pageRequest pagination.Request, cursorPosition *sessionCursor) (pagination.Page[ActiveSession], error) {
	listParameters := sessionstore.ListActiveSessionsParams{
		UserID:   userID,
		Now:      sessions.now().UTC(),
		RowLimit: int32(pageRequest.FetchLimit()),
	}
	if cursorPosition != nil {
		listParameters.CursorCreatedAt = &cursorPosition.CreatedAt
		listParameters.CursorID = &cursorPosition.ID
	}
	sessionRows, err := sessionstore.New(sessions.pool).ListActiveSessions(ctx, listParameters)
	if err != nil {
		return pagination.Page[ActiveSession]{}, err
	}

	activeSessions := make([]ActiveSession, 0, len(sessionRows))
	for _, sessionRow := range sessionRows {
		activeSessions = append(activeSessions, ActiveSession{
			ID:           sessionRow.ID,
			UserAgent:    sessionRow.UserAgent,
			IPAddress:    sessionRow.IpAddress,
			StartedAt:    sessionRow.SessionStartedAt,
			LastActiveAt: sessionRow.CreatedAt,
		})
	}
	currentSessionID, hasCurrent := sessions.currentSessionID(ctx, currentTokenHash)
	for sessionIndex := range activeSessions {
		activeSessions[sessionIndex].IsCurrent = hasCurrent && activeSessions[sessionIndex].ID == currentSessionID
	}
	return pagination.BuildPage(activeSessions, pageRequest, func(activeSession ActiveSession) sessionCursor {
		return sessionCursor{CreatedAt: activeSession.LastActiveAt, ID: activeSession.ID}
	})
}

func (sessions *Sessions) Revoke(ctx context.Context, userID, sessionID uuid.UUID, currentTokenHash []byte) (RevocationResult, error) {
	revokedTokenHash, err := sessionstore.New(sessions.pool).RevokeSessionForUser(ctx, sessionstore.RevokeSessionForUserParams{
		ID:     sessionID,
		UserID: userID,
		Now:    sessions.now().UTC(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return RevocationResult{}, errSessionNotFound
	}
	if err != nil {
		return RevocationResult{}, err
	}
	return RevocationResult{WasCurrentSession: len(currentTokenHash) > 0 && bytes.Equal(revokedTokenHash, currentTokenHash)}, nil
}

func (sessions *Sessions) RevokeOthers(ctx context.Context, database sessionstore.DBTX, userID uuid.UUID, keptTokenHash []byte) (int64, error) {
	return sessionstore.New(database).RevokeOtherSessionsForUser(ctx, sessionstore.RevokeOtherSessionsForUserParams{
		UserID:        userID,
		KeptTokenHash: keptTokenHash,
		Now:           sessions.now().UTC(),
	})
}

func (sessions *Sessions) currentSessionID(ctx context.Context, currentTokenHash []byte) (uuid.UUID, bool) {
	if len(currentTokenHash) == 0 {
		return uuid.Nil, false
	}
	storedToken, err := sessionstore.New(sessions.pool).FindRefreshTokenByHash(ctx, currentTokenHash)
	if err != nil || storedToken.RevokedAt != nil {
		return uuid.Nil, false
	}
	return storedToken.ID, true
}
