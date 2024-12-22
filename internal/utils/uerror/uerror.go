package uerror

import (
	"errors"
	"net/http"
)

type APIError struct {
	ConsumerMessage string
	ActualError     error
	StatusCode      int
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.ActualError.Error()
}

func NewNotFound(consumerMessage string, err error) error {
	if err == nil {
		err = errors.New(consumerMessage)
	}
	return &APIError{
		ConsumerMessage: consumerMessage,
		ActualError:     err,
		StatusCode:      http.StatusNotFound,
	}
}

func NewBadRequest(consumerMessage string, err error) error {
	if err == nil {
		err = errors.New(consumerMessage)
	}
	return &APIError{
		ConsumerMessage: consumerMessage,
		ActualError:     err,
		StatusCode:      http.StatusBadRequest,
	}
}

func SanitizeError(err error) (string, int) {
	if apiErr, isAPIError := isAPIError(err); isAPIError {
		return apiErr.ConsumerMessage, apiErr.StatusCode
	}
	return "unknown error occurred", http.StatusInternalServerError
}

func isAPIError(err error) (*APIError, bool) {
	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError, true
	}
	return nil, false
}
