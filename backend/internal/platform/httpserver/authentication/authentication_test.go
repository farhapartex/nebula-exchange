package authentication

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fixedVerifier struct {
	validToken string
	userID     uuid.UUID
}

func (verifier fixedVerifier) Verify(accessToken string) (uuid.UUID, error) {
	if accessToken != verifier.validToken {
		return uuid.Nil, errors.New("invalid token")
	}
	return verifier.userID, nil
}

func TestRequireUserAcceptsOnlyAValidBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	knownUserID := uuid.New()
	router := gin.New()
	router.Use(IdentifyUser(fixedVerifier{validToken: "good-token", userID: knownUserID}))
	router.GET("/me", RequireUser(), func(context *gin.Context) {
		userID, _ := UserIDFrom(context)
		context.String(http.StatusOK, userID.String())
	})

	expectations := map[string]struct {
		authorization  string
		expectedStatus int
	}{
		"no header":    {authorization: "", expectedStatus: http.StatusUnauthorized},
		"wrong token":  {authorization: "Bearer bad-token", expectedStatus: http.StatusUnauthorized},
		"not a bearer": {authorization: "Basic good-token", expectedStatus: http.StatusUnauthorized},
		"valid bearer": {authorization: "Bearer good-token", expectedStatus: http.StatusOK},
	}
	for caseName, expectation := range expectations {
		t.Run(caseName, func(t *testing.T) {
			testRequest := httptest.NewRequest(http.MethodGet, "/me", nil)
			if expectation.authorization != "" {
				testRequest.Header.Set("Authorization", expectation.authorization)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, testRequest)
			if recorder.Code != expectation.expectedStatus {
				t.Fatalf("got %d, want %d", recorder.Code, expectation.expectedStatus)
			}
			if expectation.expectedStatus == http.StatusOK && recorder.Body.String() != knownUserID.String() {
				t.Fatalf("got user %q, want %q", recorder.Body.String(), knownUserID)
			}
		})
	}
}
