package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTicker(t *testing.T) {
	ticker, err := NewTicker("AAPL")
	assert.NoError(t, err)
	assert.Equal(t, "AAPL", ticker.Symbol)
}

func TestNewTickerRejectsEmpty(t *testing.T) {
	_, err := NewTicker("")
	assert.ErrorIs(t, err, ErrInvalidCommand)
}
