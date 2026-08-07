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

func Load(paths ...string) (*Config, error) {
	_ = godotenv.Load(".env")
	k := koanf.New(".")
	for _, p := range paths {
		_ = k.Load(file.Provider(p), toml.Parser())
	}
	_ = k.Load(env.Provider("SQT_", ".", func(s string) string {
		return strings.ToLower(strings.TrimPrefix(s, "SQT_"))
	}), nil)
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50051
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}
