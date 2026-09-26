package quay

import (
	"encoding/json"
	"errors"
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

		if baseErr := errorForStatus(apiErr.Status); baseErr != nil {
			return fmt.Errorf("%w: %w", baseErr, &apiErr)
		}

		return &apiErr
	}

	trimmedBody := strings.TrimSpace(string(body))

	if baseErr := errorForStatus(resp.StatusCode); baseErr != nil {
		if trimmedBody != "" {
			return fmt.Errorf("%w: %s", baseErr, trimmedBody)
		}

		return baseErr
	}

	if trimmedBody == "" {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, trimmedBody)
}

func errorForStatus(status int) error {
	switch status {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusConflict:
		return ErrConflict
	default:
		return nil
	}
}

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
	ErrConflict     = errors.New("conflict")
)
