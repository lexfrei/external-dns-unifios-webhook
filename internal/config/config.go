package config

import (
	"strings"

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

// DebugConfig contains debug/profiling settings.
type DebugConfig struct {
	PprofEnabled bool   `mapstructure:"pprof_enabled"`
	PprofPort    string `mapstructure:"pprof_port"`
}

// Config represents the complete application configuration.
type Config struct {
	UniFi        UniFiConfig        `mapstructure:"unifi"`
	Server       ServerConfig       `mapstructure:"server"`
	Health       HealthConfig       `mapstructure:"health"`
	DomainFilter DomainFilterConfig `mapstructure:"domain_filter"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Debug        DebugConfig        `mapstructure:"debug"`
}

// Load loads configuration from environment variables and config files.
func Load() (*Config, error) {
	viperConfig := viper.New()

	viperConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viperConfig.AutomaticEnv()

	// Set defaults
	setDefaults(viperConfig)

	// Explicitly bind environment variables with hardcoded WEBHOOK_ prefix
	// AutomaticEnv() doesn't automatically bind nested keys
	_ = viperConfig.BindEnv("unifi.api_key", "WEBHOOK_UNIFI_API_KEY")
	_ = viperConfig.BindEnv("unifi.host", "WEBHOOK_UNIFI_HOST")
	_ = viperConfig.BindEnv("unifi.site", "WEBHOOK_UNIFI_SITE")
	_ = viperConfig.BindEnv("unifi.skip_tls_verify", "WEBHOOK_UNIFI_SKIP_TLS_VERIFY")
	_ = viperConfig.BindEnv("server.host", "WEBHOOK_SERVER_HOST")
	_ = viperConfig.BindEnv("server.port", "WEBHOOK_SERVER_PORT")
	_ = viperConfig.BindEnv("health.host", "WEBHOOK_HEALTH_HOST")
	_ = viperConfig.BindEnv("health.port", "WEBHOOK_HEALTH_PORT")
	_ = viperConfig.BindEnv("logging.level", "WEBHOOK_LOGGING_LEVEL")
	_ = viperConfig.BindEnv("logging.format", "WEBHOOK_LOGGING_FORMAT")

	// Try to read config file
	viperConfig.SetConfigName("config")
	viperConfig.SetConfigType("yaml")
	viperConfig.AddConfigPath(".")
	viperConfig.AddConfigPath("/etc/external-dns-unifios-webhook/")
	viperConfig.AddConfigPath("$HOME/.external-dns-unifios-webhook/")

	// Ignore config file not found error (config via env vars is fine)
	err := viperConfig.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	var cfg Config

	err = viperConfig.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// Validate required configuration
	if cfg.UniFi.Host == "" {
		return nil, errors.New("WEBHOOK_UNIFI_HOST is required")
	}

	if cfg.UniFi.APIKey == "" {
		return nil, errors.New("WEBHOOK_UNIFI_API_KEY is required")
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(viperConfig *viper.Viper) {
	// UniFi defaults
	// NOTE: unifi.host has no default - must be explicitly configured
	viperConfig.SetDefault("unifi.site", "default")
	viperConfig.SetDefault("unifi.skip_tls_verify", true)

	// Server defaults
	viperConfig.SetDefault("server.host", "localhost")
	viperConfig.SetDefault("server.port", "8888")

	// Health defaults
	viperConfig.SetDefault("health.host", "0.0.0.0")
	viperConfig.SetDefault("health.port", "8080")

	// Logging defaults
	viperConfig.SetDefault("logging.level", "info")
	viperConfig.SetDefault("logging.format", "json")
}
