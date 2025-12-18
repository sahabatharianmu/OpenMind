package response

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternalServerError = errors.New("internal server error")
	ErrConflict            = errors.New("resource already exists")
)

func NewNotFound(message string) *AppError {
	return NewAppError(consts.StatusNotFound, message, ErrNotFound)
}

func NewUnauthorized(message string) *AppError {
	return NewAppError(consts.StatusUnauthorized, message, ErrUnauthorized)
}

func NewForbidden(message string) *AppError {
	return NewAppError(consts.StatusForbidden, message, ErrForbidden)
}

func NewBadRequest(message string) *AppError {
	return NewAppError(consts.StatusBadRequest, message, ErrInvalidInput)
}

func NewInternalServerError(message string) *AppError {
	return NewAppError(consts.StatusInternalServerError, message, ErrInternalServerError)
}

func NewConflict(message string) *AppError {
	return NewAppError(consts.StatusConflict, message, ErrConflict)
}
