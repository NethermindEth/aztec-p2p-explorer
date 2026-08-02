package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	Status  int `json:"status"`
	Message any `json:"message"`
	err     error
}

func (e APIError) Error() string {
	var msg string
	switch v := e.Message.(type) {
	case map[string]string:
		j, _ := json.Marshal(v)
		msg = string(j)
	default:
		msg = fmt.Sprint(v)
	}

	return fmt.Sprintf("api error: %d (%s) error=%v", e.Status, msg, e.err)
}

func NewAPIError(status int, err error) APIError {
	return APIError{
		Status:  status,
		Message: err.Error(),
		err:     err,
	}
}

func InvalidJSON() APIError {
	return APIError{
		Status:  http.StatusBadRequest,
		Message: "invalid json",
	}
}

func InvalidParameters(err error) APIError {
	return APIError{
		Status:  http.StatusBadRequest,
		Message: "invalid parameters",
		err:     err,
	}
}

func ValidationFailed(errors map[string]string) APIError {
	return APIError{
		Status:  http.StatusUnprocessableEntity,
		Message: errors,
	}
}

func RecordNotFound(id uint64) APIError {
	err := fmt.Errorf("record with id %d not found", id)

	return APIError{
		Status:  http.StatusNotFound,
		Message: err.Error(),
	}
}

func RecordWithStringIDNotFound(id string) APIError {
	err := fmt.Errorf("record with id %s not found", id)

	return APIError{
		Status:  http.StatusNotFound,
		Message: err.Error(),
	}
}

// ServerError error is a generic error that should not be returned to the client
// but that we absolutely want to log
func ServerError(err error) APIError {
	return APIError{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
		err:     err,
	}
}
