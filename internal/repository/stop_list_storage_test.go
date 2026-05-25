package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStopListStorage_AddAndIsBanned(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "stop-list.txt")

	storage, err := NewStopListStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	err = storage.Add("  РЕКЛАМА ")
	if err != nil {
		t.Fatalf("failed to add word: %v", err)
	}

	if !storage.IsBanned("реклама") {
		t.Error("expected word 'реклама' to be banned")
	}
	if !storage.IsBanned("Реклама") {
		t.Error("expected case-insensitive match for banned word")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(data) != "реклама\n" {
		t.Errorf("unexpected file content: %q", string(data))
	}
}

func TestStopListStorage_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "stop-list.txt")

	storage, err := NewStopListStorage(filePath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	_ = storage.Add("спам")
	if !storage.IsBanned("спам") {
		t.Fatal("word should be banned")
	}

	err = storage.Remove("спам")
	if err != nil {
		t.Fatalf("failed to remove word: %v", err)
	}

	if storage.IsBanned("спам") {
		t.Error("word should not be banned after removal")
	}
}
