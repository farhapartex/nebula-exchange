package passwordhash

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	"nebula-exchange/backend/internal/platform/apierror"
)

const defaultQueueWait = 5 * time.Second

var ErrHasherBusy = apierror.New(http.StatusServiceUnavailable, apierror.CodeServiceUnavailable,
	"The server is busy. Please try again in a moment.")

type HasherOptions struct {
	Parameters        Parameters
	MaximumConcurrent int
	MaximumQueueWait  time.Duration
}

type Hasher struct {
	parameters       Parameters
	concurrencySlots chan struct{}
	maximumQueueWait time.Duration
}

func NewHasher(options HasherOptions) *Hasher {
	if options.MaximumConcurrent < 1 {
		options.MaximumConcurrent = 2 * runtime.NumCPU()
	}
	if options.MaximumQueueWait == 0 {
		options.MaximumQueueWait = defaultQueueWait
	}
	return &Hasher{
		parameters:       options.Parameters,
		concurrencySlots: make(chan struct{}, options.MaximumConcurrent),
		maximumQueueWait: options.MaximumQueueWait,
	}
}

func (hasher *Hasher) Hash(ctx context.Context, password string) (string, error) {
	if err := hasher.acquireSlot(ctx); err != nil {
		return "", err
	}
	defer hasher.releaseSlot()
	return Hash(password, hasher.parameters)
}

func (hasher *Hasher) Verify(ctx context.Context, password, encodedHash string) (bool, error) {
	if err := hasher.acquireSlot(ctx); err != nil {
		return false, err
	}
	defer hasher.releaseSlot()
	return Verify(password, encodedHash)
}

func (hasher *Hasher) acquireSlot(ctx context.Context) error {
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

func (hasher *Hasher) releaseSlot() {
	<-hasher.concurrencySlots
}
