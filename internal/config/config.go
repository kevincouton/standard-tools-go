package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	envPrefix     = "SQT_"
	envDelimiter  = "__"
	keySeparator  = "."
	defaultDotEnv = ".env"

	defaultHTTPPort = 8080
	defaultGRPCPort = 50051
	defaultLogLevel = "info"
)

type Config struct {
	HTTPPort    int           `koanf:"http_port"`
	GRPCPort    int           `koanf:"grpc_port"`
	LogLevel    string        `koanf:"log_level"`
	DatabaseURL string        `koanf:"database_url"`
	CacheDir    string        `koanf:"cache_dir"`
	AuditDir    string        `koanf:"audit_dir"`
	Polygon     PolygonConfig `koanf:"polygon"`
}

type PolygonConfig struct {
	APIKey string `koanf:"api_key"`
}

// Load reads configuration from optional TOML files and environment variables.
// Precedence (lowest to highest): defaults < TOML files < env vars.
// Env vars come from both an optional .env file and the process environment;
// both are treated with equal precedence, so SQT_HTTP_PORT and a .env HTTP_PORT
// value both override a TOML file. Nested keys in env vars are expressed with
// double underscores, e.g. SQT_POLYGON__API_KEY maps to polygon.api_key.
func Load(paths ...string) (*Config, error) {
	// .env is optional; ignore missing-file errors so local development stays simple.
	_ = godotenv.Load(defaultDotEnv)

	k := koanf.New(keySeparator)
	for _, p := range paths {
		if err := k.Load(file.Provider(p), toml.Parser()); err != nil {
			return nil, fmt.Errorf("load config file %q: %w", p, err)
		}
	}
	_ = k.Load(env.Provider(envPrefix, keySeparator, envKeyTransform), nil)

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	setDefaults(&cfg)
	return &cfg, nil
}

func envKeyTransform(s string) string {
	key := strings.ToLower(strings.TrimPrefix(s, envPrefix))
	return strings.ReplaceAll(key, envDelimiter, keySeparator)
}

func setDefaults(cfg *Config) {
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = defaultHTTPPort
	}
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = defaultGRPCPort
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
}
