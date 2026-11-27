package client

import (
	"fmt"
	"net/http"
)

type APIError struct {
	StatusCode int
	Type       string
	Message    string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s (url: %s)", e.StatusCode, e.Message, e.URL)
}

func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func IsForbidden(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusForbidden
	}
	return false
}

func IsAlreadyExists(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusBadRequest && apiErr.Type == "resource_already_exists_exception"
	}
	return false
}
