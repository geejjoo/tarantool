package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App        AppConfig        `yaml:"app"`
	HTTPServer HTTPServerConfig `yaml:"http_server"`
	Tarantool  TarantoolConfig  `yaml:"tarantool"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type HTTPServerConfig struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type TarantoolConfig struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
}

func Load(configPath string) (*Config, error) {
	_ = godotenv.Load() // Не паникуем, если файла нет

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Приоритет ENV
	config.App.Environment = getEnv("APP_ENV", config.App.Environment)
	config.HTTPServer.Port = getEnv("HTTP_PORT", config.HTTPServer.Port)
	config.HTTPServer.ReadTimeout = getEnvDuration("HTTP_READ_TIMEOUT", config.HTTPServer.ReadTimeout)
	config.HTTPServer.WriteTimeout = getEnvDuration("HTTP_WRITE_TIMEOUT", config.HTTPServer.WriteTimeout)

	config.Tarantool.Host = getEnv("TARANTOOL_HOST", config.Tarantool.Host)
	config.Tarantool.Port = getEnvInt("TARANTOOL_PORT", config.Tarantool.Port)
	config.Tarantool.Username = getEnv("TARANTOOL_USERNAME", config.Tarantool.Username)
	config.Tarantool.Password = getEnv("TARANTOOL_PASSWORD", config.Tarantool.Password)
	config.Tarantool.Timeout = getEnvDuration("TARANTOOL_TIMEOUT", config.Tarantool.Timeout)

	return &config, nil
}

func MustLoad(configPath string) *Config {
	config, err := Load(configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return fallback
}
