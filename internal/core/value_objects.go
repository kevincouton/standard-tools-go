package core

import (
	"fmt"
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

// ParseBarInterval parses a bar interval string. Empty string defaults to Daily.
// Accepted values are "daily", "weekly", and "monthly" (case-insensitive).
func ParseBarInterval(s string) (BarInterval, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "daily":
		return Daily, nil
	case "weekly":
		return Weekly, nil
	case "monthly":
		return Monthly, nil
	default:
		return Daily, fmt.Errorf("%w: invalid interval %q (want daily, weekly, or monthly)", ErrInvalidCommand, s)
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
