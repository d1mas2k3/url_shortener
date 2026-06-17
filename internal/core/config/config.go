package core_config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
)

// Config — общий конфиг приложения
type Config struct {
	// postgres или memory
	Storage string `envconfig:"STORAGE" required:"true"`
	// базовый URL сервиса для формирования коротких ссылок
	BaseURL string `envconfig:"BASE_URL" default:"http://localhost:8080"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("APP", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}
	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get app config: %w", err))
	}
	return config
}