package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Error struct {
	StatusCode int
	Code       string
	RequestID  string
	Retryable  bool
	RetryAfter time.Duration
	Message    string
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code == "" {
		return fmt.Sprintf("patchnote api error: status=%d request_id=%s retryable=%t", e.StatusCode, e.RequestID, e.Retryable)
	}
	return fmt.Sprintf("patchnote api error: status=%d code=%s request_id=%s retryable=%t", e.StatusCode, e.Code, e.RequestID, e.Retryable)
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

func parseAPIError(resp *http.Response) *Error {
	apiErr := &Error{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-ID"),
		RetryAfter: parseRetryAfter(resp.Header.Get("Retry-After")),
	}

	defer resp.Body.Close()
	var envelope errorEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return apiErr
	}

	apiErr.Code = envelope.Error.Code
	apiErr.Message = envelope.Error.Message
	apiErr.RequestID = envelope.Error.RequestID
	apiErr.Retryable = envelope.Error.Retryable
	return apiErr
}

func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
