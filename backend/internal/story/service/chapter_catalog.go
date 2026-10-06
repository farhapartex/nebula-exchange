package service

import "context"

type ChapterForSale struct {
	ID         string
	Number     int
	Title      string
	IsFree     bool
	PriceCents *int64
	LevelCount int
}

func (catalog *levelCatalog) ChaptersOnSale(ctx context.Context) ([]ChapterForSale, error) {
	orderedPlacements, err := catalog.OrderedPlacements(ctx)
	if err != nil {
		return nil, err
	}
	var chapters []ChapterForSale
	chapterIndexByID := map[string]int{}
	for _, placement := range orderedPlacements {
		chapterIndex, isKnown := chapterIndexByID[placement.ChapterID]
		if !isKnown {
			chapterIndex = len(chapters)
			chapterIndexByID[placement.ChapterID] = chapterIndex
			chapters = append(chapters, ChapterForSale{
				ID:         placement.ChapterID,
				Number:     placement.ChapterNumber,
				Title:      placement.ChapterTitle,
				IsFree:     placement.IsChapterFree,
				PriceCents: placement.ChapterPrice,
			})
		}
		chapters[chapterIndex].LevelCount++
	}
	return chapters, nil
}
