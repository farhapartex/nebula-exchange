package catalog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/platform/database/databasetest"
	"nebula-exchange/backend/internal/platform/httpserver"
	"nebula-exchange/backend/internal/platform/logger"
)

type listEnvelope[Entry any] struct {
	Data       []Entry `json:"data"`
	Pagination struct {
		NextCursor *string `json:"next_cursor"`
		Limit      int     `json:"limit"`
	} `json:"pagination"`
}

func newCatalogRouter(t *testing.T) http.Handler {
	t.Helper()
	pool := databasetest.NewPool(t)
	testLogger := logger.NewWithWriter(&bytes.Buffer{}, slog.LevelError, true)
	service := catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)
	return httpserver.NewRouter(httpserver.RouterOptions{Logger: testLogger}, catalog.NewHandler(service))
}

func getJSON(t *testing.T, router http.Handler, path string, destination any) int {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1"+path, nil))
	if destination != nil && recorder.Code == http.StatusOK {
		if err := json.Unmarshal(recorder.Body.Bytes(), destination); err != nil {
			t.Fatalf("decode %s: %v (%s)", path, err, recorder.Body.String())
		}
	}
	return recorder.Code
}

func TestSeededItemsArePagedByCursorUntilExhausted(t *testing.T) {
	router := newCatalogRouter(t)

	var collectedIDs []int
	path := "/items?limit=7"
	for pageNumber := 0; pageNumber < 10; pageNumber++ {
		var page listEnvelope[map[string]any]
		if status := getJSON(t, router, path, &page); status != http.StatusOK {
			t.Fatalf("page %d status %d", pageNumber, status)
		}
		for _, item := range page.Data {
			collectedIDs = append(collectedIDs, int(item["id"].(float64)))
		}
		if page.Pagination.NextCursor == nil {
			break
		}
		path = "/items?limit=7&cursor=" + *page.Pagination.NextCursor
	}

	if len(collectedIDs) != 21 {
		t.Fatalf("collected %d items, want the 21 seeded items: %v", len(collectedIDs), collectedIDs)
	}
	for index := 1; index < len(collectedIDs); index++ {
		if collectedIDs[index] <= collectedIDs[index-1] {
			t.Fatalf("items out of order: %v", collectedIDs)
		}
	}
}

func TestItemDetailCarriesAttributesAndNullableFields(t *testing.T) {
	router := newCatalogRouter(t)

	var envelope struct {
		Data catalog.Item `json:"data"`
	}
	if status := getJSON(t, router, "/items/10001", &envelope); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	legendaryRelic := envelope.Data
	if legendaryRelic.Category != catalog.CategoryLegendary || legendaryRelic.MaxSupply == nil || *legendaryRelic.MaxSupply != 1 {
		t.Fatalf("unexpected legendary item: %+v", legendaryRelic)
	}
	if !legendaryRelic.IsAuctionOnly {
		t.Fatal("the legendary item is auction only")
	}

	var ironEnvelope struct {
		Data catalog.Item `json:"data"`
	}
	getJSON(t, router, "/items/1", &ironEnvelope)
	if ironEnvelope.Data.Tier != nil || ironEnvelope.Data.MaxSupply != nil {
		t.Fatalf("raw resources have no tier or supply cap: %+v", ironEnvelope.Data)
	}

	if status := getJSON(t, router, "/items/999999", nil); status != http.StatusNotFound {
		t.Fatalf("missing item status %d", status)
	}
	if status := getJSON(t, router, "/items/not-a-number", nil); status != http.StatusNotFound {
		t.Fatalf("malformed item id status %d", status)
	}
}

func TestMoneyTravelsAsMicroUnitStrings(t *testing.T) {
	router := newCatalogRouter(t)

	var recipePage listEnvelope[map[string]any]
	getJSON(t, router, "/recipes?limit=100", &recipePage)
	if len(recipePage.Data) == 0 {
		t.Fatal("recipes should be seeded")
	}
	for _, recipe := range recipePage.Data {
		if _, isString := recipe["fee"].(string); !isString {
			t.Fatalf("recipe fee must be a string: %v", recipe)
		}
		if len(recipe["inputs"].([]any)) == 0 {
			t.Fatalf("recipe %v has no inputs", recipe["id"])
		}
	}

	var upgradePage listEnvelope[map[string]any]
	getJSON(t, router, "/upgrades?limit=100", &upgradePage)
	foundUnbuyableUpgrade := false
	for _, upgrade := range upgradePage.Data {
		if upgrade["buy_price"] == nil {
			foundUnbuyableUpgrade = true
		}
	}
	if !foundUnbuyableUpgrade {
		t.Fatal("the top drill upgrade has no buy price and must be null")
	}

	var shopPage listEnvelope[catalog.ShopItem]
	getJSON(t, router, "/shop-items", &shopPage)
	if len(shopPage.Data) != 4 || shopPage.Data[0].Price <= 0 {
		t.Fatalf("unexpected shop items: %+v", shopPage.Data)
	}
}

