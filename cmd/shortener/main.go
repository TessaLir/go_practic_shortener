package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/TessaLir/go_practic_shortener/cmd/shortener/config"
	"github.com/TessaLir/go_practic_shortener/internal/repository"
)

// Проверка переданного URL на валидность.
func isValidURL(str string) bool {

	// смотрим URL
	parsedURL, err := url.Parse(str)

	// Если что возвращаем false
	if err != nil {
		return false
	}

	// возвращаем результат окончательной проверки URl
	return parsedURL.Scheme != "" && parsedURL.Host != ""

}

// Генерация рандомной короткой строки длиной в 8 символов
func generateShortURL(storage *repository.Storage) string {

	// Некоторые константы, которые в теории можно вынести в глобальные, но не будем этого делать
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	// Создаем массив байтов и записываем туда символы из паттерна
	result := make([]byte, length)

	rng := storage.GetRNG()

	// цикл
	for {

		// Генерируем случайную строку
		for i := range result {
			result[i] = charset[rng.Intn(len(charset))]
		}

		// Если сгенерированная строка - уникальна, выходим из цикла
		if !storage.Exists(string(result)) {
			break
		}
	}

	// возвращаем результат
	return string(result)

}

// Метод обработки запросов для главной страницы с POST методом.
func mainPage(cfg *config.Config, storage *repository.Storage) http.HandlerFunc {
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
		case !isValidURL(urlStr):
			http.Error(res, "передан не валидный URL", http.StatusBadRequest)
			return
		}

		// Получаем уникальный хещ
		hashString := generateShortURL(storage)

		// Записываем ключ - значение в БД (импровизированную)
		storage.Save(hashString, urlStr)

		// Записываем заголовки, присваиваем статус и отдаем ответ сервера
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(cfg.BaseURL + cfg.ServerPort + "/" + hashString))

	}
}

// Метод обработки запроса получения URL и редиректа на сайт по короткой ссылке.
func urlDetailPage(storage *repository.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		// Получаем хеш из параметра
		id := chi.URLParam(req, "id")
		// Fallback для тестов, где роутер не обрабатывает запрос
		if id == "" {
			id = req.URL.Path[1:]
		}

		// Смотрим нашу БД (импровизированную), если ничего не находим, возвращаем ошибку
		urlStr, found := storage.Get(id)
		if !found {
			http.Error(res, "Сайт не найден", http.StatusBadRequest)
			return
		}

		// Если все ОК, задаем заголовок и делаем редирект
		res.Header().Set("Location", urlStr)
		res.WriteHeader(http.StatusTemporaryRedirect)

	}
}

// Настройка и возврат роутера с зарегистрированными маршрутами
func setupRouter(cfg *config.Config, storage *repository.Storage) *chi.Mux {
	r := chi.NewRouter()

	r.Post(`/`, mainPage(cfg, storage))
	r.Get(`/{id}`, urlDetailPage(storage))

	return r
}

// Точка входа
func main() {

	// Инициализируем конфигурацию из аргументов командной строки
	cfg := config.Init()

	// Создаем хранилище
	storage := repository.NewStorage()

	r := setupRouter(cfg, storage)

	err := http.ListenAndServe(cfg.ServerPort, r)
	if err != nil {
		panic(err)
	}

}
