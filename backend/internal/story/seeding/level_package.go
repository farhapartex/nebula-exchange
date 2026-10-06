package seeding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
)

const (
	levelFileName       = "level.json"
	imagesDirectoryName = "images"
)

type ChapterContent struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	IsFree      bool   `json:"is_free"`
	PriceCoins  *int64 `json:"price_coins"`
	IsPublished bool   `json:"is_published"`
}

type ArenaContent struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	Width  int                   `json:"width"`
	FloorY int                   `json:"floor_y"`
	Stage  database.JSONDocument `json:"stage"`
}

type LevelContent struct {
	ID                   string                `json:"id"`
	Number               int                   `json:"number"`
	Kind                 models.LevelKind      `json:"kind"`
	Title                string                `json:"title"`
	Teaser               string                `json:"teaser"`
	TimeLimitSeconds     int                   `json:"time_limit_seconds"`
	Difficulty           database.JSONDocument `json:"difficulty"`
	StarRules            database.JSONDocument `json:"star_rules"`
	FirstClearCoins      int64                 `json:"first_clear_coins"`
	FirstClearExperience int                   `json:"first_clear_experience"`
	ReplayExperience     int                   `json:"replay_experience"`
	IsPublished          bool                  `json:"is_published"`
}

type EnemyContent struct {
	ID    string                `json:"id"`
	Name  string                `json:"name"`
	Title string                `json:"title"`
	Stats database.JSONDocument `json:"stats"`
	Brain database.JSONDocument `json:"brain"`
	Look  database.JSONDocument `json:"look"`
}

type WaveContent struct {
	Wave      int                   `json:"wave"`
	Enemy     EnemyContent          `json:"enemy"`
	Modifiers database.JSONDocument `json:"modifiers"`
	IntroLine *string               `json:"intro_line"`
}

type SlideContent struct {
	Kind        models.SlideKind      `json:"kind"`
	Eyebrow     *string               `json:"eyebrow"`
	Heading     string                `json:"heading"`
	Body        string                `json:"body"`
	Image       *string               `json:"image"`
	Palette     database.JSONDocument `json:"palette"`
	ButtonLabel *string               `json:"button_label"`
}

type LevelPackage struct {
	Directory string         `json:"-"`
	Chapter   ChapterContent `json:"chapter"`
	Arena     ArenaContent   `json:"arena"`
	Level     LevelContent   `json:"level"`
	Enemies   []WaveContent  `json:"enemies"`
	Slides    []SlideContent `json:"slides"`
}

func LoadLevelPackage(directory string) (LevelPackage, error) {
	levelFile, err := os.ReadFile(filepath.Join(directory, levelFileName))
	if err != nil {
		return LevelPackage{}, fmt.Errorf("read %s: %w", levelFileName, err)
	}
	var levelPackage LevelPackage
	if err := json.Unmarshal(levelFile, &levelPackage); err != nil {
		return LevelPackage{}, fmt.Errorf("parse %s: %w", levelFileName, err)
	}
	levelPackage.Directory = directory
	return levelPackage, levelPackage.validate()
}

func (levelPackage LevelPackage) ImagePath(imageFileName string) string {
	return filepath.Join(levelPackage.Directory, imagesDirectoryName, imageFileName)
}

func (levelPackage LevelPackage) ImageObjectKey(imageFileName string) string {
	return "story/" + levelPackage.Level.ID + "/" + imageFileName
}

func (levelPackage LevelPackage) validate() error {
	if levelPackage.Chapter.ID == "" || levelPackage.Arena.ID == "" || levelPackage.Level.ID == "" {
		return errors.New("chapter, arena and level all need an id")
	}
	for waveIndex, wave := range levelPackage.Enemies {
		if wave.Wave != waveIndex+1 || wave.Enemy.ID == "" {
			return fmt.Errorf("enemy wave %d: waves must be numbered 1, 2, 3 in order and name an enemy id", waveIndex+1)
		}
	}
	if len(levelPackage.Slides) < 2 {
		return errors.New("a level story needs at least one slide and a call to action")
	}
	for slideIndex, slide := range levelPackage.Slides {
		isLastSlide := slideIndex == len(levelPackage.Slides)-1
		if isLastSlide != (slide.Kind == models.SlideKindCallToAction) {
			return fmt.Errorf("slide %d: only the last slide must be the call to action", slideIndex+1)
		}
		if slide.Image == nil {
			continue
		}
		if filepath.Base(*slide.Image) != *slide.Image {
			return fmt.Errorf("slide %d: image %q must be a file name inside %s", slideIndex+1, *slide.Image, imagesDirectoryName)
		}
		if _, err := os.Stat(levelPackage.ImagePath(*slide.Image)); err != nil {
			return fmt.Errorf("slide %d: %w", slideIndex+1, err)
		}
	}
	return nil
}
