package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type WonFightLevel struct {
	ID            string
	Number        int
	Title         string
	ChapterNumber int
	ChapterTitle  string
}

type WonFight struct {
	ID          uuid.UUID
	Level       WonFightLevel
	Stars       *int
	DurationMS  int
	DamageDealt int
	DamageTaken int
	FinishedAt  time.Time
}

type WonFightCursor struct {
	ChapterNumber int       `json:"chapter_number"`
	FinishedAt    time.Time `json:"finished_at"`
	ID            uuid.UUID `json:"id"`
}

type WonFightService interface {
	List(ctx context.Context, userID uuid.UUID, after *WonFightCursor, pageRequest pagination.Request) (pagination.Page[WonFight], error)
}

type chapterLevels struct {
	number   int
	title    string
	levelIDs []string
}

type wonFightService struct {
	levels        LevelCatalog
	fightSessions repository.FightSessionRepository
}

func NewWonFightService(levels LevelCatalog, fightSessions repository.FightSessionRepository) WonFightService {
	return &wonFightService{levels: levels, fightSessions: fightSessions}
}

func (wonFights *wonFightService) List(ctx context.Context, userID uuid.UUID, after *WonFightCursor, pageRequest pagination.Request) (pagination.Page[WonFight], error) {
	placements, err := wonFights.levels.OrderedPlacements(ctx)
	if err != nil {
		return pagination.Page[WonFight]{}, err
	}
	placementsByLevel := make(map[string]storyservice.LevelPlacement, len(placements))
	for _, placement := range placements {
		placementsByLevel[placement.LevelID] = placement
	}

	fetchLimit := pageRequest.FetchLimit()
	collectedWins := make([]WonFight, 0, fetchLimit)
	for _, chapter := range chaptersInStoryOrder(placements) {
		if after != nil && chapter.number < after.ChapterNumber {
			continue
		}
		var position *repository.WonFightPosition
		if after != nil && chapter.number == after.ChapterNumber {
			position = &repository.WonFightPosition{FinishedAt: after.FinishedAt, ID: after.ID}
		}
		remaining := fetchLimit - len(collectedWins)
		if remaining == 0 {
			break
		}
		chapterWins, err := wonFights.fightSessions.ListWonInLevels(ctx, userID, chapter.levelIDs, position, remaining)
		if err != nil {
			return pagination.Page[WonFight]{}, err
		}
		for _, chapterWin := range chapterWins {
			placement := placementsByLevel[chapterWin.LevelID]
			collectedWins = append(collectedWins, WonFight{
				ID: chapterWin.ID,
				Level: WonFightLevel{
					ID:            placement.LevelID,
					Number:        placement.LevelNumber,
					Title:         placement.Title,
					ChapterNumber: placement.ChapterNumber,
					ChapterTitle:  placement.ChapterTitle,
				},
				Stars:       chapterWin.Stars,
				DurationMS:  valueOrZero(chapterWin.DurationMS),
				DamageDealt: valueOrZero(chapterWin.DamageDealt),
				DamageTaken: valueOrZero(chapterWin.DamageTaken),
				FinishedAt:  *chapterWin.FinishedAt,
			})
		}
	}
	return pagination.BuildPage(collectedWins, pageRequest, func(wonFight WonFight) WonFightCursor {
		return WonFightCursor{ChapterNumber: wonFight.Level.ChapterNumber, FinishedAt: wonFight.FinishedAt, ID: wonFight.ID}
	})
}

func chaptersInStoryOrder(placements []storyservice.LevelPlacement) []chapterLevels {
	var chapters []chapterLevels
	for _, placement := range placements {
		if len(chapters) == 0 || chapters[len(chapters)-1].number != placement.ChapterNumber {
			chapters = append(chapters, chapterLevels{number: placement.ChapterNumber, title: placement.ChapterTitle})
		}
		lastChapter := &chapters[len(chapters)-1]
		lastChapter.levelIDs = append(lastChapter.levelIDs, placement.LevelID)
	}
	return chapters
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
