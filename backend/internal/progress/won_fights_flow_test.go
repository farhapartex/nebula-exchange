package progress_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type wonFightsBody struct {
	Data []struct {
		ID    string `json:"id"`
		Level struct {
			ID            string `json:"id"`
			Number        int    `json:"number"`
			Title         string `json:"title"`
			ChapterNumber int    `json:"chapter_number"`
			ChapterTitle  string `json:"chapter_title"`
		} `json:"level"`
		Stars       *int   `json:"stars"`
		DurationMS  int    `json:"duration_ms"`
		DamageDealt int    `json:"damage_dealt"`
		DamageTaken int    `json:"damage_taken"`
		FinishedAt  string `json:"finished_at"`
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

func (harness *progressHarness) listWonFights(query string, accessToken string) (int, wonFightsBody) {
	harness.t.Helper()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/fight-sessions"+query, nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := httptest.NewRecorder()
	harness.router.ServeHTTP(recorder, request)
	var body wonFightsBody
	_ = json.Unmarshal(recorder.Body.Bytes(), &body)
	return recorder.Code, body
}

func (harness *progressHarness) loseLevel(levelID string) {
	harness.t.Helper()
	fightID := harness.startedFightID(levelID)
	harness.clock.advance(95 * time.Second)
	if status, body := harness.submitResult(fightID, `{"outcome":"LOST","duration_ms":90000,"damage_dealt":70,"damage_taken":60}`, "player-token"); status != http.StatusCreated {
		harness.t.Fatalf("lose %s: got %d %v", levelID, status, body)
	}
}

func wonLevelIDs(body wonFightsBody) string {
	levelIDs := make([]string, 0, len(body.Data))
	for _, wonFight := range body.Data {
		levelIDs = append(levelIDs, wonFight.Level.ID)
	}
	return strings.Join(levelIDs, ",")
}

func TestWonFightsAreListedByChapterNewestFirst(t *testing.T) {
	harness := newProgressHarness(t)
	harness.winWithAResult("1-1")
	harness.clock.advance(time.Minute)
	harness.loseLevel("1-2")
	harness.winWithAResult("1-2")
	harness.clock.advance(time.Minute)
	harness.winWithAResult("2-1")
	harness.clock.advance(time.Minute)
	harness.winWithAResult("1-1")
	harness.requireStart("1-2", http.StatusCreated)

	status, body := harness.listWonFights("?outcome=WON", "player-token")
	if status != http.StatusOK || body.Pagination.Limit != 20 || body.Pagination.NextCursor != nil {
		t.Fatalf("got %d %+v", status, body.Pagination)
	}
	if wonLevelIDs(body) != "1-1,1-2,1-1,2-1" {
		t.Fatalf("got %s, want chapter 1 newest first, then chapter 2; losses and open fights stay out", wonLevelIDs(body))
	}
	newestWin := body.Data[0]
	if newestWin.Level.ChapterNumber != 1 || newestWin.Level.ChapterTitle != "The night they came" || newestWin.Level.Number != 1 || newestWin.Stars == nil || *newestWin.Stars != 3 || newestWin.DurationMS != 40000 || newestWin.DamageDealt != 96 || newestWin.DamageTaken != 30 || newestWin.FinishedAt == "" {
		t.Fatalf("unexpected newest win %+v", newestWin)
	}
	if !(body.Data[0].FinishedAt > body.Data[2].FinishedAt) {
		t.Fatal("wins inside a chapter must be newest first")
	}

	if _, strangerBody := harness.listWonFights("?outcome=WON", "stranger-token"); len(strangerBody.Data) != 0 {
		t.Fatal("another player must not see these wins")
	}
}

func TestWonFightsPageAcrossChapters(t *testing.T) {
	harness := newProgressHarness(t)
	harness.winWithAResult("1-1")
	harness.winWithAResult("1-2")
	harness.winWithAResult("2-1")
	harness.winWithAResult("1-1")

	var pagedLevelIDs []string
	query := "?outcome=WON&limit=1"
	for pageNumber := 1; pageNumber <= 5; pageNumber++ {
		_, page := harness.listWonFights(query, "player-token")
		if len(page.Data) == 0 {
			break
		}
		pagedLevelIDs = append(pagedLevelIDs, wonLevelIDs(page))
		if page.Pagination.NextCursor == nil {
			break
		}
		query = "?outcome=WON&limit=1&cursor=" + *page.Pagination.NextCursor
	}
	if strings.Join(pagedLevelIDs, ",") != "1-1,1-2,1-1,2-1" {
		t.Fatalf("paging one by one got %v", pagedLevelIDs)
	}
}

func TestWonFightsNeedTheWonFilterAndValidPaging(t *testing.T) {
	harness := newProgressHarness(t)
	if status, body := harness.listWonFights("", "player-token"); status != http.StatusUnprocessableEntity || body.Error.Details["outcome"] == "" {
		t.Fatalf("without outcome got %d %+v", status, body.Error)
	}
	if status, _ := harness.listWonFights("?outcome=LOST", "player-token"); status != http.StatusUnprocessableEntity {
		t.Fatalf("outcome=LOST got %d", status)
	}
	if status, body := harness.listWonFights("?outcome=WON&cursor=broken", "player-token"); status != http.StatusUnprocessableEntity || body.Error.Details["cursor"] == "" {
		t.Fatalf("a broken cursor got %d", status)
	}
	if status, _ := harness.listWonFights("?outcome=WON", ""); status != http.StatusUnauthorized {
		t.Fatalf("without a login got %d", status)
	}
	if status, body := harness.listWonFights("?outcome=WON", "player-token"); status != http.StatusOK || len(body.Data) != 0 {
		t.Fatalf("a player with no wins got %d %+v", status, body.Data)
	}
}
