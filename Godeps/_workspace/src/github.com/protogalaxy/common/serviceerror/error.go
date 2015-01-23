package serviceerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func Decode(body io.Reader) error {
	var errorResponse struct {
		Message string `json:"message"`
		Err     string `json:"error"`
	}
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&errorResponse); err != nil {
		return fmt.Errorf("Error decoding unexpected response: %s", err)
	}
	if errorResponse.Message == "" {
		return errors.New("Missing messages field")
	}
	return ErrorResponse{
		Message: errorResponse.Message,
		Err:     errors.New(errorResponse.Err),
	}
}
