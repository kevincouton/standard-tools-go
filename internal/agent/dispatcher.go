package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

type Dispatcher struct {
	marketData *marketdata.Service
}

func NewDispatcher(marketData *marketdata.Service) *Dispatcher {
	return &Dispatcher{marketData: marketData}
}

func (d *Dispatcher) Dispatch(ctx context.Context, call ToolCall) (ToolResult, error) {
	if _, ok := FindTool(call.Name); !ok {
		return ToolResult{}, fmt.Errorf("%w: unknown tool %s", core.ErrInvalidCommand, call.Name)
	}
	return d.dispatchKnown(ctx, call)
}

func (d *Dispatcher) dispatchKnown(ctx context.Context, call ToolCall) (ToolResult, error) {
	switch call.Name {
	case ToolHealth:
		return OkResult(json.RawMessage(`{"status":"ok"}`)), nil
	case ToolListTools:
		tools := ListTools()
		names := make([]string, 0, len(tools))
		for _, t := range tools {
			names = append(names, t.Name)
		}
		out, err := json.Marshal(names)
		if err != nil {
			return ToolResult{}, fmt.Errorf("%w: failed to marshal tool list: %w", core.ErrInternal, err)
		}
		return OkResult(out), nil
	case ToolFetchOhlcv:
		return d.fetchOhlcv(ctx, call.Arguments)
	default:
		return ToolResult{}, fmt.Errorf("%w: unknown tool %s", core.ErrInvalidCommand, call.Name)
	}
}

func (d *Dispatcher) fetchOhlcv(ctx context.Context, args json.RawMessage) (ToolResult, error) {
	var payload struct {
		Ticker   string `json:"ticker"`
		Start    string `json:"start"`
		End      string `json:"end"`
		Interval string `json:"interval"`
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid arguments: %w", core.ErrInvalidCommand, err)
	}
	ticker, err := core.NewTicker(payload.Ticker)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid ticker: %w", core.ErrInvalidCommand, err)
	}
	start, err := time.Parse(core.DateFormat, payload.Start)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid start date %q: %w", core.ErrInvalidCommand, payload.Start, err)
	}
	end, err := time.Parse(core.DateFormat, payload.End)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid end date %q: %w", core.ErrInvalidCommand, payload.End, err)
	}
	rng, err := core.NewDateRange(start, end)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid date range: %w", core.ErrInvalidCommand, err)
	}
	interval, err := core.ParseBarInterval(payload.Interval)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid interval: %w", core.ErrInvalidCommand, err)
	}
	series, err := d.marketData.Fetch(ctx, ticker, interval, rng, payload.Provider)
	if err != nil {
		return ToolResult{}, err
	}
	out, err := json.Marshal(series)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: failed to marshal OHLCV series: %w", core.ErrInternal, err)
	}
	return OkResult(out), nil
}
