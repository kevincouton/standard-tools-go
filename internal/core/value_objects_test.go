package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTicker(t *testing.T) {
	ticker, err := NewTicker("AAPL")
	assert.NoError(t, err)
	assert.Equal(t, "AAPL", ticker.Symbol)
}

func TestNewTickerTrimsWhitespace(t *testing.T) {
	ticker, err := NewTicker("  AAPL  ")
	assert.NoError(t, err)
	assert.Equal(t, "AAPL", ticker.Symbol)
}

func TestNewTickerRejectsEmpty(t *testing.T) {
	_, err := NewTicker("")
	assert.ErrorIs(t, err, ErrInvalidTicker)
}

func TestNewDateRange(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, err := NewDateRange(start, end)
	assert.NoError(t, err)
	assert.Equal(t, start, rng.Start)
	assert.Equal(t, end, rng.End)
}

func TestNewDateRangeRejectsInverted(t *testing.T) {
	start := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := NewDateRange(start, end)
	assert.ErrorIs(t, err, ErrInvalidDateRange)
}

func TestBarIntervalString(t *testing.T) {
	assert.Equal(t, "daily", Daily.String())
	assert.Equal(t, "weekly", Weekly.String())
	assert.Equal(t, "monthly", Monthly.String())
	assert.Equal(t, "daily", BarInterval(99).String())
}

func TestParseBarInterval(t *testing.T) {
	cases := []struct {
		input string
		want  BarInterval
	}{
		{"", Daily},
		{"daily", Daily},
		{"Daily", Daily},
		{"DAILY", Daily},
		{" weekly ", Weekly},
		{"monthly", Monthly},
	}
	for _, tc := range cases {
		got, err := ParseBarInterval(tc.input)
		assert.NoError(t, err, "input %q", tc.input)
		assert.Equal(t, tc.want, got, "input %q", tc.input)
	}
}

func TestParseBarIntervalRejectsUnknown(t *testing.T) {
	_, err := ParseBarInterval("hourly")
	assert.ErrorIs(t, err, ErrInvalidCommand)
}
