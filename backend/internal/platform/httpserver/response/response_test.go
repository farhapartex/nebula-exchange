package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

func newRecordingContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	return context, recorder
}

func decodeBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	return body
}

func TestWriteListAlwaysReturnsArrayAndPagination(t *testing.T) {
	context, recorder := newRecordingContext()

	WriteList(context, http.StatusOK, pagination.Page[string]{Info: pagination.PageInfo{Limit: 20}})

	body := decodeBody(t, recorder)
	if items, isArray := body["data"].([]any); !isArray || len(items) != 0 {
		t.Fatalf("expected empty data array, got %v", body["data"])
	}
	pageInfo := body["pagination"].(map[string]any)
	if pageInfo["next_cursor"] != nil || pageInfo["limit"] != float64(20) {
		t.Fatalf("unexpected pagination %v", pageInfo)
	}
}

func TestWriteErrorUsesAPIErrorFields(t *testing.T) {
	context, recorder := newRecordingContext()

	WriteError(context, apierror.ValidationFailed(map[string]string{"email": "is required"}))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got status %d, want 422", recorder.Code)
	}
	errorBody := decodeBody(t, recorder)["error"].(map[string]any)
	if errorBody["code"] != string(apierror.CodeValidationFailed) {
		t.Fatalf("got code %v", errorBody["code"])
	}
	if errorBody["details"].(map[string]any)["email"] != "is required" {
		t.Fatalf("got details %v", errorBody["details"])
	}
}

func TestWriteErrorHidesUnknownErrors(t *testing.T) {
	context, recorder := newRecordingContext()

	WriteError(context, errors.New("pq: connection refused to 10.0.0.5"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want 500", recorder.Code)
	}
	errorBody := decodeBody(t, recorder)["error"].(map[string]any)
	if errorBody["code"] != string(apierror.CodeInternalError) || errorBody["message"] != "Something went wrong" {
		t.Fatalf("internal details leaked: %v", errorBody)
	}
	if len(context.Errors) != 1 {
		t.Fatalf("expected the original error to be kept for logging")
	}
}
