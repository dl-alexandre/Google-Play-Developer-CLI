package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

func ClassifyAuthError(err error) *APIError {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}

	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		details := map[string]interface{}{
			"httpStatus": retrieveErr.Response.StatusCode,
		}
		resp := oauthErrorResponse{}
		if retrieveErr.Body != nil {
			_ = json.Unmarshal(retrieveErr.Body, &resp)
		}
		if resp.Error != "" {
			details["oauthError"] = resp.Error
		}
		if resp.ErrorDescription != "" {
			details["oauthErrorDescription"] = resp.ErrorDescription
		}
		addClockSkew(details, retrieveErr.Response.Header)
		switch resp.Error {
		case "invalid_grant":
			return NewAPIError(CodeAuthFailure, "refresh token expired or revoked").
				WithHint("Re-authenticate or rotate credentials").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
		case "invalid_client", "unauthorized_client":
			return NewAPIError(CodeAuthFailure, "client not authorized to refresh tokens").
				WithHint("Verify the client credentials or service account key").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
		case "access_denied":
			return NewAPIError(CodeAuthFailure, "access denied during token refresh").
				WithHint("Re-authenticate and confirm access is granted").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
		default:
			return NewAPIError(CodeAuthFailure, "authentication refresh failed").
				WithHint("Re-authenticate or verify credentials").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
		}
	}

	var gapiErr *googleapi.Error
	if errors.As(err, &gapiErr) {
		details := map[string]interface{}{
			"httpStatus": gapiErr.Code,
		}
		addClockSkew(details, gapiErr.Header)
		code := FromHTTPStatus(gapiErr.Code)
		apiErr := NewAPIError(code, gapiErr.Message).WithDetails(details).WithHTTPStatus(gapiErr.Code)
		if gapiErr.Code == http.StatusUnauthorized {
			apiErr.WithHint("Re-authenticate and retry the command")
		}
		if gapiErr.Code == http.StatusForbidden {
			apiErr.WithHint("Verify permissions and service account access")
		}
		return apiErr
	}

	return NewAPIError(CodeAuthFailure, err.Error())
}

func addClockSkew(details map[string]interface{}, header http.Header) {
	if header == nil {
		return
	}
	dateHeader := header.Get("Date")
	if dateHeader == "" {
		return
	}
	remoteTime, err := http.ParseTime(dateHeader)
	if err != nil {
		return
	}
	skew := time.Since(remoteTime)
	if skew < 0 {
		skew = -skew
	}
	if skew > 5*time.Minute {
		details["clockSkewSeconds"] = int64(skew.Seconds())
	}
}
