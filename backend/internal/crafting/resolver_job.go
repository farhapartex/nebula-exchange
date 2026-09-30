package crafting

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"nebula-exchange/backend/internal/crafting/craftingstore"
)

const resolverBatchSize = 200

type ResolverJob struct {
	service *Service
	logger  *slog.Logger
}

func NewResolverJob(service *Service, logger *slog.Logger) *ResolverJob {
	return &ResolverJob{service: service, logger: logger}
}

func (job *ResolverJob) Name() string            { return "craft_resolver" }
func (job *ResolverJob) Interval() time.Duration { return 5 * time.Second }

func (job *ResolverJob) Run(ctx context.Context) error {
	dueJobs, err := craftingstore.New(job.service.pool).ListDueCraftJobs(ctx, craftingstore.ListDueCraftJobsParams{
		Now: job.service.now(), RowLimit: resolverBatchSize,
	})
	if err != nil {
		return fmt.Errorf("list due craft jobs: %w", err)
	}
	deliveredCount := 0
	for _, dueJob := range dueJobs {
		isDelivered, err := job.service.Deliver(ctx, dueJob.ID)
		if err != nil {
			job.logger.Error("craft delivery failed", slog.String("craft_job_id", dueJob.ID.String()), slog.String("error", err.Error()))
			continue
		}
		if isDelivered {
			deliveredCount++
		}
	}
	if deliveredCount > 0 {
		job.logger.Info("delivered crafts", slog.Int("count", deliveredCount))
	}
	return nil
}
