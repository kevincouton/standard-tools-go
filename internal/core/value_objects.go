package core

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Ticker struct {
	Symbol   string
	Exchange string
}

func NewTicker(symbol string) (Ticker, error) {
	s := strings.TrimSpace(symbol)
	if s == "" {
		return Ticker{}, ErrInvalidTicker
	}
	return Ticker{Symbol: s}, nil
}

type BarInterval int

const (
	Daily BarInterval = iota
	Weekly
	Monthly
)

func (b BarInterval) String() string {
	switch b {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	default:
		return "daily"
	}
}

type DateRange struct {
	Start time.Time
	End   time.Time
}

func NewDateRange(start, end time.Time) (DateRange, error) {
	if end.Before(start) {
		return DateRange{}, ErrInvalidDateRange
	}
	return DateRange{Start: start, End: end}, nil
}

type OHLCV struct {
	Date   time.Time
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	Volume int64
}
