package client

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// HTTPError is returned when the server responds with a non-2xx HTTP status.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("api request failed with status %d: %s", e.StatusCode, e.Body)
}

// APIError is returned when HTTP succeeded but the OpenList envelope code is not 200.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("api error: %s", e.Message)
}

// IsNotFound reports whether err represents a missing object.
func IsNotFound(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == http.StatusNotFound {
			return true
		}
		msg := strings.ToLower(apiErr.Message)
		return strings.Contains(msg, "not found") ||
			strings.Contains(msg, "doesn't exist") ||
			strings.Contains(msg, "does not exist")
	}
	return false
}
