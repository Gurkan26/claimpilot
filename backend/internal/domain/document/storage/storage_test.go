package storage_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/domain/document/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalStorageService_SaveOpenDelete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "claimpilot_storage_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	svc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)

	ctx := context.Background()
	content := "Sample document text for test"
	r := strings.NewReader(content)

	path, size, err := svc.Save(ctx, "contract.pdf", r)
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Equal(t, int64(len(content)), size)

	// Open and verify content
	file, err := svc.Open(ctx, path)
	require.NoError(t, err)
	defer file.Close()

	readBytes, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, content, string(readBytes))

	// Delete and verify gone
	err = svc.Delete(ctx, path)
	require.NoError(t, err)

	_, err = svc.Open(ctx, path)
	assert.Error(t, err)
}
