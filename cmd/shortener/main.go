package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/TessaLir/go_practic_shortener/cmd/shortener/config"
	"github.com/TessaLir/go_practic_shortener/internal/handler"
	"github.com/TessaLir/go_practic_shortener/internal/repository"
	"github.com/TessaLir/go_practic_shortener/internal/service"
)

// Настройка и возврат роутера с зарегистрированными маршрутами
func setupRouter(cfg *config.Config) *chi.Mux {
	// Инициализируем слои приложения
	storage := repository.NewStorage()
	shortenerService := service.NewShortenerService(storage)
	shortenerHandler := handler.NewShortenerHandler(shortenerService, cfg)

	r := chi.NewRouter()

	r.Post(`/`, shortenerHandler.MainPage())
	r.Get(`/{id}`, shortenerHandler.URLDetailPage())

	return r
}

// Точка входа
func main() {
	// Инициализируем конфигурацию из аргументов командной строки
	cfg := config.Init()

	r := setupRouter(cfg)

	err := http.ListenAndServe(cfg.ServerPort, r)
	if err != nil {
		panic(err)
	}
}
