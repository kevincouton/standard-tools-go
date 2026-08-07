package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/kevincouton/standard-tools-go/internal/audit"
	"github.com/kevincouton/standard-tools-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyCmdEmptyChain(t *testing.T) {
	loadConfig = func(...string) (*config.Config, error) {
		return &config.Config{}, nil
	}
	newStorage = func(context.Context, *config.Config) (audit.Storage, error) {
		return audit.NewMemoryStorage(), nil
	}
	defer func() { loadConfig = config.Load; newStorage = defaultNewStorage }()

	cmd := newVerifyCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "audit chain OK")
}

func TestReportCmdExistingRecord(t *testing.T) {
	loadConfig = func(...string) (*config.Config, error) {
		return &config.Config{}, nil
	}
	store := audit.NewMemoryStorage()
	newStorage = func(context.Context, *config.Config) (audit.Storage, error) {
		return store, nil
	}
	defer func() { loadConfig = config.Load; newStorage = defaultNewStorage }()

	ctx := context.Background()
	require.NoError(t, audit.NewWriter(store).Write(ctx, audit.DecisionRecord{
		RequestID:      "req-123",
		ToolName:       "test-tool",
		Input:          map[string]any{"key": "value"},
		Output:         map[string]any{"result": "ok"},
		Status:         "success",
		GitCommitSHA:   "abc123",
		PackageVersion: "v0.1.0",
		RandomSeed:     42,
	}))

	cmd := newReportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"req-123"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "req-123")
	assert.Contains(t, out.String(), "test-tool")
}
