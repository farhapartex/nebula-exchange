package story_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database/databasetest"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/story"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/seeding"
)

const validAccessToken = "player-access-token"

var levelOneHeadings = []string{
	"At the edge of the village",
	"No one left",
	"Torches in the dark",
	"Nothing to give",
	"Fire",
	"Fight to win. Win to become a street fighter.",
}

type recordingStorage struct {
	mutex            sync.Mutex
	uploadedBytes    map[string]int
	uploadedTypes    map[string]string
	hasEnsuredBucket bool
}

func newRecordingStorage() *recordingStorage {
	return &recordingStorage{uploadedBytes: map[string]int{}, uploadedTypes: map[string]string{}}
}

func (storage *recordingStorage) EnsureBucket(context.Context) error {
	storage.hasEnsuredBucket = true
	return nil
}

func (storage *recordingStorage) Upload(_ context.Context, objectKey string, content io.Reader, _ int64, contentType string) error {
	uploadedContent, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	storage.mutex.Lock()
	defer storage.mutex.Unlock()
	storage.uploadedBytes[objectKey] = len(uploadedContent)
	storage.uploadedTypes[objectKey] = contentType
	return nil
}

func (storage *recordingStorage) PresignedDownloadURL(_ context.Context, objectKey string) (string, error) {
	return "http://localhost:9000/street-born-assets/" + objectKey + "?X-Amz-Signature=test", nil
}

type fixedTokenVerifier struct{}

func (fixedTokenVerifier) Verify(accessToken string) (uuid.UUID, error) {
	if accessToken != validAccessToken {
		return uuid.Nil, errors.New("invalid token")
	}
	return uuid.New(), nil
}

type storyListResponse struct {
	Data []struct {
		ID          string         `json:"id"`
		Position    int            `json:"position"`
		Kind        string         `json:"kind"`
		Heading     string         `json:"heading"`
		Body        string         `json:"body"`
		Image       *string        `json:"image"`
		Palette     map[string]any `json:"palette"`
		ButtonLabel *string        `json:"button_label"`
	} `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
		Limit      int     `json:"limit"`
	} `json:"pagination"`
	Error struct {
		Code    string            `json:"code"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func levelOnePackageDirectory() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "seeds", "story", "level-1-1")
}

func seedLevelOne(t *testing.T, testDatabase *gorm.DB, storage *recordingStorage) {
	t.Helper()
	levelPackage, err := seeding.LoadLevelPackage(levelOnePackageDirectory())
	if err != nil {
		t.Fatalf("load level package: %v", err)
	}
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := seeding.NewSeeder(testDatabase, storage, quietLogger).SeedLevel(context.Background(), levelPackage); err != nil {
		t.Fatalf("seed level: %v", err)
	}
}

func newStoryRouter(t *testing.T, testDatabase *gorm.DB, storage *recordingStorage) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	storyModule := story.NewModule(story.ModuleDependencies{Database: testDatabase, ImageSigner: storage})
	router, err := httpserver.NewRouter(httpserver.RouterOptions{
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		AccessTokens: fixedTokenVerifier{},
	}, storyModule.RouteRegistrars()...)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	return router
}

func listStories(router *gin.Engine, query url.Values, accessToken string) (int, storyListResponse) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/stories?"+query.Encode(), nil)
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	var body storyListResponse
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func TestSeedUploadsEveryImageAndTheStoryListsInOrder(t *testing.T) {
	testDatabase := databasetest.Open(t)
	storage := newRecordingStorage()
	seedLevelOne(t, testDatabase, storage)

	if !storage.hasEnsuredBucket || len(storage.uploadedBytes) != 6 {
		t.Fatalf("expected the bucket check and 6 uploads, got %v", storage.uploadedBytes)
	}
	for objectKey, contentType := range storage.uploadedTypes {
		if contentType != "image/webp" || storage.uploadedBytes[objectKey] == 0 {
			t.Fatalf("%s uploaded as %q with %d bytes", objectKey, contentType, storage.uploadedBytes[objectKey])
		}
	}

	status, body := listStories(newStoryRouter(t, testDatabase, storage), url.Values{"level": {"1-1"}}, validAccessToken)
	if status != http.StatusOK || len(body.Data) != len(levelOneHeadings) {
		t.Fatalf("got status %d with %d slides", status, len(body.Data))
	}
	for slideIndex, slide := range body.Data {
		if slide.Position != slideIndex+1 || slide.Heading != levelOneHeadings[slideIndex] || slide.Image == nil || slide.Palette["accent"] == nil {
			t.Fatalf("slide %d: got %+v", slideIndex+1, slide)
		}
	}
	firstSlide, lastSlide := body.Data[0], body.Data[len(body.Data)-1]
	if *firstSlide.Image != "http://localhost:9000/street-born-assets/story/1-1/01-edge-of-the-village.webp?X-Amz-Signature=test" || firstSlide.Kind != "SLIDE" {
		t.Fatalf("unexpected first slide %+v", firstSlide)
	}
	if lastSlide.Kind != "CALL_TO_ACTION" || lastSlide.ButtonLabel == nil || *lastSlide.ButtonLabel != "Play" {
		t.Fatalf("unexpected call to action %+v", lastSlide)
	}
	if body.Pagination.NextCursor != nil || body.Pagination.Limit != 20 {
		t.Fatalf("expected one page with the default limit, got %+v", body.Pagination)
	}
}

