package main

import (
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

var urlStore = make(map[string]string)
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

var HOST = "localhost"
var PORT = "8080"
var URL = "http://" + HOST + ":" + PORT
var URL_DETAIL = URL + "/{id}"

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
func generateShortURL() string {

	// Некоторые константы, которые в теории можно вынести в глобальные, но не будем этого делать
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	// Создаем массив байтов и записываем туда символы из паттерна
	result := make([]byte, length)

	// цикл
	for {

		// Генерируем случайную строку
		for i := range result {
			result[i] = charset[rng.Intn(len(charset))]
		}

		// Если сгенерированная строка - уникальна, выходим из цикла
		if _, exists := urlStore[string(result)]; !exists {
			break
		}
	}

	// возвращаем результат
	return string(result)

}

// Метод обработки запросов для главной страницы с POST методом.
func mainPage(res http.ResponseWriter, req *http.Request) {

	// Проверка метода, пропускаем только POST
	if req.Method != http.MethodPost {
		http.Error(res, "Данный запрос не поддерживает выбранный метод.", http.StatusBadRequest)
		return
	}

	// Получаем URL из данных формы
	urlStr := req.FormValue("URL")

	// Выполняем проверку полученного URL на пустую строку и является ли данный URL валидным
	switch {
	case urlStr == "":
		http.Error(res, "URL не может быть пустым", http.StatusBadRequest)
		return
	case !isValidURL(urlStr):
		http.Error(res, "передан не валидный URL", http.StatusBadRequest)
		return
	}

	// Получаем уникальный хещ
	hashString := generateShortURL()

	// Записываем ключ - значение в БД (импровизированную)
	urlStore[hashString] = urlStr

	// Записываем заголовки, присваиваем статус и отдаем ответ сервера
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(URL + "/" + hashString))

}

// Метод обработки запроса получения URL и редиректа на сайт по короткой ссылке.
func urlDetailPage(res http.ResponseWriter, req *http.Request) {

	// Получаем хеш из параметра
	id := req.URL.Path[1:]

	// Смотрим нашу БД (импровизированную), если ничего не находим, возвращаем ошибку
	urlStr, found := urlStore[id]
	if !found {
		http.Error(res, "Сайт не найден", http.StatusBadRequest)
		return
	}

	// Если все ОК, задаем заголовок и делаем редирект
	res.Header().Set("Location", urlStr)
	res.WriteHeader(http.StatusTemporaryRedirect)

}

// Точка входа
func main() {

	mux := http.NewServeMux()

	mux.HandleFunc(`/`, mainPage)
	mux.HandleFunc(`/{id}`, urlDetailPage)

	err := http.ListenAndServe(":"+PORT, mux)
	if err != nil {
		panic(err)
	}

}
