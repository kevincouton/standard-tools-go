package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
	"github.com/stretchr/testify/assert"
)

func newTestDispatcher() *Dispatcher {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	svc.Register(&marketdata.SyntheticProvider{})
	return NewDispatcher(svc)
}

func TestDispatchHealth(t *testing.T) {
	d := newTestDispatcher()
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "health", Arguments: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)
	assert.JSONEq(t, `{"status":"ok"}`, string(res.Output))
}

func TestDispatchUnknownTool(t *testing.T) {
	d := newTestDispatcher()
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "nope"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrInvalidCommand))
}

func TestDispatchListTools(t *testing.T) {
	d := newTestDispatcher()
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "list_tools", Arguments: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)

	var names []string
	assert.NoError(t, json.Unmarshal(res.Output, &names))
	assert.Equal(t, []string{"health", "list_tools", "fetch_ohlcv"}, names)
}

func TestDispatchFetchOhlcv(t *testing.T) {
	d := newTestDispatcher()
	args := json.RawMessage(`{"ticker":"AAPL","start":"2024-01-01","end":"2024-01-05","interval":"daily"}`)
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)

	var series []core.OHLCV
	assert.NoError(t, json.Unmarshal(res.Output, &series))
	assert.Len(t, series, 5)
}

func TestDispatchFetchOhlcvDefaultsToDaily(t *testing.T) {
	d := newTestDispatcher()
	args := json.RawMessage(`{"ticker":"AAPL","start":"2024-01-01","end":"2024-01-03"}`)
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)

	var series []core.OHLCV
	assert.NoError(t, json.Unmarshal(res.Output, &series))
	assert.Len(t, series, 3)
}

func TestDispatchFetchOhlcvInvalidTicker(t *testing.T) {
	d := newTestDispatcher()
	args := json.RawMessage(`{"ticker":"  ","start":"2024-01-01","end":"2024-01-05"}`)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrInvalidTicker))
}

func TestDispatchFetchOhlcvInvalidDate(t *testing.T) {
	d := newTestDispatcher()
	args := json.RawMessage(`{"ticker":"AAPL","start":"not-a-date","end":"2024-01-05"}`)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrInvalidCommand))
}

func TestDispatchFetchOhlcvInvalidInterval(t *testing.T) {
	d := newTestDispatcher()
	args := json.RawMessage(`{"ticker":"AAPL","start":"2024-01-01","end":"2024-01-05","interval":"hourly"}`)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrInvalidCommand))
}

func TestDispatchFetchOhlcvToolExecutionError(t *testing.T) {
	svc := marketdata.NewService("unregistered", marketdata.NewInMemoryCache())
	d := NewDispatcher(svc)
	args := json.RawMessage(`{"ticker":"AAPL","start":"2024-01-01","end":"2024-01-05"}`)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "fetch_ohlcv", Arguments: args})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, core.ErrProviderNotAvailable))
}
