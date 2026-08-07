package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// DateFormat is the canonical date layout used across the application.
const DateFormat = "2006-01-02"

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

// TickerInfo mirrors the Python Standard-Tools TickerInfo model.
type TickerInfo struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Sector    string `json:"sector"`
	Industry  string `json:"industry"`
	Employees int64  `json:"employees"`
	City      string `json:"city"`
	Country   string `json:"country"`
	Website   string `json:"website"`
}

// FinancialRatios mirrors the Python FinancialRatios model.
type FinancialRatios struct {
	Symbol        string          `json:"symbol"`
	ForwardPE     decimal.Decimal `json:"forward_pe"`
	TrailingPE    decimal.Decimal `json:"trailing_pe"`
	PriceToBook   decimal.Decimal `json:"price_to_book"`
	DebtToEquity  decimal.Decimal `json:"debt_to_equity"`
	ROE           decimal.Decimal `json:"roe"`
	ProfitMargins decimal.Decimal `json:"profit_margins"`
	DividendYield decimal.Decimal `json:"dividend_yield"`
	MarketCap     int64           `json:"market_cap"`
}

// DataSetMetadata describes the provenance of a fetched dataset.
type DataSetMetadata struct {
	Provider         string    `json:"provider"`
	Adjusted         bool      `json:"adjusted"`
	SurvivorshipFree bool      `json:"survivorship_free"`
	PointInTime      bool      `json:"point_in_time"`
	Frequency        string    `json:"frequency"`
	Timezone         string    `json:"timezone"`
	RetrievedAt      time.Time `json:"retrieved_at"`
}
