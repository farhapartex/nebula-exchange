package request

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

func newQueryContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return context
}

func requireAPIError(t *testing.T, err error, expectedStatus int, expectedCode apierror.Code) *apierror.Error {
	t.Helper()
	var apiError *apierror.Error
	if !errors.As(err, &apiError) {
		t.Fatalf("expected *apierror.Error, got %v", err)
	}
	if apiError.StatusCode != expectedStatus || apiError.Code != expectedCode {
		t.Fatalf("got %d %s, want %d %s", apiError.StatusCode, apiError.Code, expectedStatus, expectedCode)
	}
	return apiError
}

func TestPaginationFromQueryUsesDefaultLimit(t *testing.T) {
	paginationRequest, err := PaginationFromQuery(newQueryContext("/stories"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paginationRequest.Limit != pagination.DefaultLimit || paginationRequest.HasCursor() {
		t.Fatalf("unexpected pagination request %+v", paginationRequest)
	}
}

func TestPaginationFromQueryReadsCursorAndLimit(t *testing.T) {
	paginationRequest, err := PaginationFromQuery(newQueryContext("/stories?cursor=abc&limit=50"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paginationRequest.Limit != 50 || paginationRequest.Cursor != "abc" {
		t.Fatalf("unexpected pagination request %+v", paginationRequest)
	}
}

func TestPaginationFromQueryRejectsOutOfRangeLimits(t *testing.T) {
	for _, rawLimit := range []string{"0", "-1", "101", "ten"} {
		_, err := PaginationFromQuery(newQueryContext("/stories?limit=" + rawLimit))
		requireAPIError(t, err, http.StatusUnprocessableEntity, apierror.CodeValidationFailed)
	}
}

func TestDecodeCursorPositionRejectsInvalidCursor(t *testing.T) {
	_, err := DecodeCursorPosition[map[string]int](pagination.Request{Cursor: "%%%"})
	requireAPIError(t, err, http.StatusUnprocessableEntity, apierror.CodeValidationFailed)
}
