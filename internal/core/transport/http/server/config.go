package core_http_server

// Нужен для настройки HTTP сервера и хранения настроек из env переменных

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr            string        `envconfig:"ADDR" required:"true"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

// Заполняет config переменными окружения; в случае ошибки, просто возвращает ошибку
func NewConfig() (Config, error) {
	var config Config

	// Читаем env переменные и заполняем структуру
	if err := envconfig.Process("HTTP", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return config, nil
}

// Нужен для настройки HTTP сервера и хранения настроек из env переменных
// В случае ошибки функции NewConfig(), вызывает panic()
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err := fmt.Errorf("get HTTP server config: %w", err)
		panic(err)
	}
	return config
}
