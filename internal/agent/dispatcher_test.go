package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kevincouton/standard-tools-go/internal/marketdata"
	"github.com/stretchr/testify/assert"
)

func TestDispatchHealth(t *testing.T) {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	svc.Register(&marketdata.SyntheticProvider{})
	d := NewDispatcher(svc)
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "health", Arguments: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)
	assert.JSONEq(t, `{"status":"ok"}`, string(res.Output))
}

func TestDispatchUnknownTool(t *testing.T) {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	d := NewDispatcher(svc)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "nope"})
	assert.Error(t, err)
}
