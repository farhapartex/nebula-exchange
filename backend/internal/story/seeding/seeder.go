package seeding

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/objectstorage"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

var slideIDNamespace = uuid.MustParse("6f1d4b9e-2c0a-4e8b-9a57-3d2f1c7b8e40")

type Seeder struct {
	database *gorm.DB
	storage  objectstorage.Uploader
	logger   *slog.Logger
}

func NewSeeder(database *gorm.DB, storage objectstorage.Uploader, logger *slog.Logger) *Seeder {
	return &Seeder{database: database, storage: storage, logger: logger}
}

func (seeder *Seeder) SeedLevel(ctx context.Context, levelPackage LevelPackage) error {
	if err := seeder.uploadImages(ctx, levelPackage); err != nil {
		return err
	}
	return seeder.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := upsert(transaction, chapterRecord(levelPackage.Chapter), "number", "title", "summary", "is_free", "price_coins", "is_published"); err != nil {
			return fmt.Errorf("save chapter: %w", err)
		}
		if err := upsert(transaction, arenaRecord(levelPackage.Arena), "name", "width", "floor_y", "stage"); err != nil {
			return fmt.Errorf("save arena: %w", err)
		}
		levelColumns := []string{"chapter_id", "number", "kind", "title", "teaser", "arena_id", "time_limit_seconds", "difficulty", "star_rules", "first_clear_coins", "first_clear_experience", "replay_experience", "is_published"}
		if err := upsert(transaction, levelRecord(levelPackage), levelColumns...); err != nil {
			return fmt.Errorf("save level: %w", err)
		}
		if err := transaction.Where(map[string]any{"level_id": levelPackage.Level.ID}).Delete(&models.StorySlide{}).Error; err != nil {
			return fmt.Errorf("clear old slides: %w", err)
		}
		if err := transaction.Create(slideRecords(levelPackage)).Error; err != nil {
			return fmt.Errorf("save slides: %w", err)
		}
		seeder.logger.InfoContext(ctx, "level story seeded", slog.String("level_id", levelPackage.Level.ID), slog.Int("slides", len(levelPackage.Slides)))
		return nil
	})
}

func (seeder *Seeder) uploadImages(ctx context.Context, levelPackage LevelPackage) error {
	if err := seeder.storage.EnsureBucket(ctx); err != nil {
		return err
	}
	for _, slide := range levelPackage.Slides {
		if slide.Image == nil {
			continue
		}
		if err := seeder.uploadImage(ctx, levelPackage, *slide.Image); err != nil {
			return err
		}
	}
	return nil
}

func (seeder *Seeder) uploadImage(ctx context.Context, levelPackage LevelPackage, imageFileName string) error {
	imageFile, err := os.Open(levelPackage.ImagePath(imageFileName))
	if err != nil {
		return err
	}
	defer imageFile.Close()
	imageInfo, err := imageFile.Stat()
	if err != nil {
		return err
	}
	objectKey := levelPackage.ImageObjectKey(imageFileName)
	contentType := mime.TypeByExtension(filepath.Ext(imageFileName))
	if err := seeder.storage.Upload(ctx, objectKey, imageFile, imageInfo.Size(), contentType); err != nil {
		return err
	}
	seeder.logger.InfoContext(ctx, "image uploaded", slog.String("object_key", objectKey), slog.Int64("bytes", imageInfo.Size()))
	return nil
}

func upsert(transaction *gorm.DB, record any, updatedColumns ...string) error {
	return transaction.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns(append(updatedColumns, "updated_at")),
	}).Create(record).Error
}

func chapterRecord(chapter ChapterContent) *models.Chapter {
	return &models.Chapter{
		ID:          chapter.ID,
		Number:      chapter.Number,
		Title:       chapter.Title,
		Summary:     chapter.Summary,
		IsFree:      chapter.IsFree,
		PriceCoins:  chapter.PriceCoins,
		IsPublished: chapter.IsPublished,
	}
}

func arenaRecord(arena ArenaContent) *models.Arena {
	return &models.Arena{ID: arena.ID, Name: arena.Name, Width: arena.Width, FloorY: arena.FloorY, Stage: arena.Stage}
}

func levelRecord(levelPackage LevelPackage) *models.Level {
	level := levelPackage.Level
	chapterID := levelPackage.Chapter.ID
	levelNumber := level.Number
	return &models.Level{
		ID:                   level.ID,
		ChapterID:            &chapterID,
		Number:               &levelNumber,
		Kind:                 level.Kind,
		Title:                level.Title,
		Teaser:               level.Teaser,
		ArenaID:              levelPackage.Arena.ID,
		TimeLimitSeconds:     level.TimeLimitSeconds,
		Difficulty:           level.Difficulty,
		StarRules:            level.StarRules,
		FirstClearCoins:      level.FirstClearCoins,
		FirstClearExperience: level.FirstClearExperience,
		ReplayExperience:     level.ReplayExperience,
		IsPublished:          level.IsPublished,
	}
}

func slideRecords(levelPackage LevelPackage) []models.StorySlide {
	slides := make([]models.StorySlide, 0, len(levelPackage.Slides))
	for slideIndex, slide := range levelPackage.Slides {
		position := slideIndex + 1
		var imageKey *string
		if slide.Image != nil {
			objectKey := levelPackage.ImageObjectKey(*slide.Image)
			imageKey = &objectKey
		}
		slides = append(slides, models.StorySlide{
			ID:          uuid.NewSHA1(slideIDNamespace, fmt.Appendf(nil, "%s/%d", levelPackage.Level.ID, position)),
			LevelID:     levelPackage.Level.ID,
			Position:    position,
			Kind:        slide.Kind,
			Eyebrow:     slide.Eyebrow,
			Heading:     slide.Heading,
			Body:        slide.Body,
			ImageKey:    imageKey,
			Palette:     slide.Palette,
			ButtonLabel: slide.ButtonLabel,
		})
	}
	return slides
}
