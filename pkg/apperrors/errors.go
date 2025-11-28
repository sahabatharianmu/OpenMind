package apperrors

import "errors"

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

func New(code int, message string, err error) *AppError {
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
	return New(404, message, ErrNotFound)
}

func NewUnauthorized(message string) *AppError {
	return New(401, message, ErrUnauthorized)
}

func NewForbidden(message string) *AppError {
	return New(403, message, ErrForbidden)
}

func NewBadRequest(message string) *AppError {
	return New(400, message, ErrInvalidInput)
}

func NewInternalServerError(message string) *AppError {
	return New(500, message, ErrInternalServerError)
}

func NewConflict(message string) *AppError {
	return New(409, message, ErrConflict)
}
