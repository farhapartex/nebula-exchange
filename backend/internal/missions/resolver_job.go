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
	"nebula-exchange/backend/internal/notify/inapp"
)

const resolverBatchSize = 200

type ResolverJob struct {
	pool     *pgxpool.Pool
	notifier *inapp.Notifier
	logger   *slog.Logger
	random   RandomSource
	now      func() time.Time
}

func NewResolverJob(pool *pgxpool.Pool, notifier *inapp.Notifier, logger *slog.Logger, random RandomSource, now func() time.Time) *ResolverJob {
	return &ResolverJob{pool: pool, notifier: notifier, logger: logger, random: random, now: now}
}

func (job *ResolverJob) Name() string            { return "mission_resolver" }
func (job *ResolverJob) Interval() time.Duration { return 5 * time.Second }

func (job *ResolverJob) Run(ctx context.Context) error {
	dueMissions, err := missionsstore.New(job.pool).ListDueMissions(ctx, missionsstore.ListDueMissionsParams{Now: job.now(), RowLimit: resolverBatchSize})
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
		isResolved, err := job.completeAndNotify(ctx, dueMission, snapshot, encodedLoot)
		if err != nil {
			return err
		}
		if isResolved {
			resolvedCount++
		}
	}
	if resolvedCount > 0 {
		job.logger.Info("resolved missions", slog.Int("count", resolvedCount))
	}
	return nil
}

func (job *ResolverJob) completeAndNotify(ctx context.Context, dueMission missionsstore.Mission, snapshot ZoneSnapshot, encodedLoot []byte) (bool, error) {
	isResolved := false
	err := pgx.BeginFunc(ctx, job.pool, func(tx pgx.Tx) error {
		resolvedAt := job.now()
		_, err := missionsstore.New(tx).CompleteMission(ctx, missionsstore.CompleteMissionParams{ID: dueMission.ID, Loot: encodedLoot, ResolvedAt: &resolvedAt})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("complete mission: %w", err)
		}
		isResolved = true
		if job.notifier == nil {
			return nil
		}
		return job.notifier.Notify(ctx, tx, inapp.Notice{
			UserID: dueMission.UserID,
			Kind:   inapp.KindMissionCompleted,
			Title:  snapshot.ZoneName + " mission is back",
			Body:   "your ship returned with its haul. Collect the loot to unlock your ship and drill.",
			Link:   "/missions",
		})
	})
	return isResolved, err
}
