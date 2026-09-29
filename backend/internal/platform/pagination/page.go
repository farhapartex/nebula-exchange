package pagination

const (
	DefaultLimit = 20
	MaximumLimit = 100
)

type Request struct {
	Cursor string
	Limit  int
}

func (request Request) HasCursor() bool {
	return request.Cursor != ""
}

func (request Request) FetchLimit() int {
	return request.Limit + 1
}

type PageInfo struct {
	NextCursor *string `json:"next_cursor"`
	Limit      int     `json:"limit"`
}

type Page[Item any] struct {
	Items []Item
	Info  PageInfo
}

func BuildPage[Item any, Position any](fetchedItems []Item, request Request, positionOf func(Item) Position) (Page[Item], error) {
	if len(fetchedItems) <= request.Limit {
		return Page[Item]{Items: fetchedItems, Info: PageInfo{Limit: request.Limit}}, nil
	}

	pageItems := fetchedItems[:request.Limit]
	nextCursor, err := EncodeCursor(positionOf(pageItems[len(pageItems)-1]))
	if err != nil {
		return Page[Item]{}, err
	}
	return Page[Item]{Items: pageItems, Info: PageInfo{NextCursor: &nextCursor, Limit: request.Limit}}, nil
}
