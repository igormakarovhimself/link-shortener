package main

import (
	"context"
	"link-shortener/internal/config"
	"link-shortener/internal/repository"
	"os"
	"testing"
)

func TestRepositorySelection(t *testing.T) {
	t.Run("LocalRepository when no config", func(t *testing.T) {
		os.Clearenv()
		cfg := config.SetupConfig()

		if cfg.DatabaseDSN != "" {
			t.Error("Expected empty DatabaseDSN")
		}
		if cfg.FileStoragePath != "short-url.json" {
			t.Error("Expected default FileStoragePath")
		}

		repo := repository.NewLocalRepository()
		if repo == nil {
			t.Error("Expected LocalRepository to be created")
		}
	})

	t.Run("FileRepository initialization", func(t *testing.T) {
		testFile := "test-storage.json"
		defer os.Remove(testFile)

		repo, err := repository.NewFileRepository(testFile)
		if err != nil {
			t.Fatalf("Failed to create FileRepository: %v", err)
		}
		defer repo.Close()

		ctx := context.Background()
		err = repo.Save(ctx, "test123", "https://example.com", "test-user")
		if err != nil {
			t.Errorf("Failed to save URL: %v", err)
		}

		url, err := repo.Get(ctx, "test123")
		if err != nil {
			t.Errorf("Failed to get URL: %v", err)
		}
		if url != "https://example.com" {
			t.Errorf("Expected https://example.com, got %s", url)
		}
	})
}
