package repository

import (
	"bufio"
	"fmt"
	"os"
	"sort"
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
		word := normalizeWord(scanner.Text())
		if word != "" {
			storage.words[word] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan stop-list file: %w", err)
	}

	return storage, nil
}

func (s *StopListStorage) IsBanned(word string) bool {
	cleaned := normalizeWord(word)
	if cleaned == "" {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.words[cleaned]
	return exists
}

func (s *StopListStorage) Add(word string) error {
	cleaned := normalizeWord(word)
	if cleaned == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.words[cleaned]; exists {
		return nil
	}

	s.words[cleaned] = struct{}{}

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		delete(s.words, cleaned)
		return fmt.Errorf("failed to open stop-list file for append: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString(cleaned + "\n"); err != nil {
		delete(s.words, cleaned)
		return fmt.Errorf("failed to write word to stop-list file: %w", err)
	}

	return nil
}

func (s *StopListStorage) Remove(word string) error {
	cleaned := normalizeWord(word)
	if cleaned == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.words[cleaned]; !exists {
		return nil
	}

	delete(s.words, cleaned)

	if err := s.rewriteFileLocked(); err != nil {
		return fmt.Errorf("failed to rewrite stop-list file: %w", err)
	}

	return nil
}

func (s *StopListStorage) rewriteFileLocked() error {
	words := make([]string, 0, len(s.words))
	for w := range s.words {
		words = append(words, w)
	}
	sort.Strings(words)

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, w := range words {
		if _, err = writer.WriteString(w + "\n"); err != nil {
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		return err
	}

	return nil
}

func normalizeWord(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}
