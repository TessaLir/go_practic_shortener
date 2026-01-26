package config

import (
	"flag"
)

// Config содержит конфигурацию сервиса
type Config struct {
	ServerPort string // Адрес запуска HTTP-сервера (например, localhost:8888)
	BaseURL    string // Базовый адрес результирующего сокращённого URL (например, http://localhost:8000)
}

// Init инициализирует конфигурацию из аргументов командной строки
func Init() *Config {
	cfg := &Config{}

	// Флаг -a отвечает за адрес запуска HTTP-сервера
	flag.StringVar(&cfg.ServerPort, "a", ":8080", "address and port to run server")

	// Флаг -b отвечает за базовый адрес результирующего сокращённого URL
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost", "base URL for shortened links")

	// Парсим переданные серверу аргументы
	flag.Parse()

	return cfg
}
