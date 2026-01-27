package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/TessaLir/go_practic_shortener/cmd/shortener/config"
	"github.com/TessaLir/go_practic_shortener/internal/service"
)

// ShortenerHandler содержит HTTP обработчики
type ShortenerHandler struct {
	service *service.ShortenerService
	cfg     *config.Config
}

// NewShortenerHandler создает новый экземпляр ShortenerHandler
func NewShortenerHandler(service *service.ShortenerService, cfg *config.Config) *ShortenerHandler {
	return &ShortenerHandler{
		service: service,
		cfg:     cfg,
	}
}

// MainPage обрабатывает запросы для главной страницы с POST методом
func (h *ShortenerHandler) MainPage() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Проверка метода, пропускаем только POST
		if req.Method != http.MethodPost {
			http.Error(res, "Данный запрос не поддерживает выбранный метод.", http.StatusBadRequest)
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Ошибка чтения тела запроса", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		// Получаем URL из тела запроса
		urlStr := strings.TrimSpace(string(body))

		// Выполняем проверку полученного URL на пустую строку и является ли данный URL валидным
		switch {
		case urlStr == "":
			http.Error(res, "Невозможно распарсить полученный сокращенный URL - URL не может быть пустым", http.StatusBadRequest)
			return
		case !h.service.IsValidURL(urlStr):
			http.Error(res, "передан не валидный URL", http.StatusBadRequest)
			return
		}

		// Сохраняем URL и получаем короткий ключ
		hashString := h.service.SaveURL(urlStr)

		// Записываем заголовки, присваиваем статус и отдаем ответ сервера
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(h.cfg.BaseURL + h.cfg.ServerPort + "/" + hashString))
	}
}

// URLDetailPage обрабатывает запрос получения URL и редиректа на сайт по короткой ссылке
func (h *ShortenerHandler) URLDetailPage() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// Получаем хеш из параметра
		id := chi.URLParam(req, "id")
		// Fallback для тестов, где роутер не обрабатывает запрос
		if id == "" {
			id = req.URL.Path[1:]
		}

		// Смотрим нашу БД, если ничего не находим, возвращаем ошибку
		urlStr, found := h.service.GetURL(id)
		if !found {
			http.Error(res, "Сайт не найден", http.StatusNotFound)
			return
		}

		// Если все ОК, задаем заголовок и делаем редирект
		res.Header().Set("Location", urlStr)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
