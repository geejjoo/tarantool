package domain

import "errors"

var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyExists        = errors.New("key already exists")
	ErrInvalidKey       = errors.New("invalid key")
	ErrInvalidValue     = errors.New("invalid value")
	ErrDatabaseError    = errors.New("database error")
	ErrValidationError  = errors.New("validation error")
	ErrNotDeleted       = errors.New("value haven`t deleted")
	ErrKeyAlreadyExists = errors.New("key already exists")
)
