package request

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/pagination"
)

type signupPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Age      int    `json:"age"`
}

func newTestContext(method, target, body string) *gin.Context {
	gin.SetMode(gin.TestMode)
	RegisterJSONFieldNames()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
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

func TestBindJSONAcceptsValidPayload(t *testing.T) {
	var payload signupPayload
	context := newTestContext(http.MethodPost, "/", `{"email":"pilot@nebula.test","username":"pilot_1"}`)

	if err := BindJSON(context, &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Username != "pilot_1" {
		t.Fatalf("got username %q", payload.Username)
	}
}

func TestBindJSONReportsFieldErrorsByJSONName(t *testing.T) {
	var payload signupPayload
	context := newTestContext(http.MethodPost, "/", `{"email":"not-an-email","username":"ab"}`)

	apiError := requireAPIError(t, BindJSON(context, &payload), http.StatusUnprocessableEntity, apierror.CodeValidationFailed)

	fieldErrors, isFieldMap := apiError.Details.(map[string]string)
	if !isFieldMap {
		t.Fatalf("expected field error map, got %T", apiError.Details)
	}
	if fieldErrors["email"] != "must be a valid email address" || fieldErrors["username"] != "must be at least 3" {
		t.Fatalf("unexpected field errors %v", fieldErrors)
	}
}

func TestBindJSONRejectsBrokenBodies(t *testing.T) {
	brokenBodies := map[string]struct {
		body           string
		expectedStatus int
	}{
		"empty body":   {body: "", expectedStatus: http.StatusBadRequest},
		"invalid json": {body: `{"email":`, expectedStatus: http.StatusBadRequest},
		"wrong type":   {body: `{"email":"pilot@nebula.test","username":"pilot","age":"old"}`, expectedStatus: http.StatusUnprocessableEntity},
	}

	for caseName, testCase := range brokenBodies {
		t.Run(caseName, func(t *testing.T) {
			var payload signupPayload
			context := newTestContext(http.MethodPost, "/", testCase.body)
			requireAPIError(t, BindJSON(context, &payload), testCase.expectedStatus, apierror.CodeValidationFailed)
		})
	}
}

func TestPaginationFromQueryUsesDefaultLimit(t *testing.T) {
	paginationRequest, err := PaginationFromQuery(newTestContext(http.MethodGet, "/orders", ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paginationRequest.Limit != pagination.DefaultLimit || paginationRequest.HasCursor() {
		t.Fatalf("unexpected pagination request %+v", paginationRequest)
	}
}

func TestPaginationFromQueryReadsCursorAndLimit(t *testing.T) {
	paginationRequest, err := PaginationFromQuery(newTestContext(http.MethodGet, "/orders?cursor=abc&limit=50", ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paginationRequest.Limit != 50 || paginationRequest.Cursor != "abc" {
		t.Fatalf("unexpected pagination request %+v", paginationRequest)
	}
}

func TestPaginationFromQueryRejectsOutOfRangeLimits(t *testing.T) {
	for _, rawLimit := range []string{"0", "-1", "101", "ten"} {
		_, err := PaginationFromQuery(newTestContext(http.MethodGet, "/orders?limit="+rawLimit, ""))
		requireAPIError(t, err, http.StatusUnprocessableEntity, apierror.CodeValidationFailed)
	}
}

func TestDecodeCursorPositionRejectsInvalidCursor(t *testing.T) {
	_, err := DecodeCursorPosition[map[string]int](pagination.Request{Cursor: "%%%"})
	requireAPIError(t, err, http.StatusUnprocessableEntity, apierror.CodeValidationFailed)
}
