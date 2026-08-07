package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadDefaults assumes no .env file exists in the repository root.
// If one is added, this test should load defaults via a helper that skips .env loading.
func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, defaultHTTPPort, cfg.HTTPPort)
	assert.Equal(t, defaultGRPCPort, cfg.GRPCPort)
	assert.Equal(t, defaultLogLevel, cfg.LogLevel)
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("SQT_HTTP_PORT", "9090")
	t.Setenv("SQT_LOG_LEVEL", "debug")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.HTTPPort)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadEnvNestedKey(t *testing.T) {
	t.Setenv("SQT_POLYGON__API_KEY", "env-nested-key")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "env-nested-key", cfg.Polygon.APIKey)
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
http_port = 7070
grpc_port = 50052
log_level = "warn"
database_url = "postgres://localhost/test"
cache_dir = "/tmp/cache"
audit_dir = "/tmp/audit"

[polygon]
api_key = "file-key"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 7070, cfg.HTTPPort)
	assert.Equal(t, 50052, cfg.GRPCPort)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
	assert.Equal(t, "/tmp/cache", cfg.CacheDir)
	assert.Equal(t, "/tmp/audit", cfg.AuditDir)
	assert.Equal(t, "file-key", cfg.Polygon.APIKey)
}

func TestLoadEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(`http_port = 7070
log_level = "warn"
`), 0o600))

	t.Setenv("SQT_HTTP_PORT", "9090")
	t.Setenv("SQT_LOG_LEVEL", "debug")

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.HTTPPort)
	assert.Equal(t, "debug", cfg.LogLevel)
}

func TestLoadMissingFileReturnsError(t *testing.T) {
	path := "/nonexistent/path/config.toml"
	_, err := Load(path)
	require.Error(t, err)
	assert.ErrorContains(t, err, path)
}
