package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port               string `yaml:"port"`
		RateLimitPerMinute int    `yaml:"rate-limit-per-minute"`
		RateLimitPerHour   int    `yaml:"rate-limit-per-hour"`
		RateLimitRPS       int    `yaml:"rate-limit-rps"`
	} `yaml:"server"`

	MySQL struct {
		Addr                  string        `yaml:"addr"`
		Username              string        `yaml:"username"`
		Password              string        `yaml:"password"`
		Database              string        `yaml:"database"`
		MaxIdleConnections    int           `yaml:"max-idle-connections"`
		MaxOpenConnections    int           `yaml:"max-open-connections"`
		MaxConnectionLifeTime time.Duration `yaml:"max-connection-life-time"`
	} `yaml:"mysql"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
}

var appConfig *Config

func LoadConfig() (*Config, error) {
	configPath, explicitPath := resolveConfigPath()

	cfg := &Config{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		if explicitPath || !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	applyDefaults(cfg)
	if err := applyEnvOverrides(cfg); err != nil {
		return nil, err
	}

	appConfig = cfg
	return cfg, nil
}

func resolveConfigPath() (string, bool) {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath, true
	}

	return "./config.yaml", false
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.RateLimitPerMinute == 0 {
		cfg.Server.RateLimitPerMinute = 60
	}
	if cfg.Server.RateLimitPerHour == 0 {
		cfg.Server.RateLimitPerHour = 1000
	}
	if cfg.Server.RateLimitRPS == 0 {
		cfg.Server.RateLimitRPS = 1000
	}

	if cfg.MySQL.Addr == "" {
		cfg.MySQL.Addr = "127.0.0.1:3306"
	}
	if cfg.MySQL.Username == "" {
		cfg.MySQL.Username = "root"
	}
	if cfg.MySQL.Database == "" {
		cfg.MySQL.Database = "app"
	}
	if cfg.MySQL.MaxIdleConnections == 0 {
		cfg.MySQL.MaxIdleConnections = 100
	}
	if cfg.MySQL.MaxOpenConnections == 0 {
		cfg.MySQL.MaxOpenConnections = 100
	}
	if cfg.MySQL.MaxConnectionLifeTime == 0 {
		cfg.MySQL.MaxConnectionLifeTime = 10 * time.Second
	}

	if cfg.Redis.Addr == "" {
		cfg.Redis.Addr = "127.0.0.1:6379"
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "change-me-in-production"
		fmt.Println("WARNING: Using default JWT secret, please set jwt.secret or JWT_SECRET")
	}
}

func applyEnvOverrides(cfg *Config) error {
	overrideString(&cfg.Server.Port, "SERVER_PORT")
	if err := overrideInt(&cfg.Server.RateLimitPerMinute, "SERVER_RATE_LIMIT_PER_MINUTE"); err != nil {
		return err
	}
	if err := overrideInt(&cfg.Server.RateLimitPerHour, "SERVER_RATE_LIMIT_PER_HOUR"); err != nil {
		return err
	}
	if err := overrideInt(&cfg.Server.RateLimitRPS, "SERVER_RATE_LIMIT_RPS"); err != nil {
		return err
	}

	overrideString(&cfg.MySQL.Addr, "MYSQL_ADDR")
	overrideString(&cfg.MySQL.Username, "MYSQL_USERNAME")
	overrideString(&cfg.MySQL.Password, "MYSQL_PASSWORD")
	overrideString(&cfg.MySQL.Database, "MYSQL_DATABASE")
	if err := overrideInt(&cfg.MySQL.MaxIdleConnections, "MYSQL_MAX_IDLE_CONNECTIONS"); err != nil {
		return err
	}
	if err := overrideInt(&cfg.MySQL.MaxOpenConnections, "MYSQL_MAX_OPEN_CONNECTIONS"); err != nil {
		return err
	}
	if err := overrideDuration(&cfg.MySQL.MaxConnectionLifeTime, "MYSQL_MAX_CONNECTION_LIFE_TIME"); err != nil {
		return err
	}

	overrideString(&cfg.Redis.Addr, "REDIS_ADDR")
	overrideString(&cfg.Redis.Password, "REDIS_PASSWORD")
	if err := overrideInt(&cfg.Redis.DB, "REDIS_DB"); err != nil {
		return err
	}

	overrideString(&cfg.JWT.Secret, "JWT_SECRET")

	return nil
}

func overrideString(target *string, key string) {
	if value, ok := os.LookupEnv(key); ok {
		*target = value
	}
}

func overrideInt(target *int, key string) error {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %w", key, err)
	}
	*target = parsed
	return nil
}

func overrideDuration(target *time.Duration, key string) error {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %w", key, err)
	}
	*target = parsed
	return nil
}

func GetConfig() *Config {
	if appConfig == nil {
		if _, err := LoadConfig(); err != nil {
			panic(fmt.Errorf("failed to load config: %w", err))
		}
	}

	return appConfig
}

func GetMySQLConfig() struct {
	Addr                  string
	Username              string
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
} {
	cfg := GetConfig()
	return struct {
		Addr                  string
		Username              string
		Password              string
		Database              string
		MaxIdleConnections    int
		MaxOpenConnections    int
		MaxConnectionLifeTime time.Duration
	}{
		Addr:                  cfg.MySQL.Addr,
		Username:              cfg.MySQL.Username,
		Password:              cfg.MySQL.Password,
		Database:              cfg.MySQL.Database,
		MaxIdleConnections:    cfg.MySQL.MaxIdleConnections,
		MaxOpenConnections:    cfg.MySQL.MaxOpenConnections,
		MaxConnectionLifeTime: cfg.MySQL.MaxConnectionLifeTime,
	}
}
