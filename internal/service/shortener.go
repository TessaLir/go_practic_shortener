package service

import (
	"net/url"

	"github.com/TessaLir/go_practic_shortener/internal/repository"
)

// ShortenerService содержит бизнес-логику для работы с короткими URL
type ShortenerService struct {
	storage *repository.Storage
}

// NewShortenerService создает новый экземпляр ShortenerService
func NewShortenerService(storage *repository.Storage) *ShortenerService {
	return &ShortenerService{
		storage: storage,
	}
}

// IsValidURL проверяет переданный URL на валидность
func (s *ShortenerService) IsValidURL(str string) bool {
	parsedURL, err := url.Parse(str)
	if err != nil {
		return false
	}
	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// GenerateShortURL генерирует рандомную короткую строку длиной в 8 символов
func (s *ShortenerService) GenerateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	rng := s.storage.GetRNG()
	result := make([]byte, length)

	for {
		// Генерируем случайную строку
		for i := range result {
			result[i] = charset[rng.Intn(len(charset))]
		}

		// Если сгенерированная строка - уникальна, выходим из цикла
		if !s.storage.Exists(string(result)) {
			break
		}
	}

	return string(result)
}

// SaveURL сохраняет URL и возвращает короткий ключ
func (s *ShortenerService) SaveURL(originalURL string) string {
	shortKey := s.GenerateShortURL()
	s.storage.Save(shortKey, originalURL)
	return shortKey
}

// GetURL получает оригинальный URL по короткому ключу
func (s *ShortenerService) GetURL(shortKey string) (string, bool) {
	return s.storage.Get(shortKey)
}
