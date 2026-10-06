package service

import (
	"context"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/repository"
)

var ErrLevelStoryNotFound = apierror.NotFound("This level has no story yet")

type ImageURLSigner interface {
	PresignedDownloadURL(ctx context.Context, objectKey string) (string, error)
}

type StoryCursor struct {
	Position int `json:"position"`
}

type StorySlide struct {
	ID          string
	Position    int
	Kind        models.SlideKind
	Eyebrow     *string
	Heading     string
	Body        string
	ImageURL    *string
	Palette     database.JSONDocument
	ButtonLabel *string
}

type StoryService interface {
	ListLevelStory(ctx context.Context, levelID string, after StoryCursor, pageRequest pagination.Request) (pagination.Page[StorySlide], error)
}

type storyService struct {
	levels      repository.LevelRepository
	slides      repository.StorySlideRepository
	imageSigner ImageURLSigner
}

func NewStoryService(levels repository.LevelRepository, slides repository.StorySlideRepository, imageSigner ImageURLSigner) StoryService {
	return &storyService{levels: levels, slides: slides, imageSigner: imageSigner}
}

func (stories *storyService) ListLevelStory(ctx context.Context, levelID string, after StoryCursor, pageRequest pagination.Request) (pagination.Page[StorySlide], error) {
	level, isFound, err := stories.levels.FindPublished(ctx, levelID)
	if err != nil {
		return pagination.Page[StorySlide]{}, err
	}
	if !isFound || !isInPublishedChapter(level) {
		return pagination.Page[StorySlide]{}, ErrLevelStoryNotFound
	}

	storedSlides, err := stories.slides.ListByLevel(ctx, levelID, after.Position, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[StorySlide]{}, err
	}
	storySlides := make([]StorySlide, 0, len(storedSlides))
	for _, storedSlide := range storedSlides {
		imageURL, err := stories.signedImageURL(ctx, storedSlide.ImageKey)
		if err != nil {
			return pagination.Page[StorySlide]{}, err
		}
		storySlides = append(storySlides, StorySlide{
			ID:          storedSlide.ID.String(),
			Position:    storedSlide.Position,
			Kind:        storedSlide.Kind,
			Eyebrow:     storedSlide.Eyebrow,
			Heading:     storedSlide.Heading,
			Body:        storedSlide.Body,
			ImageURL:    imageURL,
			Palette:     storedSlide.Palette,
			ButtonLabel: storedSlide.ButtonLabel,
		})
	}
	return pagination.BuildPage(storySlides, pageRequest, func(storySlide StorySlide) StoryCursor {
		return StoryCursor{Position: storySlide.Position}
	})
}

func (stories *storyService) signedImageURL(ctx context.Context, imageKey *string) (*string, error) {
	if imageKey == nil {
		return nil, nil
	}
	signedURL, err := stories.imageSigner.PresignedDownloadURL(ctx, *imageKey)
	if err != nil {
		return nil, err
	}
	return &signedURL, nil
}

func isInPublishedChapter(level models.Level) bool {
	return level.Chapter != nil && level.Chapter.IsPublished
}
