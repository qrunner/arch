// Package config handles application configuration loading from YAML files and
// environment variables using Viper.
package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config is the top-level application configuration.
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Neo4j      Neo4jConfig      `mapstructure:"neo4j"`
	NATS       NATSConfig       `mapstructure:"nats"`
	Collectors []CollectorEntry `mapstructure:"collectors"`
	Scheduler  SchedulerConfig  `mapstructure:"scheduler"`
	Notifier   NotifierConfig   `mapstructure:"notifier"`
	Log        LogConfig        `mapstructure:"log"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// Address returns the listen address string.
func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

// Neo4jConfig holds Neo4j connection settings.
type Neo4jConfig struct {
	URI      string `mapstructure:"uri"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// NATSConfig holds NATS connection settings.
type NATSConfig struct {
	URL string `mapstructure:"url"`
}

// CollectorEntry is a named collector configuration in the config file.
type CollectorEntry struct {
	Name     string            `mapstructure:"name"`
	Type     string            `mapstructure:"type"`
	Enabled  bool              `mapstructure:"enabled"`
	Interval string            `mapstructure:"interval"`
	Settings map[string]string `mapstructure:"settings"`
}

// SchedulerConfig holds job scheduling settings.
type SchedulerConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// NotifierConfig holds notification channel settings.
type NotifierConfig struct {
	Webhook  *WebhookConfig  `mapstructure:"webhook"`
	Email    *EmailConfig    `mapstructure:"email"`
	Telegram *TelegramConfig `mapstructure:"telegram"`
}

// WebhookConfig holds webhook notification settings.
type WebhookConfig struct {
	URL string `mapstructure:"url"`
}

// EmailConfig holds SMTP email notification settings.
type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	From     string `mapstructure:"from"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// TelegramConfig holds Telegram bot notification settings.
type TelegramConfig struct {
	Token  string `mapstructure:"token"`
	ChatID string `mapstructure:"chat_id"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load reads the configuration from file and environment variables.
// It searches for config.yaml in the paths: ./configs, /etc/arch, and $HOME/.arch.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "arch")
	v.SetDefault("database.password", "arch")
	v.SetDefault("database.dbname", "arch")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("neo4j.uri", "bolt://localhost:7687")
	v.SetDefault("neo4j.user", "neo4j")
	v.SetDefault("neo4j.password", "neo4j")
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath("./configs")
	v.AddConfigPath("/etc/arch")
	v.AddConfigPath("$HOME/.arch")

	v.SetEnvPrefix("ARCH")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// Config file not found is fine - use defaults and env vars.
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
