package quay

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIError struct {
	Detail       string `json:"detail"`
	ErrorMessage string `json:"error_message"`
	ErrorType    string `json:"error_type"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	Status       int    `json:"status"`
}

func (e *APIError) Error() string {
	message := e.Detail
	if message == "" {
		message = e.ErrorMessage
	}
	if message == "" {
		message = e.Title
	}
	if message == "" {
		message = "quay api request failed"
	}

	if e.Status > 0 {
		return fmt.Sprintf("quay api returned status %d: %s", e.Status, message)
	}
	return message
}

func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("quay api returned status %d and error body could not be read: %w", resp.StatusCode, err)
	}

	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && (apiErr.Detail != "" || apiErr.ErrorMessage != "" || apiErr.Title != "") {
		if apiErr.Status == 0 {
			apiErr.Status = resp.StatusCode
		}
		return &apiErr
	}

	trimmedBody := strings.TrimSpace(string(body))
	if trimmedBody == "" {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, trimmedBody)
}
