package core

import "errors"

var (
	ErrInvalidCommand       = errors.New("invalid command")
	ErrNotFound             = errors.New("not found")
	ErrDataQuality          = errors.New("data quality")
	ErrProviderNotAvailable = errors.New("provider not available")
	ErrInternal             = errors.New("internal error")
)
