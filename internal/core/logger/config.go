package core_logger

// Суть config.go в том, чтобы мы могли на разных пк запускать logger без проблем с окружением

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// // Структура с настройками логгера: уровень логов (LOGGER_LEVEL) и папка для файлов (LOGGER_FOLDER)
type Config struct {
	Level  string `envconfig:"LEVEL"  default:"DEBUG"`
	Folder string `envconfig:"FOLDER" required:"true"`
}

// Читает переменные окружения и заполняет структуру Config для файла logger.go
func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("LOGGER", &config); err != nil {
		return Config{}, fmt.Errorf("proccess envconfig: %w", err)
	}
	return config, nil
}

// Вызывает NewConfig() при старте приложения. Если конфиг не удалось прочитать - паникует
func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Logger config: %w", err)
		panic(err)
	}
	return config
}