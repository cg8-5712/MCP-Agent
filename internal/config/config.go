package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Database    DatabaseConfig    `mapstructure:"database"`
	HealthCheck HealthCheckConfig `mapstructure:"health_check"`
	MCPServers  []MCPServerConfig `mapstructure:"mcp_servers"`
	Log         LogConfig         `mapstructure:"log"`
	LLM         LLMConfig         `mapstructure:"llm"`
	Embedding   EmbeddingConfig   `mapstructure:"embedding"`
	VectorDB    VectorDBConfig    `mapstructure:"vector_db"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	DSN      string `mapstructure:"dsn"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type HealthCheckConfig struct {
	IntervalSeconds int `mapstructure:"interval_seconds"`
	TimeoutSeconds  int `mapstructure:"timeout_seconds"`
}

type MCPServerConfig struct {
	Name string `mapstructure:"name"`
	URL  string `mapstructure:"url"`
}

type LLMConfig struct {
	Provider    string      `mapstructure:"provider"`
	APIKey      string      `mapstructure:"api_key"`
	BaseURL     string      `mapstructure:"base_url"`
	Model       ModelConfig `mapstructure:"model"`
	Temperature float64     `mapstructure:"temperature"`
	MaxTokens   int         `mapstructure:"max_tokens"`
}

type ModelConfig struct {
	Planner string `mapstructure:"planner"`
	Final   string `mapstructure:"final"`
}

type EmbeddingConfig struct {
	Provider  string `mapstructure:"provider"`
	APIKey    string `mapstructure:"api_key"`
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	Dimension int    `mapstructure:"dimension"`
}

type VectorDBConfig struct {
	Provider   string `mapstructure:"provider"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Collection string `mapstructure:"collection"`
	UseMemory  bool   `mapstructure:"use_memory"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	viper.SetDefault("server.port", 8000)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dsn", "./data/mcp_agent.db")
	viper.SetDefault("health_check.interval_seconds", 30)
	viper.SetDefault("health_check.timeout_seconds", 3)
	viper.SetDefault("log.level", "info")

	viper.SetEnvPrefix("MCP")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// 支持环境变量覆盖 JWT Secret
	if secret := viper.GetString("jwt.secret"); secret == "" {
		viper.Set("jwt.secret", viper.GetString("JWT_SECRET"))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
