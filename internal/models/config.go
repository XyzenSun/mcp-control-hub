package models

import "time"

// SystemConfig stores system-level configuration
type SystemConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Config keys for MCP client timeouts and intervals
const (
	// ConfigKeyListToolsTimeout is the timeout for ListTools operations (in seconds)
	ConfigKeyListToolsTimeout = "list_tools_timeout"
	// ConfigKeyCallToolTimeout is the timeout for CallTool operations (in seconds)
	ConfigKeyCallToolTimeout = "call_tool_timeout"
	// ConfigKeyConnectTimeout is the timeout for connecting to upstream servers (in seconds)
	ConfigKeyConnectTimeout = "connect_timeout"
	// ConfigKeyHealthCheckTimeout is the timeout for health check operations (in seconds)
	ConfigKeyHealthCheckTimeout = "health_check_timeout"
	// ConfigKeyHealthCheckInterval is the interval between health checks (in seconds)
	ConfigKeyHealthCheckInterval = "health_check_interval"
)

// DefaultConfigValues holds the default values for system config
var DefaultConfigValues = map[string]string{
	ConfigKeyListToolsTimeout:      "30",
	ConfigKeyCallToolTimeout:       "30",
	ConfigKeyConnectTimeout:       "30",
	ConfigKeyHealthCheckTimeout:    "5",
	ConfigKeyHealthCheckInterval:  "30",
}

// ConfigDefinition describes a configuration item
type ConfigDefinition struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Unit        string `json:"unit"`
	DefaultValue string `json:"default_value"`
}

// GetConfigDefinitions returns all configuration definitions with descriptions
func GetConfigDefinitions() []ConfigDefinition {
	return []ConfigDefinition{
		{
			Key:          ConfigKeyListToolsTimeout,
			Description:  "ListTools 操作超时时间",
			Category:     "timeout",
			Unit:         "秒",
			DefaultValue: DefaultConfigValues[ConfigKeyListToolsTimeout],
		},
		{
			Key:          ConfigKeyCallToolTimeout,
			Description:  "CallTool 操作超时时间",
			Category:     "timeout",
			Unit:         "秒",
			DefaultValue: DefaultConfigValues[ConfigKeyCallToolTimeout],
		},
		{
			Key:          ConfigKeyConnectTimeout,
			Description:  "连接上游服务器超时时间",
			Category:     "timeout",
			Unit:         "秒",
			DefaultValue: DefaultConfigValues[ConfigKeyConnectTimeout],
		},
		{
			Key:          ConfigKeyHealthCheckTimeout,
			Description:  "健康检查操作超时时间",
			Category:     "timeout",
			Unit:         "秒",
			DefaultValue: DefaultConfigValues[ConfigKeyHealthCheckTimeout],
		},
		{
			Key:          ConfigKeyHealthCheckInterval,
			Description:  "健康检查循环间隔",
			Category:     "interval",
			Unit:         "秒",
			DefaultValue: DefaultConfigValues[ConfigKeyHealthCheckInterval],
		},
	}
}