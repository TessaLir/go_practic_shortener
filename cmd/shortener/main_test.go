package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TessaLir/go_practic_shortener/cmd/shortener/config"
	"github.com/TessaLir/go_practic_shortener/internal/handler"
	"github.com/TessaLir/go_practic_shortener/internal/repository"
	"github.com/TessaLir/go_practic_shortener/internal/service"
)

// Создание тестовой конфигурации
func setupTestConfig() *config.Config {
	return &config.Config{
		ServerPort: ":8080",
		BaseURL:    "http://localhost",
	}
}

// Создание тестовых компонентов
func setupTestComponents() (*handler.ShortenerHandler, *repository.Storage) {
	cfg := setupTestConfig()
	storage := repository.NewStorage()
	shortenerService := service.NewShortenerService(storage)
	shortenerHandler := handler.NewShortenerHandler(shortenerService, cfg)
	return shortenerHandler, storage
}

// Тесты для mainPage (POST /)
func TestMainPage(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
		checkStore     bool
	}{
		{
			name:           "POST с валидным URL",
			method:         http.MethodPost,
			body:           "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
			checkStore:     true,
		},
		{
			name:           "POST с пустым телом",
			method:         http.MethodPost,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невозможно распарсить полученный сокращенный URL - URL не может быть пустым",
			checkStore:     false,
		},
		{
			name:           "POST с невалидным URL",
			method:         http.MethodPost,
			body:           "not-a-valid-url",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "передан не валидный URL",
			checkStore:     false,
		},
		{
			name:           "GET запрос (неправильный метод)",
			method:         http.MethodGet,
			body:           "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Данный запрос не поддерживает выбранный метод.",
			checkStore:     false,
		},
		{
			name:           "POST с URL с пробелами",
			method:         http.MethodPost,
			body:           "  https://example.com  ",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost:8080/",
			checkStore:     true,
		},
		{
			name:           "POST с URL без схемы",
			method:         http.MethodPost,
			body:           "example.com",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "передан не валидный URL",
			checkStore:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, storage := setupTestComponents()
			storage.Clear()

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.MainPage()(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, w.Code)
			}

			body := w.Body.String()
			if tt.checkStore {
				// Проверяем, что ответ содержит siteURL и что URL сохранен в хранилище
				if !strings.HasPrefix(body, tt.expectedBody) {
					t.Errorf("Ожидалось, что ответ начинается с %s, получено %s", tt.expectedBody, body)
				}
				// Извлекаем hash из ответа
				hash := strings.TrimPrefix(body, tt.expectedBody)
				if len(hash) != 8 {
					t.Errorf("Ожидался hash длиной 8 символов, получено %d", len(hash))
				}
				// Проверяем, что URL сохранен в хранилище
				storedURL, exists := storage.Get(hash)
				if !exists {
					t.Errorf("URL не был сохранен в хранилище")
				} else {
					expectedURL := strings.TrimSpace(tt.body)
					if storedURL != expectedURL {
						t.Errorf("Ожидался URL %s в хранилище, получен %s", expectedURL, storedURL)
					}
				}
			} else {
				// Для ошибок проверяем текст сообщения
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("Ожидалось сообщение содержащее '%s', получено '%s'", tt.expectedBody, body)
				}
				// Проверяем, что хранилище пустое (проверяем через метод, но это не идеально)
				// В реальности нужно добавить метод Count или проверять по-другому
			}

			// Проверяем Content-Type для успешных запросов
			if tt.expectedStatus == http.StatusCreated {
				contentType := w.Header().Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("Ожидался Content-Type 'text/plain', получен '%s'", contentType)
				}
			}
		})
	}
}

// Тесты для urlDetailPage (GET /{id})
func TestUrlDetailPage(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		setupStore       map[string]string
		expectedStatus   int
		expectedLocation string
		expectedBody     string
	}{
		{
			name:             "GET с существующим ID",
			path:             "/abc12345",
			setupStore:       map[string]string{"abc12345": "https://example.com"},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
			expectedBody:     "",
		},
		{
			name:             "GET с несуществующим ID",
			path:             "/nonexistent",
			setupStore:       map[string]string{},
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
			expectedBody:     "Сайт не найден",
		},
		{
			name:             "GET с пустым ID",
			path:             "/",
			setupStore:       map[string]string{},
			expectedStatus:   http.StatusNotFound,
			expectedLocation: "",
			expectedBody:     "Сайт не найден",
		},
		{
			name:             "GET с другим существующим ID",
			path:             "/xyz98765",
			setupStore:       map[string]string{"xyz98765": "https://google.com"},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://google.com",
			expectedBody:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, storage := setupTestComponents()
			storage.Clear()

			// Заполняем хранилище для теста
			for k, v := range tt.setupStore {
				storage.Save(k, v)
			}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.URLDetailPage()(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Ожидался статус %d, получен %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedLocation != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("Ожидался Location '%s', получен '%s'", tt.expectedLocation, location)
				}
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("Ожидалось сообщение содержащее '%s', получено '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

// Интеграционный тест: создание URL и последующее получение
func TestCreateAndRetrieveURL(t *testing.T) {
	handler, storage := setupTestComponents()
	storage.Clear()
	cfg := setupTestConfig()

	// Создаем короткую ссылку
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://test.com"))
	w1 := httptest.NewRecorder()
	handler.MainPage()(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("Ожидался статус %d при создании, получен %d", http.StatusCreated, w1.Code)
	}

	// Извлекаем hash из ответа
	responseBody := w1.Body.String()
	hash := strings.TrimPrefix(responseBody, cfg.BaseURL+cfg.ServerPort+"/")
	if len(hash) != 8 {
		t.Fatalf("Ожидался hash длиной 8 символов, получено %d", len(hash))
	}

	// Получаем URL по hash
	req2 := httptest.NewRequest(http.MethodGet, "/"+hash, nil)
	w2 := httptest.NewRecorder()
	handler.URLDetailPage()(w2, req2)

	if w2.Code != http.StatusTemporaryRedirect {
		t.Errorf("Ожидался статус %d при получении, получен %d", http.StatusTemporaryRedirect, w2.Code)
	}

	location := w2.Header().Get("Location")
	if location != "https://test.com" {
		t.Errorf("Ожидался Location 'https://test.com', получен '%s'", location)
	}
}
