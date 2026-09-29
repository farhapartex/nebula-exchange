package catalog

import (
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/pagination"
)

type catalogCursor struct {
	AfterKey string `json:"after"`
}

func paginateInOrder[Entry any](orderedEntries []Entry, pageRequest pagination.Request, keyOf func(Entry) string) (pagination.Page[Entry], error) {
	startIndex := 0
	if pageRequest.HasCursor() {
		cursorPosition, err := request.DecodeCursorPosition[catalogCursor](pageRequest)
		if err != nil {
			return pagination.Page[Entry]{}, err
		}
		startIndex = -1
		for entryIndex, entry := range orderedEntries {
			if keyOf(entry) == cursorPosition.AfterKey {
				startIndex = entryIndex + 1
				break
			}
		}
		if startIndex < 0 {
			return pagination.Page[Entry]{}, apierror.ValidationFailed(map[string]string{"cursor": "is invalid or expired"})
		}
	}

	endIndex := min(startIndex+pageRequest.FetchLimit(), len(orderedEntries))
	fetchedEntries := append([]Entry{}, orderedEntries[startIndex:endIndex]...)
	return pagination.BuildPage(fetchedEntries, pageRequest, func(entry Entry) catalogCursor {
		return catalogCursor{AfterKey: keyOf(entry)}
	})
}
