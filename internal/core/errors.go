package core

import "errors"

var (
	ErrInvalidTicker        = errors.New("invalid ticker")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrInvalidCommand       = errors.New("invalid command")
	ErrNotFound             = errors.New("not found")
	ErrDataQuality          = errors.New("data quality")
	ErrProviderNotAvailable = errors.New("provider not available")
	ErrInternal             = errors.New("internal error")
)