func TestZonesListLootAndShipRestrictions(t *testing.T) {
	router := newCatalogRouter(t)

	var zonePage listEnvelope[catalog.Zone]
	getJSON(t, router, "/zones", &zonePage)
	if len(zonePage.Data) != 4 {
		t.Fatalf("got %d zones", len(zonePage.Data))
	}
	deepVoid := zonePage.Data[len(zonePage.Data)-1]
	if deepVoid.ID != "deep-void" || len(deepVoid.AllowedShipItemIDs) != 2 || len(deepVoid.Loot) == 0 {
		t.Fatalf("unexpected deep void zone: %+v", deepVoid)
	}
}

func TestInvalidPaginationIsRejected(t *testing.T) {
	router := newCatalogRouter(t)
	for _, path := range []string{"/items?limit=0", "/items?limit=101", "/items?cursor=%25%25%25", "/zones?cursor=eyJhZnRlciI6Im5vd2hlcmUifQ"} {
		if status := getJSON(t, router, path, nil); status != http.StatusUnprocessableEntity {
			t.Fatalf("%s status %d, want 422", path, status)
		}
	}
}

type flakySource struct {
	loadCount int
	failAfter int
}

func (source *flakySource) Load(context.Context) (catalog.Snapshot, error) {
	source.loadCount++
	if source.loadCount > source.failAfter {
		return catalog.Snapshot{}, errors.New("database unavailable")
	}
	return catalog.Snapshot{Items: []catalog.Item{{ID: source.loadCount}}}, nil
}

func TestSnapshotIsCachedAndServedStaleWhenReloadFails(t *testing.T) {
	currentTime := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	source := &flakySource{failAfter: 1}
	service := catalog.NewService(source, time.Minute, func() time.Time { return currentTime })

	firstSnapshot, err := service.Snapshot(context.Background())
	if err != nil || firstSnapshot.Items[0].ID != 1 {
		t.Fatalf("first load: %v %+v", err, firstSnapshot)
	}
	service.Snapshot(context.Background())
	if source.loadCount != 1 {
		t.Fatalf("cached snapshot should not reload, loads=%d", source.loadCount)
	}

	currentTime = currentTime.Add(2 * time.Minute)
	staleSnapshot, err := service.Snapshot(context.Background())
	if err != nil || staleSnapshot.Items[0].ID != 1 || source.loadCount != 2 {
		t.Fatalf("stale fallback: %v %+v loads=%d", err, staleSnapshot, source.loadCount)
	}

	emptyService := catalog.NewService(&flakySource{failAfter: 0}, time.Minute, time.Now)
	if _, err := emptyService.Snapshot(context.Background()); err == nil {
		t.Fatal("with nothing cached the load error must surface")
	}
}

func TestItemDetailExplainsWhereAnItemComesFromAndGoes(t *testing.T) {
	router := newCatalogRouter(t)

	var crystal struct {
		Data catalog.ItemDetail `json:"data"`
	}
	getJSON(t, router, "/items/3", &crystal)
	usage := crystal.Data.Usage
	if len(usage.InputToRecipes) != 2 || len(usage.DroppedInZones) != 2 || len(usage.CraftedBy) != 0 {
		t.Fatalf("crystal usage %+v", usage)
	}

	var powerCore struct {
		Data catalog.ItemDetail `json:"data"`
	}
	getJSON(t, router, "/items/103", &powerCore)
	if len(powerCore.Data.Usage.CraftedBy) != 1 || len(powerCore.Data.Usage.InputToUpgrades) != 5 || len(powerCore.Data.Usage.InputToRecipes) != 1 {
		t.Fatalf("power core usage %+v", powerCore.Data.Usage)
	}

	var scout struct {
		Data catalog.ItemDetail `json:"data"`
	}
	getJSON(t, router, "/items/301", &scout)
	if len(scout.Data.Usage.UpgradesInto) != 2 || len(scout.Data.Usage.SoldAsSKUs) != 2 {
		t.Fatalf("scout usage %+v", scout.Data.Usage)
	}
}
