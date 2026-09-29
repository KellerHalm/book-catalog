package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, message)
}

func Internal(err error) *Error {
	return Wrap(http.StatusInternalServerError, "internal server error", err)
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
