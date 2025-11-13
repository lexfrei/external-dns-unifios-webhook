package config

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

// UniFiConfig contains UniFi controller connection settings.
type UniFiConfig struct {
	Host          string `mapstructure:"host"`
	APIKey        string `mapstructure:"api_key"`
	Site          string `mapstructure:"site"`
	SkipTLSVerify bool   `mapstructure:"skip_tls_verify"`
}

// ServerConfig contains webhook server settings.
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// HealthConfig contains health check server settings.
type HealthConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

// DomainFilterConfig contains domain filtering settings.
type DomainFilterConfig struct {
	Filters             []string `mapstructure:"filters"`
	ExcludeFilters      []string `mapstructure:"exclude_filters"`
	RegexFilters        []string `mapstructure:"regex_filters"`
	RegexExcludeFilters []string `mapstructure:"regex_exclude_filters"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Config represents the complete application configuration.
type Config struct {
	UniFi        UniFiConfig        `mapstructure:"unifi"`
	Server       ServerConfig       `mapstructure:"server"`
	Health       HealthConfig       `mapstructure:"health"`
	DomainFilter DomainFilterConfig `mapstructure:"domain_filter"`
	Logging      LoggingConfig      `mapstructure:"logging"`
}

// Load loads configuration from environment variables and config files.
func Load() (*Config, error) {
	v := viper.New()

	// Set prefix for environment variables
	v.SetEnvPrefix("WEBHOOK")
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Try to read config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/external-dns-unifios-webhook/")
	v.AddConfigPath("$HOME/.external-dns-unifios-webhook/")

	// Ignore config file not found error
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// UniFi defaults
	v.SetDefault("unifi.host", "https://unifi.local")
	v.SetDefault("unifi.site", "default")
	v.SetDefault("unifi.skip_tls_verify", false)

	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8888")

	// Health defaults
	v.SetDefault("health.host", "0.0.0.0")
	v.SetDefault("health.port", "8080")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}
