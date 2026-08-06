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
	case "health":
		return OkResult(json.RawMessage(`{"status":"ok"}`)), nil
	case "list_tools":
		names := make([]string, 0, len(ListTools()))
		for _, t := range ListTools() {
			names = append(names, t.Name)
		}
		out, _ := json.Marshal(names)
		return OkResult(out), nil
	case "fetch_ohlcv":
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
		return ToolResult{}, fmt.Errorf("%w: invalid arguments: %v", core.ErrInvalidCommand, err)
	}
	ticker, err := core.NewTicker(payload.Ticker)
	if err != nil {
		return ToolResult{}, err
	}
	start, err := time.Parse("2006-01-02", payload.Start)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid start date", core.ErrInvalidCommand)
	}
	end, err := time.Parse("2006-01-02", payload.End)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid end date", core.ErrInvalidCommand)
	}
	rng, err := core.NewDateRange(start, end)
	if err != nil {
		return ToolResult{}, err
	}
	interval := core.Daily
	switch payload.Interval {
	case "weekly":
		interval = core.Weekly
	case "monthly":
		interval = core.Monthly
	}
	series, err := d.marketData.Fetch(ctx, ticker, interval, rng, payload.Provider)
	if err != nil {
		return ErrResult(err.Error()), nil
	}
	out, _ := json.Marshal(series)
	return OkResult(out), nil
}
