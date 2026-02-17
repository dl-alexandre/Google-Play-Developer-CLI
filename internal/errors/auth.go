package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
		skewed := addClockSkew(details, retrieveErr.Response.Header)
		switch resp.Error {
		case "invalid_grant":
			apiErr := NewAPIError(CodeAuthFailure, "refresh token expired or revoked").
				WithHint("Re-authenticate or rotate credentials").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
			if skewed {
				apiErr = apiErr.WithHint(appendHint(apiErr.Hint, "Sync system clock and retry"))
			}
			return apiErr
		case "invalid_client", "unauthorized_client":
			apiErr := NewAPIError(CodeAuthFailure, "client not authorized to refresh tokens").
				WithHint("Verify the client credentials or service account key").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
			if skewed {
				apiErr = apiErr.WithHint(appendHint(apiErr.Hint, "Sync system clock and retry"))
			}
			return apiErr
		case "access_denied":
			apiErr := NewAPIError(CodeAuthFailure, "access denied during token refresh").
				WithHint("Re-authenticate and confirm access is granted").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
			if skewed {
				apiErr = apiErr.WithHint(appendHint(apiErr.Hint, "Sync system clock and retry"))
			}
			return apiErr
		default:
			apiErr := NewAPIError(CodeAuthFailure, "authentication refresh failed").
				WithHint("Re-authenticate or verify credentials").
				WithDetails(details).
				WithHTTPStatus(retrieveErr.Response.StatusCode)
			if skewed {
				apiErr = apiErr.WithHint(appendHint(apiErr.Hint, "Sync system clock and retry"))
			}
			return apiErr
		}
	}

	var gapiErr *googleapi.Error
	if errors.As(err, &gapiErr) {
		details := map[string]interface{}{
			"httpStatus": gapiErr.Code,
		}
		skewed := addClockSkew(details, gapiErr.Header)
		code := FromHTTPStatus(gapiErr.Code)
		apiErr := NewAPIError(code, gapiErr.Message).WithDetails(details).WithHTTPStatus(gapiErr.Code)
		if gapiErr.Code == http.StatusUnauthorized {
			apiErr = apiErr.WithHint("Re-authenticate and retry the command")
		}
		if gapiErr.Code == http.StatusForbidden {
			apiErr = apiErr.WithHint("Verify permissions and service account access")
		}
		if skewed {
			apiErr = apiErr.WithHint(appendHint(apiErr.Hint, "Sync system clock and retry"))
		}
		return apiErr
	}

	return NewAPIError(CodeAuthFailure, err.Error())
}

func addClockSkew(details map[string]interface{}, header http.Header) bool {
	if header == nil {
		return false
	}
	dateHeader := header.Get("Date")
	if dateHeader == "" {
		return false
	}
	remoteTime, err := http.ParseTime(dateHeader)
	if err != nil {
		return false
	}
	skew := time.Since(remoteTime)
	if skew < 0 {
		skew = -skew
	}
	if skew > 5*time.Minute {
		details["clockSkewSeconds"] = int64(skew.Seconds())
		return true
	}
	return false
}

func appendHint(current, extra string) string {
	if current == "" {
		return extra
	}
	if strings.Contains(current, extra) {
		return current
	}
	return current + "; " + extra
}
