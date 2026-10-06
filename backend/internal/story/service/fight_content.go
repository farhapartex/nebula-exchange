package service

import (
	"context"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type FightArena struct {
	Name   string
	Width  int
	FloorY int
	Stage  database.JSONDocument
}

type FightEnemy struct {
	ID    string
	Name  string
	Title string
	Stats database.JSONDocument
	Brain database.JSONDocument
	Look  database.JSONDocument
}

type EnemyWave struct {
	Wave      int
	Enemy     FightEnemy
	Modifiers database.JSONDocument
	IntroLine *string
}

type FightContent struct {
	LevelID          string
	TimeLimitSeconds int
	Arena            FightArena
	Waves            []EnemyWave
}

func (catalog *levelCatalog) FightContent(ctx context.Context, levelID string) (FightContent, error) {
	if _, err := catalog.PlayableLevel(ctx, levelID); err != nil {
		return FightContent{}, err
	}
	level, isFound, err := catalog.levels.FindWithFightContent(ctx, levelID)
	if err != nil {
		return FightContent{}, err
	}
	if !isFound || len(level.Enemies) == 0 {
		return FightContent{}, ErrLevelNotPlayable
	}
	waves := make([]EnemyWave, 0, len(level.Enemies))
	for _, levelEnemy := range level.Enemies {
		enemy := levelEnemy.Enemy
		waves = append(waves, EnemyWave{
			Wave:      levelEnemy.Wave,
			Enemy:     FightEnemy{ID: enemy.ID, Name: enemy.Name, Title: enemy.Title, Stats: enemy.Stats, Brain: enemy.Brain, Look: enemy.Look},
			Modifiers: levelEnemy.Modifiers,
			IntroLine: levelEnemy.IntroLine,
		})
	}
	return FightContent{
		LevelID:          level.ID,
		TimeLimitSeconds: level.TimeLimitSeconds,
		Arena:            FightArena{Name: level.Arena.Name, Width: level.Arena.Width, FloorY: level.Arena.FloorY, Stage: level.Arena.Stage},
		Waves:            waves,
	}, nil
}
