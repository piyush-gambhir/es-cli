package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response wraps an http.Response with convenience methods.
type Response struct {
	HTTPResponse *http.Response
	RequestURL   string
}

// JSON reads the response body and unmarshals it into v.
// It also checks for error status codes.
func (r *Response) JSON(v interface{}) error {
	defer r.HTTPResponse.Body.Close()

	body, err := io.ReadAll(r.HTTPResponse.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if r.HTTPResponse.StatusCode >= 400 {
		msg := extractErrorMessage(body)
		return &APIError{
			StatusCode: r.HTTPResponse.StatusCode,
			Message:    msg,
			URL:        r.RequestURL,
		}
	}

	if v != nil && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

// Error checks if the response indicates an error and returns an APIError.
// It always closes the response body to prevent leaks.
func (r *Response) Error() error {
	defer r.HTTPResponse.Body.Close()

	if r.HTTPResponse.StatusCode >= 400 {
		body, err := io.ReadAll(r.HTTPResponse.Body)
		if err != nil {
			return &APIError{
				StatusCode: r.HTTPResponse.StatusCode,
				Message:    fmt.Sprintf("failed to read error body: %v", err),
				URL:        r.RequestURL,
			}
		}
		msg := extractErrorMessage(body)
		return &APIError{
			StatusCode: r.HTTPResponse.StatusCode,
			Message:    msg,
			URL:        r.RequestURL,
		}
	}
	return nil
}

// StatusCode returns the HTTP status code.
func (r *Response) StatusCode() int {
	return r.HTTPResponse.StatusCode
}

// RawBody reads and returns the raw response body.
func (r *Response) RawBody() ([]byte, error) {
	defer r.HTTPResponse.Body.Close()
	return io.ReadAll(r.HTTPResponse.Body)
}

// extractErrorMessage attempts to extract a meaningful error message from
// an Elasticsearch error response body.
func extractErrorMessage(body []byte) string {
	// Elasticsearch errors typically have: {"error":{"type":"...","reason":"..."},...}
	var esErr struct {
		Error struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"error"`
		Status int `json:"status"`
	}
	if json.Unmarshal(body, &esErr) == nil && esErr.Error.Reason != "" {
		if esErr.Error.Type != "" {
			return fmt.Sprintf("[%s] %s", esErr.Error.Type, esErr.Error.Reason)
		}
		return esErr.Error.Reason
	}

	// Try simple string error: {"error":"..."}
	var simpleErr struct {
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	if json.Unmarshal(body, &simpleErr) == nil {
		if simpleErr.Error != "" {
			return simpleErr.Error
		}
		if simpleErr.Reason != "" {
			return simpleErr.Reason
		}
	}

	// Try message field: {"message":"..."}
	var msgErr struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &msgErr) == nil && msgErr.Message != "" {
		return msgErr.Message
	}

	// Fall back to raw body.
	msg := string(body)
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	return msg
}
