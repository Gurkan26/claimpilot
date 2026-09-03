package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// StorageService defines the contract for document file storage.
type StorageService interface {
	// Save writes file data from reader and returns the saved storage path and size.
	Save(ctx context.Context, originalFileName string, r io.Reader) (storagePath string, size int64, err error)

	// Open opens the file at the given storage path for reading.
	Open(ctx context.Context, storagePath string) (io.ReadCloser, error)

	// Delete removes the file at the given storage path.
	Delete(ctx context.Context, storagePath string) error
}

// LocalStorageService implements StorageService using the local filesystem.
type LocalStorageService struct {
	baseDir string
}

// NewLocalStorageService creates a local file storage service with the given base directory.
func NewLocalStorageService(baseDir string) (*LocalStorageService, error) {
	if baseDir == "" {
		baseDir = "uploads/documents"
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create storage base directory: %w", err)
	}

	return &LocalStorageService{baseDir: baseDir}, nil
}

func (s *LocalStorageService) Save(_ context.Context, originalFileName string, r io.Reader) (string, int64, error) {
	ext := filepath.Ext(originalFileName)
	cleanName := filepath.Base(originalFileName)
	uniqueName := fmt.Sprintf("%s_%d_%s", uuid.New().String()[:8], time.Now().Unix(), cleanName)
	if ext == "" {
		uniqueName += ".bin"
	}

	fullPath := filepath.Join(s.baseDir, uniqueName)
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", 0, fmt.Errorf("create file %s: %w", fullPath, err)
	}
	defer outFile.Close()

	n, err := io.Copy(outFile, r)
	if err != nil {
		_ = os.Remove(fullPath)
		return "", 0, fmt.Errorf("write file %s: %w", fullPath, err)
	}

	return fullPath, n, nil
}

func (s *LocalStorageService) Open(_ context.Context, storagePath string) (io.ReadCloser, error) {
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, fmt.Errorf("open storage file %s: %w", storagePath, err)
	}
	return file, nil
}

func (s *LocalStorageService) Delete(_ context.Context, storagePath string) error {
	if err := os.Remove(storagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete storage file %s: %w", storagePath, err)
	}
	return nil
}
