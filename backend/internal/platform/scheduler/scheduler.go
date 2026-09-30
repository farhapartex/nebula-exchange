package scheduler

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Job interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

type Options struct {
	InitialDelay time.Duration
}

type Scheduler struct {
	pool    *pgxpool.Pool
	logger  *slog.Logger
	options Options
	jobs    []Job
}

func New(pool *pgxpool.Pool, logger *slog.Logger, options Options) *Scheduler {
	return &Scheduler{pool: pool, logger: logger, options: options}
}

func (scheduler *Scheduler) Register(job Job) {
	scheduler.jobs = append(scheduler.jobs, job)
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	var runningJobs sync.WaitGroup
	for _, job := range scheduler.jobs {
		runningJobs.Add(1)
		go func(scheduledJob Job) {
			defer runningJobs.Done()
			scheduler.loop(ctx, scheduledJob)
		}(job)
	}
	runningJobs.Wait()
}

func (scheduler *Scheduler) loop(ctx context.Context, job Job) {
	initialTimer := time.NewTimer(scheduler.options.InitialDelay)
	defer initialTimer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-initialTimer.C:
	}

	intervalTicker := time.NewTicker(job.Interval())
	defer intervalTicker.Stop()
	for {
		scheduler.RunOnce(ctx, job)
		select {
		case <-ctx.Done():
			return
		case <-intervalTicker.C:
		}
	}
}

func (scheduler *Scheduler) RunOnce(ctx context.Context, job Job) bool {
	jobLogger := scheduler.logger.With(slog.String("job", job.Name()))
	lockConnection, err := scheduler.pool.Acquire(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			jobLogger.Error("acquire connection for job lock", slog.String("error", err.Error()))
		}
		return false
	}
	defer lockConnection.Release()

	lockKey := advisoryLockKey(job.Name())
	var isLockAcquired bool
	if err := lockConnection.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&isLockAcquired); err != nil {
		jobLogger.Error("take job lock", slog.String("error", err.Error()))
		return false
	}
	if !isLockAcquired {
		jobLogger.Debug("job already running elsewhere, skipping")
		return false
	}
	defer func() {
		if _, err := lockConnection.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
			jobLogger.Error("release job lock", slog.String("error", err.Error()))
		}
	}()

	startedAt := time.Now()
	if err := scheduler.runSafely(ctx, job); err != nil {
		jobLogger.Error("job failed", slog.String("error", err.Error()), slog.Duration("duration", time.Since(startedAt)))
		return true
	}
	jobLogger.Debug("job finished", slog.Duration("duration", time.Since(startedAt)))
	return true
}

func (scheduler *Scheduler) runSafely(ctx context.Context, job Job) (err error) {
	defer func() {
		if recoveredValue := recover(); recoveredValue != nil {
			err = fmt.Errorf("job panicked: %v", recoveredValue)
		}
	}()
	return job.Run(ctx)
}

func advisoryLockKey(jobName string) int64 {
	nameHasher := fnv.New64a()
	_, _ = nameHasher.Write([]byte("nebula-job:" + jobName))
	return int64(nameHasher.Sum64())
}
