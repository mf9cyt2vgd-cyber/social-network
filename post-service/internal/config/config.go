package config

// package provides configs from post-service/config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Config is a structure with all needed resources
type Config struct {
	Env              string `yaml:"env" env-default:"local"`
	ServiceName      string `yaml:"serviceName" env-default:"post-service"`
	ServiceVersion   string `yaml:"service_version" env-default:"0.0.1"`
	TelemetryEnabled string `yaml:"telemetry_enabled" env-default:"true"`
	DatabaseURL      string `env:"DATABASE_URL,required"`
	HTTPServer HTTPServer       `yaml:"http_server"`
	Kafka Kafka            `yaml:"kafka"`
	Redis Redis            `yaml:"redis"`
}

// HTTPServer includes basic variables to configure server
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// Kafka includes basic variables to configure kafka-go worker
type Kafka struct {
	Brokers []string `yaml:"brokers" env-default:"localhost:9092"`
	Topic   string   `yaml:"topic" env-default:"posts"`
	GroupID string   `yaml:"group_id" env-default:"notification-service"`
}

// Redis includes basic variables to configure redis-client
type Redis struct {
	Addr string `yaml:"address" env-default:"localhost:6379"`
	DB   int    `yaml:"db" env-default:"0"`
}

// MustLoad is a critical function that return ready config for app
func MustLoad() *Config {
	// Пробуем подгрузить .env, если он есть (в Docker его может не быть)
	_ = godotenv.Load()

	// Определяем окружение
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	// Путь до yaml
	configPath := filepath.Join("config", fmt.Sprintf("%s.yaml", env))

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file not found: %s", configPath))
	}

	// Загружаем конфиг
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	// Проверяем DATABASE_URL (берем из окружения)
	if cfg.DatabaseURL = os.Getenv("DATABASE_URL"); cfg.DatabaseURL == "" {
		panic("DATABASE_URL not set in environment")
	}

	fmt.Printf("✅ Loaded config for environment: %s\n", env)
	return &cfg
}
