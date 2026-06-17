package core_errors

// sentinel ошибки

import "errors"

var (
	// клиентские
	ErrNotFound        = errors.New("not found")
	ErrInvalidArgument = errors.New("invalid argument")

	// внутренние
	ErrCodeExists      = errors.New("code already exists")
	ErrURLExists       = errors.New("url already exists")
)
