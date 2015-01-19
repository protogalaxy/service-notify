package serviceerror

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Err        error  `json:"error,omitempty"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Message, e.Err)
}

func InternalServerError(msg string, err error) ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Message:    msg,
		Err:        err,
	}
}

func BadRequest(msg string, err error) ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Message:    msg,
		Err:        err,
	}
}
