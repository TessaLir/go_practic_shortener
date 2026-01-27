package repository

import (
	"math/rand"
	"sync"
	"time"
)

// Storage представляет хранилище для URL shortener
type Storage struct {
	urlStore map[string]string
	rng      *rand.Rand
	mu       sync.RWMutex // для потокобезопасности
}

// NewStorage создает новый экземпляр Storage
func NewStorage() *Storage {
	return &Storage{
		urlStore: make(map[string]string),
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Save сохраняет URL по ключу
func (s *Storage) Save(key, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urlStore[key] = url
}

// Get получает URL по ключу. Возвращает URL и флаг существования
func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, exists := s.urlStore[key]
	return url, exists
}

// Exists проверяет существование ключа
func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.urlStore[key]
	return exists
}

// GetRNG возвращает генератор случайных чисел (для использования в service)
func (s *Storage) GetRNG() *rand.Rand {
	return s.rng
}

// Clear очищает хранилище (для тестов)
func (s *Storage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.urlStore = make(map[string]string)
}
