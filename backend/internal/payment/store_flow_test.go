package payment_test

import (
	"net/http"
	"testing"
	"time"

	paymentseeding "github.com/farhapartex/nebula-exchange/backend/internal/payment/seeding"
	progressmodels "github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

func (harness *paymentHarness) plansByID() map[string]map[string]any {
	harness.t.Helper()
	status, body := harness.get("/plans", playerToken)
	if status != http.StatusOK {
		harness.t.Fatalf("plans: got %d", status)
	}
	plans := map[string]map[string]any{}
	for _, plan := range body.Data {
		plans[plan["id"].(string)] = plan
	}
	return plans
}

func optionTotals(plan map[string]any) map[float64]string {
	totals := map[float64]string{}
	for _, option := range plan["options"].([]any) {
		optionData := option.(map[string]any)
		totals[optionData["chapter_count"].(float64)] = optionData["total_cents"].(string)
	}
	return totals
}

func TestChaptersListPricesAndOwnershipWithPagination(t *testing.T) {
	harness := newPaymentHarness(t)

	status, firstPage := harness.get("/chapters?limit=3", playerToken)
	if status != http.StatusOK || len(firstPage.Data) != 3 || firstPage.Pagination.NextCursor == nil {
		t.Fatalf("got %d with %d chapters and cursor %v", status, len(firstPage.Data), firstPage.Pagination.NextCursor)
	}
	prologue, secondChapter := firstPage.Data[0], firstPage.Data[1]
	if prologue["is_free"] != true || prologue["price_cents"] != nil || secondChapter["price_cents"] != "499" || secondChapter["is_owned"] != false || secondChapter["level_count"] != float64(1) {
		t.Fatalf("unexpected chapters %v", firstPage.Data)
	}
	_, secondPage := harness.get("/chapters?limit=3&cursor="+*firstPage.Pagination.NextCursor, playerToken)
	if len(secondPage.Data) != 2 || secondPage.Data[0]["number"] != float64(4) || secondPage.Pagination.NextCursor != nil {
		t.Fatalf("unexpected second page %v", secondPage)
	}
	if status, _ := harness.get("/chapters", ""); status != http.StatusUnauthorized {
		t.Fatalf("got %d without a login, want 401", status)
	}
}

func TestPlansArePricedOnTheServerForTheChaptersStillToBuy(t *testing.T) {
	harness := newPaymentHarness(t)
	plans := harness.plansByID()

	if len(plans) != 3 {
		t.Fatalf("got %d plans, want 3", len(plans))
	}
	if totals := optionTotals(plans["single-chapter"]); len(totals) != 1 || totals[1] != "499" {
		t.Fatalf("unexpected single chapter totals %v", totals)
	}
	if totals := optionTotals(plans["chapter-bundle"]); len(totals) != 2 || totals[2] != "948" || totals[3] != "1347" {
		t.Fatalf("unexpected bundle totals %v", totals)
	}
	allOption := plans["all-chapters"]["options"].([]any)[0].(map[string]any)
	if allOption["chapter_count"] != float64(4) || allOption["subtotal_cents"] != "1996" || allOption["discount_cents"] != "399" || allOption["total_cents"] != "1597" {
		t.Fatalf("unexpected all chapters option %v", allOption)
	}

	ownedChapter := &progressmodels.ChapterUnlock{UserID: harness.playerID, ChapterID: "2", Source: progressmodels.ChapterUnlockSourceGrant, UnlockedAt: time.Now()}
	if err := harness.database.Create(ownedChapter).Error; err != nil {
		t.Fatalf("own chapter 2: %v", err)
	}
	afterOwning := harness.plansByID()
	singleChapters := afterOwning["single-chapter"]["options"].([]any)[0].(map[string]any)["chapters"].([]any)
	if singleChapters[0].(map[string]any)["number"] != float64(3) {
		t.Fatalf("the single plan must now offer chapter 3, got %v", singleChapters)
	}
	if totals := optionTotals(afterOwning["chapter-bundle"]); len(totals) != 1 || totals[2] != "948" {
		t.Fatalf("with 3 chapters left the bundle can only be 2 chapters, got %v", totals)
	}
	if allOption := afterOwning["all-chapters"]["options"].([]any)[0].(map[string]any); allOption["chapter_count"] != float64(3) {
		t.Fatalf("all chapters must now cover 3 chapters, got %v", allOption)
	}

	for _, chapterID := range []string{"3", "4"} {
		harness.database.Create(&progressmodels.ChapterUnlock{UserID: harness.playerID, ChapterID: chapterID, Source: progressmodels.ChapterUnlockSourceGrant, UnlockedAt: time.Now()})
	}
	oneLeft := harness.plansByID()
	if oneLeft["single-chapter"]["is_available"] != true || oneLeft["chapter-bundle"]["is_available"] != false || oneLeft["all-chapters"]["is_available"] != false {
		t.Fatalf("with one chapter left only the single plan makes sense, got %v", oneLeft)
	}
}

func TestThePlansSeedFileIsValid(t *testing.T) {
	plans, err := paymentseeding.LoadPlans(plansSeedFile())
	if err != nil || len(plans) != 3 {
		t.Fatalf("got %d plans, %v", len(plans), err)
	}
}
