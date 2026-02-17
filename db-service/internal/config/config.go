package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config содержит все конфигурации приложения
type Config struct {
	App   AppConfig   `yaml:"app" env-prefix:"APP_"`
	DB    DBConfig    `yaml:"db" env-prefix:"DB_"`
	GRPC  GRPCConfig  `yaml:"grpc" env-prefix:"GRPC_"`
	Redis RedisConfig `yaml:"redis" env-prefix:"REDIS_"`
}

// AppConfig содержит настройки приложения
type AppConfig struct {
	Name    string `yaml:"name" env:"NAME" env-default:"todo-db-service"`
	Version string `yaml:"version" env:"VERSION" env-default:"1.0.0"`
	Env     string `yaml:"env" env:"ENV" env-default:"development"`
	Debug   bool   `yaml:"debug" env:"DEBUG" env-default:"false"`
}

// DBConfig содержит настройки PostgreSQL
type DBConfig struct {
	Host     string        `yaml:"host" env:"HOST" env-default:"localhost"`
	Port     int           `yaml:"port" env:"PORT" env-default:"5432"`
	User     string        `yaml:"user" env:"USER" env-default:"postgres"`
	Password string        `yaml:"password" env:"PASSWORD" env-default:"postgres"`
	Name     string        `yaml:"name" env:"NAME" env-default:"todo_db"`
	SSLMode  string        `yaml:"ssl_mode" env:"SSLMODE" env-default:"disable"`
	MaxConns int           `yaml:"max_conns" env:"MAX_CONNS" env-default:"10"`
	Timeout  time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"5s"`
}

// GRPCConfig содержит настройки gRPC сервера
type GRPCConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"PORT" env-default:"50051"`
}

// RedisConfig содержит настройки Redis (для будущего кэширования)
type RedisConfig struct {
	Host     string        `yaml:"host" env:"HOST" env-default:"localhost"`
	Port     int           `yaml:"port" env:"PORT" env-default:"6379"`
	Password string        `yaml:"password" env:"PASSWORD" env-default:""`
	DB       int           `yaml:"db" env:"DB" env-default:"0"`
	Enabled  bool          `yaml:"enabled" env:"ENABLED" env-default:"false"`
	TTL      time.Duration `yaml:"ttl" env:"TTL" env-default:"60s"`
}

// Load загружает конфигурацию из файла и переменных окружения
func Load(configPath string) (*Config, error) {
	var cfg Config

	// Если путь к конфигу не указан, используем значения по умолчанию из env
	if configPath == "" {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read environment variables: %w", err)
		}
		return &cfg, nil
	}

	// Загружаем из файла (YAML) и переопределяем env переменными
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Читаем env переменные (они имеют приоритет над файлом)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables: %w", err)
	}

	return &cfg, nil
}

// GetConfigPath возвращает путь к конфиг файлу из env или находит его
func GetConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		return configPath
	}

	// Проверяем существующие файлы конфигурации
	possiblePaths := []string{
		"config.yaml",
		"config.yml",
		"config/config.yaml",
		"config/config.yml",
		"../config.yaml",
		"../config.yml",
		"./db-service/config.yaml",
		"./db-service/config.yml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Если файл не найден, возвращаем пустую строку
	// (будет использоваться только env)
	return ""
}

// DSN возвращает строку подключения к PostgreSQL
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// DSNWithTimeout возвращает DSN с таймаутом
func (c *DBConfig) DSNWithTimeout() string {
	return fmt.Sprintf("%s connect_timeout=%d", c.DSN(), int(c.Timeout.Seconds()))
}

// Address возвращает адрес для gRPC сервера
func (c *GRPCConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Address возвращает адрес для Redis
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsProduction проверяет, production ли среда
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

// IsDevelopment проверяет, development ли среда
func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.App.Env, "development")
}

// IsTesting проверяет, testing ли среда
func (c *Config) IsTesting() bool {
	return strings.EqualFold(c.App.Env, "test") || strings.EqualFold(c.App.Env, "testing")
}

// Print печатает конфигурацию (без чувствительных данных)
func (c *Config) Print() {
	fmt.Println("=== Application Configuration ===")
	fmt.Printf("Name: %s\n", c.App.Name)
	fmt.Printf("Version: %s\n", c.App.Version)
	fmt.Printf("Environment: %s\n", c.App.Env)
	fmt.Printf("Debug: %v\n", c.App.Debug)
	fmt.Println()

	fmt.Println("=== Database Configuration ===")
	fmt.Printf("Host: %s\n", c.DB.Host)
	fmt.Printf("Port: %d\n", c.DB.Port)
	fmt.Printf("User: %s\n", c.DB.User)
	fmt.Printf("Database: %s\n", c.DB.Name)
	fmt.Printf("SSL Mode: %s\n", c.DB.SSLMode)
	fmt.Printf("Max Connections: %d\n", c.DB.MaxConns)
	fmt.Printf("Timeout: %v\n", c.DB.Timeout)
	fmt.Println()

	fmt.Println("=== gRPC Configuration ===")
	fmt.Printf("Host: %s\n", c.GRPC.Host)
	fmt.Printf("Port: %d\n", c.GRPC.Port)
	fmt.Printf("Address: %s\n", c.GRPC.Address())
	fmt.Println()

	fmt.Println("=== Redis Configuration ===")
	fmt.Printf("Host: %s\n", c.Redis.Host)
	fmt.Printf("Port: %d\n", c.Redis.Port)
	fmt.Printf("Enabled: %v\n", c.Redis.Enabled)
	if c.Redis.Enabled {
		fmt.Printf("DB: %d\n", c.Redis.DB)
	}
	fmt.Println("============================")
}

// Validate проверяет валидность конфигурации
func (c *Config) Validate() error {
	var errors []string

	// Проверка App
	if c.App.Name == "" {
		errors = append(errors, "app.name is required")
	}

	// Проверка DB
	if c.DB.Host == "" {
		errors = append(errors, "db.host is required")
	}
	if c.DB.Port <= 0 || c.DB.Port > 65535 {
		errors = append(errors, "db.port must be between 1 and 65535")
	}
	if c.DB.User == "" {
		errors = append(errors, "db.user is required")
	}
	if c.DB.Name == "" {
		errors = append(errors, "db.name is required")
	}

	// Проверка GRPC
	if c.GRPC.Port <= 0 || c.GRPC.Port > 65535 {
		errors = append(errors, "grpc.port must be between 1 and 65535")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// GetLoggerConfig возвращает настройки для логгера
func (c *Config) GetLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Level:       c.getLogLevel(),
		JSONFormat:  c.IsProduction(),
		Development: c.IsDevelopment(),
	}
}

// getLogLevel определяет уровень логирования
func (c *Config) getLogLevel() string {
	if c.App.Debug {
		return "debug"
	}
	if c.IsProduction() {
		return "info"
	}
	return "debug"
}

// LoggerConfig содержит настройки для логгера
type LoggerConfig struct {
	Level       string
	JSONFormat  bool
	Development bool
}
