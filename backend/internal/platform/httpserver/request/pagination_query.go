package request

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

func PaginationFromQuery(context *gin.Context) (pagination.Request, error) {
	paginationRequest := pagination.Request{
		Cursor: context.Query("cursor"),
		Limit:  pagination.DefaultLimit,
	}

	rawLimit := context.Query("limit")
	if rawLimit == "" {
		return paginationRequest, nil
	}

	limit, err := strconv.Atoi(rawLimit)
	if err != nil || limit < 1 || limit > pagination.MaximumLimit {
		return pagination.Request{}, apierror.ValidationFailed(map[string]string{
			"limit": fmt.Sprintf("must be a whole number from 1 to %d", pagination.MaximumLimit),
		})
	}
	paginationRequest.Limit = limit
	return paginationRequest, nil
}

func DecodeCursorPosition[Position any](paginationRequest pagination.Request) (Position, error) {
	position, err := pagination.DecodeCursor[Position](paginationRequest.Cursor)
	if err != nil {
		return position, apierror.ValidationFailed(map[string]string{"cursor": "is invalid or expired"})
	}
	return position, nil
}
