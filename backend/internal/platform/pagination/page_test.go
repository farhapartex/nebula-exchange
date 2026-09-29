package pagination

import (
	"errors"
	"testing"
)

type orderPosition struct {
	CreatedAt string `json:"created_at"`
	ID        int    `json:"id"`
}

type testOrder struct {
	ID        int
	CreatedAt string
}

func positionOfOrder(order testOrder) orderPosition {
	return orderPosition{CreatedAt: order.CreatedAt, ID: order.ID}
}

func TestCursorRoundTrip(t *testing.T) {
	originalPosition := orderPosition{CreatedAt: "2026-09-29T10:00:00Z", ID: 42}

	cursor, err := EncodeCursor(originalPosition)
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}
	decodedPosition, err := DecodeCursor[orderPosition](cursor)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if decodedPosition != originalPosition {
		t.Fatalf("got %+v, want %+v", decodedPosition, originalPosition)
	}
}

func TestDecodeCursorRejectsGarbage(t *testing.T) {
	for _, garbageCursor := range []string{"%%%", "bm90LWpzb24"} {
		if _, err := DecodeCursor[orderPosition](garbageCursor); !errors.Is(err, ErrInvalidCursor) {
			t.Fatalf("cursor %q: got %v, want ErrInvalidCursor", garbageCursor, err)
		}
	}
}

func TestBuildPageWithoutMoreItemsHasNoNextCursor(t *testing.T) {
	fetchedOrders := []testOrder{{ID: 1}, {ID: 2}}

	page, err := BuildPage(fetchedOrders, Request{Limit: 2}, positionOfOrder)
	if err != nil {
		t.Fatalf("build page: %v", err)
	}
	if len(page.Items) != 2 || page.Info.NextCursor != nil || page.Info.Limit != 2 {
		t.Fatalf("unexpected page %+v", page)
	}
}

func TestBuildPageWithMoreItemsTrimsAndPointsAtLastItem(t *testing.T) {
	paginationRequest := Request{Limit: 2}
	fetchedOrders := []testOrder{{ID: 1, CreatedAt: "a"}, {ID: 2, CreatedAt: "b"}, {ID: 3, CreatedAt: "c"}}
	if paginationRequest.FetchLimit() != len(fetchedOrders) {
		t.Fatalf("fetch limit should be one more than the page limit")
	}

	page, err := BuildPage(fetchedOrders, paginationRequest, positionOfOrder)
	if err != nil {
		t.Fatalf("build page: %v", err)
	}
	if len(page.Items) != 2 || page.Info.NextCursor == nil {
		t.Fatalf("unexpected page %+v", page)
	}

	nextPosition, err := DecodeCursor[orderPosition](*page.Info.NextCursor)
	if err != nil {
		t.Fatalf("decode next cursor: %v", err)
	}
	if nextPosition != (orderPosition{CreatedAt: "b", ID: 2}) {
		t.Fatalf("next cursor points at %+v, want the last returned item", nextPosition)
	}
}