func TestStoryPagesFollowTheCursorWithoutGapsOrRepeats(t *testing.T) {
	testDatabase := databasetest.Open(t)
	storage := newRecordingStorage()
	seedLevelOne(t, testDatabase, storage)
	router := newStoryRouter(t, testDatabase, storage)

	var collectedHeadings []string
	query := url.Values{"level": {"1-1"}, "limit": {"4"}}
	for pageNumber := 1; pageNumber <= 3; pageNumber++ {
		status, body := listStories(router, query, validAccessToken)
		if status != http.StatusOK {
			t.Fatalf("page %d: status %d", pageNumber, status)
		}
		for _, slide := range body.Data {
			collectedHeadings = append(collectedHeadings, slide.Heading)
		}
		if body.Pagination.NextCursor == nil {
			break
		}
		query.Set("cursor", *body.Pagination.NextCursor)
	}

	if !slices.Equal(collectedHeadings, levelOneHeadings) {
		t.Fatalf("got %v across pages, want %v", collectedHeadings, levelOneHeadings)
	}
}

func TestSeedingTwiceKeepsOneCopyWithStableSlideIDs(t *testing.T) {
	testDatabase := databasetest.Open(t)
	storage := newRecordingStorage()
	seedLevelOne(t, testDatabase, storage)
	var firstSlideIDs []string
	testDatabase.Model(&models.StorySlide{}).Order("position").Pluck("id", &firstSlideIDs)

	seedLevelOne(t, testDatabase, storage)
	var secondSlideIDs []string
	testDatabase.Model(&models.StorySlide{}).Order("position").Pluck("id", &secondSlideIDs)
	var levelCount int64
	testDatabase.Model(&models.Level{}).Count(&levelCount)

	if levelCount != 1 || len(secondSlideIDs) != 6 || !slices.Equal(firstSlideIDs, secondSlideIDs) {
		t.Fatalf("got %d levels and slide ids %v then %v", levelCount, firstSlideIDs, secondSlideIDs)
	}
}

func TestStoryListRejectsBadRequests(t *testing.T) {
	testDatabase := databasetest.Open(t)
	storage := newRecordingStorage()
	seedLevelOne(t, testDatabase, storage)
	router := newStoryRouter(t, testDatabase, storage)

	if status, _ := listStories(router, url.Values{"level": {"1-1"}}, ""); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login, want 401", status)
	}
	status, body := listStories(router, url.Values{}, validAccessToken)
	if status != http.StatusUnprocessableEntity || body.Error.Details["level"] != "is required" {
		t.Fatalf("got %d %+v without a level, want 422 with a level error", status, body.Error)
	}
	if status, _ := listStories(router, url.Values{"level": {"1-1"}, "cursor": {"not-a-cursor"}}, validAccessToken); status != http.StatusUnprocessableEntity {
		t.Fatalf("got %d for a broken cursor, want 422", status)
	}
	if status, _ := listStories(router, url.Values{"level": {"9-9"}}, validAccessToken); status != http.StatusNotFound {
		t.Fatalf("got %d for an unknown level, want 404", status)
	}
	testDatabase.Model(&models.Level{}).Where(map[string]any{"id": "1-1"}).Update("is_published", false)
	if status, _ := listStories(router, url.Values{"level": {"1-1"}}, validAccessToken); status != http.StatusNotFound {
		t.Fatalf("got %d for an unpublished level, want 404", status)
	}
}
