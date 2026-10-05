package idempotency

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

const (
	KeyHeader            = "Idempotency-Key"
	ReplayedHeader       = "Idempotent-Replayed"
	defaultRetention     = 24 * time.Hour
	defaultInProgressTTL = 2 * time.Minute
	defaultMaximumBody   = 1 << 20
)

type Options struct {
	ExemptRoutes     []string
	Retention        time.Duration
	InProgressTTL    time.Duration
	MaximumBodyBytes int64
	Now              func() time.Time
}

type acquisitionOutcome int

const (
	acquiredKey acquisitionOutcome = iota
	existingKey
)

type middleware struct {
	store   Store
	logger  *slog.Logger
	options Options
}

func Middleware(store Store, logger *slog.Logger, options Options) gin.HandlerFunc {
	idempotencyMiddleware := &middleware{store: store, logger: logger, options: withDefaults(options)}
	return idempotencyMiddleware.handle
}

func withDefaults(options Options) Options {
	if options.Retention == 0 {
		options.Retention = defaultRetention
	}
	if options.InProgressTTL == 0 {
		options.InProgressTTL = defaultInProgressTTL
	}
	if options.MaximumBodyBytes == 0 {
		options.MaximumBodyBytes = defaultMaximumBody
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return options
}

func (idempotencyMiddleware *middleware) handle(context *gin.Context) {
	key := context.GetHeader(KeyHeader)
	if key == "" || !isStateChangingMethod(context.Request.Method) || slices.Contains(idempotencyMiddleware.options.ExemptRoutes, context.FullPath()) {
		context.Next()
		return
	}
	if !isAcceptedKey(key) {
		response.WriteError(context, apierror.BadRequest("Idempotency-Key must be 8 to 255 letters, digits or - _ : ."))
		return
	}

	requestBody, err := readAndRestoreBody(context.Request, idempotencyMiddleware.options.MaximumBodyBytes)
	if errors.Is(err, errRequestBodyTooLarge) {
		response.WriteError(context, apierror.New(http.StatusRequestEntityTooLarge, apierror.CodeValidationFailed, "Request body is too large"))
		return
	}
	if err != nil {
		response.WriteError(context, apierror.BadRequest("Request body could not be read"))
		return
	}

	scope := scopeFrom(context)
	requestHash := fingerprintRequest(context.Request, requestBody)
	outcome, existingRecord, err := idempotencyMiddleware.acquire(context.Request.Context(), scope, key, requestHash)
	if err != nil {
		response.WriteError(context, err)
		return
	}

	if outcome == existingKey {
		idempotencyMiddleware.respondToExistingKey(context, existingRecord, requestHash)
		return
	}

	recorder := &responseRecorder{ResponseWriter: context.Writer}
	context.Writer = recorder
	hasFinished := false
	defer func() {
		if !hasFinished {
			idempotencyMiddleware.releaseAfterPanic(context.Request.Context(), scope, key)
		}
	}()
	context.Next()
	idempotencyMiddleware.finish(context.Request.Context(), scope, key, recorder)
	hasFinished = true
}

func (idempotencyMiddleware *middleware) releaseAfterPanic(requestContext context.Context, scope, key string) {
	persistContext := context.WithoutCancel(requestContext)
	if err := idempotencyMiddleware.store.Release(persistContext, scope, key); err != nil {
		idempotencyMiddleware.logger.ErrorContext(persistContext, "release idempotency key after panic", slog.String("error", err.Error()))
	}
}

func (idempotencyMiddleware *middleware) acquire(ctx context.Context, scope, key string, requestHash []byte) (acquisitionOutcome, Record, error) {
	const maximumAttempts = 3
	for attempt := 0; attempt < maximumAttempts; attempt++ {
		hasStarted, err := idempotencyMiddleware.store.Begin(ctx, scope, key, requestHash)
		if err != nil {
			return acquiredKey, Record{}, err
		}
		if hasStarted {
			return acquiredKey, Record{}, nil
		}

		existingRecord, isFound, err := idempotencyMiddleware.store.Find(ctx, scope, key)
		if err != nil {
			return acquiredKey, Record{}, err
		}
		if !isFound {
			continue
		}

		reclaimCutoff, isReclaimable := idempotencyMiddleware.reclaimCutoff(existingRecord)
		if !isReclaimable {
			return existingKey, existingRecord, nil
		}
		if err := idempotencyMiddleware.store.DeleteIfCreatedBefore(ctx, scope, key, reclaimCutoff); err != nil {
			return acquiredKey, Record{}, err
		}
	}
	return acquiredKey, Record{}, apierror.Conflict("A request with this Idempotency-Key is in progress")
}

func (idempotencyMiddleware *middleware) reclaimCutoff(existingRecord Record) (time.Time, bool) {
	now := idempotencyMiddleware.options.Now()
	lifetime := idempotencyMiddleware.options.Retention
	if !existingRecord.IsCompleted {
		lifetime = idempotencyMiddleware.options.InProgressTTL
	}
	cutoff := now.Add(-lifetime)
	return cutoff, existingRecord.CreatedAt.Before(cutoff)
}

func (idempotencyMiddleware *middleware) respondToExistingKey(context *gin.Context, existingRecord Record, requestHash []byte) {
	if !bytes.Equal(existingRecord.RequestHash, requestHash) {
		response.WriteError(context, apierror.Conflict("Idempotency-Key was already used for a different request"))
		return
	}
	if !existingRecord.IsCompleted {
		response.WriteError(context, apierror.Conflict("A request with this Idempotency-Key is in progress"))
		return
	}

	context.Header(ReplayedHeader, "true")
	contentType := existingRecord.Response.ContentType
	if contentType == "" {
		contentType = "application/json; charset=utf-8"
	}
	context.Data(existingRecord.Response.StatusCode, contentType, existingRecord.Response.Body)
	context.Abort()
}

func (idempotencyMiddleware *middleware) finish(requestContext context.Context, scope, key string, recorder *responseRecorder) {
	persistContext := context.WithoutCancel(requestContext)
	statusCode := recorder.Status()

	var err error
	if statusCode >= http.StatusInternalServerError {
		err = idempotencyMiddleware.store.Release(persistContext, scope, key)
	} else {
		err = idempotencyMiddleware.store.Complete(persistContext, scope, key, StoredResponse{
			StatusCode:  statusCode,
			ContentType: recorder.Header().Get("Content-Type"),
			Body:        recorder.recordedBody.Bytes(),
		})
	}
	if err != nil {
		idempotencyMiddleware.logger.ErrorContext(persistContext, "store idempotent response", slog.String("error", err.Error()))
	}
}
