package errors

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

func TestClassifyAuthErrorInvalidGrant(t *testing.T) {
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"invalid_grant","error_description":"expired"}`),
	}

	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatalf("expected APIError")
	}
	if apiErr.Code != CodeAuthFailure {
		t.Fatalf("expected auth failure, got %s", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Fatalf("expected hint to be set")
	}
}

func TestClassifyAuthErrorGoogleAPI(t *testing.T) {
	gapiErr := &googleapi.Error{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
		Header:  http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
	}

	apiErr := ClassifyAuthError(gapiErr)
	if apiErr == nil {
		t.Fatalf("expected APIError")
	}
	if apiErr.Code != CodeAuthFailure {
		t.Fatalf("expected auth failure, got %s", apiErr.Code)
	}
}
