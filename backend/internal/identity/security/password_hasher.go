package security

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

const defaultQueueWait = 5 * time.Second

var ErrHasherBusy = apierror.New(http.StatusServiceUnavailable, apierror.CodeServiceUnavailable,
	"The server is busy. Please try again in a moment.")

type PasswordHasherOptions struct {
	Parameters        Argon2idParameters
	MaximumConcurrent int
	MaximumQueueWait  time.Duration
}

type PasswordHasher struct {
	parameters       Argon2idParameters
	concurrencySlots chan struct{}
	maximumQueueWait time.Duration
}

func NewPasswordHasher(options PasswordHasherOptions) *PasswordHasher {
	if options.MaximumConcurrent < 1 {
		options.MaximumConcurrent = 2 * runtime.NumCPU()
	}
	if options.MaximumQueueWait == 0 {
		options.MaximumQueueWait = defaultQueueWait
	}
	return &PasswordHasher{
		parameters:       options.Parameters,
		concurrencySlots: make(chan struct{}, options.MaximumConcurrent),
		maximumQueueWait: options.MaximumQueueWait,
	}
}

func (hasher *PasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	if err := hasher.acquireSlot(ctx); err != nil {
		return "", err
	}
	defer hasher.releaseSlot()
	return HashPassword(password, hasher.parameters)
}

func (hasher *PasswordHasher) Verify(ctx context.Context, password, encodedHash string) (bool, error) {
	if err := hasher.acquireSlot(ctx); err != nil {
		return false, err
	}
	defer hasher.releaseSlot()
	return VerifyPassword(password, encodedHash)
}

func (hasher *PasswordHasher) acquireSlot(ctx context.Context) error {
	waitContext, cancelWait := context.WithTimeout(ctx, hasher.maximumQueueWait)
	defer cancelWait()
	select {
	case hasher.concurrencySlots <- struct{}{}:
		return nil
	case <-waitContext.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		return ErrHasherBusy
	}
}

func (hasher *PasswordHasher) releaseSlot() {
	<-hasher.concurrencySlots
}
