package scheduler_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/logger"
	"nebula-exchange/backend/internal/platform/scheduler"
)

type countingJob struct {
	name        string
	runDuration time.Duration
	runCount    atomic.Int32
	activeRuns  atomic.Int32
	peakActive  atomic.Int32
	failure     error
	shouldPanic bool
}

func (job *countingJob) Name() string            { return job.name }
func (job *countingJob) Interval() time.Duration { return 10 * time.Millisecond }

func (job *countingJob) Run(ctx context.Context) error {
	currentlyActive := job.activeRuns.Add(1)
	defer job.activeRuns.Add(-1)
	for {
		peak := job.peakActive.Load()
		if currentlyActive <= peak || job.peakActive.CompareAndSwap(peak, currentlyActive) {
			break
		}
	}
	job.runCount.Add(1)
	time.Sleep(job.runDuration)
	if job.shouldPanic {
		panic("job exploded")
	}
	return job.failure
}

func newTestScheduler(t *testing.T) *scheduler.Scheduler {
	t.Helper()
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	return scheduler.New(databasetest.NewPool(t), testLogger, scheduler.Options{})
}

func TestSchedulerRunsJobsRepeatedlyUntilStopped(t *testing.T) {
	jobScheduler := newTestScheduler(t)
	job := &countingJob{name: "repeat"}
	jobScheduler.Register(job)

	schedulerContext, stopScheduler := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer stopScheduler()
	jobScheduler.Run(schedulerContext)

	if job.runCount.Load() < 3 {
		t.Fatalf("job ran %d times, want several runs", job.runCount.Load())
	}
}

func TestAdvisoryLockKeepsAJobFromOverlappingItself(t *testing.T) {
	jobScheduler := newTestScheduler(t)
	job := &countingJob{name: "exclusive", runDuration: 50 * time.Millisecond}

	var waitGroup sync.WaitGroup
	executedRuns := atomic.Int32{}
	for attemptIndex := 0; attemptIndex < 5; attemptIndex++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if jobScheduler.RunOnce(context.Background(), job) {
				executedRuns.Add(1)
			}
		}()
	}
	waitGroup.Wait()

	if job.peakActive.Load() != 1 {
		t.Fatalf("peak concurrent runs was %d, want 1", job.peakActive.Load())
	}
	if executedRuns.Load() == 5 {
		t.Fatal("concurrent attempts should have been skipped while the lock was held")
	}
}

func TestFailingAndPanickingJobsDoNotStopTheScheduler(t *testing.T) {
	jobScheduler := newTestScheduler(t)
	failingJob := &countingJob{name: "failing", failure: errors.New("boom")}
	panickingJob := &countingJob{name: "panicking", shouldPanic: true}

	if !jobScheduler.RunOnce(context.Background(), failingJob) || !jobScheduler.RunOnce(context.Background(), panickingJob) {
		t.Fatal("failing jobs still count as executed runs")
	}
	if !jobScheduler.RunOnce(context.Background(), panickingJob) {
		t.Fatal("the lock must be released after a panic")
	}
}
