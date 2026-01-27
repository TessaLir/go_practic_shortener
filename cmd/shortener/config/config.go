package config

import (
	"flag"
	"net/url"
	"strings"
)

// Config содержит конфигурацию сервиса
type Config struct {
	ServerPort string // Адрес запуска HTTP-сервера (например, :8080)
	BaseURL    string // Базовый адрес результирующего сокращённого URL (например, http://localhost)
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

	// Валидация ServerPort: должен начинаться с двоеточия и содержать только порт
	if cfg.ServerPort != "" {
		// Если ServerPort содержит хост (например, localhost:8080), извлекаем только порт
		if strings.Contains(cfg.ServerPort, ":") && !strings.HasPrefix(cfg.ServerPort, ":") {
			parts := strings.Split(cfg.ServerPort, ":")
			if len(parts) > 1 {
				cfg.ServerPort = ":" + parts[len(parts)-1]
			}
		} else if !strings.HasPrefix(cfg.ServerPort, ":") {
			// Если порт не начинается с двоеточия и не содержит двоеточие, добавляем его
			cfg.ServerPort = ":" + cfg.ServerPort
		}
	}

	// Валидация BaseURL: должен быть протокол + хост без порта
	parsedURL, err := url.Parse(cfg.BaseURL)
	if err == nil && parsedURL.Host != "" {
		// Удаляем порт из Host, если он есть
		if strings.Contains(parsedURL.Host, ":") {
			hostParts := strings.Split(parsedURL.Host, ":")
			parsedURL.Host = hostParts[0]
			// Восстанавливаем URL: scheme://host + path + query + fragment
			result := parsedURL.Scheme + "://" + parsedURL.Host
			if parsedURL.Path != "" {
				result += parsedURL.Path
			}
			if parsedURL.RawQuery != "" {
				result += "?" + parsedURL.RawQuery
			}
			if parsedURL.Fragment != "" {
				result += "#" + parsedURL.Fragment
			}
			cfg.BaseURL = result
		}
	}

	return cfg
}
