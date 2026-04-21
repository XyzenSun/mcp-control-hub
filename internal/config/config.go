package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	MCP      MCPConfig      `mapstructure:"mcp"`
	Security SecurityConfig `mapstructure:"security"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type MCPConfig struct {
	SessionTTL            time.Duration `mapstructure:"session_ttl"`
	HealthCheckInterval   time.Duration `mapstructure:"health_check_interval"`
	ReconnectBackoff      time.Duration `mapstructure:"reconnect_backoff"`
	MaxReconnectAttempts  int           `mapstructure:"max_reconnect_attempts"`
}

type SecurityConfig struct {
	APIKeyLength int `mapstructure:"api_key_length"`
	RateLimit    int `mapstructure:"rate_limit"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "gateway.db")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	v.SetDefault("mcp.session_ttl", "1h")
	v.SetDefault("mcp.health_check_interval", "30s")
	v.SetDefault("mcp.reconnect_backoff", "5s")
	v.SetDefault("mcp.max_reconnect_attempts", 10)
	v.SetDefault("security.api_key_length", 32)
	v.SetDefault("security.rate_limit", 100)

	// Environment variables
	v.SetEnvPrefix("GATEWAY")
	v.AutomaticEnv()

	// Config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
