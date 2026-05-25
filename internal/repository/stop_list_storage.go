package repository

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type StopListStorage struct {
	mu       sync.RWMutex
	words    map[string]struct{}
	filePath string
}

func NewStopListStorage(filePath string) (*StopListStorage, error) {
	storage := &StopListStorage{
		words:    make(map[string]struct{}),
		filePath: filePath,
	}

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open stop-list file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if word != "" { // исправлено: было == ""
			storage.words[word] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan stop-list file: %w", err)
	}

	return storage, nil
}

func (s *StopListStorage) IsBanned(word string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.words[strings.ToLower(strings.TrimSpace(word))]
	return exists
}

func (s *StopListStorage) Add(word string) error {
	cleanedWord := strings.ToLower(strings.TrimSpace(word))
	if cleanedWord == "" {
		return nil
	}

	s.mu.Lock()
	if _, exists := s.words[cleanedWord]; exists {
		s.mu.Unlock()
		return nil
	}
	s.words[cleanedWord] = struct{}{}
	s.mu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open stop-list file for append: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(cleanedWord + "\n"); err != nil {
		return fmt.Errorf("failed to write word to stop-list file: %w", err)
	}

	return nil
}

func (s *StopListStorage) Remove(word string) error {
	cleanedWord := strings.ToLower(strings.TrimSpace(word))
	if cleanedWord == "" {
		return nil
	}

	s.mu.Lock()
	delete(s.words, cleanedWord)
	s.mu.Unlock()

	// Вариант простой: удаляем только из RAM.
	// Если нужна полная персистентность удаления — переписывай файл целиком.
	return nil
}
