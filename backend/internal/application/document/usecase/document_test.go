package usecase_test

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/masterfabric-go/masterfabric/internal/application/document/dto"
	"github.com/masterfabric-go/masterfabric/internal/application/document/usecase"
	docModel "github.com/masterfabric-go/masterfabric/internal/domain/document/model"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/parser"
	"github.com/masterfabric-go/masterfabric/internal/domain/document/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// inMemoryDocRepo implements document.DocumentRepository for unit tests
type inMemoryDocRepo struct {
	mu   sync.Mutex
	docs map[bson.ObjectID]*docModel.Document
}

func newInMemoryDocRepo() *inMemoryDocRepo {
	return &inMemoryDocRepo{docs: make(map[bson.ObjectID]*docModel.Document)}
}

func (r *inMemoryDocRepo) Create(_ context.Context, doc *docModel.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs[doc.ID] = doc
	return nil
}

func (r *inMemoryDocRepo) FindByID(_ context.Context, id bson.ObjectID) (*docModel.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.docs[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return d, nil
}

func (r *inMemoryDocRepo) FindByUserID(_ context.Context, userID uuid.UUID, _, _ int64) ([]*docModel.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*docModel.Document
	for _, d := range r.docs {
		if d.UserID == userID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (r *inMemoryDocRepo) FindByOrganizationID(_ context.Context, orgID uuid.UUID, _, _ int64) ([]*docModel.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var list []*docModel.Document
	for _, d := range r.docs {
		if d.OrganizationID != nil && *d.OrganizationID == orgID {
			list = append(list, d)
		}
	}
	return list, nil
}

func (r *inMemoryDocRepo) UpdateStatus(_ context.Context, id bson.ObjectID, status docModel.DocumentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.Status = status
	}
	return nil
}

func (r *inMemoryDocRepo) UpdateExtractionResult(_ context.Context, id bson.ObjectID, result *docModel.ExtractionResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.ExtractionResult = result
	}
	return nil
}

func (r *inMemoryDocRepo) UpdateContent(_ context.Context, id bson.ObjectID, rawContent, redactedContent string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.docs[id]; ok {
		d.RawContent = rawContent
		d.RedactedContent = redactedContent
	}
	return nil
}

func (r *inMemoryDocRepo) Delete(_ context.Context, id bson.ObjectID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.docs, id)
	return nil
}

func (r *inMemoryDocRepo) DeleteByUserID(_ context.Context, userID uuid.UUID) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, d := range r.docs {
		if d.UserID == userID {
			delete(r.docs, id)
			count++
		}
	}
	return count, nil
}

func TestUploadDocumentUseCase_Execute(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "claimpilot_upload_uc_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	storageSvc, err := storage.NewLocalStorageService(tempDir)
	require.NoError(t, err)

	parserSvc := parser.NewDefaultParser()
	repo := newInMemoryDocRepo()

	uc := usecase.NewUploadDocumentUseCase(usecase.UploadConfig{
		DocRepo: repo,
		Storage: storageSvc,
		Parser:  parserSvc,
	})

	userID := uuid.New()
	content := "Software License Agreement\nCommitted Capacity: 100 seats\nActive Utilization: 41%"

	req := dto.UploadDocumentRequest{
		UserID:      userID,
		FileName:    "datadog_license.txt",
		FileType:    docModel.DocumentTypeOther, // test auto-detection
		MimeType:    "text/plain",
		FileContent: strings.NewReader(content),
		SyncProcess: false,
	}

	ctx := context.Background()
	resp, err := uc.Execute(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "datadog_license.txt", resp.FileName)
	assert.Equal(t, string(docModel.DocumentTypeLicense), resp.FileType)
	assert.Equal(t, string(docModel.DocumentStatusUploaded), resp.Status)

	// Test GetDocumentUseCase
	getUC := usecase.NewGetDocumentUseCase(repo)
	getResp, err := getUC.Execute(ctx, resp.ID)
	require.NoError(t, err)
	assert.Equal(t, resp.ID, getResp.ID)

	// Test ListDocumentsUseCase
	listUC := usecase.NewListDocumentsUseCase(repo)
	listResp, err := listUC.Execute(ctx, userID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, listResp.Items, 1)
	assert.Equal(t, resp.ID, listResp.Items[0].ID)
}
