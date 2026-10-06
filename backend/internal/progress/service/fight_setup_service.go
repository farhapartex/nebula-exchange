package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type FightSetup struct {
	LevelID          string
	TimeLimitSeconds int
	Arena            storyservice.FightArena
	Player           Fighter
	Enemy            storyservice.FightEnemy
}

type FightSetupService interface {
	FightSetup(ctx context.Context, userID uuid.UUID, levelID string) (FightSetup, error)
}

type fightSetupService struct {
	levels   LevelCatalog
	fighters fighterResolver
}

func NewFightSetupService(levels LevelCatalog, fighters repository.FighterRepository) FightSetupService {
	return &fightSetupService{levels: levels, fighters: fighterResolver{fighters: fighters}}
}

func (setups *fightSetupService) FightSetup(ctx context.Context, userID uuid.UUID, levelID string) (FightSetup, error) {
	fightContent, err := setups.levels.FightContent(ctx, levelID)
	if err != nil {
		return FightSetup{}, err
	}
	player, err := setups.fighters.currentFighter(ctx, userID)
	if err != nil {
		return FightSetup{}, err
	}
	return FightSetup{
		LevelID:          fightContent.LevelID,
		TimeLimitSeconds: fightContent.TimeLimitSeconds,
		Arena:            fightContent.Arena,
		Player:           player,
		Enemy:            fightContent.Waves[0].Enemy,
	}, nil
}
