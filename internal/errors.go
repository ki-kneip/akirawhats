package internal

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAlreadyExists      = errors.New("already exists")
)
