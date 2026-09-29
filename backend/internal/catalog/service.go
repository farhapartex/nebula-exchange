package catalog

import (
	"context"
	"sync"
	"time"
)

type snapshotSource interface {
	Load(ctx context.Context) (Snapshot, error)
}

type Service struct {
	source      snapshotSource
	cacheTTL    time.Duration
	now         func() time.Time
	cacheMutex  sync.Mutex
	cached      *Snapshot
	cachedUntil time.Time
}

func NewService(source snapshotSource, cacheTTL time.Duration, now func() time.Time) *Service {
	return &Service{source: source, cacheTTL: cacheTTL, now: now}
}

func (service *Service) Snapshot(ctx context.Context) (Snapshot, error) {
	service.cacheMutex.Lock()
	defer service.cacheMutex.Unlock()

	currentTime := service.now()
	if service.cached != nil && currentTime.Before(service.cachedUntil) {
		return *service.cached, nil
	}
	freshSnapshot, err := service.source.Load(ctx)
	if err != nil {
		if service.cached != nil {
			return *service.cached, nil
		}
		return Snapshot{}, err
	}
	service.cached = &freshSnapshot
	service.cachedUntil = currentTime.Add(service.cacheTTL)
	return freshSnapshot, nil
}
