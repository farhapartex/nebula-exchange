package missions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/missions/missionsstore"
)

const resolverBatchSize = 200

type ResolverJob struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
	random RandomSource
	now    func() time.Time
}

func NewResolverJob(pool *pgxpool.Pool, logger *slog.Logger, random RandomSource, now func() time.Time) *ResolverJob {
	return &ResolverJob{pool: pool, logger: logger, random: random, now: now}
}

func (job *ResolverJob) Name() string            { return "mission_resolver" }
func (job *ResolverJob) Interval() time.Duration { return 5 * time.Second }

func (job *ResolverJob) Run(ctx context.Context) error {
	queries := missionsstore.New(job.pool)
	dueMissions, err := queries.ListDueMissions(ctx, missionsstore.ListDueMissionsParams{Now: job.now(), RowLimit: resolverBatchSize})
	if err != nil {
		return fmt.Errorf("list due missions: %w", err)
	}
	resolvedCount := 0
	for _, dueMission := range dueMissions {
		var snapshot ZoneSnapshot
		if err := json.Unmarshal(dueMission.ZoneSnapshot, &snapshot); err != nil {
			job.logger.Error("mission snapshot unreadable", slog.String("mission_id", dueMission.ID.String()), slog.String("error", err.Error()))
			continue
		}
		encodedLoot, err := json.Marshal(RollLoot(snapshot.lootRules(), job.random))
		if err != nil {
			return err
		}
		resolvedAt := job.now()
		if _, err := queries.CompleteMission(ctx, missionsstore.CompleteMissionParams{ID: dueMission.ID, Loot: encodedLoot, ResolvedAt: &resolvedAt}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return fmt.Errorf("complete mission: %w", err)
		}
		resolvedCount++
	}
	if resolvedCount > 0 {
		job.logger.Info("resolved missions", slog.Int("count", resolvedCount))
	}
	return nil
}
